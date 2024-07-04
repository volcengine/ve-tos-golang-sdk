package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/codes"
)

func TestBucketCORS(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("cors")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	maxAgeSeconds := 3600
	rule1 := tos.CorsRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{http.MethodGet, http.MethodDelete, http.MethodPut},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"X-TOS-HEADER-1", "X-TOS-HEADER-2"},
		MaxAgeSeconds: maxAgeSeconds,
		ResponseVary:  true,
	}
	rule2 := tos.CorsRule{
		AllowedOrigin: []string{"https://www.volcengine.com"},
		AllowedMethod: []string{http.MethodPut, http.MethodPost},
		AllowedHeader: []string{"Authorization"},
		ExposeHeader:  []string{"X-TOS-HEADER-1", "X-TOS-HEADER-2"},
		MaxAgeSeconds: maxAgeSeconds,
		ResponseVary:  true,
	}
	putRes, err := client.PutBucketCORS(ctx, &tos.PutBucketCORSInput{
		Bucket:    bucket,
		CORSRules: []tos.CorsRule{rule1, rule2},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketCORS(context.Background(), &tos.GetBucketCORSInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, getRes)
	require.Equal(t, 2, len(getRes.CORSRules))
	require.Equal(t, 3, len(getRes.CORSRules[0].AllowedMethod))
	require.Equal(t, maxAgeSeconds, getRes.CORSRules[0].MaxAgeSeconds)
	require.Equal(t, true, getRes.CORSRules[0].ResponseVary)

	putRes, err = client.PutBucketCORS(ctx, &tos.PutBucketCORSInput{
		Bucket:    bucket,
		CORSRules: []tos.CorsRule{rule1},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err = client.GetBucketCORS(ctx, &tos.GetBucketCORSInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(getRes.CORSRules))

	deleteRes, err := client.DeleteBucketCORS(ctx, &tos.DeleteBucketCORSInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetBucketCORS(ctx, &tos.GetBucketCORSInput{
		Bucket: bucket,
	})
	require.Nil(t, getRes)
	require.NotNil(t, err)
	tosErr := err.(*tos.TosServerError)
	require.Equal(t, codes.NoSuchCORSConfiguration, tosErr.Code)
}
