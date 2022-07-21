package tests

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/tos"
)

const testPrefix = "g0lan9-5dk-t35ts"

var LongTimeOutClientOption = []tos.ClientOption{tos.WithSocketTimeout(360*time.Second, 360*time.Second), tos.WithRequestTimeout(360 * time.Second)}

type waitSetting struct {
	duration    time.Duration
	maxWaitTime time.Duration
}

type waitOption func(*waitSetting)

func withWaitDuration(duration time.Duration) waitOption {
	return func(setting *waitSetting) {
		setting.duration = duration
	}
}

func withMaxWaitTime(maxWaitTime time.Duration) waitOption {
	return func(setting *waitSetting) {
		setting.maxWaitTime = maxWaitTime
	}
}

func newWaitSetting(options ...waitOption) *waitSetting {
	setting := &waitSetting{
		duration:    3 * time.Second,
		maxWaitTime: 30 * time.Second,
	}
	for _, option := range options {
		option(setting)
	}
	return setting
}

func waitUntilBucketExist(t *testing.T, cli *tos.ClientV2, bucket string, options ...waitOption) {
	var (
		err     error
		now     = time.Now()
		setting = newWaitSetting(options...)
	)
	for {
		_, err = cli.HeadBucket(context.Background(), &tos.HeadBucketInput{Bucket: bucket})
		if err == nil {
			break
		}
		if tos.StatusCode(err) != 404 {
			break
		}
		if time.Now().After(now.Add(setting.maxWaitTime)) {
			err = errors.New("wait bucket exists timeout")
			break
		}
		time.Sleep(setting.duration)
	}
	require.Nil(t, err)
}

func waitUntilObjectExist(t *testing.T, cli *tos.ClientV2, bucket, key string, options ...waitOption) {
	var (
		err     error
		now     = time.Now()
		setting = newWaitSetting(options...)
	)
	for {
		_, err = cli.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
		if err == nil {
			break
		}
		if tos.StatusCode(err) != 404 {
			break
		}
		if time.Now().After(now.Add(setting.maxWaitTime)) {
			err = errors.New("wait object exists timeout")
			break
		}
		time.Sleep(setting.duration)
	}
	require.Nil(t, err)
}

func waitUntilBucketNoneExist(t *testing.T, cli *tos.ClientV2, bucket string, options ...waitOption) {
	var (
		err     error
		now     = time.Now()
		setting = newWaitSetting(options...)
	)
	for {
		_, err = cli.HeadBucket(context.Background(), &tos.HeadBucketInput{Bucket: bucket})
		if err == nil {
			break
		}
		if tos.StatusCode(err) != 404 {
			break
		}
		if time.Now().After(now.Add(setting.maxWaitTime)) {
			err = errors.New("wait bucket none exists timeout")
			break
		}
		time.Sleep(setting.duration)
	}
	require.Nil(t, err)
}

func waitUntilObjectNoneExist(t *testing.T, cli *tos.ClientV2, bucket, key string, options ...waitOption) {
	var (
		err     error
		now     = time.Now()
		setting = newWaitSetting(options...)
	)
	for {
		_, err = cli.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
		if err == nil {
			break
		}
		if tos.StatusCode(err) != 404 {
			break
		}
		if time.Now().After(now.Add(setting.maxWaitTime)) {
			err = errors.New("wait object exists timeout")
			break
		}
		time.Sleep(setting.duration)
	}
	require.Nil(t, err)
}

func isTosServerError(err error) bool {
	_, ok := err.(*tos.TosServerError)
	return ok
}

func isTosClientError(err error) bool {
	_, ok := err.(*tos.TosClientError)
	return ok
}

// generate a random string include only [a-b]|[1-9]
func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	b := make([]byte, n)
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

func md5s(s string) string {
	r := md5.Sum([]byte(s))
	return base64.StdEncoding.EncodeToString(r[:])
}

func generateBucketName(bucket string) string {
	return strings.Join([]string{testPrefix, randomString(8), bucket}, "-")
}

func cleanBucket(t *testing.T, client *tos.ClientV2, bucket string) {
	del, err := client.DeleteBucket(context.Background(), &tos.DeleteBucketInput{Bucket: bucket})
	if err == nil {
		require.Equal(t, http.StatusNoContent, del.StatusCode)
		return
	}
	if tos.StatusCode(err) == http.StatusNotFound {
		return
	}
	// the bucket is not clean
	if tos.StatusCode(err) == http.StatusConflict {
		// delete all multi version objects. Do this first to avoid leaving a delete marker.
		listMultiVersion, err := client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
			Bucket:                  bucket,
			ListObjectVersionsInput: tos.ListObjectVersionsInput{},
		})
		require.Nil(t, err)
		willSleep := len(listMultiVersion.Versions) > 1
		deleteMultiInput := tos.DeleteMultiObjectsInput{
			Bucket:  bucket,
			Objects: make([]tos.ObjectTobeDeleted, 0),
			Quiet:   true,
		}
		for _, version := range listMultiVersion.Versions {
			deleteMultiInput.Objects = append(deleteMultiInput.Objects, tos.ObjectTobeDeleted{
				Key:       version.Key,
				VersionID: version.VersionID,
			})
		}
		for _, marker := range listMultiVersion.DeleteMarkers {
			deleteMultiInput.Objects = append(deleteMultiInput.Objects, tos.ObjectTobeDeleted{
				Key:       marker.Key,
				VersionID: marker.VersionID,
			})
		}
		if willSleep {
			time.Sleep(time.Second * 30)
		}
		if len(deleteMultiInput.Objects) > 0 {
			_, err = client.DeleteMultiObjects(context.Background(), &deleteMultiInput)
			require.Nil(t, err)
		}
		// delete all objects
		list, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{Bucket: bucket})
		require.Nil(t, err)
		for _, object := range list.Contents {
			del, err := client.DeleteObjectV2(context.Background(), &tos.DeleteObjectV2Input{Bucket: bucket, Key: object.Key})
			checkSuccess(t, del, err, http.StatusNoContent)
		}
		// delete all uncompleted MultipartUpload
		listMulti, err := client.ListMultipartUploadsV2(context.Background(), &tos.ListMultipartUploadsV2Input{
			Bucket: bucket,
		})
		require.Nil(t, err)
		for _, upload := range listMulti.Uploads {
			abort, err := client.AbortMultipartUpload(context.Background(), &tos.AbortMultipartUploadInput{
				Bucket:   bucket,
				Key:      upload.Key,
				UploadID: upload.UploadID,
			})
			require.Nil(t, err)
			require.Equal(t, 204, abort.StatusCode)
		}
		require.Equal(t, http.StatusOK, listMulti.StatusCode)
		// now, the bucket should be clean
		del, err = client.DeleteBucket(context.Background(), &tos.DeleteBucketInput{Bucket: bucket})
		require.Nil(t, err)
		require.Equal(t, http.StatusNoContent, del.StatusCode)
		return
	}
	// something wrong
	require.Equal(t, http.StatusOK, tos.StatusCode(err))
}

func checkBucketMeta(t *testing.T, client *tos.ClientV2, bucket string, expect *tos.HeadBucketOutput) {
	head, err := client.HeadBucket(context.Background(), &tos.HeadBucketInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, http.StatusOK, head.StatusCode)
	require.Equal(t, expect.Region, head.Region)
	require.Equal(t, expect.StorageClass, head.StorageClass)
}

func testInvalidBucketName(t *testing.T, bucket string) {
	require.NotNil(t, tos.IsValidBucketName(bucket))
}

func testInvalidObjectKey(t *testing.T, client *tos.ClientV2, key string) {
	value := "test-value"
	bucket := "test-bucket"
	put, err := client.PutObject(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	require.NotNil(t, err)
	require.Nil(t, put)
	require.True(t, isTosClientError(err))
}

func testValidObjectKey(t *testing.T, client *tos.ClientV2, bucket string, key string) {
	value := ""
	put, err := client.PutObject(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
}

func putRandomObject(t *testing.T, client *tos.ClientV2, bucket string, key string, size int) {
	value := randomString(size)
	put, err := client.PutObject(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	require.Nil(t, err)
	require.Equal(t, put.StatusCode, 200)
	waitUntilObjectExist(t, client, bucket, key)
}

func cleanTestFile(t *testing.T, fileName string) {
	err := os.Remove(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		require.Nil(t, err)
	}
}

func enableMultiVersion(t *testing.T, client *tos.ClientV2, bucket string) {
	httpClient := &http.Client{}
	type VersioningConfiguration struct {
		Status string `json:"Status"`
	}
	var setting = VersioningConfiguration{Status: "Enabled"}
	settingBytes, _ := json.Marshal(setting)
	url, err := client.PreSignedURL(&tos.PreSignedURLInput{
		HTTPMethod: http.MethodPut,
		Bucket:     bucket,
		Query:      map[string]string{"versioning": ""},
	})
	req, err := http.NewRequest(http.MethodPut, url.SignedUrl, bytes.NewReader(settingBytes))
	require.Nil(t, err)
	res, err := httpClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, res.StatusCode)
	// sleep cache time
	time.Sleep(time.Second * 30)
}

func checkSuccess(t *testing.T, data interface{}, err error, expectCode int) {
	require.Nil(t, err)
	require.NotNil(t, data)
	value := reflect.ValueOf(data).Elem()
	status := value.FieldByName("StatusCode")
	require.Equal(t, expectCode, int(status.Int()))
}

func checkFail(t *testing.T, data interface{}, err error, expectCode int) {
	require.NotNil(t, err)
	require.Nil(t, data)
	require.Equal(t, expectCode, tos.StatusCode(err))
}
