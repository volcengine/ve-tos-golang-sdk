package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestObjectSetQuota(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-quota")
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
	waitUntilObjectSetReady(t, bucket, cli)

	// create object sets
	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{Bucket: bucket, ObjectSetName: "a/b/c"})
	require.Nil(t, err)
	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{Bucket: bucket, ObjectSetName: "d/e/f"})
	require.Nil(t, err)

	// scenario 1: set and verify quota
	putOut, err := cli.PutObjectSetQuota(ctx, &tos.PutObjectSetQuotaInput{
		Bucket:        bucket,
		ObjectSetName: "a/b/c",
		StorageQuota:  "1048576",
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)
	require.Equal(t, 200, putOut.StatusCode)

	getOut, err := cli.GetObjectSetQuota(ctx, &tos.GetObjectSetQuotaInput{Bucket: bucket, ObjectSetName: "a/b/c"})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, 200, getOut.StatusCode)
	require.Equal(t, "1048576", getOut.StorageQuota)

	// scenario 2: quota not set
	getOut, err = cli.GetObjectSetQuota(ctx, &tos.GetObjectSetQuotaInput{Bucket: bucket, ObjectSetName: "d/e/f"})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, 200, getOut.StatusCode)
	require.Equal(t, "", getOut.StorageQuota)

	// scenario 3: update quota
	putOut, err = cli.PutObjectSetQuota(ctx, &tos.PutObjectSetQuotaInput{
		Bucket:        bucket,
		ObjectSetName: "a/b/c",
		StorageQuota:  "0",
	})
	require.Nil(t, err)
	require.NotNil(t, putOut)
	require.Equal(t, 200, putOut.StatusCode)

	getOut, err = cli.GetObjectSetQuota(ctx, &tos.GetObjectSetQuotaInput{Bucket: bucket, ObjectSetName: "a/b/c"})
	require.Nil(t, err)
	require.NotNil(t, getOut)
	require.Equal(t, 200, getOut.StatusCode)
	require.Equal(t, "0", getOut.StorageQuota)
}

func TestObjectSetStorage(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("objectset-storage")
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
	waitUntilObjectSetReady(t, bucket, cli)

	_, err = cli.PutObjectSet(ctx, &tos.PutObjectSetInput{Bucket: bucket, ObjectSetName: "a/b/c"})
	require.Nil(t, err)

	putRandomObject(t, cli, bucket, "a/b/c/file1.txt", 16)
	putRandomObject(t, cli, bucket, "a/b/c/file2.txt", 32)

	out, err := cli.GetObjectSetStorage(ctx, &tos.GetObjectSetStorageInput{Bucket: bucket, ObjectSetName: "a/b/c"})
	require.Nil(t, err)
	require.NotNil(t, out)
	require.Equal(t, 200, out.StatusCode)
	// 约束：GetObjectSetStorage 返回的 StorageStat.Storage 恒为 "0"
	require.Equal(t, "0", out.TotalStorageStat.Storage)
	require.Equal(t, "0", out.StandardStorageStat.Storage)
	require.Equal(t, "0", out.IAStorageStat.Storage)
	require.Equal(t, "0", out.ArchiveFrStorageStat.Storage)
	require.Equal(t, "0", out.ArchiveStorageStat.Storage)
	require.Equal(t, "0", out.ColdArchiveStat.Storage)
	require.Equal(t, "0", out.DeepColdArchiveStorageStat.Storage)
	require.Equal(t, "0", out.IntelligentTieringStorageStats.HighFreqStorageStat.Storage)
	require.Equal(t, "0", out.IntelligentTieringStorageStats.LowFreqStorageStat.Storage)
	require.Equal(t, "0", out.IntelligentTieringStorageStats.ArchiveStorageStat.Storage)
}
