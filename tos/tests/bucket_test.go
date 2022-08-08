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

func TestHeadNonExistentBucket(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = "non-existent-bucket"
		client = env.prepareClient("")
	)
	head, err := client.HeadBucket(context.Background(), &tos.HeadBucketInput{Bucket: bucket})
	require.NotNil(t, err)
	require.Nil(t, head)
	terr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.True(t, strings.Contains(terr.Message, "unexpected"))
}

func TestOnlyBucketNameV2(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("only-bucket-name")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	checkBucketMeta(t, client, bucket, &tos.HeadBucketOutput{
		StorageClass: enum.StorageClassStandard,
		Region:       env.region,
	})
}

func TestAllParamsV2(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("create-bucket-all-params")
		client = env.prepareClient("")
	)
	created, err := client.CreateBucketV2(context.Background(), &tos.CreateBucketV2Input{
		Bucket:       bucket,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassStandard,
		AzRedundancy: enum.AzRedundancySingleAz,
	})
	checkSuccess(t, created, err, 200)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	checkBucketMeta(t, client, bucket, &tos.HeadBucketOutput{
		StorageClass: enum.StorageClassStandard,
		Region:       env.region,
	})
}

func TestInvalidBucketNameV2(t *testing.T) {
	testInvalidBucketName(t, "Bucket")
	testInvalidBucketName(t, "Bucket")
	testInvalidBucketName(t, "-bucket")
	testInvalidBucketName(t, "bucket-")
	testInvalidBucketName(t, randomString(64))
	testInvalidBucketName(t, "bu")
	testInvalidBucketName(t, "+bucket")
}

func TestDeleteNoneExistBucketV2(t *testing.T) {
	env := newTestEnv(t)
	bucket := "this-is-a-none-exist-bucket"
	client := env.prepareClient("")
	del, err := client.DeleteBucket(context.Background(), &tos.DeleteBucketInput{Bucket: bucket})
	checkFail(t, del, err, 404)
}

// covered by other tests
func TestDeleteBucketV2(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("delete-bucket")
		client = env.prepareClient(bucket)
	)
	del, err := client.DeleteBucket(context.Background(), &tos.DeleteBucketInput{Bucket: bucket})
	checkSuccess(t, del, err, 204)

}

func TestListBucketV2(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("list-bucket")
		client = env.prepareClient("", tos.WithSocketTimeout(360*time.Second, 360*time.Second), tos.WithRequestTimeout(360*time.Second))
	)

	listed, err := client.ListBuckets(context.Background(), &tos.ListBucketsInput{})
	checkSuccess(t, listed, err, 200)
	for _, bkt := range listed.Buckets {
		if strings.HasPrefix(bkt.Name, testPrefix) {
			cleanBucket(t, client, bkt.Name)
		}
	}
	created, err := client.CreateBucketV2(context.Background(), &tos.CreateBucketV2Input{
		Bucket:       bucket,
		ACL:          enum.ACLPrivate,
		StorageClass: enum.StorageClassStandard,
		AzRedundancy: enum.AzRedundancySingleAz,
	})
	checkSuccess(t, created, err, 200)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	listed, err = client.ListBuckets(context.Background(), &tos.ListBucketsInput{})
	var testBucket []tos.ListedBucket
	for _, bkt := range listed.Buckets {
		if strings.HasPrefix(bkt.Name, testPrefix) {
			testBucket = append(testBucket, bkt)
		}
	}
	require.Nil(t, err)
	require.Equal(t, 200, listed.StatusCode)
	require.Equal(t, 1, len(testBucket))
	require.Equal(t, bucket, testBucket[0].Name)
}
