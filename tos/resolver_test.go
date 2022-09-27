package tos

import (
	"context"
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
