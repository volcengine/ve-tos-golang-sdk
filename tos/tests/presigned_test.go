package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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

	// exceed max expires
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Key:        "put-key",
		Expires:    6048000,
	})
	require.NotNil(t, err)

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
	// put object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodPut,
		Bucket:              bucket,
		Key:                 "put-key",
		AlternativeEndpoint: endpoint,
	})
	require.Nil(t, err)
	req, _ = http.NewRequest(http.MethodPut, url.SignedUrl, strings.NewReader(randomString(4096)))
	res, err = client.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// head object
	url, err = cli.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod:          http.MethodHead,
		Bucket:              bucket,
		Key:                 "put-key",
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
		Key:                 "put-key",
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
		Key:                 "put-key",
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
