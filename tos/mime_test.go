package tos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMime(t *testing.T) {
	mm := ExtensionBasedContentTypeRecognizer{}
	typ := mm.ContentType("")
	require.Equal(t, "", typ)

	typ = mm.ContentType("a.json")
	require.Equal(t, "application/json", typ)

	me := EmptyContentTypeRecognizer{}
	typ = me.ContentType("")
	require.Equal(t, "", typ)

	typ = me.ContentType("a.json")
	require.Equal(t, "", typ)
}
