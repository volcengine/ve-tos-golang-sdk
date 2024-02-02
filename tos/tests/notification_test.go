package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestNotificationFunc(t *testing.T) {
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
	require.NotNil(t, getOutput)
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

func TestNotificationMQ(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("notification-mq")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	input := tos.PutBucketNotificationInput{
		Bucket: bucket,
		RocketMQConfigurations: []tos.RocketMQConfiguration{
			{
				ID:   "TestCreateMQ",
				Role: fmt.Sprintf("trn:iam::%s:role/%s", env.accountId, env.mqRoleName),

				Events: []string{"tos:ObjectCreated:Post", "tos:ObjectCreated:Origin"},
				Filter: tos.Filter{Key: tos.FilterKey{Rules: []tos.FilterRule{{
					Name:  "prefix",
					Value: "test-",
				}, {
					Name:  "suffix",
					Value: "-ci",
				},
				}}},
				RocketMQ: tos.RocketMQConf{
					InstanceID:  env.mqInstanceId,
					Topic:       "SDK",
					AccessKeyID: env.mqAccessKeyID,
				},
			},
		},
	}
	output, err := client.PutBucketNotification(ctx, &input)
	require.Nil(t, err)
	require.NotNil(t, output)

	getOutput, err := client.GetBucketNotification(ctx, &tos.GetBucketNotificationInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, getOutput)

	require.Equal(t, len(getOutput.RocketMQConfigurations), len(input.RocketMQConfigurations))
	require.Equal(t, len(getOutput.RocketMQConfigurations[0].Events), len(input.RocketMQConfigurations[0].Events))
	require.Equal(t, getOutput.RocketMQConfigurations[0].ID, input.RocketMQConfigurations[0].ID)
	require.Equal(t, getOutput.RocketMQConfigurations[0].Role, input.RocketMQConfigurations[0].Role)
	require.Equal(t, getOutput.RocketMQConfigurations[0].RocketMQ, input.RocketMQConfigurations[0].RocketMQ)
	require.Equal(t, len(getOutput.RocketMQConfigurations[0].Filter.Key.Rules), len(input.RocketMQConfigurations[0].Filter.Key.Rules))
	for _, rule := range getOutput.RocketMQConfigurations[0].Filter.Key.Rules {
		found := false
		for _, ir := range input.RocketMQConfigurations[0].Filter.Key.Rules {
			if rule.Name == ir.Name {
				require.Equal(t, rule.Value, ir.Value)
				found = true
			}
		}
		require.True(t, found)
	}
	input.RocketMQConfigurations = nil
	output, err = client.PutBucketNotification(ctx, &input)
	require.Nil(t, err)
	require.NotNil(t, output)
}

func TestNotificationType2Func(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("notification")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	input := tos.PutBucketNotificationType2Input{
		Bucket: bucket,
		Rules: []tos.NotificationRule{{
			RuleID: "TestCreatePrefixSuffix",
			Events: []string{"tos:ObjectCreated:Post", "tos:ObjectCreated:Origin"},
			Filter: tos.NotificationFilter{TOSKey: tos.NotificationFilterKey{FilterRules: []tos.NotificationFilterRule{{
				Name:  "prefix",
				Value: "test-",
			}, {
				Name:  "suffix",
				Value: "-ci",
			}}}},
			Destination: tos.NotificationDestination{VeFaaS: []tos.DestinationVeFaaS{{FunctionID: env.cloudFunction}}, RocketMQ: []tos.DestinationRocketMQ{{
				Role:        fmt.Sprintf("trn:iam::%s:role/%s", env.accountId, env.mqRoleName),
				InstanceID:  env.mqInstanceId,
				Topic:       "SDK",
				AccessKeyID: env.mqAccessKeyID,
			}}},
		}}}
	output, err := client.PutBucketNotificationType2(ctx, &input)
	require.Nil(t, err)
	require.NotNil(t, output)

	getOutput, err := client.GetBucketNotificationType2(ctx, &tos.GetBucketNotificationType2Input{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, getOutput)

	require.Equal(t, len(getOutput.Rules), len(input.Rules))
	require.Equal(t, len(getOutput.Rules[0].Events), len(input.Rules[0].Events))
	require.Equal(t, getOutput.Rules[0].RuleID, input.Rules[0].RuleID)
	require.Equal(t, getOutput.Rules[0].Filter, input.Rules[0].Filter)
	require.Equal(t, getOutput.Rules[0].Destination, input.Rules[0].Destination)

}
