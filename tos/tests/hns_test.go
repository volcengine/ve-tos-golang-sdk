package tests

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type Trash struct {
	TrashPath     string `json:"TrashPath"`     // must,回收站目录,输入必须是目录形式
	CleanInterval uint32 `json:"CleanInterval"` // must，天数，必须大于0
	Status        string `json:"Status"`        // must，Enabled/Disabled
}

type BucketTrash struct {
	Trash Trash `json:"Trash"`
}

func TestHNS(t *testing.T) {

	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("hns")
		client = env.prepareClient("")
		ctx    = context.Background()
	)

	_, err := client.CreateBucketV2(ctx, &tos.CreateBucketV2Input{
		Bucket:     bucket,
		BucketType: enum.BucketTypeHNS,
	})
	require.Nil(t, err)

	defer func() {
		cleanHNSBucket(t, client, bucket)
	}()

	resp, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    "a/b/c",
		},
		Content: strings.NewReader("test"),
	})
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)

	resp, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    "a/b/d",
		},
		Content: strings.NewReader("test"),
	})
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)

	preSignUrl, err := client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        "",
		Query:      map[string]string{"trash": ""},
	})
	require.Nil(t, err)

	bt := &BucketTrash{Trash{
		TrashPath:     "trash/",
		CleanInterval: 1,
		Status:        "Enabled",
	}}
	data, err := json.Marshal(bt)
	require.Nil(t, err)
	httpReq, err := http.NewRequest(http.MethodPut, preSignUrl.SignedUrl, strings.NewReader(string(data)))
	require.Nil(t, err)

	presp, err := http.DefaultClient.Do(httpReq)
	require.Nil(t, err)
	defer presp.Body.Close()
	data, err = ioutil.ReadAll(presp.Body)
	require.Nil(t, err)
	require.Equal(t, presp.StatusCode, 200)
	time.Sleep(time.Second * 30)
	listResp, err := client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{
		Bucket:           bucket,
		ListObjectsInput: tos.ListObjectsInput{Delimiter: "/"},
	})

	require.Nil(t, err)
	require.Equal(t, listResp.StatusCode, 200)
	require.Equal(t, len(listResp.CommonPrefixes), 2)
	require.NotNil(t, listResp.CommonPrefixes[0].LastModified)

	deleteOut, err := client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{
		Bucket: bucket,
		Key:    "a/b/d",
	})
	require.Nil(t, err)
	require.True(t, strings.HasSuffix(deleteOut.TrashPath, "a/b/d"))

	_, err = client.RenameObject(ctx, &tos.RenameObjectInput{
		Bucket:         bucket,
		Key:            "a/b/c",
		NewKey:         "a1/b1/c1",
		RecursiveMkdir: true,
	})
	require.Nil(t, err)
	deleteOut, err = client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{
		Bucket:    bucket,
		Key:       "a1/",
		Recursive: true,
	})
	require.Nil(t, err)
	require.Equal(t, deleteOut.StatusCode, 204)

	deleteOut, err = client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{
		Bucket:    bucket,
		Key:       "a1",
		Recursive: true,
	})
	require.Nil(t, err)
	require.Equal(t, deleteOut.StatusCode, 204)

	listResp, err = client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{
		Bucket:           bucket,
		ListObjectsInput: tos.ListObjectsInput{Delimiter: "/"},
	})

	require.Nil(t, err)
	require.Equal(t, listResp.StatusCode, 200)
	require.Equal(t, len(listResp.Contents), 0)

}

func TestList(t *testing.T) {
	env := newTestEnv(t)

	client := env.prepareClient("")
	ctx := context.Background()
	resp, err := client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{
		Bucket: "g0lan9-5dk-t39ts-wap9b9hb-hns",
	})
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)
}
