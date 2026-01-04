package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestVideoDataProcess(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = env.testBucketName
		client = env.prepareClient("")
	)
	options := []tos.ClientOption{
		tos.WithRegion(env.region2),
		tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithMaxRetryCount(5),
	}
	client, err := tos.NewClientV2(env.endpoint2, options...)
	require.Nil(t, err)
	// 3. 测试视频数据处理 - 提取快照
	processInput := &tos.VideoDataProcessInput{
		Bucket:  bucket,
		Key:     env.testVideoKey,
		Process: "x-tos-post-process=video/snapshots,f_png,m_index,w_400,h_400,index_0|10|3000&x-tos-save-object=b3V0cHV0XyR7TnVtYmVyfS5wbmc=",
	}

	output, err := client.VideoDataProcess(context.Background(), processInput)
	require.Nil(t, err)
	require.NotNil(t, output)

	// 验证输出结果
	assert.NotEmpty(t, output.OutputBucket)
	assert.Equal(t, output.RequestInfo.StatusCode, 200)
	assert.Equal(t, output.TotalFrameCount, 3)
	assert.Equal(t, output.SuccFrameCount, 2)
	assert.Equal(t, output.FailFrameCount, 1)
	assert.Equal(t, len(output.SuccFrameList), 2)
	assert.Equal(t, len(output.FailFrameList), 1)

	ouput2, err := client.VideoDataProcess(context.Background(), &tos.VideoDataProcessInput{
		Bucket:  bucket,
		Key:     env.testVideoKey,
		Process: "f_png,m_index,w_400,h_400,index_0|10|3000&x-tos-save-object=b3V0cHV0XyR7TnVtYmVyfS5wbmc=",
	})

	require.Nil(t, err)
	assert.NotEmpty(t, ouput2.OutputBucket)
	assert.Equal(t, ouput2.RequestInfo.StatusCode, 200)
	assert.Equal(t, ouput2.TotalFrameCount, 3)
	assert.Equal(t, ouput2.SuccFrameCount, 2)
	assert.Equal(t, ouput2.FailFrameCount, 1)
	assert.Equal(t, len(ouput2.SuccFrameList), 2)
	assert.Equal(t, len(ouput2.FailFrameList), 1)
}
