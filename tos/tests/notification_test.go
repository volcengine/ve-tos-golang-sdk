package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestNotification(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("notification")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	input := tos.PutBucketNotificationInput{
		Bucket: bucket,
		CloudFunctionConfigurations: []tos.CloudFunctionConfiguration{
			{
				ID:     "TestCreatePrefixSuffix",
				Events: []string{"tos:ObjectCreated:Post", "tos:ObjectCreated:Origin"},
				Filter: tos.Filter{Key: tos.FilterKey{Rules: []tos.FilterRule{{
					Name:  "prefix",
					Value: "test-",
				}, {
					Name:  "suffix",
					Value: "-ci",
				},
				}}},
				CloudFunction: env.cloudFunction,
			},
		},
	}
	output, err := client.PutBucketNotification(ctx, &input)
	require.Nil(t, err)
	require.NotNil(t, output)

	getOutput, err := client.GetBucketNotification(ctx, &tos.GetBucketNotificationInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, output)
	require.Equal(t, len(getOutput.CloudFunctionConfigurations), len(input.CloudFunctionConfigurations))
	require.Equal(t, len(getOutput.CloudFunctionConfigurations[0].Events), len(input.CloudFunctionConfigurations[0].Events))
	require.Equal(t, getOutput.CloudFunctionConfigurations[0].ID, input.CloudFunctionConfigurations[0].ID)
	require.Equal(t, getOutput.CloudFunctionConfigurations[0].CloudFunction, input.CloudFunctionConfigurations[0].CloudFunction)
	require.Equal(t, len(getOutput.CloudFunctionConfigurations[0].Filter.Key.Rules), len(input.CloudFunctionConfigurations[0].Filter.Key.Rules))
	for _, rule := range getOutput.CloudFunctionConfigurations[0].Filter.Key.Rules {
		found := false
		for _, ir := range input.CloudFunctionConfigurations[0].Filter.Key.Rules {
			if rule.Name == ir.Name {
				require.Equal(t, rule.Value, ir.Value)
				found = true
			}
		}
		require.True(t, found)
	}

}
