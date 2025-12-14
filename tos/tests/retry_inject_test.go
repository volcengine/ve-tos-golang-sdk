package tests

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// retryInjectRT injects server errors for first N calls, then delegates to base transport.
type retryInjectRT struct {
	base          http.RoundTripper
	failTimes     int
	statusCode    int
	retryAfterSec int
	counts        map[string]int
}

func newRetryInjectRT(status, failTimes, retryAfterSec int) *retryInjectRT {
	// base transport with relaxed TLS verify to match existing tests
	base := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	return &retryInjectRT{base: base, failTimes: failTimes, statusCode: status, retryAfterSec: retryAfterSec, counts: make(map[string]int)}
}

func (rt *retryInjectRT) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()
	c := rt.counts[key]
	if c < rt.failTimes {
		rt.counts[key] = c + 1
		hdr := make(http.Header)
		if rt.retryAfterSec > 0 {
			hdr.Set("Retry-After", strconv.Itoa(rt.retryAfterSec))
		}
		return &http.Response{
			StatusCode: rt.statusCode,
			Header:     hdr,
			Body:       ioutil.NopCloser(strings.NewReader("")),
			Request:    req,
		}, nil
	}
	return rt.base.RoundTrip(req)
}

func TestPutObjectRetryOn429WithRetryAfter_Seeker(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("retry-429")
		key    = randomString(6)
		ctx    = context.Background()
	)
	rt := newRetryInjectRT(429, 3, 1)
	client := env.prepareClient(bucket, tos.WithHTTPTransport(rt), tos.WithMaxRetryCount(10))
	defer cleanBucket(t, client, bucket)

	raw := strings.Repeat("RET429", 16*1024)
	listener := &dataTransferListenerTest{}
	// strings.NewReader implements io.Seeker; retry path should seek from start for each retry
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, DataTransferListener: listener},
		Content:             strings.NewReader(raw),
	})
	require.Nil(t, err)
	// retried at least failTimes times
	require.GreaterOrEqual(t, listener.RetryCount, 3)

	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, raw, string(data))
}

func TestPutObjectRetryOn408_Seeker(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("retry-408")
		key    = randomString(6)
		ctx    = context.Background()
	)
	rt := newRetryInjectRT(408, 2, 0)
	client := env.prepareClient(bucket, tos.WithHTTPTransport(rt), tos.WithMaxRetryCount(10))
	defer cleanBucket(t, client, bucket)

	raw := strings.Repeat("RET408", 8*1024)
	listener := &dataTransferListenerTest{}
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, DataTransferListener: listener},
		Content:             strings.NewReader(raw),
	})
	require.Nil(t, err)
	require.GreaterOrEqual(t, listener.RetryCount, 2)
}

func TestUploadPartRetryOn429_Seeker(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("retry-429-mp")
		key    = randomString(6)
		ctx    = context.Background()
	)
	rt := newRetryInjectRT(429, 3, 1)
	client := env.prepareClient(bucket, tos.WithHTTPTransport(rt), tos.WithMaxRetryCount(10))
	defer cleanBucket(t, client, bucket)

	// Initiate multipart
	createOut, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)

	raw := strings.Repeat("UP-RET429", 16*1024)
	listener := &dataTransferListenerTest{}
	// Upload a single part; strings.NewReader is seeker, so retry should work
	_, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: key, UploadID: createOut.UploadID, PartNumber: 1, DataTransferListener: listener},
		Content:              strings.NewReader(raw),
		ContentLength:        int64(len(raw)),
	})
	require.Nil(t, err)
	require.GreaterOrEqual(t, listener.RetryCount, 3)

	// Complete all
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{Bucket: bucket, Key: key, UploadID: createOut.UploadID, CompleteAll: true})
	require.Nil(t, err)

	// Verify content
	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, raw, string(data))
}
