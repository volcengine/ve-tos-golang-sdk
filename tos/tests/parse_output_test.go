package tests

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"net/http"
	"strings"
	"testing"
)

func TestParseOutput(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("request-by-signed-url")
		cli    = env.prepareClient(bucket)
		ctx    = context.Background()
		client = &http.Client{}
	)

	defer cleanBucket(t, cli, bucket)
	keyPrefix := "policy-"
	key := keyPrefix + "0" + randomString(5)
	key1 := keyPrefix + "1" + randomString(5)
	key2 := "policy/" + randomString(6)
	key3 := "policy/" + randomString(6)
	// put object for key key1 key2 key3
	put, err := cli.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(4096)),
	})
	checkSuccess(t, put, err, 200)
	put1, err := cli.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key1},
		Content:             strings.NewReader(randomString(2048)),
	})
	checkSuccess(t, put1, err, 200)
	put2, err := cli.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key2},
		Content:             strings.NewReader(randomString(1024)),
	})
	checkSuccess(t, put2, err, 200)
	put3, err := cli.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key3},
		Content:             strings.NewReader(randomString(1024)),
	})
	checkSuccess(t, put3, err, 200)

	// build policy url
	operatorSw := "starts-with"
	operatorEq := "eq"
	output, err := cli.PreSignedPolicyURL(ctx, &tos.PreSingedPolicyURLInput{
		Bucket:  bucket,
		Expires: 1000,
		Conditions: []tos.PolicySignatureCondition{{
			Key:      "key",
			Value:    keyPrefix,
			Operator: &operatorSw,
		}, {
			Key:      "key",
			Value:    key2,
			Operator: &operatorEq,
		}, {
			Key:   "key",
			Value: key3,
		}},
	})
	require.Nil(t, err)

	// head&get object test based policy url for key
	getUrl := output.GetSignedURLForGetOrHead(key, nil)
	req, _ := http.NewRequest(http.MethodGet, getUrl, nil)
	res, err := client.Do(req)
	require.Nil(t, err)
	get, err := tos.ParseGetObjectV2Output(res, 200)
	require.Nil(t, err)
	require.Equal(t, 200, get.StatusCode)
	require.Equal(t, int64(4096), get.ContentLength)

	req, _ = http.NewRequest(http.MethodHead, getUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	head, err := tos.ParseHeadObjectV2Output(res)
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	require.Equal(t, int64(4096), head.ContentLength)

	// get test based policy url for key1
	getUrl1 := output.GetSignedURLForGetOrHead(key1, nil)
	req, _ = http.NewRequest(http.MethodHead, getUrl1, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	head, err = tos.ParseHeadObjectV2Output(res)
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	require.Equal(t, int64(2048), head.ContentLength)

	// prefix must be subsequence for policy, prefix set "policy", but policy starts-with "policy-"
	listUrl := output.GetSignedURLForList(map[string]string{
		"prefix": "policy",
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	list, err := tos.ParseListObjectsType2Output(res)
	require.NotNil(t, err)

	// list test based policy url, can list key key1
	listUrl = output.GetSignedURLForList(map[string]string{
		"prefix": keyPrefix,
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	list, err = tos.ParseListObjectsType2Output(res)
	require.Nil(t, err)
	require.Equal(t, 200, list.StatusCode)
	require.Equal(t, len(list.Contents), 2)

	// list test based policy url, can list key key1
	listUrl = output.GetSignedURLForList(map[string]string{
		"prefix": keyPrefix,
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	listV2, err := tos.ParseListObjectsV2Output(res)
	require.Nil(t, err)
	require.Equal(t, 200, listV2.StatusCode)
	require.Equal(t, len(listV2.Contents), 2)

	// list versions test based policy url, can list key key1
	listUrl = output.GetSignedURLForList(map[string]string{
		"prefix":   keyPrefix,
		"versions": "",
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	listVersions, err := tos.ParseListObjectVersionsV2Output(res)
	require.Nil(t, err)
	require.Equal(t, 200, listVersions.StatusCode)
	require.Equal(t, len(listVersions.Versions), 2)

	// get test based policy url for key3
	getUrl3 := output.GetSignedURLForGetOrHead(key3, nil)
	req, _ = http.NewRequest(http.MethodGet, getUrl3, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	get, err = tos.ParseGetObjectV2Output(res, 200)
	require.Nil(t, err)
	require.Equal(t, 200, get.StatusCode)
	require.Equal(t, int64(1024), get.ContentLength)
	require.Equal(t, "binary/octet-stream", get.Header.Get("Content-Type"))

	// get with additionQuery test based policy url for key3
	getUrl3 = output.GetSignedURLForGetOrHead(key3, map[string]string{"response-content-type": "text/plain"})
	req, _ = http.NewRequest(http.MethodGet, getUrl3, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	get, err = tos.ParseGetObjectV2Output(res, 200)
	require.Nil(t, err)
	require.Equal(t, 200, get.StatusCode)
	require.Equal(t, "text/plain", get.Header.Get("Content-Type"))
}
