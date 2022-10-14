package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketPolicy(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("policy-v2")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	policy := map[string]interface{}{
		"Statement": []map[string]interface{}{
			{
				"Sid":       "internal public",
				"Effect":    "Allow",
				"Action":    []string{"*"},
				"Principal": "*",
				"Resource": []string{
					fmt.Sprintf("trn:tos:::%s/*", bucket),
					fmt.Sprintf("trn:tos:::%s", bucket),
				},
			},
		}}
	data, err := json.Marshal(policy)
	require.Nil(t, err)
	putRes, err := client.PutBucketPolicyV2(ctx, &tos.PutBucketPolicyV2Input{
		Bucket: bucket,
		Policy: string(data),
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketPolicyV2(ctx, &tos.GetBucketPolicyV2Input{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, getRes.Policy != "")

	deleteRes, err := client.DeleteBucketPolicyV2(ctx, &tos.DeleteBucketPolicyV2Input{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

}
