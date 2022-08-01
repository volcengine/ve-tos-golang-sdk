package tests

import (
	"context"
	"testing"

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
		Bucket:      bucket,
		Key:         key,
		ContentType: contentType,
		Meta:        meta,
	})
	head, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	require.Equal(t, contentType, head.Header.Get(tos.HeaderContentType))
	for k, v := range meta {
		require.Equal(t, v, head.Meta.Get(k))
	}
}
