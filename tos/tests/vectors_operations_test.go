package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func waitForVectorsCount(t *testing.T, client *tos.TosVectorsClient, vectorBucketName, accountID, indexName string, keys []string, expect int) {
	ctx := context.Background()
	deadline := time.Now().Add(5 * time.Minute)
	lastCount := -1
	for {
		out, err := client.GetVectors(ctx, &tos.GetVectorsInput{
			VectorBucketName: vectorBucketName,
			AccountID:        accountID,
			IndexName:        indexName,
			Keys:             keys,
			ReturnData:       false,
			ReturnMetadata:   false,
		})
		if err == nil && out != nil {
			lastCount = len(out.Vectors)
			t.Logf("get vectors count: %d", lastCount)
			if lastCount == expect {
				return
			}
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(5 * time.Second)
	}
}

func TestVectorsOperations(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector-ops-" + randomString(8)
		accountID        = e.accountId
		indexName        = "test-index-" + randomString(8)
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(ctx, createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, createBucketOutput.StatusCode, 200)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 创建向量索引
	createIndexInput := &tos.CreateIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		DataType:         enum.DataTypeFloat32,
		Dimension:        128,
		DistanceMetric:   enum.DistanceMetricEuclidean,
	}
	createIndexOutput, err := client.CreateIndex(ctx, createIndexInput)
	require.Nil(t, err)
	require.NotNil(t, createIndexOutput)
	require.Equal(t, createIndexOutput.StatusCode, 200)

	// 3. 测试 PutVectors - 批量插入向量
	vectors := []tos.Vector{
		{
			Key: "vector-key-1",
			Data: tos.VectorData{
				Value: generateRandomVector(128),
			},
			Metadata: map[string]interface{}{
				"category":  "electronics",
				"timestamp": 1234567890,
				"source":    "user-upload",
			},
		},
		{
			Key: "vector-key-2",
			Data: tos.VectorData{
				Value: generateRandomVector(128),
			},
			Metadata: map[string]interface{}{
				"category":  "clothing",
				"timestamp": 1234567891,
				"source":    "batch-import",
			},
		},
		{
			Key: "vector-key-3",
			Data: tos.VectorData{
				Value: generateRandomVector(128),
			},
			Metadata: map[string]interface{}{
				"category":  "books",
				"timestamp": 1234567892,
				"source":    "api-call",
			},
		},
	}

	putVectorsInput := &tos.PutVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Vectors:          vectors,
	}
	putVectorsOutput, err := client.PutVectors(ctx, putVectorsInput)
	require.Nil(t, err)
	require.NotNil(t, putVectorsOutput)
	require.Equal(t, putVectorsOutput.StatusCode, 200)
	require.NotEmpty(t, putVectorsOutput.RequestID)

	waitForVectorsCount(t, client, vectorBucketName, accountID, indexName, []string{"vector-key-1", "vector-key-2", "vector-key-3"}, 3)

	// 4. 测试 GetVectors - 查询向量（只返回基本信息）
	getVectorsInput := &tos.GetVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Keys:             []string{"vector-key-1", "vector-key-2"},
	}
	getVectorsOutput, err := client.GetVectors(ctx, getVectorsInput)
	require.Nil(t, err)
	require.NotNil(t, getVectorsOutput)
	require.Equal(t, getVectorsOutput.StatusCode, 200)
	require.NotEmpty(t, getVectorsOutput.RequestID)
	require.NotNil(t, getVectorsOutput.Vectors)
	require.Equal(t, len(getVectorsOutput.Vectors), 2)

	// 验证返回的向量基本信息
	vectorMap := make(map[string]tos.Vector)
	for _, vec := range getVectorsOutput.Vectors {
		vectorMap[vec.Key] = vec
	}
	require.Contains(t, vectorMap, "vector-key-1")
	require.Contains(t, vectorMap, "vector-key-2")

	// 5. 测试 GetVectors - 查询向量（返回数据和元数据）
	getVectorsWithDataInput := &tos.GetVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Keys:             []string{"vector-key-1", "vector-key-2", "vector-key-3"},
		ReturnData:       true,
		ReturnMetadata:   true,
	}
	getVectorsWithDataOutput, err := client.GetVectors(ctx, getVectorsWithDataInput)
	require.Nil(t, err)
	require.NotNil(t, getVectorsWithDataOutput)
	require.Equal(t, getVectorsWithDataOutput.StatusCode, 200)
	require.NotNil(t, getVectorsWithDataOutput.Vectors)
	require.Equal(t, len(getVectorsWithDataOutput.Vectors), 3)

	// 验证返回的向量数据和元数据
	vectorWithDataMap := make(map[string]tos.Vector)
	for _, vec := range getVectorsWithDataOutput.Vectors {
		vectorWithDataMap[vec.Key] = vec
	}

	// 验证 vector-key-1
	vec1, exists := vectorWithDataMap["vector-key-1"]
	require.True(t, exists)
	require.Equal(t, len(vec1.Data.Value), 128)
	require.Equal(t, vec1.Metadata["category"], "electronics")
	require.Equal(t, vec1.Metadata["source"], "user-upload")

	// 验证 vector-key-2
	vec2, exists := vectorWithDataMap["vector-key-2"]
	require.True(t, exists)
	require.Equal(t, len(vec2.Data.Value), 128)
	require.Equal(t, vec2.Metadata["category"], "clothing")
	require.Equal(t, vec2.Metadata["source"], "batch-import")

	// 验证 vector-key-3
	vec3, exists := vectorWithDataMap["vector-key-3"]
	require.True(t, exists)
	require.Equal(t, len(vec3.Data.Value), 128)
	require.Equal(t, vec3.Metadata["category"], "books")
	require.Equal(t, vec3.Metadata["source"], "api-call")

	// 6. 测试 DeleteVectors - 删除向量
	deleteVectorsInput := &tos.DeleteVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Keys:             []string{"vector-key-1", "vector-key-3"},
	}
	deleteVectorsOutput, err := client.DeleteVectors(ctx, deleteVectorsInput)
	require.Nil(t, err)
	require.NotNil(t, deleteVectorsOutput)
	require.Equal(t, deleteVectorsOutput.StatusCode, 200)
	require.NotEmpty(t, deleteVectorsOutput.RequestID)

	waitForVectorsCount(t, client, vectorBucketName, accountID, indexName, []string{"vector-key-1", "vector-key-2", "vector-key-3"}, 1)

	// 7. 验证删除后的向量查询
	getVectorsAfterDeleteInput := &tos.GetVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Keys:             []string{"vector-key-1", "vector-key-2", "vector-key-3"},
		ReturnData:       true,
		ReturnMetadata:   true,
	}
	getVectorsAfterDeleteOutput, err := client.GetVectors(ctx, getVectorsAfterDeleteInput)
	require.Nil(t, err)
	require.NotNil(t, getVectorsAfterDeleteOutput)
	require.Equal(t, getVectorsAfterDeleteOutput.StatusCode, 200)
	require.NotNil(t, getVectorsAfterDeleteOutput.Vectors)
	require.Equal(t, len(getVectorsAfterDeleteOutput.Vectors), 1) // 只有 vector-key-2 应该还存在
	require.Equal(t, getVectorsAfterDeleteOutput.Vectors[0].Key, "vector-key-2")
}

// 生成随机向量数据
func generateRandomVector(dimension int) []float32 {
	vector := make([]float32, dimension)
	for i := 0; i < dimension; i++ {
		vector[i] = float32(i) * 0.1 // 简单的测试数据
	}
	return vector
}

func TestQueryVectors(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-query-" + randomString(8)
		accountID        = e.accountId
		indexName        = "test-index-" + randomString(8)
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(ctx, createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, 200, createBucketOutput.StatusCode)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 创建向量索引
	createIndexInput := &tos.CreateIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		DataType:         enum.DataTypeFloat32,
		Dimension:        128,
		DistanceMetric:   enum.DistanceMetricCosine,
	}
	createIndexOutput, err := client.CreateIndex(ctx, createIndexInput)
	require.Nil(t, err)
	require.NotNil(t, createIndexOutput)
	require.Equal(t, 200, createIndexOutput.StatusCode)

	// 3. 准备测试向量数据
	testVectors := make([]tos.Vector, 10)
	for i := 0; i < 10; i++ {
		vectorData := make([]float32, 128)
		for j := 0; j < 128; j++ {
			vectorData[j] = float32(i)*0.1 + float32(j)*0.01
		}

		category := "A"
		if i%2 == 0 {
			category = "B"
		}

		testVectors[i] = tos.Vector{
			Key: "test-vector-" + string(rune('0'+i)),
			Data: tos.VectorData{
				Value: vectorData,
			},
			Metadata: map[string]interface{}{
				"category":  category,
				"timestamp": i * 1000,
				"index":     i,
			},
		}
	}

	// 4. 上传测试向量
	putVectorsInput := &tos.PutVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Vectors:          testVectors,
	}
	putVectorsOutput, err := client.PutVectors(ctx, putVectorsInput)
	require.Nil(t, err)
	require.NotNil(t, putVectorsOutput)
	require.Equal(t, 200, putVectorsOutput.StatusCode)
	waitForVectorsCount(t, client, vectorBucketName, accountID, indexName, []string{testVectors[0].Key}, 1)

	// 5. 测试基本向量查询
	queryVector := make([]float32, 128)
	for i := 0; i < 128; i++ {
		queryVector[i] = 2 + float32(i)*0.01
	}

	queryInput := &tos.QueryVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		ReturnDistance:   true,
		ReturnMetadata:   true,
		TopK:             5,
		QueryVector: tos.VectorData{
			Value: queryVector,
		},
	}

	queryOutput, err := client.QueryVectors(ctx, queryInput)
	require.Nil(t, err)
	require.NotNil(t, queryOutput)
	require.Equal(t, 200, queryOutput.StatusCode)
	require.NotEmpty(t, queryOutput.RequestID)
	require.NotNil(t, queryOutput.Vectors)
	require.LessOrEqual(t, len(queryOutput.Vectors), 5)

	// 验证返回的向量结构
	vector := queryOutput.Vectors[0]
	require.NotEmpty(t, vector.Key)
	require.NotNil(t, vector.Data)
	require.NotNil(t, vector.Data.Value)
	require.Equal(t, 128, len(vector.Data.Value))
	require.NotZero(t, vector.Distance)
	require.NotNil(t, vector.Metadata)

	// 6. 测试带过滤条件的查询
	queryInputWithFilter := &tos.QueryVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		ReturnDistance:   true,
		ReturnMetadata:   true,
		TopK:             10,
		QueryVector: tos.VectorData{
			Value: queryVector,
		},
		Filter: map[string]interface{}{
			"category": "A",
		},
	}

	queryOutputWithFilter, err := client.QueryVectors(ctx, queryInputWithFilter)
	require.Nil(t, err)
	require.NotNil(t, queryOutputWithFilter)
	require.Equal(t, 200, queryOutputWithFilter.StatusCode)
	require.NotNil(t, queryOutputWithFilter.Vectors)

	// 验证过滤后的结果
	for _, vector := range queryOutputWithFilter.Vectors {
		require.Equal(t, "A", vector.Metadata["category"])
	}

	// 7. 测试不返回距离的查询
	queryInputNoDistance := &tos.QueryVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		ReturnDistance:   false,
		ReturnMetadata:   true,
		TopK:             3,
		QueryVector: tos.VectorData{
			Value: queryVector,
		},
	}

	queryOutputNoDistance, err := client.QueryVectors(ctx, queryInputNoDistance)
	require.Nil(t, err)
	require.NotNil(t, queryOutputNoDistance)
	require.Equal(t, 200, queryOutputNoDistance.StatusCode)
	require.NotNil(t, queryOutputNoDistance.Vectors)

	vector = queryOutputNoDistance.Vectors[0]
	require.NotEmpty(t, vector.Key)
	require.NotNil(t, vector.Data)
	require.Zero(t, vector.Distance) // 距离应该为0（不返回）
	require.NotNil(t, vector.Metadata)

	// 8. 测试不返回元数据的查询
	queryInputNoMetadata := &tos.QueryVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		ReturnDistance:   true,
		ReturnMetadata:   false,
		TopK:             3,
		QueryVector: tos.VectorData{
			Value: queryVector,
		},
	}

	queryOutputNoMetadata, err := client.QueryVectors(ctx, queryInputNoMetadata)
	require.Nil(t, err)
	require.NotNil(t, queryOutputNoMetadata)
	require.Equal(t, 200, queryOutputNoMetadata.StatusCode)
	require.NotNil(t, queryOutputNoMetadata.Vectors)

	vector = queryOutputNoMetadata.Vectors[0]
	require.NotEmpty(t, vector.Key)
	require.NotNil(t, vector.Data)
	require.NotZero(t, vector.Distance)
	require.Nil(t, vector.Metadata) // 元数据应该为nil

	// 9. 测试不同的TopK值
	topKValues := []int{1, 3, 5, 10}
	for _, topK := range topKValues {
		queryInputTopK := &tos.QueryVectorsInput{
			VectorBucketName: vectorBucketName,
			AccountID:        accountID,
			IndexName:        indexName,
			ReturnDistance:   true,
			ReturnMetadata:   true,
			TopK:             topK,
			QueryVector: tos.VectorData{
				Value: queryVector,
			},
		}

		queryOutputTopK, err := client.QueryVectors(ctx, queryInputTopK)
		require.Nil(t, err)
		require.NotNil(t, queryOutputTopK)
		require.Equal(t, 200, queryOutputTopK.StatusCode)
		require.NotNil(t, queryOutputTopK.Vectors)
		require.LessOrEqual(t, len(queryOutputTopK.Vectors), topK)
	}

	// 10. 测试空结果查询
	queryInputEmpty := &tos.QueryVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		ReturnDistance:   true,
		ReturnMetadata:   true,
		TopK:             5,
		QueryVector: tos.VectorData{
			Value: queryVector,
		},
		Filter: map[string]interface{}{
			"category": "NonExistentCategory",
		},
	}

	queryOutputEmpty, err := client.QueryVectors(ctx, queryInputEmpty)
	require.Nil(t, err)
	require.NotNil(t, queryOutputEmpty)
	require.Equal(t, 200, queryOutputEmpty.StatusCode)
	require.Nil(t, queryOutputEmpty.Vectors) // 应该返回nil或空数组
}

func TestListVectors(t *testing.T) {
	var (
		e                = newTestEnv(t)
		client           = e.prepareVectorsClient("")
		vectorBucketName = "test-vector-bucket-" + randomString(8)
		accountID        = e.accountId
		indexName        = "test-index-" + randomString(8)
		ctx              = context.Background()
	)

	// 1. 创建向量桶
	createBucketInput := &tos.CreateVectorBucketInput{
		VectorBucketName: vectorBucketName,
	}
	createBucketOutput, err := client.CreateVectorBucket(ctx, createBucketInput)
	require.Nil(t, err)
	require.NotNil(t, createBucketOutput)
	require.Equal(t, createBucketOutput.StatusCode, 200)

	defer func() {
		deleteVectorBucket(client, vectorBucketName, accountID)
	}()

	// 2. 创建向量索引
	createIndexInput := &tos.CreateIndexInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		DataType:         enum.DataTypeFloat32,
		Dimension:        128,
		DistanceMetric:   enum.DistanceMetricEuclidean,
	}
	createIndexOutput, err := client.CreateIndex(ctx, createIndexInput)
	require.Nil(t, err)
	require.NotNil(t, createIndexOutput)
	require.Equal(t, createIndexOutput.StatusCode, 200)

	// 3. 创建测试向量数据
	testVectorsCount := 15
	testVectors := make([]tos.Vector, testVectorsCount)
	for i := 0; i < testVectorsCount; i++ {
		vectorData := make([]float32, 128)
		for j := 0; j < 128; j++ {
			vectorData[j] = float32(i) * 0.1
		}

		testVectors[i] = tos.Vector{
			Key: fmt.Sprintf("test-vector-%03d-%d", i, time.Now().Unix()),
			Data: tos.VectorData{
				Value: vectorData,
			},
			Metadata: map[string]interface{}{
				"index":     i,
				"category":  i % 2,
				"timestamp": time.Now().Unix(),
			},
		}
	}

	// 4. 批量写入测试向量
	putVectorsInput := &tos.PutVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		Vectors:          testVectors,
	}
	putVectorsOutput, err := client.PutVectors(ctx, putVectorsInput)
	require.Nil(t, err)
	require.NotNil(t, putVectorsOutput)
	require.Equal(t, putVectorsOutput.StatusCode, 200)

	keys := make([]string, 0, testVectorsCount)
	for _, vector := range testVectors {
		keys = append(keys, vector.Key)
	}
	waitForVectorsCount(t, client, vectorBucketName, accountID, indexName, keys, testVectorsCount)

	// 5. 测试基本列举功能
	listInput := &tos.ListVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
	}
	listOutput, err := client.ListVectors(ctx, listInput)
	require.Nil(t, err)
	require.NotNil(t, listOutput)
	require.Equal(t, listOutput.StatusCode, 200)
	require.NotEmpty(t, listOutput.RequestID)
	require.NotNil(t, listOutput.Vectors)
	require.Equal(t, len(listOutput.Vectors), testVectorsCount, "Should have at least %d vectors", testVectorsCount)
	vector := listOutput.Vectors[0]
	require.NotEmpty(t, vector.Key, "Vector key should not be empty")
	require.Zero(t, vector.Data.Value)
	require.Zero(t, vector.Metadata)

	// 6. 测试分页功能
	pageSize := 5

	// 第一页
	page1Input := &tos.ListVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		MaxResults:       pageSize,
		ReturnData:       true,
		ReturnMetadata:   true,
	}
	page1Output, err := client.ListVectors(ctx, page1Input)
	require.Nil(t, err)
	require.NotNil(t, page1Output)
	require.Equal(t, page1Output.StatusCode, 200)
	require.Equal(t, len(page1Output.Vectors), pageSize, "First page should have exactly %d vectors", pageSize)
	require.NotEmpty(t, page1Output.NextToken, "NextToken should be present for pagination")

	// 验证第一页的数据完整性
	for _, vector := range page1Output.Vectors {
		require.NotEmpty(t, vector.Key, "Vector key should not be empty")
		require.NotNil(t, vector.Data.Value, "Vector data should not be nil when ReturnData=true")
		require.Equal(t, len(vector.Data.Value), 128, "Vector dimension should be 128")
		require.NotNil(t, vector.Metadata, "Vector metadata should not be nil when ReturnMetadata=true")
		require.Contains(t, vector.Metadata, "index")
		require.Contains(t, vector.Metadata, "category")
		require.Contains(t, vector.Metadata, "timestamp")
	}

	// 第二页
	page2Input := &tos.ListVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		MaxResults:       pageSize,
		NextToken:        page1Output.NextToken,
		ReturnData:       true,
		ReturnMetadata:   true,
	}
	page2Output, err := client.ListVectors(ctx, page2Input)
	require.Nil(t, err)
	require.NotNil(t, page2Output)
	require.Equal(t, page2Output.StatusCode, 200)
	require.Equal(t, len(page2Output.Vectors), pageSize, "Second page should have exactly %d vectors", pageSize)
	require.NotEmpty(t, page2Output.NextToken, "NextToken should be present for pagination")

	// 验证没有重复向量
	page1Keys := make(map[string]bool)
	for _, vector := range page1Output.Vectors {
		page1Keys[vector.Key] = true
	}
	for _, vector := range page2Output.Vectors {
		require.False(t, page1Keys[vector.Key], "Vector %s should not be in both pages", vector.Key)
	}

	// 第三页（可能不足pageSize）
	page3Input := &tos.ListVectorsInput{
		VectorBucketName: vectorBucketName,
		AccountID:        accountID,
		IndexName:        indexName,
		MaxResults:       pageSize,
		NextToken:        page2Output.NextToken,
		ReturnData:       true,
		ReturnMetadata:   true,
	}
	page3Output, err := client.ListVectors(ctx, page3Input)
	require.Nil(t, err)
	require.NotNil(t, page3Output)
	require.Equal(t, page3Output.StatusCode, 200)
	require.GreaterOrEqual(t, len(page3Output.Vectors), 1, "Third page should have at least 1 vector")
}
