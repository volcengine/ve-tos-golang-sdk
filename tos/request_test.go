package tos

import (
	"encoding/json"
	"fmt"
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

// func TestSetHeaders(t *testing.T) {
//
//	input := PutObjectV2Input{
//		PutObjectBasicInput: PutObjectBasicInput{
//			Bucket: "bucket",
//			Key:    "key",
//			CommonHeaders: CommonHeaders{
//				ContentLength: 123,
//				ContentMD5:    "test_md5",
//				ContentSHA256: "test_sha256",
//				CacheControl:  "test_cache",
//				Expires:       time.Now(),
//				ACL:           ACLType("test_acl"),
//				StorageClass:  StorageClassStandard,
//			},
//			SSEHeaders: SSEHeaders{
//				SSECAlgorithm: "test_sse_algorithm",
//			},
//			Meta:                 nil,
//			DataTransferListener: nil,
//			RateLimiter:          nil,
//		},
//		Content: nil,
//	}
//
// }

func TestEncodingContentDisposition(t *testing.T) {
	res := encodeContentDisposition("attachement; filename=\"中文.pdf\"")
	require.Equal(t, res, fmt.Sprintf("attachement; filename=\"%s\"", url.QueryEscape("中文.pdf")))

	res = encodeContentDisposition("attachment; filename=\"filename.pdf\"")
	require.Equal(t, res, "attachment; filename=\"filename.pdf\"")
	res = encodeContentDisposition("attachment; filename='中文.pdf'")
	require.Equal(t, res, fmt.Sprintf("attachment; filename='%s'", url.QueryEscape("中文.pdf")))

	res = encodeContentDisposition("attachment; filename*=UTF-8''%E6%96%87%E4%BB%B6%E5%90%8D%E5%AD%97.txt")
	require.Equal(t, res, "attachment; filename*=UTF-8''%E6%96%87%E4%BB%B6%E5%90%8D%E5%AD%97.txt")
}
