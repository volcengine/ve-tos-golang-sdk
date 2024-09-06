package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

func TestSoftSymlink(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("symlink")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := int64(6)
	key := "raw-" + randomString(int(length))
	data := randomString(int(length))
	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(data),
	})
	require.NoError(t, err)
	softKey := "软连接-" + randomString(6)
	acl := enum.ACLPublicRead
	storageClass := enum.StorageClassIa
	tagging := "key1=value1&key2=value2&key3"
	ct := "text/plain; charset=utf-8"
	cacheControl := "public, max-age=86400"
	contentDisposition := "attachment; filename=\"" + softKey + "\""
	contentLanguage := "en"
	expires := time.Now()
	metaKey := "aaaa"
	meta := map[string]string{metaKey: "bbb"}
	_, err = client.PutSymlink(ctx, &tos.PutSymlinkInput{
		Bucket:              bucket,
		Key:                 softKey,
		SymlinkTargetKey:    key,
		SymlinkTargetBucket: bucket,
		ACL:                 acl,
		Meta:                meta,
		StorageClass:        storageClass,
		Tagging:             tagging,
		ContentType:         ct,
		Expires:             expires,
		CacheControl:        cacheControl,
		ContentDisposition:  contentDisposition,
		ContentLanguage:     contentLanguage,
	})
	require.NoError(t, err)

	out, err := client.GetSymlink(context.Background(), &tos.GetSymlinkInput{
		Bucket: bucket,
		Key:    softKey,
	})
	require.NoError(t, err)
	require.Equal(t, key, out.SymlinkTargetKey)
	require.Equal(t, bucket, out.SymlinkTargetBucket)

	headOut, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    softKey,
	})
	require.NoError(t, err)

	require.Equal(t, contentDisposition, headOut.ContentDisposition)
	require.Equal(t, contentLanguage, headOut.ContentLanguage)
	require.Equal(t, ct, headOut.ContentType)
	require.Equal(t, cacheControl, headOut.CacheControl)
	require.Equal(t, expires.Second(), headOut.Expires.Second())
	require.Equal(t, length, headOut.SymlinkTargetSize)
	require.Equal(t, key, out.SymlinkTargetKey)
	require.Equal(t, bucket, out.SymlinkTargetBucket)
	v, _ := headOut.Meta.Get(metaKey)
	require.Equal(t, v, meta[metaKey])
	require.Equal(t, headOut.SymlinkTargetSize, length)
	require.Equal(t, int64(3), headOut.TaggingCount)

	getResp, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: softKey})
	require.NoError(t, err)
	softData, err := ioutil.ReadAll(getResp.Content)
	require.NoError(t, err)
	require.Equal(t, data, string(softData))

	taggingResp, err := client.GetObjectTagging(ctx, &tos.GetObjectTaggingInput{Bucket: bucket, Key: softKey})
	require.NoError(t, err)
	require.Equal(t, len(taggingResp.TagSet.Tags), 3)
	require.Equal(t, taggingResp.TagSet.Tags[0].Key, "key1")
	require.Equal(t, taggingResp.TagSet.Tags[0].Value, "value1")
	require.Equal(t, taggingResp.TagSet.Tags[1].Key, "key2")
	require.Equal(t, taggingResp.TagSet.Tags[1].Value, "value2")
	require.Equal(t, taggingResp.TagSet.Tags[2].Key, "key3")
	require.Equal(t, taggingResp.TagSet.Tags[2].Value, "")

	listResp, err := client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{Bucket: bucket})
	require.NoError(t, err)
	for _, object := range listResp.Contents {
		if object.Key == softKey {
			require.Equal(t, object.ObjectType, "Symlink")
		}
	}

	listType2, err := client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{Bucket: bucket})
	require.NoError(t, err)
	for _, object := range listType2.Contents {
		if object.Key == softKey {
			require.Equal(t, object.ObjectType, "Symlink")
		}
	}
	listVersion, err := client.ListObjectVersionsV2(ctx, &tos.ListObjectVersionsV2Input{Bucket: bucket})
	require.NoError(t, err)
	for _, object := range listVersion.Versions {
		if object.Key == softKey {
			require.Equal(t, object.ObjectType, "Symlink")
		}
	}
}
