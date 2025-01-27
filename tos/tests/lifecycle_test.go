package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestBucketLifecycle(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("lifecycle")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	rule := tos.LifecycleRule{
		ID:     "1",
		Prefix: "test",
		Status: enum.LifecycleStatusEnabled,
		Transitions: []tos.Transition{{
			Date:         time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			StorageClass: enum.StorageClassIa,
		}},
		Expiration: &tos.Expiration{
			Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		NonCurrentVersionTransition: []tos.NonCurrentVersionTransition{{
			NonCurrentDays: 30,
			StorageClass:   enum.StorageClassIa,
		}},
		NoCurrentVersionExpiration: &tos.NoCurrentVersionExpiration{NoCurrentDays: 70},
		Tag: []tos.Tag{{
			Key:   "1",
			Value: "2",
		}},
		Filter: &tos.LifecycleRuleFilter{
			ObjectSizeGreaterThan:   1000,
			GreaterThanIncludeEqual: enum.StatusEnabled,
			ObjectSizeLessThan:      2000,
			LessThanIncludeEqual:    enum.StatusEnabled,
		},
	}

	ctx := context.Background()
	putRes, err := client.PutBucketLifecycle(ctx, &tos.PutBucketLifecycleInput{
		Bucket: bucket,
		Rules:  []tos.LifecycleRule{rule},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketLifecycle(ctx, &tos.GetBucketLifecycleInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getRes.Rules) == 1)
	require.Equal(t, getRes.Rules[0], rule)

	deleteRes, err := client.DeleteBucketLifecycle(ctx, &tos.DeleteBucketLifecycleInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetBucketLifecycle(ctx, &tos.GetBucketLifecycleInput{Bucket: bucket})
	require.NotNil(t, err)
	nonCurrentDate := time.Date(2099, 11, 31, 0, 0, 0, 0, time.UTC)
	noCurrentVersionExpiration := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	rule = tos.LifecycleRule{
		ID:     "1",
		Prefix: "test",
		Status: enum.LifecycleStatusEnabled,
		Transitions: []tos.Transition{{
			Date:         time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			StorageClass: enum.StorageClassIa,
		}},
		Expiration: &tos.Expiration{
			Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		NonCurrentVersionTransition: []tos.NonCurrentVersionTransition{{
			NonCurrentDate: &nonCurrentDate,
			StorageClass:   enum.StorageClassIa,
		}},
		NoCurrentVersionExpiration: &tos.NoCurrentVersionExpiration{NonCurrentDate: &noCurrentVersionExpiration},
		Tag: []tos.Tag{{
			Key:   "1",
			Value: "2",
		}},
		Filter: &tos.LifecycleRuleFilter{
			ObjectSizeGreaterThan:   1000,
			GreaterThanIncludeEqual: enum.StatusEnabled,
			ObjectSizeLessThan:      2000,
			LessThanIncludeEqual:    enum.StatusEnabled,
		},
	}

	putRes, err = client.PutBucketLifecycle(ctx, &tos.PutBucketLifecycleInput{
		Bucket: bucket,
		Rules:  []tos.LifecycleRule{rule},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err = client.GetBucketLifecycle(ctx, &tos.GetBucketLifecycleInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getRes.Rules) == 1)
	require.Equal(t, getRes.Rules[0], rule)

}

func TestBucketLifeCycleOverlap(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("lifecycle")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()

	rule1 := tos.LifecycleRule{
		ID:     "1",
		Prefix: "test/",
		Status: enum.LifecycleStatusEnabled,
		Transitions: []tos.Transition{{
			Date:         time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			StorageClass: enum.StorageClassIa,
		}},
		Expiration: &tos.Expiration{
			Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		NonCurrentVersionTransition: []tos.NonCurrentVersionTransition{{
			NonCurrentDays: 30,
			StorageClass:   enum.StorageClassIa,
		}},
		NoCurrentVersionExpiration: &tos.NoCurrentVersionExpiration{NoCurrentDays: 70},
	}
	rule2 := tos.LifecycleRule{
		ID:     "2",
		Prefix: "test/an",
		Status: enum.LifecycleStatusEnabled,
		Transitions: []tos.Transition{{
			Date:         time.Date(2022, 12, 31, 0, 0, 0, 0, time.UTC),
			StorageClass: enum.StorageClassIa,
		}},
		Expiration: &tos.Expiration{
			Date: time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		NonCurrentVersionTransition: []tos.NonCurrentVersionTransition{{
			NonCurrentDays: 30,
			StorageClass:   enum.StorageClassIa,
		}},
		NoCurrentVersionExpiration: &tos.NoCurrentVersionExpiration{NoCurrentDays: 70},
	}

	ctx := context.Background()
	_, err := client.PutBucketLifecycle(ctx, &tos.PutBucketLifecycleInput{
		Bucket: bucket,
		Rules:  []tos.LifecycleRule{rule1, rule2},
	})
	require.NotNil(t, err)
	t.Log(err.Error())

	_, err = client.PutBucketLifecycle(ctx, &tos.PutBucketLifecycleInput{
		Bucket:                 bucket,
		Rules:                  []tos.LifecycleRule{rule1, rule2},
		AllowSameActionOverlap: true,
	})
	require.Nil(t, err)
}
