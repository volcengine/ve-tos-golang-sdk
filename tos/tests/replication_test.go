package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestReplication(t *testing.T) {
	var (
		env = newTestEnv(t)

		ctx = context.Background()
	)
	bucket := generateBucketName("replication")
	bucket2 := generateBucketName("replication")
	client := env.prepareClient(bucket)
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.TextFormatter{DisableQuote: true}
	options := []tos.ClientOption{
		tos.WithRegion(env.region2),
		tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithLogger(log),
		tos.WithMaxRetryCount(5),
	}
	client2, err := tos.NewClientV2(env.endpoint2, options...)
	require.Nil(t, err)
	_, err = client2.CreateBucketV2(ctx, &tos.CreateBucketV2Input{
		Bucket: bucket2,
	})
	require.Nil(t, err)
	defer func() {
		cleanBucket(t, client2, bucket2)
		cleanBucket(t, client, bucket)
	}()
	rule1 := tos.ReplicationRule{
		ID:        "1",
		Status:    enum.StatusEnabled,
		PrefixSet: []string{"prefix1", "prefix2"},
		Destination: tos.Destination{
			Bucket:       bucket2,
			Location:     env.region2,
			StorageClass: enum.StorageClassIa,
		},
		HistoricalObjectReplication: enum.StatusEnabled,
	}
	input := &tos.PutBucketReplicationInput{
		Bucket: bucket,
		Role:   "ServiceRoleforReplicationAccessTOS",
		Rules:  []tos.ReplicationRule{rule1},
	}
	putOutput, err := client.PutBucketReplication(ctx, input)
	require.Nil(t, err)
	t.Log("PutReplication Request ID:", putOutput.RequestID)

	getOut, err := client.GetBucketReplication(ctx, &tos.GetBucketReplicationInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.Equal(t, len(getOut.Rules), 1)
	require.Equal(t, getOut.Rules[0].HistoricalObjectReplication, rule1.HistoricalObjectReplication)
	require.Equal(t, getOut.Rules[0].ID, rule1.ID)
	require.Equal(t, getOut.Rules[0].Status, rule1.Status)
	require.Equal(t, getOut.Rules[0].PrefixSet, rule1.PrefixSet)
	require.Equal(t, getOut.Rules[0].Destination, rule1.Destination)
	t.Log("GetBucketReplication Request ID:", getOut.RequestID)

	output, err := client.DeleteBucketReplication(ctx, &tos.DeleteBucketReplicationInput{Bucket: bucket})
	require.Nil(t, err)
	t.Log("DeleteBucketReplication Request ID:", output.RequestID)

	getOut, err = client.GetBucketReplication(ctx, &tos.GetBucketReplicationInput{
		Bucket: bucket,
	})
	require.NotNil(t, err)
}

func TestTime(t *testing.T) {
	tm := 1672371186296234489 / int64(time.Millisecond)
	t.Log(tm < 1672371046165)
}
