package tests

import (
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/codes"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestGetNotExistObject(t *testing.T) {
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
		value        = randomString(4 * 1024)
		md5Sum       = md5s(value)
		expires      = time.Now().UTC().Add(time.Hour)
		acl          = enum.ACLAuthRead
		meta         = map[string]string{"Hello": "world"}
		ssecKey      = randomString(32)
		ssecMd5      = md5s(ssecKey)
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
		require.Equal(t, v, head.Meta.Get(k))
	}
	require.Equal(t, "中文测试", get.ContentDisposition)
	require.Equal(t, expires.Format(time.UnixDate), get.Expires.Format(time.UnixDate))
	require.Equal(t, storageClass, get.StorageClass)
}

func TestPutEmptyObject(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-empty-object")
		key    = "key123"
		value  = ""
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
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

func TestCopyObject(t *testing.T) {
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
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
		Content:             strings.NewReader(value),
	})
	checkSuccess(t, put, err, 200)
	_, err = client.CopyObject(context.Background(), &tos.CopyObjectInput{
		Bucket:    bucket,
		Key:       copyedKey,
		SrcBucket: bucket,
		SrcKey:    key,
	})
	require.Nil(t, err)
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

func TestInvalidObjectKey(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("test-invalid-object-key")
		client = env.prepareClient(bucket)
	)
	testInvalidObjectKey(t, client, "	key")
	testInvalidObjectKey(t, client, randomString(1001))
	testInvalidObjectKey(t, client, "/key")
	testInvalidObjectKey(t, client, "\\key")
	key1 := make([]byte, 5)
	for i := 0; i < len(key1); i++ {
		key1[i] = byte(i)
	}
	testInvalidObjectKey(t, client, string(key1))
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
	for k := range meta {
		require.Equal(t, meta[k], get.Meta.Get(k))
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
	require.Equal(t, 400, tos.StatusCode(err))
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
	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:            bucket,
		Key:               key,
		IfUnmodifiedSince: now,
	})
	checkFail(t, get, err, 412)

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:          bucket,
		Key:             key,
		IfModifiedSince: now,
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
	appendOutput, err = client.AppendObjectV2(context.Background(), &tos.AppendObjectV2Input{
		Bucket:           bucket,
		Key:              key,
		Offset:           appendOutput.NextAppendOffset,
		ContentMD5:       md5s(value2),
		Content:          strings.NewReader(value2),
		PreHashCrc64ecma: appendOutput.HashCrc64ecma,
	})
	checkSuccess(t, appendOutput, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	buffer, err := ioutil.ReadAll(get.Content)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	for k, v := range meta {
		require.Equal(t, v, get.Meta.Get(k))
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
		require.Equal(t, v, get.Meta.Get(k))
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
		require.Equal(t, v, get.Meta.Get(k))
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
		require.Equal(t, v, get.Meta.Get(k))
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
