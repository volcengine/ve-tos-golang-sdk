package tos

import "encoding/json"

// PutConvertWorkflowInput 创建或更新工作流的入参。
type PutConvertWorkflowInput struct {
	Bucket string
	Role   string         `json:"Role,omitempty"`
	Rules  []WorkflowRule `json:"Rules,omitempty"`
	GenericInput
}

type PutConvertWorkflowOutput struct {
	RequestInfo
}

type GetConvertWorkflowInput struct {
	Bucket string
	GenericInput
}

type GetConvertWorkflowOutput struct {
	RequestInfo
	Role  string         `json:"Role,omitempty"`
	Rules []WorkflowRule `json:"Rules,omitempty"`
}

type DeleteConvertWorkflowInput struct {
	Bucket string
	GenericInput
}

type DeleteConvertWorkflowOutput struct {
	RequestInfo
}

// WorkflowRule 描述一条工作流规则，定义触发前缀、文件类型过滤、拓扑和操作。
type WorkflowRule struct {
	ID           string                `json:"ID,omitempty"`
	Enabled      bool                  `json:"Enabled,omitempty"`
	Prefix       string                `json:"Prefix,omitempty"`
	ExtFilter    *WorkflowExtFilter    `json:"ExtFilter,omitempty"`
	Topology     [][]string            `json:"Topology,omitempty"`
	Operations   WorkflowOperations    `json:"Operations,omitempty"`
	NotifyConfig *WorkflowNotifyConfig `json:"NotifyConfig,omitempty"`
}

type WorkflowExtFilter struct {
	AudioExts []string `json:"AudioExts,omitempty"`
	VideoExts []string `json:"VideoExts,omitempty"`
}

type WorkflowNotifyConfig struct {
	Enabled bool   `json:"Enabled,omitempty"`
	URL     string `json:"URL,omitempty"`
}

type WorkflowOperations struct {
	AudioTranscode []WorkflowOperationAudioTranscode `json:"AudioTranscode,omitempty"`
	Transcode      []WorkflowOperationTranscode      `json:"Transcode,omitempty"`
}

type WorkflowJobOutput struct {
	Region string `json:"Region,omitempty"`
	Bucket string `json:"Bucket,omitempty"`
	Object string `json:"Object,omitempty"`
}

type WorkflowOperationAudioTranscode struct {
	OperationID string            `json:"OperationID,omitempty"`
	TemplateID  string            `json:"TemplateID,omitempty"`
	Output      WorkflowJobOutput `json:"Output,omitempty"`
}

type WorkflowOperationTranscode struct {
	OperationID         string            `json:"OperationID,omitempty"`
	TemplateID          string            `json:"TemplateID,omitempty"`
	WatermarkTemplateID []string          `json:"WatermarkTemplateID,omitempty"`
	Output              WorkflowJobOutput `json:"Output,omitempty"`
}

type GetWorkflowExecutionInput struct {
	Bucket      string
	ExecutionID string
	GenericInput
}

type GetWorkflowExecutionOutput struct {
	RequestInfo
	Items []*WorkflowExecution `json:"Items,omitempty"`
}

type ListWorkflowExecutionInput struct {
	Bucket    string
	PageSize  int
	PageToken string
	StartTime int64
	EndTime   int64
	GenericInput
}

type ListWorkflowExecutionOutput struct {
	RequestInfo
	Items         []*WorkflowExecution `json:"Items,omitempty"`
	NextPageToken string               `json:"NextPageToken,omitempty"`
}

type WorkflowExecution struct {
	ExecutionID string          `json:"ExecutionID,omitempty"`
	Object      string          `json:"Object,omitempty"`
	Workflow    WorkflowRule    `json:"Workflow,omitempty"`
	State       string          `json:"State,omitempty"`
	CreateTime  string          `json:"CreateTime,omitempty"`
	StartTime   string          `json:"StartTime,omitempty"`
	EndTime     string          `json:"EndTime,omitempty"`
	Tasks       []*WorkflowTask `json:"Tasks,omitempty"`
}

type WorkflowTask struct {
	Type        string `json:"Type,omitempty"`
	JobID       string `json:"JobID,omitempty"`
	OperationID string `json:"OperationID,omitempty"`
	State       string `json:"State,omitempty"`
	CreateTime  string `json:"CreateTime,omitempty"`
	StartTime   string `json:"StartTime,omitempty"`
	EndTime     string `json:"EndTime,omitempty"`
	Error       string `json:"Error,omitempty"`
	Code        int    `json:"Code,omitempty"`
	Message     string `json:"Message,omitempty"`
}

// CreateWorkflowJobInput 提交工作流 Job 的入参。JobDetail 为具体的作业体类型（如 WorkflowAudioJobDetail、WorkflowVideoJobDetail、WorkflowFileJobDetail）。
type CreateWorkflowJobInput struct {
	Bucket    string
	JobType   WorkflowJobType
	JobDetail interface{}
	GenericInput
}

type CreateWorkflowJobOutput struct {
	RequestInfo
	Code    string `json:"Code,omitempty"`
	Message string `json:"Message,omitempty"`
	JobID   string `json:"JobId,omitempty"`
}

type QueryWorkflowJobsInput struct {
	Bucket    string
	JobType   string
	JobID     string
	PageSize  int
	PageToken string
	StartTime int64
	EndTime   int64
	GenericInput
}

// QueryWorkflowJobsOutput 查询结果，Items 为 json.RawMessage 列表，由调用方按 JobType 反序列化为对应类型。
type QueryWorkflowJobsOutput struct {
	RequestInfo
	JobType       string            `json:"JobType,omitempty"`
	Items         []json.RawMessage `json:"Items,omitempty"`
	NextPageToken string            `json:"NextPageToken,omitempty"`
}

type WorkflowJobType string

const (
	WorkflowJobTypeAudioConvert   WorkflowJobType = "AudioConvert"
	WorkflowJobTypeAudioConcat    WorkflowJobType = "AudioConcat"
	WorkflowJobTypeTranscode      WorkflowJobType = "Transcode"
	WorkflowJobTypeVideoSnapshots WorkflowJobType = "VideoSnapshots"
	WorkflowJobTypeRemux          WorkflowJobType = "Remux"
	WorkflowJobTypeConcat         WorkflowJobType = "Concat"
	WorkflowJobTypeDocConvert     WorkflowJobType = "DocConvert"
	WorkflowJobTypeFileCompress   WorkflowJobType = "FileCompress"
	WorkflowJobTypeFileUncompress WorkflowJobType = "FileUncompress"
)

type WorkflowDocType string

const (
	WorkflowDocTypePDF  WorkflowDocType = "pdf"
	WorkflowDocTypeJPG  WorkflowDocType = "jpg"
	WorkflowDocTypePNG  WorkflowDocType = "png"
	WorkflowDocTypeHTML WorkflowDocType = "html"
)

type WorkflowBucketJobsInfo struct {
	Bucket string `json:"Bucket,omitempty"`
}

type WorkflowDocConvertProcessInput struct {
	Input            WorkflowDocConvertInput  `json:"Input"`
	DocConvertConfig WorkflowDocConvertConfig `json:"DocProcessConfig"`
	Output           WorkflowDocConvertOutput `json:"Output"`
}

type WorkflowDocConvertInput struct {
	Key string `json:"Key,omitempty"`
}

type WorkflowDocConvertConfig struct {
	SrcType   WorkflowDocType `json:"SrcType,omitempty"`
	TgtType   WorkflowDocType `json:"TgtType,omitempty"`
	StartPage *int            `json:"StartPage,omitempty"`
	EndPage   *int            `json:"EndPage,omitempty"`
}

type WorkflowDocConvertOutput struct {
	Region string `json:"Region,omitempty"`
	Bucket string `json:"Bucket,omitempty"`
	Object string `json:"Object,omitempty"`
}

type CreateDocConvertToPDFJobParams struct {
	Bucket       string
	Region       string
	SourceKey    string
	SourceType   WorkflowDocType
	OutputBucket string
	OutputObject string
}

type CreateDocConvertToImageJobParams struct {
	Bucket       string
	Region       string
	SourceKey    string
	SourceType   WorkflowDocType
	OutputBucket string
	OutputObject string
	StartPage    *int
	EndPage      *int
}

type WorkflowAudioJobDetail struct {
	Input              ProcessJobInput     `json:"Input"`
	TemplateID         string              `json:"TemplateID,omitempty"`
	AudioConvertConfig *AudioConvertParams `json:"AudioConvertConfig,omitempty"`
	AudioConcatConfig  *AudioConcatParams  `json:"AudioConcatConfig,omitempty"`
	Output             ProcessJobOutput    `json:"Output"`
	Tag                string              `json:"Tag"`
}

type WorkflowVideoJobDetail struct {
	Input           ProcessJobInput               `json:"Input"`
	TranscodeConfig *WorkflowVideoTranscodeConfig `json:"TranscodeConfig,omitempty"`
	RemuxConfig     *WorkflowVideoRemuxConfig     `json:"RemuxConfig,omitempty"`
	Output          ProcessJobOutput              `json:"Output"`
	Tag             string                        `json:"Tag"`
}

type WorkflowVideoTranscodeConfig struct {
	TemplateID          string                 `json:"TemplateID,omitempty"`
	Transcode           *VideoTranscodeDetail  `json:"Transcode,omitempty"`
	WatermarkTemplateID []string               `json:"WatermarkTemplateID,omitempty"`
	Watermark           []VideoWatermark       `json:"Watermark,omitempty"`
	DigitalWatermark    *VideoDigitalWatermark `json:"DigitalWatermark,omitempty"`
}

type WorkflowVideoRemuxConfig struct {
	Format         string        `json:"Format"`
	Duration       int           `json:"Duration,omitempty"`
	TranscodeIndex string        `json:"TranscodeIndex,omitempty"`
	AIGCMetadata   *AIGCMetadata `json:"AIGCMetadata,omitempty"`
}

type WorkflowFileJobDetail struct {
	Input                WorkflowFileJobInput           `json:"Input"`
	FileCompressConfig   *WorkflowFileCompressConfig   `json:"FileCompressConfig,omitempty"`
	FileUncompressConfig *WorkflowFileUncompressConfig `json:"FileUncompressConfig,omitempty"`
	Output               WorkflowFileJobOutput         `json:"Output"`
}

// WorkflowFileJobInput 文件 Job 的输入，通过 CompressInput/UncompressInput 二选一。
// 自定义 MarshalJSON 保证序列化时只输出实际设置的字段。
type WorkflowFileJobInput struct {
	CompressInput   *WorkflowFileCompressInput   `json:"-"`
	UncompressInput *WorkflowFileUncompressInput `json:"-"`
}

func (i WorkflowFileJobInput) MarshalJSON() ([]byte, error) {
	if i.CompressInput != nil {
		return json.Marshal(i.CompressInput)
	}
	return json.Marshal(i.UncompressInput)
}

type WorkflowFileCompressConfig struct {
	Format  string `json:"Format"`
	Flatten int    `json:"Flatten,omitempty"`
}

type WorkflowFileUncompressConfig struct {
	Prefix         string `json:"Prefix"`
	PrefixReplaced int    `json:"PrefixReplaced,omitempty"`
}

type WorkflowFileCompressInput struct {
	Prefix    string                          `json:"Prefix,omitempty"`
	KeyConfig []WorkflowFileCompressKeyConfig `json:"KeyConfig,omitempty"`
}

type WorkflowFileCompressKeyConfig struct {
	Key string `json:"Key"`
}

type WorkflowFileUncompressInput struct {
	Object string `json:"Object"`
}

type WorkflowFileJobOutput struct {
	Region string `json:"Region"`
	Bucket string `json:"Bucket"`
	Object string `json:"Object,omitempty"`
}

type CreateAudioConvertWorkflowJobParams struct {
	Bucket        string
	SourceKey     string
	Region        string
	OutputBucket  string
	OutputObject  string
	TemplateID    string
	ConvertParams *AudioConvertParams
}

type CreateAudioConcatWorkflowJobParams struct {
	Bucket       string
	SourceKey    string
	Region       string
	OutputBucket string
	OutputObject string
	ConcatParams *AudioConcatParams
}

type CreateVideoTranscodeWorkflowJobParams struct {
	Bucket              string
	SourceKey           string
	Region              string
	OutputBucket        string
	OutputObject        string
	TemplateID          string
	WatermarkTemplateID []string
	TranscodeConfig     *VideoTranscodeConfig
}

type CreateVideoRemuxWorkflowJobParams struct {
	Bucket       string
	SourceKey    string
	Region       string
	OutputBucket string
	OutputObject string
	RemuxConfig  *WorkflowVideoRemuxConfig
}

type CreateFileCompressWorkflowJobParams struct {
	Bucket       string
	Region       string
	OutputBucket string
	OutputObject string
	SourceKeys   []string
	SourcePrefix string
	Format       string
	Flatten      int
}

type CreateFileUncompressWorkflowJobParams struct {
	Bucket         string
	Region         string
	OutputBucket   string
	SourceKey      string
	Prefix         string
	PrefixReplaced int
}

type ListProcessTemplateInput struct {
	Bucket   string
	Tag      string
	Category string
	GenericInput
}

type ListProcessTemplateOutput struct {
	RequestInfo
	Templates []json.RawMessage `json:"Templates,omitempty"`
}

type GetProcessTemplateInput struct {
	Bucket   string
	Tag      string
	ID       string
	Category string
	GenericInput
}

type GetProcessTemplateOutput struct {
	RequestInfo
	Template json.RawMessage `json:"Template,omitempty"`
}

type PutProcessTemplateInput struct {
	Bucket         string
	Tag            string
	TemplateConfig interface{}
	GenericInput
}

type PutProcessTemplateOutput struct {
	RequestInfo
	TemplateID string `json:"TemplateID,omitempty"`
}

type DeleteProcessTemplateInput struct {
	Bucket string
	ID     string
	GenericInput
}

type DeleteProcessTemplateOutput struct {
	RequestInfo
}
