package tests

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type dataTransferListenerTest struct {
	TotalBytes       int64
	CurBytes         int64 // bytes read/written
	StartedTime      int64
	SuccessTime      int64
	AlreadyConsumer  int64
	RWTime           int64
	DataTransferType enum.DataTransferType
	RetryCount       int
}

func (d *dataTransferListenerTest) DataTransferStatusChange(status *tos.DataTransferStatus) {
	switch status.Type {
	case enum.DataTransferStarted:
		d.StartedTime += 1
		d.DataTransferType = enum.DataTransferStarted
	case enum.DataTransferRW:
		d.TotalBytes = status.TotalBytes
		atomic.AddInt64(&d.CurBytes, status.RWOnceBytes)
		d.AlreadyConsumer = status.ConsumedBytes
		d.RWTime += 1
		d.DataTransferType = enum.DataTransferRW
		d.RetryCount = status.RetryCount
	case enum.DataTransferSucceed:
		d.SuccessTime += 1
		d.TotalBytes = status.TotalBytes
		d.DataTransferType = enum.DataTransferSucceed
		d.RetryCount = status.RetryCount
	case enum.DataTransferFailed:
		fmt.Println("Failed")
		d.DataTransferType = enum.DataTransferFailed
	}
}

func TestUploadFile(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file")
		key      = "key123"
		value1   = randomString(20 * 1024 * 1024)
		client   = env.prepareClient(bucket, LongTimeOutClientOption...)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	transferListener := &dataTransferListenerTest{}
	upload, err := client.UploadFile(context.Background(), &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:             fileName,
		PartSize:             5 * 1024 * 1024,
		TaskNum:              4,
		EnableCheckpoint:     false,
		DataTransferListener: transferListener,
	})
	checkDataListener(t, transferListener)
	checkSuccess(t, upload, err, 200)
	require.Equal(t, transferListener.StartedTime, int64(1))
	require.Equal(t, transferListener.TotalBytes, transferListener.CurBytes)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, md5s(string(data)), md5s(value1))
}

func TestUploadEmptyFile(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file")
		key      = "key123"
		value1   = ""
		client   = env.prepareClient(bucket, LongTimeOutClientOption...)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
	}()
	value1 = ""
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	upload, err := client.UploadFile(context.Background(), &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:         fileName,
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		EnableCheckpoint: false,
	})
	checkSuccess(t, upload, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, md5s(string(data)), md5s(value1))
	_, err = client.DownloadFile(context.Background(), &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{Bucket: bucket, Key: key},
		FilePath:          fileName,
	})
	require.Nil(t, err)
	data, err = ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, md5s(string(data)), md5s(value1))
}

func TestUploadFileWithCheckpoint(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file-with-checkpoint")
		key      = "key123"
		value1   = randomString(20 * 1024 * 1024)
		client   = env.prepareClient(bucket, LongTimeOutClientOption...)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()

	upload, err := client.UploadFile(context.Background(), &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:         fileName,
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		EnableCheckpoint: true,
		// DataTransferListener: &dataTransferListenerTest{},
	})
	checkSuccess(t, upload, err, 200)
	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
}

type uploadFileListenerTest struct {
	count   int
	cancel  tos.CancelHook
	maxTime int
}

func (l *uploadFileListenerTest) EventChange(event *tos.UploadEvent) {
	if event.Type == enum.UploadEventUploadPartSucceed {
		l.count++
	}
	if l.count == l.maxTime {
		l.cancel.Cancel(false)
	}
}

func getCrc(value []byte) uint64 {
	var checker hash.Hash64
	checker = crc64.New(crc64.MakeTable(crc64.ECMA))
	checker.Write(value)
	return checker.Sum64()
}

func TestUploadFileCancelHook(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file-cancel-hook")
		key      = "key123"
		value1   = randomString(22 * 1024 * 1024)
		md5sum   = md5s(value1)
		client   = env.prepareClient(bucket, LongTimeOutClientOption...)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".checkpoint")
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	hook := tos.NewCancelHook()
	input := &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:         fileName,
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		EnableCheckpoint: true,
		CancelHook:       hook,
	}
	listener := &uploadFileListenerTest{
		count:   0,
		cancel:  hook,
		maxTime: 2,
	}
	input.UploadEventListener = listener
	upload, err := client.UploadFile(context.Background(), input)
	require.Nil(t, upload)
	require.NotNil(t, err)
	// checkpoint file still exist
	stat, err := os.Stat(input.CheckpointFile)
	require.Nil(t, err)
	require.True(t, listener.count >= 2)

	input.CancelHook = tos.NewCancelHook()
	listener.maxTime = 3
	d := &dataTransferListenerTest{}
	input.DataTransferListener = d
	upload, err = client.UploadFile(context.Background(), input)
	require.Nil(t, err)
	file, err = os.Open(fileName)
	require.Nil(t, err)
	fileData, err := ioutil.ReadAll(file)
	require.Equal(t, upload.HashCrc64ecma, getCrc(fileData))
	require.Equal(t, d.AlreadyConsumer, d.TotalBytes)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)

	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, md5sum, md5s(string(data)))
	_ = os.Remove(stat.Name())
	require.Equal(t, 5, listener.count)

}

func TestUploadFileUpdate(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file-checkpoint-update")
		key      = "key123"
		value1   = randomString(22 * 1024 * 1024)
		client   = env.prepareClient(bucket, LongTimeOutClientOption...)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".checkpoint")
	}()

	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	file.Sync()
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	hook := tos.NewCancelHook()
	input := &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:         fileName,
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		EnableCheckpoint: true,
		CheckpointFile:   fileName + ".checkpoint",
		CancelHook:       hook,
	}
	listener := &uploadFileListenerTest{
		count:   0,
		cancel:  hook,
		maxTime: 2,
	}
	input.UploadEventListener = listener
	upload, err := client.UploadFile(context.Background(), input)
	require.Nil(t, upload)
	require.NotNil(t, err)
	// checkpoint file still exist
	stat, err := os.Stat(fileName + ".checkpoint")
	require.Nil(t, err)
	require.True(t, listener.count >= 2)

	value1 = randomString(23 * 1024 * 1024)
	md5sum := md5s(value1)

	os.Remove(fileName)
	file, err = os.Create(fileName)
	require.Nil(t, err)
	n, err = file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	file.Sync()
	defer file.Close()

	time.Sleep(time.Second)
	input.CancelHook = tos.NewCancelHook()
	listener.count = 0
	listener.maxTime = 5
	listener.cancel = input.CancelHook
	upload, err = client.UploadFile(context.Background(), input)
	require.Nil(t, err)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)

	data, err := ioutil.ReadAll(get.Content)
	require.Nil(t, err)
	require.Equal(t, md5sum, md5s(string(data)))
	_ = os.Remove(stat.Name())
	require.Equal(t, 5, listener.count)

}

func TestDownloadFile(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("download-file")
		key      = "key123"
		value1   = randomString(20 * 1024 * 1024)
		md5Sum   = md5s(value1)
		client   = env.prepareClient(bucket)
		fileName = randomString(16) + ".file"
	)

	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".file")
		cleanTestFile(t, fileName+".file"+".temp")
		cleanTestFile(t, strings.Join([]string{fileName + ".file", bucket, key, "download"}, "."))
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	file.Sync()

	upload, err := client.PutObjectFromFile(context.Background(), &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		FilePath: fileName,
	})

	checkSuccess(t, upload, err, 200)
	listener := &dataTransferListenerTest{}
	download, err := client.DownloadFile(context.Background(), &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    key,
		},
		PartSize:             5 * 1024 * 1024,
		TaskNum:              4,
		FilePath:             fileName + ".file", // xxx.file.file
		EnableCheckpoint:     false,
		DataTransferListener: listener,
	})
	checkDataListener(t, listener)
	checkSuccess(t, download, err, 200)
	_, err = file.Seek(0, io.SeekStart)
	require.Nil(t, err)
	downFile, err := os.Open(fileName + ".file")
	require.Nil(t, err)
	allContent, err := ioutil.ReadAll(downFile)
	require.Nil(t, nil)
	require.Equal(t, md5Sum, md5s(string(allContent)))

}

func TestDownloadFileWithCheckpoint(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("download-file-with-checkpoint")
		key      = "key123"
		value1   = randomString(20 * 1024 * 1024)
		md5Sum   = md5s(value1)
		client   = env.prepareClient(bucket)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".file")
		cleanTestFile(t, fileName+".file"+".temp")
		cleanTestFile(t, strings.Join([]string{fileName + ".file", bucket, key, "download"}, "."))
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()

	upload, err := client.PutObjectFromFile(context.Background(), &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		FilePath: fileName,
	})
	checkSuccess(t, upload, err, 200)
	download, err := client.DownloadFile(context.Background(), &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    key,
		},
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		FilePath:         fileName + ".file", // xxx.file.file
		EnableCheckpoint: true,
	})
	checkSuccess(t, download, err, 200)
	file, err = os.Open(fileName + ".file")
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, md5Sum, md5s(string(buffer)))
}

type DownloadCancelListenerTest struct {
	count   int
	input   *tos.DownloadFileInput
	maxTime int
}

func (l *DownloadCancelListenerTest) EventChange(event *tos.DownloadEvent) {
	if event.Type == enum.DownloadEventDownloadPartSucceed {
		l.count++
	}
	if l.count == l.maxTime {
		l.input.CancelHook.Cancel(false)
	}
}

func TestDownloadCancelHook(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("download-file-cancel-hook")
		key      = "key123"
		value1   = randomString(82 * 1024 * 1024)
		client   = env.prepareClient(bucket)
		fileName = randomString(16) + ".file"
		md5Sum   = md5s(value1)
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".file")
		cleanTestFile(t, fileName+".checkpoint")
		cleanTestFile(t, fileName+".file.temp")
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	upload, err := client.PutObjectFromFile(context.Background(), &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		FilePath: fileName,
	})
	checkSuccess(t, upload, err, 200)
	hook := tos.NewCancelHook()

	input := &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    key,
		},
		PartSize:         20 * 1024 * 1024,
		TaskNum:          4,
		FilePath:         fileName + ".file", // xxx.file.file
		CheckpointFile:   fileName + ".checkpoint",
		EnableCheckpoint: true,
		CancelHook:       hook,
	}
	listener := &DownloadCancelListenerTest{
		count:   0,
		input:   input,
		maxTime: 2,
	}
	input.DownloadEventListener = listener
	_, err = client.DownloadFile(context.Background(), input)
	require.True(t, listener.count >= 2)
	t.Logf("listener.count:%d", listener.count)
	_, err = os.Stat(fileName + ".file" + ".temp")
	require.Nil(t, err)

	input.CancelHook = tos.NewCancelHook()
	listener.input = input
	listener.maxTime = 5
	d := &dataTransferListenerTest{}
	input.DataTransferListener = d
	_, err = client.DownloadFile(context.Background(), input)
	require.Nil(t, err)
	file, err = os.Open(fileName + ".file")
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	require.Equal(t, 5, listener.count)
	require.Equal(t, d.AlreadyConsumer, d.TotalBytes)
	require.Equal(t, d.DataTransferType, enum.DataTransferSucceed)
	require.Equal(t, d.SuccessTime, int64(1))
}

func TestLargeFile(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("upload-file")
		key    = "key123"
		value1 = randomString(1000 * 1024 * 1024)
		md5Sum = md5s(value1)
	)
	config := tos.DefaultTransportConfig()
	highLatencyLogThreshold := 100 * 1024 * 1024
	config.HighLatencyLogThreshold = &highLatencyLogThreshold
	ops := make([]tos.ClientOption, 0)
	ops = append(ops, tos.WithTransportConfig(&config))
	ops = append(ops, LongTimeOutClientOption...)
	client := env.prepareClient(bucket, ops...)
	fileName := randomString(16) + ".file"
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()
	transferListener := &dataTransferListenerTest{}
	upload, err := client.UploadFile(context.Background(), &tos.UploadFileInput{
		CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
			Bucket: bucket,
			Key:    key,
		},
		FilePath:             fileName,
		PartSize:             5 * 1024 * 1024,
		TaskNum:              5,
		EnableCheckpoint:     false,
		DataTransferListener: transferListener,
	})
	checkSuccess(t, upload, err, 200)
	checkDataListener(t, transferListener)
	require.Equal(t, transferListener.StartedTime, int64(1))
	require.Equal(t, transferListener.TotalBytes, transferListener.CurBytes)
	transferListener = &dataTransferListenerTest{}
	download, err := client.DownloadFile(context.Background(), &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    key,
		},
		PartSize:             5 * 1024 * 1024,
		TaskNum:              5,
		FilePath:             fileName + ".file", // xxx.file.file
		EnableCheckpoint:     true,
		DataTransferListener: transferListener,
	})
	checkSuccess(t, download, err, 200)
	checkDataListener(t, transferListener)

	file, err = os.Open(fileName + ".file")
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, md5Sum, md5s(string(buffer)))
}

func TestDownloadFileWithUpdate(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("download-file-with-checkpoint")
		key      = "key123"
		value1   = randomString(20 * 1024 * 1024)
		client   = env.prepareClient(bucket)
		fileName = randomString(16) + ".file"
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".file")
		cleanTestFile(t, fileName+".file"+".temp")
		cleanTestFile(t, fileName+".checkpoint")
		cleanTestFile(t, strings.Join([]string{fileName + ".file", bucket, key, "download"}, "."))
	}()
	file, err := os.Create(fileName)
	require.Nil(t, err)
	n, err := file.Write([]byte(value1))
	require.Nil(t, err)
	require.Equal(t, len(value1), n)
	defer file.Close()

	upload, err := client.PutObjectFromFile(context.Background(), &tos.PutObjectFromFileInput{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		FilePath: fileName,
	})
	checkSuccess(t, upload, err, 200)
	hook := tos.NewCancelHook()
	input := &tos.DownloadFileInput{
		HeadObjectV2Input: tos.HeadObjectV2Input{
			Bucket: bucket,
			Key:    key,
		},
		PartSize:         5 * 1024 * 1024,
		TaskNum:          4,
		FilePath:         fileName + ".file", // xxx.file.file
		CheckpointFile:   fileName + ".checkpoint",
		EnableCheckpoint: true,
		CancelHook:       hook,
	}
	listener := DownloadCancelListenerTest{
		count:   0,
		input:   input,
		maxTime: 2,
	}
	input.DownloadEventListener = &listener
	_, err = client.DownloadFile(context.Background(), input)
	require.True(t, listener.count >= 2)
	// Checkpoint 文件还存在
	_, err = os.Stat(fileName + ".checkpoint")
	require.Nil(t, err)

	// 重现上传新文件
	value1 = randomString(20 * 1024 * 1024)
	_, err = client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket: bucket,
			Key:    key,
		},
		Content: bytes.NewBufferString(value1),
	})
	require.Nil(t, err)

	hook = tos.NewCancelHook()
	input.CancelHook = hook
	input.DownloadEventListener = nil
	download, err := client.DownloadFile(context.Background(), input)
	checkSuccess(t, download, err, 200)
	file, err = os.Open(fileName + ".file")
	require.Nil(t, err)
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, md5s(value1), md5s(string(buffer)))
}

func TestDownloadFileWithDirPath(t *testing.T) {
	var (
		env            = newTestEnv(t)
		bucket         = generateBucketName("download-file-with-checkpoint")
		value1         = randomString(5 * 1024)
		client         = env.prepareClient(bucket)
		filePath       = "/tmp/gosdk/"
		checkpointPath = "/tmp/gosdk/checkpoint"
		ctx            = context.Background()
	)
	defer cleanBucket(t, client, bucket)
	keyList := []string{"/a/b.file", "/a/b/c.file", "/a/d/e/f/g.file", "a/b/d.file", "/a/g/", "/a/f"}
	for _, key := range keyList {
		_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
			PutObjectBasicInput: tos.PutObjectBasicInput{
				Bucket: bucket,
				Key:    key,
			},
			Content: bytes.NewReader([]byte(value1)),
		})
		require.Nil(t, err)
	}

	// filePath dir not exist
	path := filePath + randomString(4) + "/"
	for _, key := range keyList {

		_, err := client.DownloadFile(ctx, &tos.DownloadFileInput{
			HeadObjectV2Input: tos.HeadObjectV2Input{Bucket: bucket, Key: key},
			FilePath:          path,
			CheckpointFile:    checkpointPath,
			EnableCheckpoint:  true,
		})
		require.Nil(t, err)
	}

	// filePath dir exist
	path = filePath + randomString(4)
	os.MkdirAll(path, os.ModePerm)
	for _, key := range keyList {

		_, err := client.DownloadFile(ctx, &tos.DownloadFileInput{
			HeadObjectV2Input: tos.HeadObjectV2Input{Bucket: bucket, Key: key},
			FilePath:          path,
			EnableCheckpoint:  true,
			CheckpointFile:    "/tmp/gosdk/checkpoint/",
		})
		require.Nil(t, err)
	}
}
