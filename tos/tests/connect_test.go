package tests

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

// 验证连接复用
// 请求 2 个 sleep 30 秒后再发两个请求连接是复用的
func TestClientConnect(t *testing.T) {
	env := newTestEnv(t)
	options := []tos.ClientOption{
		tos.WithRegion(env.region),
		tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithMaxRetryCount(5),
	}
	client, err := tos.NewClientV2(env.endpoint, options...)
	require.Nil(t, err)
	for i := 0; i < 2; i++ {
		go func() {
			_, err = client.HeadBucket(context.Background(), &tos.HeadBucketInput{
				Bucket: "bzy" + strconv.Itoa(i),
			})
			require.NotNil(t, err)
			fmt.Println(err.Error())
		}()
		go func() {
			_, err = client.HeadBucket(context.Background(), &tos.HeadBucketInput{
				Bucket: "bzy" + strconv.Itoa(i),
			})
			require.NotNil(t, err)
			fmt.Println(err.Error())
		}()
		fmt.Println("OK")
		time.Sleep(10 * time.Second)

	}

}
