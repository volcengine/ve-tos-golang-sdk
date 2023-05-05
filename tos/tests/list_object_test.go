package tests

//
//import (
//	"context"
//	"fmt"
//	"github.com/stretchr/testify/require"
//	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
//	"net/http"
//	"strings"
//	"testing"
//)
//

//func TestListObjWithMeta(t *testing.T) {
//
//	client := newClient(t)
//	key := "meta-" + randomString(6)
//	data := strings.NewReader(randomString(1024))
//	res, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
//		PutObjectBasicInput: tos.PutObjectBasicInput{
//			Bucket: bucket,
//			Key:    key,
//			Meta:   map[string]string{"meta-key": "meta-value", "中文key": "中文value"},
//		},
//		Content: data,
//	})
//	require.Nil(t, err)
//	require.Equal(t, res.StatusCode, http.StatusOK)
//
//	listOut, err := client.ListObjectsType2(context.Background(), &tos.ListObjectsType2Input{
//		Bucket:    bucket,
//		Prefix:    "meta",
//		FetchMeta: true,
//	})
//	require.Nil(t, err)
//
//	for index, obj := range listOut.Contents {
//		obj.Meta.Range(func(key, value string) bool {
//			fmt.Println(index, ". ", "Key:", key, " Value:", value)
//			return true
//		})
//	}
//
//	fmt.Println("===List Object Type End===")
//	fmt.Println("")
//
//	output, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{
//		Bucket: bucket,
//		ListObjectsInput: tos.ListObjectsInput{
//			FetchMeta: true,
//		},
//	})
//	require.Nil(t, err)
//	for index, obj := range output.Contents {
//		if obj.Meta != nil {
//			obj.Meta.Range(func(key, value string) bool {
//				fmt.Println(index, ". ", "Key:", key, " Value:", value)
//				return true
//			})
//		}
//	}
//
//	fmt.Println("===ListObjectsV2 Type End===")
//	fmt.Println("")
//
//	clientv1, err := tos.NewClient(endpoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(ak, sk)), tos.WithEnableVerifySSL(false))
//	require.Nil(t, err)
//	bkt, err := clientv1.Bucket(bucket)
//	resV1, err := bkt.ListObjects(context.Background(), &tos.ListObjectsInput{
//		FetchMeta: true,
//	})
//	require.Nil(t, err)
//	for index, obj := range resV1.Contents {
//		if obj.Meta != nil {
//			obj.Meta.Range(func(key, value string) bool {
//				fmt.Println(index, ". ", "Key:", key, " Value:", value)
//				return true
//			})
//		}
//	}
//
//	fmt.Println("=== List Objects Type End===")
//	fmt.Println("")
//	fmt.Println("=== List Object Versions Type Start===")
//	v1Res, err := bkt.ListObjectVersions(context.Background(), &tos.ListObjectVersionsInput{FetchMeta: true})
//	require.Nil(t, err)
//	for index, obj := range v1Res.Versions {
//		if obj.Meta != nil {
//			obj.Meta.Range(func(key, value string) bool {
//				fmt.Println(index, ". ", "Key:", key, " Value:", value)
//				return true
//			})
//		}
//	}
//
//	fmt.Println("=== List Object Versions   End===")
//	fmt.Println("")
//
//	fmt.Println("=== List Object Versions v2 Start===")
//	v2Res, err := client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
//		Bucket:                  bucket,
//		ListObjectVersionsInput: tos.ListObjectVersionsInput{FetchMeta: true},
//	})
//	require.Nil(t, err)
//	for index, obj := range v2Res.Versions {
//		if obj.Meta != nil {
//			obj.Meta.Range(func(key, value string) bool {
//				fmt.Println(index, ". ", "Key:", key, " Value:", value)
//				return true
//			})
//		}
//	}
//
//	fmt.Println("=== List Object Versions V2 End===")
//	fmt.Println("")
//
//}
//func newClient(t *testing.T) *tos.ClientV2 {
//	client, err := tos.NewClientV2(endpoint, tos.WithRegion(region), tos.WithCredentials(tos.NewStaticCredentials(ak, sk)), tos.WithEnableVerifySSL(false))
//	require.Nil(t, err)
//	return client
//}
