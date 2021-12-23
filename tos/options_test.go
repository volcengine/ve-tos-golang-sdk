package tos

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	options := []Option{WithContentType("type")}
	builder := requestBuilder{
		Query:  make(url.Values),
		Header: make(http.Header),
	}

	for _, option := range options {
		option(&builder)
	}

	require.Equal(t, "type", builder.Header.Get(HeaderContentType))
}
