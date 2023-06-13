package tests

import (
	"context"
	"encoding/base64"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"net/http"
	"os"
	"strings"
	"testing"
)

type callback struct {
	CallbackUrl      string `json:"callbackUrl,omitempty"`
	CallbackBodyType string `json:"callbackBodyType,omitempty"`
	CallbackBody     string `json:"callbackBody"`
}

func TestCallback(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("callback")
		client   = env.prepareClient(bucket)
		ctx      = context.Background()
		key      = randomString(6)
		fileName = randomString(16) + ".file"
		value1   = randomString(2 * 1024 * 1024)
	)
	defer func() { cleanBucket(t, client, bucket) }()

	originInput := `
	{
		"callbackUrl" : "` + env.callbackUrl + `", 
		"callbackBody" : "{\"bucket\": ${bucket}, \"object\": ${object}, \"key1\": ${x:key1}}", 
		"callbackBodyType" : "application/json"                
	}`

	originVarInput := `
	{
		"x:key1" : "ceshi"
	}`

	out, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, Callback: base64.StdEncoding.EncodeToString([]byte(originInput)), CallbackVar: base64.StdEncoding.EncodeToString([]byte(originVarInput))},
		Content:             strings.NewReader(randomString(6)),
	})
	require.Nil(t, err)
	require.NotNil(t, out)
	require.NotEqual(t, out.CallbackResult, "")
	require.Contains(t, out.CallbackResult, "ok")

	partKey := "part-" + randomString(6)
	initPart, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: partKey})
	require.Nil(t, err)
	_, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: partKey, UploadID: initPart.UploadID, PartNumber: 1},
		Content:              strings.NewReader(randomString(1024 * 1024 * 5)),
	})
	require.Nil(t, err)
	completeOut, err := client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:      bucket,
		Key:         partKey,
		CompleteAll: true,
		UploadID:    initPart.UploadID,
		Callback:    base64.StdEncoding.EncodeToString([]byte(originInput)),
		CallbackVar: base64.StdEncoding.EncodeToString([]byte(originVarInput)),
	})
	require.Nil(t, err)
	require.NotEqual(t, completeOut.CallbackResult, "")
	require.Contains(t, completeOut.CallbackResult, "ok")

	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	file.Sync()

	putout, err := client.PutObjectFromFile(context.Background(), &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      bucket,
			Key:         key,
			Callback:    base64.StdEncoding.EncodeToString([]byte(originInput)),
			CallbackVar: base64.StdEncoding.EncodeToString([]byte(originVarInput)),
		},
		FilePath: fileName,
	})
	require.Nil(t, err)
	require.NotEqual(t, putout.CallbackResult, "")
	require.Contains(t, putout.CallbackResult, "ok")

}

func TestCallbackErr(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("callback")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
		key    = randomString(6)
	)
	defer func() { cleanBucket(t, client, bucket) }()

	originInput := `
	{
		"callbackUrl" : "` + env.callbackUrl + `", 
		"callbackBody" : "{\"bucket\": ${bucket, \"object\": ${object}, \"key1\": ${x:key1}}", 
		"callbackBodyType" : "application/json"                
	}`

	originVarInput := `
	{
		"x:key1" : "ceshi"
	}`

	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, Callback: base64.StdEncoding.EncodeToString([]byte(originInput)), CallbackVar: base64.StdEncoding.EncodeToString([]byte(originVarInput))},
		Content:             strings.NewReader(randomString(6)),
	})
	require.NotNil(t, err)
	sErr := err.(*tos.TosServerError)
	require.Equal(t, sErr.StatusCode, http.StatusNonAuthoritativeInfo)
	t.Log(err.Error())

	partKey := "part-" + randomString(6)
	initPart, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: partKey})
	require.Nil(t, err)
	_, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: bucket, Key: partKey, UploadID: initPart.UploadID, PartNumber: 1},
		Content:              strings.NewReader(randomString(1024 * 1024 * 5)),
	})
	require.Nil(t, err)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:      bucket,
		Key:         partKey,
		CompleteAll: true,
		UploadID:    initPart.UploadID,
		Callback:    base64.StdEncoding.EncodeToString([]byte(originInput)),
		CallbackVar: base64.StdEncoding.EncodeToString([]byte(originVarInput)),
	})
	require.NotNil(t, err)
	sErr = err.(*tos.TosServerError)
	require.Equal(t, sErr.StatusCode, http.StatusNonAuthoritativeInfo)
	t.Log(err.Error())
	require.Equal(t, sErr.Code, "CallbackFailed")

}
