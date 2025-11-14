package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestAccessPoint(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("access-point")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	apName := generateBucketName("ap")

	resp, err := client.CreateAccessPoint(ctx, &tos.CreateAccessPointInput{AccountID: env.accountId, BucketAccountID: env.accountId, AccessPointName: apName, Bucket: bucket, NetworkOrigin: enum.NetworkOriginInternet})
	require.NoError(t, err)
	require.NotEmpty(t, resp)
	require.NotEqual(t, resp.Alias, "")
	require.NotEqual(t, resp.AccessPointTrn, "")

	getResp, err := client.GetAccessPoint(ctx, &tos.GetAccessPointInput{AccountID: env.accountId, AccessPointName: apName})
	require.NoError(t, err)
	require.NotEmpty(t, getResp)

	listResp, err := client.ListAccessPoints(ctx, &tos.ListAccessPointsInput{
		Bucket:    bucket,
		AccountID: env.accountId,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listResp)
	require.Equal(t, len(listResp.AccessPoints), 1)

	listAccResp, err := client.ListBindAcceleratorForAccessPoint(ctx, &tos.ListBindAcceleratorForAccessPointInput{
		AccountID:       env.accountId,
		AccessPointName: apName,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listAccResp)

	deleteResp, err := client.DeleteAccessPoint(ctx, &tos.DeleteAccessPointInput{AccountID: env.accountId, AccessPointName: apName})
	require.NoError(t, err)
	require.NotEmpty(t, deleteResp)
}
