package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketQuota(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("bucket-quota")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer client.Close()
	defer cleanBucket(t, client, bucket)

	const quota = int64(64 * 1024 * 1024)
	putOutput, err := client.PutBucketQuota(ctx, &tos.PutBucketQuotaInput{
		Bucket:       bucket,
		StorageQuota: quota,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, putOutput.StatusCode)
	require.NotEmpty(t, putOutput.RequestID)

	getOutput, err := client.GetBucketQuota(ctx, &tos.GetBucketQuotaInput{Bucket: bucket})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getOutput.StatusCode)
	require.NotEmpty(t, getOutput.RequestID)
	require.Equal(t, quota, getOutput.StorageQuota)

	restoreOutput, err := client.PutBucketQuota(ctx, &tos.PutBucketQuotaInput{
		Bucket:       bucket,
		StorageQuota: 0,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, restoreOutput.StatusCode)
	require.NotEmpty(t, restoreOutput.RequestID)
}
