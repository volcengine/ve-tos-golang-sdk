package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketCustomDomain(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("custom-domain")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	input := &tos.PutBucketCustomDomainInput{
		Bucket: bucket,
		Rule: tos.CustomDomainRule{
			Domain: bucket + ".volcengine.com",
		},
	}
	output, err := client.PutBucketCustomDomain(ctx, input)
	require.Nil(t, err)
	t.Log(output)
	listOutput, err := client.ListBucketCustomDomain(ctx, &tos.ListBucketCustomDomainInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, len(listOutput.Rules), 1)
	rule := listOutput.Rules[0]
	require.Equal(t, rule.Domain, bucket+".volcengine.com")
	_, err = client.DeleteBucketCustomDomain(ctx, &tos.DeleteBucketCustomDomainInput{Bucket: bucket, Domain: bucket + ".volcengine.com"})
	require.Nil(t, err)

}
