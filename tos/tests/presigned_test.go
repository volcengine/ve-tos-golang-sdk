package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestPreSignedURLWithExpires(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("pre-signed-url-expires")
		cli    = env.prepareClient("")
		client = &http.Client{}
	)
	defer cleanBucket(t, cli, bucket)

	url, err := cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Expires:    1000,
	})
	require.Nil(t, err)
	req, _ := http.NewRequest(http.MethodPut, url.SignedUrl, nil)
	res, err := client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)

	// put object expire
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        "put-key",
		Expires:    2,
	})
	require.Nil(t, err)
	time.Sleep(time.Second * 3)
	req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(1024)))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 403, res.StatusCode)

	// put object without expires
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        "put-key",
		Expires:    7 * 24 * 60 * 60, // 604800
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(1024)))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)

}

func TestPreSignedURL(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("pre-signed-url")
		cli    = env.prepareClient("")
		client = &http.Client{}
	)
	defer cleanBucket(t, cli, bucket)
	// create bucket
	url, err := cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
	})
	require.Nil(t, err)
	req, _ := http.NewRequest(http.MethodPut, url.SignedUrl, nil)
	res, err := client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// put object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        "put-key",
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(4096)))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// head object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodHead,
		Bucket:     bucket,
		Key:        "put-key",
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodHead, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// get object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodGet,
		Bucket:     bucket,
		Key:        "put-key",
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodGet, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// delete object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodDelete,
		Bucket:     bucket,
		Key:        "put-key",
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodDelete, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 204, res.StatusCode)
	// create multipart upload
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPost,
		Bucket:     bucket,
		Key:        "multipart",
		Query:      map[string]string{"uploads": ""},
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPost, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	var output tos.CreateMultipartUploadV2Output
	jsonBytes, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	json.Unmarshal(jsonBytes, &output)
	// upload 3 parts
	type Part struct {
		PartNumber int
		ETag       string
	}
	type Parts struct {
		Parts []Part
	}
	parts := make([]Part, 0, 3)
	for i := 1; i <= 3; i++ {
		url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
			HTTPMethod: http.MethodPut,
			Bucket:     bucket,
			Key:        "multipart",
			Query:      map[string]string{"uploadId": output.UploadID, "partNumber": fmt.Sprint(i)},
		})
		require.Nil(t, err)
		req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(5*1024*1024)))
		res, err = client.Do(req)
		require.Nil(t, err)
		require.Equal(t, 200, res.StatusCode)
		parts = append(parts, Part{
			PartNumber: i,
			ETag:       res.Header.Get("ETag"),
		})
	}
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPost,
		Bucket:     bucket,
		Key:        "multipart",
		Query:      map[string]string{"uploadId": output.UploadID},
	})
	// cli.CompleteMultipartUploadV2()
	require.Nil(t, err)
	partsBytes, err := json.Marshal(Parts{Parts: parts})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPost, url.SignedUrl, bytes.NewReader(partsBytes))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)

}

func TestPreSignedURLEndpoint(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("pre-signed-url-endpoint")
		cli      = env.prepareClient("")
		client   = &http.Client{}
		endpoint = os.Getenv("TOS_GO_SDK_ALTERNATIVE_ENDPOINT")
	)
	defer func() {
		env.endpoint = endpoint
		cli = env.prepareClient("")
		cleanBucket(t, cli, bucket)
	}()
	// create bucket
	url, err := cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodPut,
		Bucket:              bucket,
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ := http.NewRequest(http.MethodPut, url.SignedUrl, nil)
	res, err := client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	key := "put/key/test"
	// put object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodPut,
		Bucket:              bucket,
		Key:                 key,
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	require.Contains(t, url.SignedUrl, key)
	req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(4096)))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// head object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodHead,
		Bucket:              bucket,
		Key:                 key,
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodHead, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// get object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodGet,
		Bucket:              bucket,
		Key:                 key,
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodGet, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// delete object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodDelete,
		Bucket:              bucket,
		Key:                 key,
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodDelete, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 204, res.StatusCode)
	// create multipart upload
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodPost,
		Bucket:              bucket,
		Key:                 "multipart",
		Query:               map[string]string{"uploads": ""},
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPost, url.SignedUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	var output tos.CreateMultipartUploadV2Output
	jsonBytes, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	json.Unmarshal(jsonBytes, &output)
	// upload 3 parts
	type Part struct {
		PartNumber int
		ETag       string
	}
	type Parts struct {
		Parts []Part
	}
	parts := make([]Part, 0, 3)
	for i := 1; i <= 3; i++ {
		url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
			HTTPMethod:          http.MethodPut,
			Bucket:              bucket,
			Key:                 "multipart",
			Query:               map[string]string{"uploadId": output.UploadID, "partNumber": fmt.Sprint(i)},
			AlternativeEndpoint: endpoint,
		})
		require.Nil(t, err)
		req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(5*1024*1024)))
		res, err = client.Do(req)
		require.Nil(t, err)
		require.Equal(t, 200, res.StatusCode)
		parts = append(parts, Part{
			PartNumber: i,
			ETag:       res.Header.Get("ETag"),
		})
	}
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodPost,
		Bucket:              bucket,
		Key:                 "multipart",
		Query:               map[string]string{"uploadId": output.UploadID},
		AlternativeEndpoint: endpoint,
	})
	// cli.CompleteMultipartUploadV2()
	require.Nil(t, err)
	partsBytes, err := json.Marshal(Parts{Parts: parts})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPost, url.SignedUrl, bytes.NewReader(partsBytes))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
}

func newPostRequest(bucket, key string, endpoint string, acl enum.ACLType, input *tos.PreSingedPostSignatureOutput) (*http.Request, string, error) {
	url := fmt.Sprintf("http://%s.%s", bucket, endpoint)
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	w.WriteField("key", key)
	w.WriteField("x-tos-algorithm", input.Algorithm)
	w.WriteField("x-tos-date", input.Date)
	w.WriteField("x-tos-credential", input.Credential)
	w.WriteField("policy", input.Policy)
	w.WriteField("x-tos-signature", input.Signature)
	if acl != "" {
		w.WriteField("x-tos-acl", string(acl))
	}
	fileWrite, _ := w.CreateFormFile("file", "my.test")
	value := randomString(1024)
	fileWrite.Write([]byte(value))
	w.Close()
	request, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, "", err
	}
	md5 := md5s(value)

	request.Header.Set("Content-Type", w.FormDataContentType())
	return request, md5, nil
}

func TestPreSignedPostSignature(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("pre-post-signed")
		cli    = env.prepareClient(bucket)
		client = &http.Client{}
		key    = randomString(6)
		ctx    = context.Background()
	)
	defer cleanBucket(t, cli, bucket)
	// base upload
	res, err := cli.PreSignedPostSignature(ctx, &tos.PreSingedPostSignatureInput{
		Bucket:  bucket,
		Key:     key,
		Expires: 3600,
	})
	require.Nil(t, err)
	fmt.Println(res)
	request, md5, err := newPostRequest(bucket, key, env.endpoint, "", res)

	httpRes, err := client.Do(request)
	require.Nil(t, err)
	fmt.Println(httpRes)
	data, err := ioutil.ReadAll(httpRes.Body)
	fmt.Println(string(data))

	getRes, err := cli.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	assert.Nil(t, err)
	data, err = ioutil.ReadAll(getRes.Content)
	assert.Nil(t, err)
	assert.Equal(t, md5s(string(data)), md5)

	// with content length range
	key = randomString(6)
	res, err = cli.PreSignedPostSignature(ctx, &tos.PreSingedPostSignatureInput{
		Bucket:  bucket,
		Key:     key,
		Expires: 3600,
		ContentLengthRange: &tos.ContentLengthRange{
			RangeStart: 50,
			RangeEnd:   1025,
		},
	})
	request, md5, err = newPostRequest(bucket, key, env.endpoint, "", res)
	httpRes, err = client.Do(request)
	assert.Nil(t, err)
	getRes, err = cli.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	assert.Nil(t, err)
	data, err = ioutil.ReadAll(getRes.Content)
	assert.Nil(t, err)
	assert.Equal(t, md5s(string(data)), md5)

	// expires
	res, err = cli.PreSignedPostSignature(ctx, &tos.PreSingedPostSignatureInput{
		Bucket:  bucket,
		Key:     key,
		Expires: 1,
	})
	require.Nil(t, err)
	time.Sleep(time.Second * 4)
	request, md5, err = newPostRequest(bucket, key, env.endpoint, "", res)

	httpRes, err = client.Do(request)
	assert.Nil(t, err)
	assert.Equal(t, httpRes.StatusCode, http.StatusForbidden)

	// exceed content range
	key = randomString(6)
	res, err = cli.PreSignedPostSignature(ctx, &tos.PreSingedPostSignatureInput{
		Bucket:  bucket,
		Key:     key,
		Expires: 3600,
		ContentLengthRange: &tos.ContentLengthRange{
			RangeStart: 50,
			RangeEnd:   1023,
		},
	})
	request, md5, err = newPostRequest(bucket, key, env.endpoint, "", res)
	httpRes, err = client.Do(request)
	assert.Nil(t, err)
	getRes, err = cli.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	assert.NotNil(t, err)
}

func TestPreSignedPostSignatureWithCondition(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("pre-post-signed-condition")
		cli    = env.prepareClient(bucket)
		client = &http.Client{}
		ctx    = context.Background()
	)
	defer cleanBucket(t, cli, bucket)
	// base upload
	keyPrefix := "post-"
	key := keyPrefix + randomString(6)

	operator := "starts-with"
	res, err := cli.PreSignedPostSignature(ctx, &tos.PreSingedPostSignatureInput{
		Bucket:  bucket,
		Key:     key,
		Expires: 3600,
		Conditions: []tos.PostSignatureCondition{{
			Key:   "x-tos-acl",
			Value: "public-read",
		}, {
			Key:      "key",
			Value:    keyPrefix,
			Operator: &operator,
		}},
	})
	require.Nil(t, err)
	fmt.Println(res)
	request, md5, err := newPostRequest(bucket, key, env.endpoint, enum.ACLPublicRead, res)
	httpRes, err := client.Do(request)
	require.Nil(t, err)
	assert.Equal(t, httpRes.StatusCode, 204)

	getRes, err := cli.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	assert.Nil(t, err)
	data, err := ioutil.ReadAll(getRes.Content)
	require.Equal(t, md5, md5s(string(data)))

	// Invalid ACl
	request, md5, err = newPostRequest(bucket, key, env.endpoint, enum.ACLPrivate, res)
	httpRes, err = client.Do(request)
	require.Nil(t, err)
	assert.Equal(t, httpRes.StatusCode, 403)

	// Invalid Key
	request, md5, err = newPostRequest(bucket, randomString(6), env.endpoint, enum.ACLPublicRead, res)
	httpRes, err = client.Do(request)
	require.Nil(t, err)
	assert.Equal(t, httpRes.StatusCode, 403)

}

func TestPreSignedPolicyURLWithExpires(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("pre-signed-policy-url")
		cli    = env.prepareClient(bucket)
		client = &http.Client{}
		ctx    = context.Background()
	)
	defer cleanBucket(t, cli, bucket)
	keyPrefix := "policy-"
	key := keyPrefix + randomString(6)
	key1 := keyPrefix + randomString(6)
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
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(4096), res.ContentLength)

	req, _ = http.NewRequest(http.MethodHead, getUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(4096), res.ContentLength)

	// get test based policy url for key1
	getUrl1 := output.GetSignedURLForGetOrHead(key1, nil)
	req, _ = http.NewRequest(http.MethodHead, getUrl1, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(2048), res.ContentLength)

	// prefix must be subsequence for policy, prefix set "policy", but policy starts-with "policy-"
	listUrl := output.GetSignedURLForList(map[string]string{
		"prefix": "policy",
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 403, res.StatusCode)

	// list test based policy url, can list key key1
	listUrl = output.GetSignedURLForList(map[string]string{
		"prefix": keyPrefix,
	})
	req, _ = http.NewRequest(http.MethodGet, listUrl, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// Unmarshal ListObjectsOutput
	data, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	data = bytes.TrimSpace(data)
	jsonOut := tos.ListObjectsOutput{}
	err = json.Unmarshal(data, &jsonOut)
	require.Nil(t, err)
	require.Equal(t, len(jsonOut.Contents), 2)

	// head test based policy url for key2
	getUrl2 := output.GetSignedURLForGetOrHead(key2, nil)
	req, _ = http.NewRequest(http.MethodHead, getUrl2, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(1024), res.ContentLength)

	// get test based policy url for key3
	getUrl3 := output.GetSignedURLForGetOrHead(key3, nil)
	req, _ = http.NewRequest(http.MethodGet, getUrl3, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, int64(1024), res.ContentLength)
	require.Equal(t, "binary/octet-stream", res.Header.Get("Content-Type"))

	// get with additionQuery test based policy url for key3
	getUrl3 = output.GetSignedURLForGetOrHead(key3, map[string]string{"response-content-type": "text/plain"})
	req, _ = http.NewRequest(http.MethodGet, getUrl3, nil)
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	require.Equal(t, "text/plain", res.Header.Get("Content-Type"))
}
