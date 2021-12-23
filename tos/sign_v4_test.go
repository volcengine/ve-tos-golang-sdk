package tos

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestURIEncode(t *testing.T) {
	out := URIEncode("23i23+___", true)
	require.Equal(t, "23i23%2B___", string(out))

	out = URIEncode("23i23 ___", true)
	require.Equal(t, "23i23%20___", string(out))

	out = URIEncode("23i23 /___", true)
	require.Equal(t, "23i23%20%2F___", string(out))

	out = URIEncode("23i23 /___", false)
	require.Equal(t, "23i23%20/___", string(out))
}

func TestAlgV4(t *testing.T) {
	date, err := time.Parse(iso8601Layout, "20210721T104454Z")
	require.Nil(t, err)

	cred := NewStaticCredentials("AKIAIOSFODNN7EXAMPLE", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")

	sv := SignV4{
		credentials:   cred,
		region:        "cn-north-1",
		signingHeader: defaultSigningHeaderV4,
		signingQuery:  defaultSigningQueryV4,
		now:           func() time.Time { return date.UTC() },
		signingKey:    SigningKey,
	}

	uri, err := url.Parse("https://test.tos.com:8080/test.txt")
	require.Nil(t, err)

	req := &Request{
		Method: http.MethodGet,
		Host:   uri.Host,
		Path:   uri.Path,
		Query:  uri.Query(),
		Header: make(http.Header),
	}

	query := sv.SignQuery(req, time.Hour)
	require.Nil(t, err)

	require.Equal(t, 6, len(query))
	require.Equal(t, "20210721T104454Z", query.Get(v4Date))
	require.Equal(t, "TOS4-HMAC-SHA256", query.Get(v4Algorithm))
	require.Equal(t, "host", query.Get(v4SignedHeaders))
	require.Equal(t, "AKIAIOSFODNN7EXAMPLE/20210721/cn-north-1/tos/request", query.Get(v4Credential))
	require.Equal(t, "3600", query.Get(v4Expires))
	require.Equal(t, "decc75e2b2d453117f81e53954eb2cd3a2f56db2951e9b2257863db4c4921111", query.Get(v4Signature))

	header := sv.SignHeader(req)
	require.Equal(t, 3, len(header))
	require.Equal(t, "TOS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/20210721/cn-north-1/tos/request,SignedHeaders=date;host;x-tos-date,Signature=8851df83fd33af6fb4a28addbe938d88139715dca85bbb43ee739c3eb58eddf2",
		header.Get(authorization))
	require.Equal(t, "20210721T104454Z", header.Get(v4Date))
	require.Equal(t, "20210721T104454Z", header.Get("Date"))
	require.Equal(t, "", header.Get(v4ContentSHA256))
}
