package tos

import (
	"container/heap"
	"context"
	"net"
	"sync"
	"time"
)

const (
	DefaultCacheCap = 100
	VolceHostSuffix = "volces.com"
	HostSplitSep    = "."
	HostSplitLength = 4
)

var RefreshInterval = time.Second * 30

const dnsRefreshTimeout = 10 * time.Second

type cacheItem struct {
	host       string
	ipList     []string
	expireAt   time.Time
	refreshAt  time.Time
	revision   uint64
	heapIndex  int
	keepAlive  bool
	refreshing bool
	stalePins  int
}
type priorityQueue []*cacheItem

func (p priorityQueue) Len() int {
	return len(p)
}

func (p priorityQueue) Peek() *cacheItem {
	if p.Len() > 0 {
		return p[0]
	}
	return nil
}

func (p priorityQueue) Less(i, j int) bool {
	return p[i].expireAt.Before(p[j].expireAt)
}

func (p priorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	p[i].heapIndex = i
	p[j].heapIndex = j
}

func (p *priorityQueue) Push(x interface{}) {
	n := len(*p)
	item := x.(*cacheItem)
	item.heapIndex = n
	*p = append(*p, item)
}

func (p *priorityQueue) Pop() interface{} {
	old := *p
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.heapIndex = -1
	*p = old[0 : n-1]
	return item
}

type cache struct {
	lock       sync.RWMutex
	heap       *priorityQueue
	cleanTime  time.Time
	data       map[string]cacheItem
	expiration time.Duration
	revision   uint64
}

// nextRevisionLocked must be called while c.lock is held for writing.
func (c *cache) nextRevisionLocked() uint64 {
	c.revision++
	return c.revision
}

func (c *cache) setKeepAliveIfRevision(key string, revision uint64, nextRefresh time.Time) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok || data.revision != revision {
		return false
	}
	data.keepAlive = true
	data.refreshing = false
	data.refreshAt = nextRefresh
	data.stalePins = 0
	data.revision = c.nextRevisionLocked()
	c.data[key] = data
	return true
}

// ClaimRefresh advances the next refresh time and returns a cache revision when
// the caller should refresh key. Updating refreshAt under the cache lock
// prevents concurrent dials from starting duplicate refresh goroutines.
func (c *cache) ClaimRefresh(key string, now time.Time, interval time.Duration) (uint64, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok || data.refreshing || now.Before(data.refreshAt) {
		return 0, false
	}
	data.refreshing = true
	data.refreshAt = now.Add(interval)
	data.stalePins = 0
	data.revision = c.nextRevisionLocked()
	c.data[key] = data
	return data.revision, true
}

// GetStale returns a cached value even after it expires. It is used only as a
// fallback when a DNS refresh fails.
func (c *cache) GetStale(key string) ([]string, uint64, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	data, ok := c.data[key]
	if !ok {
		return nil, 0, false
	}
	ipList := make([]string, len(data.ipList))
	copy(ipList, data.ipList)
	return ipList, data.revision, true
}

// getStaleAndPin prevents ordinary expiry cleanup from removing the stale
// fallback while a synchronous DNS lookup is using it. The returned revision
// makes release safe when another lookup replaces the cache entry first.
func (c *cache) getStaleAndPin(key string) ([]string, uint64, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok {
		return nil, 0, false
	}
	data.stalePins++
	c.data[key] = data
	ipList := make([]string, len(data.ipList))
	copy(ipList, data.ipList)
	return ipList, data.revision, true
}

func (c *cache) releaseStalePinIfRevision(key string, revision uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok || data.revision != revision || data.stalePins == 0 {
		return
	}
	data.stalePins--
	c.data[key] = data
}

func (c *cache) Remove(key string, removeIp string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok {
		return
	}
	value := make([]string, 0, len(data.ipList))
	for _, ip := range data.ipList {
		if ip == removeIp {
			continue
		}
		value = append(value, ip)
	}

	// 没有有效的 IP 将缓存删除
	if len(value) == 0 {
		c.removeHeapItems(key)
		delete(c.data, key)
		return
	}

	hadInFlightLookup := data.refreshing || data.stalePins > 0
	data.ipList = value
	data.refreshing = false
	data.stalePins = 0
	if hadInFlightLookup {
		// Removing an address invalidates the in-flight lookup revision. Preserve
		// the remaining addresses as fallback because that lookup can no longer
		// publish either a refreshed value or its failure state.
		data.keepAlive = true
	}
	data.revision = c.nextRevisionLocked()
	c.data[key] = data

}

func (c *cache) GetWithRefreshState(key string, now time.Time) ([]string, bool, bool) {
	c.lock.RLock()
	data, ok := c.data[key]
	c.lock.RUnlock()
	if !ok {
		return nil, false, false
	}
	if !data.keepAlive && data.expireAt.Before(now) {
		return nil, false, false
	}
	// 返回拷贝，避免调用方对 cache 内部切片做原地修改（例如 rand.Shuffle）
	// 造成多 goroutine 间的 data race。
	ipList := make([]string, len(data.ipList))
	copy(ipList, data.ipList)
	refreshDue := !data.refreshing && !now.Before(data.refreshAt)
	return ipList, refreshDue, true
}

func (c *cache) Get(key string) ([]string, bool) {
	ipList, _, ok := c.GetWithRefreshState(key, time.Now())
	return ipList, ok
}

// removeHeapItems removes heap nodes tracking a cache entry. The
// cache is small and bounded, so a linear scan keeps the data representation
// simple while ensuring repeated updates cannot leave stale duplicate nodes.
func (c *cache) removeHeapItems(key string) {
	for {
		index := -1
		for i, item := range *c.heap {
			if item.host == key {
				index = i
				break
			}
		}
		if index < 0 {
			return
		}
		heap.Remove(c.heap, index)
	}
}

func (c *cache) evictOverflow() {
	for len(c.data) > DefaultCacheCap && c.heap.Len() > 0 {
		item := heap.Pop(c.heap).(*cacheItem)
		data, ok := c.data[item.host]
		if !ok || data.expireAt != item.expireAt {
			continue
		}
		// keepAlive protects stale addresses from time-based cleanup, but it must
		// not bypass the cache's hard capacity limit.
		delete(c.data, item.host)
	}
}

func (c *cache) cleanCache() {
	now := time.Now()
	c.cleanTime = now.Add(c.expiration)
	maxCleanCount := 5
	maxScanCount := c.heap.Len()
	inFlightItems := make([]*cacheItem, 0, maxCleanCount)
	defer func() {
		for _, item := range inFlightItems {
			heap.Push(c.heap, item)
		}
	}()
	cleanedCount := 0
	for scannedCount := 0; scannedCount < maxScanCount && cleanedCount < maxCleanCount; scannedCount++ {
		item := c.heap.Peek()
		if item == nil {
			return
		}

		if item.expireAt.Before(now) {
			heap.Pop(c.heap)
			data, ok := c.data[item.host]
			if !ok || data.expireAt != item.expireAt {
				cleanedCount++
				continue
			}
			if data.refreshing || data.stalePins > 0 {
				// An in-flight lookup still needs this revision to publish either the
				// refreshed addresses or the stale-address fallback. Keep the original
				// expiry so cancellation cannot make an expired entry fresh again.
				inFlightItems = append(inFlightItems, item)
				continue
			}
			cleanedCount++
			if data.keepAlive {
				// Keep the stale-address fallback available after ordinary expiry,
				// while retaining a heap node so capacity eviction can still find it.
				data.expireAt = now.Add(c.expiration)
				c.data[item.host] = data
				tracked := data
				heap.Push(c.heap, &tracked)
				continue
			}
			delete(c.data, item.host)
		} else {
			return
		}
	}
}

func (c *cache) put(key string, ipList []string, refreshInterval time.Duration) {
	now := time.Now()
	if _, ok := c.data[key]; ok {
		c.removeHeapItems(key)
	}
	stored := make([]string, len(ipList))
	copy(stored, ipList)
	item := cacheItem{
		ipList:    stored,
		expireAt:  now.Add(c.expiration),
		refreshAt: now.Add(refreshInterval),
		revision:  c.nextRevisionLocked(),
		host:      key,
	}
	c.data[key] = item
	heap.Push(c.heap, &item)
	c.evictOverflow()

	if time.Now().After(c.cleanTime) {
		c.cleanCache()
	}
}

func (c *cache) Put(key string, ipList []string, refreshInterval time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.put(key, ipList, refreshInterval)
}

func (c *cache) putIfRevision(key string, revision uint64, ipList []string, refreshInterval time.Duration) bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok || data.revision != revision {
		return false
	}
	c.put(key, ipList, refreshInterval)
	return true
}

func (c *cache) releaseRefreshIfRevision(key string, revision uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	data, ok := c.data[key]
	if !ok || data.revision != revision {
		return
	}
	data.refreshing = false
	c.data[key] = data
}

type resolver struct {
	cache           *cache
	closer          chan struct{}
	closeOnce       sync.Once
	ctx             context.Context
	cancel          context.CancelFunc
	lookupIP        func(context.Context, string) ([]net.IP, error)
	refreshInterval time.Duration
	refreshTimeout  time.Duration
}

func (r *resolver) Close() {
	r.closeOnce.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
		if r.closer != nil {
			close(r.closer)
		}
	})
}

func newResolver(expiration time.Duration) *resolver {
	pq := make(priorityQueue, 0)
	ctx, cancel := context.WithCancel(context.Background())
	cacheResolver := &resolver{cache: &cache{
		heap:       &pq,
		cleanTime:  time.Now().Add(expiration),
		data:       make(map[string]cacheItem),
		expiration: expiration,
	},
		closer:          make(chan struct{}),
		ctx:             ctx,
		cancel:          cancel,
		lookupIP:        LookupIP,
		refreshInterval: RefreshInterval,
		refreshTimeout:  dnsRefreshTimeout,
	}
	return cacheResolver
}

func ipToStringList(ips []net.IP) []string {
	res := make([]string, len(ips))
	for i, ip := range ips {
		res[i] = ip.String()
	}
	return res
}

func LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(addrs))
	for i, ia := range addrs {
		ips[i] = ia.IP
	}
	return ips, nil
}

func (r *resolver) lookup(ctx context.Context, host string) ([]string, error) {
	lookupIP := r.lookupIP
	if lookupIP == nil {
		lookupIP = LookupIP
	}
	ips, err := lookupIP(ctx, host)
	if err != nil {
		return nil, err
	}
	return ipToStringList(ips), nil
}

func (r *resolver) GetIpListWithoutCache(ctx context.Context, host string) ([]string, error) {
	ipsStr, err := r.lookup(ctx, host)
	if err != nil {
		return nil, err
	}
	r.cache.Put(host, ipsStr, r.refreshInterval)

	return ipsStr, nil
}

func (r *resolver) GetIpList(ctx context.Context, host string) ([]string, error) {
	now := time.Now()
	ipList, refreshDue, ok := r.cache.GetWithRefreshState(host, now)
	if ok {
		if refreshDue {
			if revision, claimed := r.cache.ClaimRefresh(host, now, r.refreshInterval); claimed {
				r.refresh(host, revision)
			}
		}
		return ipList, nil
	}

	stale, revision, hasStale := r.cache.getStaleAndPin(host)
	if hasStale {
		defer r.cache.releaseStalePinIfRevision(host, revision)
	}
	value, err := r.lookup(ctx, host)
	if err != nil {
		// A canceled request must retain its cancellation semantics and must not
		// turn the cached entry into a permanent stale fallback.
		if ctx.Err() != nil {
			return nil, err
		}
		if hasStale && r.cache.setKeepAliveIfRevision(host, revision, time.Now().Add(r.refreshInterval)) {
			return stale, nil
		}
		// Another lookup may have refreshed the cache while this lookup failed.
		if current, exists := r.cache.Get(host); exists {
			return current, nil
		}
		return nil, err
	}
	r.cache.Put(host, value, r.refreshInterval)
	return value, nil
}

func (r *resolver) handleRefreshFailure(refreshCtx context.Context, host string, revision uint64) {
	if refreshCtx.Err() != nil {
		r.cache.releaseRefreshIfRevision(host, revision)
		return
	}
	r.cache.setKeepAliveIfRevision(host, revision, time.Now().Add(r.refreshInterval))
}

// refresh performs a request-driven, bounded asynchronous refresh. The cache's
// atomic refresh claim guarantees that at most one refresh is started per host
// during each refresh interval.
func (r *resolver) refresh(host string, revision uint64) {
	refreshCtx := r.ctx
	if refreshCtx == nil {
		refreshCtx = context.Background()
	}
	select {
	case <-refreshCtx.Done():
		r.cache.releaseRefreshIfRevision(host, revision)
		return
	default:
	}

	go func() {
		defer func() {
			if recover() != nil {
				// A panic in an asynchronous refresh must not crash the caller's
				// process or leave this host permanently marked as refreshing.
				r.handleRefreshFailure(refreshCtx, host, revision)
			}
		}()

		timeout := r.refreshTimeout
		if timeout <= 0 {
			timeout = dnsRefreshTimeout
		}
		ctx, cancel := context.WithTimeout(refreshCtx, timeout)
		defer cancel()
		ipList, err := r.lookup(ctx, host)
		if err != nil {
			r.handleRefreshFailure(refreshCtx, host, revision)
			return
		}
		r.cache.putIfRevision(host, revision, ipList, r.refreshInterval)
	}()
}

func (r *resolver) Remove(host string, ip string) {
	r.cache.Remove(host, ip)
}
