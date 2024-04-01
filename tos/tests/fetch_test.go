package tests

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestFetchObjectWithMeta(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("fetch-object-with-meta")
		client   = env.prepareClient(bucket)
		key      = "key123"
		fetchKey = randomString(6)
		ctx      = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := 1024
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLPublicRead},
		Content:             bytes.NewReader([]byte(randomString(length))),
	})
	require.Nil(t, err)
	fetchRes, err := client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:       bucket,
		Key:          fetchKey,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassIa,
		Meta:         map[string]string{"test-key": "test-value"},
		URL:          "https://" + bucket + "." + env.endpoint + "/" + key,
	})
	require.Nil(t, err)

	headRes, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    fetchKey,
	})
	require.Nil(t, err)
	actualValue, _ := headRes.Meta.Get("test-key")
	require.Equal(t, actualValue, "test-value")
	require.Equal(t, headRes.StorageClass, enum.StorageClassIa)
	require.Equal(t, headRes.ETag, fetchRes.Etag)
	require.Equal(t, headRes.ContentLength, int64(length))
}

func TestFetchObject(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("fetch-object")
		client   = env.prepareClient(bucket)
		key      = "key123"
		fetchKey = randomString(6)
		ctx      = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := 1024
	value := randomString(length)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLPublicRead},
		Content:             bytes.NewReader([]byte(value)),
	})
	require.Nil(t, err)
	_, err = client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:       bucket,
		Key:          fetchKey,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassIa,
		Meta:         map[string]string{"test-key": "test-value"},
		URL:          "https://" + bucket + "." + env.endpoint + "/" + key,
	})
	require.Nil(t, err)

	_, err = client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:        bucket,
		Key:           key,
		ACL:           enum.ACLPrivate,
		StorageClass:  enum.StorageClassIa,
		Meta:          map[string]string{"test-key": "test-value"},
		URL:           "https://" + bucket + "." + env.endpoint + "/" + key,
		IgnoreSameKey: true,
	})
	require.NotNil(t, err)

}

func TestPutFetchTaskV2(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("fetch-object-task")
		client   = env.prepareClient(bucket)
		key      = "key123"
		fetchKey = randomString(6)
		ctx      = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := 1024
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLPublicRead},
		Content:             bytes.NewReader([]byte(randomString(length))),
	})
	require.Nil(t, err)
	res, err := client.PutFetchTaskV2(ctx, &tos.PutFetchTaskInputV2{
		Bucket: bucket,
		Key:    fetchKey,
		ACL:    enum.ACLPrivate,
		Meta:   map[string]string{"test-key": "test-value"},
		URL:    "https://" + bucket + "." + env.endpoint + "/" + key})
	fmt.Println(res.TaskID)
	var headRes *tos.HeadObjectV2Output
	for i := 0; i < 20; i++ {

		headRes, err = client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    fetchKey,
		})
		if err != nil {
			time.Sleep(time.Second * 6)
			t.Log(err.Error())
			continue
		}
	}
	require.Nil(t, err)
	actualValue, _ := headRes.Meta.Get("test-key")
	require.Equal(t, actualValue, "test-value")
	require.Equal(t, headRes.ContentLength, int64(length))
}
