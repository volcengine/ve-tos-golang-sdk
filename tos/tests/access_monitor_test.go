package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestBucketAccessMonitor(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("access-monitor")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()

	// 1. 调用 GetBucketAccessMonitor，断言默认状态为 Disabled
	getOutput, err := client.GetBucketAccessMonitor(context.Background(), &tos.GetBucketAccessMonitorInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOutput)
	require.Equal(t, enum.StatusDisabled, getOutput.Status)

	// 2. 调用 PutBucketAccessMonitor 开启 AccessMonitor
	putOutput, err := client.PutBucketAccessMonitor(context.Background(), &tos.PutBucketAccessMonitorInput{
		Bucket: bucket,
		Status: enum.StatusEnabled,
	})
	require.Nil(t, err)
	require.NotNil(t, putOutput)

	// 3. 调用 GetBucketAccessMonitor，断言状态为 Enabled
	getOutput, err = client.GetBucketAccessMonitor(context.Background(), &tos.GetBucketAccessMonitorInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOutput)
	require.Equal(t, enum.StatusEnabled, getOutput.Status)

	rule := tos.LifecycleRule{
		ID:     "1",
		Prefix: "test",
		Status: enum.LifecycleStatusEnabled,

		AccessTimeTransitions: []tos.AccessTimeTransition{{
			StorageClass: enum.StorageClassIa,
			Days:         20,
		}},
		NonCurrentVersionAccessTimeTransitions: []tos.NonCurrentVersionAccessTimeTransition{{
			StorageClass:   enum.StorageClassIa,
			NonCurrentDays: 20,
		}},
	}

	putRes, err := client.PutBucketLifecycle(ctx, &tos.PutBucketLifecycleInput{
		Bucket: bucket,
		Rules:  []tos.LifecycleRule{rule},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketLifecycle(ctx, &tos.GetBucketLifecycleInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getRes.Rules) == 1)
	require.Equal(t, getRes.Rules[0], rule)

	_, err = client.DeleteBucketLifecycle(ctx, &tos.DeleteBucketLifecycleInput{
		Bucket: bucket,
	})
	require.Nil(t, err)

	// 4. 调用 PutBucketAccessMonitor 关闭 AccessMonitor
	putOutput, err = client.PutBucketAccessMonitor(context.Background(), &tos.PutBucketAccessMonitorInput{
		Bucket: bucket,
		Status: enum.StatusDisabled,
	})
	require.Nil(t, err)
	require.NotNil(t, putOutput)

	// 5. 调用 GetBucketAccessMonitor，断言状态为 Disabled
	getOutput, err = client.GetBucketAccessMonitor(context.Background(), &tos.GetBucketAccessMonitorInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOutput)
	require.Equal(t, enum.StatusDisabled, getOutput.Status)
}
