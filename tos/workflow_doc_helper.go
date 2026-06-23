package tos

import "fmt"

// NewDocConvertToPDFWorkflowJobInput 构造文档转 PDF 的工作流 Job 输入。
func NewDocConvertToPDFWorkflowJobInput(params CreateDocConvertToPDFJobParams) (*CreateWorkflowJobInput, error) {
	return newDocWorkflowJobInput(workflowDocJobBuildParams{
		bucket:       params.Bucket,
		region:       params.Region,
		jobType:      WorkflowJobTypeDocConvert,
		sourceKey:    params.SourceKey,
		sourceType:   params.SourceType,
		targetType:   WorkflowDocTypePDF,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
	})
}

// NewDocConvertToImageWorkflowJobInput 构造文档转图片的工作流 Job 输入，支持指定页码范围。
func NewDocConvertToImageWorkflowJobInput(params CreateDocConvertToImageJobParams) (*CreateWorkflowJobInput, error) {
	return newDocWorkflowJobInput(workflowDocJobBuildParams{
		bucket:       params.Bucket,
		region:       params.Region,
		jobType:      WorkflowJobTypeDocConvert,
		sourceKey:    params.SourceKey,
		sourceType:   params.SourceType,
		targetType:   WorkflowDocTypeJPG,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
		startPage:    params.StartPage,
		endPage:      params.EndPage,
	})
}

type workflowDocJobBuildParams struct {
	bucket       string
	region       string
	jobType      WorkflowJobType
	sourceKey    string
	sourceType   WorkflowDocType
	targetType   WorkflowDocType
	outputBucket string
	outputObject string
	startPage    *int
	endPage      *int
}

func newDocWorkflowJobInput(params workflowDocJobBuildParams) (*CreateWorkflowJobInput, error) {
	if params.bucket == "" {
		return nil, fmt.Errorf("tos: Bucket is required")
	}
	if params.sourceKey == "" {
		return nil, fmt.Errorf("tos: SourceKey is required")
	}
	if params.outputObject == "" {
		return nil, fmt.Errorf("tos: OutputObject is required")
	}

	outputBucket := params.outputBucket
	if outputBucket == "" {
		outputBucket = params.bucket
	}

	jobDetail := &WorkflowDocConvertProcessInput{
		Input: WorkflowDocConvertInput{
			Key: params.sourceKey,
		},
		DocConvertConfig: WorkflowDocConvertConfig{
			SrcType: params.sourceType,
			TgtType: params.targetType,
		},
		Output: WorkflowDocConvertOutput{
			Region: params.region,
			Bucket: outputBucket,
			Object: params.outputObject,
		},
	}

	if params.startPage != nil {
		jobDetail.DocConvertConfig.StartPage = params.startPage
	}
	if params.endPage != nil {
		jobDetail.DocConvertConfig.EndPage = params.endPage
	}

	return &CreateWorkflowJobInput{
		Bucket:    params.bucket,
		JobType:   params.jobType,
		JobDetail: jobDetail,
	}, nil
}

// NewAudioConvertWorkflowJobInput 构造音频转码的工作流 Job 输入。
func NewAudioConvertWorkflowJobInput(params CreateAudioConvertWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.ConvertParams == nil {
		return nil, fmt.Errorf("tos: ConvertParams is required")
	}
	return newMediaWorkflowJobInput(workflowMediaJobBuildParams{
		bucket:       params.Bucket,
		jobType:      WorkflowJobTypeAudioConvert,
		sourceKey:    params.SourceKey,
		region:       params.Region,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
		jobDetail: &WorkflowAudioJobDetail{
			TemplateID:         params.TemplateID,
			AudioConvertConfig: params.ConvertParams,
		},
	})
}

// NewAudioConcatWorkflowJobInput 构造音频拼接的工作流 Job 输入。
func NewAudioConcatWorkflowJobInput(params CreateAudioConcatWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.ConcatParams == nil {
		return nil, fmt.Errorf("tos: ConcatParams is required")
	}
	return newMediaWorkflowJobInput(workflowMediaJobBuildParams{
		bucket:       params.Bucket,
		jobType:      WorkflowJobTypeAudioConcat,
		sourceKey:    params.SourceKey,
		region:       params.Region,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
		jobDetail: &WorkflowAudioJobDetail{
			AudioConcatConfig: params.ConcatParams,
		},
	})
}

// NewVideoTranscodeWorkflowJobInput 构造视频转码的工作流 Job 输入。
func NewVideoTranscodeWorkflowJobInput(params CreateVideoTranscodeWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.TranscodeConfig == nil {
		return nil, fmt.Errorf("tos: TranscodeConfig is required")
	}
	return newMediaWorkflowJobInput(workflowMediaJobBuildParams{
		bucket:       params.Bucket,
		jobType:      WorkflowJobTypeTranscode,
		sourceKey:    params.SourceKey,
		region:       params.Region,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
		jobDetail: &WorkflowVideoJobDetail{
			TranscodeConfig: &WorkflowVideoTranscodeConfig{
				TemplateID:          params.TemplateID,
				Transcode:           params.TranscodeConfig.Transcode,
				WatermarkTemplateID: params.WatermarkTemplateID,
				Watermark:           params.TranscodeConfig.Watermark,
				DigitalWatermark:    params.TranscodeConfig.DigitalWatermark,
			},
		},
	})
}

// NewVideoRemuxWorkflowJobInput 构造视频转封装的工作流 Job 输入。
func NewVideoRemuxWorkflowJobInput(params CreateVideoRemuxWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.RemuxConfig == nil {
		return nil, fmt.Errorf("tos: RemuxConfig is required")
	}
	return newMediaWorkflowJobInput(workflowMediaJobBuildParams{
		bucket:       params.Bucket,
		jobType:      WorkflowJobTypeRemux,
		sourceKey:    params.SourceKey,
		region:       params.Region,
		outputBucket: params.OutputBucket,
		outputObject: params.OutputObject,
		jobDetail: &WorkflowVideoJobDetail{
			RemuxConfig: params.RemuxConfig,
		},
	})
}

// NewFileCompressWorkflowJobInput 构造文件压缩的工作流 Job 输入，SourceKeys 和 SourcePrefix 二选一。
func NewFileCompressWorkflowJobInput(params CreateFileCompressWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.Region == "" {
		return nil, fmt.Errorf("tos: Region is required")
	}
	if params.OutputObject == "" {
		return nil, fmt.Errorf("tos: OutputObject is required")
	}
	if len(params.SourceKeys) == 0 && params.SourcePrefix == "" {
		return nil, fmt.Errorf("tos: SourceKeys or SourcePrefix is required")
	}
	if params.Format == "" {
		params.Format = "zip"
	}

	outputBucket := params.OutputBucket
	if outputBucket == "" {
		outputBucket = params.Bucket
	}

	input := WorkflowFileCompressInput{}
	if len(params.SourceKeys) > 0 {
		keys := make([]WorkflowFileCompressKeyConfig, len(params.SourceKeys))
		for i, k := range params.SourceKeys {
			keys[i] = WorkflowFileCompressKeyConfig{Key: k}
		}
		input.KeyConfig = keys
	}
	if params.SourcePrefix != "" {
		input.Prefix = params.SourcePrefix
	}

	return &CreateWorkflowJobInput{
		Bucket:  params.Bucket,
		JobType: WorkflowJobTypeFileCompress,
		JobDetail: &WorkflowFileJobDetail{
			Input: WorkflowFileJobInput{CompressInput: &input},
			FileCompressConfig: &WorkflowFileCompressConfig{
				Format:  params.Format,
				Flatten: params.Flatten,
			},
			Output: WorkflowFileJobOutput{
				Region: params.Region,
				Bucket: outputBucket,
				Object: params.OutputObject,
			},
		},
	}, nil
}

// NewFileUncompressWorkflowJobInput 构造文件解压缩的工作流 Job 输入。
func NewFileUncompressWorkflowJobInput(params CreateFileUncompressWorkflowJobParams) (*CreateWorkflowJobInput, error) {
	if params.Region == "" {
		return nil, fmt.Errorf("tos: Region is required")
	}
	if params.SourceKey == "" {
		return nil, fmt.Errorf("tos: SourceKey is required")
	}

	outputBucket := params.OutputBucket
	if outputBucket == "" {
		outputBucket = params.Bucket
	}

	return &CreateWorkflowJobInput{
		Bucket:  params.Bucket,
		JobType: WorkflowJobTypeFileUncompress,
		JobDetail: &WorkflowFileJobDetail{
			Input: WorkflowFileJobInput{UncompressInput: &WorkflowFileUncompressInput{
				Object: params.SourceKey,
			}},
			FileUncompressConfig: &WorkflowFileUncompressConfig{
				Prefix:         params.Prefix,
				PrefixReplaced: params.PrefixReplaced,
			},
			Output: WorkflowFileJobOutput{
				Region: params.Region,
				Bucket: outputBucket,
			},
		},
	}, nil
}

type workflowMediaJobBuildParams struct {
	bucket       string
	jobType      WorkflowJobType
	sourceKey    string
	region       string
	outputBucket string
	outputObject string
	jobDetail    interface{}
}

func newMediaWorkflowJobInput(params workflowMediaJobBuildParams) (*CreateWorkflowJobInput, error) {
	if params.bucket == "" {
		return nil, fmt.Errorf("tos: Bucket is required")
	}
	if params.sourceKey == "" {
		return nil, fmt.Errorf("tos: SourceKey is required")
	}
	if params.region == "" {
		return nil, fmt.Errorf("tos: Region is required")
	}
	if params.outputObject == "" {
		return nil, fmt.Errorf("tos: OutputObject is required")
	}
	if params.jobDetail == nil {
		return nil, fmt.Errorf("tos: JobDetail is required")
	}

	outputBucket := params.outputBucket
	if outputBucket == "" {
		outputBucket = params.bucket
	}

	switch detail := params.jobDetail.(type) {
	case *WorkflowAudioJobDetail:
		detail.Input.Object = params.sourceKey
		detail.Output = ProcessJobOutput{
			Region: params.region,
			Bucket: outputBucket,
			Object: params.outputObject,
		}
		detail.Tag = string(params.jobType)
	case *WorkflowVideoJobDetail:
		detail.Input.Object = params.sourceKey
		detail.Output = ProcessJobOutput{
			Region: params.region,
			Bucket: outputBucket,
			Object: params.outputObject,
		}
		detail.Tag = string(params.jobType)
	default:
		return nil, fmt.Errorf("tos: unsupported workflow media job detail type %T", params.jobDetail)
	}

	return &CreateWorkflowJobInput{
		Bucket:    params.bucket,
		JobType:   params.jobType,
		JobDetail: params.jobDetail,
	}, nil
}
