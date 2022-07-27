package tos

import (
	"context"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"os"
	"path/filepath"
	"strings"
)

// initUploadPartsInfo initialize parts info from file stat,return TosClientError if failed
func initUploadPartsInfo(uploadFileStat os.FileInfo, partSize int64) ([]uploadPartInfo, error) {
	partCount := uploadFileStat.Size() / partSize
	lastPartSize := uploadFileStat.Size() % partSize
	if lastPartSize != 0 {
		partCount++
	}
	if partCount > 10000 {
		return nil, newTosClientError("tos: part count too many", nil)
	}
	parts := make([]uploadPartInfo, 0, partCount)
	for i := int64(0); i < partCount; i++ {
		part := uploadPartInfo{
			PartNumber: int(i + 1),
			PartSize:   partSize,
			Offset:     uint64(i * partSize),
		}
		parts = append(parts, part)
	}
	if lastPartSize != 0 {
		parts[partCount-1].PartSize = lastPartSize
	}
	return parts, nil
}

// initUploadCheckpoint initialize checkpoint file, return TosClientError if failed
func initUploadCheckpoint(input *UploadFileInput, created *CreateMultipartUploadV2Output) (*uploadCheckpoint, error) {
	stat, err := os.Stat(input.FilePath)
	if err != nil {
		return nil, newTosClientError(err.Error(), err)
	}
	parts, err := initUploadPartsInfo(stat, input.PartSize)
	if err != nil {
		return nil, err
	}
	checkPoint := &uploadCheckpoint{
		checkpointPath: input.CheckpointFile,
		UploadID:       created.UploadID,
		PartsInfo:      parts,
		Bucket:         input.Bucket,
		Key:            input.Key,
		PartSize:       input.PartSize,
		SSECAlgorithm:  input.SSECAlgorithm,
		SSECKeyMD5:     input.SSECKeyMD5,
		EncodingType:   input.ContentEncoding,
		FilePath:       input.FilePath,
		FileInfo: fileInfo{
			Size:         stat.Size(),
			LastModified: stat.ModTime().Unix(),
		},
	}
	return checkPoint, nil
}

// validateUploadInput validate upload input, return TosClientError failed
func validateUploadInput(input *UploadFileInput) error {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return err
	}
	if input.PartSize == 0 {
		input.PartSize = MinPartSize
	}
	if input.PartSize < MinPartSize || input.PartSize > MaxPartSize {
		return newTosClientError("tos: the input part size is invalid, please set it range from 5MB to 5GB.", nil)
	}
	stat, err := os.Stat(input.FilePath)
	if err != nil {
		return newTosClientError("tos: stat file to upload failed", err)
	}
	if stat.IsDir() {
		return newTosClientError("tos: does not support directory, please specific your file path.", nil)
	}
	if input.EnableCheckpoint {
		// get correct checkpoint path
		if len(input.CheckpointFile) == 0 {
			dirName, _ := filepath.Split(input.FilePath)
			fileName := strings.Join([]string{input.FilePath, input.Bucket, input.Key, "upload"}, ".")
			input.CheckpointFile = filepath.Join(dirName, fileName)
		} else {
			mustFile(&input.CheckpointFile, strings.Join([]string{input.FilePath, input.Bucket, input.Key, "upload"}, "."))
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

func postUploadEvent(listener UploadEventListener, event *UploadEvent) {
	if listener != nil {
		listener.EventChange(event)
	}
}

// getUploadCheckpoint get struct checkpoint from checkpoint file if checkpointPath is valid,
// or initialize from scratch with function init
func getUploadCheckpoint(enabled bool, checkpointPath string,
	init func() (*uploadCheckpoint, error)) (checkpoint *uploadCheckpoint, err error) {
	if enabled {
		_, err = os.Stat(checkpointPath)
		// if err is not empty, assume checkpoint not exists
		if err == nil {
			loadCheckPoint(checkpointPath, checkpoint)
			if checkpoint != nil {
				return
			}
		}
		_, err = os.Create(checkpointPath)
		if err != nil {
			return nil, newTosClientError("tos: create checkpoint file failed", err)
		}
	}
	checkpoint, err = init()
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

func bindCancelHookWithAborter(hook CancelHook, aborter func() error) {
	if hook == nil {
		return
	}
	cancel := hook.(*canceler)
	cancel.aborter = aborter
}

func bindCancelHookWithCleaner(hook CancelHook, cleaner func()) {
	if hook == nil {
		return
	}
	cancel := hook.(*canceler)
	cancel.cleaner = cleaner
}

func (cli *ClientV2) UploadFile(ctx context.Context, input *UploadFileInput) (output *UploadFileOutput, err error) {
	// avoid modifying on origin pointer
	input = &(*input)
	if err = validateUploadInput(input); err != nil {
		return nil, err
	}
	// create multipart upload task
	created, err := cli.CreateMultipartUploadV2(ctx, &input.CreateMultipartUploadV2Input)
	if err != nil {
		postUploadEvent(input.UploadEventListener, &UploadEvent{
			Type:           enum.UploadEventCreateMultipartUploadFailed,
			Err:            err,
			Bucket:         input.Bucket,
			Key:            input.Key,
			CheckpointFile: &input.CheckpointFile,
		})
		return nil, err
	}
	postUploadEvent(input.UploadEventListener, &UploadEvent{
		Type:           enum.UploadEventCreateMultipartUploadSucceed,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       &created.UploadID,
		CheckpointFile: &input.CheckpointFile,
	})
	init := func() (*uploadCheckpoint, error) {
		return initUploadCheckpoint(input, created)
	}
	// if the checkpoint file not exist, here we will create it
	checkpoint, err := getUploadCheckpoint(input.EnableCheckpoint, input.CheckpointFile, init)
	if err != nil {
		return nil, err
	}
	cleaner := func() {
		_ = os.Remove(input.CheckpointFile)
	}
	bindCancelHookWithCleaner(input.CancelHook, cleaner)
	return cli.uploadPart(ctx, checkpoint, input)
}

func prepareUploadTasks(cli *ClientV2, ctx context.Context, checkpoint *uploadCheckpoint, input *UploadFileInput) []task {
	tasks := make([]task, 0)
	consumed := int64(0)
	subtotal := int64(0)
	for _, part := range checkpoint.PartsInfo {
		if !part.IsCompleted {
			tasks = append(tasks, &uploadTask{
				cli:        cli,
				ctx:        ctx,
				input:      input,
				total:      checkpoint.FileInfo.Size,
				UploadID:   checkpoint.UploadID,
				PartNumber: part.PartNumber,
				subtotal:   &subtotal,
				consumed:   &consumed,
				Offset:     part.Offset,
				PartSize:   part.PartSize,
			})
		}
	}
	return tasks
}

func newUploadPartSucceedEvent(input *UploadFileInput, part uploadPartInfo) *UploadEvent {
	return &UploadEvent{
		Type:           enum.UploadEventUploadPartSucceed,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       part.uploadID,
		CheckpointFile: &input.CheckpointFile,
		UploadPartInfo: &UploadPartInfo{
			PartNumber:    part.PartNumber,
			PartSize:      part.PartSize,
			Offset:        int64(part.Offset),
			ETag:          &part.ETag,
			HashCrc64ecma: &part.HashCrc64ecma,
		},
	}
}

func newUploadPartAbortedEvent(input *UploadFileInput, uploadID string, err error) *UploadEvent {
	return &UploadEvent{
		Type:           enum.UploadEventUploadPartAborted,
		Err:            err,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       &uploadID,
		CheckpointFile: &input.CheckpointFile,
	}
}

func newUploadPartFailedEvent(input *UploadFileInput, uploadID string, err error) *UploadEvent {
	return &UploadEvent{
		Type:           enum.UploadEventUploadPartFailed,
		Err:            err,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       &uploadID,
		CheckpointFile: &input.CheckpointFile,
	}
}

func newCompleteMultipartUploadFailedEvent(input *UploadFileInput, uploadID string, err error) *UploadEvent {
	return &UploadEvent{
		Type:           enum.UploadEventCompleteMultipartUploadFailed,
		Err:            err,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       &uploadID,
		CheckpointFile: &input.CheckpointFile,
	}
}

func newCompleteMultipartUploadSucceedEvent(input *UploadFileInput, uploadID string) *UploadEvent {
	return &UploadEvent{
		Type:           enum.UploadEventCompleteMultipartUploadSucceed,
		Bucket:         input.Bucket,
		Key:            input.Key,
		UploadID:       &uploadID,
		CheckpointFile: &input.CheckpointFile,
	}
}

// checkFileCrc64 check if crc64 checksum of file is expected. Return TosClientError if open file failed, or return
// TosServerError if check sum mismatch
//func checkFileCrc64(filepath string, want uint64) error {
//	fd, err := os.Open(filepath)
//	if err != nil {
//		return newTosClientError(err.Error(), err)
//	}
//	defer fd.Close()
//	crc := crc64.New(DefaultCrcTable())
//	_, err = io.Copy(crc, fd)
//	if err != nil {
//		return err
//	}
//	if crc.Sum64() != want {
//		// data returned by server is invalid, or we encounter a bug in sdk
//		return &TosServerError{
//			TosError: TosError{"tos: crc of entire file mismatch."},
//		}
//	}
//	return nil
//}

func postDataTransferStatus(listener DataTransferListener, status *DataTransferStatus) {
	if listener != nil {
		listener.DataTransferStatusChange(status)
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func getCancelHandle(hook CancelHook) chan struct{} {
	if c, ok := hook.(*canceler); ok {
		return c.cancelHandle
	}
	return make(chan struct{})
}

// combineCRCInParts calculates the total CRC of continuous parts
func combineCRCInParts(parts []uploadPartInfo) uint64 {
	if parts == nil || len(parts) == 0 {
		return 0
	}
	crc := parts[0].HashCrc64ecma
	for i := 1; i < len(parts); i++ {
		crc = CRC64Combine(crc, parts[i].HashCrc64ecma, uint64(parts[i].PartSize))
	}
	return crc
}

func (cli *ClientV2) uploadPart(ctx context.Context, checkpoint *uploadCheckpoint, input *UploadFileInput) (*UploadFileOutput, error) {
	// prepare tasks
	// if amount of tasks >= 10000, err "tos: part count too many" will be raised.
	tasks := prepareUploadTasks(cli, ctx, checkpoint, input)
	routinesNum := min(input.TaskNum, len(tasks))
	taskBufferSize := min(routinesNum, DefaultTaskBufferSize)
	tasksCh := make(chan task, taskBufferSize)
	resultsCh := make(chan uploadPartInfo)
	errCh := make(chan error)
	cancelHandle := getCancelHandle(input.CancelHook)
	abortHandle := make(chan struct{})
	worker := func() {
		for {
			select {
			case <-cancelHandle:
				return
			case <-abortHandle:
				return
			case t, ok := <-tasksCh:
				if !ok {
					return
				}
				result, err := t.do()
				if err != nil {
					errCh <- err
				}
				if part, ok := result.(uploadPartInfo); ok {
					resultsCh <- part
				}
			}
		}
	}
	scheduler := func() {
		func() {
			for _, t := range tasks {
				select {
				case <-cancelHandle:
					return
				case <-abortHandle:
					return
				default:
					tasksCh <- t
				}
			}
		}()
		close(tasksCh)
	}

	aborter := func() error {
		_, err := cli.AbortMultipartUpload(ctx,
			&AbortMultipartUploadInput{
				Bucket:   input.Bucket,
				Key:      input.Key,
				UploadID: checkpoint.UploadID})
		return err
	}
	bindCancelHookWithAborter(input.CancelHook, aborter)

	// start running workers
	for i := 0; i < routinesNum; i++ {
		go worker()
	}
	// start adding tasks
	postDataTransferStatus(input.DataTransferListener, &DataTransferStatus{
		TotalBytes: checkpoint.FileInfo.Size,
		Type:       enum.DataTransferStarted,
	})
	go scheduler()
	success := 0
	fails := 0
	// processing tasks
Loop:
	for success+fails < len(tasks) {
		select {
		case <-abortHandle:
			break Loop
		case <-cancelHandle:
			break Loop
		case part := <-resultsCh:
			success++
			checkpoint.UpdatePartsInfo(part)
			if input.EnableCheckpoint {
				checkpoint.WriteToFile()
			}
			postUploadEvent(input.UploadEventListener, newUploadPartSucceedEvent(input, part))
		case taskErr := <-errCh:
			if StatusCode(taskErr) == 403 || StatusCode(taskErr) == 404 || StatusCode(taskErr) == 405 {
				close(abortHandle)
				_ = os.Remove(input.CheckpointFile)
				if err := aborter(); err != nil {
					// TODO: log abort err
					return nil, taskErr
				}
				postUploadEvent(input.UploadEventListener, newUploadPartAbortedEvent(input, checkpoint.UploadID, taskErr))
				break Loop
			} else {
				postUploadEvent(input.UploadEventListener, newUploadPartFailedEvent(input, checkpoint.UploadID, taskErr))
				fails++
			}
		}
	}
	// handle results
	if success < len(tasks) {
		return nil, newTosClientError("tos: some upload tasks failed.", nil)
	}
	complete, err := cli.CompleteMultipartUploadV2(ctx, &CompleteMultipartUploadV2Input{
		Bucket:   input.Bucket,
		Key:      input.Key,
		UploadID: checkpoint.UploadID,
		Parts:    checkpoint.GetParts(),
	})
	if err != nil {
		postUploadEvent(input.UploadEventListener, newCompleteMultipartUploadFailedEvent(input, checkpoint.UploadID, err))
		return nil, err
	}
	postUploadEvent(input.UploadEventListener, newCompleteMultipartUploadSucceedEvent(input, checkpoint.UploadID))

	if combineCRCInParts(checkpoint.PartsInfo) != complete.HashCrc64ecma {
		return nil, &TosServerError{
			TosError: TosError{"tos: crc of entire file mismatch."},
		}

	}
	_ = os.Remove(input.CheckpointFile)

	return &UploadFileOutput{
		RequestInfo:   complete.RequestInfo,
		Bucket:        complete.Bucket,
		Key:           complete.Key,
		UploadID:      checkpoint.UploadID,
		ETag:          complete.ETag,
		Location:      complete.Location,
		VersionID:     complete.VersionID,
		HashCrc64ecma: complete.HashCrc64ecma,
		SSECAlgorithm: checkpoint.SSECAlgorithm,
		SSECKeyMD5:    checkpoint.SSECKeyMD5,
		EncodingType:  checkpoint.EncodingType,
	}, nil
}
