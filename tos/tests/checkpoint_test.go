package tests

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type dataTransferListenerTest struct {
	TotalBytes      int64
	CurBytes        int64 // bytes read/written
	StartedTime     int64
	SuccessTime     int64
	AlreadyConsumer int64
	RWTime          int64
}

func (d *dataTransferListenerTest) DataTransferStatusChange(status *tos.DataTransferStatus) {
	switch status.Type {
	case enum.DataTransferStarted:
		d.StartedTime += 1
	case enum.DataTransferRW:
		d.TotalBytes = status.TotalBytes
		atomic.AddInt64(&d.CurBytes, status.RWOnceBytes)
		d.AlreadyConsumer = status.ConsumedBytes
		d.RWTime += 1
	case enum.DataTransferSucceed:
		d.SuccessTime += 1
	case enum.DataTransferFailed:
		fmt.Println("Failed")
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

	get, err = client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket:     bucket,
		Key:        key,
		PartNumber: 1,
	})
	checkSuccess(t, get, err, 206)
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
	count  int
	cancel tos.CancelHook
}

func (l *uploadFileListenerTest) EventChange(event *tos.UploadEvent) {
	if event.Type == enum.UploadEventUploadPartSucceed {
		l.count++
	}
	if l.count == 2 {
		l.cancel.Cancel(false)
	}
}

func TestUploadFileCancelHook(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("upload-file-cancel-hook")
		key      = "key123"
		value1   = randomString(22 * 1024 * 1024)
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
		count:  0,
		cancel: hook,
	}
	input.UploadEventListener = listener
	upload, err := client.UploadFile(context.Background(), input)
	require.Nil(t, upload)
	require.NotNil(t, err)
	// checkpoint file still exist
	stat, err := os.Stat(strings.Join([]string{fileName, bucket, key, "upload"}, "."))
	require.Nil(t, err)
	require.Equal(t, 2, listener.count)

	input.CancelHook = tos.NewCancelHook()
	upload, err = client.UploadFile(context.Background(), input)
	require.Nil(t, err)

	get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{
		Bucket: bucket,
		Key:    key,
	})
	checkSuccess(t, get, err, 200)
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

type DownloadListenerTest struct {
	count int
	input *tos.DownloadFileInput
}

func (l *DownloadListenerTest) EventChange(event *tos.DownloadEvent) {
	if event.Type == enum.DownloadEventDownloadPartSucceed {
		l.count++
	}
	if l.count == 2 {
		l.input.CancelHook.Cancel(false)
	}
}

//
func TestDownloadCancelHook(t *testing.T) {
	var (
		env      = newTestEnv(t)
		bucket   = generateBucketName("download-file-cancel-hook")
		key      = "key123"
		value1   = randomString(22 * 1024 * 1024)
		client   = env.prepareClient(bucket)
		fileName = randomString(16) + ".file"
		md5Sum   = md5s(value1)
	)
	defer func() {
		cleanBucket(t, client, bucket)
		cleanTestFile(t, fileName)
		cleanTestFile(t, fileName+".file")
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
		EnableCheckpoint: true,
		CancelHook:       hook,
	}
	listener := DownloadListenerTest{
		count: 0,
		input: input,
	}
	input.DownloadEventListener = &listener
	_, err = client.DownloadFile(context.Background(), input)
	require.Equal(t, 2, listener.count)

	stat, err := os.Stat(fileName + ".file" + ".temp")
	require.Nil(t, err)
	input.CancelHook = tos.NewCancelHook()
	_, err = client.DownloadFile(context.Background(), input)
	require.Nil(t, err)
	file, err = os.Open(fileName + ".file")
	buffer, err := ioutil.ReadAll(file)
	require.Nil(t, err)
	require.Equal(t, md5Sum, md5s(string(buffer)))
	_ = os.Remove(stat.Name())
	stat, err = os.Stat(strings.Join([]string{fileName + ".file", bucket, key, "download"}, "."))
	require.Nil(t, err)
	_ = os.Remove(stat.Name())
	require.Equal(t, 5, listener.count)
}

func TestLargeFile(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("upload-file")
		key    = "key123"
		value1 = randomString(1000 * 1024 * 1024)
		md5Sum = md5s(value1)

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
		TaskNum:              100,
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
		TaskNum:              100,
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
