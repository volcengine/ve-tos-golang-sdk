package tos

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// GetDataProcessHelper 将 GetDataProcessParams 结构化参数转换为 x-tos-process 查询参数字符串。
// SaveAs 由调用方直接设到 GetObjectV2Input.SaveBucket/SaveObject，配合 GetDataProcess 使用时
// SDK 内部自动处理 base64 编码和不支持 SaveAs 的算子兼容性。
// 注意：若配合 GetObjectV2 使用，SaveAs 的 base64 编码和兼容性屏蔽不会生效。
func GetDataProcessHelper(ctx context.Context, params GetDataProcessParams) (string, error) {
	switch params.GetProcessType {
	case enum.GetProcessTypeImage:
		return buildImageProcessString(params.ImageProcessParams)
	case enum.GetProcessTypeVideo:
		if params.VideoProcessParams == nil {
			return "", fmt.Errorf("tos: VideoProcessParams is required for video process type")
		}
		if params.VideoProcessParams.ConvertParams != nil || params.VideoProcessParams.RemuxParams != nil {
			return "", fmt.Errorf("tos: video/convert and video/remux are only supported by PostDataProcess")
		}
		if params.VideoProcessParams.SnapshotsParams != nil {
			return "", fmt.Errorf("tos: video/snapshots is only supported by PostDataProcess")
		}
		if params.VideoProcessParams.PCMParams != nil {
			return "", fmt.Errorf("tos: PCM extraction is only supported by PostDataProcess")
		}
		if params.VideoProcessParams.TranscodeParams != nil {
			return "", fmt.Errorf("tos: video/transcode is only supported by PostDataProcess or async Job mode")
		}
		return buildVideoProcessString(params.VideoProcessParams)
	case enum.GetProcessTypeDoc:
		if params.DocProcessParams == nil {
			return "", fmt.Errorf("tos: DocProcessParams is required for doc process type")
		}
		return buildDocProcessString(params.DocProcessParams), nil
	case enum.GetProcessTypeHls:
		if params.HlsProcessParams == nil {
			return "", fmt.Errorf("tos: HlsProcessParams is required for hls process type")
		}
		return buildHlsProcessString(params.HlsProcessParams)
	case enum.GetProcessTypeFile:
		if params.McapProcessParams == nil {
			return "", fmt.Errorf("tos: McapProcessParams is required for file process type")
		}
		return buildMcapProcessString(params.McapProcessParams)
	case enum.GetProcessTypePointCloud:
		if params.PointCloudProcessParams == nil {
			return "", fmt.Errorf("tos: PointCloudProcessParams is required for pointcloud process type")
		}
		return "pointcloud/compress", nil
	default:
		return "", fmt.Errorf("tos: unsupported GetProcessType: %s", params.GetProcessType)
	}
}

// PutDataProcessHelper 将 PutDataProcessParams 结构化参数转换为数据处理字符串。
// 图片优先走 ImageOperations（JSON，支持 rules 多结果落盘），兼容 process string；
// 视频走 JSON 序列化；文档走 process string。
// 返回的 string 配合 PutObjectBasicInput.ProcessType + Process 使用，SDK 内部根据 ProcessType 自动填对应 header。
func PutDataProcessHelper(ctx context.Context, params PutDataProcessParams) (string, error) {
	switch params.PutProcessType {
	case enum.PutProcessTypeImage:
		if params.ImageOperations != nil {
			data, err := json.Marshal(params.ImageOperations)
			if err != nil {
				return "", fmt.Errorf("tos: failed to marshal ImageOperations: %w", err)
			}
			return string(data), nil
		}
		if len(params.ImageProcessParams) > 0 {
			return buildImageProcessString(params.ImageProcessParams)
		}
		return "", fmt.Errorf("tos: ImageOperations or ImageProcessParams is required for image put process type")
	case enum.PutProcessTypeVideo:
		if params.VideoProcessParams == nil {
			return "", fmt.Errorf("tos: VideoProcessParams.TranscodeParams is required for video put process type")
		}
		if count := countVideoProcessOperations(params.VideoProcessParams); count > 1 {
			return "", fmt.Errorf("tos: only one video operation can be specified")
		}
		if params.VideoProcessParams.ConvertParams != nil || params.VideoProcessParams.RemuxParams != nil {
			return "", fmt.Errorf("tos: video/convert and video/remux are only supported by PostDataProcess")
		}
		if params.VideoProcessParams.TranscodeParams != nil {
			data, err := json.Marshal(params.VideoProcessParams.TranscodeParams)
			if err != nil {
				return "", fmt.Errorf("tos: failed to marshal VideoTranscodeParams: %w", err)
			}
			return string(data), nil
		}
		return "", fmt.Errorf("tos: VideoProcessParams.TranscodeParams is required for video put process type")
	case enum.PutProcessTypeDoc:
		if params.DocProcessParams != nil {
			return buildDocProcessString(params.DocProcessParams), nil
		}
		return "", fmt.Errorf("tos: DocProcessParams is required for doc put process type")
	default:
		return "", fmt.Errorf("tos: unsupported PutProcessType: %s", params.PutProcessType)
	}
}

// PostDataProcessHelper 将 PostDataProcessParams 结构化参数转换为 query-string 风格的 PostProcess 字符串。
func PostDataProcessHelper(ctx context.Context, params PostDataProcessParams) (string, error) {
	switch params.PostProcessType {
	case enum.PostProcessTypeImage:
		process, err := buildImageProcessString(params.ImageProcessParams)
		if err != nil {
			return "", err
		}
		if params.SaveAsParams != nil {
			saveAs, err := buildPostImageSaveAsQuery(params.SaveAsParams)
			if err != nil {
				return "", err
			}
			process += saveAs
		}
		return process, nil
	case enum.PostProcessTypeVideo:
		if params.VideoProcessParams == nil {
			return "", fmt.Errorf("tos: VideoProcessParams is required for video post process type")
		}
		if count := countVideoProcessOperations(params.VideoProcessParams); count > 1 {
			return "", fmt.Errorf("tos: only one video operation can be specified")
		}
		if params.VideoProcessParams.ConvertParams != nil || params.VideoProcessParams.RemuxParams != nil {
			var (
				process string
				err     error
			)
			if params.VideoProcessParams.ConvertParams != nil {
				process, err = buildVideoConvertPostProcessString(params.VideoProcessParams.ConvertParams)
			} else {
				process, err = buildVideoRemuxPostProcessString(params.VideoProcessParams.RemuxParams)
			}
			if err != nil {
				return "", err
			}
			saveAs, err := buildPostVideoSaveAsQuery(params.SaveAsParams)
			if err != nil {
				return "", err
			}
			return process + saveAs, nil
		}
		if params.VideoProcessParams.InfoParams != nil ||
			params.VideoProcessParams.PM3U8Params != nil ||
			params.VideoProcessParams.EmbeddingParams != nil ||
			params.VideoProcessParams.UnderstandingParams != nil ||
			params.VideoProcessParams.AIGCMetadataParams != nil ||
			params.VideoProcessParams.C2PAMetadataParams != nil {
			return "", fmt.Errorf("tos: video info/pm3u8/embedding/understanding/metadata are only supported by GetDataProcess")
		}
		if params.VideoProcessParams.SnapshotsParams != nil {
			process, err := buildVideoSnapshotsPostProcessString(params.VideoProcessParams.SnapshotsParams)
			if err != nil {
				return "", err
			}
			if params.SaveAsParams == nil || params.SaveAsParams.SaveObject == "" {
				return "", fmt.Errorf("tos: SaveAsParams.SaveObject is required for video snapshots")
			}
			return process + buildPostSaveAsQuery(params.SaveAsParams), nil
		}
		if params.VideoProcessParams.PCMParams != nil {
			process, err := buildVideoPCMPostProcessString(params.VideoProcessParams.PCMParams)
			if err != nil {
				return "", err
			}
			if params.SaveAsParams != nil && params.SaveAsParams.SaveObject != "" {
				process += buildPostSaveAsQuery(params.SaveAsParams)
			}
			return process, nil
		}
		if params.VideoProcessParams.TranscodeParams != nil {
			return "", fmt.Errorf("tos: VideoProcessParams.TranscodeParams uses legacy JSON and is not supported by PostDataProcessHelper; use ConvertParams for synchronous processing or PostDataProcessAsync with TranscodeJobBody")
		}
		// 视频 snapshot 走 query-string 风格
		process, err := buildVideoProcessString(params.VideoProcessParams)
		if err != nil {
			return "", err
		}
		if params.SaveAsParams != nil {
			process += buildSaveAsString(params.SaveAsParams)
		}
		return process, nil
	case enum.PostProcessTypeDoc:
		if params.DocProcessParams == nil {
			return "", fmt.Errorf("tos: DocProcessParams is required for doc post process type")
		}
		process := buildDocProcessString(params.DocProcessParams)
		if params.SaveAsParams != nil {
			process += buildSaveAsString(params.SaveAsParams)
		}
		return process, nil
	default:
		return "", fmt.Errorf("tos: unsupported PostProcessType: %s", params.PostProcessType)
	}
}

// PostDataProcessAsyncHelper 将 JobBody 序列化为 Post 异步处理使用的 JSON 请求体。
// 返回值可直接赋给 PostDPAsyncInput.JobBody，并配合相同的 JobType 提交。
func PostDataProcessAsyncHelper(ctx context.Context, params PostDataProcessAsyncParams) (string, error) {
	if params.JobType == "" {
		return "", fmt.Errorf("tos: JobType is required for Post async processing")
	}
	if params.JobBody == nil {
		return "", fmt.Errorf("tos: JobBody is required for Post async processing")
	}
	data, _, err := marshalInput("PostDataProcessAsyncParams.JobBody", params.JobBody)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ==================== 内部拼接函数 ====================

// buildImageProcessString 将多个 ImageProcessParams 拼接为 "image/op1,kv1/op2,kv2" 格式。
func buildImageProcessString(params []ImageProcessParams) (string, error) {
	if len(params) == 0 {
		return "", fmt.Errorf("tos: ImageProcessParams is required for image process type")
	}

	var parts []string
	for _, p := range params {
		part, err := buildSingleImageOperation(p)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "image/" + strings.Join(parts, "/"), nil
}

func buildSingleImageOperation(p ImageProcessParams) (string, error) {
	op := string(p.Operation)
	var kvParts []string

	switch p.Operation {
	case enum.ImageOperationResize:
		if p.ResizeParams == nil {
			return op, nil
		}
		r := p.ResizeParams
		if r.M != "" {
			kvParts = append(kvParts, "m_"+r.M)
		}
		if r.W != nil {
			kvParts = append(kvParts, "w_"+strconv.Itoa(*r.W))
		}
		if r.H != nil {
			kvParts = append(kvParts, "h_"+strconv.Itoa(*r.H))
		}
		if r.L != nil {
			kvParts = append(kvParts, "l_"+strconv.Itoa(*r.L))
		}
		if r.S != nil {
			kvParts = append(kvParts, "s_"+strconv.Itoa(*r.S))
		}
		if r.P != nil {
			kvParts = append(kvParts, "p_"+strconv.Itoa(*r.P))
		}
		if r.Limit != nil {
			kvParts = append(kvParts, "limit_"+strconv.Itoa(*r.Limit))
		}
		if r.Color != "" {
			kvParts = append(kvParts, "color_"+r.Color)
		}

	case enum.ImageOperationWatermark:
		if p.WatermarkParams == nil {
			return op, nil
		}
		w := p.WatermarkParams
		if w.T != nil {
			kvParts = append(kvParts, "t_"+strconv.Itoa(*w.T))
		}
		if w.G != "" {
			kvParts = append(kvParts, "g_"+w.G)
		}
		if w.X != nil {
			kvParts = append(kvParts, "x_"+strconv.Itoa(*w.X))
		}
		if w.Y != nil {
			kvParts = append(kvParts, "y_"+strconv.Itoa(*w.Y))
		}
		if w.VOffset != nil {
			kvParts = append(kvParts, "voffset_"+strconv.Itoa(*w.VOffset))
		}
		if w.Image != "" {
			kvParts = append(kvParts, "image_"+w.Image)
		}
		if w.ImageP != nil {
			kvParts = append(kvParts, "P_"+strconv.Itoa(*w.ImageP))
		}
		if w.Text != "" {
			kvParts = append(kvParts, "text_"+w.Text)
		}
		if w.Type != "" {
			kvParts = append(kvParts, "type_"+w.Type)
		}
		if w.Color != "" {
			kvParts = append(kvParts, "color_"+w.Color)
		}
		if w.Size != nil {
			kvParts = append(kvParts, "size_"+strconv.Itoa(*w.Size))
		}
		if w.Shadow != nil {
			kvParts = append(kvParts, "shadow_"+strconv.Itoa(*w.Shadow))
		}
		if w.Rotate != nil {
			kvParts = append(kvParts, "rotate_"+strconv.Itoa(*w.Rotate))
		}
		if w.Fill != nil {
			kvParts = append(kvParts, "fill_"+strconv.Itoa(*w.Fill))
		}
		if w.Order != nil {
			kvParts = append(kvParts, "order_"+strconv.Itoa(*w.Order))
		}
		if w.Align != nil {
			kvParts = append(kvParts, "align_"+strconv.Itoa(*w.Align))
		}
		if w.Interval != nil {
			kvParts = append(kvParts, "interval_"+strconv.Itoa(*w.Interval))
		}

	case enum.ImageOperationBlindWatermark:
		if p.BlindWatermarkParams == nil {
			return op, nil
		}
		b := p.BlindWatermarkParams
		if b.Text != "" {
			kvParts = append(kvParts, "text_"+b.Text)
		}
		if b.Version != nil {
			kvParts = append(kvParts, "version_"+strconv.Itoa(*b.Version))
		}
		if b.Level != nil {
			kvParts = append(kvParts, "level_"+strconv.Itoa(*b.Level))
		}

	case enum.ImageOperationDeBlindWatermark:
		if p.DeBlindWatermarkParams == nil {
			return op, nil
		}
		if p.DeBlindWatermarkParams.Version != nil {
			kvParts = append(kvParts, "version_"+strconv.Itoa(*p.DeBlindWatermarkParams.Version))
		}

	case enum.ImageOperationSetAIGCMetadata:
		if p.AIGCMetadataParams == nil {
			return op, nil
		}
		a := p.AIGCMetadataParams
		if a.Label != "" {
			kvParts = append(kvParts, "Label_"+a.Label)
		}
		if a.ContentProducer != "" {
			kvParts = append(kvParts, "ContentProducer_"+a.ContentProducer)
		}
		if a.ProduceID != "" {
			kvParts = append(kvParts, "ProduceID_"+a.ProduceID)
		}
		if a.ContentPropagator != "" {
			kvParts = append(kvParts, "ContentPropagator_"+a.ContentPropagator)
		}
		if a.PropagateID != "" {
			kvParts = append(kvParts, "PropagateID_"+a.PropagateID)
		}
		if a.ReservedCode1 != "" {
			kvParts = append(kvParts, "ReservedCode1_"+a.ReservedCode1)
		}
		if a.ReservedCode2 != "" {
			kvParts = append(kvParts, "ReservedCode2_"+a.ReservedCode2)
		}

	case enum.ImageOperationGetAIGCMetadata, enum.ImageOperationGetC2PAMetadata, enum.ImageOperationAiTag:
		return op, nil

	case enum.ImageOperationSetC2PAMetadata:
		if p.SetC2PAMetadataParams == nil {
			return op, nil
		}
		if p.SetC2PAMetadataParams.AppID != "" {
			kvParts = append(kvParts, "AppID_"+p.SetC2PAMetadataParams.AppID)
		}
		if p.SetC2PAMetadataParams.Manifest != "" {
			kvParts = append(kvParts, "Manifest_"+p.SetC2PAMetadataParams.Manifest)
		}

	case enum.ImageOperationEmbedding:
		if p.EmbeddingParams == nil {
			return op, nil
		}
		e := p.EmbeddingParams
		if e.Embedder != "" {
			kvParts = append(kvParts, "embedder_"+e.Embedder)
		}
		if e.Dimension != nil {
			kvParts = append(kvParts, "dimension_"+strconv.Itoa(*e.Dimension))
		}
		if e.Instructions != nil {
			kvParts = append(kvParts, "instructions_"+*e.Instructions)
		}
		if e.EmbeddingType != nil {
			kvParts = append(kvParts, "embeddingType_"+strconv.Itoa(*e.EmbeddingType))
		}

	case enum.ImageOperationUnderstanding:
		if p.UnderstandingParams == nil {
			return op, nil
		}
		u := p.UnderstandingParams
		if u.Model != "" {
			kvParts = append(kvParts, "m_"+encodeBase64URLSafe(u.Model))
		}
		if u.Prompt != "" {
			kvParts = append(kvParts, "p_"+encodeBase64URLSafe(u.Prompt))
		}
		if u.Detail != "" {
			kvParts = append(kvParts, "d_"+u.Detail)
		}

	case enum.ImageOperationOCR:
		if p.OCRParams == nil {
			return op, nil
		}
		if p.OCRParams.Model != "" {
			kvParts = append(kvParts, "m_"+encodeBase64URLSafe(p.OCRParams.Model))
		}

	case enum.ImageOperationCrop:
		if p.CropParams == nil {
			return op, nil
		}
		c := p.CropParams
		if c.G != "" {
			kvParts = append(kvParts, "g_"+c.G)
		}
		if c.W != nil {
			kvParts = append(kvParts, "w_"+strconv.Itoa(*c.W))
		}
		if c.H != nil {
			kvParts = append(kvParts, "h_"+strconv.Itoa(*c.H))
		}
		if c.X != nil {
			kvParts = append(kvParts, "x_"+strconv.Itoa(*c.X))
		}
		if c.Y != nil {
			kvParts = append(kvParts, "y_"+strconv.Itoa(*c.Y))
		}

	case enum.ImageOperationCircle:
		if p.CircleParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, "r_"+strconv.Itoa(p.CircleParams.R))

	case enum.ImageOperationIndexcrop:
		if p.IndexcropParams == nil {
			return op, nil
		}
		ic := p.IndexcropParams
		if ic.X != nil {
			kvParts = append(kvParts, "x_"+strconv.Itoa(*ic.X))
		}
		if ic.Y != nil {
			kvParts = append(kvParts, "y_"+strconv.Itoa(*ic.Y))
		}
		if ic.I != nil {
			kvParts = append(kvParts, "i_"+strconv.Itoa(*ic.I))
		}

	case enum.ImageOperationClip:
		if p.ClipParams == nil {
			return op, nil
		}
		c := p.ClipParams
		if c.Frame != nil {
			kvParts = append(kvParts, "frame_"+strconv.Itoa(*c.Frame))
		}
		if c.First != nil {
			kvParts = append(kvParts, "first_"+strconv.Itoa(*c.First))
		}
		if c.Step != nil {
			kvParts = append(kvParts, "step_"+strconv.Itoa(*c.Step))
		}

	case enum.ImageOperationRoundedCorners:
		if p.RoundedCorners == nil {
			return op, nil
		}
		kvParts = append(kvParts, "r_"+strconv.Itoa(p.RoundedCorners.R))

	case enum.ImageOperationAutoOrient:
		if p.AutoOrientParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.AutoOrientParams.Value))

	case enum.ImageOperationAutoOrientInternal:
		return op, nil

	case enum.ImageOperationBlur:
		if p.BlurParams == nil {
			return op, nil
		}
		if p.BlurParams.Radius != nil {
			kvParts = append(kvParts, "r_"+strconv.Itoa(*p.BlurParams.Radius))
		}
		if p.BlurParams.Sigma != nil {
			kvParts = append(kvParts, "s_"+strconv.Itoa(*p.BlurParams.Sigma))
		}
		if len(kvParts) == 0 && p.BlurParams.Value != 0 {
			kvParts = append(kvParts, strconv.Itoa(p.BlurParams.Value))
		}

	case enum.ImageOperationRotate:
		if p.RotateParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.RotateParams.Value))

	case enum.ImageOperationBright:
		if p.BrightParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.BrightParams.Value))

	case enum.ImageOperationSharpen:
		if p.SharpenParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.SharpenParams.Value))

	case enum.ImageOperationContrast:
		if p.ContrastParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.ContrastParams.Value))

	case enum.ImageOperationGrayscale:
		op = "colorspace"
		kvParts = append(kvParts, "gray")

	case enum.ImageOperationColorspace:
		if p.ColorspaceParams == nil {
			return op, nil
		}
		if p.ColorspaceParams.Value != "" {
			kvParts = append(kvParts, p.ColorspaceParams.Value)
		}

	case enum.ImageOperationDraw:
		if p.DrawParams == nil {
			return op, nil
		}
		d := p.DrawParams
		if len(d.Points) > 0 {
			var pointStrs []string
			for _, pt := range d.Points {
				pointStrs = append(pointStrs, strconv.Itoa(pt.X)+"x"+strconv.Itoa(pt.Y))
			}
			kvParts = append(kvParts, "p_"+strings.Join(pointStrs, "-"))
		}
		if d.Radius != nil {
			kvParts = append(kvParts, "r_"+strconv.Itoa(*d.Radius))
		}
		if d.DrawLine != nil {
			if *d.DrawLine {
				kvParts = append(kvParts, "l_true")
			} else {
				kvParts = append(kvParts, "l_false")
			}
		}
		if d.LineWidth != nil {
			kvParts = append(kvParts, "lw_"+strconv.Itoa(*d.LineWidth))
		}
		if d.ColorRGB != nil {
			kvParts = append(kvParts, "color_"+*d.ColorRGB)
		}

	case enum.ImageOperationMosaic:
		if p.MosaicParams == nil {
			return op, nil
		}
		m := p.MosaicParams
		if m.G != "" {
			kvParts = append(kvParts, "g_"+m.G)
		}
		if m.W != nil {
			kvParts = append(kvParts, "w_"+strconv.Itoa(*m.W))
		}
		if m.H != nil {
			kvParts = append(kvParts, "h_"+strconv.Itoa(*m.H))
		}
		if m.X != nil {
			kvParts = append(kvParts, "x_"+strconv.Itoa(*m.X))
		}
		if m.Y != nil {
			kvParts = append(kvParts, "y_"+strconv.Itoa(*m.Y))
		}
		if m.T != "" {
			kvParts = append(kvParts, "t_"+m.T)
		}
		if m.R != nil {
			kvParts = append(kvParts, "r_"+strconv.Itoa(*m.R))
		}
		if m.S != nil {
			kvParts = append(kvParts, "s_"+strconv.Itoa(*m.S))
		}

	case enum.ImageOperationQuality:
		if p.QualityParams == nil {
			return op, nil
		}
		q := p.QualityParams
		if q.Q != nil {
			kvParts = append(kvParts, "q_"+strconv.Itoa(*q.Q))
		}
		if q.QQ != nil {
			kvParts = append(kvParts, "Q_"+strconv.Itoa(*q.QQ))
		}

	case enum.ImageOperationInterlace:
		if p.InterlaceParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, strconv.Itoa(p.InterlaceParams.Value))

	case enum.ImageOperationFormat:
		if p.FormatParams == nil {
			return op, nil
		}
		kvParts = append(kvParts, p.FormatParams.Format)

	case enum.ImageOperationSlim:
		if p.SlimParams == nil {
			return op, nil
		}
		if p.SlimParams.ZLevel != nil {
			kvParts = append(kvParts, "zlevel_"+strconv.Itoa(*p.SlimParams.ZLevel))
		}

	case enum.ImageOperationStrip, enum.ImageOperationInfo, enum.ImageOperationInspect, enum.ImageOperationAverageHue:
		// 这些操作没有参数，直接返回操作名
		return op, nil

	default:
		return op, nil
	}

	if len(kvParts) > 0 {
		return op + "," + strings.Join(kvParts, ","), nil
	}
	return op, nil
}

func buildVideoProcessString(params *VideoProcessParams) (string, error) {
	opCount := 0
	if params.InfoParams != nil {
		opCount++
	}
	if params.SnapshotParams != nil {
		opCount++
	}
	if params.PM3U8Params != nil {
		opCount++
	}
	if params.EmbeddingParams != nil {
		opCount++
	}
	if params.UnderstandingParams != nil {
		opCount++
	}
	if params.AIGCMetadataParams != nil {
		opCount++
	}
	if params.C2PAMetadataParams != nil {
		opCount++
	}
	if params.TranscodeParams != nil {
		opCount++
	}
	if opCount == 0 {
		return "", fmt.Errorf("tos: one video operation is required")
	}
	if opCount > 1 {
		return "", fmt.Errorf("tos: only one video operation can be specified")
	}
	if params.InfoParams != nil {
		return "video/info", nil
	}
	if params.SnapshotParams != nil {
		return buildVideoSnapshotString(params.SnapshotParams), nil
	}
	if params.PM3U8Params != nil {
		return "video/pm3u8", nil
	}
	if params.EmbeddingParams != nil {
		return buildVideoEmbeddingString(params.EmbeddingParams)
	}
	if params.UnderstandingParams != nil {
		return buildVideoUnderstandingString(params.UnderstandingParams)
	}
	if params.AIGCMetadataParams != nil {
		return "video/aigcmetadata", nil
	}
	if params.C2PAMetadataParams != nil {
		return "video/c2pametadata", nil
	}
	if params.TranscodeParams != nil {
		data, err := json.Marshal(params.TranscodeParams)
		if err != nil {
			return "", fmt.Errorf("tos: failed to marshal VideoTranscodeParams: %w", err)
		}
		return string(data), nil
	}
	return "", fmt.Errorf("tos: one video operation is required")
}

func buildVideoEmbeddingString(params *VideoEmbeddingParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: VideoEmbeddingParams is required")
	}
	parts := []string{"video/embedding"}
	if params.Embedder != "" {
		parts = append(parts, "embedder_"+params.Embedder)
	}
	if params.Dimension != nil {
		parts = append(parts, "dimension_"+strconv.Itoa(*params.Dimension))
	}
	if params.Instructions != nil {
		parts = append(parts, "instructions_"+*params.Instructions)
	}
	if len(parts) == 1 {
		return "", fmt.Errorf("tos: VideoEmbeddingParams is required")
	}
	return strings.Join(parts, ","), nil
}

func buildVideoUnderstandingString(params *VideoUnderstandingParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: VideoUnderstandingParams is required")
	}
	if params.Model == "" || params.Prompt == "" {
		return "", fmt.Errorf("tos: VideoUnderstandingParams.Model and Prompt are required")
	}
	parts := []string{"video/understanding", "m_" + encodeBase64URLSafe(params.Model), "p_" + encodeBase64URLSafe(params.Prompt)}
	if params.FPS != nil {
		parts = append(parts, "fps_"+strconv.FormatFloat(*params.FPS, 'f', -1, 64))
	}
	return strings.Join(parts, ","), nil
}

func buildVideoSnapshotsPostProcessString(params *VideoSnapshotsParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: VideoSnapshotsParams is required")
	}
	var parts []string
	parts = append(parts, "video/snapshots")
	if params.Format != "" {
		parts = append(parts, "f_"+params.Format)
	}
	if params.Mode != "" {
		parts = append(parts, "m_"+params.Mode)
	}
	if params.Width != nil {
		parts = append(parts, "w_"+strconv.Itoa(*params.Width))
	}
	if params.Height != nil {
		parts = append(parts, "h_"+strconv.Itoa(*params.Height))
	}
	if params.Index != "" {
		parts = append(parts, "index_"+params.Index)
	}
	if params.Num != nil {
		parts = append(parts, "num_"+strconv.Itoa(*params.Num))
	}
	if params.DHashThreshold != nil {
		parts = append(parts, "dhash_"+strconv.Itoa(*params.DHashThreshold))
	}
	if len(parts) == 1 {
		return "", fmt.Errorf("tos: VideoSnapshotsParams is required")
	}
	return strings.Join(parts, ","), nil
}

func buildVideoPCMPostProcessString(params *VideoPCMParams) (string, error) {
	sampleRate := 16000
	channels := 1
	if params != nil {
		if params.SampleRate != nil {
			sampleRate = *params.SampleRate
		}
		if params.Channels != nil {
			channels = *params.Channels
		}
	}

	if !isValidPCMSampleRate(sampleRate) {
		return "", fmt.Errorf("tos: unsupported PCM sample rate: %d", sampleRate)
	}
	if channels != 1 && channels != 2 {
		return "", fmt.Errorf("tos: unsupported PCM channels: %d", channels)
	}

	return fmt.Sprintf("video/convert,f_pcm,acodec_pcm_s16le,ac_%d,ar_%d,vn_1", channels, sampleRate), nil
}

func countVideoProcessOperations(params *VideoProcessParams) int {
	if params == nil {
		return 0
	}
	count := 0
	if params.InfoParams != nil {
		count++
	}
	if params.SnapshotParams != nil {
		count++
	}
	if params.PM3U8Params != nil {
		count++
	}
	if params.EmbeddingParams != nil {
		count++
	}
	if params.UnderstandingParams != nil {
		count++
	}
	if params.AIGCMetadataParams != nil {
		count++
	}
	if params.C2PAMetadataParams != nil {
		count++
	}
	if params.SnapshotsParams != nil {
		count++
	}
	if params.TranscodeParams != nil {
		count++
	}
	if params.PCMParams != nil {
		count++
	}
	if params.ConvertParams != nil {
		count++
	}
	if params.RemuxParams != nil {
		count++
	}
	return count
}

func buildVideoConvertPostProcessString(params *VideoConvertParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: VideoConvertParams is required")
	}
	if params.Format == "" {
		return "", fmt.Errorf("tos: VideoConvertParams.Format is required")
	}

	parts := []string{"video/convert", "f_" + params.Format}
	if params.StartTime != nil {
		if *params.StartTime < 0 {
			return "", fmt.Errorf("tos: VideoConvertParams.StartTime must be non-negative")
		}
		parts = append(parts, "ss_"+strconv.Itoa(*params.StartTime))
	}
	if params.Duration != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.Duration", *params.Duration, 0, 900000); err != nil {
			return "", err
		}
		parts = append(parts, "t_"+strconv.Itoa(*params.Duration))
	}
	if params.HLSSegmentDuration != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.HLSSegmentDuration", *params.HLSSegmentDuration, 0, 3600000); err != nil {
			return "", err
		}
		parts = append(parts, "st_"+strconv.Itoa(*params.HLSSegmentDuration))
	}
	if params.RemoveVideo != nil {
		if err := validatePostVideoFlag("VideoConvertParams.RemoveVideo", *params.RemoveVideo); err != nil {
			return "", err
		}
		parts = append(parts, "vn_"+strconv.Itoa(*params.RemoveVideo))
	}
	if params.VideoCodec != "" {
		parts = append(parts, "vcodec_"+params.VideoCodec)
	}
	if params.FPS != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.FPS", *params.FPS, 0, 60); err != nil {
			return "", err
		}
		parts = append(parts, "fps_"+strconv.Itoa(*params.FPS))
	}
	if params.PixelFormat != "" {
		parts = append(parts, "pixfmt_"+params.PixelFormat)
	}
	if params.Width != nil {
		if err := validatePostVideoDimension("VideoConvertParams.Width", *params.Width); err != nil {
			return "", err
		}
		parts = append(parts, "w_"+strconv.Itoa(*params.Width))
	}
	if params.Height != nil {
		if err := validatePostVideoDimension("VideoConvertParams.Height", *params.Height); err != nil {
			return "", err
		}
		parts = append(parts, "h_"+strconv.Itoa(*params.Height))
	}
	if params.VideoBitRate != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.VideoBitRate", *params.VideoBitRate, 10000, 50000000); err != nil {
			return "", err
		}
		parts = append(parts, "vb_"+strconv.Itoa(*params.VideoBitRate))
	}
	if params.MaxRate != nil {
		if params.VideoBitRate == nil {
			return "", fmt.Errorf("tos: VideoConvertParams.MaxRate requires VideoBitRate")
		}
		if err := validatePostVideoIntRange("VideoConvertParams.MaxRate", *params.MaxRate, 10000, 50000000); err != nil {
			return "", err
		}
		if *params.MaxRate < *params.VideoBitRate {
			return "", fmt.Errorf("tos: VideoConvertParams.MaxRate must be greater than or equal to VideoBitRate")
		}
		parts = append(parts, "maxrate_"+strconv.Itoa(*params.MaxRate))
	}
	if params.BufferSize != nil {
		if params.MaxRate == nil {
			return "", fmt.Errorf("tos: VideoConvertParams.BufferSize requires MaxRate")
		}
		if err := validatePostVideoIntRange("VideoConvertParams.BufferSize", *params.BufferSize, 1000000, 128000000); err != nil {
			return "", err
		}
		parts = append(parts, "bufsize_"+strconv.Itoa(*params.BufferSize))
	}
	if params.CRF != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.CRF", *params.CRF, 0, 51); err != nil {
			return "", err
		}
		parts = append(parts, "crf_"+strconv.Itoa(*params.CRF))
	}
	if params.RemoveAudio != nil {
		if err := validatePostVideoFlag("VideoConvertParams.RemoveAudio", *params.RemoveAudio); err != nil {
			return "", err
		}
		parts = append(parts, "an_"+strconv.Itoa(*params.RemoveAudio))
	}
	if params.AudioCodec != "" {
		parts = append(parts, "acodec_"+params.AudioCodec)
	}
	if params.SampleRate != nil {
		if *params.SampleRate < 0 {
			return "", fmt.Errorf("tos: VideoConvertParams.SampleRate must be non-negative")
		}
		parts = append(parts, "ar_"+strconv.Itoa(*params.SampleRate))
	}
	if params.Channels != nil {
		if *params.Channels < 0 {
			return "", fmt.Errorf("tos: VideoConvertParams.Channels must be non-negative")
		}
		parts = append(parts, "ac_"+strconv.Itoa(*params.Channels))
	}
	if params.AudioBitRate != nil {
		if err := validatePostVideoIntRange("VideoConvertParams.AudioBitRate", *params.AudioBitRate, 8000, 1000000); err != nil {
			return "", err
		}
		parts = append(parts, "ab_"+strconv.Itoa(*params.AudioBitRate))
	}
	if params.SampleFormat != "" {
		parts = append(parts, "af_"+params.SampleFormat)
	}

	var err error
	if params.AIGCMetadata != nil {
		parts, err = appendPostVideoJSONParam(parts, "aigc_", params.AIGCMetadata)
		if err != nil {
			return "", err
		}
	}
	if params.C2PAMetadata != nil {
		parts, err = appendPostVideoJSONParam(parts, "c2pa_", params.C2PAMetadata)
		if err != nil {
			return "", err
		}
	}
	if len(params.Watermarks) > 0 {
		parts, err = appendPostVideoJSONParam(parts, "watermark_", params.Watermarks)
		if err != nil {
			return "", err
		}
	}
	if params.BlindWatermark != nil {
		parts, err = appendPostVideoJSONParam(parts, "blindwatermark_", params.BlindWatermark)
		if err != nil {
			return "", err
		}
	}
	return strings.Join(parts, ","), nil
}

func buildVideoRemuxPostProcessString(params *VideoRemuxParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: VideoRemuxParams is required")
	}
	if params.Format == "" {
		return "", fmt.Errorf("tos: VideoRemuxParams.Format is required")
	}

	parts := []string{"video/remux", "f_" + params.Format}
	if params.HLSSegmentDuration != nil {
		if err := validatePostVideoIntRange("VideoRemuxParams.HLSSegmentDuration", *params.HLSSegmentDuration, 0, 3600000); err != nil {
			return "", err
		}
		parts = append(parts, "st_"+strconv.Itoa(*params.HLSSegmentDuration))
	}
	if params.StreamIndex != nil {
		if *params.StreamIndex < 0 {
			return "", fmt.Errorf("tos: VideoRemuxParams.StreamIndex must be non-negative")
		}
		parts = append(parts, "ti_"+strconv.Itoa(*params.StreamIndex))
	}

	var err error
	if params.AIGCMetadata != nil {
		parts, err = appendPostVideoJSONParam(parts, "aigc_", params.AIGCMetadata)
		if err != nil {
			return "", err
		}
	}
	if params.C2PAMetadata != nil {
		parts, err = appendPostVideoJSONParam(parts, "c2pa_", params.C2PAMetadata)
		if err != nil {
			return "", err
		}
	}
	return strings.Join(parts, ","), nil
}

func appendPostVideoJSONParam(parts []string, prefix string, value interface{}) ([]string, error) {
	encoded, err := encodePostVideoJSON(value)
	if err != nil {
		return nil, err
	}
	return append(parts, prefix+encoded), nil
}

func encodePostVideoJSON(value interface{}) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("tos: failed to marshal Post video parameter: %w", err)
	}
	return encodeBase64URLSafe(string(data)), nil
}

func validatePostVideoFlag(name string, value int) error {
	if value != 0 && value != 1 {
		return fmt.Errorf("tos: %s must be 0 or 1", name)
	}
	return nil
}

func validatePostVideoIntRange(name string, value, min, max int) error {
	if value < min || value > max {
		return fmt.Errorf("tos: %s must be between %d and %d", name, min, max)
	}
	return nil
}

func validatePostVideoDimension(name string, value int) error {
	if value < 128 || value > 4096 || value%2 != 0 {
		return fmt.Errorf("tos: %s must be an even number between 128 and 4096", name)
	}
	return nil
}

func isValidPCMSampleRate(sampleRate int) bool {
	switch sampleRate {
	case 8000, 11025, 16000, 22050, 24000, 32000, 44100, 48000, 88200, 96000:
		return true
	default:
		return false
	}
}

func buildVideoSnapshotString(params *VideoSnapshotParams) string {
	var kvParts []string
	kvParts = append(kvParts, "t_"+strconv.Itoa(params.T))
	if params.W != nil {
		kvParts = append(kvParts, "w_"+strconv.Itoa(*params.W))
	}
	if params.H != nil {
		kvParts = append(kvParts, "h_"+strconv.Itoa(*params.H))
	}
	if params.M != "" {
		kvParts = append(kvParts, "m_"+params.M)
	}
	if params.F != "" {
		kvParts = append(kvParts, "f_"+params.F)
	}
	if params.Rotate != nil {
		kvParts = append(kvParts, "rotate_"+strconv.Itoa(*params.Rotate))
	}
	return "video/snapshot," + strings.Join(kvParts, ",")
}

func buildDocProcessString(params *DocProcessParams) string {
	parts := []string{"doc-preview"}
	if params.SrcType != "" {
		parts = append(parts, "src_type_"+string(params.SrcType))
	}
	if params.DstType != "" {
		parts = append(parts, "dst_type_"+string(params.DstType))
	}
	if params.DocPage != nil {
		parts = append(parts, "doc_page_"+strconv.Itoa(*params.DocPage))
	}
	if params.StartPage != nil {
		parts = append(parts, "start_page_"+strconv.Itoa(*params.StartPage))
	}
	if params.EndPage != nil {
		parts = append(parts, "end_page_"+strconv.Itoa(*params.EndPage))
	}
	if params.ImageMode != nil {
		parts = append(parts, "image_mode_"+strconv.Itoa(int(*params.ImageMode)))
	}
	if len(parts) == 1 {
		return "doc-preview"
	}
	return strings.Join(parts, ",")
}

func buildHlsProcessString(params *HlsProcessParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: HlsProcessParams is required")
	}

	opCount := 0
	if params.M3U8Params != nil {
		opCount++
	}
	if params.TSParams != nil {
		opCount++
	}
	if opCount != 1 {
		return "", fmt.Errorf("tos: exactly one hls operation must be specified")
	}

	if params.M3U8Params != nil {
		return buildHlsM3U8ProcessString(params.M3U8Params)
	}
	return buildHlsTSProcessString(params.TSParams)
}

func buildHlsM3U8ProcessString(params *HlsM3U8Params) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: HlsM3U8Params is required")
	}
	if params.Width <= 0 || params.Height <= 0 {
		return "", fmt.Errorf("tos: HlsM3U8Params.Width and Height must be positive")
	}
	if params.EncodeFormat == "" {
		return "", fmt.Errorf("tos: HlsM3U8Params.EncodeFormat is required")
	}
	if params.PixFmt == "" {
		return "", fmt.Errorf("tos: HlsM3U8Params.PixFmt is required")
	}

	parts := []string{
		"hls/m3u8",
		"w_" + strconv.Itoa(params.Width),
		"h_" + strconv.Itoa(params.Height),
		"ef_" + params.EncodeFormat,
		"pf_" + params.PixFmt,
	}
	if params.SegmentDuration != nil {
		parts = append(parts, "sd_"+strconv.Itoa(*params.SegmentDuration))
	}
	for _, wm := range params.WatermarkTemplateIDs {
		if wm != "" {
			parts = append(parts, "wm_"+wm)
		}
	}
	return strings.Join(parts, ","), nil
}

func buildHlsTSProcessString(params *HlsTSParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: HlsTSParams is required")
	}
	if params.FromObject == "" {
		return "", fmt.Errorf("tos: HlsTSParams.FromObject is required")
	}
	if params.Width <= 0 || params.Height <= 0 {
		return "", fmt.Errorf("tos: HlsTSParams.Width and Height must be positive")
	}
	if params.EncodeFormat == "" {
		return "", fmt.Errorf("tos: HlsTSParams.EncodeFormat is required")
	}
	if params.PixFmt == "" {
		return "", fmt.Errorf("tos: HlsTSParams.PixFmt is required")
	}

	downloadFlag := 0
	if params.NeedDownload != nil && *params.NeedDownload {
		downloadFlag = 1
	}

	parts := []string{
		"hls/ts",
		"from_" + base64.URLEncoding.EncodeToString([]byte(params.FromObject)),
		"w_" + strconv.Itoa(params.Width),
		"h_" + strconv.Itoa(params.Height),
		"ef_" + params.EncodeFormat,
		"pf_" + params.PixFmt,
		"st_" + strconv.FormatInt(params.StartTime, 10),
		"et_" + strconv.FormatInt(params.EndTime, 10),
		"dl_" + strconv.Itoa(downloadFlag),
	}
	for _, wm := range params.WatermarkTemplateIDs {
		if wm != "" {
			parts = append(parts, "wm_"+wm)
		}
	}
	return strings.Join(parts, ","), nil
}

// buildSaveAsString 拼接 Get/Post 路径下 process 字符串末尾的 "|sys/saveas,o_xxx,b_xxx" 段。
func buildSaveAsString(params *SaveAsParams) string {
	var parts []string
	if params.SaveObject != "" {
		encoded := base64.URLEncoding.EncodeToString([]byte(params.SaveObject))
		parts = append(parts, "o_"+encoded)
	}
	if params.SaveBucket != "" {
		encoded := base64.URLEncoding.EncodeToString([]byte(params.SaveBucket))
		parts = append(parts, "b_"+encoded)
	}
	if len(parts) == 0 {
		return ""
	}
	return "|sys/saveas," + strings.Join(parts, ",")
}

// buildPostSaveAsQuery 拼接 Post 路径下 video/snapshots 和 video/PCM 的 SaveAs query 参数段（"&x-tos-save-object=xxx&x-tos-save-bucket=xxx"）。
func buildPostSaveAsQuery(params *SaveAsParams) string {
	if params == nil {
		return ""
	}
	var parts []string
	if params.SaveObject != "" {
		parts = append(parts, "x-tos-save-object="+base64.URLEncoding.EncodeToString([]byte(params.SaveObject)))
	}
	if params.SaveBucket != "" {
		parts = append(parts, "x-tos-save-bucket="+base64.URLEncoding.EncodeToString([]byte(params.SaveBucket)))
	}
	if len(parts) == 0 {
		return ""
	}
	return "&" + strings.Join(parts, "&")
}

func buildPostVideoSaveAsQuery(params *SaveAsParams) (string, error) {
	if params == nil || params.SaveObject == "" {
		return "", fmt.Errorf("tos: SaveAsParams.SaveObject is required for Post video processing")
	}

	parts := []string{"x-tos-save-object=" + encodeBase64URLSafe(params.SaveObject)}
	if params.SaveBucket != "" {
		parts = append(parts, "x-tos-save-bucket="+encodeBase64URLSafe(params.SaveBucket))
	}
	return "&" + strings.Join(parts, "&"), nil
}

func buildPostImageSaveAsQuery(params *SaveAsParams) (string, error) {
	if params == nil || params.SaveObject == "" {
		return "", fmt.Errorf("tos: SaveAsParams.SaveObject is required for Post image processing")
	}

	parts := []string{"x-tos-save-object=" + encodeBase64URLSafe(params.SaveObject)}
	if params.SaveBucket != "" {
		parts = append(parts, "x-tos-save-bucket="+encodeBase64URLSafe(params.SaveBucket))
	}
	return "&" + strings.Join(parts, "&"), nil
}

// ==================== MCAP 处理字符串拼接 ====================

func buildMcapProcessString(params *McapProcessParams) (string, error) {
	if params == nil {
		return "", fmt.Errorf("tos: McapProcessParams is required")
	}
	switch params.Operation {
	case McapOperationInfo:
		return "file/mcap-info", nil
	case McapOperationDoctor:
		return "file/mcap-doctor", nil
	case McapOperationList:
		if params.ListResource == "" {
			return "", fmt.Errorf("tos: McapProcessParams.ListResource is required when Operation is mcap-list")
		}
		return "file/mcap-list," + string(params.ListResource), nil
	default:
		return "", fmt.Errorf("tos: unsupported McapOperation: %s", params.Operation)
	}
}

func encodeBase64URLSafe(s string) string {
	encoded := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.TrimRight(encoded, "=")
}
