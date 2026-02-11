package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketObjectSetConfiguration(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	pathLevel := 3
	customDelimiter := "/"
	enableDefaultObjectSet := true
	readsQps := 1000
	writesQps := 2000
	listQps := 3000
	readsRate := 4000
	writesRate := 5000

	putInput := &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              pathLevel,
		CustomDelimiter:        customDelimiter,
		EnableDefaultObjectSet: enableDefaultObjectSet,
		Qos: tos.QosConfig{
			ReadsQps:   readsQps,
			WritesQps:  writesQps,
			ListQps:    listQps,
			ReadsRate:  readsRate,
			WritesRate: writesRate,
		},
	}

	putOut, err := cli.PutBucketObjectSetConfiguration(ctx, putInput)
	require.Nil(t, err)
	require.NotNil(t, putOut)

	getOut, err := cli.GetBucketObjectSetConfiguration(ctx, &tos.GetBucketObjectSetConfigurationInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)

	require.Equal(t, pathLevel, getOut.PathLevel)
	require.Equal(t, customDelimiter, getOut.CustomDelimiter)
	require.Equal(t, enableDefaultObjectSet, getOut.EnableDefaultObjectSet)
	require.Equal(t, readsQps, getOut.Qos.ReadsQps)
	require.Equal(t, writesQps, getOut.Qos.WritesQps)
	require.Equal(t, listQps, getOut.Qos.ListQps)
	require.Equal(t, readsRate, getOut.Qos.ReadsRate)
	require.Equal(t, writesRate, getOut.Qos.WritesRate)
}

func TestBucketObjectSetConfigurationWithDefaultValues(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	pathLevel := 3

	putInput := &tos.PutBucketObjectSetConfigurationInput{
		Bucket:    bucket,
		PathLevel: pathLevel,
	}

	putOut, err := cli.PutBucketObjectSetConfiguration(ctx, putInput)
	require.Nil(t, err)
	require.NotNil(t, putOut)

	getOut, err := cli.GetBucketObjectSetConfiguration(ctx, &tos.GetBucketObjectSetConfigurationInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)

	require.Equal(t, pathLevel, getOut.PathLevel)
	require.Equal(t, getOut.EnableDefaultObjectSet, false)

	jsonV, _ := json.Marshal(getOut)
	t.Logf("%s", jsonV)

}

func TestObjectSetCRUD(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-crud")
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

	// create object set
	putOut, err := cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/c",
		TagSet: tos.TagSet{
			Tags: []tos.Tag{
				{Key: "env", Value: "dev"},
				{Key: "team", Value: "alpha"},
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)

	// get and assert
	getOut, err := cli.GetObjectSet(ctx, &tos.GetObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/c",
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, "tenant-a/b/c/", getOut.ObjectSetName)

	tags := tagsToMap(getOut.TagSet.Tags)
	require.Equal(t, "dev", tags["env"])
	require.Equal(t, "alpha", tags["team"])

	// delete
	delOut, err := cli.DeleteObjectSet(ctx, &tos.DeleteObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/c",
	})
	require.Nil(t, err)
	require.NotNil(t, delOut)

	// get expect not found
	getOut, err = cli.GetObjectSet(ctx, &tos.GetObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/c",
	})
	require.Nil(t, getOut)
	require.NotNil(t, err)
	require.Equal(t, http.StatusNotFound, tos.StatusCode(err))

	putOutWithNoTags, err := cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/no-tags-c",
	})
	require.Nil(t, err)
	require.NotNil(t, putOutWithNoTags)

	getOutWithNoTags, err := cli.GetObjectSet(ctx, &tos.GetObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "tenant-a/b/no-tags-c",
	})
	require.Nil(t, err)
	require.NotNil(t, getOutWithNoTags)
	require.Equal(t, "tenant-a/b/no-tags-c/", getOutWithNoTags.ObjectSetName)
	jsonV, _ := json.Marshal(getOutWithNoTags)
	t.Logf("%s", jsonV)
}

func TestListObjectSet(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-list")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, cli, bucket)
	}()

	// enable ObjectSet
	_, err := cli.PutBucketObjectSetConfiguration(ctx, &tos.PutBucketObjectSetConfigurationInput{
		Bucket:                 bucket,
		PathLevel:              4,
		CustomDelimiter:        "/",
		EnableDefaultObjectSet: true,
	})
	require.Nil(t, err)

	// create object sets
	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "dev/a/b/c",
		TagSet: tos.TagSet{
			Tags: []tos.Tag{
				{Key: "env", Value: "dev"},
				{Key: "owner", Value: "alice"},
			},
		},
	})
	require.Nil(t, err)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "prod/a/b/c",
		TagSet: tos.TagSet{
			Tags: []tos.Tag{
				{Key: "env", Value: "prod"},
				{Key: "owner", Value: "bob"},
			},
		},
	})
	require.Nil(t, err)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: "dev/x/y/z",
		TagSet: tos.TagSet{
			Tags: []tos.Tag{
				{Key: "env", Value: "dev"},
				{Key: "owner", Value: "charlie"},
			},
		},
	})
	require.Nil(t, err)

	// list with pagination
	firstPage, err := cli.ListObjectSet(ctx, &tos.ListObjectSetInput{
		Bucket:  bucket,
		MaxKeys: 2,
	})
	require.Nil(t, err)
	require.NotNil(t, firstPage)
	require.Equal(t, 2, len(firstPage.ObjectSets))
	require.True(t, firstPage.IsTruncated)
	require.NotEmpty(t, firstPage.NextMarker)

	names := make(map[string]struct{})
	for _, s := range firstPage.ObjectSets {
		names[s.ObjectSetName] = struct{}{}
	}

	marker := firstPage.NextMarker
	for {
		page, err := cli.ListObjectSet(ctx, &tos.ListObjectSetInput{
			Bucket:  bucket,
			Marker:  marker,
			MaxKeys: 2,
		})
		require.Nil(t, err)
		require.NotNil(t, page)
		for _, s := range page.ObjectSets {
			names[s.ObjectSetName] = struct{}{}
		}
		if !page.IsTruncated {
			break
		}
		marker = page.NextMarker
	}

	t.Log("names ===> ", names)

	// assert all object sets are listed
	require.Len(t, names, 3)
	_, ok := names["dev/a/b/c/"]
	require.True(t, ok)
	_, ok = names["prod/a/b/c/"]
	require.True(t, ok)
	_, ok = names["dev/x/y/z/"]
	require.True(t, ok)

	// list with Prefix filter
	prefixPage, err := cli.ListObjectSet(ctx, &tos.ListObjectSetInput{
		Bucket: bucket,
		Prefix: "dev/",
	})
	require.Nil(t, err)
	require.NotNil(t, prefixPage)
	require.True(t, len(prefixPage.ObjectSets) >= 1)
	require.Equal(t, 2, len(prefixPage.ObjectSets))
	for _, s := range prefixPage.ObjectSets {
		require.True(t, len(s.ObjectSetName) >= len("dev/"))
		require.Equal(t, "dev/", s.ObjectSetName[:len("dev/")])
	}

	// list with Tags filter
	tagPage, err := cli.ListObjectSet(ctx, &tos.ListObjectSetInput{
		Bucket: bucket,
		Tags:   "env=prod",
	})
	require.Nil(t, err)
	require.NotNil(t, tagPage)
	require.True(t, len(tagPage.ObjectSets) >= 1)
	for _, s := range tagPage.ObjectSets {
		tagMap := tagsToMap(s.TagSet.Tags)
		require.Equal(t, "prod", tagMap["env"])
	}
	require.Equal(t, 1, len(tagPage.ObjectSets))

	listWithDefaultOut, err := cli.ListObjectSet(ctx, &tos.ListObjectSetInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, listWithDefaultOut)
	require.Equal(t, 3, len(listWithDefaultOut.ObjectSets))
}

func TestObjectSetTagging(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-tagging")
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

	objectSetName := "test/object/set"

	// create object set
	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)

	// set tagging for object set
	expectedTagSet := tos.TagSet{
		Tags: []tos.Tag{
			{Key: "owner", Value: "zhangsan"},
			{Key: "env", Value: "dev"},
		},
	}
	putOut, err := cli.PutObjectSetTagging(ctx, &tos.PutObjectSetTaggingInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
		TagSet:        expectedTagSet,
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)

	// head bucket to ensure bucket is accessible
	headOut, err := cli.HeadBucket(ctx, &tos.HeadBucketInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.NotNil(t, headOut)

	// get object set tagging and assert
	getOut, err := cli.GetObjectSetTagging(ctx, &tos.GetObjectSetInput{
		Bucket:        bucket,
		ObjectSetName: objectSetName,
	})
	require.Nil(t, err)
	require.NotNil(t, getOut)

	require.Equal(t, objectSetName+"/", getOut.ObjectSetName)
	tagMap := tagsToMap(getOut.TagSet.Tags)
	require.Equal(t, "zhangsan", tagMap["owner"])
	require.Equal(t, "dev", tagMap["env"])
}

func tagsToMap(tags []tos.Tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, tag := range tags {
		m[tag.Key] = tag.Value
	}
	return m
}
