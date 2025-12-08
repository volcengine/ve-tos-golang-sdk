package tests

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestBucketRename(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("rename")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	resp, err := client.PutBucketRename(ctx, &tos.PutBucketRenameInput{Bucket: bucket, RenameEnable: true})
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK)

	getResp, err := client.GetBucketRename(ctx, &tos.GetBucketRenameInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, getResp.RenameEnable, true)

	time.Sleep(time.Second * 60)

	key := "key-1"
	rawData := randomString(1024)
	putObj, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(rawData),
	})
	require.Nil(t, err)
	require.Equal(t, putObj.StatusCode, http.StatusOK)

	newKey := randomString(8)
	renameObj, err := client.RenameObject(ctx, &tos.RenameObjectInput{
		Bucket: bucket,
		Key:    key,
		NewKey: newKey,
	})
	require.Nil(t, err)
	require.Equal(t, renameObj.StatusCode, http.StatusNoContent)

	getObject, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    newKey,
	})
	require.Nil(t, err)
	require.Equal(t, getObject.StatusCode, 200)

	data, err := ioutil.ReadAll(getObject.Content)
	require.Nil(t, err)
	require.Equal(t, string(data), rawData)

	deleteResp, err := client.DeleteBucketRename(ctx, &tos.DeleteBucketRenameInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, deleteResp.StatusCode, http.StatusNoContent)

}
