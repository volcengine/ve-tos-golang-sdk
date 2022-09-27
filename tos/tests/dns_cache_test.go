package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestDnsCache(t *testing.T) {

	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("dns-cache")
		client = env.prepareClient(bucket, tos.WithTransportConfig(&tos.TransportConfig{
			InsecureSkipVerify: true,
			KeepAlive:          time.Second,
			IdleConnTimeout:    time.Millisecond,
		}), tos.WithDNSCacheTime(1))
		value = randomString(1024)
	)

	defer func() {
		cleanBucket(t, client, bucket)
	}()
	var wg sync.WaitGroup
	count := 10
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer func() {
				fmt.Println("done")
				wg.Done()
			}()
			key := randomString(8)
			ctx := context.Background()
			time.Sleep(time.Second)
			putRes, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
				PutObjectBasicInput: tos.PutObjectBasicInput{
					Bucket: bucket,
					Key:    key,
				},
				Content: bytes.NewBufferString(value),
			})
			require.Nil(t, err)
			require.Equal(t, putRes.StatusCode, http.StatusOK)
			time.Sleep(time.Second)

			getRes, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
				Bucket: bucket,
				Key:    key,
			})
			require.Nil(t, err)
			require.Equal(t, getRes.StatusCode, http.StatusOK)
		}()
	}
	wg.Wait()

}
