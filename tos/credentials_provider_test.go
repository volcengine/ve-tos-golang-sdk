package tos

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStaticCredentialsProvider_PreSign(t *testing.T) {
	endpoint := SupportedRegion()["cn-beijing"]
	cli, err := NewClientV2(endpoint,
		WithRegion("cn-beijing"),
		WithCredentialsProvider(NewStaticCredentialsProvider("AK_STATIC", "SK_STATIC", "")))
	require.Nil(t, err)

	out, err := cli.PreSignedURL(&PreSignedURLInput{
		HTTPMethod: http.MethodGet,
		Bucket:     "unit-bucket",
		Key:        "unit-key",
		Expires:    60,
	})
	require.Nil(t, err)
	require.NotEmpty(t, out.SignedUrl)

	u, err := url.Parse(out.SignedUrl)
	require.Nil(t, err)
	q := u.Query()
	cred := q.Get(v4Credential)
	require.True(t, strings.HasPrefix(cred, "AK_STATIC/"))
	require.True(t, strings.HasSuffix(cred, "/cn-beijing/tos/request"))

	// No security token expected
	require.Equal(t, "", q.Get(v4SecurityToken))
}
