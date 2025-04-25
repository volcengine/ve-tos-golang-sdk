package tests

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type TestQuota struct {
	WritesQps  string `json:"WritesQps"`
	WritesRate string `json:"WritesRate"`
}

type TestStatement struct {
	Sid       string    `json:"Sid"`
	Quota     TestQuota `json:"Quota"`
	Principal []string  `json:"Principal"`
	Resource  string    `json:"Resource"`
}

type TestPolicy struct {
	Version   string          `json:"Version"`
	Statement []TestStatement `json:"Statement"`
}

func TestQosPolicy(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("qos-policy")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()
	_, err := cli.DeleteQosPolicy(ctx, &tos.DeleteQosPolicyInput{AccountID: env.accountId})
	require.Nil(t, err)

	tmpQosPolicy := &TestPolicy{
		Version: "2012-08-01",
		Statement: []TestStatement{{
			Sid: "gosdk-qospolicy-123",
			Quota: TestQuota{
				WritesQps:  "10",
				WritesRate: "10000",
			},
			Principal: []string{"trn:iam::*"},
			Resource:  "trn:tos:::gosdkqospolicynousebucket/*",
		}},
	}
	tmpData, err := json.Marshal(tmpQosPolicy)
	require.Nil(t, err)
	_, err = cli.PutQosPolicy(ctx, &tos.PutQosPolicyInput{
		AccountID: env.accountId,
		Policy:    string(tmpData),
	})
	require.Nil(t, err)

	qosOut, err := cli.GetQosPolicy(ctx, &tos.GetQosPolicyInput{
		AccountID: env.accountId,
	})
	require.Nil(t, err)
	resQospolicy := &TestPolicy{}
	err = json.Unmarshal([]byte(qosOut.Policy), resQospolicy)
	require.Nil(t, err)
	require.Equal(t, tmpQosPolicy.Statement[0].Quota.WritesRate, resQospolicy.Statement[0].Quota.WritesRate)
	require.Equal(t, tmpQosPolicy.Statement[0].Quota.WritesQps, resQospolicy.Statement[0].Quota.WritesQps)
	_, err = cli.DeleteQosPolicy(ctx, &tos.DeleteQosPolicyInput{AccountID: env.accountId})
	require.Nil(t, err)
}
