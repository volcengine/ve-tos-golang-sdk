package tests

import (
	"bytes"
	"context"
	"fmt"
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

func TestGetBucketLocation(t *testing.T) {
	var (
		env = newTestEnv(t)
	)
	bucket := generateBucketName("bucket-location")
	client := env.prepareClient(bucket)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	res, err := client.GetBucketLocation(context.Background(), &tos.GetBucketLocationInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, res.Region, env.region)
	require.Equal(t, res.ExtranetEndpoint, env.endpoint)
}

func TestPutBucketStorageClass(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("storage-class")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	headRes, err := client.HeadBucket(ctx, &tos.HeadBucketInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, headRes.StorageClass != enum.StorageClassIa)

	output, err := client.PutBucketStorageClass(ctx, &tos.PutBucketStorageClassInput{
		Bucket:       bucket,
		StorageClass: enum.StorageClassIa,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	headRes, err = client.HeadBucket(ctx, &tos.HeadBucketInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, headRes.StorageClass, enum.StorageClassIa)

	output, err = client.PutBucketStorageClass(ctx, &tos.PutBucketStorageClassInput{
		Bucket:       bucket,
		StorageClass: enum.StorageClassArchiveFr,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	headRes, err = client.HeadBucket(ctx, &tos.HeadBucketInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, headRes.StorageClass, enum.StorageClassArchiveFr)
	
	output, err = client.PutBucketStorageClass(ctx, &tos.PutBucketStorageClassInput{
		Bucket:       bucket,
		StorageClass: enum.StorageClassColdArchive,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	headRes, err = client.HeadBucket(ctx, &tos.HeadBucketInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, headRes.StorageClass, enum.StorageClassColdArchive)

	output, err = client.PutBucketStorageClass(ctx, &tos.PutBucketStorageClassInput{
		Bucket:       bucket,
		StorageClass: "ci-test-storage",
	})
	require.NotNil(t, err)
}

func TestListObjectType2(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("list-object-type2")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
		value  = randomString(1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := 3
	for i := 0; i < length; i++ {
		for j := 0; j < length; j++ {
			for k := 0; k < length; k++ {
				_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
					PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: fmt.Sprintf("%d/%d/%d", i, j, k)},
					Content:             bytes.NewBufferString(value),
				})
				require.Nil(t, err)
			}
		}
	}

	out, err := client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{
		Bucket:     bucket,
		Prefix:     "0",
		StartAfter: "0/1",
		MaxKeys:    2,
	})
	require.Nil(t, err)
	require.Equal(t, len(out.Contents), 2)
	for _, obj := range out.Contents {
		require.True(t, strings.HasPrefix(obj.Key, "0/1"))
	}

	out, err = client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{
		Bucket:            bucket,
		Prefix:            "0",
		StartAfter:        "0/1",
		ContinuationToken: out.NextContinuationToken,
		MaxKeys:           2,
	})
	require.Nil(t, err)
	require.NotNil(t, out)
	require.Equal(t, len(out.Contents), 2)
	require.Equal(t, out.Contents[0].Owner.ID != "", true)
	require.Equal(t, out.Contents[0].Key, "0/1/2")
	require.Equal(t, out.Contents[1].Key, "0/2/0")
}

func TestListObjectType2MaxKeys(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("list-object-type2")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
		value  = randomString(1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	length := 5
	for i := 0; i < length; i++ {
		for j := 0; j < length; j++ {
			for k := 0; k < length; k++ {
				_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
					PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: fmt.Sprintf("%d/%d/%d", i, j, k)},
					Content:             bytes.NewBufferString(value),
				})
				require.Nil(t, err)
			}
		}
	}
	maxKeys := 100
	input := &tos.ListObjectsType2Input{
		Bucket:  bucket,
		MaxKeys: maxKeys,
	}
	out, err := client.ListObjectsType2(ctx, input)
	require.Nil(t, err)
	require.Equal(t, len(out.Contents), maxKeys)
	require.Equal(t, input.MaxKeys, maxKeys)
	out, err = client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{Bucket: bucket,
		MaxKeys:           maxKeys,
		ContinuationToken: out.NextContinuationToken})
	require.Nil(t, err)
	require.Equal(t, len(out.Contents), length*length*length-maxKeys)

	input = &tos.ListObjectsType2Input{Bucket: bucket}
	out, err = client.ListObjectsType2(ctx, input)
	require.Nil(t, err)
	require.Equal(t, len(out.Contents), length*length*length)
	require.Equal(t, input.MaxKeys, 0)
}

func TestEnableBucketVersion(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("bucket-version")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	enableMultiVersion(t, client, bucket)
	output, err := client.PutBucketVersioning(ctx, &tos.PutBucketVersioningInput{
		Bucket: bucket,
		Status: enum.VersioningStatusEnable,
	})
	require.Nil(t, err)
	t.Log(output.RequestID)

	getoutput, err := client.GetBucketVersioning(ctx, &tos.GetBucketVersioningInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	t.Log(getoutput.RequestID)
	require.Equal(t, getoutput.Status, enum.VersioningStatusEnable)

	output, err = client.PutBucketVersioning(ctx, &tos.PutBucketVersioningInput{
		Bucket: bucket,
		Status: enum.VersioningStatusSuspended,
	})
	require.Nil(t, err)
	t.Log(output.RequestID)

	getoutput, err = client.GetBucketVersioning(ctx, &tos.GetBucketVersioningInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	t.Log(getoutput.RequestID)
	require.Equal(t, getoutput.Status, enum.VersioningStatusSuspended)
}
