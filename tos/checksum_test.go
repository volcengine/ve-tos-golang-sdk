package tos

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestETagCheckReadCloser(t *testing.T) {
	buf := make([]byte, 1024)
	rand.Read(buf)
	hash := md5.Sum(buf)
	sum := hex.EncodeToString(hash[:])

	rc := NewETagCheckReadCloser(ioutil.NopCloser(bytes.NewReader(buf)), sum, "xxx")
	n, err := io.Copy(ioutil.Discard, rc)
	require.Nil(t, err)
	require.Equal(t, int(n), len(buf))

	rc = NewETagCheckReadCloser(ioutil.NopCloser(bytes.NewReader(buf)), "xxx", "xxx")
	n, err = io.Copy(ioutil.Discard, rc)
	require.NotNil(t, err)

	_, ok := err.(*ChecksumError)
	require.True(t, ok)
	require.Equal(t, int(n), len(buf))

	rc = NewETagCheckReadCloser(ioutil.NopCloser(bytes.NewReader(buf)), "", "xxx")
	n, err = io.Copy(ioutil.Discard, rc)
	require.Nil(t, err)
	require.Equal(t, int(n), len(buf))

	rc = NewETagCheckReadCloser(ioutil.NopCloser(bytes.NewReader(buf)), `"abc"`, "xxx")
	n, err = io.Copy(ioutil.Discard, rc)
	_, ok = err.(*ChecksumError)
	require.True(t, ok)
	require.Equal(t, int(n), len(buf))
	require.Equal(t, rc.eTag, "abc")
}
