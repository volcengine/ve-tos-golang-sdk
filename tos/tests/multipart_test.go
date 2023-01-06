package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// TestCreateMultipartUploadV2 test CreateMultipartUploadV2,UploadPartV2,ListPartsV2,UploadPartCopyV2,CompleteMultipartUploadV2
func TestMultipartUpload(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("multi-part-upload")
		client  = env.prepareClient(bucket)
		copyKey = "key-copyKey"
		key     = "key-test-create-multipart-upload"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	upload, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	buf := make([]byte, 5<<20)
	part1, err := client.UploadPartV2(context.Background(), &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:     bucket,
			Key:        key,
			UploadID:   upload.UploadID,
			PartNumber: 1,
		},
		Content: bytes.NewReader(buf),
	})
	require.Nil(t, err)
	part2, err := client.UploadPartV2(context.Background(), &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:     bucket,
			Key:        key,
			UploadID:   upload.UploadID,
			PartNumber: 2,
		},
		Content: bytes.NewReader(buf),
	})
	require.Nil(t, err)

	putRandomObject(t, client, bucket, copyKey, 6*1024*1024)

	part3, err := client.UploadPartCopyV2(context.Background(), &tos.UploadPartCopyV2Input{
		Bucket:          bucket,
		Key:             key,
		UploadID:        upload.UploadID,
		PartNumber:      3,
		CopySourceRange: fmt.Sprintf("bytes=0-%d", 5*1024*1024-1),
		SrcBucket:       bucket,
		SrcKey:          copyKey,
	})
	require.Nil(t, err)

	putRandomObject(t, client, bucket, copyKey, 4*1024)

	part4, err := client.UploadPartCopyV2(context.Background(), &tos.UploadPartCopyV2Input{
		Bucket:     bucket,
		Key:        key,
		UploadID:   upload.UploadID,
		PartNumber: 4,
		SrcBucket:  bucket,
		SrcKey:     copyKey,
	})
	require.Nil(t, err)

	parts, err := client.ListParts(context.Background(), &tos.ListPartsInput{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
	})
	require.Equal(t, 4, len(parts.Parts))
	require.Equal(t, part1.ETag, parts.Parts[0].ETag)
	require.Equal(t, part2.ETag, parts.Parts[1].ETag)
	require.Equal(t, part3.ETag, parts.Parts[2].ETag)
	require.Equal(t, part4.ETag, parts.Parts[3].ETag)

	complete, err := client.CompleteMultipartUploadV2(context.Background(), &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
		Parts: []tos.UploadedPartV2{{
			PartNumber: part1.PartNumber,
			ETag:       part1.ETag,
		}, {
			PartNumber: part2.PartNumber,
			ETag:       part2.ETag,
		}, {
			PartNumber: part3.PartNumber,
			ETag:       part3.ETag,
		}, {
			PartNumber: part4.PartNumber,
			ETag:       part4.ETag,
		}},
	})
	require.Nil(t, err)
	require.NotEqual(t, len(complete.ETag), 0)
	require.NotEqual(t, len(complete.Location), 0)
	require.NotEqual(t, len(complete.Bucket), 0)
	require.NotEqual(t, len(complete.Key), 0)
}

func TestUploadPartCopyRange(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("upload-copy-range")
		client = env.prepareClient(bucket)
		key    = randomString(6)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	rowKey := randomString(6)
	value := randomString(6)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: rowKey},
		Content:             strings.NewReader(value),
	})
	require.Nil(t, err)
	upload, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	part, err := client.UploadPartCopyV2(ctx, &tos.UploadPartCopyV2Input{
		Bucket:          bucket,
		Key:             key,
		UploadID:        upload.UploadID,
		PartNumber:      1,
		SrcBucket:       bucket,
		SrcKey:          rowKey,
		CopySourceRange: "bytes=0-0",
	})
	require.Nil(t, err)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
		Parts: []tos.UploadedPartV2{{
			PartNumber: 1,
			ETag:       part.ETag,
		}},
	})
	require.Nil(t, err)

	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, get.GetObjectBasicOutput.ContentLength, int64(1))
	b, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(value[0]), string(b))
}

func TestConcurrenceMultipartUpload(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("multi-part-upload")
		client = env.prepareClient(bucket)
		key    = "key-test-create-multipart-upload"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	upload, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	data := randomString(20 * 1024 * 1024)
	concurNum := 30
	var wg sync.WaitGroup
	wg.Add(concurNum)
	parts := make([]tos.UploadedPartV2, 0)
	for i := 1; i < concurNum+1; i++ {
		go func(i int) {
			defer wg.Done()
			part, err := client.UploadPartV2(context.Background(), &tos.UploadPartV2Input{
				UploadPartBasicInput: tos.UploadPartBasicInput{
					Bucket:     bucket,
					Key:        key,
					UploadID:   upload.UploadID,
					PartNumber: i,
				},
				Content: strings.NewReader(data),
			})
			require.Nil(t, err)
			fmt.Println("Current:", i)
			parts = append(parts, tos.UploadedPartV2{
				PartNumber: i,
				ETag:       part.ETag,
			})
		}(i)
	}
	wg.Wait()

	complete, err := client.CompleteMultipartUploadV2(context.Background(), &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
		Parts:    parts})
	require.Nil(t, err)
	require.NotEqual(t, len(complete.ETag), 0)
	require.NotEqual(t, len(complete.Location), 0)
	require.NotEqual(t, len(complete.Bucket), 0)
	require.NotEqual(t, len(complete.Key), 0)

}

func TestAbortMultipartUpload(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("abort-multi-part-upload")
		key    = "key-test-create-multipart-upload"
		client = env.prepareClient(bucket)
	)
	defer func() {}()
	upload1, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)

	upload2, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})

	require.Nil(t, err)
	listMulti, err := client.ListMultipartUploadsV2(context.Background(), &tos.ListMultipartUploadsV2Input{
		Bucket: bucket,
	})
	require.Equal(t, 200, listMulti.StatusCode)
	require.Equal(t, 2, len(listMulti.Uploads))
	sort1 := []string{upload1.UploadID, upload2.UploadID}
	sort2 := []string{listMulti.Uploads[0].UploadID, listMulti.Uploads[1].UploadID}
	sort.Strings(sort1)
	sort.Strings(sort2)
	require.Equal(t, sort1[0], sort2[0])
	require.Equal(t, sort1[1], sort2[1])

	for _, upload := range listMulti.Uploads {
		abort, err := client.AbortMultipartUpload(context.Background(), &tos.AbortMultipartUploadInput{
			Bucket:   bucket,
			Key:      upload.Key,
			UploadID: upload.UploadID,
		})
		require.Nil(t, err)
		require.Equal(t, 204, abort.StatusCode)
	}
}

func TestUploadPartFromFile(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-part-from-file")
		client   = env.prepareClient(bucket)
		key      = "key123"
		value1   = randomString(5 * 1024 * 1024)
		value2   = randomString(5 * 1024 * 1024)
		md5Sum   = md5s(value1 + value2)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	defer cleanTestFile(t, fileName)
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	upload, err := client.CreateMultipartUploadV2(context.Background(), &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	part1, err := client.UploadPartFromFile(context.Background(), &tos.UploadPartFromFileInput{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:     bucket,
			Key:        key,
			UploadID:   upload.UploadID,
			PartNumber: 1,
		},
		FilePath: fileName,
		Offset:   0,
		PartSize: int64(len(value1)),
	})
	require.Nil(t, err)
	part2, err := client.UploadPartV2(context.Background(), &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:     bucket,
			Key:        key,
			UploadID:   upload.UploadID,
			PartNumber: 2,
		},
		Content: strings.NewReader(value2),
	})
	require.Nil(t, err)
	parts, err := client.ListParts(context.Background(), &tos.ListPartsInput{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
	})
	require.Equal(t, 2, len(parts.Parts))
	require.Equal(t, part1.ETag, parts.Parts[0].ETag)
	require.Equal(t, part2.ETag, parts.Parts[1].ETag)
	complete, err := client.CompleteMultipartUploadV2(context.Background(), &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      key,
		UploadID: upload.UploadID,
		Parts: []tos.UploadedPartV2{{
			PartNumber: part1.PartNumber,
			ETag:       part1.ETag,
		}, {
			PartNumber: part2.PartNumber,
			ETag:       part2.ETag,
		}},
	})
	checkSuccess(t, complete, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	content, err := ioutil.ReadAll(get.Content)
	require.Equal(t, md5Sum, md5s(string(content)))
	require.NotEqual(t, len(complete.ETag), 0)
	require.NotEqual(t, len(complete.Location), 0)
	require.NotEqual(t, len(complete.Bucket), 0)
	require.NotEqual(t, len(complete.Key), 0)
}
