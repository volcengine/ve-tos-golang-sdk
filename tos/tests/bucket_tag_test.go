package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"testing"
)

func TestBktTagging(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("bucket-tagging")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	putRes, err := client.PutBucketTagging(ctx, &tos.PutBucketTaggingInput{
		Bucket: bucket,
		TagSet: tos.TagSet{Tags: []tos.Tag{{
			Key:   "1",
			Value: "2",
		},
		}},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketTagging(ctx, &tos.GetBucketTaggingInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 1)

	deleteRes, err := client.DeleteBucketTagging(ctx, &tos.DeleteBucketTaggingInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetBucketTagging(ctx, &tos.GetBucketTaggingInput{Bucket: bucket})
	require.NotNil(t, err)

}
