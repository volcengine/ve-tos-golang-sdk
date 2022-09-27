package tos

import (
	"container/heap"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPriority(t *testing.T) {
	p := priorityQueue{}
	t3 := &cacheItem{expireAt: time.Now().Add(time.Second), host: "host3"}
	t1 := &cacheItem{expireAt: time.Now().Add(time.Second), host: "host1"}
	t2 := &cacheItem{expireAt: time.Now().Add(time.Second), host: "host2"}
	heap.Push(&p, t1)
	require.True(t, p.Peek() == t1)
	heap.Push(&p, t2)
	require.True(t, p.Peek() == t1)
	heap.Push(&p, t3)
	require.True(t, p.Peek() == t3)

	require.True(t, heap.Pop(&p) == t3)
	require.True(t, heap.Pop(&p) == t1)
	require.True(t, heap.Pop(&p) == t2)
}
