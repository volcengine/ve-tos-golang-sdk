package tests

import (
    "context"
    "errors"
    "io"
    "io/ioutil"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// Retryable + offset reader for multipart
type mpRetryReaderWithOffset struct {
    base   string
    reader io.ReadSeeker
    count  int
    offset int64
}

func (r *mpRetryReaderWithOffset) Reset() error {
    r.reader = strings.NewReader(r.base)
    _, err := r.reader.Seek(r.offset, io.SeekStart)
    return err
}

func (r *mpRetryReaderWithOffset) Read(p []byte) (n int, err error) {
    if r.count == 5 {
        return r.reader.Read(p)
    }
    r.count++
    time.Sleep(3 * time.Second)
    return 0, errors.New("time out")
}

func (r *mpRetryReaderWithOffset) Seek(offset int64, whence int) (int64, error) {
    return r.reader.Seek(offset, whence)
}

func TestMultipartWithRetry_RetryableOffset(t *testing.T) {
    var (
        env    = newTestEnv(t)
        bucket = generateBucketName("mp-retry-offset")
        key    = randomString(6)
        ctx    = context.Background()
    )
    client := env.prepareClient(bucket, tos.WithMaxRetryCount(5))
    defer cleanBucket(t, client, bucket)

    // Prepare payload and offset
    raw := strings.Repeat("MP-RETRY-OFFSET-", 32*1024)
    offset := int64(1000)
    rr := &mpRetryReaderWithOffset{base: raw, reader: strings.NewReader(raw), offset: offset}
    // initial seek to non-zero offset
    _, err := rr.Seek(offset, io.SeekStart)
    require.Nil(t, err)

    // Initiate multipart
    createOut, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)

    listener := &dataTransferListenerTest{}
    // Upload single part from offset, use explicit ContentLength for the remaining
    partLen := int64(len(raw)) - offset
    _, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
        UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: key, UploadID: createOut.UploadID, PartNumber: 1, DataTransferListener: listener},
        Content:              rr,
        ContentLength:        partLen,
    })
    require.Nil(t, err)
    require.Equal(t, 5, rr.count)
    require.Equal(t, 5, listener.RetryCount)

    // Complete
    _, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{Bucket: bucket, Key: key, UploadID: createOut.UploadID, CompleteAll: true})
    require.Nil(t, err)

    // Verify object content equals substring from offset
    get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    data, err := ioutil.ReadAll(get.Content)
    require.Nil(t, err)
    require.Equal(t, raw[offset:], string(data))
}

func TestAppendRetry_RetryableOffset_CreateOnHNS(t *testing.T) {
    var (
        env    = newTestEnv(t)
        bucket = generateBucketName("append-retry-hns")
        client = env.prepareClient("")
        ctx    = context.Background()
    )
    // Create HNS bucket
    _, err := client.CreateBucketV2(ctx, &tos.CreateBucketV2Input{Bucket: bucket, BucketType: enum.BucketTypeHNS})
    require.Nil(t, err)
    // Best-effort cleanup to avoid test flakiness in certain environments
    // defer cleanHNSBucket(t, client, bucket)

    key := "key1-" + randomString(6)
    raw := strings.Repeat("APPEND-RETRY-OFFSET-", 16*1024)
    offset := int64(512)
    // Use a seeker reader positioned at non-zero offset
    ar := strings.NewReader(raw)
    _, err = ar.Seek(offset, io.SeekStart)
    require.Nil(t, err)

    listener := &dataTransferListenerTest{}
    // HNS append at offset 0 with non-existent object triggers PutObjectV2 internally (with retry)
    out, err := client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{Bucket: bucket, Key: key, Content: ar, Offset: 0, DataTransferListener: listener})
    require.Nil(t, err)
    require.Equal(t, 0, listener.RetryCount)

    // Verify content equals substring from offset
    get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    data, err := ioutil.ReadAll(get.Content)
    require.Nil(t, err)
    require.Equal(t, raw[offset:], string(data))

    // Subsequent append on HNS follows modify path with NoRetry; listener RetryCount should remain 0
    listener = &dataTransferListenerTest{}
    _, err = client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{Bucket: bucket, Key: key, Content: strings.NewReader(raw), Offset: out.NextAppendOffset, PreHashCrc64ecma: out.HashCrc64ecma, DataTransferListener: listener})
    require.Nil(t, err)
    require.Equal(t, 0, listener.RetryCount)

    // Verify final length
    head, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    require.Equal(t, int64(len(raw[offset:])+len(raw)), head.ContentLength)
}
