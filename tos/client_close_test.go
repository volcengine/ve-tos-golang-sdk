package tos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientCloseClosesDefaultTransportResolver(t *testing.T) {
	client, err := NewClient("tos-cn-beijing.volces.com", WithDNSCacheTime(1))
	require.NoError(t, err)

	transport, ok := client.transport.(*DefaultTransport)
	require.True(t, ok)
	require.NotNil(t, transport.resolver)

	select {
	case <-transport.resolver.closer:
		t.Fatal("resolver was closed before Client.Close")
	default:
	}

	client.Close()
	select {
	case <-transport.resolver.closer:
	default:
		t.Fatal("resolver was not closed by Client.Close")
	}

	require.NotPanics(t, func() {
		client.Close()
	})
}

func TestClientCloseWithoutResolver(t *testing.T) {
	client, err := NewClient("tos-cn-beijing.volces.com", WithDNSCacheTime(0))
	require.NoError(t, err)

	transport, ok := client.transport.(*DefaultTransport)
	require.True(t, ok)
	require.Nil(t, transport.resolver)

	require.NotPanics(t, func() {
		client.Close()
	})
}
