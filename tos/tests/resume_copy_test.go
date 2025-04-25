package tests

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestResumeCopyObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("resume-copy")
		client = env.prepareClient(bucket)
		//	rowSSECKey = randomString(32)
		// rowCopySSECKey = randomString(32)
		//	ssecKey = base64.StdEncoding.EncodeToString([]byte(rowSSECKey))
		// ssecCopyKey    = base64.StdEncoding.EncodeToString([]byte(rowCopySSECKey))
		//	ssecMd5 = md5s(rowSSECKey)
		// ssecCopyMd5    = md5s(rowCopySSECKey)
		// algorithm = "AES256"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	data := randomString(20 * 1024 * 1024)
	key := randomString(6)
	putout, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
			// SSECAlgorithm: algorithm,
			// SSECKey:       ssecKey,
			// SSECKeyMD5:    ssecMd5,
		},
		Content: strings.NewReader(data),
	})
	require.Nil(t, err)
	copyKey := randomString(6)
	_, err = client.ResumableCopyObject(ctx, &tos.ResumableCopyObjectInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    copyKey,
		},
		CopySourceIfNoneMatch: putout.ETag,
		SrcBucket:             bucket,
		SrcKey:                key,
		PartSize:              5 * 1024 * 1024,
		TaskNum:               4,
		EnableCheckpoint:      true,
	})
	require.NotNil(t, err)

	_, err = client.ResumableCopyObject(ctx, &tos.ResumableCopyObjectInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    copyKey,

			// SSECAlgorithm: algorithm,
			// SSECKeyMD5:    ssecCopyMd5,
			// SSECKey:       ssecCopyKey,

		},
		CopySourceIfMatch: putout.ETag,
		SrcBucket:         bucket,
		SrcKey:            key,
		PartSize:          5 * 1024 * 1024,
		TaskNum:           4,
		EnableCheckpoint:  true,
		// CopySourceSSECKey:       ssecKey,
		// CopySourceSSECAlgorithm: algorithm,
		// CopySourceSSECKeyMD5:    ssecMd5,
	})
	require.Nil(t, err)
	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    copyKey,
		// SSECKey:       ssecCopyKey,
		// SSECKeyMD5:    ssecCopyMd5,
		// SSECAlgorithm: algorithm,
	})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, len(string(getData)), len(data))
}

type copyCancelEvent struct {
	total  int64
	cancel tos.CancelHook
}

func (c *copyCancelEvent) EventChange(event *tos.CopyEvent) {
	switch event.Type {
	case enum.CopyEventUploadPartCopySuccess:
		c.total++
	default:
	}
	if c.total == 2 {
		c.cancel.Cancel(true)
	}
}

func TestResumeCopyObjectCancel(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("resume-copy")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	data := randomString(100 * 1024 * 1024)
	key := randomString(6)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(data),
	})
	require.Nil(t, err)
	copyKey := randomString(6)
	cancel := tos.NewCancelHook()
	event := &copyCancelEvent{
		cancel: cancel,
	}
	_, err = client.ResumableCopyObject(ctx, &tos.ResumableCopyObjectInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    copyKey,
		},
		SrcBucket:         bucket,
		SrcKey:            key,
		PartSize:          20 * 1024 * 1024,
		TaskNum:           2,
		EnableCheckpoint:  true,
		CheckpointFile:    "./checkpoint",
		CopyEventListener: event,
		CancelHook:        cancel,
	})
	require.NotNil(t, err)
	require.Equal(t, event.total < 5, true)
	require.Equal(t, event.total >= 2, true)

}

type copyEvent struct {
	partSuccess            int
	completeSuccess        int
	createMultipartSuccess int
}

func (c *copyEvent) EventChange(event *tos.CopyEvent) {
	switch event.Type {
	case enum.CopyEventCompleteMultipartUploadSucceed:
		c.completeSuccess++
	case enum.CopyEventUploadPartCopySuccess:
		c.partSuccess++
	case enum.CopyEventCreateMultipartUploadSucceed:
		c.createMultipartSuccess++
	default:
	}
}

func TestResumeCopyEvent(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("resume-copy")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	data := randomString(100 * 1024 * 1024)
	key := randomString(6)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(data),
	})
	require.Nil(t, err)
	copyKey := randomString(6)
	event := &copyEvent{}
	_, err = client.ResumableCopyObject(ctx, &tos.ResumableCopyObjectInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    copyKey,
		},
		SrcBucket:         bucket,
		SrcKey:            key,
		PartSize:          20 * 1024 * 1024,
		TaskNum:           2,
		EnableCheckpoint:  true,
		CheckpointFile:    "./checkpoint",
		CopyEventListener: event,
	})
	require.Nil(t, err)
	require.Equal(t, event.createMultipartSuccess, 1)
	require.Equal(t, event.completeSuccess, 1)
	require.Equal(t, event.partSuccess, 5)

}

func TestResumeCopyEmpty(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("resume-copy")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	// Empty Object
	data := ""
	key := randomString(6)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(data),
	})
	require.Nil(t, err)
	copyKey := randomString(6)
	event := &copyEvent{}
	_, err = client.ResumableCopyObject(ctx, &tos.ResumableCopyObjectInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    copyKey,
		},
		SrcBucket:         bucket,
		SrcKey:            key,
		PartSize:          20 * 1024 * 1024,
		TaskNum:           2,
		EnableCheckpoint:  true,
		CheckpointFile:    "./checkpoint",
		CopyEventListener: event,
	})
	require.Nil(t, err)
	require.Equal(t, event.createMultipartSuccess, 1)
	require.Equal(t, event.completeSuccess, 1)
	require.Equal(t, event.partSuccess, 1)

	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: copyKey})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, string(getData), data)
}
