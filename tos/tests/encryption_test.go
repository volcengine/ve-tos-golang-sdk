package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"testing"
)

func TestEncryption(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("encryption")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer cleanBucket(t, client, bucket)
	_, err := client.PutBucketEncryption(ctx, &tos.PutBucketEncryptionInput{
		Bucket: bucket,
		Rule: tos.BucketEncryptionRule{
			ApplyServerSideEncryptionByDefault: tos.ApplyServerSideEncryptionByDefault{
				SSEAlgorithm: "AES256",
			},
		},
	})
	require.NoError(t, err)

	getResp, err := client.GetBucketEncryption(ctx, &tos.GetBucketEncryptionInput{
		Bucket: bucket,
	})
	require.NoError(t, err)
	require.Equal(t, getResp.Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm, "AES256")

	infoResp, err := client.GetBucketInfo(ctx, &tos.GetBucketInfoInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, infoResp)
	require.Equal(t, infoResp.Bucket.ServerSideEncryptionConfiguration.Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm, "AES256")

	_, err = client.DeleteBucketEncryption(ctx, &tos.DeleteBucketEncryptionInput{Bucket: bucket})
	require.NoError(t, err)

	_, err = client.GetBucketEncryption(ctx, &tos.GetBucketEncryptionInput{Bucket: bucket})
	require.NotNil(t, err)
}
