package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestObjectSetQuotaByTag(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-quota-by-tag")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	// enable ObjectSet
	_, err := cli.PutBucketObjectSetConfiguration(ctx, &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              3,
		CustomDelimiter:        "/",
		EnableDefaultObjectSet: true,
	})
	require.Nil(t, err)

	// put quota by tag with two rules
	ruleA := tos.ObjectSetQuotaRule{
		Tag: tos.Tag{
			Key:   "owner",
			Value: "userA",
		},
		Qos: tos.QosConfig{
			ReadsQps:   100,
			WritesQps:  200,
			ListQps:    50,
			ReadsRate:  10,
			WritesRate: 20,
		},
	}
	ruleB := tos.ObjectSetQuotaRule{
		Tag: tos.Tag{
			Key:   "owner",
			Value: "userB",
		},
		Qos: tos.QosConfig{
			ReadsQps:   300,
			WritesQps:  400,
			ListQps:    150,
			ReadsRate:  30,
			WritesRate: 40,
		},
	}

	putOut, err := cli.PutObjectSetQuotaByTag(ctx, &tos.PutObjectSetQuotaByTagInput{
		Bucket: bucket,
		Rules:  []tos.ObjectSetQuotaRule{ruleA, ruleB},
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)

	// get and verify
	getOut, err := cli.GetObjectSetQuotaByTag(ctx, &tos.GetObjectSetQuotaByTagInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)

	require.Len(t, getOut.Rules, 2)


	rulesByOwner := make(map[string]tos.QosConfig)
	for _, r := range getOut.Rules {
		if r.Tag.Key == "owner" {
			rulesByOwner[r.Tag.Value] = r.Qos
		}
	}

	qos, ok := rulesByOwner["userA"]
	require.True(t, ok)
	require.Equal(t, ruleA.Qos, qos)

	qos, ok = rulesByOwner["userB"]
	require.True(t, ok)
	require.Equal(t, ruleB.Qos, qos)

	require.Equal(t, []tos.ObjectSetQuotaRule{ruleA, ruleB}, getOut.Rules)

	// delete quota rules
	delOut, err := cli.DeleteObjectSetQuotaByTag(ctx, &tos.DeleteObjectSetQuotaByTagInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, delOut)

	// get again expect not found
	getOut, err = cli.GetObjectSetQuotaByTag(ctx, &tos.GetObjectSetQuotaByTagInput{
		Bucket: bucket,
	})
	require.Nil(t, getOut)
	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, tos.StatusCode(err))
}
