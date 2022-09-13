package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

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
	//cli.CompleteMultipartUploadV2()
	require.Nil(t, err)
	partsBytes, err := json.Marshal(Parts{Parts: parts})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPost, url.SignedUrl, bytes.NewReader(partsBytes))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
}
