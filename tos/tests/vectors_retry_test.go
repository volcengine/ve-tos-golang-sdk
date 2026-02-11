package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestVectorsRetry(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = "retry-429" + randomString(6)
		ctx    = context.Background()
	)
	rt := newRetryInjectRT(429, 3, 1)
	client := env.prepareVectorsClient(bucket, tos.WithVectorsHTTPTransport(rt))
	client.CreateVectorBucket(ctx, &tos.CreateVectorBucketInput{
		VectorBucketName: bucket,
	})
	require.Equal(t, len(rt.counts), 1)

	var onlyKey string
	var onlyVal int
	for k, v := range rt.counts {
		onlyKey, onlyVal = k, v
		break
	}

	t.Log("key:", onlyKey, "val:", onlyVal)
	require.Equal(t, onlyVal, 3)

	out, err := client.GetVectorBucket(ctx, &tos.GetVectorBucketInput{
		VectorBucketName: bucket,
		AccountID:        env.accountId,
	})
	require.Nil(t, err)
	require.Equal(t, out.VectorBucket.VectorBucketName, bucket)
	client.DeleteVectorBucket(ctx, &tos.DeleteVectorBucketInput{
		VectorBucketName: bucket,
		AccountID:        env.accountId,
	})
	require.Equal(t, len(rt.counts), 3)
	for k, v := range rt.counts {
		t.Log("key:", k, "val:", v)
		require.Equal(t, v, 3)
	}
}
