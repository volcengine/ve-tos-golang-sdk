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

func TestRestoreInfo(t *testing.T) {
	res := &Response{Header: map[string][]string{}}
	res.Header.Set(HeaderRestore, "ongoing-request=\"false\", expiry-date=\"Mon, 02 Dec 2024 00:00:00 GMT\"")
	restoreInfo := parseRestoreInfo(res)
	require.Equal(t, false, restoreInfo.RestoreStatus.OngoingRequest)
	require.Equal(t, true, restoreInfo.RestoreStatus.ExpiryDate.String() == "2024-12-02 00:00:00 +0000 UTC")
}
