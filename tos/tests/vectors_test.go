package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestVectorsBucketCrud(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector-bucket-" + randomString(8)
		accountID        = e.accountId
	)

	// 1. 创建向量桶
	createInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createOutput, err := client.CreateVectorBucket(context.Background(), createInput)
	require.Nil(t, err)
	require.NotNil(t, createOutput)
	require.NotEmpty(t, createOutput.RequestID)
	require.Equal(t, createOutput.StatusCode, 200)

	// 2. 获取向量桶信息
	getInput := &tos.GetVectorBucketInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
	}
	getOutput, err := client.GetVectorBucket(context.Background(), getInput)
	require.Nil(t, err)
	require.NotNil(t, getOutput)
	require.Equal(t, getOutput.StatusCode, 200)
	require.NotEmpty(t, getOutput.RequestID)
	require.Equal(t, vectorBucketName, getOutput.VectorBucket.VectorBucketName)
	require.NotEmpty(t, getOutput.VectorBucket.VectorBucketTrn)
	require.Equal(t, "default", getOutput.VectorBucket.ProjectName)
	require.NotZero(t, getOutput.VectorBucket.CreationTime)

	// 3. 删除向量桶
	deleteInput := &tos.DeleteVectorBucketInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
	}
	deleteOutput, err := client.DeleteVectorBucket(context.Background(), deleteInput)
	require.Nil(t, err)
	require.NotNil(t, deleteOutput)
	require.Equal(t, deleteOutput.StatusCode, 200)
	require.NotEmpty(t, deleteOutput.RequestID)

	// 4. 再次获取向量桶信息，应该返回404
	getOutput2, err := client.GetVectorBucket(context.Background(), getInput)
	require.NotNil(t, err)
	require.Nil(t, getOutput2)
	serr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.Equal(t, serr.StatusCode, 404)
}

func TestListVectorBuckets(t *testing.T) {
	var (
		e                      = newTestEnv(t)
		client                 = e.prepareVectorsClient("")
		vectorBucketNamePrefix = "test-list-bucket-" + randomString(8)
		accountID              = e.accountId
	)

	// 创建多个测试向量桶
	bucketNames := make([]string, 3)
	for i := 0; i < 3; i++ {
		bucketNames[i] = vectorBucketNamePrefix + "-" + randomString(4)
		createInput := &tos.CreateVectorBucketInput{
			VectorBucketName: bucketNames[i],
		}
		createOutput, err := client.CreateVectorBucket(context.Background(), createInput)
		require.Nil(t, err)
		require.NotNil(t, createOutput)
		require.Equal(t, createOutput.StatusCode, 200)
	}

	defer func() {
		// 清理测试向量桶
		for _, bucketName := range bucketNames {
			deleteInput := &tos.DeleteVectorBucketInput{
				VectorBucketName: bucketName,
				AccountID:        accountID,
			}
			client.DeleteVectorBucket(context.Background(), deleteInput)
		}
	}()

	// 测试列举所有向量桶
	listInput := &tos.ListVectorBucketsInput{}
	listOutput, err := client.ListVectorBuckets(context.Background(), listInput)
	require.Nil(t, err)
	require.NotNil(t, listOutput)
	require.Equal(t, listOutput.StatusCode, 200)
	require.NotEmpty(t, listOutput.RequestID)
	require.NotNil(t, listOutput.VectorBuckets)

	// 验证创建的桶在列表中
	foundBuckets := make(map[string]bool)
	for _, bucket := range listOutput.VectorBuckets {
		foundBuckets[bucket.VectorBucketName] = true
	}

	for _, bucketName := range bucketNames {
		require.True(t, foundBuckets[bucketName], "Bucket %s should be found in list", bucketName)
	}

	// 测试使用前缀过滤
	listInputWithPrefix := &tos.ListVectorBucketsInput{
		Prefix: vectorBucketNamePrefix,
	}
	listOutputWithPrefix, err := client.ListVectorBuckets(context.Background(), listInputWithPrefix)
	require.Nil(t, err)
	require.NotNil(t, listOutputWithPrefix)
	require.Equal(t, listOutputWithPrefix.StatusCode, 200)

	// 验证前缀过滤生效
	for _, bucket := range listOutputWithPrefix.VectorBuckets {
		require.True(t, strings.HasPrefix(bucket.VectorBucketName, vectorBucketNamePrefix),
			"Bucket %s should start with prefix %s", bucket.VectorBucketName, vectorBucketNamePrefix)
	}

	// 测试分页功能
	listInputWithLimit := &tos.ListVectorBucketsInput{
		MaxResults: 2,
	}
	listOutputWithLimit, err := client.ListVectorBuckets(context.Background(), listInputWithLimit)
	require.Nil(t, err)
	require.NotNil(t, listOutputWithLimit)
	require.Equal(t, listOutputWithLimit.StatusCode, 200)
	require.LessOrEqual(t, len(listOutputWithLimit.VectorBuckets), 2, "Should return at most 2 buckets")
	require.NotEmpty(t, listOutputWithLimit.NextToken, "NextToken should be present for pagination")

	nextOutput, err := client.ListVectorBuckets(context.Background(), &tos.ListVectorBucketsInput{
		MaxResults: 2,
		NextToken:  listOutputWithLimit.NextToken,
	})
	require.Nil(t, err)

	for _, bucket := range nextOutput.VectorBuckets {
		for _, bucket2 := range listOutputWithLimit.VectorBuckets {
			require.NotEqual(t, bucket.VectorBucketName, bucket2.VectorBucketName,
				"Bucket %s should not be in both lists", bucket.VectorBucketName)
		}
	}
}

func TestVectorsIndexCrud(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test" + randomString(8)
		accountID        = e.accountId
		indexName        = "test-index-" + randomString(8)
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(context.Background(), createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, createBucketOutput.StatusCode, 200)

	defer func() {
		// 清理：删除向量桶
		deleteBucketInput := &tos.DeleteVectorBucketInput{
			VectorBucketName: vectorBucketName,
			AccountID:        accountID,
		}
		client.DeleteVectorBucket(context.Background(), deleteBucketInput)
	}()

	// 2. 创建向量索引
	createIndexInput := &tos.CreateIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		DataType:         enum.DataTypeFloat32,
		Dimension:        128,
		DistanceMetric:   enum.DistanceMetricEuclidean,
		MetadataConfiguration: tos.MetadataConfiguration{
			NonFilterableMetadataKeys: []string{"timestamp", "source"},
		},
	}
	createIndexOutput, err := client.CreateIndex(context.Background(), createIndexInput)
	require.Nil(t, err)
	require.NotNil(t, createIndexOutput)
	require.Equal(t, createIndexOutput.StatusCode, 200)
	require.NotEmpty(t, createIndexOutput.RequestID)

	// 3. 获取向量索引信息
	getIndexInput := &tos.GetIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
	}
	getIndexOutput, err := client.GetIndex(context.Background(), getIndexInput)
	require.Nil(t, err)
	require.NotNil(t, getIndexOutput)
	require.Equal(t, getIndexOutput.StatusCode, 200)
	require.NotEmpty(t, getIndexOutput.RequestID)
	require.Equal(t, indexName, getIndexOutput.Index.IndexName)
	require.Equal(t, vectorBucketName, getIndexOutput.Index.VectorBucketName)
	require.Equal(t, enum.DataTypeFloat32, getIndexOutput.Index.DataType)
	require.Equal(t, 128, getIndexOutput.Index.Dimension)
	require.Equal(t, enum.DistanceMetricEuclidean, getIndexOutput.Index.DistanceMetric)
	require.NotEmpty(t, getIndexOutput.Index.IndexTrn)
	require.NotZero(t, getIndexOutput.Index.CreationTime)
	require.Equal(t, []string{"timestamp", "source"}, getIndexOutput.Index.MetadataConfiguration.NonFilterableMetadataKeys)

	// 4. 删除向量索引
	deleteIndexInput := &tos.DeleteIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
	}
	deleteIndexOutput, err := client.DeleteIndex(context.Background(), deleteIndexInput)
	require.Nil(t, err)
	require.NotNil(t, deleteIndexOutput)
	require.Equal(t, deleteIndexOutput.StatusCode, 200)
	require.NotEmpty(t, deleteIndexOutput.RequestID)

	// 5. 再次获取向量索引，应该返回404
	getIndexOutput2, err := client.GetIndex(context.Background(), getIndexInput)
	require.NotNil(t, err)
	require.Nil(t, getIndexOutput2)
	serr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.Equal(t, serr.StatusCode, 404)
}

func TestListIndexes(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector" + randomString(8)
		accountID        = e.accountId
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	client.CreateVectorBucket(ctx, createBucketInput)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 创建多个测试索引用于列表测试
	indexNames := []string{
		"test-index-1",
		"test-index-2",
		"prefix-index-1",
		"prefix-index-2",
	}
	for _, indexName := range indexNames {
		createIndexInput := &tos.CreateIndexInput{
			VectorBucketName: vectorBucketName,
			AccountID:        accountID,
			IndexName:        indexName,
			DataType:         enum.DataTypeFloat32,
			Dimension:        128,
			DistanceMetric:   enum.DistanceMetricEuclidean,
		}
		client.CreateIndex(ctx, createIndexInput)
	}

	// 3. 测试基本列举功能
	listInput := &tos.ListIndexesInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
	}
	listOutput, err := client.ListIndexes(ctx, listInput)
	require.Nil(t, err)
	require.NotNil(t, listOutput)
	require.Equal(t, listOutput.StatusCode, 200)
	require.NotEmpty(t, listOutput.RequestID)
	require.NotNil(t, listOutput.Indexes)
	require.GreaterOrEqual(t, len(listOutput.Indexes), 4, "Should have at least 4 indexes")

	// 验证返回的索引数据结构
	index := listOutput.Indexes[0]
	require.False(t, index.CreationTime.IsZero(), "CreationTime should not be zero")
	require.NotEmpty(t, index.IndexName, "IndexName should not be empty")
	require.NotEmpty(t, index.IndexTrn, "IndexTrn should not be empty")
	require.Equal(t, vectorBucketName, index.VectorBucketName, "VectorBucketName should match")

	// 4. 测试maxResults参数限制
	listInputWithLimit := &tos.ListIndexesInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		MaxResults:       2,
	}
	listOutputWithLimit, err := client.ListIndexes(ctx, listInputWithLimit)
	require.Nil(t, err)
	require.NotNil(t, listOutputWithLimit)
	require.Equal(t, listOutputWithLimit.StatusCode, 200)
	require.NotNil(t, listOutputWithLimit.Indexes)
	require.Equal(t, len(listOutputWithLimit.Indexes), 2, "Should return at most 2 indexes")

	// 5. 测试prefix参数过滤
	listInputWithPrefix := &tos.ListIndexesInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		Prefix:           "prefix-",
	}
	listOutputWithPrefix, err := client.ListIndexes(ctx, listInputWithPrefix)
	require.Nil(t, err)
	require.NotNil(t, listOutputWithPrefix)
	require.Equal(t, listOutputWithPrefix.StatusCode, 200)
	require.NotNil(t, listOutputWithPrefix.Indexes)

	// 验证所有返回的索引名称都以指定前缀开头
	for _, index := range listOutputWithPrefix.Indexes {
		require.True(t, strings.HasPrefix(index.IndexName, "prefix-"),
			"Index name %s should start with 'prefix-'", index.IndexName)
	}

	// 6. 测试分页功能
	// 先列举一次获取 nextToken
	firstPageInput := &tos.ListIndexesInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		MaxResults:       2,
	}
	firstPage, err := client.ListIndexes(ctx, firstPageInput)
	require.Nil(t, err)
	require.NotNil(t, firstPage)
	require.Equal(t, firstPage.StatusCode, 200)
	require.NotNil(t, firstPage.Indexes)
	require.Equal(t, 2, len(firstPage.Indexes), "Should return exactly 2 indexes")
	require.NotEmpty(t, firstPage.NextToken, "NextToken should not be empty")

	// 如果有 nextToken，继续列举下一页
	secondPageInput := &tos.ListIndexesInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		MaxResults:       2,
		NextToken:        firstPage.NextToken,
	}
	secondPage, err := client.ListIndexes(ctx, secondPageInput)
	require.Nil(t, err)
	require.NotNil(t, secondPage)
	require.Equal(t, secondPage.StatusCode, 200)
	require.NotNil(t, secondPage.Indexes)
	require.LessOrEqual(t, len(secondPage.Indexes), 2, "Should return at most 2 indexes")

	// 验证第一页和第二页数据不重复
	firstPageNames := make(map[string]bool)
	for _, idx := range firstPage.Indexes {
		firstPageNames[idx.IndexName] = true
	}
	for _, idx := range secondPage.Indexes {
		require.False(t, firstPageNames[idx.IndexName],
			"Index %s should not be in both pages", idx.IndexName)
	}
}
