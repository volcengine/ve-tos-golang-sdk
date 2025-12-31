package tests

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
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
			Domain:   bucket + ".volcengine.com",
			Protocol: enum.AuthProtocolS3,
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
	require.Equal(t, rule.Protocol, enum.AuthProtocolS3)
	_, err = client.DeleteBucketCustomDomain(ctx, &tos.DeleteBucketCustomDomainInput{Bucket: bucket, Domain: bucket + ".volcengine.com"})
	require.Nil(t, err)

}

func TestCustomDomain(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket1 = generateBucketName("custom-domain")
		bucket2 = generateBucketName("custom-domain")
		ctx     = context.Background()
		c       = env.prepareClient(bucket1)
		key     = randomString(6)
	)
	_, err := c.CreateBucketV2(ctx, &tos.CreateBucketV2Input{
		Bucket: bucket2,
	})
	require.Nil(t, err)
	defer func() {
		cleanBucket(t, c, bucket1)
		cleanBucket(t, c, bucket2)
	}()
	endpoint := strings.Replace(strings.Replace(env.endpoint, "http://", "", 1), "https://", "", 1)
	client, err := tos.NewClientV2(bucket2+"."+endpoint, tos.WithRegion(env.region),
		tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)),
		tos.WithCustomDomain(true))
	require.Nil(t, err)

	put, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket1, Key: key},
		Content:             strings.NewReader(randomString(1024)),
	})
	require.Nil(t, err)
	require.Equal(t, put.StatusCode, http.StatusOK)
	head, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: bucket1,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, head.StatusCode, http.StatusOK)

	get, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket1,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, get.StatusCode, http.StatusOK)
	defer get.Content.Close()

	headNotExist, err := c.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: bucket1, Key: key})
	require.NotNil(t, err)
	t.Log(err)
	require.Nil(t, headNotExist)

	putacl, err := c.PutBucketACL(context.Background(), &tos.PutBucketACLInput{Bucket: bucket1, ACLType: enum.ACLPublicRead})
	require.Nil(t, err)
	require.Equal(t, putacl.StatusCode, http.StatusOK)
}
