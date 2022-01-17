package tos

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestURL(t *testing.T) {
	req := Request{
		Scheme: "https",
		Method: http.MethodGet,
		Host:   "localhost",
		Path:   "/abc/😊?/😭#~!.txt",
		Query: url.Values{
			"versionId": []string{"abc123"},
		},
	}

	u := req.URL()
	require.Equal(t, "https://localhost/abc/%F0%9F%98%8A%3F/%F0%9F%98%AD%23~%21.txt?versionId=abc123", u)
}

func TestNotMarshalInfo(t *testing.T) {
	output := PutObjectOutput{
		RequestInfo: RequestInfo{
			RequestID: "bbb",
			Header: http.Header{
				HeaderContentType: []string{"application/json"},
			},
		},
		ETag: "ccc",
	}
	data := `{
		"RequestInfo": {
			"RequestID": "aaa"
		},
		"RequestID": "ddd",
		"ETag": "abs",
		"VersionId": "vid"
	}
	`
	out, err := json.Marshal(&output)
	require.Nil(t, err)
	t.Logf("%s", out)
	err = json.Unmarshal([]byte(data), &output)
	require.Nil(t, err)
	require.Equal(t, output.ETag, "abs")
	require.Equal(t, output.RequestID, "bbb")
	require.Equal(t, "application/json", output.Header.Get(HeaderContentType))
}

func TestTryResolveLength(t *testing.T) {
	file, err := os.Open("./request.go")
	require.Nil(t, err)

	size := tryResolveLength(file)
	require.Greater(t, size, int64(0))

	buffers := net.Buffers{make([]byte, 1024), make([]byte, 1024)}
	size = tryResolveLength(&buffers)
	require.Equal(t, size, int64(2048))
}

func TestFileUnreadSize(t *testing.T) {
	file, err := os.Open("./request.go")
	require.Nil(t, err)

	stat, err := file.Stat()
	require.Nil(t, err)

	size, err := fileUnreadLength(file)
	require.Nil(t, err)
	require.Equal(t, size, stat.Size())

	n, err := file.Read(make([]byte, 8))
	require.Nil(t, err)
	require.Equal(t, n, 8)

	size, err = fileUnreadLength(file)
	require.Nil(t, err)
	require.Equal(t, size, stat.Size()-8)

	data, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, size, int64(len(data)))

	size, err = fileUnreadLength(file)
	require.Nil(t, err)
	require.Equal(t, size, int64(0))
}
