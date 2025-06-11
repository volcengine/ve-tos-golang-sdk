package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestBucketObjectLock(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-lock")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()

	resp, err := client.PutBucketObjectLock(ctx, &tos.PutBucketObjectLockInput{Bucket: bucket, Configuration: tos.ObjectLockConfiguration{
		ObjectLockEnabled: enum.StatusDisabled,
	}})
	require.NotNil(t, err)
	require.Equal(t, err.(*tos.TosServerError).StatusCode, http.StatusBadRequest)

	conf := tos.ObjectLockConfiguration{
		ObjectLockEnabled: enum.ObjectLockEnabled,
		Rule: &tos.RetentionRule{
			DefaultRetention: tos.DefaultRetention{
				Days: 10,
				Mode: enum.RetentionModeCompliance,
			},
		},
	}
	resp, err = client.PutBucketObjectLock(ctx, &tos.PutBucketObjectLockInput{Bucket: bucket, Configuration: conf})
	require.Nil(t, err)
	require.True(t, len(resp.RequestID) > 0)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	getResp, err := client.GetBucketObjectLock(ctx, &tos.GetBucketObjectLockInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getResp.RequestID) > 0)
	require.Equal(t, getResp.Configuration.ObjectLockEnabled, enum.ObjectLockEnabled)
	require.Equal(t, getResp.Configuration.Rule.DefaultRetention.Years, int64(0))
	require.Equal(t, getResp.Configuration.Rule.DefaultRetention.Mode, enum.RetentionModeCompliance)
	require.Equal(t, getResp.Configuration.Rule.DefaultRetention.Days, int64(10))
}
