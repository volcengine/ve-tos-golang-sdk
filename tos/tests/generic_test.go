package tests

import (
	"bytes"
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"io"
	"io/ioutil"
	"testing"
)

func TestGeneric(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("generic-bucket")
		client = env.prepareClient(bucket, LongTimeOutClientOption...)
	)
	key := "test-" + randomString(8)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	content := randomString(1024)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		GenericInput: tos.GenericInput{
			RequestHeader: map[string]string{
				"X-Tos-Test-Header": "value",
				"X-test-demo":       "test",
				"Content-Length":    "1",
			},
			RequestQuery: map[string]string{
				"X-Tos-Query-测试": "value-测试",
			},
		},
		Content: bytes.NewReader([]byte(content)),
	})

	getResp, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
		GenericInput: tos.GenericInput{
			RequestHeader: map[string]string{
				"X-Tos-Test-Header": "value",
				"X-test-demo":       "test",
			},
			RequestQuery: map[string]string{
				"X-Tos-Query-测试": "value-测试",
			},
		},
	})
	require.Nil(t, err)
	body, err := ioutil.ReadAll(getResp.Content)
	require.Nil(t, err)
	require.Equal(t, string(body), content)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             &mockErrReader{},
	})
	require.NotNil(t, err)
}

type mockErrReader struct {
	readCount int
}

func (m *mockErrReader) Read(p []byte) (n int, err error) {
	s := "1234567890"
	if m.readCount == 0 {
		m.readCount++
		copy(p, s)
		return len(s), errors.New("mock error")
	}
	if m.readCount == 10 {
		return 0, io.EOF
	}
	m.readCount++
	copy(p, s)

	return len(s), nil
}
