package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"strings"
	"testing"
)

func TestTurbo(t *testing.T) {
	var (
		env = newTestEnv(t)
		// bucket = generateBucketName("turbo")
		bucket = "ywt-turbo-1"
		key    = randomString(10)
		value1 = randomString(1024 * 1024)
		value2 = randomString(4 * 1024)
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	// 1. OpenTurbo
	openOutput, err := client.OpenTurbo(context.Background(), &tos.OpenTurboInput{
		Bucket:  bucket,
		Key:     key,
		Mode:    enum.OpenCreate,
		Content: strings.NewReader(value1),
	})
	checkSuccess(t, openOutput, err, 200)
	require.NotNil(t, openOutput)
	require.True(t, openOutput.NextTurboOffset > 0)
	require.True(t, openOutput.TurboToken != "")
	// 2. AppendTurbo
	appendOutput, err := client.AppendTurbo(context.Background(), &tos.AppendTurboInput{
		Bucket:     bucket,
		Key:        key,
		TurboToken: openOutput.TurboToken,
		Content:    strings.NewReader(value2),
	})
	checkSuccess(t, appendOutput, err, 200)
	require.NotNil(t, appendOutput)
	require.True(t, appendOutput.NextTurboOffset > openOutput.NextTurboOffset)
	require.True(t, appendOutput.TurboToken != "")
	// 3. ListOpenedTurbo
	listOutput, err := client.ListOpenedTurbo(context.Background(), &tos.ListOpenedTurboInput{
		Bucket: bucket,
	})
	checkSuccess(t, listOutput, err, 200)
	t.Log(listOutput)
	t.Log("ywt")
	require.NotNil(t, listOutput)
	require.Equal(t, listOutput.IsTruncated, true)
	require.True(t, listOutput.NextContinuationToken != "")
	require.True(t, listOutput.ContinuationToken == "")
	found := false
	for _, obj := range listOutput.Contents {
		t.Log(obj.Key)
		if obj.Key == key {
			found = true
			break
		}
	}
	require.True(t, found, "should find the opened turbo object")
	// 4. CloseTurbo
	closeOutput, err := client.CloseTurbo(context.Background(), &tos.CloseTurboInput{
		Bucket: bucket,
		Key:    key,
		Mode:   enum.TemporaryClose,
	})
	checkSuccess(t, closeOutput, err, 200)
	require.NotNil(t, closeOutput)
	// 1. Open no data
	key = randomString(10)
	openOutput, err = client.OpenTurbo(context.Background(), &tos.OpenTurboInput{
		Bucket: bucket,
		Key:    key,
		Mode:   enum.OpenCreate,
	})
	checkSuccess(t, openOutput, err, 200)
	require.NotNil(t, openOutput)
	require.True(t, openOutput.NextTurboOffset == 0)
	require.True(t, openOutput.TurboToken != "")
	// 2. AppendTurbo
	appendOutput, err = client.AppendTurbo(context.Background(), &tos.AppendTurboInput{
		Bucket:     bucket,
		Key:        key,
		TurboToken: openOutput.TurboToken,
		Content:    strings.NewReader(value2),
	})
	checkSuccess(t, appendOutput, err, 200)
	require.NotNil(t, appendOutput)
	require.True(t, appendOutput.NextTurboOffset > openOutput.NextTurboOffset)
	require.True(t, appendOutput.TurboToken != "")
	closeOutput, err = client.CloseTurbo(context.Background(), &tos.CloseTurboInput{
		Bucket: bucket,
		Key:    key,
		Mode:   enum.PermanentClose,
	})
	checkSuccess(t, closeOutput, err, 200)
	require.NotNil(t, closeOutput)
}
