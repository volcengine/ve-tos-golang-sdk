package tos

import (
	"encoding/json"
	"io"
	"strconv"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// ==================== 图片处理参数 ====================

// ImageProcessParams 描述单个图片操作。多个 ImageProcessParams 可串联，对应 x-tos-process 中 image/ 下的多操作组合。
type ImageProcessParams struct {
	Operation enum.ImageOperation

	ResizeParams           *ImageResizeParams
	WatermarkParams        *ImageWatermarkParams
	BlindWatermarkParams   *ImageBlindWatermarkParams
	DeBlindWatermarkParams *ImageDeBlindWatermarkParams
	AIGCMetadataParams     *ImageAIGCMetadataParams
	GetAIGCMetadataParams  *ImageGetAIGCMetadataParams
	SetC2PAMetadataParams  *ImageSetC2PAMetadataParams
	GetC2PAMetadataParams  *ImageGetC2PAMetadataParams
	AiTagParams            *ImageAiTagParams
	EmbeddingParams        *ImageEmbeddingParams
	UnderstandingParams    *ImageUnderstandingParams
	OCRParams              *ImageOCRParams

	CropParams      *ImageCropParams
	CircleParams    *ImageCircleParams
	IndexcropParams *ImageIndexcropParams
	ClipParams      *ImageClipParams

	RoundedCorners           *ImageRoundedCornersParams
	AutoOrientParams         *ImageAutoOrientParams
	AutoOrientInternalParams *ImageAutoOrientInternalParams
	BlurParams               *ImageBlurParams
	RotateParams             *ImageRotateParams
	BrightParams             *ImageBrightParams
	SharpenParams            *ImageSharpenParams
	ContrastParams           *ImageContrastParams
	GrayscaleParams          *ImageGrayscaleParams
	ColorspaceParams         *ImageColorspaceParams

	DrawParams      *ImageDrawParams
	MosaicParams    *ImageMosaicParams
	QualityParams   *ImageQualityParams
	InterlaceParams *ImageInterlaceParams
	FormatParams    *ImageFormatParams
	SlimParams      *ImageSlimParams
}

type ImageResizeParams struct {
	M     string // lfit, mfit, fill, pad, fixed
	W     *int   // [1,16384]
	H     *int   // [1,16384]
	L     *int   // [1,16384]
	S     *int   // [1,16384]
	P     *int   // [1,1000]
	Limit *int   // 0 或 1
	Color string // 十六进制颜色码
}

type ImageWatermarkParams struct {
	// 基础参数
	T       *int   // 透明度 [0,100]
	G       string // 位置
	X       *int   // 水平边距 [0,4096]
	Y       *int   // 垂直边距 [0,4096]
	VOffset *int   // 中线偏移 [-1000,1000]

	// 图片水印
	Image  string // URL-safe Base64 编码的水印图对象名
	ImageP *int   // 水印图相对原图比例

	// 文字水印
	Text   string // URL-safe Base64 编码的水印文字
	Type   string // URL-safe Base64 编码的字体名
	Color  string // 十六进制颜色码
	Size   *int   // 文字大小 (0,1000]
	Shadow *int   // 阴影透明度 [0,100]
	Rotate *int   // 旋转角度 [0,360]
	Fill   *int   // 0 或 1

	// 图文混排
	Order    *int // 0 或 1
	Align    *int // 0、1、2
	Interval *int // [0,1000]
}

type ImageBlindWatermarkParams struct {
	Text    string // URL-safe Base64
	Version *int   // 1 或 2（建议 2）
	Level   *int   // 1-3（建议 2）
}

type ImageDeBlindWatermarkParams struct {
	Version *int // 1 或 2，默认 1
}

type ImageAIGCMetadataParams struct {
	Label             string // URL-safe Base64
	ContentProducer   string // URL-safe Base64
	ProduceID         string // URL-safe Base64
	ContentPropagator string // URL-safe Base64
	PropagateID       string // URL-safe Base64
	ReservedCode1     string // URL-safe Base64
	ReservedCode2     string // URL-safe Base64
}

type ImageGetAIGCMetadataParams struct{}

type ImageSetC2PAMetadataParams struct {
	AppID    string // 原样传递
	Manifest string // URL-safe Base64
}

type ImageGetC2PAMetadataParams struct{}

type ImageAiTagParams struct{}

type ImageEmbeddingParams struct {
	Embedder      string  // embedding 模型名称
	Dimension     *int    // 降维维度
	Instructions  *string // URL-safe Base64
	EmbeddingType *int    // 0-3
}

type ImageUnderstandingParams struct {
	Model  string
	Prompt string
	Detail string
}

type ImageOCRParams struct {
	Model string
}

type ImageCropParams struct {
	G string
	W *int
	H *int
	X *int
	Y *int
}

type ImageCircleParams struct {
	R int
}

type ImageIndexcropParams struct {
	X *int
	Y *int
	I *int
}

type ImageClipParams struct {
	Frame *int
	First *int
	Step  *int
}

type ImageRoundedCornersParams struct {
	R int
}

type ImageAutoOrientParams struct {
	Value int // 0 或 1
}

type ImageAutoOrientInternalParams struct{}

type ImageBlurParams struct {
	Value  int
	Radius *int
	Sigma  *int
}

type ImageRotateParams struct {
	Value int
}

type ImageBrightParams struct {
	Value int // [-100,100]
}

type ImageSharpenParams struct {
	Value int
}

type ImageContrastParams struct {
	Value int
}

type ImageGrayscaleParams struct {
	Value int // 0 或 1
}

type ImageColorspaceParams struct {
	Value string // gray
}

type ImagePoint struct {
	X int
	Y int
}

type ImageDrawParams struct {
	Points    []ImagePoint // p，必填，序列化为 "50x50-100x100"
	Radius    *int         // r，默认 6
	DrawLine  *bool        // l，true/false
	LineWidth *int         // lw，默认 3
	ColorRGB  *string      // color，RRGGBB，不带 #；默认 000000
}

type ImageMosaicParams struct {
	G string // nw,north,ne,west,center,east,sw,south,se；默认 nw
	W *int   // 裁剪宽度
	H *int   // 裁剪高度
	X *int   // 横向偏移（相对 g）
	Y *int   // 纵向偏移（相对 g）
	T string // square(默认) 或 blur

	// 仅当 t=blur 生效
	R *int // [0,50]
	S *int // [1,50]，默认 20
}

type ImageQualityParams struct {
	Q  *int // 相对质量
	QQ *int // 绝对质量
}

type ImageInterlaceParams struct {
	Value int // 0 或 1
}

type ImageFormatParams struct {
	Format string // jpg, png, webp, bmp, gif, tiff, heic
}

type ImageSlimParams struct {
	ZLevel *int // 0-10，默认 3
}

// ==================== 视频处理参数 ====================

// VideoProcessParams 描述统一视频处理操作，同一时刻只能指定一个操作分支。
// Get 路径支持：Info/Snapshot/PM3U8/Embedding/Understanding/AIGCMetadata/C2PAMetadata。
// Put 路径支持 Transcode；Post 同步路径额外支持 Snapshots/PCM/Convert/Remux。
type VideoProcessParams struct {
	InfoParams          *VideoInfoParams
	SnapshotParams      *VideoSnapshotParams
	PM3U8Params         *VideoPM3U8Params
	EmbeddingParams     *VideoEmbeddingParams
	UnderstandingParams *VideoUnderstandingParams
	AIGCMetadataParams  *VideoAIGCMetadataParams
	C2PAMetadataParams  *VideoC2PAMetadataParams
	SnapshotsParams     *VideoSnapshotsParams
	TranscodeParams     *VideoTranscodeParams
	PCMParams           *VideoPCMParams
	ConvertParams       *VideoConvertParams
	RemuxParams         *VideoRemuxParams
}

type VideoInfoParams struct{}

type VideoSnapshotParams struct {
	T      int    // 截帧时间，毫秒
	W      *int   // 宽度
	H      *int   // 高度
	M      string // accurate 或 fast
	F      string // jpg 或 png
	Rotate *int   // 旋转角度
}

type VideoPM3U8Params struct{}

type VideoEmbeddingParams struct {
	Embedder     string
	Dimension    *int
	Instructions *string // URL-safe Base64
}

type VideoUnderstandingParams struct {
	Model  string // URL-safe Base64
	Prompt string // URL-safe Base64
	FPS    *float64
}

type VideoAIGCMetadataParams struct{}

type VideoC2PAMetadataParams struct{}

type VideoPCMParams struct {
	SampleRate *int
	Channels   *int
}

type VideoSnapshotsParams struct {
	Format         string
	Mode           string
	Width          *int
	Height         *int
	Index          string
	Num            *int
	DHashThreshold *int
}

// VideoTranscodeParams 描述 Put 视频处理及旧版同步 Post 使用的 JSON 转码参数。
// 新版同步 Post 请使用 VideoConvertParams，异步 Post 请使用 TranscodeJobBody。
type VideoTranscodeParams struct {
	Tag             string               `json:"Tag"`
	Name            string               `json:"Name"`
	TranscodeConfig VideoTranscodeConfig `json:"TranscodeConfig"`
	Output          *ProcessJobOutput    `json:"Output,omitempty"`
}

// VideoConvertParams 描述 Post 视频转码（video/convert）参数。
// 参数最终会按 Post 视频接口要求拼接为逗号分隔的处理流水线，而不是 JSON。
type VideoConvertParams struct {
	Format             string
	StartTime          *int // ss，截取开始时间，毫秒
	Duration           *int // t，截取时长，毫秒
	HLSSegmentDuration *int // st，HLS 分片时长，毫秒
	RemoveVideo        *int // vn，0 保留视频流，1 移除视频流
	VideoCodec         string
	FPS                *int
	PixelFormat        string
	Width              *int
	Height             *int
	VideoBitRate       *int
	MaxRate            *int
	BufferSize         *int
	CRF                *int
	RemoveAudio        *int // an，0 保留音频流，1 移除音频流
	AudioCodec         string
	SampleRate         *int
	Channels           *int
	AudioBitRate       *int
	SampleFormat       string
	AIGCMetadata       *AIGCMetadata
	C2PAMetadata       *C2PAMetadata
	Watermarks         []VideoWatermark
	BlindWatermark     *VideoDigitalWatermark
}

// VideoRemuxParams 描述 Post 视频转封装（video/remux）参数。
// video/remux 不重新编码，也不支持可见水印和暗水印。
type VideoRemuxParams struct {
	Format             string
	HLSSegmentDuration *int // st，HLS 分片时长，毫秒
	StreamIndex        *int // ti，保留的输入流索引
	AIGCMetadata       *AIGCMetadata
	C2PAMetadata       *C2PAMetadata
}

type VideoTranscodeConfig struct {
	Transcode        *VideoTranscodeDetail  `json:"Transcode,omitempty"`
	Watermark        []VideoWatermark       `json:"Watermark,omitempty"`
	DigitalWatermark *VideoDigitalWatermark `json:"DigitalWatermark,omitempty"`
}

type VideoTranscodeDetail struct {
	TimeInterval *TimeInterval     `json:"TimeInterval,omitempty"`
	Container    *Container        `json:"Container,omitempty"`
	Video        *VideoConfig      `json:"Video,omitempty"`
	Audio        *AudioConfig      `json:"Audio,omitempty"`
	Options      *TranscodeOptions `json:"Options,omitempty"`
}

type TimeInterval struct {
	Start    int `json:"Start"`
	Duration int `json:"Duration"`
}

type Container struct {
	Format     string      `json:"Format"`
	ClipConfig *ClipConfig `json:"ClipConfig,omitempty"`
}

type ClipConfig struct {
	Duration *int `json:"Duration,omitempty"`
}

type VideoConfig struct {
	Codec   string `json:"Codec,omitempty"`
	Width   *int   `json:"Width,omitempty"`
	Height  *int   `json:"Height,omitempty"`
	Crf     *int   `json:"Crf,omitempty"`
	PixFmt  string `json:"PixFmt,omitempty"`
	BitRate *int   `json:"BitRate,omitempty"`
	Fps     *int   `json:"Fps,omitempty"`
	Remove  *bool  `json:"Remove,omitempty"`
}

type AudioConfig struct {
	Codec        string `json:"Codec,omitempty"`
	SampleRate   *int   `json:"SampleRate,omitempty"`
	BitRate      *int   `json:"BitRate,omitempty"`
	SampleFormat string `json:"SampleFormat,omitempty"`
	Channels     *int   `json:"Channels,omitempty"`
	Remove       *bool  `json:"Remove,omitempty"`
}

type TranscodeOptions struct {
	AIGCMetadata *AIGCMetadata `json:"AIGCMetadata,omitempty"`
}

type AIGCMetadata struct {
	Label             string `json:"Label,omitempty"`
	ContentProducer   string `json:"ContentProducer,omitempty"`
	ProduceID         string `json:"ProduceID,omitempty"`
	ContentPropagator string `json:"ContentPropagator,omitempty"`
	PropagateID       string `json:"PropagateID,omitempty"`
	ReservedCode1     string `json:"ReservedCode1,omitempty"`
	ReservedCode2     string `json:"ReservedCode2,omitempty"`
}

// C2PAMetadata 描述 Post 视频接口使用的 C2PA 元数据。
// Manifest 使用 interface{} 保留 C2PA Manifest 的扩展性，调用方可传入结构体或 map。
type C2PAMetadata struct {
	AppID    string      `json:"AppID,omitempty"`
	Manifest interface{} `json:"Manifest,omitempty"`
}

type VideoWatermark struct {
	Type      string `json:"Type,omitempty"`
	Pos       string `json:"Pos,omitempty"`
	LocMode   string `json:"LocMode,omitempty"`
	Dx        *int   `json:"Dx,omitempty"`
	Dy        *int   `json:"Dy,omitempty"`
	StartTime *int   `json:"StartTime,omitempty"`
	EndTime   *int   `json:"EndTime,omitempty"`

	Text  *VideoWatermarkText  `json:"Text,omitempty"`
	Image *VideoWatermarkImage `json:"Image,omitempty"`
}

type VideoWatermarkText struct {
	FontSize     *int   `json:"FontSize,omitempty"`
	FontType     string `json:"FontType,omitempty"`
	FontColor    string `json:"FontColor,omitempty"`
	Transparency *int   `json:"Transparency,omitempty"`
	Text         string `json:"Text,omitempty"`
}

type VideoWatermarkImage struct {
	Url          string `json:"Url,omitempty"`
	Mode         string `json:"Mode,omitempty"`
	Width        *int   `json:"Width,omitempty"`
	Height       *int   `json:"Height,omitempty"`
	Transparency *int   `json:"Transparency,omitempty"`
	Background   *bool  `json:"Background,omitempty"`
}

type VideoDigitalWatermark struct {
	Type          string `json:"Type,omitempty"`
	Version       string `json:"Version,omitempty"`
	Message       string `json:"Message,omitempty"`
	FrameInterval *int   `json:"FrameInterval,omitempty"`
}

// ==================== 文档处理参数 ====================

type DocProcessParams struct {
	SrcType   enum.DocPreviewSrcType
	DstType   enum.DocPreviewDstType
	DocPage   *int
	StartPage *int
	EndPage   *int
	ImageMode *enum.ImageModeType
}

// ==================== HLS 处理参数 ====================

type HlsProcessParams struct {
	M3U8Params *HlsM3U8Params
	TSParams   *HlsTSParams
}

type HlsM3U8Params struct {
	SegmentDuration      *int
	Width                int
	Height               int
	EncodeFormat         string
	PixFmt               string
	WatermarkTemplateIDs []string
}

type HlsTSParams struct {
	FromObject           string
	Width                int
	Height               int
	EncodeFormat         string
	PixFmt               string
	StartTime            int64
	EndTime              int64
	NeedDownload         *bool
	WatermarkTemplateIDs []string
}

// ==================== MCAP 文件处理参数 ====================

type McapOperation string

const (
	McapOperationInfo   McapOperation = "mcap-info"
	McapOperationDoctor McapOperation = "mcap-doctor"
	McapOperationList   McapOperation = "mcap-list"
)

type McapListResource string

const (
	McapListAttachments McapListResource = "attachments"
	McapListChannels    McapListResource = "channels"
	McapListChunks      McapListResource = "chunks"
	McapListMetadata    McapListResource = "metadata"
	McapListSchemas     McapListResource = "schemas"
)

type McapProcessParams struct {
	Operation    McapOperation
	ListResource McapListResource // 当 Operation = McapOperationList 时必填
}

// MCAP 响应类型

type McapInfoChunk struct {
	Num              int     `json:"num"`
	TotalSize        uint64  `json:"total_size"`
	CompressionRatio float64 `json:"compression_ratio"`
}

type McapInfoResult struct {
	Library       string        `json:"library"`
	Profile       string        `json:"profile"`
	MessageNum    uint64        `json:"message_num"`
	Duration      string        `json:"duration"`
	Start         string        `json:"start"`
	End           string        `json:"end"`
	Compression   []string      `json:"compression"`
	Chunks        McapInfoChunk `json:"chunks"`
	Channels      []string      `json:"channels"`
	ChannelNum    int           `json:"channel_num"`
	AttachmentNum int           `json:"attachment_num"`
	MetadataNum   int           `json:"metadata_num"`
}

type McapDoctorResult struct {
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
	Fatals   []string `json:"fatals"`
}

type McapAttachment struct {
	Name          string `json:"name"`
	MediaType     string `json:"media_type"`
	LogTime       uint64 `json:"log_time"`
	CreationTime  uint64 `json:"creation_time"`
	ContentLength uint64 `json:"content_length"`
	Offset        uint64 `json:"offset"`
}

type McapAttachmentsResult struct {
	Attachments []McapAttachment `json:"attachments"`
}

type McapChannel struct {
	Id              uint16 `json:"id"`
	SchemaId        uint16 `json:"schema_id"`
	Topic           string `json:"topic"`
	MessageEncoding string `json:"message_encoding"`
	Metadata        string `json:"metadata"`
}

type McapChannelsResult struct {
	Channels []McapChannel `json:"channels"`
}

type McapChunk struct {
	Offset             uint64  `json:"offset"`
	Length             uint64  `json:"length"`
	Start              uint64  `json:"start"`
	End                uint64  `json:"end"`
	Compression        string  `json:"compression"`
	CompressedSize     uint64  `json:"compressed_size"`
	UncompressedSize   uint64  `json:"uncompressed_size"`
	CompressionRatio   float64 `json:"compression_ratio"`
	MessageIndexLength uint64  `json:"message_index_length"`
}

type McapChunksResult struct {
	Chunks []McapChunk `json:"chunks"`
}

type McapMetaData struct {
	Name     string `json:"name"`
	Offset   uint64 `json:"offset"`
	Length   uint64 `json:"length"`
	MetaData string `json:"metadata"`
}

type McapMetaDatasResult struct {
	MetaDatas []McapMetaData `json:"metadatas"`
}

type McapSchema struct {
	Id       uint16 `json:"id"`
	Name     string `json:"name"`
	Encoding string `json:"encoding"`
	Data     string `json:"data"`
}

type McapSchemasResult struct {
	Schemas []McapSchema `json:"schemas"`
}

// ==================== 点云压缩参数 ====================

type PointCloudCompressParams struct {
	Format           string   // 文件格式，默认 "pcd"，当前仅支持 pcd
	Method           string   // 压缩方式，默认 "octree"，当前仅支持 octree
	Fields           string   // 点云数据类型，默认 "eHI6"(即 xyz 的 base64)
	Lib              string   // 压缩库，默认 "pcl"
	PointResolution  *float64 // 点云分辨率，取值 [0,1]，默认 0.01
	OctreeResolution *float64 // 八叉树最小块(voxel)边长，取值 [0,1]，默认 0.01
	DownSampling     *bool    // 是否使用下采样，默认 true
}

// QueryParams 返回点云压缩的额外 query 参数，用于设置到 GenericInput.RequestQuery
func (p *PointCloudCompressParams) QueryParams() map[string]string {
	q := make(map[string]string)

	format := p.Format
	if format == "" {
		format = "pcd"
	}
	q["format"] = format

	method := p.Method
	if method == "" {
		method = "octree"
	}
	q["method"] = method

	fields := p.Fields
	if fields == "" {
		fields = "eHI6"
	}
	q["fields"] = fields

	lib := p.Lib
	if lib == "" {
		lib = "pcl"
	}
	q["lib"] = lib

	pr := 0.01
	if p.PointResolution != nil {
		pr = *p.PointResolution
	}
	q["point-resolution"] = strconv.FormatFloat(pr, 'f', -1, 64)

	or := 0.01
	if p.OctreeResolution != nil {
		or = *p.OctreeResolution
	}
	q["octree-resolution"] = strconv.FormatFloat(or, 'f', -1, 64)

	ds := 1
	if p.DownSampling != nil && !*p.DownSampling {
		ds = 0
	}
	q["down-sampling"] = strconv.Itoa(ds)

	return q
}

// ==================== 另存为参数 ====================

// SaveAsParams 描述另存为目标位置。Get 路径通过 x-tos-save-* query 传递；
// Post 图片及视频 convert/remux 通过请求体中的 x-tos-save-* 参数传递，旧协议仍使用 sys/saveas。
type SaveAsParams struct {
	SaveBucket string // 传原始字符串，SDK 内部编码
	SaveObject string // 传原始字符串，SDK 内部编码
}

// ==================== Get 接口相关类型 ====================

// GetDataProcessParams 是 GetDataProcessHelper 的入参，按 GetProcessType 选择对应的处理参数分支。
// SaveAs 由调用方直接设到 GetObjectV2Input.SaveBucket/SaveObject，SDK 内部自动处理兼容性（不支持的算子会静默忽略）。
type GetDataProcessParams struct {
	GetProcessType          enum.GetProcessType
	ImageProcessParams      []ImageProcessParams // 数组，支持多操作串联
	VideoProcessParams      *VideoProcessParams
	DocProcessParams        *DocProcessParams
	HlsProcessParams        *HlsProcessParams
	McapProcessParams       *McapProcessParams
	PointCloudProcessParams *PointCloudCompressParams
}

// ==================== Put 接口相关类型 ====================

// PutDataProcessParams 是 PutDataProcessHelper 的入参，按 PutProcessType 选择对应的处理参数分支。
type PutDataProcessParams struct {
	PutProcessType enum.PutProcessType

	// 图片上传处理：对应 header x-tos-image-operations
	ImageOperations *PutImageOperationsParams

	// 兼容/预留：如后端仍支持 X-Tos-Process
	ImageProcessParams []ImageProcessParams

	// 视频上传处理：对应 header x-tos-video-operations（base64(json)）
	VideoProcessParams *VideoProcessParams

	DocProcessParams *DocProcessParams
}

type PutImageOperationsParams struct {
	IsImageInfo *int                     `json:"is_image_info,omitempty"`
	Rules       []PutImageOperationsRule `json:"rules,omitempty"`
}

type PutImageOperationsRule struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Rule   string `json:"rule"`
}

// ==================== Post 同步接口类型 ====================

// PostDPInput 是 PostDataProcess 的入参。PostProcess 支持 Post 处理流水线字符串，
// 同时兼容旧版 JSON body。
type PostDPInput struct {
	GenericInput
	Bucket      string
	Key         string
	PostProcess string
}

// PostDPOutput 按请求类型填充图片或视频响应；同一次请求最多只有一个分支非 nil。
type PostDPOutput struct {
	RequestInfo
	ImageProcessOutput *ImageProcessOutput
	VideoProcessOutput *VideoProcessOutput
}

// ImageProcessOutput 描述 Post 图片处理响应。SaveAs 响应填充对象字段；
// 非 SaveAs 响应无论 Content-Type 为何，均由调用方读取并关闭 Content。
type ImageProcessOutput struct {
	Bucket   string `json:"bucket,omitempty"`
	Object   string `json:"object,omitempty"`
	FileSize int64  `json:"fileSize,string,omitempty"`
	Status   string `json:"status,omitempty"`

	Content       io.ReadCloser `json:"-"`
	ContentType   string        `json:"-"`
	ContentLength int64         `json:"-"`
}

func (o *PostDPOutput) UnmarshalJSON(data []byte) error {
	var video VideoProcessOutput
	if err := json.Unmarshal(data, &video); err != nil {
		return err
	}
	o.ImageProcessOutput = nil
	o.VideoProcessOutput = nil
	if video.OutputBucket != "" ||
		video.TotalFrameCount != 0 ||
		video.SuccFrameCount != 0 ||
		video.FailFrameCount != 0 ||
		len(video.SuccFrameList) > 0 ||
		len(video.FailFrameList) > 0 ||
		video.PcmBucket != "" ||
		video.PcmObject != "" ||
		video.PcmStatus != "" {
		o.VideoProcessOutput = &video
	}
	return nil
}

// PostDataProcessParams 是 PostDataProcessHelper 的入参，按 PostProcessType 选择对应的处理参数分支。
type PostDataProcessParams struct {
	PostProcessType    enum.PostProcessType
	ImageProcessParams []ImageProcessParams
	VideoProcessParams *VideoProcessParams
	DocProcessParams   *DocProcessParams
	SaveAsParams       *SaveAsParams
}

type VideoProcessOutput struct {
	OutputBucket    string
	TotalFrameCount int
	SuccFrameCount  int
	FailFrameCount  int
	SuccFrameList   []SuccFrame
	FailFrameList   []FailFrame
	PcmDataProcessOutput

	// Convert/Remux SaveAs 与 PCM 使用相同的响应字段，由请求类型决定映射位置。
	SaveAsBucket     string
	SaveAsObject     string
	SaveAsObjectSize int64
	SaveAsStatus     enum.VideoDataProcessStatus
}

// ==================== 音频异步接口类型 ====================

// AudioConvertParams 描述工作流及旧版 x-tos-async-process 路径中的音频转码参数。
// 新版 Post 异步请求请使用 AudioConvertJobConfig。
type AudioConvertParams struct {
	ContainerFormat string
	TimeInterval    *TimeInterval
	BitRate         *int
	BitRateOpt      *int
	SampleRate      *int
	Channels        *int
	SampleFormat    string
}

type AudioConcatFragment struct {
	Object   string
	Start    *int
	Duration *int
}

// AudioConcatParams 描述音频拼接参数。PreFragments 为前拼接片段，SurFragments 为后拼接片段。
type AudioConcatParams struct {
	ContainerFormat string
	TimeInterval    *TimeInterval
	BitRate         *int
	BitRateOpt      *int
	SampleRate      *int
	Channels        *int
	SampleFormat    string
	PreFragments    []AudioConcatFragment
	SurFragments    []AudioConcatFragment
}

// ==================== Post 异步接口类型 ====================

// PostDPAsyncInput 是 PostDataProcessAsync 的入参。新调用应通过 JobType + JobBody 提交 JSON Job。
type PostDPAsyncInput struct {
	GenericInput
	Bucket string
	// Deprecated: 仅用于兼容旧版 x-tos-async-process query-string 请求。
	Key string
	// Deprecated: 仅用于兼容旧版 x-tos-async-process query-string 请求。
	PostProcess string
	JobType     ProcessJobType
	// JobBody 可传结构化 Job 类型，或 PostDataProcessAsyncHelper 返回的 JSON 字符串。
	JobBody interface{}
}

type PostDPAsyncOutput struct {
	RequestInfo
	Code    string `json:"Code"`
	Message string `json:"Message"`
	JobId   string `json:"JobId"`
	Status  string `json:"status,omitempty"` // Deprecated: async-process 路径已改为走 CommitJob，不再返回 status 字段，请使用 Code 判断成功
}

// PostDataProcessAsyncParams 是 PostDataProcessAsyncHelper 的入参，JobBody 将被序列化为 JSON。
type PostDataProcessAsyncParams struct {
	// Deprecated: 请使用 JobType + JobBody。
	PostProcessAsyncType enum.PostProcessAsyncType
	// Deprecated: 请使用 AudioConvertJobBody 或 AudioConcatJobBody。
	AudioProcessAsyncParams *AudioProcessAsyncParams
	JobType                 ProcessJobType
	JobBody                 interface{}
}

type DPAsyncStatus string

const (
	DPAsyncStatusQueued  DPAsyncStatus = "queued"
	DPAsyncStatusRunning DPAsyncStatus = "running"
	DPAsyncStatusSuccess DPAsyncStatus = "success"
	DPAsyncStatusFailed  DPAsyncStatus = "failed"
)

type DPAsyncResult struct {
	Status   DPAsyncStatus   `json:"Status"`
	Progress *int            `json:"Progress,omitempty"`
	Raw      json.RawMessage `json:"Raw,omitempty"`
	Code     *string         `json:"Code,omitempty"`
	Message  *string         `json:"Message,omitempty"`
}

// AudioProcessAsyncParams 描述旧版 x-tos-async-process 音频参数。
// Deprecated: 请使用 AudioConvertJobBody 或 AudioConcatJobBody。
type AudioProcessAsyncParams struct {
	Bucket        string
	Region        string
	TargetObject  string
	SaveAsParams  *SaveAsParams
	ConvertParams *AudioConvertParams
	ConcatParams  *AudioConcatParams
}

type GetDPAsyncResultInput struct {
	GenericInput
	PostProcessAsyncType enum.PostProcessAsyncType
	JobId                string
	Bucket               string
	JobType              ProcessJobType
}

type GetDPAsyncResultOutput struct {
	RequestInfo
	ImageProcessAsyncOutput *ImageProcessAsyncOutput
	VideoProcessAsyncOutput *VideoProcessAsyncOutput
	DocProcessAsyncOutput   *DocProcessAsyncOutput
	AudioProcessAsyncOutput *AudioProcessAsyncOutput
	JobResult               *ProcessJobResult `json:"JobResult,omitempty"`
}

type ImageProcessAsyncOutput struct {
	Result *DPAsyncResult
}

type VideoProcessAsyncOutput struct {
	Result *DPAsyncResult
}

type DocProcessAsyncOutput struct {
	Result *DPAsyncResult
}

type AudioProcessAsyncOutput struct {
	Result *DPAsyncResult
}

// ==================== 统一 Job 类型 ====================

// ProcessJobType 统一的异步 Job 类型，同时用于 PostDataProcessAsync 和 CreateWorkflowJob。
type ProcessJobType string

const (
	ProcessJobTypeTranscode      ProcessJobType = "Transcode"
	ProcessJobTypeRemux          ProcessJobType = "Remux"
	ProcessJobTypeAnimation      ProcessJobType = "Animation"
	ProcessJobTypeAudioConvert   ProcessJobType = "AudioConvert"
	ProcessJobTypeAudioConcat    ProcessJobType = "AudioConcat"
	ProcessJobTypeDocConvert     ProcessJobType = "DocConvert"
	ProcessJobTypeDocAnalyze     ProcessJobType = "DocAnalyze"
	ProcessJobTypeFileCompress   ProcessJobType = "FileCompress"
	ProcessJobTypeFileUncompress ProcessJobType = "FileUncompress"
)

// jobsQueryParam 根据 JobType 返回对应的 query 参数名（media_jobs 或 file_jobs）。
func (t ProcessJobType) jobsQueryParam() string {
	switch t {
	case ProcessJobTypeTranscode, ProcessJobTypeRemux, ProcessJobTypeAnimation,
		ProcessJobTypeAudioConvert, ProcessJobTypeAudioConcat:
		return "media_jobs"
	default:
		return "file_jobs"
	}
}

type ProcessJobInput struct {
	Object string `json:"Object"`
}

type ProcessJobOutput struct {
	Region string `json:"Region"`
	Bucket string `json:"Bucket"`
	Object string `json:"Object"`
}

type ProcessBucketJobsInfo struct {
	BucketID      int64  `json:"BucketId,omitempty"`
	Bucket        string `json:"Bucket,omitempty"`
	AccountID     string `json:"AccountID,omitempty"`
	BucketOwnerID string `json:"BucketOwnerId,omitempty"`
}

type ProcessDocConvertInput struct {
	Key string `json:"Key,omitempty"`
}

type ProcessDocConvertConfig struct {
	SrcType    string `json:"SrcType,omitempty"`
	TgtType    string `json:"TgtType,omitempty"`
	StartPage  int    `json:"StartPage"`
	EndPage    int    `json:"EndPage"`
	ImageDpi   int    `json:"ImageDpi,omitempty"`
	Quality    int    `json:"Quality,omitempty"`
	RenderHTML bool   `json:"RenderHTML,omitempty"`
}

type ProcessDocConvertOutput struct {
	Region string `json:"Region"`
	Bucket string `json:"Bucket,omitempty"`
	Object string `json:"Object,omitempty"`
}

type DocConvertJobBody struct {
	Input            ProcessDocConvertInput  `json:"Input"`
	DocConvertConfig ProcessDocConvertConfig `json:"DocProcessConfig"`
	Output           ProcessDocConvertOutput `json:"Output"`
}

type ProcessJobResult struct {
	JobID      string           `json:"JobID"`
	CreateTime string           `json:"CreateTime"`
	StartTime  string           `json:"StartTime"`
	EndTime    string           `json:"EndTime"`
	State      string           `json:"State"`
	Code       int              `json:"Code"`
	Message    string           `json:"Message"`
	Input      ProcessJobInput  `json:"Input"`
	Output     ProcessJobOutput `json:"Output"`
}

// ==================== Job 模式请求体类型 ====================

// TranscodeJobBody 视频异步转码 Job 请求体
type TranscodeJobBody struct {
	Input           ProcessJobInput      `json:"Input"`
	Output          ProcessJobOutput     `json:"Output"`
	TranscodeConfig VideoTranscodeConfig `json:"TranscodeConfig"`
}

// RemuxConfig 视频转封装 Job 配置
type RemuxConfig struct {
	Format string `json:"Format"`
}

// RemuxJobBody 视频转封装 Job 请求体
type RemuxJobBody struct {
	Input       ProcessJobInput  `json:"Input"`
	Output      ProcessJobOutput `json:"Output"`
	RemuxConfig RemuxConfig      `json:"RemuxConfig"`
}

// AudioConvertJobConfig 音频异步转码 Job 配置
type AudioConvertJobConfig struct {
	ContainerFormat string `json:"ContainerFormat,omitempty"`
	BitRate         *int   `json:"BitRate,omitempty"`
	BitRateOpt      *int   `json:"BitRateOpt,omitempty"`
	SampleRate      *int   `json:"SampleRate,omitempty"`
	Channels        *int   `json:"Channels,omitempty"`
	SampleFormat    string `json:"SampleFormat,omitempty"`
}

// AudioConvertJobBody 音频异步转码 Job 请求体
type AudioConvertJobBody struct {
	Input              ProcessJobInput       `json:"Input"`
	Output             ProcessJobOutput      `json:"Output"`
	AudioConvertConfig AudioConvertJobConfig `json:"AudioConvertConfig"`
}

// AudioConcatInput 音频异步拼接 Job 输入。
type AudioConcatInput struct {
	Object       string                   `json:"Object"`
	PreFragments []AudioConcatPreFragment `json:"PreFragments,omitempty"`
}

type AudioConcatPreFragment struct {
	Object string `json:"Object"`
}

type AudioConcatConfig struct {
	ContainerFormat string `json:"ContainerFormat"`
}

// AudioConcatJobBody 音频异步拼接 Job 请求体。
type AudioConcatJobBody struct {
	Input             AudioConcatInput  `json:"Input"`
	Output            ProcessJobOutput  `json:"Output"`
	AudioConcatConfig AudioConcatConfig `json:"AudioConcatConfig"`
}

// ==================== Put 接口返回类型 ====================

// ==================== FileCompress / FileUncompress 类型 ====================

// FileCompressKeyConfig 指定压缩的单个文件
type FileCompressKeyConfig struct {
	Key string `json:"Key"`
}

// FileCompressInput 压缩源文件配置
type FileCompressInput struct {
	Prefix    string                  `json:"Prefix,omitempty"`    // 文件前缀（与 KeyConfig 二选一）
	KeyConfig []FileCompressKeyConfig `json:"KeyConfig,omitempty"` // 指定文件列表（最多1000个）
}

// FileCompressConfig 压缩规则
type FileCompressConfig struct {
	Format  string `json:"Format"`  // 压缩格式，目前仅支持 "zip"
	Flatten int    `json:"Flatten"` // 目录结构：0/1/2/3
}

// FileJobOutput 压缩/解压输出位置
type FileJobOutput struct {
	Region string `json:"Region"`
	Bucket string `json:"Bucket"`
	Object string `json:"Object,omitempty"` // 压缩包名称（压缩时必填，解压时不需要）
}

// FileCompressJobBody FileCompress 请求体
type FileCompressJobBody struct {
	Input              FileCompressInput  `json:"Input"`
	FileCompressConfig FileCompressConfig `json:"FileCompressConfig"`
	Output             FileJobOutput      `json:"Output"`
}

// FileUncompressInput 解压源文件配置
type FileUncompressInput struct {
	Object string `json:"Object"` // 压缩包名称
}

// FileUncompressConfig 解压规则
type FileUncompressConfig struct {
	Prefix         string `json:"Prefix"`                   // 解压后存储路径前缀
	PrefixReplaced int    `json:"PrefixReplaced,omitempty"` // 0/1/2，默认0
}

// FileUncompressJobBody FileUncompress 请求体
type FileUncompressJobBody struct {
	Input                FileUncompressInput  `json:"Input"`
	FileUncompressConfig FileUncompressConfig `json:"FileUncompressConfig"`
	Output               FileJobOutput        `json:"Output"`
}

// QueryFileJobOutput 查询文件压缩/解压任务状态的响应，复用 GetDPAsyncResultOutput.JobResult
