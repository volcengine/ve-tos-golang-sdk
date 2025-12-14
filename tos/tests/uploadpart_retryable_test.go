package tests

import (
    "context"
    "io"
    "io/ioutil"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// nonSeekerRetryable for UploadPart: does not implement io.Seeker; Reset recreates underlying reader
type upNonSeekerRetryable struct {
    base   string
    reader io.Reader
    count  int
}

func (r *upNonSeekerRetryable) Reset() error {
    r.reader = strings.NewReader(r.base)
    return nil
}

func (r *upNonSeekerRetryable) Read(p []byte) (n int, err error) {
    if r.count == 5 {
        return r.reader.Read(p)
    }
    r.count++
    // induce timeout on write to trigger retry
    time.Sleep(3 * time.Second)
    return r.reader.Read(p)
}

func TestUploadPartWithRetry_RetryableNonSeeker(t *testing.T) {
    var (
        env    = newTestEnv(t)
        bucket = generateBucketName("uploadpart-retry-nonseek")
        key    = randomString(6)
        ctx    = context.Background()
    )
    tsConfig := tos.DefaultTransportConfig()
    tsConfig.ReadTimeout = time.Millisecond * 500
    tsConfig.WriteTimeout = time.Millisecond * 500
    client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5))
    defer cleanBucket(t, client, bucket)

    // init multipart upload
    initOut, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)

    raw := strings.Repeat("UP-NONSEEK-RETRY", 32*1024)
    data := &upNonSeekerRetryable{base: raw, reader: strings.NewReader(raw)}
    listener := &dataTransferListenerTest{}

    _, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
        UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: key, UploadID: initOut.UploadID, PartNumber: 1, DataTransferListener: listener},
        Content:              data,
        ContentLength:        int64(len(raw)),
    })
    require.Nil(t, err)
    require.Equal(t, 5, data.count)
    require.Equal(t, 5, listener.RetryCount)

    // complete all
    _, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{Bucket: bucket, Key: key, UploadID: initOut.UploadID, CompleteAll: true})
    require.Nil(t, err)

    // verify object content
    getOut, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    content, err := ioutil.ReadAll(getOut.Content)
    require.Nil(t, err)
    require.Equal(t, raw, string(content))
}

