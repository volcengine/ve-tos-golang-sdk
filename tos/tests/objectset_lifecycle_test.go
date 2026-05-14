package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestObjectSetLifecycle(t *testing.T) {
	var (
		env           = newTestEnv(t)
		bucket        = generateBucketName("objectset-lifecycle")
		cli           = env.prepareClient(bucket)
		ctx           = context.Background()
		objectSetName = "life/cycle/set"
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	_, err := cli.PutBucketObjectSetConfiguration(ctx, &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              3,
		CustomDelimiter:        "/",
		EnableDefaultObjectSet: true,
	})
	require.Nil(t, err)
	waitUntilObjectSetReady(t, bucket, cli)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)

	rule := tos.LifecycleRule{
		ID:     "rule1",
		Prefix: objectSetName + "/",
		Status: enum.LifecycleStatusEnabled,
		Expiration: &tos.Expiration{
			Days: 7,
		},
	}

	putOut, err := cli.PutObjectSetLifecycle(ctx, &tos.PutObjectSetLifecycleInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
		Rules:         []tos.LifecycleRule{rule},
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)
	require.Equal(t, http.StatusOK, putOut.StatusCode)

	getOut, err := cli.GetObjectSetLifecycle(ctx, &tos.GetObjectSetLifecycleInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, 1, len(getOut.Rules))
	require.Equal(t, rule, getOut.Rules[0])

	delOut, err := cli.DeleteObjectSetLifecycle(ctx, &tos.DeleteObjectSetLifecycleInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)
	require.NotNil(t, delOut)
	require.Equal(t, http.StatusNoContent, delOut.StatusCode)

	getOut, err = cli.GetObjectSetLifecycle(ctx, &tos.GetObjectSetLifecycleInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.NotNil(t, err)
	require.Nil(t, getOut)
	require.Equal(t, http.StatusNotFound, tos.StatusCode(err))
}

func TestObjectSetLifecycleByTag(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-lifecycle-bytag")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	_, err := cli.PutBucketObjectSetConfiguration(ctx, &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              3,
		CustomDelimiter:        "/",
		EnableDefaultObjectSet: true,
	})
	require.Nil(t, err)
	waitUntilObjectSetReady(t, bucket, cli)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tag/a/b",
		TagSet:        tos.TagSet{Tags: []tos.Tag{{Key: "key1", Value: "value1"}}},
	})
	require.Nil(t, err)

	tagRule := tos.ObjectSetTagLifecycleRule{
		Tag: tos.Tag{Key: "key1", Value: "value1"},
		Rules: []tos.LifecycleRule{{
			ID:     "rule1",
			Prefix: "prefix",
			Status: enum.LifecycleStatusEnabled,
			Expiration: &tos.Expiration{
				Days: 1,
			},
		}},
	}

	putOut, err := cli.PutObjectSetLifecycleByTag(ctx, &tos.PutObjectSetLifecycleByTagInput{
		Bucket:            bucket,
		ObjectSetTagRules: []tos.ObjectSetTagLifecycleRule{tagRule},
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)
	require.Equal(t, http.StatusOK, putOut.StatusCode)

	getOut, err := cli.GetObjectSetLifecycleByTag(ctx, &tos.GetObjectSetLifecycleByTagInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, 1, len(getOut.ObjectSetTagRules))
	require.Equal(t, tagRule, getOut.ObjectSetTagRules[0])

	delOut, err := cli.DeleteObjectSetLifecycleByTag(ctx, &tos.DeleteObjectSetLifecycleByTagInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, delOut)
	require.Equal(t, http.StatusNoContent, delOut.StatusCode)

	getOut, err = cli.GetObjectSetLifecycleByTag(ctx, &tos.GetObjectSetLifecycleByTagInput{Bucket: bucket})
	require.NotNil(t, err)
	require.Nil(t, getOut)
	require.Equal(t, http.StatusNotFound, tos.StatusCode(err))
}

func TestObjectSetEndpoint(t *testing.T) {
	var (
		env           = newTestEnv(t)
		bucket        = generateBucketName("objectset-endpoint")
		cli           = env.prepareClient(bucket)
		ctx           = context.Background()
		objectSetName = "ep/a/b"
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	_, err := cli.PutBucketObjectSetConfiguration(ctx, &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              3,
		CustomDelimiter:        "/",
		EnableDefaultObjectSet: true,
	})
	require.Nil(t, err)
	waitUntilObjectSetReady(t, bucket, cli)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)

	out, err := cli.GetObjectSetEndpoint(ctx, &tos.GetObjectSetEndpointInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName + "/",
	})
	if err == nil {
		require.NotNil(t, out)
		require.GreaterOrEqual(t, len(out.Endpoints), 1)
		return
	}

	endpoint := out.Endpoints[0]
	require.NotNil(t, endpoint)
	require.NotEmpty(t, endpoint.Endpoint)
	require.NotEmpty(t, endpoint.CapName)
	require.NotEmpty(t, endpoint.S3Endpoint)

	code := tos.StatusCode(err)
	require.True(t, code == http.StatusNotFound || code == http.StatusForbidden || code == http.StatusBadRequest)
}
