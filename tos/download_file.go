package tos

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// func getDownloadCheckpoint(enabled bool, checkpointPath string, init func() (*downloadCheckpoint, error)) (checkpoint *downloadCheckpoint, err error) {
// 	if enabled {
// 		_, err = os.Stat(checkpointPath)
// 		// if err is not empty, assume checkpoint not exists
// 		if err == nil {
// 			loadCheckPoint(checkpointPath, checkpoint)
// 			if checkpoint != nil {
// 				return
// 			}
// 		}
// 		_, err = os.Create(checkpointPath)
// 		if err != nil {
// 			return nil, newTosClientError(err.Error(), err)
// 		}
// 	}
// 	checkpoint, err = init()
// 	if err != nil {
// 		return nil, err
// 	}
// 	if enabled {
// 		err = checkpoint.WriteToFile()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return
// }

// func (cli *ClientV2) DownloadFile(ctx context.Context, input *DownloadFileInput) (*DownloadFileOutput, error) {
// 	err := validateDownloadInput(input)
// 	if err != nil {
// 		return nil, err
// 	}
// 	headOutput, err := cli.HeadObjectV2(ctx, &input.HeadObjectV2Input)
// 	if err != nil {
// 		return nil, err
// 	}
// 	init := func() (*downloadCheckpoint, error) {
// 		err := createTempFile(input.tempFile, input.Bucket, input.Key,
// 			input.VersionID, input.FilePath, input.DownloadEventListener)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return initDownloadCheckpoint(input, headOutput)
// 	}
// 	checkpoint, err := getDownloadCheckpoint(input.EnableCheckpoint, input.CheckpointFile, init)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return cli.downloadFile(ctx, headOutput, checkpoint, input)
// }

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

// func validateDownloadInput(input *DownloadFileInput) error {
// 	if err := isValidNames(input.Bucket, input.Key); err != nil {
// 		return err
// 	}
// 	if input.PartSize == 0 {
// 		input.PartSize = MinPartSize
// 	}
// 	if input.PartSize < MinPartSize || input.PartSize > MaxPartSize {
// 		return newTosClientError("The input part size is invalid, please set it range from 5MB to 5GB", nil)
// 	}
// 	// if directory, append object key at end
// 	mustFile(&input.FilePath, input.Key)
// 	input.tempFile = input.FilePath + TempFileSuffix
// 	if input.EnableCheckpoint {
// 		// get correct checkpoint path
// 		if len(input.CheckpointFile) == 0 {
// 			dirName, _ := filepath.Split(input.FilePath)
// 			fileName := strings.Join([]string{input.FilePath, input.Bucket, input.Key, "download"}, ".")
// 			input.CheckpointFile = filepath.Join(dirName, fileName)
// 		} else {
// 			mustFile(&input.CheckpointFile, strings.Join([]string{input.FilePath, input.Bucket, input.Key, "download"}, "."))
// 		}
// 	}
// 	if input.TaskNum < 1 {
// 		input.TaskNum = 1
// 	}
// 	if input.TaskNum > 1000 {
// 		input.TaskNum = 1000
// 	}
// 	return nil
// }

// func initDownloadCheckpoint(input *DownloadFileInput, headOutput *HeadObjectV2Output) (*downloadCheckpoint, error) {
// 	partsNum := headOutput.ContentLength / input.PartSize
// 	remainder := headOutput.ContentLength % input.PartSize
// 	if remainder != 0 {
// 		partsNum++
// 	}
// 	parts := make([]downloadPartInfo, partsNum)
// 	for i := int64(0); i < partsNum; i++ {
// 		parts[i] = downloadPartInfo{
// 			PartNumber: int(i + 1),
// 			RangeStart: i * input.PartSize,
// 			RangeEnd:   (i+1)*input.PartSize - 1,
// 		}
// 	}
// 	if remainder != 0 {
// 		parts[partsNum-1].RangeEnd = (partsNum-1)*input.PartSize + remainder - 1
// 	}
// 	if len(parts) > 10000 {
// 		return nil, newTosClientError("tos: part count too many", nil)
// 	}
// 	return &downloadCheckpoint{
// 		checkpointPath:    input.CheckpointFile,
// 		Bucket:            input.Bucket,
// 		Key:               input.Key,
// 		VersionID:         input.VersionID,
// 		PartSize:          input.PartSize,
// 		IfMatch:           input.IfMatch,
// 		IfModifiedSince:   input.IfModifiedSince,
// 		IfNoneMatch:       input.IfNoneMatch,
// 		IfUnmodifiedSince: input.IfUnmodifiedSince,
// 		SSECAlgorithm:     input.SSECAlgorithm,
// 		SSECKeyMD5:        input.SSECKey,
// 		ObjectInfo: downloadObjectInfo{
// 			Etag:          headOutput.ETag,
// 			HashCrc64ecma: headOutput.HashCrc64ecma,
// 			LastModified:  headOutput.LastModified,
// 			ObjectSize:    headOutput.ContentLength,
// 		},
// 		FileInfo: downloadFileInfo{
// 			FilePath:     input.FilePath,
// 			TempFilePath: input.tempFile,
// 		},
// 		PartsInfo: parts,
// 	}, nil
// }
//
// func createTempFile(tempFilePath string, bucket, key, versionID string, filePath string, listener DownloadEventListener) error {
// 	_, err := os.Create(tempFilePath)
// 	if err != nil {
// 		postDownloadEvent(listener, &DownloadEvent{
// 			Type:      enum.DownloadEventCreateTempFileFailed,
// 			Bucket:    bucket,
// 			Key:       key,
// 			VersionID: versionID,
// 			FilePath:  filePath,
// 		})
// 		return newTosClientError("tos: create temp file failed.", err)
// 	}
// 	postDownloadEvent(listener, &DownloadEvent{
// 		Type:         enum.DownloadEventCreateTempFileSucceed,
// 		Bucket:       bucket,
// 		Key:          key,
// 		VersionID:    versionID,
// 		FilePath:     filePath,
// 		TempFilePath: &tempFilePath,
// 	})
// 	return nil
// }

// func getDownloadTasks(cli *ClientV2, ctx context.Context, headOutput *HeadObjectV2Output,
// 	checkpoint *downloadCheckpoint, input *DownloadFileInput) []task {
// 	tasks := make([]task, 0)
// 	consumed := int64(0)
// 	mutex := sync.Mutex{}
// 	for _, part := range checkpoint.PartsInfo {
// 		if !part.IsCompleted {
// 			tasks = append(tasks, &downloadTask{
// 				cli:        cli,
// 				ctx:        ctx,
// 				input:      input,
// 				PartNumber: part.PartNumber,
// 				RangeStart: part.RangeStart,
// 				RangeEnd:   part.RangeEnd,
// 				consumed:   &consumed,
// 				total:      headOutput.ContentLength,
// 				mutex:      &mutex,
// 			})
// 		}
// 	}
// 	return tasks
// }
//
// func newDownloadEvent(input *DownloadFileInput) *DownloadEvent {
// 	return &DownloadEvent{
// 		Bucket:         input.Bucket,
// 		Key:            input.Key,
// 		VersionID:      input.VersionID,
// 		FilePath:       input.FilePath,
// 		CheckpointFile: &input.CheckpointFile,
// 		TempFilePath:   &input.tempFile,
// 	}
// }
//
// func newDownloadPartSucceedEvent(part downloadPartInfo, input *DownloadFileInput) *DownloadEvent {
// 	event := newSucceedEvent(enum.DownloadEventDownloadPartSucceed, input)
// 	event.DowloadPartInfo = &DownloadPartInfo{
// 		PartNumber: part.PartNumber,
// 		RangeStart: part.RangeStart,
// 		RangeEnd:   part.RangeEnd,
// 	}
// 	return event
// }
//
// func newSucceedEvent(eventType enum.DownloadEventType, input *DownloadFileInput) *DownloadEvent {
// 	event := newDownloadEvent(input)
// 	event.Type = eventType
// 	return event
// }
//
// func newFailedEvent(err error, eventType enum.DownloadEventType, input *DownloadFileInput) *DownloadEvent {
// 	event := newDownloadEvent(input)
// 	event.Type = eventType
// 	event.Err = err
// 	return event
// }
//
// func postDownloadEvent(listener DownloadEventListener, event *DownloadEvent) {
// 	if listener != nil {
// 		listener.EventChange(event)
// 	}
// }
//
// func (cli *ClientV2) downloadFile(ctx context.Context,
// 	headOutput *HeadObjectV2Output, checkpoint *downloadCheckpoint, input *DownloadFileInput) (*DownloadFileOutput, error) {
// 	// prepare tasks
// 	tasks := getDownloadTasks(cli, ctx, headOutput, checkpoint, input)
// 	if len(tasks) < input.TaskNum {
// 		input.TaskNum = len(tasks)
// 	}
// 	manager := newTaskManager(input.TaskNum)
// 	manager.addTask(tasks...)
// 	input.withCancelHook(canceler{
// 		called: false,
// 		cancelHandle: manager.cancelHandle,
// 		files:  []string{input.tempFile, input.CheckpointFile},
// 	})
// 	postDataTransferStatus(input.DataTransferListener, &DataTransferStatus{
// 		TotalBytes: checkpoint.ObjectInfo.ObjectSize,
// 		Type:       enum.DataTransferStarted,
// 	})
// 	manager.run()
// 	success := 0
// 	fail := make([]error, 0, len(tasks))
// Loop:
// 	for success+len(fail) < len(tasks) {
// 		select {
// 		case <-manager.cancelHandle:
// 			break Loop
// 		case result := <-manager.results:
// 			if part, ok := result.(downloadPartInfo); ok {
// 				success++
// 				checkpoint.UpdatePartsInfo(part)
// 				if input.EnableCheckpoint {
// 					err := checkpoint.WriteToFile()
// 					if err != nil {
// 						return nil, err
// 					}
// 				}
// 				postDownloadEvent(input.DownloadEventListener, newDownloadPartSucceedEvent(part, input))
// 				if success == len(tasks) {
// 					break Loop
// 				}
// 			}
// 		case err := <-manager.failed:
// 			if StatusCode(err) == 403 || StatusCode(err) == 404 || StatusCode(err) == 405 {
// 				close(manager.cancelHandle)
// 				postDownloadEvent(input.DownloadEventListener, newFailedEvent(err, enum.DownloadEventDownloadPartAborted, input))
// 				_ = os.Remove(input.CheckpointFile)
// 				_ = os.Remove(input.tempFile)
// 				break Loop
// 			} else {
// 				postDownloadEvent(input.DownloadEventListener, newFailedEvent(err, enum.DownloadEventDownloadPartFailed, input))
// 				fail = append(fail, err)
// 			}
// 		}
// 	}
//
// 	if success < len(tasks) {
// 		return nil, newTosClientError("tos: some download task failed.", nil)
// 	}
// 	err := checkFileCrc64(input.tempFile, headOutput.HashCrc64ecma)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = os.Rename(input.tempFile, input.FilePath)
// 	if err != nil {
// 		postDownloadEvent(input.DownloadEventListener, newFailedEvent(err, enum.DownloadEventRenameTempFileFailed, input))
// 		return nil, err
// 	}
// 	postDownloadEvent(input.DownloadEventListener, newSucceedEvent(enum.DownloadEventRenameTempFileSucceed, input))
// 	_ = os.Remove(checkpoint.checkpointPath)
// 	return &DownloadFileOutput{*headOutput}, nil
// }
