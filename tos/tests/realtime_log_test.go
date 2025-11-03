package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketRealTimeLog(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("realtime-log")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	input := &tos.PutBucketRealTimeLogInput{
		Bucket: bucket,
		Configuration: tos.RealTimeLogConfiguration{
			Role: "TOSLogArchiveTLSRole",
			Configuration: tos.AccessLogConfiguration{
				UseServiceTopic: true,
			},
		},
	}
	_, err := client.PutBucketRealTimeLog(ctx, input)
	require.Nil(t, err)

	getOutput, err := client.GetBucketRealTimeLog(ctx, &tos.GetBucketRealTimeLogInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, getOutput.Configuration.Role, input.Configuration.Role)
	require.Equal(t, getOutput.Configuration.Configuration.TLSProjectID != "", true)
	require.Equal(t, getOutput.Configuration.Configuration.TLSTopicID != "", true)

	_, err = client.DeleteBucketRealTimeLog(ctx, &tos.DeleteBucketRealTimeLogInput{Bucket: bucket})
	require.Nil(t, err)

	_, err = client.GetBucketRealTimeLog(ctx, &tos.GetBucketRealTimeLogInput{Bucket: bucket})
	require.NotNil(t, err)
}
