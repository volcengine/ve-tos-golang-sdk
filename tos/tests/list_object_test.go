package tests

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestListObjWithMeta(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("bucket-version")
		client = env.prepareClient(bucket)
	)
	defer cleanBucket(t, client, bucket)
	key := "meta-" + randomString(6)
	data := strings.NewReader(randomString(1024))
	metaKey := "中文key"
	metaValue := "!@#$%^&*()_+-=[]{}|;':\"%2f%fg,     「」：-=+、\n./<>?中文测试编码%20%%%^&abcd /\\"
	res, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
			Meta:   map[string]string{"meta-key": "meta-value", metaKey: metaValue},
		},
		Content: data,
	})
	require.Nil(t, err)
	require.Equal(t, res.StatusCode, http.StatusOK)

	listOut, err := client.ListObjectsType2(context.Background(), &tos.ListObjectsType2Input{
		Bucket:    bucket,
		Prefix:    "meta",
		FetchMeta: true,
	})
	require.Nil(t, err)

	for index, obj := range listOut.Contents {
		obj.Meta.Range(func(key, value string) bool {
			fmt.Println(index, ". ", "Key:", key, " Value:", value)
			return true
		})
		mValue, ok := obj.Meta.Get(metaKey)
		require.True(t, ok)
		require.Equal(t, mValue, metaValue)
	}

	fmt.Println("===List Object Type End===")
	fmt.Println("")

	output, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{
		Bucket: bucket,
		ListObjectsInput: tos.ListObjectsInput{
			FetchMeta: true,
		},
	})
	require.Nil(t, err)
	for index, obj := range output.Contents {
		if obj.Meta != nil {
			obj.Meta.Range(func(key, value string) bool {
				fmt.Println(index, ". ", "Key:", key, " Value:", value)
				return true
			})
			mValue, ok := obj.Meta.Get(metaKey)
			require.True(t, ok)
			require.Equal(t, mValue, metaValue)
		}
	}

	fmt.Println("===ListObjectsV2 Type End===")
	fmt.Println("")

	clientv1, err := tos.NewClient(env.endpoint, tos.WithRegion(env.region), tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)), tos.WithEnableVerifySSL(false))
	require.Nil(t, err)
	bkt, err := clientv1.Bucket(bucket)
	resV1, err := bkt.ListObjects(context.Background(), &tos.ListObjectsInput{
		FetchMeta: true,
	})
	require.Nil(t, err)
	for index, obj := range resV1.Contents {
		if obj.Meta != nil {
			obj.Meta.Range(func(key, value string) bool {
				fmt.Println(index, ". ", "Key:", key, " Value:", value)
				return true
			})
			mValue, ok := obj.Meta.Get(metaKey)
			require.True(t, ok)
			require.Equal(t, mValue, metaValue)
			require.True(t, obj.HashCrc64ecma > 0)

		}
	}

	fmt.Println("=== List Objects Type End===")
	fmt.Println("")
	fmt.Println("=== List Object Versions Type Start===")
	v1Res, err := bkt.ListObjectVersions(context.Background(), &tos.ListObjectVersionsInput{FetchMeta: true})
	require.Nil(t, err)
	for index, obj := range v1Res.Versions {
		if obj.Meta != nil {
			obj.Meta.Range(func(key, value string) bool {
				fmt.Println(index, ". ", "Key:", key, " Value:", value)
				return true
			})
			mValue, ok := obj.Meta.Get(metaKey)
			require.True(t, ok)
			require.Equal(t, mValue, metaValue)
			require.True(t, obj.HashCrc64ecma > 0)
		}
	}

	fmt.Println("=== List Object Versions   End===")
	fmt.Println("")

	fmt.Println("=== List Object Versions v2 Start===")
	v2Res, err := client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
		Bucket:                  bucket,
		ListObjectVersionsInput: tos.ListObjectVersionsInput{FetchMeta: true},
	})
	require.Nil(t, err)
	for index, obj := range v2Res.Versions {
		if obj.Meta != nil {
			obj.Meta.Range(func(key, value string) bool {
				fmt.Println(index, ". ", "Key:", key, " Value:", value)
				return true
			})
			mValue, ok := obj.Meta.Get(metaKey)
			require.True(t, ok)
			require.Equal(t, mValue, metaValue)
		}
	}

	fmt.Println("=== List Object Versions V2 End===")
	fmt.Println("")

}
