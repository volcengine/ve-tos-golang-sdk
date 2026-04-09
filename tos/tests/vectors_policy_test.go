package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type VectorPolicy struct {
	Version   string            `json:"Version"`
	Statement []PolicyStatement `json:"Statement"`
}

type PolicyStatement struct {
	Sid       string   `json:"Sid"`
	Effect    string   `json:"Effect"`
	Principal []string `json:"Principal"`
	Action    string   `json:"Action"`
	Resource  string   `json:"Resource"`
}

func TestVectorBucketPolicyOperations(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector-policy-" + randomString(8)
		accountID        = e.accountId
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(ctx, createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, createBucketOutput.StatusCode, 200)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 构造测试policy
	testPolicy := map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect":    "Allow",
				"Principal": []string{accountID},
				"Action":    "tosvectors:GetVectorBucket",
				"Resource":  "trn:tosvectors:" + e.region + ":" + accountID + ":bucket/" + vectorBucketName,
			},
		},
	}
	policyJSON, err := json.Marshal(testPolicy)
	require.Nil(t, err)

	// 3. 测试 PutVectorBucketPolicy
	putPolicyInput := &tos.PutVectorBucketPolicyInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		Policy:           string(policyJSON),
	}
	putPolicyOutput, err := client.PutVectorBucketPolicy(ctx, putPolicyInput)
	require.Nil(t, err)
	require.NotNil(t, putPolicyOutput)
	require.Equal(t, putPolicyOutput.StatusCode, 200)
	require.NotEmpty(t, putPolicyOutput.RequestID)

	// 4. 测试 GetVectorBucketPolicy
	getPolicyInput := &tos.GetVectorBucketPolicyInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
	}
	getPolicyOutput, err := client.GetVectorBucketPolicy(ctx, getPolicyInput)
	require.Nil(t, err)
	require.NotNil(t, getPolicyOutput)
	require.Equal(t, getPolicyOutput.StatusCode, 200)
	require.NotEmpty(t, getPolicyOutput.RequestID)
	require.NotEmpty(t, getPolicyOutput.Policy)

	// 验证返回的policy符合预期
	var returnedPolicy VectorPolicy
	err = json.Unmarshal([]byte(getPolicyOutput.Policy), &returnedPolicy)
	require.Nil(t, err)
	require.True(t, len(returnedPolicy.Version) > 0)
	require.True(t, len(returnedPolicy.Statement) > 0)
	require.Equal(t, returnedPolicy.Statement[0].Effect, "Allow")
	require.True(t, len(returnedPolicy.Statement[0].Principal) > 0)
	require.True(t, len(returnedPolicy.Statement[0].Action) > 0)
	require.True(t, len(returnedPolicy.Statement[0].Resource) > 0)

	// 5. 测试 DeleteVectorBucketPolicy
	deletePolicyInput := &tos.DeleteVectorBucketPolicyInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
	}
	deletePolicyOutput, err := client.DeleteVectorBucketPolicy(ctx, deletePolicyInput)
	require.Nil(t, err)
	require.NotNil(t, deletePolicyOutput)
	require.Equal(t, deletePolicyOutput.StatusCode, 200)
	require.NotEmpty(t, deletePolicyOutput.RequestID)

	// 6. 再次 GetVectorBucketPolicy，应该返回404或空policy
	getPolicyOutput2, err := client.GetVectorBucketPolicy(ctx, getPolicyInput)
	require.NotNil(t, err)
	require.Nil(t, getPolicyOutput2)
	serr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.Equal(t, serr.StatusCode, 404)
}

func TestPutVectorBucketPolicyWithEmptyPolicy(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector-policy-empty-" + randomString(6)
		accountID        = e.accountId
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(ctx, createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, createBucketOutput.StatusCode, 200)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 测试 PutVectorBucketPolicy 空policy，应该返回错误
	putPolicyInput := &tos.PutVectorBucketPolicyInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		Policy:           "",
	}
	putPolicyOutput, err := client.PutVectorBucketPolicy(ctx, putPolicyInput)
	require.NotNil(t, err)
	require.Nil(t, putPolicyOutput)
	require.Contains(t, err.Error(), "Policy is empty")
}

func TestVectorBucketPolicyWithInvalidBucketName(t *testing.T) {
	var (
		e         = newTestEnv(t)
		client    = e.prepareVectorsClient("")
		accountID = e.accountId
		ctx       = context.Background()
	)

	// 测试无效的bucket name
	invalidBucketName := "invalid-bucket-name-!@#$%"

	putPolicyInput := &tos.PutVectorBucketPolicyInput{
		VectorBucketName: invalidBucketName,
		AccountID:        accountID,
		Policy:           `{"Version":"2012-10-17","Statement":[]}`,
	}
	putPolicyOutput, err := client.PutVectorBucketPolicy(ctx, putPolicyInput)
	require.NotNil(t, err)
	require.Nil(t, putPolicyOutput)

	getPolicyInput := &tos.GetVectorBucketPolicyInput{
		VectorBucketName: invalidBucketName,
		AccountID:        accountID,
	}
	getPolicyOutput, err := client.GetVectorBucketPolicy(ctx, getPolicyInput)
	require.NotNil(t, err)
	require.Nil(t, getPolicyOutput)

	deletePolicyInput := &tos.DeleteVectorBucketPolicyInput{
		VectorBucketName: invalidBucketName,
		AccountID:        accountID,
	}
	deletePolicyOutput, err := client.DeleteVectorBucketPolicy(ctx, deletePolicyInput)
	require.NotNil(t, err)
	require.Nil(t, deletePolicyOutput)
}

func TestVectorBucketPolicyWithNilInput(t *testing.T) {
	var (
		e      = newTestEnv(t)
		client = e.prepareVectorsClient("")
		ctx    = context.Background()
	)

	// 测试 nil input
	putPolicyOutput, err := client.PutVectorBucketPolicy(ctx, nil)
	require.NotNil(t, err)
	require.Nil(t, putPolicyOutput)
	require.Equal(t, err, tos.InputIsNilClientError)

	getPolicyOutput, err := client.GetVectorBucketPolicy(ctx, nil)
	require.NotNil(t, err)
	require.Nil(t, getPolicyOutput)
	require.Equal(t, err, tos.InputIsNilClientError)

	deletePolicyOutput, err := client.DeleteVectorBucketPolicy(ctx, nil)
	require.NotNil(t, err)
	require.Nil(t, deletePolicyOutput)
	require.Equal(t, err, tos.InputIsNilClientError)
}
