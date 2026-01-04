package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestHnsAppend(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("hns-append")
		client = env.prepareClient("")
		ctx    = context.Background()
	)

	_, err := client.CreateBucketV2(ctx, &tos.CreateBucketV2Input{
		Bucket:     bucket,
		BucketType: enum.BucketTypeHNS,
	})
	require.Nil(t, err)
	defer func() {
		cleanHNSBucket(t, client, bucket)
	}()
	key := "key1-" + randomString(8)
	body := randomString(1024 * 100)
	out, err := client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{
		Bucket:     bucket,
		Key:        key,
		Content:    strings.NewReader(body),
		ACL:        enum.ACLPublicRead,
		Meta:       map[string]string{"key": "value"},
		ContentMD5: md5s(body),
	})
	require.Nil(t, err)

	out, err = client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{
		Bucket:           bucket,
		Key:              key,
		Content:          strings.NewReader(body),
		PreHashCrc64ecma: out.HashCrc64ecma,
		Offset:           out.NextAppendOffset,
		ContentMD5:       md5s(body),
	})
	require.Nil(t, err)

	headOut, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, headOut.ContentLength, int64(len(body)*2))
	require.Equal(t, headOut.HashCrc64ecma, out.HashCrc64ecma)
	metaKey, _ := headOut.Meta.Get("key")
	require.Equal(t, metaKey, "value")
	require.True(t, headOut.LastModifyTimestamp.UnixNano() >= headOut.LastModified.UnixNano())

	key = "key2-" + randomString(6)
	// 已经有了 0字节的文件可以正常 append
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLPublicRead}, Content: strings.NewReader("")})
	require.Nil(t, err)

	out, err = client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{
		Bucket:     bucket,
		Key:        key,
		Content:    strings.NewReader(body),
		ContentMD5: md5s(body),
	})
	require.Nil(t, err)

	out, err = client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{
		Bucket:           bucket,
		Key:              key,
		Content:          strings.NewReader(body),
		PreHashCrc64ecma: out.HashCrc64ecma,
		Offset:           out.NextAppendOffset,
		ContentMD5:       md5s(body),
	})
	require.Nil(t, err)

	headOutKey2, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, headOutKey2.ContentLength, int64(len(body)*2))
	require.Equal(t, headOutKey2.HashCrc64ecma, headOut.HashCrc64ecma)

}

func TestHnsSetObjectTime(t *testing.T) {

	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("hns-set-object-time")
		client = env.prepareClient("")
		ctx    = context.Background()
	)
	_, err := client.CreateBucketV2(ctx, &tos.CreateBucketV2Input{
		Bucket:     bucket,
		BucketType: enum.BucketTypeHNS,
	})
	require.Nil(t, err)
	defer cleanHNSBucket(t, client, bucket)
	key := "key1-" + randomString(8)
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(1024 * 100)),
	})
	require.Nil(t, err)
	now := time.Now()
	_, err = client.SetObjectTime(ctx, &tos.SetObjectTimeInput{
		Bucket:          bucket,
		Key:             key,
		ModifyTimestamp: now,
	})
	require.Nil(t, err)
	resp, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, resp.LastModifyTimestamp.UnixNano(), now.UnixNano())
}
