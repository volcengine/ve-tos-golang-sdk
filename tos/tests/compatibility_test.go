package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func cleanBucketV1(t *testing.T, client *tos.Client, bucket string) {

	del, err := client.DeleteBucket(context.Background(), bucket)
	if err == nil {
		require.Equal(t, 204, del.StatusCode)
		return
	}
	if tos.StatusCode(err) == 404 {
		return
	}
	// the bucket is not clean
	if tos.StatusCode(err) == 409 {
		// delete all objects
		handle, err := client.Bucket(bucket)
		require.Nil(t, err)
		list, err := handle.ListObjects(context.Background(), &tos.ListObjectsInput{MaxKeys: 1000})
		require.Nil(t, err)
		for _, object := range list.Contents {
			del, err := handle.DeleteObject(context.Background(), object.Key)
			require.Nil(t, err)
			require.Equal(t, 204, del.StatusCode)
		}
		// delete all uncompleted MultipartUpload
		listMulti, err := handle.ListMultipartUploads(context.Background(), &tos.ListMultipartUploadsInput{})
		for _, upload := range listMulti.Upload {
			abort, err := handle.AbortMultipartUpload(context.Background(), &tos.AbortMultipartUploadInput{
				Bucket:   bucket,
				Key:      upload.Key,
				UploadID: upload.UploadId,
			})
			require.Nil(t, err)
			require.Equal(t, 204, abort.StatusCode)
		}
		require.Equal(t, 200, listMulti.StatusCode)
		// now, the bucket should be clean
		del, err = client.DeleteBucket(context.Background(), bucket)
		require.Nil(t, err)
		require.Equal(t, 204, del.StatusCode)
		return
	}
	// something wrong
	require.Equal(t, 200, tos.StatusCode(err))

}

func TestCurdV1(t *testing.T) {
	var (
		endpoint  = os.Getenv("TOS_GO_SDK_ENDPOINT")
		accessKey = os.Getenv("TOS_GO_SDK_AK")
		secretKey = os.Getenv("TOS_GO_SDK_SK")
		region    = os.Getenv("TOS_GO_SDK_REGION")
		bucket    = testPrefix + "curd" + "-" + randomString(16)
	)

	client, err := tos.NewClient(endpoint,
		tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)), tos.WithMaxRetryCount(5), tos.WithRegion(region))
	require.Nil(t, err)
	created, err := client.CreateBucket(context.Background(), &tos.CreateBucketInput{Bucket: bucket})
	checkSuccess(t, created, err, 200)
	defer cleanBucketV1(t, client, bucket)
	buckets, err := client.ListBuckets(context.Background(), &tos.ListBucketsInput{})
	checkSuccess(t, buckets, err, 200)
	head, err := client.HeadBucket(context.Background(), bucket)
	checkSuccess(t, head, err, 200)
	handle, err := client.Bucket(bucket)
	require.Nil(t, err)
	// basic
	objectKey := "test-key"
	content := randomString(4 * 1024)
	put, err := handle.PutObject(context.Background(), objectKey, strings.NewReader(content))
	checkSuccess(t, put, err, 200)
	url, err := client.PreSignedURL(http.MethodGet, bucket, "", time.Hour)
	require.Nil(t, err)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	require.Nil(t, err)
	data, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(data))
	headObject, err := handle.HeadObject(context.Background(), objectKey)
	checkSuccess(t, headObject, err, 200)
	get, err := handle.GetObject(context.Background(), objectKey)
	checkSuccess(t, get, err, 200)
	del, err := handle.DeleteObject(context.Background(), objectKey)
	checkSuccess(t, del, err, 204)
	headObject, err = handle.HeadObject(context.Background(), objectKey)
	checkFail(t, headObject, err, 404)
	// multipart
	multiKey := "test-multi-key"
	part1 := randomString(5 * 1024 * 1024)
	part2 := randomString(5 * 1024 * 1024)
	part3 := randomString(5 * 1024 * 1024)
	multi, err := handle.CreateMultipartUpload(context.Background(), multiKey)
	checkSuccess(t, multi, err, 200)
	upload1, err := handle.UploadPart(context.Background(), &tos.UploadPartInput{
		Key:        multiKey,
		UploadID:   multi.UploadID,
		PartNumber: 1,
		Content:    strings.NewReader(part1),
	})
	checkSuccess(t, upload1, err, 200)
	upload2, err := handle.UploadPart(context.Background(), &tos.UploadPartInput{
		Key:        multiKey,
		UploadID:   multi.UploadID,
		PartNumber: 2,
		Content:    strings.NewReader(part2),
	})
	checkSuccess(t, upload2, err, 200)
	upload3, err := handle.UploadPart(context.Background(), &tos.UploadPartInput{
		Key:        multiKey,
		UploadID:   multi.UploadID,
		PartNumber: 3,
		Content:    strings.NewReader(part3),
	})
	checkSuccess(t, upload3, err, 200)
	complete, err := handle.CompleteMultipartUpload(context.Background(), &tos.CompleteMultipartUploadInput{
		Key:           multiKey,
		UploadID:      multi.UploadID,
		UploadedParts: []tos.MultipartUploadedPart{upload1, upload2, upload3},
	})
	checkSuccess(t, complete, err, 200)
	// copy
	srcKey := "test-src-key"
	dstKey := "test-dst-key"
	dstBucket := testPrefix + "copy" + "-" + randomString(16)
	created, err = client.CreateBucket(context.Background(), &tos.CreateBucketInput{Bucket: dstBucket})
	checkSuccess(t, created, err, 200)
	defer cleanBucketV1(t, client, dstBucket)
	put, err = handle.PutObject(context.Background(), srcKey, strings.NewReader(content))
	checkSuccess(t, put, err, 200)
	copyOutput, err := handle.CopyObject(context.Background(), srcKey, dstKey)
	checkSuccess(t, copyOutput, err, 200)
	headObject, err = handle.HeadObject(context.Background(), dstKey)
	checkSuccess(t, headObject, err, 200)
	copyTo, err := handle.CopyObjectTo(context.Background(), dstBucket, dstKey, srcKey)
	checkSuccess(t, copyTo, err, 200)
	dstHandle, err := client.Bucket(dstBucket)
	require.Nil(t, err)
	headObject, err = dstHandle.HeadObject(context.Background(), dstKey)
	checkSuccess(t, headObject, err, 200)
	fromKey := "test-from-key"
	copyFrom, err := handle.CopyObjectFrom(context.Background(), dstBucket, dstKey, fromKey)
	checkSuccess(t, copyFrom, err, 200)
	headObject, err = handle.HeadObject(context.Background(), fromKey)
	checkSuccess(t, headObject, err, 200)
	// part copy
	partCopyKey := "test-part-copy-key"
	partCopySrcKey := "test-part-copy-src-key"
	put, err = dstHandle.PutObject(context.Background(), partCopySrcKey, strings.NewReader(content))
	checkSuccess(t, put, err, 200)
	multi, err = handle.CreateMultipartUpload(context.Background(), partCopyKey)
	checkSuccess(t, multi, err, 200)
	upload1, err = handle.UploadPart(context.Background(), &tos.UploadPartInput{
		Key:        partCopyKey,
		UploadID:   multi.UploadID,
		PartNumber: 1,
		Content:    strings.NewReader(part1),
	})
	checkSuccess(t, upload1, err, 200)
	partCopyOutput, err := handle.UploadPartCopy(context.Background(), &tos.UploadPartCopyInput{
		UploadID:       multi.UploadID,
		DestinationKey: partCopyKey,
		SourceBucket:   dstBucket,
		SourceKey:      partCopySrcKey,
		PartNumber:     2,
	})
	checkSuccess(t, upload2, err, 200)
	complete, err = handle.CompleteMultipartUpload(context.Background(), &tos.CompleteMultipartUploadInput{
		Key:           partCopyKey,
		UploadID:      multi.UploadID,
		UploadedParts: []tos.MultipartUploadedPart{upload1, partCopyOutput},
	})
	checkSuccess(t, complete, err, 200)
	// acl
	userID := "test-user-id"
	put, err = handle.PutObject(context.Background(), objectKey, strings.NewReader(content), tos.WithACLGrantRead("id="+userID))
	checkSuccess(t, put, err, 200)
	getAcl, err := handle.GetObjectAcl(context.Background(), objectKey)
	checkSuccess(t, getAcl, err, 200)
	require.Equal(t, userID, getAcl.Grants[0].Grantee.ID)

}
