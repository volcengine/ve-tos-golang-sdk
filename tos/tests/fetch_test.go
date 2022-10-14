package tests

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err)
	fetchRes, err := client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:       bucket,
		Key:          fetchKey,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassIa,
		Meta:         map[string]string{"test-key": "test-value"},
		URL:          "https://" + bucket + "." + env.endpoint + "/" + key,
	})
	assert.Nil(t, err)

	headRes, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    fetchKey,
	})
	assert.Nil(t, err)
	actualValue, _ := headRes.Meta.Get("test-key")
	assert.Equal(t, actualValue, "test-value")
	assert.Equal(t, headRes.StorageClass, enum.StorageClassIa)
	assert.Equal(t, headRes.ETag, fetchRes.Etag)
	assert.Equal(t, headRes.ContentLength, int64(length))
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
	assert.Nil(t, err)
	_, err = client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:       bucket,
		Key:          fetchKey,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassIa,
		Meta:         map[string]string{"test-key": "test-value"},
		URL:          "https://" + bucket + "." + env.endpoint + "/" + key,
	})
	assert.Nil(t, err)

	_, err = client.FetchObjectV2(ctx, &tos.FetchObjectInputV2{
		Bucket:        bucket,
		Key:           key,
		ACL:           enum.ACLPrivate,
		StorageClass:  enum.StorageClassIa,
		Meta:          map[string]string{"test-key": "test-value"},
		URL:           "https://" + bucket + "." + env.endpoint + "/" + key,
		IgnoreSameKey: true,
	})
	assert.Nil(t, err)

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
	assert.Nil(t, err)
	res, err := client.PutFetchTaskV2(ctx, &tos.PutFetchTaskInputV2{
		Bucket: bucket,
		Key:    fetchKey,
		ACL:    enum.ACLPrivate,
		Meta:   map[string]string{"test-key": "test-value"},
		URL:    "https://" + bucket + "." + env.endpoint + "/" + key})
	fmt.Println(res.TaskID)
	time.Sleep(time.Second * 5)

	headRes, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    fetchKey,
	})
	assert.Nil(t, err)
	actualValue, _ := headRes.Meta.Get("test-key")
	assert.Equal(t, actualValue, "test-value")
	assert.Equal(t, headRes.ContentLength, int64(length))

}
