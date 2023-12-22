package tos

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	r := newResolver(time.Minute)
	endPoint := os.Getenv("TOS_GO_SDK_ENDPOINT")
	ipList, err := r.GetIpList(context.Background(), endPoint)
	require.Nil(t, err)
	require.True(t, len(r.cache.data) == 1)

	for _, ip := range ipList {
		r.Remove(endPoint, ip)
	}

	require.True(t, len(r.cache.data) == 0)

}

func TestResolverConcurrency(t *testing.T) {
	r := newResolver(time.Minute)
	endPoint := os.Getenv("TOS_GO_SDK_ENDPOINT")
	var wg sync.WaitGroup
	total := 20
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			ipList, err := r.GetIpList(context.Background(), endPoint)
			require.Nil(t, err)
			require.True(t, len(ipList) > 0)
			r.Remove(endPoint, ipList[0])
			ipList, err = r.GetIpList(context.Background(), endPoint)
			require.Nil(t, err)
			require.True(t, len(ipList) > 0)
		}()
	}
	wg.Wait()
}

func TestFreshKey(t *testing.T) {
	pq := make(priorityQueue, 0)
	ctx, cancel := context.WithCancel(context.Background())
	expire := time.Second
	r := &resolver{cache: &cache{
		heap:       &pq,
		cleanTime:  time.Now().Add(expire),
		data:       make(map[string]cacheItem),
		expiration: expire,
	}, ctx: ctx, closer: make(chan struct{})}
	endPoint := os.Getenv("TOS_GO_SDK_ENDPOINT")
	fmt.Println(r)
	ipList, err := r.GetIpList(context.Background(), endPoint)
	require.Nil(t, err)
	require.True(t, len(ipList) > 0)
	data, _ := r.cache.data[endPoint]
	require.False(t, data.keepAlive)

	// 刷新成功
	RefreshInterval = time.Millisecond * 200
	r.refresh()
	time.Sleep(RefreshInterval * 2)
	refreshData, _ := r.cache.data[endPoint]
	require.True(t, refreshData.expireAt.After(data.expireAt))
	require.False(t, refreshData.keepAlive)

	// 刷新失败
	cancel()
	time.Sleep(RefreshInterval * 2)
	refreshData, _ = r.cache.data[endPoint]
	require.True(t, refreshData.keepAlive)

	// 过期后可以从缓存中获取到
	time.Sleep(expire)
	ip, exist := r.cache.Get(endPoint)
	require.True(t, exist)
	require.True(t, len(ip) > 0)

	// 恢复后可以正常刷新
	r.ctx = context.Background()
	time.Sleep(RefreshInterval * 2)
	refreshData, _ = r.cache.data[endPoint]
	require.False(t, refreshData.keepAlive)
	require.True(t, refreshData.expireAt.After(time.Now()))

	r.Close()

}
