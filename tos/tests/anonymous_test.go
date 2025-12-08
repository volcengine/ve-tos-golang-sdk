package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestAnonymous(t *testing.T) {
	var (
		env       = newTestEnv(t)
		bucket    = generateBucketName("anonymous-bucket")
		oldClient = env.prepareClient(bucket)
		ctx       = context.Background()
	)
	defer func() {
		cleanBucket(t, oldClient, bucket)
	}()
	policy := map[string]interface{}{
		"Statement": []map[string]interface{}{
			{
				"Sid":       "internal public",
				"Effect":    "Allow",
				"Action":    []string{"*"},
				"Principal": "*",
				"Resource": []string{
					fmt.Sprintf("trn:tos:::%s/*", bucket),
					fmt.Sprintf("trn:tos:::%s", bucket),
				},
			},
		}}
	data, err := json.Marshal(policy)
	require.Nil(t, err)
	_, err = oldClient.PutBucketPolicyV2(context.Background(), &tos.PutBucketPolicyV2Input{
		Bucket: bucket,
		Policy: string(data),
	})
	require.Nil(t, err)
	time.Sleep(60 * time.Second)

	// 匿名读写
	key := randomString(7)
	client, err := tos.NewClientV2(env.endpoint, tos.WithRegion(env.region))
	require.Nil(t, err)
	_, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(6)),
	})
	require.Nil(t, err)
	_, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)

	_, err = client.CopyObject(ctx, &tos.CopyObjectInput{
		Bucket:    bucket,
		Key:       randomString(6),
		SrcBucket: bucket,
		SrcKey:    key,
	})
	require.Nil(t, err)

	newKey := randomString(7)

	multi, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: newKey})
	require.Nil(t, err)
	partOut, err := client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:     bucket,
			Key:        newKey,
			UploadID:   multi.UploadID,
			PartNumber: 1,
		},
		Content: strings.NewReader(randomString(1024)),
	})
	require.Nil(t, err)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      newKey,
		UploadID: multi.UploadID,
		Parts: []tos.UploadedPartV2{{
			PartNumber: 1,
			ETag:       partOut.ETag,
		}},
	})
	require.Nil(t, err)
	_, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    newKey,
	})
	require.Nil(t, err)

	appendKey := randomString(6)
	appendLength := 1024
	_, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:  bucket,
		Key:     appendKey,
		Offset:  0,
		Content: strings.NewReader(randomString(appendLength)),
	})
	require.Nil(t, err)

	preSignUrl, err := client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodGet,
		Bucket:     bucket,
		Key:        appendKey,
		Expires:    3600,
	})
	require.Nil(t, err)
	resp, err := http.Get(preSignUrl.SignedUrl)
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)
	require.Equal(t, resp.ContentLength, int64(appendLength))

	newKey = randomString(8)
	preSignUrl, err = client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        newKey,
		Expires:    3600,
	})
	req, err := http.NewRequest(http.MethodPut, preSignUrl.SignedUrl, strings.NewReader(randomString(appendLength)))
	require.Nil(t, err)
	resp, err = http.DefaultClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, resp.StatusCode, 200)
	operatorEq := "eq"
	policyUrl, err := client.PreSignedPolicyURL(context.Background(), &tos.PreSignedPolicyURLInput{
		Bucket:  bucket,
		Expires: 3600,
		Conditions: []tos.PolicySignatureCondition{{
			Key:      "key",
			Value:    newKey,
			Operator: &operatorEq,
		}}})

	require.Nil(t, err)
	getUrl := policyUrl.GetSignedURLForGetOrHead(newKey, nil)
	req, _ = http.NewRequest(http.MethodGet, getUrl, nil)
	res, err := http.DefaultClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(appendLength), res.ContentLength)

}
