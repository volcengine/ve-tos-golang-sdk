package tos

import (
	"bytes"
	"context"
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

type requestRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f requestRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

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

func TestPutObjectV2EmptyBodySkipsTosChunked(t *testing.T) {
	for _, test := range []struct {
		name    string
		options []ClientOption
	}{
		{name: "default"},
		{name: "trailer enabled", options: []ClientOption{WithDisableTrailerHeader(false)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			options := append([]ClientOption{}, test.options...)
			options = append(options, WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				require.Equal(t, int64(0), req.ContentLength)
				require.Empty(t, req.TransferEncoding)
				require.Empty(t, req.Header.Get(HeaderContentEncoding))
				require.NotContains(t, req.Header.Get(HeaderContentEncoding), tosChunkedContentEncodingHeaderValue)
				require.Empty(t, req.Header.Get(tosTrailerHeaderName))
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       ioutil.NopCloser(bytes.NewReader(nil)),
					Request:    req,
				}, nil
			})))

			cli, err := NewClientV2("tos-cn-beijing.volces.com", options...)
			require.Nil(t, err)

			_, err = cli.PutObjectV2(context.Background(), &PutObjectV2Input{
				PutObjectBasicInput: PutObjectBasicInput{
					Bucket: "bucket",
					Key:    "key",
				},
				Content: bytes.NewReader(nil),
			})
			require.Nil(t, err)
		})
	}
}

func TestPutObjectFromFileLimitsFileGrowth(t *testing.T) {
	original := []byte("0123456789-original-file-content")
	extra := []byte("-appended-after-request-build")
	file := mustCreateTempFile(t, original)
	defer os.Remove(file)

	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		appendToFile(t, file, extra)
		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		require.Equal(t, int64(len(original)), req.ContentLength)
		require.Equal(t, original, body)
		requireCurrentFileContent(t, file, append(append([]byte{}, original...), extra...))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	_, err = cli.PutObjectFromFile(context.Background(), &PutObjectFromFileInput{
		PutObjectBasicInput: PutObjectBasicInput{
			Bucket: "bucket",
			Key:    "key",
		},
		FilePath: file,
	})
	require.Nil(t, err)
}

func TestUploadPartFromFileLimitsFileGrowth(t *testing.T) {
	original := []byte("prefix-expected-part-suffix")
	offset := uint64(len("prefix-"))
	expected := []byte("expected-part")
	extra := []byte("-appended-after-request-build")
	file := mustCreateTempFile(t, original)
	defer os.Remove(file)

	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		appendToFile(t, file, extra)
		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		require.Equal(t, int64(len(expected)), req.ContentLength)
		require.Equal(t, expected, body)
		requireCurrentFileContent(t, file, append(append([]byte{}, original...), extra...))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	_, err = cli.UploadPartFromFile(context.Background(), &UploadPartFromFileInput{
		UploadPartBasicInput: UploadPartBasicInput{
			Bucket:     "bucket",
			Key:        "key",
			UploadID:   "upload-id",
			PartNumber: 1,
		},
		FilePath: file,
		Offset:   offset,
		PartSize: int64(len(expected)),
	})
	require.Nil(t, err)
}

func TestUploadPartFromFileWithZeroPartSizeLimitsResolvedFileGrowth(t *testing.T) {
	original := []byte("head-rest-of-file")
	offset := uint64(len("head-"))
	expected := []byte("rest-of-file")
	extra := []byte("-appended-after-request-build")
	file := mustCreateTempFile(t, original)
	defer os.Remove(file)

	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		appendToFile(t, file, extra)
		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		require.Equal(t, int64(len(expected)), req.ContentLength)
		require.Equal(t, expected, body)
		requireCurrentFileContent(t, file, append(append([]byte{}, original...), extra...))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	_, err = cli.UploadPartFromFile(context.Background(), &UploadPartFromFileInput{
		UploadPartBasicInput: UploadPartBasicInput{
			Bucket:     "bucket",
			Key:        "key",
			UploadID:   "upload-id",
			PartNumber: 1,
		},
		FilePath: file,
		Offset:   offset,
	})
	require.Nil(t, err)
}

func TestUploadPartV2UsesResolvedContentLength(t *testing.T) {
	content := []byte("part content")
	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		require.Equal(t, int64(len(content)), req.ContentLength)
		require.Equal(t, content, body)
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	_, err = cli.UploadPartV2(context.Background(), &UploadPartV2Input{
		UploadPartBasicInput: UploadPartBasicInput{
			Bucket:     "bucket",
			Key:        "key",
			UploadID:   "upload-id",
			PartNumber: 1,
		},
		Content: bytes.NewReader(content),
	})
	require.Nil(t, err)
}

func TestUploadPartV2LimitsFileGrowth(t *testing.T) {
	original := []byte("upload-part-v2-file-content")
	extra := []byte("-appended-after-request-build")
	fileName := mustCreateTempFile(t, original)
	defer os.Remove(fileName)
	file, err := os.Open(fileName)
	require.Nil(t, err)
	defer file.Close()

	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		appendToFile(t, fileName, extra)
		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		require.Equal(t, int64(len(original)), req.ContentLength)
		require.Equal(t, original, body)
		requireCurrentFileContent(t, fileName, append(append([]byte{}, original...), extra...))
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	_, err = cli.UploadPartV2(context.Background(), &UploadPartV2Input{
		UploadPartBasicInput: UploadPartBasicInput{
			Bucket:     "bucket",
			Key:        "key",
			UploadID:   "upload-id",
			PartNumber: 1,
		},
		Content: file,
	})
	require.Nil(t, err)
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

func mustCreateTempFile(t *testing.T, content []byte) string {
	t.Helper()
	file, err := ioutil.TempFile("", "tos-sdk-test-*")
	require.Nil(t, err)
	_, err = file.Write(content)
	require.Nil(t, err)
	require.Nil(t, file.Close())
	return file.Name()
}

func appendToFile(t *testing.T, fileName string, content []byte) {
	t.Helper()
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_APPEND, 0)
	require.Nil(t, err)
	_, err = file.Write(content)
	require.Nil(t, err)
	require.Nil(t, file.Close())
}

func requireCurrentFileContent(t *testing.T, fileName string, expected []byte) {
	t.Helper()
	actual, err := ioutil.ReadFile(fileName)
	require.Nil(t, err)
	require.Equal(t, expected, actual)
}
