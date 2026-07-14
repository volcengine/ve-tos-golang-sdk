package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestAccelerator(t *testing.T) {
	var (
		env    = newTestEnv(t)
		client = env.prepareClient("")
		ctx    = context.Background()
	)
	if env.acceleratorAZ == "" {
		t.Skip("skip accelerator integration test: missing TOS_GO_SDK_ACCELERATOR_AZ")
	}

	acceleratorName := acceleratorTestName(env.region, env.acceleratorAZ)
	deleted := false
	putResp, err := client.PutAccelerator(ctx, &tos.PutAcceleratorInput{
		AccountID:       env.accountId,
		AcceleratorName: acceleratorName,
		Region:          env.region,
		AZ:              env.acceleratorAZ,
		Type:            enum.AcceleratorTypeRO,
		CacheCapacity: tos.AcceleratorCacheCapacity{
			Value: 50,
			Unit:  "GiB",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, putResp.RequestID)
	require.NotEmpty(t, putResp.AcceleratorID)
	t.Log("PutAccelerator Request ID:", putResp.RequestID)
	defer func() {
		if !deleted {
			_, _ = client.DeleteAccelerator(context.Background(), &tos.DeleteAcceleratorInput{
				AccountID:     env.accountId,
				AcceleratorID: putResp.AcceleratorID,
			})
		}
	}()

	getResp, err := client.GetAccelerator(ctx, &tos.GetAcceleratorInput{
		AccountID:     env.accountId,
		AcceleratorID: putResp.AcceleratorID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, getResp.RequestID)
	require.Equal(t, putResp.AcceleratorID, getResp.AcceleratorID)
	require.Equal(t, acceleratorName, getResp.AcceleratorName)
	require.Equal(t, env.accountId, getResp.Account)
	require.Equal(t, env.region, getResp.Region)
	require.Equal(t, env.acceleratorAZ, getResp.AZ)
	require.NotZero(t, getResp.CreateTime)
	require.Equal(t, int64(50), getResp.CacheCapacity.Value)
	require.Equal(t, "GiB", getResp.CacheCapacity.Unit)
	t.Log("GetAccelerator Request ID:", getResp.RequestID)

	listResp, err := client.ListAccelerator(ctx, &tos.ListAcceleratorInput{
		AccountID:  env.accountId,
		MaxResults: 10,
	})
	require.NoError(t, err)
	require.NotEmpty(t, listResp.RequestID)
	accelerator := acceleratorByID(listResp.Accelerators, putResp.AcceleratorID)
	require.NotNil(t, accelerator)
	require.Equal(t, acceleratorName, accelerator.AcceleratorName)
	require.Equal(t, env.accountId, accelerator.Account)
	require.Equal(t, env.region, accelerator.Region)
	require.Equal(t, env.acceleratorAZ, accelerator.AZ)
	require.Equal(t, int64(50), accelerator.CacheCapacity.Value)
	require.Equal(t, "GiB", accelerator.CacheCapacity.Unit)
	t.Log("ListAccelerator Request ID:", listResp.RequestID)

	deleteResp, err := client.DeleteAccelerator(ctx, &tos.DeleteAcceleratorInput{
		AccountID:     env.accountId,
		AcceleratorID: putResp.AcceleratorID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, deleteResp.RequestID)
	deleted = true
	t.Log("DeleteAccelerator Request ID:", deleteResp.RequestID)

	deleteResp, err = client.DeleteAccelerator(ctx, &tos.DeleteAcceleratorInput{
		AccountID:     env.accountId,
		AcceleratorID: putResp.AcceleratorID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, deleteResp.RequestID)
	t.Log("DeleteAccelerator Idempotent Request ID:", deleteResp.RequestID)
}

func acceleratorTestName(region, az string) string {
	name := "go-sdk-test-acc_" + region + "-" + az
	return strings.ToLower(name)
}

func acceleratorByID(accelerators []tos.Accelerator, id string) *tos.Accelerator {
	for i := range accelerators {
		if accelerators[i].AcceleratorID == id {
			return &accelerators[i]
		}
	}
	return nil
}
