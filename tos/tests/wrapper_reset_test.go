package tests

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestWrapperReset_TrailerEnabled_MemoryOffset(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("wrapper-reset-trailer")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	// 启用 Trailer，使封装链包含 listener/CRC 等，从而触发封装层 Reset
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithDisableTrailerHeader(false))
	defer cleanBucket(t, client, bucket)

	rawData := strings.Repeat("WRAP-RESET-OFFSET", 64*1024)
	offset := int64(1000)

	dataListener := &dataTransferListenerTest{}
	rr := &retryResetReaderWithOffset{base: rawData, reader: strings.NewReader(rawData), offset: offset}
	// 设置非零起始偏移，验证底层 Retryable.Reset 被调用后仍从同一偏移开始
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

	// 校验内容从偏移位置开始
	dataListener = &dataTransferListenerTest{}
	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, DataTransferListener: dataListener})
	require.Nil(t, err)
	got, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, rawData[offset:], string(got))
	require.Equal(t, 0, dataListener.RetryCount)
}

func TestWrapperReset_TrailerDisabled_ContentMD5(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("wrapper-reset-md5")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithDisableTrailerHeader(false))
	defer cleanBucket(t, client, bucket)

	rawData := strings.Repeat("WRAP-RESET-MD5", 16*1024)
	offset := int64(512)
	sum := md5.Sum([]byte(rawData[offset:]))
	md5b64 := base64.StdEncoding.EncodeToString(sum[:])

	dataListener := &dataTransferListenerTest{}
	rr := &retryResetReaderWithOffset{base: rawData, reader: strings.NewReader(rawData), offset: offset}
	_, err := rr.Seek(offset, io.SeekStart)
	require.Nil(t, err)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			ContentMD5:           md5b64,
			ContentLength:        int64(len(rawData) - int(offset)),
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

func TestSeekFallback_NoWrapper(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("seek-fallback")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	// 关闭 CRC，且不传 listener/limiter，确保封装层不实现 Retryable
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithEnableCRC(false))
	defer cleanBucket(t, client, bucket)

	raw := strings.Repeat("SEEK-FALLBACK", 16*1024)
	rr := &retryReader{reader: strings.NewReader(raw), t: t}
	offset := int64(256)
	_, err := rr.Seek(offset, io.SeekStart)
	require.Nil(t, err)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             rr,
	})
	require.Nil(t, err)
	require.Equal(t, 5, rr.count)

	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	got, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, raw[offset:], string(got))
}

type countingRetryableReader struct {
	base       string
	reader     io.Reader
	resetCount int
	readCount  int
}

// timeout error for classifier
// net.Error with Timeout=true, so http wraps it and classifier recognizes retry
type retryTimeoutNetErr struct{}

func (e *retryTimeoutNetErr) Error() string   { return "client request time out" }
func (e *retryTimeoutNetErr) Timeout() bool   { return true }
func (e *retryTimeoutNetErr) Temporary() bool { return true }

func (r *countingRetryableReader) Reset() error {
	r.resetCount++
	r.reader = strings.NewReader(r.base)
	return nil
}

func (r *countingRetryableReader) Read(p []byte) (n int, err error) {
	if r.readCount == 5 {
		return r.reader.Read(p)
	}
	r.readCount++
	time.Sleep(3 * time.Second)
	// 返回带 Timeout() 的错误以触发重试
	return 0, &retryTimeoutNetErr{}
}

func TestWrapperReset_PropagatesUnderlyingReset(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("wrapper-reset-propagate")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5), tos.WithDisableTrailerHeader(false))
	defer cleanBucket(t, client, bucket)

	raw := strings.Repeat("WRAP-COUNTING-RETRY", 16*1024)
	dataListener := &dataTransferListenerTest{}
	cr := &countingRetryableReader{base: raw, reader: strings.NewReader(raw)}

	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, DataTransferListener: dataListener},
		Content:             cr,
	})
	require.Nil(t, err)
	// 重试 5 次：封装层 Reset 应触发底层 Reset 5 次
	require.Equal(t, 5, cr.resetCount)

	// 读取验证内容
	getOut, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	body, err := ioutil.ReadAll(getOut.Content)
	require.Nil(t, err)
	require.Equal(t, raw, string(body))
}
