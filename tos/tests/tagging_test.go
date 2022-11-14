package tests

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketTagging(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("mirror-back")
		key    = "key123"
		client = env.prepareClient(bucket)
		value  = randomString(1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: bytes.NewBufferString(value),
	})
	require.Nil(t, err)
	putRes, err := client.PutObjectTagging(ctx, &tos.PutObjectTaggingInput{
		Bucket: bucket,
		Key:    key,
		TagSet: tos.TagSet{Tags: []tos.Tag{{
			Key:   "1",
			Value: "2",
		},
		}},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 1)

	deleteRes, err := client.DeleteObjectTagging(ctx, &tos.DeleteObjectTaggingInput{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 0)

}

func TestBucketTaggingVersion(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("mirror-back")
		key    = "key123"
		client = env.prepareClient(bucket)
		value  = randomString(1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	enableMultiVersion(t, client, bucket)
	time.Sleep(time.Minute)
	ctx := context.Background()
	putObjectRes, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: bytes.NewBufferString(value),
	})
	require.Nil(t, err)

	version := putObjectRes.VersionID
	putObjectRes, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: bytes.NewBufferString(value),
	})
	putRes, err := client.PutObjectTagging(ctx, &tos.PutObjectTaggingInput{
		Bucket:    bucket,
		Key:       key,
		VersionID: version,
		TagSet: tos.TagSet{Tags: []tos.Tag{{
			Key:   "1",
			Value: "2",
		},
		}},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)
	getRes, err := client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 0)

	getRes, err = client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{Bucket: bucket, Key: key, VersionID: version})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 1)

	deleteRes, err := client.DeleteObjectTagging(ctx, &tos.DeleteObjectTaggingInput{Bucket: bucket, Key: key, VersionID: version})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{Bucket: bucket, Key: key, VersionID: version})
	require.Nil(t, err)
	require.True(t, len(getRes.TagSet.Tags) == 0)

}
