package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"testing"
)

func TestMrap(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("mrap-test")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer cleanBucket(t, client, bucket)

	resp, err := client.ListMultiRegionAccessPoints(ctx, &tos.ListMultiRegionAccessPointsInput{AccountID: env.accountId})
	require.Nil(t, err)

	for _, accessPoint := range resp.AccessPoints {
		_, err = client.DeleteMultiRegionAccessPoint(ctx, &tos.DeleteMultiRegionAccessPointInput{
			AccountID: env.accountId,
			Name:      accessPoint.Name,
		})
		require.Nil(t, err)
	}
	name := randomString(6)
	_, err = client.CreateMultiRegionAccessPoint(ctx, &tos.CreateMultiRegionAccessPointInput{
		AccountID: env.accountId,
		Name:      name,
		Regions: []tos.BucketAccount{{
			Bucket:          bucket,
			BucketAccountID: env.accountId,
		}},
	})
	require.Nil(t, err)

	resp, err = client.ListMultiRegionAccessPoints(ctx, &tos.ListMultiRegionAccessPointsInput{AccountID: env.accountId})
	require.Nil(t, err)
	require.Equal(t, len(resp.AccessPoints), 1)
	require.Equal(t, resp.AccessPoints[0].Name, name)
	require.Equal(t, resp.AccessPoints[0].Regions[0].Bucket, bucket)
	require.Equal(t, resp.AccessPoints[0].Regions[0].BucketAccountID, env.accountId)
	require.Equal(t, resp.AccessPoints[0].Regions[0].Region, env.region)
	require.True(t, len(resp.AccessPoints[0].Status) > 0)
	require.True(t, resp.AccessPoints[0].CreatedAt > 0)

	getResp, err := client.GetMultiRegionAccessPoint(ctx, &tos.GetMultiRegionAccessPointInput{
		AccountID: env.accountId,
		Name:      name,
	})
	require.Nil(t, err)
	require.Equal(t, getResp.Name, name)
	require.Equal(t, len(getResp.Regions), 1)
	require.True(t, len(getResp.Alias) > 0)
	require.Equal(t, getResp.Regions[0].Bucket, bucket)
	require.Equal(t, getResp.Regions[0].BucketAccountID, env.accountId)
	require.Equal(t, getResp.Regions[0].Region, env.region)
	require.True(t, len(getResp.Status) > 0)
	require.True(t, getResp.CreatedAt > 0)

	_, err = client.DeleteMultiRegionAccessPoint(ctx, &tos.DeleteMultiRegionAccessPointInput{
		AccountID: env.accountId,
		Name:      name,
	})
	require.Nil(t, err)

	getResp, err = client.GetMultiRegionAccessPoint(ctx, &tos.GetMultiRegionAccessPointInput{
		AccountID: env.accountId,
		Name:      name,
	})
	require.NotNil(t, err)
}
