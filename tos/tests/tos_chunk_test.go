package tests

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

type unknowSize struct {
	body io.Reader
}

func (u unknowSize) Read(p []byte) (n int, err error) {
	return u.body.Read(p)
}

func TestTosContentLengthChunk(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("tos-chunk")
		ctx    = context.Background()
		client = env.prepareClient(bucket, tos.WithDisableTrailerHeader(false))
	)
	contentEncoding := "gzip, br"
	rawData := randomString(5 * 1024)
	key := "test-ceshi-key"
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ContentEncoding: contentEncoding},
		Content:             strings.NewReader(rawData),
	})
	require.Nil(t, err)

	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)
	require.Equal(t, get.ContentEncoding, contentEncoding)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=0-0"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData[:1])
	require.Equal(t, get.ContentLength, int64(1))
	require.Equal(t, get.ContentEncoding, contentEncoding)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(rawData),
	})
	require.Nil(t, err)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)

	partKey := "multi-part-key"
	createMultiPart, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    partKey,
	})
	require.Nil(t, err)
	uploadResp, err := client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: partKey, UploadID: createMultiPart.UploadID, PartNumber: 1},
		Content:              strings.NewReader(rawData),
	})
	require.Nil(t, err)
	fmt.Println(uploadResp)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:      bucket,
		Key:         partKey,
		CompleteAll: true,
		UploadID:    createMultiPart.UploadID,
	})
	require.Nil(t, err)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=1-57"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData[1:58])

	appendData := randomString(1024 * 1024)
	appendKey := "append-key"
	_, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:  bucket,
		Key:     appendKey,
		Content: strings.NewReader(appendData),
	})
	require.Nil(t, err)
	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: appendKey, Range: "bytes=0-0"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), appendData[:1])
	require.Equal(t, get.ContentLength, int64(1))

}

func TestTosContentChunk(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("tos-chunk")
		ctx    = context.Background()
		client = env.prepareClient(bucket, tos.WithDisableTrailerHeader(false))
	)
	contentEncoding := "gzip, br"
	rawData := randomString(5 * 1024)
	key := "test-ceshi-key"
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ContentEncoding: contentEncoding},
		Content:             unknowSize{body: strings.NewReader(rawData)},
	})
	require.Nil(t, err)

	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)
	require.Equal(t, get.ContentEncoding, contentEncoding)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=0-0"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData[:1])
	require.Equal(t, get.ContentLength, int64(1))
	require.Equal(t, get.ContentEncoding, contentEncoding)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(rawData),
	})
	require.Nil(t, err)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)

	partKey := "multi-part-key"
	createMultiPart, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    partKey,
	})
	require.Nil(t, err)
	uploadResp, err := client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: partKey, UploadID: createMultiPart.UploadID, PartNumber: 1},
		Content:              unknowSize{strings.NewReader(rawData)},
	})
	require.Nil(t, err)
	fmt.Println(uploadResp)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:      bucket,
		Key:         partKey,
		CompleteAll: true,
		UploadID:    createMultiPart.UploadID,
	})
	require.Nil(t, err)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=1-57"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), rawData[1:58])

	appendData := randomString(1024 * 1024)
	appendKey := "append-key"
	_, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:  bucket,
		Key:     appendKey,
		Content: strings.NewReader(appendData),
	})
	require.Nil(t, err)
	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: appendKey})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), appendData)

	get, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: appendKey, Range: "bytes=0-0"})
	require.Nil(t, err)
	body, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), appendData[:1])
	require.Equal(t, get.ContentLength, int64(1))

	key = randomString(7)
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: unknowSize{strings.NewReader("")},
	})
	require.Nil(t, err)
	resp, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, resp.ContentLength, int64(0))

	_, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=0-0"})
	require.NotNil(t, err)

	key = randomString(7)
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: unknowSize{strings.NewReader(key)},
	})
	require.Nil(t, err)
	resp, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, resp.ContentLength, int64(len(key)))

	resp, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: "bytes=1-2"})
	require.Nil(t, err)
	require.Equal(t, resp.ContentLength, int64(2))
	body, err = ioutil.ReadAll(resp.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), key[1:3])

}
