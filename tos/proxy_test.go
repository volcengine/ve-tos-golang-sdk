package tos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewProxy(t *testing.T) {
	host := "a.b.c"
	proxy, err := NewProxy(host, 8080)
	require.Nil(t, err)
	require.Equal(t, proxy.proxyHost, "http://"+host)

	host = "http://a.b.c"
	proxy, err = NewProxy(host, 8080)
	require.Nil(t, err)
	require.Equal(t, proxy.proxyHost, host)

	host = "HTTP://a.b.c"
	proxy, err = NewProxy(host, 8080)
	require.Nil(t, err)
	require.Equal(t, proxy.proxyHost, host)

	host = "https://a.b.c"
	proxy, err = NewProxy(host, 8080)
	require.NotNil(t, err)
	require.Equal(t, err, ProxyNotSupportHttps)

	host = "HTTPS://a.b.c"
	proxy, err = NewProxy(host, 8080)
	require.NotNil(t, err)
	require.Equal(t, err, ProxyNotSupportHttps)
}
