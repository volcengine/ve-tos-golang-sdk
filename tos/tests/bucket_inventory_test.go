package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"testing"
	"time"
)

func TestBucketInventory(t *testing.T) {

	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("inventory-bucket")
		client = env.prepareClient(bucket, tos.WithSocketTimeout(360*time.Second, 360*time.Second), tos.WithRequestTimeout(360*time.Second))
		ctx    = context.Background()
	)
	defer func() { cleanBucket(t, client, bucket) }()

	outResp, err := client.PutBucketInventory(ctx, &tos.PutBucketInventoryInput{
		Bucket: bucket,
		ID:     "inventory-id",
		Destination: &tos.InventoryDestination{
			TOSBucketDestination: &tos.TOSBucketDestination{
				Format:    enum.InventoryFormatCsv,
				AccountID: env.accountId,
				Role:      env.mqRoleName,
				Bucket:    bucket,
				Prefix:    "prefix",
			},
		},
		Filter: &tos.InventoryFilter{
			Prefix: "prefix",
		},
		Schedule: &tos.InventorySchedule{
			Frequency: enum.InventoryFrequencyTypeDaily,
		},
		IncludedObjectVersions: enum.InventoryIncludedObjTypeAll,
		OptionalFields: &tos.InventoryOptionalFields{
			Field: []string{"LastModifiedDate", "Size", "ETag"},
		},
		IsUnCompressed: true,
		IsEnabled:      true,
	})

	require.Nil(t, err)
	require.Equal(t, 200, outResp.StatusCode)
	getOutResp, err := client.GetBucketInventory(ctx, &tos.GetBucketInventoryInput{
		Bucket: bucket,
		ID:     "inventory-id",
	})
	require.Nil(t, err)
	require.Equal(t, 200, outResp.StatusCode)
	require.Equal(t, "inventory-id", getOutResp.BucketInventoryConfiguration.ID)
	require.Equal(t, true, getOutResp.BucketInventoryConfiguration.IsEnabled)
	require.Equal(t, "prefix", getOutResp.BucketInventoryConfiguration.Filter.Prefix)
	require.Equal(t, enum.InventoryFormatCsv, getOutResp.BucketInventoryConfiguration.Destination.TOSBucketDestination.Format)
	require.Equal(t, env.accountId, getOutResp.BucketInventoryConfiguration.Destination.TOSBucketDestination.AccountID)
	require.Equal(t, env.mqRoleName, getOutResp.BucketInventoryConfiguration.Destination.TOSBucketDestination.Role)
	require.Equal(t, bucket, getOutResp.BucketInventoryConfiguration.Destination.TOSBucketDestination.Bucket)
	require.Equal(t, "prefix", getOutResp.BucketInventoryConfiguration.Destination.TOSBucketDestination.Prefix)
	require.Equal(t, enum.InventoryFrequencyTypeDaily, getOutResp.BucketInventoryConfiguration.Schedule.Frequency)
	require.Equal(t, enum.InventoryIncludedObjTypeAll, getOutResp.BucketInventoryConfiguration.IncludedObjectVersions)
	require.Equal(t, true, getOutResp.BucketInventoryConfiguration.IsUnCompressed)

	listOut, err := client.ListBucketInventory(ctx, &tos.ListBucketInventoryInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.Equal(t, 200, outResp.StatusCode)
	require.Equal(t, 1, len(listOut.Configurations))
	require.Equal(t, "inventory-id", listOut.Configurations[0].ID)
	require.Equal(t, true, listOut.Configurations[0].IsEnabled)
	require.Equal(t, "prefix", listOut.Configurations[0].Filter.Prefix)
	require.Equal(t, enum.InventoryFormatCsv, listOut.Configurations[0].Destination.TOSBucketDestination.Format)
	require.Equal(t, env.accountId, listOut.Configurations[0].Destination.TOSBucketDestination.AccountID)
	require.Equal(t, env.mqRoleName, listOut.Configurations[0].Destination.TOSBucketDestination.Role)
	require.Equal(t, bucket, listOut.Configurations[0].Destination.TOSBucketDestination.Bucket)
	require.Equal(t, "prefix", listOut.Configurations[0].Destination.TOSBucketDestination.Prefix)
	require.Equal(t, enum.InventoryFrequencyTypeDaily, listOut.Configurations[0].Schedule.Frequency)
	require.Equal(t, enum.InventoryIncludedObjTypeAll, listOut.Configurations[0].IncludedObjectVersions)
	require.Equal(t, true, listOut.Configurations[0].IsUnCompressed)
	_, err = client.DeleteBucketInventory(ctx, &tos.DeleteBucketInventoryInput{
		Bucket: bucket,
		ID:     "inventory-id",
	})
	require.Nil(t, err)

	listOut, err = client.ListBucketInventory(ctx, &tos.ListBucketInventoryInput{
		Bucket: bucket,
	})
	require.NotNil(t, err)
	require.Nil(t, listOut)

}
