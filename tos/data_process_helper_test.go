package tos

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestGetDataProcessHelper_ImageResize(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationResize,
				ResizeParams: &ImageResizeParams{
					M: "lfit",
					W: intPtr(100),
					H: intPtr(200),
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "image/resize,m_lfit,w_100,h_200"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_ImageMultipleOperations(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationResize,
				ResizeParams: &ImageResizeParams{
					W: intPtr(200),
				},
			},
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &ImageFormatParams{Format: "jpg"},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "image/resize,w_200/format,jpg"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_ImageWithSaveAs(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationResize,
				ResizeParams: &ImageResizeParams{
					W: intPtr(100),
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Get 路径下 SaveAs 不拼入 process 字符串，由调用方直接设到 GetObjectV2Input
	expected := "image/resize,w_100"
	if result != expected {
		t.Errorf("expected process %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_InspectSaveAsIgnored(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationInspect},
		},
	}
	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "image/inspect" {
		t.Errorf("expected process 'image/inspect', got %q", result)
	}
}

func TestGetDataProcessHelper_VideoSnapshot(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			SnapshotParams: &VideoSnapshotParams{
				T: 1000,
				W: intPtr(640),
				F: "jpg",
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "video/snapshot,t_1000,w_640,f_jpg"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_VideoExtendedOps(t *testing.T) {
	ctx := context.Background()
	dimension := 512
	instructions := "aW5zdHJ1Y3Rpb25z"
	fps := 1.5

	cases := []struct {
		name     string
		params   *VideoProcessParams
		expected string
	}{
		{name: "info", params: &VideoProcessParams{InfoParams: &VideoInfoParams{}}, expected: "video/info"},
		{name: "pm3u8", params: &VideoProcessParams{PM3U8Params: &VideoPM3U8Params{}}, expected: "video/pm3u8"},
		{name: "embedding", params: &VideoProcessParams{EmbeddingParams: &VideoEmbeddingParams{Embedder: "doubao-video-embedding", Dimension: &dimension, Instructions: &instructions}}, expected: "video/embedding,embedder_doubao-video-embedding,dimension_512,instructions_aW5zdHJ1Y3Rpb25z"},
		{name: "understanding", params: &VideoProcessParams{UnderstandingParams: &VideoUnderstandingParams{Model: "model", Prompt: "prompt", FPS: &fps}}, expected: "video/understanding,m_bW9kZWw,p_cHJvbXB0,fps_1.5"},
		{name: "aigcmetadata", params: &VideoProcessParams{AIGCMetadataParams: &VideoAIGCMetadataParams{}}, expected: "video/aigcmetadata"},
		{name: "c2pametadata", params: &VideoProcessParams{C2PAMetadataParams: &VideoC2PAMetadataParams{}}, expected: "video/c2pametadata"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetDataProcessHelper(ctx, GetDataProcessParams{
				GetProcessType:     enum.GetProcessTypeVideo,
				VideoProcessParams: tc.params,
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGetDataProcessHelper_PostVideoOpsError(t *testing.T) {
	ctx := context.Background()
	_, err := GetDataProcessHelper(ctx, GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			ConvertParams: &VideoConvertParams{Format: "mp4"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "only supported by PostDataProcess") {
		t.Fatalf("expected Post-only video operation error, got %v", err)
	}
}

func TestGetDataProcessHelper_Doc(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeDoc,
		DocProcessParams: &DocProcessParams{
			SrcType:   "docx",
			DstType:   "pdf",
			StartPage: intPtr(1),
			EndPage:   intPtr(10),
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "doc-preview") {
		t.Errorf("expected doc-preview prefix, got %q", result)
	}
	if !strings.Contains(result, "src_type_docx") {
		t.Errorf("expected src_type_docx in result: %q", result)
	}
	if !strings.Contains(result, "dst_type_pdf") {
		t.Errorf("expected dst_type_pdf in result: %q", result)
	}
}

func TestGetDataProcessHelper_InvalidType(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: "invalid",
	}

	_, err := GetDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error for invalid process type")
	}
}

func TestGetDataProcessHelper_ImageNoParams(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
	}

	_, err := GetDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error when no ImageProcessParams provided")
	}
}

func TestGetDataProcessHelper_ImageInfo(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationInfo},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "image/info"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_ImageWatermark(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationWatermark,
				WatermarkParams: &ImageWatermarkParams{
					Text:  "dGVzdA==",
					Size:  intPtr(30),
					G:     "south",
					Color: "FF0000",
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "watermark") {
		t.Errorf("expected watermark in result: %q", result)
	}
	if !strings.Contains(result, "text_dGVzdA==") {
		t.Errorf("expected text param in result: %q", result)
	}
	if !strings.Contains(result, "size_30") {
		t.Errorf("expected size param in result: %q", result)
	}
}

func TestGetDataProcessHelper_ImageBlindWatermark(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &ImageBlindWatermarkParams{
					Text:    "YmxpbmQtd2F0ZXJtYXJr",
					Version: intPtr(2),
					Level:   intPtr(2),
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "blindwatermark") {
		t.Errorf("expected blindwatermark in result: %q", result)
	}
	if !strings.Contains(result, "text_YmxpbmQtd2F0ZXJtYXJr") {
		t.Errorf("expected text param in result: %q", result)
	}
	if !strings.Contains(result, "version_2") {
		t.Errorf("expected version param in result: %q", result)
	}
	if !strings.Contains(result, "level_2") {
		t.Errorf("expected level param in result: %q", result)
	}
}

func TestGetDataProcessHelper_ImageExtendedBasicOps(t *testing.T) {
	ctx := context.Background()
	frame := 2
	first := 1
	step := 3
	grayscale := 1
	radius := 2
	sigma := 3
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationAutoOrientInternal, AutoOrientInternalParams: &ImageAutoOrientInternalParams{}},
			{Operation: enum.ImageOperationClip, ClipParams: &ImageClipParams{Frame: &frame, First: &first, Step: &step}},
			{Operation: enum.ImageOperationIndexcrop, IndexcropParams: &ImageIndexcropParams{X: intPtr(1), Y: intPtr(1), I: intPtr(0)}},
			{Operation: enum.ImageOperationGrayscale, GrayscaleParams: &ImageGrayscaleParams{Value: grayscale}},
			{Operation: enum.ImageOperationColorspace, ColorspaceParams: &ImageColorspaceParams{Value: "gray"}},
			{Operation: enum.ImageOperationBlur, BlurParams: &ImageBlurParams{Radius: &radius, Sigma: &sigma}},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{
		"auto-orient-internal",
		"clip,frame_2,first_1,step_3",
		"indexcrop,x_1,y_1,i_0",
		"colorspace,gray",
		"blur,r_2,s_3",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected %q in result: %q", expected, result)
		}
	}
}

func TestGetDataProcessHelper_ImageExtendedMetadataOps(t *testing.T) {
	ctx := context.Background()
	version := 2
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationDeBlindWatermark, DeBlindWatermarkParams: &ImageDeBlindWatermarkParams{Version: &version}},
			{Operation: enum.ImageOperationGetAIGCMetadata, GetAIGCMetadataParams: &ImageGetAIGCMetadataParams{}},
			{Operation: enum.ImageOperationSetC2PAMetadata, SetC2PAMetadataParams: &ImageSetC2PAMetadataParams{AppID: "app-1", Manifest: "bWFuaWZlc3Q="}},
			{Operation: enum.ImageOperationGetC2PAMetadata, GetC2PAMetadataParams: &ImageGetC2PAMetadataParams{}},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{
		"deblindwatermark,version_2",
		"getaigcmetadata",
		"setc2pametadata,AppID_app-1,Manifest_bWFuaWZlc3Q=",
		"getc2pametadata",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected %q in result: %q", expected, result)
		}
	}
}

func TestGetDataProcessHelper_ImageExtendedCompressionOps(t *testing.T) {
	ctx := context.Background()
	zlevel := 5
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationSlim, SlimParams: &ImageSlimParams{ZLevel: &zlevel}},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{
		"slim,zlevel_5",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected %q in result: %q", expected, result)
		}
	}
}

func TestGetDataProcessHelper_ImageExtendedAIOps(t *testing.T) {
	ctx := context.Background()
	dimension := 512
	embeddingType := 2
	instructions := "aW5zdHJ1Y3Rpb25z"
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{Operation: enum.ImageOperationAiTag, AiTagParams: &ImageAiTagParams{}},
			{
				Operation: enum.ImageOperationEmbedding,
				EmbeddingParams: &ImageEmbeddingParams{
					Embedder:      "doubaoEmbeddingVision250328",
					Dimension:     &dimension,
					Instructions:  &instructions,
					EmbeddingType: &embeddingType,
				},
			},
			{
				Operation: enum.ImageOperationUnderstanding,
				UnderstandingParams: &ImageUnderstandingParams{
					Model:  "model",
					Prompt: "prompt",
					Detail: "high",
				},
			},
			{
				Operation: enum.ImageOperationOCR,
				OCRParams: &ImageOCRParams{
					Model: "model",
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, expected := range []string{
		"aitag",
		"embedding,embedder_doubaoEmbeddingVision250328,dimension_512,instructions_aW5zdHJ1Y3Rpb25z,embeddingType_2",
		"understanding,m_bW9kZWw,p_cHJvbXB0,d_high",
		"ocr,m_bW9kZWw",
	} {
		if !strings.Contains(result, expected) {
			t.Errorf("expected %q in result: %q", expected, result)
		}
	}
}

func TestGetDataProcessHelper_ImageQuality(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation:     enum.ImageOperationQuality,
				QualityParams: &ImageQualityParams{Q: intPtr(80)},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "image/quality,q_80"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_ImageDraw(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationDraw,
				DrawParams: &ImageDrawParams{
					Points:   []ImagePoint{{X: 50, Y: 50}, {X: 100, Y: 100}},
					Radius:   intPtr(8),
					DrawLine: boolPtr(true),
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "draw") {
		t.Errorf("expected draw in result: %q", result)
	}
	if !strings.Contains(result, "p_50x50-100x100") {
		t.Errorf("expected points in result: %q", result)
	}
	if !strings.Contains(result, "r_8") {
		t.Errorf("expected radius in result: %q", result)
	}
	if !strings.Contains(result, "l_true") {
		t.Errorf("expected drawline in result: %q", result)
	}
}

func TestGetDataProcessHelper_HlsM3U8(t *testing.T) {
	ctx := context.Background()
	segmentDuration := 10
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeHls,
		HlsProcessParams: &HlsProcessParams{
			M3U8Params: &HlsM3U8Params{
				SegmentDuration: &segmentDuration,
				Width:           1280,
				Height:          720,
				EncodeFormat:    "h264",
				PixFmt:          "yuv420p",
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hls/m3u8,w_1280,h_720,ef_h264,pf_yuv420p,sd_10"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_HlsTS(t *testing.T) {
	ctx := context.Background()
	needDownload := true
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeHls,
		HlsProcessParams: &HlsProcessParams{
			TSParams: &HlsTSParams{
				FromObject:   "test.mp4",
				Width:        1280,
				Height:       720,
				EncodeFormat: "h264",
				PixFmt:       "yuv420p",
				StartTime:    0,
				EndTime:      1000000,
				NeedDownload: &needDownload,
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "hls/ts,from_dGVzdC5tcDQ=,w_1280,h_720,ef_h264,pf_yuv420p,st_0,et_1000000,dl_1"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestPutDataProcessHelper_ImageOperations(t *testing.T) {
	ctx := context.Background()
	params := PutDataProcessParams{
		PutProcessType: enum.PutProcessTypeImage,
		ImageOperations: &PutImageOperationsParams{
			IsImageInfo: intPtr(1),
			Rules: []PutImageOperationsRule{
				{Bucket: "mybucket", Key: "output.jpg", Rule: "image/resize,w_100"},
			},
		},
	}

	result, err := PutDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "{") {
		t.Fatalf("expected JSON string for ImageOperations, got: %s", result)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsed); err != nil {
		t.Fatalf("ImageOperations result is not valid JSON: %v", err)
	}
}

func TestPutDataProcessHelper_VideoTranscode(t *testing.T) {
	ctx := context.Background()
	params := PutDataProcessParams{
		PutProcessType: enum.PutProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			TranscodeParams: &VideoTranscodeParams{
				Tag:  "Transcode",
				Name: "test-transcode",
			},
		},
	}

	result, err := PutDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "{") {
		t.Fatalf("expected JSON string for VideoTranscode, got: %s", result)
	}
}

func TestPutDataProcessHelper_PostVideoOpsError(t *testing.T) {
	ctx := context.Background()
	_, err := PutDataProcessHelper(ctx, PutDataProcessParams{
		PutProcessType: enum.PutProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			RemuxParams: &VideoRemuxParams{Format: "mp4"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "only supported by PostDataProcess") {
		t.Fatalf("expected Post-only video operation error, got %v", err)
	}
}

func TestPostDataProcessHelper_VideoSnapshot(t *testing.T) {
	ctx := context.Background()
	params := PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			SnapshotParams: &VideoSnapshotParams{
				T: 2000,
				W: intPtr(320),
				H: intPtr(240),
				F: "png",
			},
		},
	}

	result, err := PostDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "video/snapshot,t_2000,w_320,h_240,f_png"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestPostDataProcessHelper_VideoSnapshots(t *testing.T) {
	ctx := context.Background()
	width := 400
	height := 400
	params := PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			SnapshotsParams: &VideoSnapshotsParams{
				Format: "png",
				Mode:   "index",
				Width:  &width,
				Height: &height,
				Index:  "0|10|3000",
			},
		},
		SaveAsParams: &SaveAsParams{
			SaveObject: "dp-test/video/post_snapshot_${Number}.png",
		},
	}

	result, err := PostDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "video/snapshots,f_png,m_index,w_400,h_400,index_0|10|3000"
	if !strings.Contains(result, expected) {
		t.Errorf("expected result to contain %q, got %q", expected, result)
	}
	if !strings.Contains(result, "&x-tos-save-object=") {
		t.Errorf("expected save object in result: %q", result)
	}
}

func TestPostDataProcessHelper_VideoPCM(t *testing.T) {
	ctx := context.Background()
	sampleRate := 16000
	channels := 1
	params := PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			PCMParams: &VideoPCMParams{
				SampleRate: &sampleRate,
				Channels:   &channels,
			},
		},
		SaveAsParams: &SaveAsParams{
			SaveBucket: "target-bucket",
			SaveObject: "audio/output.pcm",
		},
	}

	result, err := PostDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "video/convert,f_pcm,acodec_pcm_s16le,ac_1,ar_16000,vn_1"
	if !strings.Contains(result, expected) {
		t.Errorf("expected result to contain %q, got %q", expected, result)
	}
	if !strings.Contains(result, "&x-tos-save-object=") {
		t.Errorf("expected save object in result: %q", result)
	}
	if !strings.Contains(result, "&x-tos-save-bucket=") {
		t.Errorf("expected save bucket in result: %q", result)
	}
}

func TestPostDataProcessHelper_VideoPCMWithoutSaveAs(t *testing.T) {
	ctx := context.Background()
	sampleRate := 16000
	channels := 1
	params := PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			PCMParams: &VideoPCMParams{
				SampleRate: &sampleRate,
				Channels:   &channels,
			},
		},
	}

	result, err := PostDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "video/convert,f_pcm,acodec_pcm_s16le,ac_1,ar_16000,vn_1"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestPostDataProcessHelper_VideoConvert(t *testing.T) {
	ctx := context.Background()
	start := 1000
	duration := 60000
	width := 1920
	height := 1080
	videoBitRate := 2000000
	maxRate := 2500000
	bufferSize := 5000000
	crf := 23
	removeAudio := 0
	frameInterval := 40

	result, err := PostDataProcessHelper(ctx, PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			ConvertParams: &VideoConvertParams{
				Format:       "mp4",
				StartTime:    &start,
				Duration:     &duration,
				VideoCodec:   "h264",
				Width:        &width,
				Height:       &height,
				VideoBitRate: &videoBitRate,
				MaxRate:      &maxRate,
				BufferSize:   &bufferSize,
				CRF:          &crf,
				RemoveAudio:  &removeAudio,
				AIGCMetadata: &AIGCMetadata{Label: "1"},
				Watermarks: []VideoWatermark{{
					Type:    "Text",
					Pos:     "TopRight",
					LocMode: "Relative",
					Text:    &VideoWatermarkText{Text: "test watermark"},
				}},
				BlindWatermark: &VideoDigitalWatermark{
					Type:          "Text",
					Version:       "V1",
					Message:       "ATIAAa6bqATuSQ-_394AAAAA",
					FrameInterval: &frameInterval,
				},
			},
		},
		SaveAsParams: &SaveAsParams{
			SaveBucket: "target-bucket",
			SaveObject: "output.mp4",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(result, "video/convert,f_mp4,ss_1000,t_60000,vcodec_h264,w_1920,h_1080,vb_2000000,maxrate_2500000,bufsize_5000000,crf_23,an_0,aigc_") {
		t.Fatalf("unexpected convert process: %s", result)
	}
	if !strings.Contains(result, "&x-tos-save-object=b3V0cHV0Lm1wNA") {
		t.Errorf("expected unpadded save object in result: %s", result)
	}
	if !strings.Contains(result, "&x-tos-save-bucket=dGFyZ2V0LWJ1Y2tldA") {
		t.Errorf("expected unpadded save bucket in result: %s", result)
	}
	if strings.Contains(result, "aigc_eyJMYWJlbCI6IjEifQ==") {
		t.Errorf("expected URL-safe base64 without padding: %s", result)
	}
	if !strings.Contains(result, "watermark_") || !strings.Contains(result, "blindwatermark_") {
		t.Errorf("expected watermark parameters in result: %s", result)
	}
}

func TestPostDataProcessHelper_VideoRemux(t *testing.T) {
	ctx := context.Background()
	segmentDuration := 5000
	streamIndex := 1

	result, err := PostDataProcessHelper(ctx, PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			RemuxParams: &VideoRemuxParams{
				Format:             "mp4",
				HLSSegmentDuration: &segmentDuration,
				StreamIndex:        &streamIndex,
				C2PAMetadata: &C2PAMetadata{
					AppID: "3006",
					Manifest: map[string]interface{}{
						"claim_generator": "TestApp/1.0",
					},
				},
			},
		},
		SaveAsParams: &SaveAsParams{SaveObject: "remux/output.mp4"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, "video/remux,f_mp4,st_5000,ti_1,c2pa_") {
		t.Fatalf("unexpected remux process: %s", result)
	}
	if !strings.Contains(result, "&x-tos-save-object=cmVtdXgvb3V0cHV0Lm1wNA") {
		t.Errorf("expected unpadded save object in result: %s", result)
	}
}

func TestPostDataProcessHelper_VideoLegacyTranscodeRejected(t *testing.T) {
	_, err := PostDataProcessHelper(context.Background(), PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			TranscodeParams: &VideoTranscodeParams{Tag: "Transcode", Name: "legacy-transcode"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "use ConvertParams for synchronous processing") {
		t.Fatalf("expected legacy Transcode rejection, got %v", err)
	}
}

func TestPostDataProcessHelper_VideoConvertValidation(t *testing.T) {
	ctx := context.Background()
	width := 127
	_, err := PostDataProcessHelper(ctx, PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			ConvertParams: &VideoConvertParams{
				Format: "mp4",
				Width:  &width,
			},
		},
		SaveAsParams: &SaveAsParams{SaveObject: "output.mp4"},
	})
	if err == nil || !strings.Contains(err.Error(), "Width") {
		t.Fatalf("expected width validation error, got %v", err)
	}
}

func TestPostDataProcessHelper_VideoMultipleOpsError(t *testing.T) {
	ctx := context.Background()
	params := PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			SnapshotParams: &VideoSnapshotParams{T: 1000},
			PCMParams:      &VideoPCMParams{},
		},
	}

	_, err := PostDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error when multiple video operations are specified")
	}

	_, err = PostDataProcessHelper(ctx, PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &VideoProcessParams{
			ConvertParams: &VideoConvertParams{Format: "mp4"},
			RemuxParams:   &VideoRemuxParams{Format: "mp4"},
		},
		SaveAsParams: &SaveAsParams{SaveObject: "output.mp4"},
	})
	if err == nil {
		t.Fatal("expected error when convert and remux are both specified")
	}
}

func TestPostDPOutputUnmarshalPCM(t *testing.T) {
	output := PostDPOutput{ImageProcessOutput: &ImageProcessOutput{Bucket: "stale-image-bucket"}}
	err := json.Unmarshal([]byte(`{"bucket":"target-bucket","object":"audio/output.pcm","status":"OK"}`), &output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.VideoProcessOutput == nil {
		t.Fatal("expected VideoProcessOutput to be populated")
	}
	if output.ImageProcessOutput != nil {
		t.Fatal("video response must clear stale ImageProcessOutput")
	}
	if output.VideoProcessOutput.PcmBucket != "target-bucket" {
		t.Fatalf("unexpected bucket: %s", output.VideoProcessOutput.PcmBucket)
	}
	if output.VideoProcessOutput.PcmObject != "audio/output.pcm" {
		t.Fatalf("unexpected object: %s", output.VideoProcessOutput.PcmObject)
	}
	if output.VideoProcessOutput.PcmStatus != "OK" {
		t.Fatalf("unexpected status: %s", output.VideoProcessOutput.PcmStatus)
	}
}

func TestPostDataProcessHelper_ImageCombinedWithSaveAs(t *testing.T) {
	result, err := PostDataProcessHelper(context.Background(), PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation:       enum.ImageOperationWatermark,
				WatermarkParams: &ImageWatermarkParams{Text: "dG9zIHRlc3Q"},
			},
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &ImageBlindWatermarkParams{
					Text: "dG9zdGVzdDEyMw", Version: intPtr(2), Level: intPtr(2),
				},
			},
			{
				Operation: enum.ImageOperationSetAIGCMetadata,
				AIGCMetadataParams: &ImageAIGCMetadataParams{
					Label:             "dGVzdF9sYWJlbA",
					ContentProducer:   "dGVzdF9wcm9kdWNlcg",
					ProduceID:         "dGVzdF9wcm9kdWNlcl9pZA",
					ContentPropagator: "dGVzdF9wcm9wYWdhdG9y",
					PropagateID:       "dGVzdF9wcm9wYWdhdG9yX2lk",
					ReservedCode1:     "dGVzdF9yZXNlcnZlZF9jb2RlXzE",
					ReservedCode2:     "dGVzdF9yZXNlcnZlZF9jb2RlXzI",
				},
			},
			{
				Operation: enum.ImageOperationSetC2PAMetadata,
				SetC2PAMetadataParams: &ImageSetC2PAMetadataParams{
					AppID:    "tos-functional-test",
					Manifest: "eyJjbGFpbSI6InBvc3QtaW1hZ2UtcHJvY2VzcyJ9",
				},
			},
		},
		SaveAsParams: &SaveAsParams{
			SaveBucket: "target-bucket",
			SaveObject: "post/image/combined.jpg",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "image/watermark,text_dG9zIHRlc3Q/" +
		"blindwatermark,text_dG9zdGVzdDEyMw,version_2,level_2/" +
		"setaigcmetadata,Label_dGVzdF9sYWJlbA,ContentProducer_dGVzdF9wcm9kdWNlcg,ProduceID_dGVzdF9wcm9kdWNlcl9pZA," +
		"ContentPropagator_dGVzdF9wcm9wYWdhdG9y,PropagateID_dGVzdF9wcm9wYWdhdG9yX2lk," +
		"ReservedCode1_dGVzdF9yZXNlcnZlZF9jb2RlXzE,ReservedCode2_dGVzdF9yZXNlcnZlZF9jb2RlXzI/" +
		"setc2pametadata,AppID_tos-functional-test,Manifest_eyJjbGFpbSI6InBvc3QtaW1hZ2UtcHJvY2VzcyJ9" +
		"&x-tos-save-object=cG9zdC9pbWFnZS9jb21iaW5lZC5qcGc" +
		"&x-tos-save-bucket=dGFyZ2V0LWJ1Y2tldA"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestPostDataProcessHelper_ImageSaveAsRequiresObject(t *testing.T) {
	_, err := PostDataProcessHelper(context.Background(), PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{{
			Operation:    enum.ImageOperationResize,
			ResizeParams: &ImageResizeParams{W: intPtr(50)},
		}},
		SaveAsParams: &SaveAsParams{SaveBucket: "target-bucket"},
	})
	if err == nil || !strings.Contains(err.Error(), "SaveAsParams.SaveObject") {
		t.Fatalf("expected required SaveObject error, got %v", err)
	}
}

func TestPostDataProcessAsyncHelper_RequiresJobFields(t *testing.T) {
	_, err := PostDataProcessAsyncHelper(context.Background(), PostDataProcessAsyncParams{})
	if err == nil || !strings.Contains(err.Error(), "JobType is required") {
		t.Fatalf("expected JobType error, got %v", err)
	}

	_, err = PostDataProcessAsyncHelper(context.Background(), PostDataProcessAsyncParams{
		JobType: ProcessJobTypeTranscode,
	})
	if err == nil || !strings.Contains(err.Error(), "JobBody is required") {
		t.Fatalf("expected JobBody error, got %v", err)
	}
}

func TestPostDataProcessAsyncHelper_AudioConvertJSON(t *testing.T) {
	result, err := PostDataProcessAsyncHelper(context.Background(), PostDataProcessAsyncParams{
		JobType: ProcessJobTypeAudioConvert,
		JobBody: &AudioConvertJobBody{
			Input:              ProcessJobInput{Object: "audio/input.mp3"},
			Output:             ProcessJobOutput{Region: "cn-beijing", Bucket: "target-bucket", Object: "audio/output.wav"},
			AudioConvertConfig: AudioConvertJobConfig{ContainerFormat: "wav"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "x-tos-async-process=") {
		t.Fatalf("async helper must return JSON, got %s", result)
	}
	var body struct {
		Input              ProcessJobInput       `json:"Input"`
		Output             ProcessJobOutput      `json:"Output"`
		AudioConvertConfig AudioConvertJobConfig `json:"AudioConvertConfig"`
	}
	if err := json.Unmarshal([]byte(result), &body); err != nil {
		t.Fatalf("async helper result is not JSON: %v", err)
	}
	if body.Input.Object != "audio/input.mp3" || body.Output.Object != "audio/output.wav" ||
		body.AudioConvertConfig.ContainerFormat != "wav" {
		t.Fatalf("unexpected audio convert JSON: %+v", body)
	}
}

func TestPostDataProcessAsyncHelper_AudioConcatJSON(t *testing.T) {
	result, err := PostDataProcessAsyncHelper(context.Background(), PostDataProcessAsyncParams{
		JobType: ProcessJobTypeAudioConcat,
		JobBody: &AudioConcatJobBody{
			Input: AudioConcatInput{
				Object:       "audio/input.mp3",
				PreFragments: []AudioConcatPreFragment{{Object: "audio/pre.mp3"}},
			},
			Output:            ProcessJobOutput{Region: "cn-beijing", Bucket: "target-bucket", Object: "audio/output.mp3"},
			AudioConcatConfig: AudioConcatConfig{ContainerFormat: "mp3"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(result, "x-tos-async-process=") {
		t.Fatalf("async helper must return JSON, got %s", result)
	}
	var body AudioConcatJobBody
	if err := json.Unmarshal([]byte(result), &body); err != nil {
		t.Fatalf("async helper result is not JSON: %v", err)
	}
	if body.Input.Object != "audio/input.mp3" || body.Output.Object != "audio/output.mp3" ||
		body.AudioConcatConfig.ContainerFormat != "mp3" || len(body.Input.PreFragments) != 1 {
		t.Fatalf("unexpected audio concat JSON: %+v", body)
	}
}

func TestDocConvertJobBody_ToPDF(t *testing.T) {
	body := &DocConvertJobBody{
		Input: ProcessDocConvertInput{Key: "input.docx"},
		DocConvertConfig: ProcessDocConvertConfig{
			SrcType:   "docx",
			TgtType:   "pdf",
			StartPage: 1,
			EndPage:   -1,
		},
		Output: ProcessDocConvertOutput{
			Region: "cn-beijing",
			Bucket: "dst-bucket",
			Object: "output.pdf",
		},
	}
	if body.Input.Key != "input.docx" {
		t.Fatalf("unexpected source key: %q", body.Input.Key)
	}
	if body.DocConvertConfig.TgtType != "pdf" {
		t.Fatalf("unexpected target type: %q", body.DocConvertConfig.TgtType)
	}
	if body.Output.Bucket != "dst-bucket" || body.Output.Object != "output.pdf" {
		t.Fatalf("unexpected output: %+v", body.Output)
	}
	if body.Output.Region != "cn-beijing" {
		t.Fatalf("unexpected output region: %q", body.Output.Region)
	}
	if body.DocConvertConfig.StartPage != 1 || body.DocConvertConfig.EndPage != -1 {
		t.Fatalf("unexpected default page range: %+v", body.DocConvertConfig)
	}
}

func TestDocConvertJobBody_ToImage(t *testing.T) {
	body := &DocConvertJobBody{
		Input: ProcessDocConvertInput{Key: "input.pdf"},
		DocConvertConfig: ProcessDocConvertConfig{
			SrcType:   "pdf",
			TgtType:   "jpg",
			StartPage: 1,
			EndPage:   3,
		},
		Output: ProcessDocConvertOutput{
			Region: "cn-beijing",
			Bucket: "dst-bucket",
			Object: "output{Page}.jpg",
		},
	}
	if body.DocConvertConfig.TgtType != "jpg" {
		t.Fatalf("unexpected target type: %q", body.DocConvertConfig.TgtType)
	}
	if body.DocConvertConfig.StartPage != 1 || body.DocConvertConfig.EndPage != 3 {
		t.Fatalf("unexpected page range: %+v", body.DocConvertConfig)
	}
}

func TestGetDataProcessHelper_ImageMosaic(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation: enum.ImageOperationMosaic,
				MosaicParams: &ImageMosaicParams{
					G: "center",
					W: intPtr(100),
					H: intPtr(100),
					T: "blur",
					R: intPtr(10),
					S: intPtr(20),
				},
			},
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "mosaic") {
		t.Errorf("expected mosaic in result: %q", result)
	}
	if !strings.Contains(result, "g_center") {
		t.Errorf("expected g_center in result: %q", result)
	}
	if !strings.Contains(result, "t_blur") {
		t.Errorf("expected t_blur in result: %q", result)
	}
}

func TestPipelineConcatenation(t *testing.T) {
	ctx := context.Background()
	p1Params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &ImageResizeParams{W: intPtr(100)},
			},
		},
	}
	p2Params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []ImageProcessParams{
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &ImageFormatParams{Format: "webp"},
			},
		},
	}

	p1, err := GetDataProcessHelper(ctx, p1Params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p2, err := GetDataProcessHelper(ctx, p2Params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	combined := p1 + "|" + p2
	expected := "image/resize,w_100|image/format,webp"
	if combined != expected {
		t.Errorf("expected %q, got %q", expected, combined)
	}
}

// ==================== MCAP 测试 ====================

func TestGetDataProcessHelper_McapInfo(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &McapProcessParams{
			Operation: McapOperationInfo,
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "file/mcap-info"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_McapDoctor(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &McapProcessParams{
			Operation: McapOperationDoctor,
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "file/mcap-doctor"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_McapListChannels(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &McapProcessParams{
			Operation:    McapOperationList,
			ListResource: McapListChannels,
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "file/mcap-list,channels"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestGetDataProcessHelper_McapListAllResources(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name     string
		resource McapListResource
		expected string
	}{
		{"attachments", McapListAttachments, "file/mcap-list,attachments"},
		{"channels", McapListChannels, "file/mcap-list,channels"},
		{"chunks", McapListChunks, "file/mcap-list,chunks"},
		{"metadata", McapListMetadata, "file/mcap-list,metadata"},
		{"schemas", McapListSchemas, "file/mcap-list,schemas"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := GetDataProcessParams{
				GetProcessType: enum.GetProcessTypeFile,
				McapProcessParams: &McapProcessParams{
					Operation:    McapOperationList,
					ListResource: tc.resource,
				},
			}
			result, err := GetDataProcessHelper(ctx, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGetDataProcessHelper_McapListMissingResource(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &McapProcessParams{
			Operation: McapOperationList,
			// ListResource 未设置
		},
	}

	_, err := GetDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error when ListResource is missing for mcap-list")
	}
}

func TestGetDataProcessHelper_McapNilParams(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		// McapProcessParams 未设置
	}

	_, err := GetDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error when McapProcessParams is nil")
	}
}

// ==================== 点云压缩测试 ====================

func TestGetDataProcessHelper_PointCloudCompress(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypePointCloud,
		PointCloudProcessParams: &PointCloudCompressParams{
			PointResolution:  float64Ptr(0.05),
			OctreeResolution: float64Ptr(0.02),
		},
	}

	result, err := GetDataProcessHelper(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "pointcloud/compress"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestPointCloudCompressParams_QueryParamsDefaults(t *testing.T) {
	params := &PointCloudCompressParams{}
	q := params.QueryParams()

	if q["format"] != "pcd" {
		t.Errorf("expected format=pcd, got %q", q["format"])
	}
	if q["method"] != "octree" {
		t.Errorf("expected method=octree, got %q", q["method"])
	}
	if q["fields"] != "eHI6" {
		t.Errorf("expected fields=eHI6, got %q", q["fields"])
	}
	if q["lib"] != "pcl" {
		t.Errorf("expected lib=pcl, got %q", q["lib"])
	}
	if q["point-resolution"] != "0.01" {
		t.Errorf("expected point-resolution=0.01, got %q", q["point-resolution"])
	}
	if q["octree-resolution"] != "0.01" {
		t.Errorf("expected octree-resolution=0.01, got %q", q["octree-resolution"])
	}
	if q["down-sampling"] != "1" {
		t.Errorf("expected down-sampling=1, got %q", q["down-sampling"])
	}
}

func TestPointCloudCompressParams_QueryParamsCustom(t *testing.T) {
	params := &PointCloudCompressParams{
		Format:           "pcd",
		Method:           "octree",
		Fields:           "eHI6",
		Lib:              "pcl",
		PointResolution:  float64Ptr(0.05),
		OctreeResolution: float64Ptr(0.02),
		DownSampling:     boolPtr(false),
	}
	q := params.QueryParams()

	if q["point-resolution"] != "0.05" {
		t.Errorf("expected point-resolution=0.05, got %q", q["point-resolution"])
	}
	if q["octree-resolution"] != "0.02" {
		t.Errorf("expected octree-resolution=0.02, got %q", q["octree-resolution"])
	}
	if q["down-sampling"] != "0" {
		t.Errorf("expected down-sampling=0, got %q", q["down-sampling"])
	}
}

func TestGetDataProcessHelper_PointCloudNilParams(t *testing.T) {
	ctx := context.Background()
	params := GetDataProcessParams{
		GetProcessType: enum.GetProcessTypePointCloud,
		// PointCloudProcessParams 未设置
	}

	_, err := GetDataProcessHelper(ctx, params)
	if err == nil {
		t.Fatal("expected error when PointCloudProcessParams is nil")
	}
}

func float64Ptr(v float64) *float64 { return &v }

// ==================== FileCompress / FileUncompress Helper 测试 ====================

func TestFileCompressJobBody_KeyConfig(t *testing.T) {
	body := &FileCompressJobBody{
		Input: FileCompressInput{
			KeyConfig: []FileCompressKeyConfig{
				{Key: "input/obj1"},
				{Key: "input/obj2"},
			},
		},
		FileCompressConfig: FileCompressConfig{
			Format:  "zip",
			Flatten: 0,
		},
		Output: FileJobOutput{
			Region: "cn-beijing",
			Bucket: "testbkt",
			Object: "output/test.zip",
		},
	}

	if body.FileCompressConfig.Format != "zip" {
		t.Errorf("expected Format 'zip', got %q", body.FileCompressConfig.Format)
	}
	if body.FileCompressConfig.Flatten != 0 {
		t.Errorf("expected Flatten 0, got %d", body.FileCompressConfig.Flatten)
	}
	if len(body.Input.KeyConfig) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(body.Input.KeyConfig))
	}
	if body.Input.KeyConfig[0].Key != "input/obj1" {
		t.Errorf("expected key 'input/obj1', got %q", body.Input.KeyConfig[0].Key)
	}
	if body.Output.Object != "output/test.zip" {
		t.Errorf("expected output object 'output/test.zip', got %q", body.Output.Object)
	}
}

func TestFileCompressJobBody_Prefix(t *testing.T) {
	body := &FileCompressJobBody{
		Input: FileCompressInput{
			Prefix: "test/",
		},
		FileCompressConfig: FileCompressConfig{
			Format:  "zip",
			Flatten: 2,
		},
		Output: FileJobOutput{
			Region: "cn-beijing",
			Bucket: "testbkt",
			Object: "output/test.zip",
		},
	}

	if body.Input.Prefix != "test/" {
		t.Errorf("expected Prefix 'test/', got %q", body.Input.Prefix)
	}
	if body.FileCompressConfig.Flatten != 2 {
		t.Errorf("expected Flatten 2, got %d", body.FileCompressConfig.Flatten)
	}
}

func TestFileUncompressJobBody(t *testing.T) {
	body := &FileUncompressJobBody{
		Input: FileUncompressInput{
			Object: "input/test.zip",
		},
		FileUncompressConfig: FileUncompressConfig{
			Prefix:         "output/",
			PrefixReplaced: 1,
		},
		Output: FileJobOutput{
			Region: "cn-beijing",
			Bucket: "testbkt",
		},
	}

	if body.Input.Object != "input/test.zip" {
		t.Errorf("expected input object 'input/test.zip', got %q", body.Input.Object)
	}
	if body.FileUncompressConfig.Prefix != "output/" {
		t.Errorf("expected Prefix 'output/', got %q", body.FileUncompressConfig.Prefix)
	}
	if body.FileUncompressConfig.PrefixReplaced != 1 {
		t.Errorf("expected PrefixReplaced 1, got %d", body.FileUncompressConfig.PrefixReplaced)
	}
	if body.Output.Region != "cn-beijing" {
		t.Errorf("expected Region 'cn-beijing', got %q", body.Output.Region)
	}
}

func TestFileCompressJobBody_JSONMarshal(t *testing.T) {
	body := FileCompressJobBody{
		Input: FileCompressInput{
			KeyConfig: []FileCompressKeyConfig{
				{Key: "a.txt"},
			},
		},
		FileCompressConfig: FileCompressConfig{
			Format:  "zip",
			Flatten: 0,
		},
		Output: FileJobOutput{
			Region: "cn-beijing",
			Bucket: "mybucket",
			Object: "out.zip",
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]interface{}
	if err = json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if parsed["Input"] == nil {
		t.Error("expected 'Input' field in JSON")
	}
	if parsed["FileCompressConfig"] == nil {
		t.Error("expected 'FileCompressConfig' field in JSON")
	}
	if parsed["Output"] == nil {
		t.Error("expected 'Output' field in JSON")
	}
}

func TestFileUncompressJobBody_JSONMarshal(t *testing.T) {
	body := FileUncompressJobBody{
		Input: FileUncompressInput{
			Object: "archive.zip",
		},
		FileUncompressConfig: FileUncompressConfig{
			Prefix:         "extracted/",
			PrefixReplaced: 0,
		},
		Output: FileJobOutput{
			Region: "cn-beijing",
			Bucket: "mybucket",
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]interface{}
	if err = json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	if parsed["Input"] == nil {
		t.Error("expected 'Input' field in JSON")
	}
	if parsed["FileUncompressConfig"] == nil {
		t.Error("expected 'FileUncompressConfig' field in JSON")
	}
	if parsed["Output"] == nil {
		t.Error("expected 'Output' field in JSON")
	}
}
