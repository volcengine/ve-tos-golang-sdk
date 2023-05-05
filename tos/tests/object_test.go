package tests

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/codes"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestHeadNoneExistentObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-basic")
		client = env.prepareClient(bucket)
		key    = "none-exist-key"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	head, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.NotNil(t, err)
	require.Nil(t, head)
	terr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.True(t, strings.Contains(terr.Message, "unexpected"))
}

func TestGetNoneExistentObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-basic")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	_, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    "not-exist-key",
	})
	require.NotNil(t, err)
	terr, ok := err.(*tos.TosServerError)
	require.True(t, ok)
	require.Equal(t, codes.NoSuchKey, terr.Code)
}

func TestGetObjectWithCloseBody(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("get-basic")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	key := "get-" + randomString(5)
	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(randomString(20 * 1024 * 1024)),
	})
	require.Nil(t, err)
	getOut, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	bufSize := 50
	buf := make([]byte, bufSize)
	n, err := io.ReadFull(getOut.Content, buf)
	require.Nil(t, err)
	require.Equal(t, n, bufSize)
	start := time.Now()
	getOut.Content.Close()
	now := time.Now()
	t.Log("cost:", now.Sub(start).Milliseconds(), " ms")
	require.True(t, time.Now().Sub(start).Milliseconds() < int64(5*time.Millisecond))

	buf = make([]byte, bufSize)
	n, err = io.ReadFull(getOut.Content, buf)
	require.NotNil(t, err)
	require.Equal(t, n, 0)
}

func TestPutBasic(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-basic")
		client = env.prepareClient(bucket)
		key    = "key123"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putRandomObject(t, client, bucket, key, 4096)
}

func TestPutLargeObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-basic")
		client = env.prepareClient(bucket, LongTimeOutClientOption...)
		key    = "key123"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putRandomObject(t, client, bucket, key, 100*4096*4096)
}

func TestPutWithAllParams(t *testing.T) {
	var (
		env          = newTestEnv(t)
		bucket       = generateBucketName("put-all-params")
		client       = env.prepareClient(bucket)
		key          = "key123"
		value        = randomString(5 * 1024 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		ssecKey      = randomString(32)
		ssecCopyKey  = randomString(32)
		ssecMd5      = md5s(ssecKey)
		ssecCopyMd5  = md5s(ssecCopyKey)
		storageClass = enum.StorageClassIa
		// sse          = "AES256"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:             bucket,
			Key:                key,
			ContentMD5:         md5Sum,
			Expires:            expires,
			ACL:                acl,
			ContentDisposition: "中文测试",
			StorageClass:       storageClass,
			SSECAlgorithm:      "AES256",
			SSECKey:            base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:         ssecMd5,
			Meta:               meta,
		},
		Content: strings.NewReader(value),
	}
	put, err := client.PutObjectV2(context.Background(), input)
	require.Nil(t, err)
	require.NotNil(t, put)
	require.Equal(t, 200, put.StatusCode)
	head, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key, SSECAlgorithm: "AES256",
		SSECKey:    base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		SSECKeyMD5: ssecMd5})
	require.Nil(t, err)
	require.Equal(t, 200, head.StatusCode)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:        bucket,
		Key:           key,
		SSECAlgorithm: "AES256",
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		SSECKeyMD5:    ssecMd5,
	})
	buffer, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := head.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Equal(t, "中文测试", get.ContentDisposition)
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
	ctx := context.Background()
	copyKey := "ssec_copy_key"
	algorithm := "AES256"

	multi, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{Bucket: bucket, Key: copyKey,
		SSECAlgorithm: algorithm,
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecCopyKey)),
		SSECKeyMD5:    ssecCopyMd5})
	require.Nil(t, err)
	partOut, err := client.UploadPartCopyV2(ctx, &tos.UploadPartCopyV2Input{
		Bucket:                  bucket,
		Key:                     copyKey,
		UploadID:                multi.UploadID,
		PartNumber:              1,
		SrcBucket:               bucket,
		SrcKey:                  key,
		CopySourceSSECAlgorithm: algorithm,
		CopySourceSSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		CopySourceSSECKeyMD5:    ssecMd5,
		SSECKey:                 base64.StdEncoding.EncodeToString([]byte(ssecCopyKey)),
		SSECKeyMD5:              ssecCopyMd5,
		SSECAlgorithm:           algorithm,
	})
	require.Nil(t, err)
	_, err = client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{
		Bucket:   bucket,
		Key:      copyKey,
		UploadID: multi.UploadID,
		Parts: []tos.UploadedPartV2{{
			PartNumber: 1,
			ETag:       partOut.ETag,
		}},
	})
	require.Nil(t, err)
	obj, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket:        bucket,
		Key:           copyKey,
		SSECAlgorithm: "AES256",
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecCopyKey)),
		SSECKeyMD5:    ssecCopyMd5,
	})
	require.Nil(t, err)
	buffer, err = ioutil.ReadAll(obj.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))

	copyKey = randomString(6)
	_, err = client.CopyObject(ctx, &tos.CopyObjectInput{
		Bucket:                  bucket,
		Key:                     copyKey,
		SrcBucket:               bucket,
		SrcKey:                  key,
		CopySourceSSECAlgorithm: algorithm,
		CopySourceSSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		CopySourceSSECKeyMD5:    ssecMd5,
		SSECAlgorithm:           algorithm,
		SSECKey:                 base64.StdEncoding.EncodeToString([]byte(ssecCopyKey)),
		SSECKeyMD5:              ssecCopyMd5,
	})
	require.Nil(t, err)
	obj, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket:        bucket,
		Key:           copyKey,
		SSECAlgorithm: "AES256",
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecCopyKey)),
		SSECKeyMD5:    ssecCopyMd5,
	})
	require.Nil(t, err)
	buffer, err = ioutil.ReadAll(obj.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))
}

func TestPutEmptyObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-empty-object")
		key    = "key123"
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             nil,
	})
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	buffer, err := ioutil.ReadAll(get.Content)
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, len(buffer), 0)
	require.Nil(t, err)
}

func TestListObjects(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("list-objects")
		key1   = "key1"
		key2   = "key2"
		key3   = "key3"
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key1},
		Content:             strings.NewReader(randomString(4096)),
	})
	checkSuccess(t, put, err, 200)
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key2},
		Content:             strings.NewReader(randomString(4096)),
	})
	checkSuccess(t, put, err, 200)
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key3},
		Content:             strings.NewReader(randomString(4096)),
	})
	objects, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{
		Bucket: bucket,
	})
	checkSuccess(t, objects, err, 200)
	require.Equal(t, 3, len(objects.Contents))
}

func TestListObjectVersions(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("list-objects")
		key1   = "key-a-1"
		key2   = "key-a-2"
		key3   = "key-b-3"
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key1},
		Content:             strings.NewReader(randomString(4096)),
	})
	checkSuccess(t, put, err, 200)
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key2},
		Content:             strings.NewReader(randomString(4096)),
	})
	checkSuccess(t, put, err, 200)
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key3},
		Content:             strings.NewReader(randomString(4096)),
	})
	objects, err := client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
		Bucket: bucket,
	})
	checkSuccess(t, objects, err, 200)
	require.Equal(t, 3, len(objects.Versions))

	objects, err = client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
		Bucket: bucket,
		ListObjectVersionsInput: tos.ListObjectVersionsInput{
			MaxKeys: 2,
		},
	})
	checkSuccess(t, objects, err, 200)
	require.Equal(t, 2, len(objects.Versions))

	objects, err = client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
		Bucket: bucket,
		ListObjectVersionsInput: tos.ListObjectVersionsInput{
			Prefix: "key-a-",
		},
	})
	checkSuccess(t, objects, err, 200)
	require.Equal(t, 2, len(objects.Versions))
}

func TestCopyObject(t *testing.T) {
	var (
		env       = newTestEnv(t)
		bucket    = generateBucketName("copy-object")
		key       = "1.jpg"
		value     = "value123"
		copyedKey = "copyedKey"
		client    = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	copyRes, err := client.CopyObject(context.Background(), &tos.CopyObjectInput{
		Bucket:    bucket,
		Key:       copyedKey,
		SrcBucket: bucket,
		SrcKey:    key,
	})
	require.Nil(t, err)
	require.NotNil(t, copyRes.ETag)
	require.NotNil(t, copyRes.LastModified)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    copyedKey,
	})
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(get.Content)
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Nil(t, err)
	require.Equal(t, get.ContentType, "image/jpeg")
}

func TestCopyObjectContentType(t *testing.T) {
	var (
		env       = newTestEnv(t)
		bucket    = generateBucketName("copy-object")
		key       = "key123"
		value     = "value123"
		copyedKey = "copyedKey"
		client    = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	// Case 1: 源对象是 jpeg
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ContentType: "image/jpeg"},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	copyRes, err := client.CopyObject(context.Background(), &tos.CopyObjectInput{
		Bucket:      bucket,
		Key:         copyedKey,
		SrcBucket:   bucket,
		SrcKey:      key,
		ContentType: "image/jpeg",
	})
	require.Nil(t, err)
	require.NotNil(t, copyRes.ETag)
	require.NotNil(t, copyRes.LastModified)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    copyedKey,
	})
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(get.Content)
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Nil(t, err)
	require.Equal(t, get.ContentType, "image/jpeg")

	// Case2: Copy 时指定新的ContentType
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	copyRes, err = client.CopyObject(context.Background(), &tos.CopyObjectInput{
		Bucket:            bucket,
		Key:               copyedKey,
		SrcBucket:         bucket,
		SrcKey:            key,
		ContentType:       "image/jpeg",
		MetadataDirective: enum.MetadataDirectiveReplace,
	})
	require.Nil(t, err)
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    copyedKey,
	})
	require.Nil(t, err)
	require.Equal(t, get.ContentType, "image/jpeg")
}

func TestCopyObjectVersion(t *testing.T) {
	var (
		env       = newTestEnv(t)
		bucket    = generateBucketName("copy-object-version")
		key       = "key123"
		value     = "value123"
		copyedKey = "copyedKey"
		client    = env.prepareClient(bucket)
	)
	enableMultiVersion(t, client, bucket)
	time.Sleep(time.Minute)
	defer func() {
		cleanBucket(t, client, bucket)

	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	versionId := put.VersionID
	value2 := randomString(8)
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value2),
	})
	checkSuccess(t, put, err, 200)

	copyRes, err := client.CopyObject(context.Background(), &tos.CopyObjectInput{
		Bucket:       bucket,
		Key:          copyedKey,
		SrcVersionID: versionId,
		SrcBucket:    bucket,
		SrcKey:       key,
	})
	require.Nil(t, err)

	require.NotNil(t, copyRes.ETag)
	require.NotNil(t, copyRes.LastModified)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    copyedKey,
	})

	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(get.Content)
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Nil(t, err)

}

func TestValidObjectKey(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("test-invalid-object-key")
		client = env.prepareClient(bucket)
	)
	testValidObjectKey(t, client, bucket, ".")
	testValidObjectKey(t, client, bucket, "..")
}

func TestUnmatchedMD5(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("unmatched-md5")
		key    = "key123"
		value  = "value123"
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	}
	input.ContentMD5 = "unmatched md5"
	_, err := client.PutObjectV2(context.Background(), input)
	require.NotNil(t, err)
	require.True(t, isTosServerError(err))
}

func TestUrlEncodeChineseInMeta(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("url-encode-chinese-in-meta")
		client = env.prepareClient(bucket)
		key    = "key123"
		value  = "value123"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	meta := make(map[string]string)
	meta["中文开头的键"] = "中文值"
	meta["中文开头的键-test-key"] = "中文值-test-val"
	meta["test-key带中文的键"] = "test-val-中文值"
	meta["test-key-带中文的键"] = "test-val-中文值"
	// same key
	meta["test-key"] = "TEST-VAL"
	meta["TEST-KEY"] = "TEST-VAL"
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	}
	input.Meta = meta
	put, err := client.PutObjectV2(context.Background(), input)
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.NotNil(t, get)
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Nil(t, err)
}

func encryptAES256(key []byte, plaintext string) string {
	c, err := aes.NewCipher(key)
	_ = err
	out := make([]byte, len(plaintext))
	c.Encrypt(out, []byte(plaintext))
	return hex.EncodeToString(out)
}

func TestSSEC(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("supported-ssec")
		key     = "key123"
		value   = randomString(4 * 1024)
		ssecKey = randomString(32)
		ssecMd5 = md5s(ssecKey)
		client  = env.prepareClient(bucket)
	)
	value = encryptAES256([]byte(ssecKey), value)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
		},
		Content: strings.NewReader(value),
	})
	require.Nil(t, err)
	require.NotNil(t, put)
	require.Equal(t, 200, put.StatusCode)
	_, err = client.CopyObject(context.Background(), &tos.CopyObjectInput{Bucket: bucket, Key: key + "1", SrcBucket: bucket, SrcKey: key})
	if serr, ok := err.(*tos.TosServerError); ok {
		t.Log(serr.Code)
		t.Log(serr.Resource)
	}
	t.Log(err)
	// GetObjectV2 without SSEC will fail
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, get)
	require.NotNil(t, err)
	require.Equal(t, 400, tos.StatusCode(err))
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:        bucket,
		Key:           key,
		SSECAlgorithm: "AES256",
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		SSECKeyMD5:    ssecMd5,
	})
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, 200, get.StatusCode)

	buffer, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5s(value), md5s(string(buffer)))
}

func TestSSE(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("supported-sse")
		key    = "key123"
		value  = randomString(4 * 1024)
		md5Sum = md5s(value)
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			ServerSideEncryption: "AES256",
		},
		Content: strings.NewReader(value),
	})
	require.Nil(t, err)
	require.NotNil(t, put)
	require.Equal(t, 200, put.StatusCode)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.NotNil(t, get)
	require.Nil(t, err)
	require.Equal(t, 200, get.StatusCode)
	buffer, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))
}

func TestUnsupportedSSEC(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("unsupported-ssec")
		key     = "key123"
		value   = randomString(4 * 1024)
		ssecKey = randomString(16)
		ssecMd5 = md5s(ssecKey)
		client  = env.prepareClient(bucket)
	)
	value = encryptAES256([]byte(ssecKey), value)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			SSECAlgorithm: "AES128",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
		},
		Content: strings.NewReader(value),
	})
	require.Nil(t, put)
	require.Equal(t, tos.InvalidSSECAlgorithm, err)
}

func TestDeleteMultiObjects(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("delete-multi-objects")
		key    = "key"
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	for i := 0; i < 5; i++ {
		putRandomObject(t, client, bucket, key+strconv.Itoa(i), 4*1024)
	}
	list, err := client.ListObjectsV2(context.Background(), &tos.ListObjectsV2Input{Bucket: bucket})
	require.Nil(t, err)

	var toDelete []tos.ObjectTobeDeleted
	for _, object := range list.Contents {
		toDelete = append(toDelete, tos.ObjectTobeDeleted{
			Key: object.Key,
		})
	}
	mulDelete, err := client.DeleteMultiObjects(context.Background(), &tos.DeleteMultiObjectsInput{
		Bucket:  bucket,
		Objects: toDelete,
		Quiet:   false,
	})
	require.Nil(t, err)

	var deleted []string
	for _, object := range mulDelete.Deleted {
		deleted = append(deleted, object.Key)
	}
	sort.Strings(deleted)
	for i, object := range deleted {
		require.Equal(t, key+strconv.Itoa(i), object)
	}
}

func TestCAS(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("cas")
		key    = "key"
		value  = randomString(4 * 1024)
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	// wait for a while to make sure tos write data indeed
	time.Sleep(10 * time.Second)
	now := time.Now().UTC()
	eTag := put.ETag
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:  bucket,
		Key:     key,
		IfMatch: eTag,
	})
	require.Nil(t, err)
	require.Equal(t, get.StatusCode, 200)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:  bucket,
		Key:     key,
		IfMatch: "none-match" + eTag,
	})
	checkFail(t, get, err, 412)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:      bucket,
		Key:         key,
		IfNoneMatch: "none-match" + eTag,
	})
	checkSuccess(t, get, err, 200)
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:      bucket,
		Key:         key,
		IfNoneMatch: eTag,
	})
	checkFail(t, get, err, 304)
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:            bucket,
		Key:               key,
		IfUnmodifiedSince: now,
	})
	checkSuccess(t, get, err, 200)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:          bucket,
		Key:             key,
		IfModifiedSince: now,
	})
	checkFail(t, get, err, 304)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	put, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: strings.NewReader(value + "123"),
	})
	checkSuccess(t, put, err, 200)

	headObject, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:            bucket,
		Key:               key,
		IfUnmodifiedSince: headObject.LastModified.Add(-1 * time.Second),
	})
	checkFail(t, get, err, 412)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:          bucket,
		Key:             key,
		IfModifiedSince: headObject.LastModified.Add(-1 * time.Second),
	})
	checkSuccess(t, get, err, 200)
}

func TestEscapeCharacters(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("escape-characters")
		key    = "#?~` /\\中文测试"
		value  = randomString(4*1024) + "#?~` 中文测试"
		md5Sum = md5s(value)
		client = env.prepareClient(bucket, tos.WithSocketTimeout(60*time.Second, 60*time.Second))
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putInput := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	}
	putInput.ContentMD5 = md5Sum
	put, err := client.PutObjectV2(context.Background(), putInput)
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	buffer, err := ioutil.ReadAll(get.Content)
	require.Equal(t, len(value), len(buffer))
	require.Equal(t, md5Sum, md5s(string(buffer)))
	require.Nil(t, err)
}

func TestGetWithRange(t *testing.T) {
	var (
		env        = newTestEnv(t)
		bucket     = generateBucketName("get-with-range")
		key        = "key123"
		value      = randomString(2 * 4096)
		md5Sum     = md5s(value)
		partMD5Sum = md5s(value[:4096])
		client     = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putInput := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	}
	putInput.ContentMD5 = md5Sum
	put, err := client.PutObjectV2(context.Background(), putInput)
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{Bucket: bucket, Key: key, RangeStart: 0, RangeEnd: 4095})
	checkSuccess(t, get, err, 206)
	buffer, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, 4096, len(buffer))
	require.Equal(t, value[:4096], string(buffer))
	require.Equal(t, partMD5Sum, md5s(string(buffer)))
}

func TestAppend(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("append")
		key    = "key123"
		value1 = randomString(4 * 1024)
		value2 = randomString(4 * 1024)
		md5Sum = md5s(value1 + value2)
		client = env.prepareClient(bucket)
		meta   = map[string]string{"my-key": "长风破浪会有时"}
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	// none exist bucket
	appendOutput, err := client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:  "none-exsist-bucket",
		Key:     key,
		Offset:  0,
		Content: strings.NewReader(value1),
	})
	checkFail(t, appendOutput, err, 404)
	appendOutput, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:           bucket,
		Key:              key,
		Offset:           0,
		ContentMD5:       md5s(value1),
		GrantRead:        "id=0",
		GrantWriteAcp:    "id=1",
		GrantFullControl: "id=2",
		GrantReadAcp:     "id=3",
		Content:          strings.NewReader(value1),
		Meta:             meta,
	})
	checkSuccess(t, appendOutput, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	buffer, err := ioutil.ReadAll(get.Content)
	require.Equal(t, md5s(value1), md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}

	appendOutput, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:           bucket,
		Key:              key,
		Offset:           appendOutput.NextAppendOffset,
		ContentMD5:       md5s(value2),
		Content:          strings.NewReader(value2),
		PreHashCrc64ecma: appendOutput.HashCrc64ecma,
	})
	checkSuccess(t, appendOutput, err, 200)
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	buffer, err = ioutil.ReadAll(get.Content)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
}

func TestPutObjectFromFile(t *testing.T) {
	var (
		env          = newTestEnv(t)
		bucket       = generateBucketName("put-from-file")
		key          = "new123"
		value        = randomString(4 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		client       = env.prepareClient(bucket)
		ssecKey      = randomString(32)
		ssecMd5      = md5s(ssecKey)
		storageClass = enum.StorageClassIa
		fileName     = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	defer cleanTestFile(t, fileName)

	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value))
	require.Nil(t, err)
	require.Equal(t, len(value), n)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	input := &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			ContentMD5:    md5Sum,
			Expires:       expires,
			ACL:           acl,
			StorageClass:  storageClass,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
			Meta:          meta,
		},
		FilePath: fileName,
	}
	put, err := client.PutObjectFromFile(context.Background(), input)
	require.Nil(t, err)
	require.NotNil(t, put)
	require.Equal(t, 200, put.StatusCode)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:        bucket,
		Key:           key,
		SSECAlgorithm: "AES256",
		SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
		SSECKeyMD5:    ssecMd5,
	})
	buffer, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
}

func TestGetObjectToFile(t *testing.T) {
	var (
		env          = newTestEnv(t)
		bucket       = generateBucketName("get-to-file")
		key          = "new123"
		value        = randomString(4 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		client       = env.prepareClient(bucket)
		ssecKey      = randomString(32)
		ssecMd5      = md5s(ssecKey)
		storageClass = enum.StorageClassIa
		fileName     = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	defer cleanTestFile(t, fileName)
	input := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			ContentMD5:    md5Sum,
			Expires:       expires,
			ACL:           acl,
			StorageClass:  storageClass,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
			Meta:          meta,
		},
		Content: strings.NewReader(value),
	}
	put, err := client.PutObjectV2(context.Background(), input)
	require.Nil(t, err)
	require.NotNil(t, put)
	require.Equal(t, 200, put.StatusCode)

	get, err := client.GetObjectToFile(context.Background(), &tos.GetObjectToFileInput{
		GetObjectV2Input: tos.GetObjectV2Input{
			Bucket:        bucket,
			Key:           key,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
		},
		FilePath: fileName,
	})
	file, err := os.Open(fileName)
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, string(buffer), value)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
}

func TestPutAndGetFile(t *testing.T) {
	var (
		env          = newTestEnv(t)
		bucket       = generateBucketName("put-from-file")
		key          = "new123"
		value        = randomString(4 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		client       = env.prepareClient(bucket)
		ssecKey      = randomString(32)
		ssecMd5      = md5s(ssecKey)
		storageClass = enum.StorageClassIa
		fileName     = randomString(16) + ".file"
		downFileName = fileName + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, downFileName)
	}()

	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value))
	require.Nil(t, err)
	require.Equal(t, len(value), n)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	input := &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			ContentMD5:    md5Sum,
			Expires:       expires,
			ACL:           acl,
			StorageClass:  storageClass,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
			Meta:          meta,
		},
		FilePath: fileName,
	}
	put, err := client.PutObjectFromFile(context.Background(), input)
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectToFile(context.Background(), &tos.GetObjectToFileInput{
		GetObjectV2Input: tos.GetObjectV2Input{
			Bucket:        bucket,
			Key:           key,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
		},
		FilePath: downFileName,
	})
	checkSuccess(t, get, err, 200)
	downFile, err := os.Open(downFileName)
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(downFile)
	require.Equal(t, value, string(buffer))
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
}

func TestPutAndGetFileDir(t *testing.T) {
	var (
		env          = newTestEnv(t)
		bucket       = generateBucketName("put-from-file-dir")
		key          = "/new123"
		value        = randomString(4 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		client       = env.prepareClient(bucket)
		ssecKey      = randomString(32)
		ssecMd5      = md5s(ssecKey)
		storageClass = enum.StorageClassIa
		fileName     = randomString(16) + ".file"
		downFileName = "/tmp/gosdk/download"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, downFileName)
	}()

	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value))
	require.Nil(t, err)
	require.Equal(t, len(value), n)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	input := &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			ContentMD5:    md5Sum,
			Expires:       expires,
			ACL:           acl,
			StorageClass:  storageClass,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
			Meta:          meta,
		},
		FilePath: fileName,
	}
	put, err := client.PutObjectFromFile(context.Background(), input)
	checkSuccess(t, put, err, 200)
	get, err := client.GetObjectToFile(context.Background(), &tos.GetObjectToFileInput{
		GetObjectV2Input: tos.GetObjectV2Input{
			Bucket:        bucket,
			Key:           key,
			SSECAlgorithm: "AES256",
			SSECKey:       base64.StdEncoding.EncodeToString([]byte(ssecKey)),
			SSECKeyMD5:    ssecMd5,
		},
		FilePath: downFileName,
	})
	checkSuccess(t, get, err, 200)
	downFile, err := os.Open(downFileName)
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(downFile)
	require.Equal(t, value, string(buffer))
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		val, ok := get.Meta.Get(k)
		require.Equal(t, ok, true)
		require.Equal(t, v, val)
	}
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
}

func TestMultiVersion(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("multi-version")
		client  = env.prepareClient(bucket)
		key     = "new123"
		values  = []string{randomString(4 * 1024), randomString(4 * 1024), randomString(4 * 1024)}
		md5Sums = func(v []string) []string {
			r := make([]string, len(v))
			for i := 0; i < len(v); i++ {
				r[i] = md5s(v[i])
			}
			return r
		}(values)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	enableMultiVersion(t, client, bucket)
	var putted = make([]*tos.PutObjectV2Output, 0)
	map1 := make(map[string]string)
	map2 := make(map[string]string)
	time.Sleep(time.Minute)
	// put multi version objects
	for i := 0; i < 3; i++ {
		put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
			PutObjectBasicInput: tos.PutObjectBasicInput{
				Bucket: bucket,
				Key:    key,
			},
			Content: strings.NewReader(values[i]),
		})
		checkSuccess(t, put, err, 200)
		putted = append(putted, put)
		map1[put.VersionID] = put.ETag
		// time.Sleep(5 * time.Second)
	}
	require.Equal(t, 3, len(putted))
	// list multi version objects
	listed, err := client.ListObjectVersionsV2(context.Background(), &tos.ListObjectVersionsV2Input{
		Bucket: bucket,
		ListObjectVersionsInput: tos.ListObjectVersionsInput{
			Prefix: key,
		},
	})
	checkSuccess(t, listed, err, 200)
	require.Equal(t, 3, len(listed.Versions))
	for _, version := range listed.Versions {
		map2[version.VersionID] = version.ETag
	}
	require.Equal(t, map1, map2)
	// get specific version object
	for i := 0; i < 3; i++ {
		get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
			Bucket:    bucket,
			Key:       key,
			VersionID: putted[i].VersionID,
		})
		require.Nil(t, err)
		content, err := ioutil.ReadAll(get.Content)
		require.Nil(t, err)
		require.Equal(t, md5s(string(content)), md5Sums[i])
	}
}

func TestPutWithMeta(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-with-meta")
		key    = "key&123"
		client = env.prepareClient(bucket)
		value  = randomString(4 * 1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	metaValue := "abc=*%()%2f^!@#$%^&*_+"

	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, Meta: map[string]string{"key": metaValue}, ContentDisposition: "attachment; filename=\"中文.txt\""},
		Content:             bytes.NewBufferString(value),
	})

	require.Nil(t, err)
	output, err := client.HeadObjectV2(context.Background(), &tos.HeadObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	metaResValue, ok := output.Meta.Get("key")
	require.True(t, ok)
	require.Equal(t, metaResValue, metaValue)
}

func TestGetWithModify(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("get-with-modify")
		key    = "key123"
		client = env.prepareClient(bucket)
		value  = randomString(4 * 1024)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             bytes.NewBufferString(value),
	})
	require.Nil(t, err)
	_, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:          bucket,
		Key:             key,
		IfModifiedSince: time.Now(),
	})
	require.NotNil(t, err)
	tosErr := err.(*tos.TosServerError)
	require.Equal(t, tosErr.StatusCode, http.StatusNotModified)
}
func checkDataListener(t *testing.T, listener *dataTransferListenerTest) {
	require.Equal(t, listener.TotalBytes, listener.CurBytes)
	require.Equal(t, listener.TotalBytes, listener.AlreadyConsumer)
	require.Equal(t, listener.StartedTime, int64(1))
	require.Equal(t, listener.SuccessTime, int64(1))
	require.True(t, listener.TotalBytes != 0)
	require.True(t, listener.DataTransferType == enum.DataTransferSucceed)
}

func TestWithDataListener(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("data-with-listener")
		key    = "key123"
		client = env.prepareClient(bucket)
		value  = randomString(4 * 1024)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	listener := &dataTransferListenerTest{}
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, DataTransferListener: listener},
		Content:             bytes.NewBufferString(value),
	})
	require.Nil(t, err)
	checkDataListener(t, listener)

	listener = &dataTransferListenerTest{}
	getRes, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket:               bucket,
		Key:                  key,
		DataTransferListener: listener,
	})
	require.Nil(t, err)
	_, _ = ioutil.ReadAll(getRes.Content)
	checkDataListener(t, listener)

	uploadRes, err := client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: bucket,
		Key:    "upload" + key,
	})
	require.Nil(t, err)
	listener = &dataTransferListenerTest{}
	defer func() {
		client.AbortMultipartUpload(ctx, &tos.AbortMultipartUploadInput{
			Bucket:   bucket,
			Key:      uploadRes.Key,
			UploadID: uploadRes.UploadID,
		})
	}()
	_, err = client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{
			Bucket:               bucket,
			UploadID:             uploadRes.UploadID,
			PartNumber:           1,
			Key:                  key,
			DataTransferListener: listener,
		}, Content: bytes.NewBufferString(value),
	})
	checkDataListener(t, listener)
	listener = &dataTransferListenerTest{}

	_, _ = client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{
		Bucket:               bucket,
		Key:                  key,
		Content:              bytes.NewBufferString(value),
		DataTransferListener: listener,
	})
	checkDataListener(t, listener)
}

type CiIoReader struct {
	buff *bytes.Buffer
}

func (c CiIoReader) Read(p []byte) (n int, err error) {
	return c.buff.Read(p)
}

func TestObjectWithIoReader(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-basic-io-reader")
		client = env.prepareClient(bucket)
		key    = "key123"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	totalSize := 4 * 1024 * 1024
	ioReader := CiIoReader{buff: bytes.NewBufferString(randomString(totalSize))}
	d := &dataTransferListenerTest{}
	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:               bucket,
			Key:                  key,
			DataTransferListener: d,
		},
		Content: ioReader,
	})
	require.Nil(t, err)
	assert.Equal(t, d.TotalBytes, int64(totalSize))
	assert.Equal(t, d.TotalBytes, d.AlreadyConsumer)
	assert.Equal(t, d.DataTransferType, enum.DataTransferSucceed)
	assert.Equal(t, d.SuccessTime, int64(1))
}

func TestObjectWithRootPath(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-object-with-root")
		client = env.prepareClient(bucket)
		key    = "/key123"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	value := randomString(4 * 1024 * 1024)
	md5 := md5s(value)
	ioReader := CiIoReader{buff: bytes.NewBufferString(value)}
	_, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: ioReader,
	})
	require.Nil(t, err)
	out, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	data, err := ioutil.ReadAll(out.Content)
	require.Nil(t, err)
	require.Equal(t, md5s(string(data)), md5)
}

func TestObjectKey(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-object-test-key")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer cleanBucket(t, client, bucket)

	validKeys := []string{"a\\aa", "b/b\\a", "a?ac=asd", "$%&^^&$¥**^&(^", "ab·d/test.test", "a" + string(rune(32)), "a" + string(rune(127)), "a" + string(rune(183)), "a", "  a"}
	for _, key := range validKeys {
		value := randomString(4 * 1024)
		md5 := md5s(value)
		_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
			PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
			Content:             bytes.NewBufferString(value),
		})
		require.Nil(t, err)
		getRes, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
		require.Nil(t, err)
		data, err := ioutil.ReadAll(getRes.Content)
		require.Nil(t, err)
		require.Equal(t, md5s(string(data)), md5)
	}
	listRes, err := client.ListObjectsV2(ctx, &tos.ListObjectsV2Input{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, len(listRes.Contents), len(validKeys))

	invalidKey := []string{""}
	for _, key := range invalidKey {
		value := randomString(4 * 1024)
		_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
			PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
			Content:             bytes.NewBufferString(value),
		})
		require.NotNil(t, err)
		_, ok := err.(*tos.TosClientError)
		require.True(t, ok)
	}

}

func TestGetObjectWithResponseParams(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-object-test-key")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
		key    = "key123"
	)
	defer cleanBucket(t, client, bucket)
	putRandomObject(t, client, bucket, key, 1024)
	responseCacheControl := "10"
	responseContentType := "application/json"
	responseContentLanguage := "zh-cn"
	responseContentDisposition := "abc;def"
	responseContentEncoding := "deflate"
	res, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key,
		ResponseCacheControl:       responseCacheControl,
		ResponseExpires:            time.Now().Add(time.Hour),
		ResponseContentEncoding:    responseContentEncoding,
		ResponseContentLanguage:    responseContentLanguage,
		ResponseContentType:        responseContentType,
		ResponseContentDisposition: responseContentDisposition,
	})
	require.Nil(t, err)
	require.Equal(t, res.ContentLanguage, responseContentLanguage)
	require.Equal(t, res.ContentType, responseContentType)
	require.Equal(t, res.ContentDisposition, responseContentDisposition)
	require.Equal(t, res.CacheControl, responseCacheControl)
	require.Equal(t, res.ContentEncoding, responseContentEncoding)
	require.NotNil(t, res.Expires)
}

type CountReader struct {
	count int
}

func (c *CountReader) Read(p []byte) (n int, err error) {
	time.Sleep(time.Second)
	c.count++
	fmt.Println("Now Count:", c.count)
	if c.count == 35 {
		return 0, io.EOF
	}
	return rand.Read(p)

}

func TestUpload(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-with-reader-timeout")
		client = env.prepareClient(bucket)
		key    = randomString(6)
		ctx    = context.Background()
	)
	c := &CountReader{}
	defer cleanBucket(t, client, bucket)
	out, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             c,
	})
	require.Nil(t, err)
	t.Log("request id:", out.RequestID)
	require.Equal(t, c.count > 33, true)
	defer cleanBucket(t, client, bucket)
	putRandomObject(t, client, bucket, key, 4096)

	// When Range and RangeStart & RangeEnd appear together, range is preferred
	res, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: fmt.Sprintf("bytes=0-0"), RangeStart: 0, RangeEnd: 100})
	require.Nil(t, err)
	require.Equal(t, res.ContentLength, int64(1))
	body, err := ioutil.ReadAll(res.Content)
	require.Nil(t, err)
	require.Equal(t, len(body), 1)
	defer res.Content.Close()

	// only range
	res, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, Range: fmt.Sprintf("bytes=0-1")})
	require.Nil(t, err)
	require.Equal(t, res.ContentLength, int64(2))
	body, err = ioutil.ReadAll(res.Content)
	require.Nil(t, err)
	require.Equal(t, len(body), 2)
	defer res.Content.Close()

	// only RangeStart & RangeEnd
	res, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, RangeStart: 0, RangeEnd: 1})
	require.Nil(t, err)
	require.Equal(t, res.ContentLength, int64(2))
	body, err = ioutil.ReadAll(res.Content)
	require.Nil(t, err)
	require.Equal(t, len(body), 2)
	defer res.Content.Close()
}

func TestObjectWithRateLimit(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-with-rate-limit")
		client = env.prepareClient(bucket)
		key    = randomString(6)
		ctx    = context.Background()
	)
	defer cleanBucket(t, client, bucket)
	rowData := randomString(1024 * 1024 * 7)
	data := strings.NewReader(rowData)
	now := time.Now()
	limiter := tos.NewDefaultRateLimit(1024*1024, 1024*1024)
	putoutput, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      bucket,
			Key:         key,
			RateLimiter: limiter,
		},
		Content: data,
	})
	require.Nil(t, err)
	t.Log(putoutput.RequestID)
	t.Logf("putobject cost: %v", time.Now().Sub(now).Seconds())
	require.True(t, time.Now().Sub(now) >= time.Second*5)
	require.True(t, time.Now().Sub(now) <= time.Second*10)

	// 限流耗时应大于不限流
	start := time.Now()
	getoutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key, RateLimiter: limiter})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getoutput.Content)
	require.Nil(t, err)
	require.Equal(t, string(getData), rowData)
	t.Logf("getobject cost: %v", time.Now().Sub(start).Seconds())
	limiterCost := time.Now().Sub(start)

	noLimiterStart := time.Now()
	getoutput, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	getData, err = ioutil.ReadAll(getoutput.Content)
	require.Nil(t, err)
	require.Equal(t, string(getData), rowData)
	t.Logf("getobject cost: %v", time.Now().Sub(noLimiterStart).Seconds())
	require.True(t, time.Now().Sub(noLimiterStart) < limiterCost)

}

type retryReader struct {
	count  int
	reader io.Reader
	n      int
	t      *testing.T
}

func (r *retryReader) Read(p []byte) (n int, err error) {
	if r.count == 5 {

		r.t.Log("Already Read:", r.n)
		require.True(r.t, r.n > 0)
		return r.reader.Read(p)
	}

	time.Sleep(time.Second * 3)
	n, err = r.reader.Read(p)
	r.n += n
	return n, err
}

func (r *retryReader) Seek(offset int64, whence int) (int64, error) {
	r.count++
	seek := r.reader.(io.Seeker)
	return seek.Seek(offset, whence)
}

type noRetryReader struct {
	count  int
	reader io.Reader
}

func (r *noRetryReader) Read(p []byte) (n int, err error) {
	time.Sleep(time.Second * 3)
	return r.reader.Read(p)
}

func TestObjectWithRetry(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-with-retry")
		key    = randomString(6)
		ctx    = context.Background()
	)
	tsConfig := tos.DefaultTransportConfig()
	tsConfig.ReadTimeout = time.Millisecond * 500
	tsConfig.WriteTimeout = time.Millisecond * 500
	client := env.prepareClient(bucket, tos.WithTransportConfig(&tsConfig), tos.WithMaxRetryCount(5))
	defer cleanBucket(t, client, bucket)
	// Case 1: 测试内存流
	rawData := "hello world"
	data := &retryReader{reader: strings.NewReader(rawData), t: t}
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: data,
	})
	require.Nil(t, err)
	require.Equal(t, data.count, 5)

	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	getData, err := ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	require.Equal(t, string(getData), rawData)

	// case 2: 测试文件流重试
	fileName := randomString(5)
	file, err := os.Create(fileName)
	require.Nil(t, err)
	defer os.Remove(fileName)
	fileSize := 1024
	readOffset := 1000
	value := randomString(fileSize)
	_, err = file.Write([]byte(value))
	require.Nil(t, err)
	err = file.Sync()
	require.Nil(t, err)
	// Seek 到 1000
	_, err = file.Seek(int64(readOffset), io.SeekStart)
	require.Nil(t, err)
	data = &retryReader{reader: file, t: t}
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: data,
	})
	require.Nil(t, err)
	require.Equal(t, data.count, 5)
	getOutput, err = client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: key})
	require.Nil(t, err)
	getData, err = ioutil.ReadAll(getOutput.Content)
	require.Nil(t, err)
	t.Log("getData:", string(getData), "RealValue:", value[1000:])
	require.Equal(t, string(getData), string(value[1000:]))

	// Case3: 网络流不可重试
	res, err := http.Get("https://www.volcengine.com/")
	require.Nil(t, err)
	nrReader := &noRetryReader{reader: res.Body}
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: nrReader,
	})
	require.NotNil(t, err)

}

type contentType struct {
	bucket            string
	client            *tos.ClientV2
	keySuffix         string
	expectContentType string
	inputContentType  string
}

func (c *contentType) testContentType(t *testing.T, ctx context.Context) {
	key := randomString(6) + "." + c.keySuffix
	_, err := c.client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: c.bucket, Key: key, ContentType: c.inputContentType},
		Content:             strings.NewReader(randomString(1)),
	})
	require.Nil(t, err)
	houtput, err := c.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: c.bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, houtput.ContentType, c.expectContentType)

	// 1.2: AppendObject
	key = randomString(6) + "." + c.keySuffix
	_, err = c.client.AppendObjectV2(ctx, &tos.AppendObjectV2Input{Bucket: c.bucket,
		Key:         key,
		ContentType: c.inputContentType,
		Content:     strings.NewReader(randomString(128 * 1024)),
	})
	require.Nil(t, err)
	houtput, err = c.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: c.bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, houtput.ContentType, c.expectContentType)

	// 1.3: MultiPart
	key = randomString(6) + "." + c.keySuffix
	createMultiOutput, err := c.client.CreateMultipartUploadV2(ctx, &tos.CreateMultipartUploadV2Input{
		Bucket: c.bucket, Key: key,
		ContentType: c.inputContentType,
	})
	require.Nil(t, err)
	part, err := c.client.UploadPartV2(ctx, &tos.UploadPartV2Input{
		UploadPartBasicInput: tos.UploadPartBasicInput{Bucket: c.bucket, Key: key, UploadID: createMultiOutput.UploadID, PartNumber: 1},
		Content:              strings.NewReader(randomString(1024)),
	})
	require.Nil(t, err)
	_, err = c.client.CompleteMultipartUploadV2(ctx, &tos.CompleteMultipartUploadV2Input{Bucket: c.bucket, Key: key, UploadID: createMultiOutput.UploadID, Parts: []tos.UploadedPartV2{{
		PartNumber: 1,
		ETag:       part.ETag,
	}}})
	require.Nil(t, err)
	houtput, err = c.client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: c.bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, houtput.ContentType, c.expectContentType)

}

func TestContentType(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("object-with-content-type")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer cleanBucket(t, client, bucket)
	ct := contentType{
		bucket:            bucket,
		client:            client,
		keySuffix:         "",
		expectContentType: "binary/octet-stream",
		inputContentType:  "",
	}
	// Case 1: 测试 default content type
	ct.testContentType(t, ctx)

	// Case 2: 测试 SDK 推断 content type
	ct.keySuffix = ".jpg"
	ct.expectContentType = "image/jpeg"
	ct.testContentType(t, ctx)

	// Case 3: 测试指定 ContentType
	ct.inputContentType = "text/html"
	ct.expectContentType = "text/html"
	ct.testContentType(t, ctx)

}
