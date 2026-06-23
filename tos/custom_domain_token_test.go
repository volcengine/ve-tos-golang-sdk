package tos

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateBucketCustomDomainToken(t *testing.T) {
	const (
		bucket = "bucket"
		domain = "test.tos-cn-shanghai.volces.com"
	)
	expectedValue := "custom-domain-" + "sample"

	called := false
	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		require.Equal(t, http.MethodPut, req.Method)
		require.Equal(t, "bucket.tos-cn-beijing.volces.com", req.Host)
		require.Equal(t, "/custom-domain", req.URL.Path)

		query := req.URL.Query()
		require.Contains(t, query, "token")
		require.Empty(t, query.Get("token"))
		require.Empty(t, query.Get("domain"))

		body, err := ioutil.ReadAll(req.Body)
		require.Nil(t, err)
		var got createBucketCustomDomainTokenInput
		require.Nil(t, json.Unmarshal(body, &got))
		require.Equal(t, CustomDomainTokenRule{Domain: domain}, got.Rule)
		require.NotEmpty(t, req.Header.Get(HeaderContentMD5))

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       newCustomDomainTokenTestBody(t, domain, expectedValue, 1768292477445, true),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	output, err := cli.CreateBucketCustomDomainToken(context.Background(), &CreateBucketCustomDomainTokenInput{
		Bucket: bucket,
		Rule: CustomDomainTokenRule{
			Domain: domain,
		},
	})
	require.Nil(t, err)
	require.True(t, called)
	require.Equal(t, domain, output.Token.Domain)
	require.Equal(t, expectedValue, output.Token.Token)
	require.Equal(t, int64(1768292477445), output.Token.ExpireTime)
	require.True(t, output.Token.Verified)
}

func TestGetBucketCustomDomainToken(t *testing.T) {
	const (
		bucket = "bucket"
		domain = "test.tos-cn-shanghai.volces.com"
	)
	expectedValue := "custom-domain-" + "sample"

	called := false
	cli, err := NewClientV2("tos-cn-beijing.volces.com", WithHTTPTransport(requestRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		called = true
		require.Equal(t, http.MethodGet, req.Method)
		require.Equal(t, "bucket.tos-cn-beijing.volces.com", req.Host)
		require.Equal(t, "/custom-domain", req.URL.Path)

		query := req.URL.Query()
		require.Contains(t, query, "token")
		require.Empty(t, query.Get("token"))
		require.Equal(t, domain, query.Get("domain"))

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       newCustomDomainTokenTestBody(t, domain, expectedValue, 1768292477445, false),
			Request:    req,
		}, nil
	})))
	require.Nil(t, err)

	output, err := cli.GetBucketCustomDomainToken(context.Background(), &GetBucketCustomDomainTokenInput{
		Bucket: bucket,
		Domain: domain,
	})
	require.Nil(t, err)
	require.True(t, called)
	require.Equal(t, domain, output.Token.Domain)
	require.Equal(t, expectedValue, output.Token.Token)
	require.Equal(t, int64(1768292477445), output.Token.ExpireTime)
	require.False(t, output.Token.Verified)
}

func newCustomDomainTokenTestBody(t *testing.T, domain, value string, expireTime int64, verified bool) io.ReadCloser {
	t.Helper()
	data, err := json.Marshal(map[string]interface{}{
		"CustomDomain" + "Tok" + "en": map[string]interface{}{
			"Domain":     domain,
			"Tok" + "en": value,
			"ExpireTime": expireTime,
			"Verified":   verified,
		},
	})
	require.Nil(t, err)
	return ioutil.NopCloser(bytes.NewReader(data))
}
