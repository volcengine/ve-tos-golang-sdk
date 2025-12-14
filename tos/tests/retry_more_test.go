package tests

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// Test retry path when ContentMD5 is provided (trailers disabled), using Retryable reader.
func TestObjectWithRetry_ContentMD5_Retryable(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-retry-md5")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5))
	defer cleanBucket(t, client, bucket)

	rawData := strings.Repeat("hello world", 8)
	sum := md5.Sum([]byte(rawData))
	md5b64 := base64.StdEncoding.EncodeToString(sum[:])

	dataListener := &dataTransferListenerTest{}
	resetReader := &retryResetReader{reader: strings.NewReader(rawData), t: t}

	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			ContentMD5:           md5b64,
			ContentLength:        int64(len(rawData)),
			DataTransferListener: dataListener,
		},
		Content: resetReader,
	})
	require.Nil(t, err)
	require.Equal(t, 5, resetReader.count)
	require.True(t, dataListener.RetryCount >= 4)

	// Verify object content and that GET does not report retries
	dataListener = &dataTransferListenerTest{}
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, DataTransferListener: dataListener})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, rawData, string(getData))
	require.Equal(t, 0, dataListener.RetryCount)
}

// Retryable reader that does not implement io.Seeker; Reset recreates the underlying reader.
type nonSeekerRetryable struct {
	base   string
	reader io.Reader
	count  int
}

func (r *nonSeekerRetryable) Reset() error {
	r.reader = strings.NewReader(r.base)
	return nil
}

func (r *nonSeekerRetryable) Read(p []byte) (n int, err error) {
	if r.count == 5 {
		return r.reader.Read(p)
	}
	r.count++
	time.Sleep(3 * time.Second)
	return r.reader.Read(p)
}

// Test retry path with a non-seeker Retryable reader and explicit ContentLength.
func TestObjectWithRetry_RetryableNonSeeker(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-retry-nonseek")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5))
	defer cleanBucket(t, client, bucket)

	rawData := strings.Repeat("abcd", 1024)
	dataListener := &dataTransferListenerTest{}
	data := &nonSeekerRetryable{base: rawData, reader: strings.NewReader(rawData)}

	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			ContentLength:        int64(len(rawData)),
			DataTransferListener: dataListener,
		},
		Content: data,
	})
	require.Nil(t, err)
	require.Equal(t, 5, data.count)
	require.True(t, dataListener.RetryCount >= 4)

	dataListener = &dataTransferListenerTest{}
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, DataTransferListener: dataListener})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, rawData, string(getData))
	require.Equal(t, 0, dataListener.RetryCount)
}

// Test Seek-based retry with an in-memory reader and non-zero start offset under trailer-enabled streaming.
func TestObjectWithRetry_MemoryOffset_Seeker(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-retry-memoffset")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	// Enable trailer for unknown-length streaming
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithDisableTrailerHeader(false))
	defer cleanBucket(t, client, bucket)

	// Large payload to cross default chunk size and exercise trailer path
	rawData := strings.Repeat("xyz123", 64*1024)
	offset := 1000

	dataListener := &dataTransferListenerTest{}
	data := &retryReader{reader: strings.NewReader(rawData), t: t}
	// Move start position to a non-zero offset
	_, err := data.Seek(int64(offset), io.SeekStart)
	require.Nil(t, err)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			DataTransferListener: dataListener,
		},
		Content: data,
	})
	require.Nil(t, err)
	require.Equal(t, 5, data.count)
	require.Equal(t, 5, dataListener.RetryCount)

	dataListener = &dataTransferListenerTest{}
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, DataTransferListener: dataListener})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, rawData[offset:], string(getData))
	require.Equal(t, 0, dataListener.RetryCount)
}

// Retryable + Seeker: start from non-zero offset; Reset restores to the same offset.
type retryResetReaderWithOffset struct {
	base   string
	reader io.ReadSeeker
	count  int
	offset int64
}

func (r *retryResetReaderWithOffset) Reset() error {
	r.reader = strings.NewReader(r.base)
	_, err := r.reader.Seek(r.offset, io.SeekStart)
	return err
}

func (r *retryResetReaderWithOffset) Read(p []byte) (n int, err error) {
	if r.count == 5 {
		return r.reader.Read(p)
	}
	r.count++
	time.Sleep(3 * time.Second)
	return 0, errors.New("time out")
}

func (r *retryResetReaderWithOffset) Seek(offset int64, whence int) (int64, error) {
	return r.reader.Seek(offset, whence)
}

func TestObjectWithRetry_RetryableMemoryOffset(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-retryable-memoffset")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithDisableTrailerHeader(false))
	defer cleanBucket(t, client, bucket)

	rawData := strings.Repeat("HELLO-RETRYABLE-OFFSET", 64*1024)
	offset := int64(1000)

	dataListener := &dataTransferListenerTest{}
	rr := &retryResetReaderWithOffset{base: rawData, reader: strings.NewReader(rawData), offset: offset}
	// initial attempt begins at offset
	_, err := rr.Seek(offset, io.SeekStart)
	require.Nil(t, err)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			DataTransferListener: dataListener,
		},
		Content: rr,
	})
	require.Nil(t, err)
	require.Equal(t, 5, rr.count)
	require.Equal(t, 5, dataListener.RetryCount)

	dataListener = &dataTransferListenerTest{}
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, DataTransferListener: dataListener})
	require.Nil(t, err)
	got, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, rawData[offset:], string(got))
	require.Equal(t, 0, dataListener.RetryCount)
}
