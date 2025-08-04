package tests

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestSetObjectMetaV2(t *testing.T) {
	var (
		env         = newTestEnv(t)
		bucket      = generateBucketName("set-object-meta")
		client      = env.prepareClient(bucket)
		key         = "key"
		contentType = "application/x-www-form-urlencoded"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putRandomObject(t, client, bucket, key, 4*1024)
	meta := make(map[string]string)
	meta["Test-床前明月光"] = "疑是地上霜"
	meta["Test-Key"] = "Value"
	_, err := client.SetObjectMeta(context.Background(), &tos.SetObjectMetaInput{
		Bucket:       bucket,
		Key:          key,
		ContentType:  contentType,
		Meta:         meta,
		Expires:      time.Now().Add(24 * time.Hour),
		CacheControl: "no-cache",
	})
	head, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	require.Equal(t, contentType, head.Header.Get(tos.HeaderContentType))
	for k, v := range meta {
		val, ok := head.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.True(t, head.Header.Get("X-Tos-Expiration") == "")

	out, err := client.SetObjectExpires(context.Background(), &tos.SetObjectExpiresInput{Bucket: bucket, Key: key, ObjectExpires: 1})
	require.Nil(t, err)
	require.Equal(t, out.StatusCode, 200)
	head, err = client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.True(t, head.Header.Get("X-Tos-Expiration") != "")
}

func TestSetObjectMetaV2Version(t *testing.T) {
	var (
		env         = newTestEnv(t)
		bucket      = generateBucketName("set-object-meta-version")
		client      = env.prepareClient(bucket)
		key         = "key"
		contentType = "application/x-www-form-urlencoded"
		ctx         = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	enableMultiVersion(t, client, bucket)
	res, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(8)),
	})
	require.Nil(t, err)
	versionID := res.VersionID
	res, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(8)),
	})
	require.Nil(t, err)
	meta := make(map[string]string)
	meta["Test-Key"] = "Value"
	_, err = client.SetObjectMeta(context.Background(), &tos.SetObjectMetaInput{
		Bucket:       bucket,
		Key:          key,
		ContentType:  contentType,
		VersionID:    versionID,
		Meta:         meta,
		Expires:      time.Now().Add(24 * time.Hour),
		CacheControl: "no-cache",
	})
	head, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key, VersionID: versionID})
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	for k, v := range meta {
		val, ok := head.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}

	_, err = client.DeleteObjectV2(context.Background(), &tos.DeleteObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)

	resp, err := client.SetObjectExpires(context.Background(), &tos.SetObjectExpiresInput{Bucket: bucket, Key: key, ObjectExpires: 1, VersionID: versionID})
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)

	head, err = client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key, VersionID: versionID})
	require.Nil(t, err)
	require.True(t, head.Header.Get("X-Tos-Expiration") != "")
}

func TestMeta(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("meta-with")
		client = env.prepareClient(bucket)
		key    = "key"
		ctx    = context.Background()
		data   = randomString(12)
		meta   = "key"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	rawStr := "!@#$%^&*()_+-=[]{}|;':\"%2f%fg,     「」：-=+、\n./<>?中文测试编码%20%%%^&abcd /\\"
	out, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, Meta: map[string]string{meta: rawStr}},
		Content:             strings.NewReader(data),
	})
	require.Nil(t, err)
	require.Equal(t, out.StatusCode, http.StatusOK)

	head, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	r, ok := head.Meta.Get(meta)
	require.True(t, ok)
	require.Equal(t, r, rawStr)

	out, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, Meta: map[string]string{meta: url.PathEscape(rawStr)}},
		Content:             strings.NewReader(data),
	})
	require.Nil(t, err)
	require.Equal(t, out.StatusCode, http.StatusOK)

	head, err = client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	r, ok = head.Meta.Get(meta)
	require.True(t, ok)
	r, err = url.PathUnescape(r)
	require.Nil(t, err)
	require.Equal(t, r, rawStr)
}
