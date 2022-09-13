package tos

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func getDownloadCheckpoint(enabled bool, checkpointPath string, init func(input *HeadObjectV2Output) (*downloadCheckpoint, error), output *HeadObjectV2Output) (checkpoint *downloadCheckpoint, err error) {
	if enabled {
		_, err = os.Stat(checkpointPath)
		// if err is not empty, assume checkpoint not exists
		if err == nil {
			checkpoint = &downloadCheckpoint{}
			loadCheckPoint(checkpointPath, checkpoint)
			if checkpoint != nil && checkpoint.ObjectInfo.Etag == output.ETag {
				return
			}
		}
		_, err = os.Create(checkpointPath)
		if err != nil {
			return nil, newTosClientError(err.Error(), err)
		}
	}
	checkpoint, err = init(output)
	if err != nil {
		return nil, err
	}
	if enabled {
		err = checkpoint.WriteToFile()
		if err != nil {
			return nil, err
		}
	}
	return
}

func (cli *ClientV2) DownloadFile(ctx context.Context, input *DownloadFileInput) (*DownloadFileOutput, error) {
	err := validateDownloadInput(input)
	if err != nil {
		return nil, err
	}
	headOutput, err := cli.HeadObjectV2(ctx, &input.HeadObjectV2Input)
	if err != nil {
		return nil, err
	}
	event := downloadEvent{input: input}
	init := func(output *HeadObjectV2Output) (*downloadCheckpoint, error) {
		err := createTempFile(input, event)
		if err != nil {
			return nil, err
		}
		return initDownloadCheckpoint(input, headOutput)
	}
	checkpoint, err := getDownloadCheckpoint(input.EnableCheckpoint, input.CheckpointFile, init, headOutput)
	if err != nil {
		return nil, err
	}
	cleaner := func() {
		_ = os.Remove(input.CheckpointFile)
	}
	bindCancelHookWithCleaner(input.CancelHook, cleaner)
	return cli.downloadFile(ctx, headOutput, checkpoint, input, event)
}

// loadCheckPoint load UploadFile checkpoint or DownloadFile checkpoint.
// checkpoint must be a pointer
func loadCheckPoint(path string, checkpoint interface{}) {
	contents, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if len(contents) == 0 {
		return
	}
	json.Unmarshal(contents, &checkpoint)
}

// if file is a directory, append suffix to it to make a file name
func mustFile(file *string, suffix string) {
	stat, _ := os.Stat(*file)
	if stat != nil && stat.IsDir() {
		*file = filepath.Join(*file, suffix)
	}
}

func validateDownloadInput(input *DownloadFileInput) error {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return err
	}
	if input.PartSize == 0 {
		input.PartSize = MinPartSize
	}
	if input.PartSize < MinPartSize || input.PartSize > MaxPartSize {
		return newTosClientError("The input part size is invalid, please set it range from 5MB to 5GB", nil)
	}
	// if directory, append object key at end
	mustFile(&input.FilePath, input.Key)
	input.tempFile = input.FilePath + TempFileSuffix
	if input.EnableCheckpoint {
		// get correct checkpoint path
		if len(input.CheckpointFile) == 0 {
			dirName, _ := filepath.Split(input.FilePath)
			fileName := strings.Join([]string{input.FilePath, input.Bucket, input.Key, "download"}, ".")
			input.CheckpointFile = filepath.Join(dirName, fileName)
		} else {
			mustFile(&input.CheckpointFile, strings.Join([]string{input.FilePath, input.Bucket, input.Key, "download"}, "."))
		}
	}
	if input.TaskNum < 1 {
		input.TaskNum = 1
	}
	if input.TaskNum > 1000 {
		input.TaskNum = 1000
	}
	return nil
}

func initDownloadCheckpoint(input *DownloadFileInput, headOutput *HeadObjectV2Output) (*downloadCheckpoint, error) {
	partsNum := headOutput.ContentLength / input.PartSize
	remainder := headOutput.ContentLength % input.PartSize
	if remainder != 0 {
		partsNum++
	}
	parts := make([]downloadPartInfo, partsNum)
	for i := int64(0); i < partsNum; i++ {
		parts[i] = downloadPartInfo{
			PartNumber: int(i + 1),
			RangeStart: i * input.PartSize,
			RangeEnd:   (i+1)*input.PartSize - 1,
		}
	}
	if remainder != 0 {
		parts[partsNum-1].RangeEnd = (partsNum-1)*input.PartSize + remainder - 1
	}
	if len(parts) > 10000 {
		return nil, newTosClientError("tos: part count too many", nil)
	}
	return &downloadCheckpoint{
		checkpointPath:    input.CheckpointFile,
		Bucket:            input.Bucket,
		Key:               input.Key,
		VersionID:         input.VersionID,
		PartSize:          input.PartSize,
		IfMatch:           input.IfMatch,
		IfModifiedSince:   input.IfModifiedSince,
		IfNoneMatch:       input.IfNoneMatch,
		IfUnmodifiedSince: input.IfUnmodifiedSince,
		SSECAlgorithm:     input.SSECAlgorithm,
		SSECKeyMD5:        input.SSECKey,
		ObjectInfo: downloadObjectInfo{
			Etag:          headOutput.ETag,
			HashCrc64ecma: headOutput.HashCrc64ecma,
			LastModified:  headOutput.LastModified,
			ObjectSize:    headOutput.ContentLength,
		},
		FileInfo: downloadFileInfo{
			FilePath:     input.FilePath,
			TempFilePath: input.tempFile,
		},
		PartsInfo: parts,
	}, nil
}

func createTempFile(input *DownloadFileInput, event downloadEvent) error {
	_, err := os.Create(input.tempFile)
	if err != nil {
		event.postDownloadEvent(&DownloadEvent{
			Type:      enum.DownloadEventCreateTempFileFailed,
			Bucket:    input.Bucket,
			Key:       input.Key,
			VersionID: input.VersionID,
			FilePath:  input.FilePath,
		})
		return newTosClientError("tos: create temp file failed.", err)
	}
	event.postDownloadEvent(&DownloadEvent{
		Type:         enum.DownloadEventCreateTempFileSucceed,
		Bucket:       input.Bucket,
		Key:          input.Key,
		VersionID:    input.VersionID,
		FilePath:     input.FilePath,
		TempFilePath: &input.FilePath,
	})
	return nil
}

func getDownloadTasks(cli *ClientV2, ctx context.Context, headOutput *HeadObjectV2Output,
	checkpoint *downloadCheckpoint, input *DownloadFileInput) []task {
	tasks := make([]task, 0)
	consumed := int64(0)
	subtotal := int64(0)
	for _, part := range checkpoint.PartsInfo {
		if !part.IsCompleted {
			tasks = append(tasks, &downloadTask{
				cli:         cli,
				ctx:         ctx,
				input:       input,
				partNumber:  part.PartNumber,
				rangeStart:  part.RangeStart,
				rangeEnd:    part.RangeEnd,
				consumed:    &consumed,
				subtotal:    &subtotal,
				total:       headOutput.ContentLength,
				enableCRC64: cli.enableCRC,
			})
		}
	}
	return tasks
}

func (d downloadEvent) newDownloadEvent() *DownloadEvent {
	return &DownloadEvent{
		Bucket:         d.input.Bucket,
		Key:            d.input.Key,
		VersionID:      d.input.VersionID,
		FilePath:       d.input.FilePath,
		CheckpointFile: &d.input.CheckpointFile,
		TempFilePath:   &d.input.tempFile,
	}
}

func (d downloadEvent) newDownloadPartSucceedEvent(part downloadPartInfo) *DownloadEvent {
	event := d.newSucceedEvent(enum.DownloadEventDownloadPartSucceed)
	event.DowloadPartInfo = &DownloadPartInfo{
		PartNumber: part.PartNumber,
		RangeStart: part.RangeStart,
		RangeEnd:   part.RangeEnd,
	}
	return event
}

func (d downloadEvent) newSucceedEvent(eventType enum.DownloadEventType) *DownloadEvent {
	event := d.newDownloadEvent()
	event.Type = eventType
	return event
}

func (d downloadEvent) newFailedEvent(err error, eventType enum.DownloadEventType) *DownloadEvent {
	event := d.newDownloadEvent()
	event.Type = eventType
	event.Err = err
	return event
}

func (d downloadEvent) postDownloadEvent(event *DownloadEvent) {
	if d.input.DownloadEventListener != nil {
		d.input.DownloadEventListener.EventChange(event)
	}
}

func (cli *ClientV2) downloadFile(ctx context.Context,
	headOutput *HeadObjectV2Output, checkpoint *downloadCheckpoint, input *DownloadFileInput, event downloadEvent) (*DownloadFileOutput, error) {
	// prepare tasks
	tasks := getDownloadTasks(cli, ctx, headOutput, checkpoint, input)
	routinesNum := min(input.TaskNum, len(tasks))
	tg := newTaskGroup(getCancelHandle(input.CancelHook), routinesNum, checkpoint, event, input.EnableCheckpoint, tasks)
	tg.RunWorker()
	// start adding tasks
	postDataTransferStatus(input.DataTransferListener, &DataTransferStatus{
		Type: enum.DataTransferStarted,
	})
	tg.Scheduler()
	success, err := tg.Wait()
	if err != nil {
		_ = os.Remove(input.tempFile)
	}

	if success < len(tasks) {
		return nil, newTosClientError("tos: some download task failed.", nil)
	}
	// Check CRC64
	if cli.enableCRC && headOutput.HashCrc64ecma != 0 && combineCRCInDownload(checkpoint.PartsInfo) != headOutput.HashCrc64ecma {
		return nil, newTosClientError("tos: crc of entire file mismatch.", nil)
	}
	err = os.Rename(input.tempFile, input.FilePath)
	if err != nil {
		event.postDownloadEvent(event.newFailedEvent(err, enum.DownloadEventRenameTempFileFailed))
		return nil, err
	}
	event.postDownloadEvent(event.newSucceedEvent(enum.DownloadEventRenameTempFileSucceed))
	_ = os.Remove(checkpoint.checkpointPath)
	return &DownloadFileOutput{*headOutput}, nil
}
