package tests

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type turboRoundTripFunc func(*http.Request) (*http.Response, error)

func (f turboRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestTurbo(t *testing.T) {
	var (
		env = newTestEnv(t)
		// bucket = generateBucketName("turbo")
		bucket = "ywt-turbo-2"
		// key    = randomString(10)
		// value1 = randomString(1024 * 1024)
		// value2 = randomString(4 * 1024)
		client = env.prepareClient(bucket)
	)
	// defer func() {
	// 	cleanBucket(t, client, bucket)
	// }()
	// 1. OpenTurbo
	// openOutput, err := client.OpenTurbo(context.Background(), &tos.OpenTurboInput{
	// 	Bucket:  bucket,
	// 	Key:     key,
	// 	Mode:    enum.OpenCreate,
	// 	Content: strings.NewReader(value1),
	// })
	// checkSuccess(t, openOutput, err, 200)
	// require.NotNil(t, openOutput)
	// require.True(t, openOutput.NextTurboOffset > 0)
	// require.True(t, openOutput.TurboToken != "")
	// // 2. AppendTurbo
	// appendOutput, err := client.AppendTurbo(context.Background(), &tos.AppendTurboInput{
	// 	Bucket:     bucket,
	// 	Key:        key,
	// 	TurboToken: openOutput.TurboToken,
	// 	Content:    strings.NewReader(value2),
	// })
	// checkSuccess(t, appendOutput, err, 200)
	// require.NotNil(t, appendOutput)
	// require.True(t, appendOutput.NextTurboOffset > openOutput.NextTurboOffset)
	// require.True(t, appendOutput.TurboToken != "")
	// // 3. ListOpenedTurbo
	// listOutput, err := client.ListOpenedTurbo(context.Background(), &tos.ListOpenedTurboInput{
	// 	Bucket: bucket,
	// })
	// checkSuccess(t, listOutput, err, 200)
	// t.Log(listOutput)
	// t.Log("ywt")
	// require.NotNil(t, listOutput)
	// require.Equal(t, listOutput.IsTruncated, true)
	// require.True(t, listOutput.NextContinuationToken != "")
	// require.True(t, listOutput.ContinuationToken == "")
	// found := false
	// for _, obj := range listOutput.Contents {
	// 	t.Log(obj.Key)
	// 	if obj.Key == key {
	// 		found = true
	// 		break
	// 	}
	// }
	// require.True(t, found, "should find the opened turbo object")
	// // 4. CloseTurbo
	// closeOutput, err := client.CloseTurbo(context.Background(), &tos.CloseTurboInput{
	// 	Bucket: bucket,
	// 	Key:    key,
	// 	Mode:   enum.TemporaryClose,
	// })
	// checkSuccess(t, closeOutput, err, 200)
	// require.NotNil(t, closeOutput)
	// // 1. Open no data
	// key = randomString(10)
	// openOutput, err = client.OpenTurbo(context.Background(), &tos.OpenTurboInput{
	// 	Bucket: bucket,
	// 	Key:    key,
	// 	Mode:   enum.OpenCreate,
	// })
	// checkSuccess(t, openOutput, err, 200)
	// require.NotNil(t, openOutput)
	// require.True(t, openOutput.NextTurboOffset == 0)
	// require.True(t, openOutput.TurboToken != "")
	// // 2. AppendTurbo
	// appendOutput, err = client.AppendTurbo(context.Background(), &tos.AppendTurboInput{
	// 	Bucket:     bucket,
	// 	Key:        key,
	// 	TurboToken: openOutput.TurboToken,
	// 	Content:    strings.NewReader(value2),
	// })
	// checkSuccess(t, appendOutput, err, 200)
	// require.NotNil(t, appendOutput)
	// require.True(t, appendOutput.NextTurboOffset > openOutput.NextTurboOffset)
	// require.True(t, appendOutput.TurboToken != "")
	// closeOutput, err = client.CloseTurbo(context.Background(), &tos.CloseTurboInput{
	// 	Bucket: bucket,
	// 	Key:    key,
	// 	Mode:   enum.PermanentClose,
	// })
	// checkSuccess(t, closeOutput, err, 200)
	// require.NotNil(t, closeOutput)

	getCountOutput, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: bucket,
	})
	checkSuccess(t, getCountOutput, err, 200)
	require.True(t, getCountOutput.OpenedCount > 0)
	t.Logf("getCountOutput.OpenedCount %v", getCountOutput.OpenedCount)
}

func TestGetOpenedTurboCount(t *testing.T) {
	var gotReq *http.Request
	client, err := tos.NewClientV2("https://tos-cn-beijing.volces.com",
		tos.WithRegion("cn-beijing"),
		tos.WithCredentials(tos.NewStaticCredentials("ak", "sk")),
		tos.WithHTTPTransport(turboRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotReq = req
			header := http.Header{}
			header.Set(tos.HeaderTurboOpenedCount, "123")
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		})))
	require.NoError(t, err)

	output, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: "bucket",
	})

	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, int64(123), output.OpenedCount)
	require.NotNil(t, gotReq)
	require.Equal(t, http.MethodGet, gotReq.Method)
	require.Equal(t, "", gotReq.URL.Query().Get("getopenedturbocount"))
	require.Contains(t, gotReq.URL.RawQuery, "getopenedturbocount=")
}

func TestGetOpenedTurboCount_InvalidBucket(t *testing.T) {
	called := false
	client, err := tos.NewClientV2("https://tos-cn-beijing.volces.com",
		tos.WithRegion("cn-beijing"),
		tos.WithCredentials(tos.NewStaticCredentials("ak", "sk")),
		tos.WithHTTPTransport(turboRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})))
	require.NoError(t, err)

	output, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: "-bad-",
	})
	require.Error(t, err)
	require.Nil(t, output)
	require.False(t, called, "transport should not be invoked when bucket name is invalid")
}

func TestGetOpenedTurboCount_TransportError(t *testing.T) {
	transportErr := errors.New("network down")
	client, err := tos.NewClientV2("https://tos-cn-beijing.volces.com",
		tos.WithRegion("cn-beijing"),
		tos.WithCredentials(tos.NewStaticCredentials("ak", "sk")),
		tos.WithHTTPTransport(turboRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			return nil, transportErr
		})))
	require.NoError(t, err)

	output, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: "bucket",
	})
	require.Error(t, err)
	require.Nil(t, output)
}

func TestGetOpenedTurboCount_MissingHeader(t *testing.T) {
	client, err := tos.NewClientV2("https://tos-cn-beijing.volces.com",
		tos.WithRegion("cn-beijing"),
		tos.WithCredentials(tos.NewStaticCredentials("ak", "sk")),
		tos.WithHTTPTransport(turboRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		})))
	require.NoError(t, err)

	output, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: "bucket",
	})
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, int64(0), output.OpenedCount)
}

func TestGetOpenedTurboCount_NonNumericHeader(t *testing.T) {
	client, err := tos.NewClientV2("https://tos-cn-beijing.volces.com",
		tos.WithRegion("cn-beijing"),
		tos.WithCredentials(tos.NewStaticCredentials("ak", "sk")),
		tos.WithHTTPTransport(turboRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			header := http.Header{}
			header.Set(tos.HeaderTurboOpenedCount, "not-a-number")
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader("")),
				Request:    req,
			}, nil
		})))
	require.NoError(t, err)

	output, err := client.GetOpenedTurboCount(context.Background(), &tos.GetOpenedTurboCountInput{
		Bucket: "bucket",
	})
	require.NoError(t, err)
	require.NotNil(t, output)
	require.Equal(t, int64(0), output.OpenedCount)
}
