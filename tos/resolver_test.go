package tos

import (
	"container/heap"
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func newTestResolver(expiration time.Duration, lookup func(context.Context, string) ([]net.IP, error)) *resolver {
	r := newResolver(expiration)
	r.lookupIP = lookup
	r.refreshInterval = time.Minute
	r.refreshTimeout = time.Second
	return r
}

func setRefreshDue(r *resolver, host string) {
	r.cache.lock.Lock()
	item := r.cache.data[host]
	item.refreshAt = time.Now().Add(-time.Second)
	r.cache.data[host] = item
	r.cache.lock.Unlock()
}

func expireCacheEntryForCleanup(r *resolver, host string) {
	r.cache.lock.Lock()
	defer r.cache.lock.Unlock()

	expired := time.Now().Add(-time.Second)
	item, ok := r.cache.data[host]
	if !ok {
		panic("cache entry not found: " + host)
	}
	item.expireAt = expired
	r.cache.data[host] = item

	heapIndex := -1
	for i, tracked := range *r.cache.heap {
		if tracked.host == host {
			tracked.expireAt = expired
			heapIndex = i
			break
		}
	}
	if heapIndex < 0 {
		panic("cache heap entry not found: " + host)
	}
	heap.Fix(r.cache.heap, heapIndex)
	r.cache.cleanTime = expired
}

func cachedItem(r *resolver, host string) (cacheItem, bool) {
	r.cache.lock.RLock()
	defer r.cache.lock.RUnlock()
	item, ok := r.cache.data[host]
	return item, ok
}

func waitForCondition(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}

func waitForSignal(t *testing.T, signal <-chan struct{}, message string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(time.Second):
		t.Fatal(message)
	}
}

func TestNewResolverDoesNotStartBackgroundGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()
	resolvers := make([]*resolver, 100)
	for i := range resolvers {
		resolvers[i] = newResolver(time.Minute)
	}
	runtime.Gosched()
	after := runtime.NumGoroutine()
	for _, r := range resolvers {
		r.Close()
	}

	// Allow a small amount of runtime noise while still detecting the previous
	// one-per-resolver refresh loop.
	require.Less(t, after-before, 10)
}

func TestResolverCachesAndRemovesAddresses(t *testing.T) {
	var lookups int32
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		atomic.AddInt32(&lookups, 1)
		return []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")}, nil
	})
	defer r.Close()

	ipList, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, []string{"192.0.2.1", "192.0.2.2"}, ipList)

	_, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, int32(1), atomic.LoadInt32(&lookups))

	for _, ip := range ipList {
		r.Remove("example.com", ip)
	}
	_, exists := cachedItem(r, "example.com")
	require.False(t, exists)
}

func TestResolverFreshCacheHitDoesNotTakeWriteLock(t *testing.T) {
	var lookups int32
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		atomic.AddInt32(&lookups, 1)
		return []net.IP{net.ParseIP("192.0.2.1")}, nil
	})
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)

	result := make(chan error, 1)
	r.cache.lock.RLock()
	go func() {
		ipList, err := r.GetIpList(context.Background(), "example.com")
		if err == nil && (len(ipList) != 1 || ipList[0] != "192.0.2.1") {
			err = fmt.Errorf("unexpected IP list: %v", ipList)
		}
		result <- err
	}()

	select {
	case err = <-result:
		r.cache.lock.RUnlock()
		require.NoError(t, err)
	case <-time.After(time.Second):
		r.cache.lock.RUnlock()
		t.Fatal("fresh cache hit waited for the cache write lock")
	}
	require.Equal(t, int32(1), atomic.LoadInt32(&lookups))
}

func TestResolverConcurrentAccessStartsSingleRefresh(t *testing.T) {
	var lookups int32
	refreshStarted := make(chan struct{})
	refreshFinished := make(chan struct{})
	releaseRefresh := make(chan struct{})
	r := newTestResolver(time.Minute, func(ctx context.Context, _ string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		defer close(refreshFinished)
		close(refreshStarted)
		select {
		case <-releaseRefresh:
			return []net.IP{net.ParseIP("192.0.2.2")}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")

	const total = 20
	var wg sync.WaitGroup
	results := make(chan error, total)
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			ipList, err := r.GetIpList(context.Background(), "example.com")
			if err != nil {
				results <- err
				return
			}
			if len(ipList) != 1 || ipList[0] != "192.0.2.1" {
				results <- fmt.Errorf("unexpected IP list: %v", ipList)
			}
		}()
	}
	waitForSignal(t, refreshStarted, "asynchronous refresh did not start")
	accessDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(accessDone)
	}()
	waitForSignal(t, accessDone, "concurrent cache access did not finish")
	close(releaseRefresh)
	waitForSignal(t, refreshFinished, "asynchronous refresh did not finish")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && len(item.ipList) == 1 && item.ipList[0] == "192.0.2.2"
	})
	close(results)

	for err := range results {
		require.NoError(t, err)
	}
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))
}

func TestResolverDoesNotOverlapRefreshes(t *testing.T) {
	var lookups int32
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	refreshFinished := make(chan struct{})
	r := newTestResolver(time.Minute, func(ctx context.Context, _ string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		defer close(refreshFinished)
		close(refreshStarted)
		select {
		case <-releaseRefresh:
			return []net.IP{net.ParseIP("192.0.2.2")}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	r.refreshInterval = 10 * time.Millisecond
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")
	_, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	waitForSignal(t, refreshStarted, "asynchronous refresh did not start")

	// Even after another refresh interval elapses, the in-flight refresh owns
	// the host until it finishes.
	time.Sleep(20 * time.Millisecond)
	_, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))

	close(releaseRefresh)
	waitForSignal(t, refreshFinished, "asynchronous refresh did not finish")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && !item.refreshing && len(item.ipList) == 1 && item.ipList[0] == "192.0.2.2"
	})
}

func TestResolverRefreshesOnlyAfterCacheAccess(t *testing.T) {
	var lookups int32
	refreshDone := make(chan struct{})
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		call := atomic.AddInt32(&lookups, 1)
		if call == 2 {
			defer close(refreshDone)
			return []net.IP{net.ParseIP("192.0.2.2")}, nil
		}
		return []net.IP{net.ParseIP("192.0.2.1")}, nil
	})
	defer r.Close()

	first, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, []string{"192.0.2.1"}, first)
	setRefreshDue(r, "example.com")

	// An expired refresh deadline alone must not start background work.
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int32(1), atomic.LoadInt32(&lookups))

	stale, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale, "refresh must not block the current dial")
	waitForSignal(t, refreshDone, "asynchronous refresh did not finish")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && len(item.ipList) == 1 && item.ipList[0] == "192.0.2.2"
	})

	refreshed, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, []string{"192.0.2.2"}, refreshed)
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))
}

func TestResolverOldRefreshDoesNotOverwriteNewerCache(t *testing.T) {
	r := newTestResolver(time.Minute, LookupIP)
	defer r.Close()
	r.cache.Put("example.com", []string{"192.0.2.1"}, r.refreshInterval)
	setRefreshDue(r, "example.com")
	revision, claimed := r.cache.ClaimRefresh("example.com", time.Now(), r.refreshInterval)
	require.True(t, claimed)

	// Simulate the cached address failing while the old refresh is in flight.
	r.Remove("example.com", "192.0.2.1")
	r.cache.Put("example.com", []string{"192.0.2.3"}, r.refreshInterval)
	require.False(t, r.cache.putIfRevision("example.com", revision, []string{"192.0.2.2"}, r.refreshInterval))

	current, exists := r.cache.Get("example.com")
	require.True(t, exists)
	require.Equal(t, []string{"192.0.2.3"}, current)
}

func TestResolverRefreshFailureKeepsStaleAddresses(t *testing.T) {
	var lookups int32
	refreshDone := make(chan struct{})
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		defer close(refreshDone)
		return nil, errors.New("dns unavailable")
	})
	defer r.Close()

	first, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")
	stale, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale)

	waitForSignal(t, refreshDone, "failed refresh did not finish")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && item.keepAlive
	})

	r.cache.lock.Lock()
	item := r.cache.data["example.com"]
	item.expireAt = time.Now().Add(-time.Second)
	r.cache.data["example.com"] = item
	r.cache.lock.Unlock()

	stale, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale)
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))
}

func TestResolverRefreshPanicIsRecoveredAndCanRetry(t *testing.T) {
	var lookups int32
	panicStarted := make(chan struct{})
	retryFinished := make(chan struct{})
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		switch atomic.AddInt32(&lookups, 1) {
		case 1:
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		case 2:
			close(panicStarted)
			panic("dns refresh panic")
		default:
			close(retryFinished)
			return []net.IP{net.ParseIP("192.0.2.2")}, nil
		}
	})
	defer r.Close()

	first, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")
	stale, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale)
	waitForSignal(t, panicStarted, "panicking refresh did not start")

	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && item.keepAlive && !item.refreshing
	})

	setRefreshDue(r, "example.com")
	stale, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale)
	waitForSignal(t, retryFinished, "refresh did not retry after recovering from panic")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && !item.keepAlive && !item.refreshing && len(item.ipList) == 1 && item.ipList[0] == "192.0.2.2"
	})
	require.Equal(t, int32(3), atomic.LoadInt32(&lookups))
}

func TestResolverCleanupPreservesEntryDuringRefresh(t *testing.T) {
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	refreshFinished := make(chan struct{})
	r := newTestResolver(time.Hour, func(ctx context.Context, _ string) ([]net.IP, error) {
		close(refreshStarted)
		defer close(refreshFinished)
		select {
		case <-releaseRefresh:
			return nil, errors.New("dns unavailable")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	defer r.Close()

	r.cache.Put("refreshing.example.com", []string{"192.0.2.1"}, r.refreshInterval)
	setRefreshDue(r, "refreshing.example.com")
	stale, err := r.GetIpList(context.Background(), "refreshing.example.com")
	require.NoError(t, err)
	require.Equal(t, []string{"192.0.2.1"}, stale)
	waitForSignal(t, refreshStarted, "asynchronous refresh did not start")

	expireCacheEntryForCleanup(r, "refreshing.example.com")
	// A successful write for another host triggers ordinary expiry cleanup.
	r.cache.Put("other.example.com", []string{"192.0.2.2"}, r.refreshInterval)

	item, exists := cachedItem(r, "refreshing.example.com")
	require.True(t, exists, "cleanup deleted an in-flight refresh")
	require.True(t, item.refreshing)
	_, usable := r.cache.Get("refreshing.example.com")
	require.False(t, usable, "cleanup extended the logical cache lifetime")

	r.cache.lock.RLock()
	dataLen := len(r.cache.data)
	heapLen := r.cache.heap.Len()
	trackedCount := 0
	for _, tracked := range *r.cache.heap {
		if tracked.host == "refreshing.example.com" {
			trackedCount++
		}
	}
	r.cache.lock.RUnlock()
	require.Equal(t, dataLen, heapLen)
	require.Equal(t, 1, trackedCount)

	close(releaseRefresh)
	waitForSignal(t, refreshFinished, "failed refresh did not finish")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "refreshing.example.com")
		return ok && item.keepAlive && !item.refreshing
	})

	stale, exists = r.cache.Get("refreshing.example.com")
	require.True(t, exists)
	require.Equal(t, []string{"192.0.2.1"}, stale)
}

func TestResolverRefreshingEntriesDoNotStarveExpiryCleanup(t *testing.T) {
	r := newTestResolver(time.Hour, LookupIP)
	defer r.Close()

	const refreshingCount = 6
	for i := 0; i < refreshingCount; i++ {
		host := fmt.Sprintf("refreshing-%d.example.com", i)
		r.cache.Put(host, []string{"192.0.2.1"}, r.refreshInterval)
		setRefreshDue(r, host)
		_, claimed := r.cache.ClaimRefresh(host, time.Now(), r.refreshInterval)
		require.True(t, claimed)
		expireCacheEntryForCleanup(r, host)
	}

	r.cache.Put("expired.example.com", []string{"192.0.2.2"}, r.refreshInterval)
	expireCacheEntryForCleanup(r, "expired.example.com")
	// Trigger cleanup with more refreshing entries than the per-pass cleanup limit.
	r.cache.Put("current.example.com", []string{"192.0.2.3"}, r.refreshInterval)

	_, exists := cachedItem(r, "expired.example.com")
	require.False(t, exists, "refreshing entries starved ordinary expiry cleanup")

	r.cache.lock.RLock()
	dataLen := len(r.cache.data)
	heapLen := r.cache.heap.Len()
	refreshingEntries := 0
	trackedCounts := make([]int, 0, refreshingCount)
	for host, item := range r.cache.data {
		if item.refreshing {
			refreshingEntries++
			trackedCount := 0
			for _, tracked := range *r.cache.heap {
				if tracked.host == host {
					trackedCount++
				}
			}
			trackedCounts = append(trackedCounts, trackedCount)
		}
	}
	r.cache.lock.RUnlock()
	require.Equal(t, dataLen, heapLen)
	require.Equal(t, refreshingCount, refreshingEntries)
	for _, trackedCount := range trackedCounts {
		require.Equal(t, 1, trackedCount)
	}
}

func TestResolverCleanupPreservesEntryDuringSynchronousRefresh(t *testing.T) {
	lookupStarted := make(chan struct{})
	releaseLookup := make(chan struct{})
	r := newTestResolver(time.Hour, func(ctx context.Context, _ string) ([]net.IP, error) {
		close(lookupStarted)
		select {
		case <-releaseLookup:
			return nil, errors.New("dns unavailable")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	defer r.Close()

	r.cache.Put("expired.example.com", []string{"192.0.2.1"}, r.refreshInterval)
	expireCacheEntryForCleanup(r, "expired.example.com")

	type lookupResult struct {
		ipList []string
		err    error
	}
	result := make(chan lookupResult, 1)
	go func() {
		ipList, err := r.GetIpList(context.Background(), "expired.example.com")
		result <- lookupResult{ipList: ipList, err: err}
	}()
	waitForSignal(t, lookupStarted, "synchronous refresh did not start")

	// A successful write for another host triggers ordinary expiry cleanup.
	r.cache.Put("other.example.com", []string{"192.0.2.2"}, r.refreshInterval)
	item, exists := cachedItem(r, "expired.example.com")
	require.True(t, exists, "cleanup deleted a synchronously refreshed entry")
	require.Equal(t, 1, item.stalePins)

	close(releaseLookup)
	var lookup lookupResult
	select {
	case lookup = <-result:
	case <-time.After(time.Second):
		t.Fatal("synchronous refresh did not finish")
	}
	require.NoError(t, lookup.err)
	require.Equal(t, []string{"192.0.2.1"}, lookup.ipList)
	item, exists = cachedItem(r, "expired.example.com")
	require.True(t, exists)
	require.True(t, item.keepAlive)
	require.Equal(t, 0, item.stalePins)
}

func TestResolverSynchronousRefreshPinReleasedOnCancellation(t *testing.T) {
	lookupStarted := make(chan struct{})
	r := newTestResolver(time.Hour, func(ctx context.Context, _ string) ([]net.IP, error) {
		close(lookupStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	defer r.Close()

	r.cache.Put("expired.example.com", []string{"192.0.2.1"}, r.refreshInterval)
	expireCacheEntryForCleanup(r, "expired.example.com")
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := r.GetIpList(ctx, "expired.example.com")
		result <- err
	}()
	waitForSignal(t, lookupStarted, "synchronous refresh did not start")

	r.cache.Put("other.example.com", []string{"192.0.2.2"}, r.refreshInterval)
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("canceled synchronous refresh did not finish")
	}

	item, exists := cachedItem(r, "expired.example.com")
	require.True(t, exists)
	require.False(t, item.keepAlive)
	require.Equal(t, 0, item.stalePins)
	_, usable := r.cache.Get("expired.example.com")
	require.False(t, usable)
}

func TestResolverRemoveDuringAsyncRefreshPreservesRemainingAddresses(t *testing.T) {
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	refreshFinished := make(chan struct{})
	r := newTestResolver(time.Hour, func(ctx context.Context, _ string) ([]net.IP, error) {
		close(refreshStarted)
		defer close(refreshFinished)
		select {
		case <-releaseRefresh:
			return nil, errors.New("dns unavailable")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	defer r.Close()

	r.cache.Put("refreshing.example.com", []string{"192.0.2.1", "192.0.2.2"}, r.refreshInterval)
	setRefreshDue(r, "refreshing.example.com")
	_, err := r.GetIpList(context.Background(), "refreshing.example.com")
	require.NoError(t, err)
	waitForSignal(t, refreshStarted, "asynchronous refresh did not start")

	expireCacheEntryForCleanup(r, "refreshing.example.com")
	r.Remove("refreshing.example.com", "192.0.2.1")
	r.cache.Put("other.example.com", []string{"192.0.2.3"}, r.refreshInterval)

	item, exists := cachedItem(r, "refreshing.example.com")
	require.True(t, exists)
	require.True(t, item.keepAlive)
	require.False(t, item.refreshing)
	require.Equal(t, []string{"192.0.2.2"}, item.ipList)

	close(releaseRefresh)
	waitForSignal(t, refreshFinished, "failed refresh did not finish")
	remaining, exists := r.cache.Get("refreshing.example.com")
	require.True(t, exists)
	require.Equal(t, []string{"192.0.2.2"}, remaining)
}

func TestResolverRemoveDuringSynchronousRefreshPreservesRemainingAddresses(t *testing.T) {
	lookupStarted := make(chan struct{})
	releaseLookup := make(chan struct{})
	r := newTestResolver(time.Hour, func(ctx context.Context, _ string) ([]net.IP, error) {
		close(lookupStarted)
		select {
		case <-releaseLookup:
			return nil, errors.New("dns unavailable")
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	defer r.Close()

	r.cache.Put("expired.example.com", []string{"192.0.2.1", "192.0.2.2"}, r.refreshInterval)
	expireCacheEntryForCleanup(r, "expired.example.com")
	type lookupResult struct {
		ipList []string
		err    error
	}
	result := make(chan lookupResult, 1)
	go func() {
		ipList, err := r.GetIpList(context.Background(), "expired.example.com")
		result <- lookupResult{ipList: ipList, err: err}
	}()
	waitForSignal(t, lookupStarted, "synchronous refresh did not start")

	r.Remove("expired.example.com", "192.0.2.1")
	r.cache.Put("other.example.com", []string{"192.0.2.3"}, r.refreshInterval)
	close(releaseLookup)

	select {
	case lookup := <-result:
		require.NoError(t, lookup.err)
		require.Equal(t, []string{"192.0.2.2"}, lookup.ipList)
	case <-time.After(time.Second):
		t.Fatal("synchronous refresh did not finish")
	}
	item, exists := cachedItem(r, "expired.example.com")
	require.True(t, exists)
	require.True(t, item.keepAlive)
	require.Equal(t, 0, item.stalePins)
	require.Equal(t, []string{"192.0.2.2"}, item.ipList)
}

func TestResolverCleanupKeepsStaleAddressesAfterRefreshFailure(t *testing.T) {
	r := newTestResolver(20*time.Millisecond, LookupIP)
	defer r.Close()

	r.cache.Put("stale.example.com", []string{"192.0.2.1"}, r.refreshInterval)
	_, revision, exists := r.cache.GetStale("stale.example.com")
	require.True(t, exists)
	require.True(t, r.cache.setKeepAliveIfRevision("stale.example.com", revision, time.Now().Add(r.refreshInterval)))

	expireCacheEntryForCleanup(r, "stale.example.com")

	// A successful lookup for another host can trigger expiry cleanup.
	r.cache.Put("other.example.com", []string{"192.0.2.2"}, r.refreshInterval)

	stale, exists := r.cache.Get("stale.example.com")
	require.True(t, exists)
	require.Equal(t, []string{"192.0.2.1"}, stale)
}

func TestResolverRepeatedUpdatesKeepStaleAddressesAfterRefreshFailure(t *testing.T) {
	r := newTestResolver(time.Hour, LookupIP)
	defer r.Close()

	r.cache.Put("stale.example.com", []string{"192.0.2.1"}, r.refreshInterval)
	_, revision, exists := r.cache.GetStale("stale.example.com")
	require.True(t, exists)
	require.True(t, r.cache.setKeepAliveIfRevision("stale.example.com", revision, time.Now().Add(r.refreshInterval)))

	// Repeated updates of one host must not create duplicate heap nodes or evict
	// the stale fallback for another host.
	for i := 0; i < DefaultCacheCap; i++ {
		r.cache.Put("other.example.com", []string{"192.0.2.2"}, r.refreshInterval)
	}

	stale, exists := r.cache.Get("stale.example.com")
	require.True(t, exists)
	require.Equal(t, []string{"192.0.2.1"}, stale)
	require.Equal(t, 2, r.cache.heap.Len())
}

func TestResolverCapacityRemainsBoundedWithKeepAliveEntries(t *testing.T) {
	r := newTestResolver(time.Hour, LookupIP)
	defer r.Close()

	for i := 0; i < DefaultCacheCap+25; i++ {
		host := fmt.Sprintf("stale-%03d.example.com", i)
		r.cache.Put(host, []string{"192.0.2.1"}, r.refreshInterval)
		_, revision, exists := r.cache.GetStale(host)
		require.True(t, exists)
		require.True(t, r.cache.setKeepAliveIfRevision(host, revision, time.Now().Add(r.refreshInterval)))
	}

	r.cache.lock.RLock()
	dataLen := len(r.cache.data)
	heapLen := r.cache.heap.Len()
	keepAliveCount := 0
	for _, item := range r.cache.data {
		if item.keepAlive {
			keepAliveCount++
		}
	}
	r.cache.lock.RUnlock()

	require.Equal(t, DefaultCacheCap, dataLen)
	require.Equal(t, dataLen, heapLen)
	require.Equal(t, dataLen, keepAliveCount)
}

func TestResolverExpiredEntryFallsBackToStaleAddresses(t *testing.T) {
	var lookups int32
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		return nil, errors.New("dns unavailable")
	})
	defer r.Close()

	first, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	r.cache.lock.Lock()
	item := r.cache.data["example.com"]
	item.expireAt = time.Now().Add(-time.Second)
	item.refreshAt = time.Now().Add(-time.Second)
	r.cache.data["example.com"] = item
	r.cache.lock.Unlock()

	stale, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, first, stale)
	item, exists := cachedItem(r, "example.com")
	require.True(t, exists)
	require.True(t, item.keepAlive)
	require.True(t, item.refreshAt.After(time.Now()))
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))
}

func TestResolverCanceledLookupDoesNotKeepStaleAddressesAlive(t *testing.T) {
	r := newTestResolver(time.Minute, func(ctx context.Context, _ string) ([]net.IP, error) {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return []net.IP{net.ParseIP("192.0.2.1")}, nil
	})
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	r.cache.lock.Lock()
	item := r.cache.data["example.com"]
	item.expireAt = time.Now().Add(-time.Second)
	r.cache.data["example.com"] = item
	r.cache.lock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = r.GetIpList(ctx, "example.com")
	require.ErrorIs(t, err, context.Canceled)
	item, exists := cachedItem(r, "example.com")
	require.True(t, exists)
	require.False(t, item.keepAlive)
}

func TestResolverExpiredEntryRefreshesSynchronously(t *testing.T) {
	var lookups int32
	r := newTestResolver(time.Minute, func(context.Context, string) ([]net.IP, error) {
		call := atomic.AddInt32(&lookups, 1)
		if call == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		return []net.IP{net.ParseIP("192.0.2.2")}, nil
	})
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	r.cache.lock.Lock()
	item := r.cache.data["example.com"]
	item.expireAt = time.Now().Add(-time.Second)
	r.cache.data["example.com"] = item
	r.cache.lock.Unlock()

	refreshed, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	require.Equal(t, []string{"192.0.2.2"}, refreshed)
	require.Equal(t, int32(2), atomic.LoadInt32(&lookups))
}

func TestResolverCloseCancelsInFlightRefresh(t *testing.T) {
	var lookups int32
	refreshStarted := make(chan struct{})
	refreshCanceled := make(chan struct{})
	r := newTestResolver(time.Minute, func(ctx context.Context, _ string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		close(refreshStarted)
		<-ctx.Done()
		close(refreshCanceled)
		return nil, ctx.Err()
	})
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")
	_, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	waitForSignal(t, refreshStarted, "asynchronous refresh did not start")
	r.Close()

	waitForSignal(t, refreshCanceled, "Close did not cancel the in-flight refresh")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && !item.refreshing
	})
	item, exists := cachedItem(r, "example.com")
	require.True(t, exists)
	require.False(t, item.keepAlive)
	require.NotPanics(t, r.Close)
}

func TestResolverRefreshIsBoundedByTimeout(t *testing.T) {
	var lookups int32
	refreshFinished := make(chan struct{})
	r := newTestResolver(time.Minute, func(ctx context.Context, _ string) ([]net.IP, error) {
		if atomic.AddInt32(&lookups, 1) == 1 {
			return []net.IP{net.ParseIP("192.0.2.1")}, nil
		}
		<-ctx.Done()
		close(refreshFinished)
		return nil, ctx.Err()
	})
	r.refreshTimeout = 20 * time.Millisecond
	defer r.Close()

	_, err := r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)
	setRefreshDue(r, "example.com")
	_, err = r.GetIpList(context.Background(), "example.com")
	require.NoError(t, err)

	waitForSignal(t, refreshFinished, "asynchronous refresh exceeded its timeout")
	waitForCondition(t, func() bool {
		item, ok := cachedItem(r, "example.com")
		return ok && !item.refreshing && item.keepAlive
	})
}
