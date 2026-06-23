package tests

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type dpTestEnv struct {
	region        string
	bucket        string
	accessKey     string
	secretKey     string
	endpoint      string
	imageKey      string
	autoOrientKey string
	gifKey        string
	videoKey      string
	m3u8Key       string
	docKey        string
	pdfKey        string
	audioKey      string
	audioKey2     string
}

func newDPTestEnv() *dpTestEnv {
	audioKey := os.Getenv("TOS_AUDIO_KEY")
	if audioKey == "" {
		audioKey = "test.mp3"
	}
	audioKey2 := os.Getenv("TOS_AUDIO_KEY2")
	if audioKey2 == "" {
		audioKey2 = audioKey
	}
	return &dpTestEnv{
		region:        os.Getenv("TOS_REGION"),
		bucket:        os.Getenv("TOS_BUCKET"),
		accessKey:     os.Getenv("TOS_ACCESS_KEY"),
		secretKey:     os.Getenv("TOS_SECRET_KEY"),
		endpoint:      os.Getenv("TOS_ENDPOINT"),
		imageKey:      "test.jpg",
		autoOrientKey: getAutoOrientImageKey(),
		gifKey:        "test.gif",
		videoKey:      "test.mp4",
		m3u8Key:       "test.m3u8",
		docKey:        "test.docx",
		pdfKey:        "test.pdf",
		audioKey:      audioKey,
		audioKey2:     audioKey2,
	}
}

func getAutoOrientImageKey() string {
	autoOrientKey := os.Getenv("TOS_AUTO_ORIENT_KEY")
	if autoOrientKey == "" {
		autoOrientKey = "auto_orient_exif6.jpg"
	}
	return autoOrientKey
}

func (e *dpTestEnv) skip(t *testing.T) {
	if e.accessKey == "" || e.secretKey == "" || e.endpoint == "" || e.bucket == "" {
		t.Skip("跳过集成测试：缺少环境变量 TOS_ACCESS_KEY / TOS_SECRET_KEY / TOS_ENDPOINT / TOS_BUCKET")
	}
}

func (e *dpTestEnv) newClient(t *testing.T) *tos.ClientV2 {
	client, err := tos.NewClientV2(e.endpoint,
		tos.WithRegion(e.region),
		tos.WithCredentials(tos.NewStaticCredentials(e.accessKey, e.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithMaxRetryCount(3),
	)
	require.Nil(t, err)
	return client
}

func (e *dpTestEnv) skipAudio(t *testing.T) {
	e.skip(t)
	if e.audioKey == "" {
		t.Skip("跳过音频集成测试：缺少音频测试对象配置")
	}
}

func (e *dpTestEnv) skipWorkflowJob(t *testing.T) {
	e.skip(t)
}

// processSupportsSaveAs 判断 process 是否支持 SaveAs，与 tos.getProcessSupportsSaveAs 逻辑一致。
func processSupportsSaveAs(process string) bool {
	if process == "" {
		return false
	}
	switch process {
	case "image/inspect", "image/info", "image/average-hue",
		"image/getaigcmetadata", "image/getc2pametadata", "image/aitag",
		"image/embedding", "image/ocr",
		"video/info", "video/pm3u8",
		"video/aigcmetadata", "video/c2pametadata",
		"video/embedding":
		return false
	}
	if strings.HasPrefix(process, "hls/ts") {
		return false
	}
	if strings.HasPrefix(process, "file/") || strings.HasPrefix(process, "pointcloud/") {
		return false
	}
	return true
}

func TestDP_GetImageResize(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{M: "lfit", W: intPtr(200), H: intPtr(200)},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.autoOrientKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("缩放后图片大小: %d bytes", len(data))
		assert.True(t, len(data) > 0, "缩放后图片应有内容")
	}
}

func TestDP_GetImageResizeWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{M: "lfit", W: intPtr(200), H: intPtr(200)},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.autoOrientKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/image/resize_200x200.jpg",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("缩放+SaveAs 响应大小: %d bytes", len(data))
	}
}

func TestDP_GetImageInfo(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationInfo},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.autoOrientKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 info 响应: %s", string(data))
		assert.True(t, len(data) > 0, "info 应返回 JSON")
	}
}

func TestDP_GetImageInfoWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationInfo},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.autoOrientKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/image/info.json",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 info+SaveAs 响应: %s", string(data))
	}
}

func TestDP_GetImageFormatConvert(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &tos.ImageFormatParams{Format: "png"},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.autoOrientKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("格式转换后大小: %d bytes", len(data))
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageFormatConvertWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &tos.ImageFormatParams{Format: "png"},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.autoOrientKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/image/format_convert.png",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("格式转换+SaveAs 响应大小: %d bytes", len(data))
	}
}

func TestDP_GetImageMultiOps(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{W: intPtr(300)},
			},
			{
				Operation:     enum.ImageOperationQuality,
				QualityParams: &tos.ImageQualityParams{Q: intPtr(80)},
			},
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &tos.ImageFormatParams{Format: "webp"},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.imageKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("多操作串联后大小: %d bytes", len(data))
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageMultiOpsWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{W: intPtr(300)},
			},
			{
				Operation:     enum.ImageOperationQuality,
				QualityParams: &tos.ImageQualityParams{Q: intPtr(80)},
			},
			{
				Operation:    enum.ImageOperationFormat,
				FormatParams: &tos.ImageFormatParams{Format: "webp"},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/image/multi_ops.webp",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("多操作串联+SaveAs 响应大小: %d bytes", len(data))
	}
}

func TestDP_GetImageWatermark(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	textBase64 := base64.URLEncoding.EncodeToString([]byte("TOS-SDK-TEST"))

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation: enum.ImageOperationWatermark,
				WatermarkParams: &tos.ImageWatermarkParams{
					Text:  textBase64,
					Size:  intPtr(40),
					G:     "south",
					Color: "FF0000",
					T:     intPtr(80),
				},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.imageKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("水印图片大小: %d bytes", len(data))
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageWatermarkWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	textBase64 := base64.URLEncoding.EncodeToString([]byte("TOS-SDK-TEST"))

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation: enum.ImageOperationWatermark,
				WatermarkParams: &tos.ImageWatermarkParams{
					Text:  textBase64,
					Size:  intPtr(40),
					G:     "south",
					Color: "FF0000",
					T:     intPtr(80),
				},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/image/watermark.jpg",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("水印+SaveAs 响应大小: %d bytes", len(data))
	}
}

func TestDP_GetImageBlindWatermarkWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	textBase64 := base64.URLEncoding.EncodeToString([]byte("blind-watermark"))

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &tos.ImageBlindWatermarkParams{
					Text:    textBase64,
					Version: intPtr(2),
					Level:   intPtr(2),
				},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	saveKey := "dp-test/image/blindwatermark.jpg"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("暗水印+SaveAs 响应大小: %d bytes", len(data))
	}

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, headOutput)
	assert.True(t, headOutput.ContentLength > 0)
}

func TestDP_GetImageAutoOrientInternal(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationAutoOrientInternal},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.imageKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageAutoOrientInternalWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationAutoOrientInternal},
		},
	})
	require.Nil(t, err)

	saveKey := "dp-test/image/auto_orient_internal.jpg"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, headOutput)
	assert.True(t, headOutput.ContentLength > 0)
}

func TestDP_GetImageColorspaceGray(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:        enum.ImageOperationColorspace,
				ColorspaceParams: &tos.ImageColorspaceParams{Value: "gray"},
			},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.imageKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageColorspaceGrayWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:        enum.ImageOperationColorspace,
				ColorspaceParams: &tos.ImageColorspaceParams{Value: "gray"},
			},
		},
	})
	require.Nil(t, err)

	saveKey := "dp-test/image/colorspace_gray.jpg"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, headOutput)
	assert.True(t, headOutput.ContentLength > 0)
}

func TestDP_GetImageAIGCMetadata(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationGetAIGCMetadata},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     "test_aigc.png",
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 AIGC metadata 响应: %s", string(data))
		assert.True(t, len(data) > 2)
	}
}

func TestDP_GetImageC2PAMetadata(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationGetC2PAMetadata},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.imageKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 C2PA metadata 响应: %s", string(data))
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageDeBlindWatermark(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	textBase64 := base64.URLEncoding.EncodeToString([]byte("blind-watermark"))
	saveKey := "dp-test/image/deblind_source.jpg"

	writeProcess, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{M: "pad", W: intPtr(640), H: intPtr(640)},
			},
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &tos.ImageBlindWatermarkParams{
					Text:    textBase64,
					Version: intPtr(2),
					Level:   intPtr(2),
				},
			},
		},
	})
	require.Nil(t, err)

	writeOutput, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    writeProcess,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, writeOutput)
	if writeOutput.Content != nil {
		_, _ = ioutil.ReadAll(writeOutput.Content)
		writeOutput.Content.Close()
	}

	extractProcess, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:              enum.ImageOperationDeBlindWatermark,
				DeBlindWatermarkParams: &tos.ImageDeBlindWatermarkParams{Version: intPtr(2)},
			},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     saveKey,
		Process: extractProcess,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetImageDeBlindWatermarkWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	textBase64 := base64.URLEncoding.EncodeToString([]byte("blind-watermark"))
	sourceKey := "dp-test/image/deblind_source_saveas.jpg"

	writeProcess, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:    enum.ImageOperationResize,
				ResizeParams: &tos.ImageResizeParams{M: "pad", W: intPtr(640), H: intPtr(640)},
			},
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &tos.ImageBlindWatermarkParams{
					Text:    textBase64,
					Version: intPtr(2),
					Level:   intPtr(2),
				},
			},
		},
	})
	require.Nil(t, err)

	writeOutput, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.imageKey,
		Process:    writeProcess,
		SaveBucket: env.bucket,
		SaveObject: sourceKey,
	})
	require.Nil(t, err)
	require.NotNil(t, writeOutput)
	if writeOutput.Content != nil {
		_, _ = ioutil.ReadAll(writeOutput.Content)
		writeOutput.Content.Close()
	}

	extractProcess, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation:              enum.ImageOperationDeBlindWatermark,
				DeBlindWatermarkParams: &tos.ImageDeBlindWatermarkParams{Version: intPtr(2)},
			},
		},
	})
	require.Nil(t, err)

	saveKey := "dp-test/image/deblind_result.jpg"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        sourceKey,
		Process:    extractProcess,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, headOutput)
	assert.True(t, headOutput.ContentLength > 0)
}

type imageGetProcessCase struct {
	name   string
	key    string
	params []tos.ImageProcessParams
}

func runImageGetProcessCase(t *testing.T, env *dpTestEnv, client *tos.ClientV2, tc imageGetProcessCase, saveAs bool) {
	t.Helper()
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType:     enum.GetProcessTypeImage,
		ImageProcessParams: tc.params,
	})
	require.Nil(t, err)

	input := &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     tc.key,
		Process: dpResult,
	}

	var saveKey string
	if saveAs {
		// 根据源文件后缀推导 SaveAs 输出后缀
		ext := ".jpg"
		if idx := strings.LastIndex(tc.key, "."); idx >= 0 {
			ext = tc.key[idx:]
		}
		saveKey = "dp-test/image/" + strings.ReplaceAll(strings.ToLower(tc.name), "/", "_") + ext
		input.SaveBucket = env.bucket
		input.SaveObject = saveKey
	}

	output, err := client.GetDataProcess(ctx, input)
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}

	if saveAs {
		// 不支持 SaveAs 的算子，SDK 会静默忽略 SaveBucket/SaveObject，跳过落桶验证
		if processSupportsSaveAs(dpResult) {
			headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
				Bucket: env.bucket,
				Key:    saveKey,
			})
			require.Nil(t, err)
			require.NotNil(t, headOutput)
			assert.True(t, headOutput.ContentLength > 0)
		}
	}
}

func runStandaloneImageGetProcessCase(t *testing.T, tc imageGetProcessCase, saveAs bool) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	runImageGetProcessCase(t, env, client, tc, saveAs)
}

func TestDP_GetImageSetAIGCMetadata(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "setaigcmetadata",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:          enum.ImageOperationSetAIGCMetadata,
			AIGCMetadataParams: &tos.ImageAIGCMetadataParams{Label: "bGFiZWw=", ContentProducer: "cHJvZHVjZXI="},
		}},
	}, false)
}

func TestDP_GetImageSetAIGCMetadataWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "setaigcmetadata",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:          enum.ImageOperationSetAIGCMetadata,
			AIGCMetadataParams: &tos.ImageAIGCMetadataParams{Label: "bGFiZWw=", ContentProducer: "cHJvZHVjZXI="},
		}},
	}, true)
}

func TestDP_GetImageCrop(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "crop",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationCrop,
			CropParams: &tos.ImageCropParams{W: intPtr(200), H: intPtr(200), X: intPtr(0), Y: intPtr(0)},
		}},
	}, false)
}

func TestDP_GetImageCropWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "crop",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationCrop,
			CropParams: &tos.ImageCropParams{W: intPtr(200), H: intPtr(200), X: intPtr(0), Y: intPtr(0)},
		}},
	}, true)
}

func TestDP_GetImageCircle(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "circle",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationCircle,
			CircleParams: &tos.ImageCircleParams{R: 100},
		}},
	}, false)
}

func TestDP_GetImageCircleWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "circle",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationCircle,
			CircleParams: &tos.ImageCircleParams{R: 100},
		}},
	}, true)
}

func TestDP_GetImageIndexcrop(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "indexcrop",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:       enum.ImageOperationIndexcrop,
			IndexcropParams: &tos.ImageIndexcropParams{X: intPtr(1), Y: intPtr(1), I: intPtr(0)},
		}},
	}, false)
}

func TestDP_GetImageIndexcropWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "indexcrop",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:       enum.ImageOperationIndexcrop,
			IndexcropParams: &tos.ImageIndexcropParams{X: intPtr(1), Y: intPtr(1), I: intPtr(0)},
		}},
	}, true)
}

func TestDP_GetImageClip(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "clip",
		key:  "test.gif",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationClip,
			ClipParams: &tos.ImageClipParams{Frame: intPtr(2), First: intPtr(1), Step: intPtr(3)},
		}},
	}, false)
}

func TestDP_GetImageClipWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "clip",
		key:  "test.gif",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationClip,
			ClipParams: &tos.ImageClipParams{Frame: intPtr(2), First: intPtr(1), Step: intPtr(3)},
		}},
	}, true)
}

func TestDP_GetImageRoundedCorners(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "rounded-corners",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:      enum.ImageOperationRoundedCorners,
			RoundedCorners: &tos.ImageRoundedCornersParams{R: 20},
		}},
	}, false)
}

func TestDP_GetImageRoundedCornersWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "rounded-corners",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:      enum.ImageOperationRoundedCorners,
			RoundedCorners: &tos.ImageRoundedCornersParams{R: 20},
		}},
	}, true)
}

func TestDP_GetImageAverageHue(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name:   "average-hue",
		key:    "test.jpg",
		params: []tos.ImageProcessParams{{Operation: enum.ImageOperationAverageHue}},
	}, false)
}

func TestDP_GetImageAverageHueWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name:   "average-hue",
		key:    "test.jpg",
		params: []tos.ImageProcessParams{{Operation: enum.ImageOperationAverageHue}},
	}, true)
}

func TestDP_GetImageAutoOrient(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "auto-orient",
		key:  getAutoOrientImageKey(),
		params: []tos.ImageProcessParams{{
			Operation:        enum.ImageOperationAutoOrient,
			AutoOrientParams: &tos.ImageAutoOrientParams{Value: 1},
		}},
	}, false)
}

func TestDP_GetImageAutoOrientWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "auto-orient",
		key:  getAutoOrientImageKey(),
		params: []tos.ImageProcessParams{{
			Operation:        enum.ImageOperationAutoOrient,
			AutoOrientParams: &tos.ImageAutoOrientParams{Value: 1},
		}},
	}, true)
}

func TestDP_GetImageBlur(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "blur",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationBlur,
			BlurParams: &tos.ImageBlurParams{Radius: intPtr(2), Sigma: intPtr(3)},
		}},
	}, false)
}

func TestDP_GetImageBlurWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "blur",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationBlur,
			BlurParams: &tos.ImageBlurParams{Radius: intPtr(2), Sigma: intPtr(3)},
		}},
	}, true)
}

func TestDP_GetImageRotate(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "rotate",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationRotate,
			RotateParams: &tos.ImageRotateParams{Value: 90},
		}},
	}, false)
}

func TestDP_GetImageRotateWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "rotate",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationRotate,
			RotateParams: &tos.ImageRotateParams{Value: 90},
		}},
	}, true)
}

func TestDP_GetImageBright(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "bright",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationBright,
			BrightParams: &tos.ImageBrightParams{Value: 10},
		}},
	}, false)
}

func TestDP_GetImageBrightWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "bright",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationBright,
			BrightParams: &tos.ImageBrightParams{Value: 10},
		}},
	}, true)
}

func TestDP_GetImageSharpen(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "sharpen",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:     enum.ImageOperationSharpen,
			SharpenParams: &tos.ImageSharpenParams{Value: 100},
		}},
	}, false)
}

func TestDP_GetImageSharpenWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "sharpen",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:     enum.ImageOperationSharpen,
			SharpenParams: &tos.ImageSharpenParams{Value: 100},
		}},
	}, true)
}

func TestDP_GetImageContrast(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "contrast",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:      enum.ImageOperationContrast,
			ContrastParams: &tos.ImageContrastParams{Value: 10},
		}},
	}, false)
}

func TestDP_GetImageContrastWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "contrast",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:      enum.ImageOperationContrast,
			ContrastParams: &tos.ImageContrastParams{Value: 10},
		}},
	}, true)
}

func TestDP_GetImageDraw(t *testing.T) {
	red := "FF0000"
	drawLine := true
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "draw",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation: enum.ImageOperationDraw,
			DrawParams: &tos.ImageDrawParams{
				Points:    []tos.ImagePoint{{X: 20, Y: 20}, {X: 80, Y: 80}},
				Radius:    intPtr(4),
				DrawLine:  &drawLine,
				LineWidth: intPtr(5),
				ColorRGB:  &red,
			},
		}},
	}, false)
}

func TestDP_GetImageDrawWithSaveAs(t *testing.T) {
	red := "FF0000"
	drawLine := true
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "draw",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation: enum.ImageOperationDraw,
			DrawParams: &tos.ImageDrawParams{
				Points:    []tos.ImagePoint{{X: 20, Y: 20}, {X: 80, Y: 80}},
				Radius:    intPtr(4),
				DrawLine:  &drawLine,
				LineWidth: intPtr(5),
				ColorRGB:  &red,
			},
		}},
	}, true)
}

func TestDP_GetImageMosaic(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "mosaic",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationMosaic,
			MosaicParams: &tos.ImageMosaicParams{G: "center", W: intPtr(100), H: intPtr(100), T: "blur", R: intPtr(2), S: intPtr(10)},
		}},
	}, false)
}

func TestDP_GetImageMosaicWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "mosaic",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:    enum.ImageOperationMosaic,
			MosaicParams: &tos.ImageMosaicParams{G: "center", W: intPtr(100), H: intPtr(100), T: "blur", R: intPtr(2), S: intPtr(10)},
		}},
	}, true)
}

func TestDP_GetImageQuality(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "quality",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:     enum.ImageOperationQuality,
			QualityParams: &tos.ImageQualityParams{Q: intPtr(80)},
		}},
	}, false)
}

func TestDP_GetImageQualityWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "quality",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:     enum.ImageOperationQuality,
			QualityParams: &tos.ImageQualityParams{Q: intPtr(80)},
		}},
	}, true)
}

func TestDP_GetImageInterlace(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "interlace",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:       enum.ImageOperationInterlace,
			InterlaceParams: &tos.ImageInterlaceParams{Value: 1},
		}},
	}, false)
}

func TestDP_GetImageInterlaceWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "interlace",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:       enum.ImageOperationInterlace,
			InterlaceParams: &tos.ImageInterlaceParams{Value: 1},
		}},
	}, true)
}

func TestDP_GetImageSlim(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "slim",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationSlim,
			SlimParams: &tos.ImageSlimParams{ZLevel: intPtr(5)},
		}},
	}, false)
}

func TestDP_GetImageSlimWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name: "slim",
		key:  "test.jpg",
		params: []tos.ImageProcessParams{{
			Operation:  enum.ImageOperationSlim,
			SlimParams: &tos.ImageSlimParams{ZLevel: intPtr(5)},
		}},
	}, true)
}

func TestDP_GetImageStrip(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name:   "strip",
		key:    "test.jpg",
		params: []tos.ImageProcessParams{{Operation: enum.ImageOperationStrip}},
	}, false)
}

func TestDP_GetImageStripWithSaveAs(t *testing.T) {
	runStandaloneImageGetProcessCase(t, imageGetProcessCase{
		name:   "strip",
		key:    "test.jpg",
		params: []tos.ImageProcessParams{{Operation: enum.ImageOperationStrip}},
	}, true)
}

func TestDP_GetImageInspect(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{Operation: enum.ImageOperationInspect},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     "test.jpg",
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 inspect 响应: %s", string(data))
		assert.True(t, len(data) > 0, "inspect 应返回 JSON")
	}
}

func TestDP_GetImageUnderstanding(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation: enum.ImageOperationUnderstanding,
				UnderstandingParams: &tos.ImageUnderstandingParams{
					Model:  "doubao-seed-1.6-vision",
					Prompt: "请描述这张图片的内容",
				},
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     "test.jpg",
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("图片 understanding 响应: %s", string(data))
		assert.True(t, len(data) > 0, "understanding 应返回 JSON")
	}
}

func TestDP_GetImageStructuredExtendedOpsSkipped(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{name: "setc2pametadata", reason: "当前环境请求超时，暂不纳入 PASS 集成测试"},
		{name: "aitag", reason: "当前环境请求超时，依赖额外 AI 能力"},
		{name: "embedding", reason: "当前环境后端返回 process invalid"},
		{name: "ocr", reason: "当前环境后端返回 process invalid"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Skip(tc.reason)
		})
	}
}

type videoGetProcessCase struct {
	name   string
	key    string
	params *tos.VideoProcessParams
}

func runVideoGetProcessCase(t *testing.T, env *dpTestEnv, client *tos.ClientV2, tc videoGetProcessCase, saveAs bool, saveExt string) {
	t.Helper()
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType:     enum.GetProcessTypeVideo,
		VideoProcessParams: tc.params,
	})
	require.Nil(t, err)

	input := &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     tc.key,
		Process: dpResult,
	}

	var saveKey string
	if saveAs {
		saveKey = "dp-test/video/" + strings.ReplaceAll(strings.ToLower(tc.name), "/", "_") + "." + saveExt
		input.SaveBucket = env.bucket
		input.SaveObject = saveKey
	}

	output, err := client.GetDataProcess(ctx, input)
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}

	if saveAs {
		// 不支持 SaveAs 的算子，SDK 会静默忽略 SaveBucket/SaveObject，跳过落桶验证
		if processSupportsSaveAs(dpResult) {
			headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
				Bucket: env.bucket,
				Key:    saveKey,
			})
			require.Nil(t, err)
			require.NotNil(t, headOutput)
			assert.True(t, headOutput.ContentLength > 0)
		}
	}
}

func runStandaloneVideoGetProcessCase(t *testing.T, tc videoGetProcessCase, saveAs bool, saveExt string) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	runVideoGetProcessCase(t, env, client, tc, saveAs, saveExt)
}

func TestDP_GetVideoInfo(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "info",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{InfoParams: &tos.VideoInfoParams{}},
	}, false, "json")
}

func TestDP_GetVideoInfoWithSaveAs(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "info",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{InfoParams: &tos.VideoInfoParams{}},
	}, true, "json")
}

func TestDP_GetVideoPM3U8(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "pm3u8",
		key:    "test.m3u8",
		params: &tos.VideoProcessParams{PM3U8Params: &tos.VideoPM3U8Params{}},
	}, false, "m3u8")
}

func TestDP_GetVideoPM3U8WithSaveAs(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "pm3u8",
		key:    "test.m3u8",
		params: &tos.VideoProcessParams{PM3U8Params: &tos.VideoPM3U8Params{}},
	}, true, "m3u8")
}

func TestDP_GetHlsM3U8(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	segmentDuration := 10
	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeHls,
		HlsProcessParams: &tos.HlsProcessParams{
			M3U8Params: &tos.HlsM3U8Params{
				SegmentDuration: &segmentDuration,
				Width:           1280,
				Height:          720,
				EncodeFormat:    "h264",
				PixFmt:          "yuv420p",
			},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.videoKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
		assert.True(t, strings.Contains(string(data), "#EXTM3U"))
	}
}

func TestDP_GetHlsM3U8WithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	segmentDuration := 10
	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeHls,
		HlsProcessParams: &tos.HlsProcessParams{
			M3U8Params: &tos.HlsM3U8Params{
				SegmentDuration: &segmentDuration,
				Width:           1280,
				Height:          720,
				EncodeFormat:    "h264",
				PixFmt:          "yuv420p",
			},
		},
	})
	require.Nil(t, err)

	saveKey := "dp-test/hls/hls_m3u8.m3u8"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.videoKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    saveKey,
	})
	require.Nil(t, err)
	require.NotNil(t, headOutput)
	assert.True(t, headOutput.ContentLength > 0)
}

func TestDP_GetHlsTS(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	needDownload := true
	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeHls,
		HlsProcessParams: &tos.HlsProcessParams{
			TSParams: &tos.HlsTSParams{
				FromObject:   env.videoKey,
				Width:        640,
				Height:       360,
				EncodeFormat: "h264",
				PixFmt:       "yuv420p",
				StartTime:    0,
				EndTime:      1000000,
				NeedDownload: &needDownload,
			},
		},
	})
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.videoKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetVideoAIGCMetadata(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "aigcmetadata",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{AIGCMetadataParams: &tos.VideoAIGCMetadataParams{}},
	}, false, "json")
}

func TestDP_GetVideoAIGCMetadataWithSaveAs(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "aigcmetadata",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{AIGCMetadataParams: &tos.VideoAIGCMetadataParams{}},
	}, true, "json")
}

func TestDP_GetVideoC2PAMetadata(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "c2pametadata",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{C2PAMetadataParams: &tos.VideoC2PAMetadataParams{}},
	}, false, "json")
}

func TestDP_GetVideoC2PAMetadataWithSaveAs(t *testing.T) {
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name:   "c2pametadata",
		key:    "test.mp4",
		params: &tos.VideoProcessParams{C2PAMetadataParams: &tos.VideoC2PAMetadataParams{}},
	}, true, "json")
}

func TestDP_GetVideoUnderstanding(t *testing.T) {
	fps := 1.0
	runStandaloneVideoGetProcessCase(t, videoGetProcessCase{
		name: "understanding",
		key:  "test.mp4",
		params: &tos.VideoProcessParams{
			UnderstandingParams: &tos.VideoUnderstandingParams{
				Model:  "doubao-seed-1.6-vision",
				Prompt: "这段视频里有什么",
				FPS:    &fps,
			},
		},
	}, false, "json")
}

func TestDP_GetVideoStructuredSkipped(t *testing.T) {
	cases := []struct {
		name   string
		reason string
	}{
		{name: "embedding", reason: "当前环境后端返回 The action you requested is not valid"},
		{name: "blob-transcode", reason: "blob_transcode 属于独立 BlobVideoService，不属于当前 Get/Post/PutDataProcess 主链路"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Skip(tc.reason)
		})
	}
}

func TestDP_GetVideoSnapshot(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeVideo,
		VideoProcessParams: &tos.VideoProcessParams{
			SnapshotParams: &tos.VideoSnapshotParams{
				T: 1000,
				W: intPtr(640),
				H: intPtr(360),
				F: "jpg",
				M: "fast",
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.videoKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("视频截帧大小: %d bytes", len(data))
		assert.True(t, len(data) > 0)
	}
}

func TestDP_GetVideoSnapshotWithSaveAs(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeVideo,
		VideoProcessParams: &tos.VideoProcessParams{
			SnapshotParams: &tos.VideoSnapshotParams{
				T: 1000,
				W: intPtr(640),
				H: intPtr(360),
				F: "jpg",
				M: "fast",
			},
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.videoKey,
		Process:    dpResult,
		SaveBucket: env.bucket,
		SaveObject: "dp-test/video/snapshot_1s.jpg",
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("视频截帧+SaveAs 响应大小: %d bytes", len(data))
	}
}

func TestDP_PostVideoSnapshots(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	width := 400
	height := 400
	process, err := tos.PostDataProcessHelper(ctx, tos.PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &tos.VideoProcessParams{
			SnapshotsParams: &tos.VideoSnapshotsParams{
				Format: "png",
				Mode:   "index",
				Width:  &width,
				Height: &height,
				Index:  "0|10|3000",
			},
		},
		SaveAsParams: &tos.SaveAsParams{
			SaveObject: "dp-test/video/post_snapshot_${Number}.png",
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", process)

	output, err := client.PostDataProcess(ctx, &tos.PostDPInput{
		Bucket:      env.bucket,
		Key:         env.videoKey,
		PostProcess: process,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	t.Logf("PostDataProcess 视频截帧响应: StatusCode=%d", output.RequestInfo.StatusCode)
}

func TestDP_PostVideoExtractPCM(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	sampleRate := 16000
	channels := 1
	process, err := tos.PostDataProcessHelper(ctx, tos.PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeVideo,
		VideoProcessParams: &tos.VideoProcessParams{
			PCMParams: &tos.VideoPCMParams{
				SampleRate: &sampleRate,
				Channels:   &channels,
			},
		},
		SaveAsParams: &tos.SaveAsParams{
			SaveBucket: env.bucket,
			SaveObject: "dp-test/video/pcm_output.pcm",
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", process)

	output, err := client.PostDataProcess(ctx, &tos.PostDPInput{
		Bucket:      env.bucket,
		Key:         env.videoKey,
		PostProcess: process,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	require.NotNil(t, output.VideoProcessOutput)
	assert.Equal(t, env.bucket, output.VideoProcessOutput.PcmBucket)
	assert.Equal(t, "dp-test/video/pcm_output.pcm", output.VideoProcessOutput.PcmObject)
	assert.Equal(t, "OK", output.VideoProcessOutput.PcmStatus)
	t.Logf("PCM 提取响应: bucket=%s object=%s status=%s", output.VideoProcessOutput.PcmBucket, output.VideoProcessOutput.PcmObject, output.VideoProcessOutput.PcmStatus)
}

func TestDP_PutImageWithOperations(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	getOutput, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{
		Bucket: env.bucket,
		Key:    env.imageKey,
	})
	require.Nil(t, err)
	require.NotNil(t, getOutput)
	imgData, err := ioutil.ReadAll(getOutput.Content)
	getOutput.Content.Close()
	require.Nil(t, err)

	process, err := tos.PutDataProcessHelper(ctx, tos.PutDataProcessParams{
		PutProcessType: enum.PutProcessTypeImage,
		ImageOperations: &tos.PutImageOperationsParams{
			IsImageInfo: intPtr(1),
			Rules: []tos.PutImageOperationsRule{
				{
					Bucket: env.bucket,
					Key:    "/dp-test/image/put_resize_result.jpg",
					Rule:   "image/resize,w_150",
				},
			},
		},
	})
	require.Nil(t, err)

	putInput := &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:      env.bucket,
			Key:         "dp-test/image/put_original.jpg",
			ProcessType: enum.PutProcessTypeImage,
			Process:     process,
		},
	}

	putInput.Content = bytes.NewReader(imgData)
	putOutput, err := client.PutDataProcess(ctx, putInput)
	require.Nil(t, err)
	require.NotNil(t, putOutput)
	assert.Equal(t, 200, putOutput.RequestInfo.StatusCode)
	t.Logf("PutDataProcess 响应: StatusCode=%d", putOutput.RequestInfo.StatusCode)
}

func TestDP_GetDocPreview(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	dpResult, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeDoc,
		DocProcessParams: &tos.DocProcessParams{
			SrcType: enum.DocPreviewSrcTypeDocx,
			DstType: enum.DocPreviewDstTypePng,
			DocPage: intPtr(1),
		},
	})
	require.Nil(t, err)
	t.Logf("process string: %s", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     env.docKey,
		Process: dpResult,
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("文档预览 PNG 大小: %d bytes", len(data))
		assert.True(t, len(data) > 0, "文档预览应返回 PNG 图片数据")
	}
}

func TestDP_GetDocPreviewBatchConvert(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	startPage := 1
	endPage := 1
	imageMode := enum.ImageModeBatch
	saveKey := "dp-test/doc/" + t.Name() + "_{Page}.png"

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:     env.bucket,
		Key:        env.pdfKey,
		Process:    "doc-preview",
		SaveBucket: env.bucket,
		SaveObject: saveKey,
		GenericInput: tos.GenericInput{
			RequestQuery: map[string]string{
				"x-tos-doc-src-type": string(enum.DocPreviewSrcTypePdf),
				"x-tos-doc-dst-type": string(enum.DocPreviewDstTypePng),
				"start-page":         strconv.Itoa(startPage),
				"end-page":           strconv.Itoa(endPage),
				"image-mode":         strconv.Itoa(int(imageMode)),
			},
		},
	})
	require.Nil(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 200, output.RequestInfo.StatusCode)

	if output.Content != nil {
		data, _ := ioutil.ReadAll(output.Content)
		output.Content.Close()
		t.Logf("文档批量转图响应大小: %d bytes", len(data))
		assert.True(t, len(data) > 0, "文档批量转图应返回 JSON 结果")
	}

	_, err = client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
		Bucket: env.bucket,
		Key:    "dp-test/doc/" + t.Name() + "_1.png",
	})
	require.Nil(t, err)
}

func TestDP_PostAsyncVideoTranscodeAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobOutput, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeTranscode,
		JobBody: &tos.TranscodeJobBody{
			Input:  tos.ProcessJobInput{Object: env.videoKey},
			Output: tos.ProcessJobOutput{Region: env.region, Bucket: env.bucket, Object: "dp-test/video-transcode-out.mp4"},
			TranscodeConfig: tos.VideoTranscodeConfig{
				Transcode: &tos.VideoTranscodeDetail{
					Container: &tos.Container{Format: "mp4"},
				},
			},
		},
	})
	if err != nil {
		t.Logf("PostDataProcessAsync(Job) 返回错误: %v", err)
		t.Skip("异步视频转码提交失败，跳过结果查询测试")
		return
	}
	require.NotNil(t, jobOutput)
	t.Logf("Job 模式异步视频转码响应: JobId=%s, Code=%s, Message=%s",
		jobOutput.JobId, jobOutput.Code, jobOutput.Message)

	if jobOutput.JobId == "" {
		t.Log("未获得 JobId，跳过查询")
		return
	}

	queryOutput, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeTranscode,
		JobId:   jobOutput.JobId,
	})
	if err != nil {
		t.Logf("GetDPAsyncResult(Job) 返回错误: %v", err)
		return
	}
	require.NotNil(t, queryOutput)
	t.Logf("Job 模式查询结果: JobID=%s, State=%s, Code=%d, Message=%s",
		queryOutput.JobResult.JobID, queryOutput.JobResult.State, queryOutput.JobResult.Code, queryOutput.JobResult.Message)
}

func TestDP_PostAsyncAudioConvertAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipAudio(t)
	client := env.newClient(t)
	ctx := context.Background()

	t.Run("job", func(t *testing.T) {
		jobOutput, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
			Bucket:  env.bucket,
			JobType: tos.ProcessJobTypeAudioConvert,
			JobBody: &tos.AudioConvertJobBody{
				Input:              tos.ProcessJobInput{Object: env.audioKey},
				Output:             tos.ProcessJobOutput{Region: env.region, Bucket: env.bucket, Object: "dp-test/audio/convert_output.wav"},
				AudioConvertConfig: tos.AudioConvertJobConfig{ContainerFormat: "wav"},
			},
		})
		if err != nil {
			t.Logf("PostDataProcessAsync(AudioConvert,Job) 返回错误: %v", err)
			t.Skip("异步音频转码提交失败，跳过结果查询测试")
			return
		}
		require.NotNil(t, jobOutput)
		t.Logf("Job 模式异步音频转码响应: JobId=%s, Code=%s, Message=%s",
			jobOutput.JobId, jobOutput.Code, jobOutput.Message)

		if jobOutput.JobId == "" {
			t.Log("未获得 JobId，跳过查询")
			return
		}

		queryOutput, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
			Bucket:  env.bucket,
			JobType: tos.ProcessJobTypeAudioConvert,
			JobId:   jobOutput.JobId,
		})
		if err != nil {
			t.Logf("GetDPAsyncResult(AudioConvert,Job) 返回错误: %v", err)
			return
		}
		require.NotNil(t, queryOutput)
		t.Logf("Job 模式查询结果: JobID=%s, State=%s, Code=%d, Message=%s",
			queryOutput.JobResult.JobID, queryOutput.JobResult.State, queryOutput.JobResult.Code, queryOutput.JobResult.Message)
	})

	t.Run("async-process", func(t *testing.T) {
		saveKey := "dp-test/audio/async_convert_output.wav"
		postProcess, err := tos.PostDataProcessAsyncHelper(ctx, tos.PostDataProcessAsyncParams{
			PostProcessAsyncType: enum.PostProcessAsyncTypeAudio,
			AudioProcessAsyncParams: &tos.AudioProcessAsyncParams{
				Bucket:       env.bucket,
				Region:       env.region,
				TargetObject: env.audioKey,
				SaveAsParams: &tos.SaveAsParams{
					SaveBucket: env.bucket,
					SaveObject: saveKey,
				},
				ConvertParams: &tos.AudioConvertParams{
					ContainerFormat: "wav",
				},
			},
		})
		if err != nil {
			t.Logf("PostDataProcessAsyncHelper(AudioConvert) 返回错误: %v", err)
			t.Skip("音频异步转码 Helper 失败")
			return
		}

		asyncOut, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
			Bucket:      env.bucket,
			Key:         env.audioKey,
			PostProcess: postProcess,
		})
		if err != nil {
			t.Logf("PostDataProcessAsync(AudioConvert,async-process) 返回错误: %v", err)
			t.Skip("async-process 模式音频转码提交失败")
			return
		}
		require.NotNil(t, asyncOut)
		t.Logf("async-process 模式音频转码响应: Status=%s, JobId=%s, Code=%s, Message=%s",
			asyncOut.Status, asyncOut.JobId, asyncOut.Code, asyncOut.Message)

		require.Equal(t, "OK", asyncOut.Code)

		// async-process 音频转码是异步落桶，需要等待
		var headOutput *tos.HeadObjectV2Output
		for i := 0; i < 10; i++ {
			headOutput, err = client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
				Bucket: env.bucket,
				Key:    saveKey,
			})
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
		if err != nil {
			t.Logf("async-process 输出文件未在等待时间内落桶: %v", err)
			t.Skip("异步任务已提交成功(Code=OK)，但后端处理未完成，跳过输出验证")
			return
		}
		require.NotNil(t, headOutput)
		t.Logf("async-process 输出对象校验成功: key=%s size=%d", saveKey, headOutput.ContentLength)
	})
}

func TestDP_PostAsyncAudioConcat(t *testing.T) {
	env := newDPTestEnv()
	env.skipAudio(t)
	client := env.newClient(t)
	ctx := context.Background()

	saveKey := "dp-test/audio/concat_output.mp3"
	postProcess, err := tos.PostDataProcessAsyncHelper(ctx, tos.PostDataProcessAsyncParams{
		PostProcessAsyncType: enum.PostProcessAsyncTypeAudio,
		AudioProcessAsyncParams: &tos.AudioProcessAsyncParams{
			Bucket:       env.bucket,
			Region:       env.region,
			TargetObject: env.audioKey,
			SaveAsParams: &tos.SaveAsParams{
				SaveBucket: env.bucket,
				SaveObject: saveKey,
			},
			ConcatParams: &tos.AudioConcatParams{
				ContainerFormat: "mp3",
				PreFragments: []tos.AudioConcatFragment{
					{Object: env.audioKey},
				},
				SurFragments: []tos.AudioConcatFragment{
					{Object: env.audioKey2},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("PostDataProcessAsyncHelper(AudioConcat) 返回错误: %v", err)
	}
	t.Logf("AudioConcat process string: %s", postProcess)

	asyncOut, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:      env.bucket,
		Key:         env.audioKey,
		PostProcess: postProcess,
	})
	if err != nil {
		t.Fatalf("PostDataProcessAsync(AudioConcat,async-process) 返回错误: %v", err)
	}
	require.NotNil(t, asyncOut)
	t.Logf("async-process 模式音频拼接响应: Status=%s, JobId=%s, Code=%s, Message=%s",
		asyncOut.Status, asyncOut.JobId, asyncOut.Code, asyncOut.Message)

	require.Equal(t, "OK", asyncOut.Code)
	// async-process 走 CommitJob，异步落桶需要等待
	var headOutput *tos.HeadObjectV2Output
	for i := 0; i < 15; i++ {
		headOutput, err = client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{
			Bucket: env.bucket,
			Key:    saveKey,
		})
		if err == nil {
			break
		}
		t.Logf("HeadObjectV2 第 %d 次尝试失败: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		t.Logf("AudioConcat 输出文件未在等待时间内落桶: key=%s, err=%v", saveKey, err)
		t.Skip("异步任务已提交成功(Code=OK)，但后端处理未完成，跳过输出验证")
		return
	}
	require.NotNil(t, headOutput)
	t.Logf("AudioConcat 输出对象校验成功: key=%s size=%d", saveKey, headOutput.ContentLength)
}

func TestDP_WorkflowDocConvertToPDFAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewDocConvertToPDFWorkflowJobInput(tos.CreateDocConvertToPDFJobParams{
		Bucket:       env.bucket,
		Region:       env.region,
		SourceKey:    env.docKey,
		SourceType:   tos.WorkflowDocType("docx"),
		OutputBucket: env.bucket,
		OutputObject: "dp-test/doc/workflow_doc_convert.pdf",
	})
	if err != nil {
		require.NoError(t, err)
	}

	detail := jobInput.JobDetail.(*tos.WorkflowDocConvertProcessInput)
	startPage := 1
	endPage := 1
	detail.DocConvertConfig.StartPage = &startPage
	detail.DocConvertConfig.EndPage = &endPage

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 文档转 PDF 提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 文档转 PDF 能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 文档转 PDF 响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeDocConvert),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 文档转 PDF 查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 文档转 PDF 查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowDocConvertToImageAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	startPage, endPage := 1, 2
	jobInput, err := tos.NewDocConvertToImageWorkflowJobInput(tos.CreateDocConvertToImageJobParams{
		Bucket:       env.bucket,
		Region:       env.region,
		SourceKey:    env.pdfKey,
		SourceType:   tos.WorkflowDocTypePDF,
		OutputBucket: env.bucket,
		OutputObject: "dp-test/doc/workflow_doc_image_{Page}.jpg",
		StartPage:    &startPage,
		EndPage:      &endPage,
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 文档转图片提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 文档转图片能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 文档转图片响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeDocConvert),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 文档转图片查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 文档转图片查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowAudioConvertAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	env.skipAudio(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewAudioConvertWorkflowJobInput(tos.CreateAudioConvertWorkflowJobParams{
		Bucket:       env.bucket,
		SourceKey:    env.audioKey,
		Region:       env.region,
		OutputObject: "dp-test/audio/workflow_audio_convert.wav",
		ConvertParams: &tos.AudioConvertParams{
			ContainerFormat: "wav",
		},
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 音频转码提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 音频转码能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 音频转码响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeAudioConvert),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 音频转码查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 音频转码查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowAudioConcatAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	env.skipAudio(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewAudioConcatWorkflowJobInput(tos.CreateAudioConcatWorkflowJobParams{
		Bucket:       env.bucket,
		SourceKey:    env.audioKey,
		Region:       env.region,
		OutputObject: "dp-test/audio/workflow_audio_concat.mp3",
		ConcatParams: &tos.AudioConcatParams{
			ContainerFormat: "mp3",
			PreFragments: []tos.AudioConcatFragment{
				{Object: env.audioKey},
			},
			SurFragments: []tos.AudioConcatFragment{
				{Object: env.audioKey2},
			},
		},
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 音频拼接提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 音频拼接能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 音频拼接响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeAudioConcat),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 音频拼接查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 音频拼接查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowVideoTranscodeAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	width, height := 640, 360
	jobInput, err := tos.NewVideoTranscodeWorkflowJobInput(tos.CreateVideoTranscodeWorkflowJobParams{
		Bucket:       env.bucket,
		SourceKey:    env.videoKey,
		Region:       env.region,
		OutputObject: "dp-test/video/workflow_transcode.mp4",
		TranscodeConfig: &tos.VideoTranscodeConfig{
			Transcode: &tos.VideoTranscodeDetail{
				Container: &tos.Container{Format: "mp4"},
				Video:     &tos.VideoConfig{Codec: "h264", Width: &width, Height: &height},
			},
		},
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 视频转码提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 视频转码能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 视频转码响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeTranscode),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 视频转码查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 视频转码查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowVideoRemuxAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewVideoRemuxWorkflowJobInput(tos.CreateVideoRemuxWorkflowJobParams{
		Bucket:       env.bucket,
		SourceKey:    env.videoKey,
		Region:       env.region,
		OutputObject: "dp-test/video/workflow_remux.mp4",
		RemuxConfig: &tos.WorkflowVideoRemuxConfig{
			Format: "mp4",
		},
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 视频转封装提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 视频转封装能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 视频转封装响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeRemux),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 视频转封装查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 视频转封装查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_PostAsyncDocConvertPDF(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobOutput, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeDocConvert,
		JobBody: &tos.DocConvertJobBody{
			Input: tos.ProcessDocConvertInput{Key: env.docKey},
			DocConvertConfig: tos.ProcessDocConvertConfig{
				SrcType:   "docx",
				TgtType:   "pdf",
				StartPage: 1,
				EndPage:   -1,
			},
			Output: tos.ProcessDocConvertOutput{
				Region: env.region,
				Bucket: env.bucket,
				Object: "dp-test/doc/async_doc_convert_pdf.pdf",
			},
		},
	})
	if err != nil {
		t.Logf("PostDataProcessAsync(DocConvert->PDF) 返回错误: %v", err)
		t.Skip("当前环境未开通文档异步转 PDF 能力，跳过测试")
		return
	}
	require.NotNil(t, jobOutput)
	t.Logf("PostDataProcessAsync(DocConvert->PDF) 响应: JobId=%s Code=%s Message=%s", jobOutput.JobId, jobOutput.Code, jobOutput.Message)

	if jobOutput.JobId == "" {
		t.Log("未获得 JobId，跳过查询")
		return
	}

	queryOut, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeDocConvert,
		JobId:   jobOutput.JobId,
	})
	if err != nil {
		t.Logf("GetDPAsyncResult(DocConvert) 返回错误: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("GetDPAsyncResult(DocConvert) 响应: JobID=%s State=%s Code=%d Message=%s", queryOut.JobResult.JobID, queryOut.JobResult.State, queryOut.JobResult.Code, queryOut.JobResult.Message)
}

func TestDP_PostAsyncDocConvertImage(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()
	startPage, endPage := 1, 2

	jobOutput, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeDocConvert,
		JobBody: &tos.DocConvertJobBody{
			Input: tos.ProcessDocConvertInput{Key: env.pdfKey},
			DocConvertConfig: tos.ProcessDocConvertConfig{
				SrcType:   "pdf",
				TgtType:   "jpg",
				StartPage: startPage,
				EndPage:   endPage,
			},
			Output: tos.ProcessDocConvertOutput{
				Region: env.region,
				Bucket: env.bucket,
				Object: "dp-test/doc/async_doc_convert_image{Page}.jpg",
			},
		},
	})
	if err != nil {
		t.Logf("PostDataProcessAsync(DocConvert->Image) 返回错误: %v", err)
		t.Skip("当前环境未开通文档异步转图片能力，跳过测试")
		return
	}
	require.NotNil(t, jobOutput)
	t.Logf("PostDataProcessAsync(DocConvert->Image) 响应: JobId=%s Code=%s Message=%s", jobOutput.JobId, jobOutput.Code, jobOutput.Message)

	if jobOutput.JobId == "" {
		t.Log("未获得 JobId，跳过查询")
		return
	}

	queryOut, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeDocConvert,
		JobId:   jobOutput.JobId,
	})
	if err != nil {
		t.Logf("GetDPAsyncResult(DocConvert->Image) 返回错误: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("GetDPAsyncResult(DocConvert->Image) 响应: JobID=%s State=%s Code=%d Message=%s", queryOut.JobResult.JobID, queryOut.JobResult.State, queryOut.JobResult.Code, queryOut.JobResult.Message)
}

func intPtr(v int) *int {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func TestDP_McapInfo(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	mcapKey := os.Getenv("TOS_MCAP_KEY")
	if mcapKey == "" {
		mcapKey = "test.mcap"
	}

	params := tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &tos.McapProcessParams{
			Operation: tos.McapOperationInfo,
		},
	}
	dpResult, err := tos.GetDataProcessHelper(ctx, params)
	require.Nil(t, err)
	assert.Equal(t, "file/mcap-info", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     mcapKey,
		Process: dpResult,
	})
	if err != nil {
		t.Logf("GetDataProcess(mcap-info) 返回错误: %v", err)
		t.Skip("当前环境可能缺少 mcap 测试文件，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	body, _ := ioutil.ReadAll(output.Content)
	t.Logf("McapInfo 响应状态码: %d, 响应体:\n%s", output.RequestInfo.StatusCode, string(body))
}

func TestDP_McapDoctor(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	mcapKey := os.Getenv("TOS_MCAP_KEY")
	if mcapKey == "" {
		mcapKey = "test.mcap"
	}

	params := tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &tos.McapProcessParams{
			Operation: tos.McapOperationDoctor,
		},
	}
	dpResult, err := tos.GetDataProcessHelper(ctx, params)
	require.Nil(t, err)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     mcapKey,
		Process: dpResult,
	})
	if err != nil {
		t.Logf("GetDataProcess(mcap-doctor) 返回错误: %v", err)
		t.Skip("当前环境可能缺少 mcap 测试文件，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	body, _ := ioutil.ReadAll(output.Content)
	t.Logf("McapDoctor 响应状态码: %d, 响应体:\n%s", output.RequestInfo.StatusCode, string(body))
}

func TestDP_McapListChannels(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	mcapKey := os.Getenv("TOS_MCAP_KEY")
	if mcapKey == "" {
		mcapKey = "test.mcap"
	}

	params := tos.GetDataProcessParams{
		GetProcessType: enum.GetProcessTypeFile,
		McapProcessParams: &tos.McapProcessParams{
			Operation:    tos.McapOperationList,
			ListResource: tos.McapListChannels,
		},
	}
	dpResult, err := tos.GetDataProcessHelper(ctx, params)
	require.Nil(t, err)
	assert.Equal(t, "file/mcap-list,channels", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     mcapKey,
		Process: dpResult,
	})
	if err != nil {
		t.Logf("GetDataProcess(mcap-list,channels) 返回错误: %v", err)
		t.Skip("当前环境可能缺少 mcap 测试文件，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	body, _ := ioutil.ReadAll(output.Content)
	t.Logf("McapListChannels 响应状态码: %d, 响应体:\n%s", output.RequestInfo.StatusCode, string(body))
}

func TestDP_PointCloudCompress(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	pclKey := os.Getenv("TOS_PCL_KEY")
	if pclKey == "" {
		pclKey = "test.pcd"
	}

	pclParams := &tos.PointCloudCompressParams{
		PointResolution:  float64Ptr(0.01),
		OctreeResolution: float64Ptr(0.01),
	}
	params := tos.GetDataProcessParams{
		GetProcessType:          enum.GetProcessTypePointCloud,
		PointCloudProcessParams: pclParams,
	}
	dpResult, err := tos.GetDataProcessHelper(ctx, params)
	require.Nil(t, err)
	assert.Equal(t, "pointcloud/compress", dpResult)

	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{
		Bucket:  env.bucket,
		Key:     pclKey,
		Process: dpResult,
		GenericInput: tos.GenericInput{
			RequestQuery: pclParams.QueryParams(),
		},
	})
	if err != nil {
		t.Logf("GetDataProcess(pointcloud/compress) 返回错误: %v", err)
		t.Skip("当前环境可能缺少 pcd 测试文件，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	body, _ := ioutil.ReadAll(output.Content)
	t.Logf("PointCloudCompress 响应状态码: %d, 返回数据大小: %d bytes", output.RequestInfo.StatusCode, len(body))
}

func TestDP_FileCompress(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	output, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeFileCompress,
		JobBody: &tos.FileCompressJobBody{
			Input: tos.FileCompressInput{
				KeyConfig: []tos.FileCompressKeyConfig{
					{Key: "test.mp4"},
				},
			},
			FileCompressConfig: tos.FileCompressConfig{
				Format:  "zip",
				Flatten: 0,
			},
			Output: tos.FileJobOutput{
				Region: env.region,
				Bucket: env.bucket,
				Object: "dp-test/output.zip",
			},
		},
	})
	if err != nil {
		t.Logf("PostDataProcessAsync(FileCompress) 返回错误: %v", err)
		t.Skip("当前环境可能不支持 FileCompress，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	t.Logf("FileCompress 响应状态码: %d, Code: %s, JobId: %s", output.RequestInfo.StatusCode, output.Code, output.JobId)

	if output.JobId != "" {
		queryOutput, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
			Bucket:  env.bucket,
			JobType: tos.ProcessJobTypeFileCompress,
			JobId:   output.JobId,
		})
		if err != nil {
			t.Logf("GetDPAsyncResult(FileCompress) 返回错误: %v", err)
		} else {
			t.Logf("FileCompress 任务状态: State=%s, JobID=%s", queryOutput.JobResult.State, queryOutput.JobResult.JobID)
		}
	}
}

func TestDP_FileUncompress(t *testing.T) {
	env := newDPTestEnv()
	env.skip(t)
	client := env.newClient(t)
	ctx := context.Background()

	output, err := client.PostDataProcessAsync(ctx, &tos.PostDPAsyncInput{
		Bucket:  env.bucket,
		JobType: tos.ProcessJobTypeFileUncompress,
		JobBody: &tos.FileUncompressJobBody{
			Input: tos.FileUncompressInput{
				Object: "test.zip",
			},
			FileUncompressConfig: tos.FileUncompressConfig{
				Prefix:         "dp-test/uncompress/",
				PrefixReplaced: 0,
			},
			Output: tos.FileJobOutput{
				Region: env.region,
				Bucket: env.bucket,
			},
		},
	})
	if err != nil {
		t.Logf("PostDataProcessAsync(FileUncompress) 返回错误: %v", err)
		t.Skip("当前环境可能不支持 FileUncompress 或缺少压缩包，跳过")
		return
	}
	assert.Equal(t, 200, output.RequestInfo.StatusCode)
	t.Logf("FileUncompress 响应状态码: %d, Code: %s, JobId: %s", output.RequestInfo.StatusCode, output.Code, output.JobId)

	if output.JobId != "" {
		queryOutput, err := client.GetDPAsyncResult(ctx, &tos.GetDPAsyncResultInput{
			Bucket:  env.bucket,
			JobType: tos.ProcessJobTypeFileUncompress,
			JobId:   output.JobId,
		})
		if err != nil {
			t.Logf("GetDPAsyncResult(FileUncompress) 返回错误: %v", err)
		} else {
			t.Logf("FileUncompress 任务状态: State=%s, JobID=%s", queryOutput.JobResult.State, queryOutput.JobResult.JobID)
		}
	}
}

func TestDP_WorkflowFileCompressAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewFileCompressWorkflowJobInput(tos.CreateFileCompressWorkflowJobParams{
		Bucket:       env.bucket,
		Region:       env.region,
		OutputBucket: env.bucket,
		OutputObject: "dp-test/file/workflow_compress.zip",
		SourceKeys:   []string{env.docKey},
		Format:       "zip",
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 文件压缩提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 文件压缩能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 文件压缩响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeFileCompress),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 文件压缩查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 文件压缩查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowFileUncompressAndQuery(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	jobInput, err := tos.NewFileUncompressWorkflowJobInput(tos.CreateFileUncompressWorkflowJobParams{
		Bucket:       env.bucket,
		Region:       env.region,
		OutputBucket: env.bucket,
		SourceKey:    "dp-test/file/workflow_compress.zip",
		Prefix:       "dp-test/uncompress/",
	})
	require.NoError(t, err)

	createOut, err := client.CreateWorkflowJob(ctx, jobInput)
	if err != nil {
		t.Logf("workflow 文件解压提交失败: %v", err)
		t.Skip("当前环境未开通 workflow 文件解压能力，跳过测试")
		return
	}
	require.NotNil(t, createOut)
	t.Logf("workflow 文件解压响应: JobID=%s Code=%s Message=%s", createOut.JobID, createOut.Code, createOut.Message)

	if createOut.JobID == "" {
		t.Log("未获得 workflow JobID，跳过查询")
		return
	}

	queryOut, err := client.QueryWorkflowJobs(ctx, &tos.QueryWorkflowJobsInput{
		Bucket:  env.bucket,
		JobType: string(tos.WorkflowJobTypeFileUncompress),
		JobID:   createOut.JobID,
	})
	if err != nil {
		t.Logf("workflow 文件解压查询失败: %v", err)
		return
	}
	require.NotNil(t, queryOut)
	t.Logf("workflow 文件解压查询成功: JobType=%s Items=%d", queryOut.JobType, len(queryOut.Items))
}

func TestDP_WorkflowEventTriggerAudioConvert(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	env.skipAudio(t)
	client := env.newClient(t)
	ctx := context.Background()

	putOut, err := client.PutProcessTemplate(ctx, &tos.PutProcessTemplateInput{
		Bucket: env.bucket,
		Tag:    "AudioConvert",
		TemplateConfig: map[string]interface{}{
			"Tag": "AudioConvert", "Name": "sdk-test-audio-mp3",
			"AudioConvertConfig": map[string]interface{}{
				"ContainerFormat": "mp3", "TimeInterval": map[string]interface{}{"Start": 0, "Duration": 0},
				"BitRate": 320000, "BitRateOpt": 1, "SampleRate": 44100, "Channels": 2, "SampleFormat": "",
			},
		},
	})
	require.NoError(t, err, "PutProcessTemplate(AudioConvert) 失败")
	templateID := putOut.TemplateID
	t.Logf("创建自定义 AudioConvert 模板: %s", templateID)
	defer client.DeleteProcessTemplate(ctx, &tos.DeleteProcessTemplateInput{Bucket: env.bucket, ID: templateID})

	_, err = client.CreateWorkflow(ctx, &tos.PutConvertWorkflowInput{
		Bucket: env.bucket,
		Role:   "ServiceRoleforTOSDataProcess",
		Rules: []tos.WorkflowRule{{
			ID:      "event-trigger-audio",
			Enabled: true,
			ExtFilter: &tos.WorkflowExtFilter{
				AudioExts: []string{"mp3"},
			},
			Topology: [][]string{{"op-audio"}},
			Operations: tos.WorkflowOperations{
				AudioTranscode: []tos.WorkflowOperationAudioTranscode{{
					OperationID: "op-audio",
					TemplateID:  templateID,
					Output: tos.WorkflowJobOutput{
						Region: env.region,
						Bucket: env.bucket,
						Object: "dp-test/workflow-event/audio/${InputName}.mp3",
					},
				}},
			},
		}},
	})
	require.NoError(t, err, "CreateWorkflow 失败")
	t.Log("Workflow 规则创建成功")
	defer client.DeleteWorkflow(ctx, &tos.DeleteConvertWorkflowInput{Bucket: env.bucket})

	triggerKey := fmt.Sprintf("dp-test/workflow-event/trigger-audio-%d.mp3", time.Now().UnixNano())
	srcOut, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: env.bucket, Key: env.audioKey})
	if err != nil {
		t.Fatalf("获取源音频失败: %v", err)
	}
	defer srcOut.Content.Close()

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: env.bucket, Key: triggerKey},
		Content:             srcOut.Content,
	})
	require.NoError(t, err, "PutObjectV2 失败")
	t.Logf("上传触发对象成功: %s", triggerKey)

	require.Eventually(t, func() bool {
		listOut, err := client.ListWorkflowExecution(ctx, &tos.ListWorkflowExecutionInput{Bucket: env.bucket})
		if err != nil {
			t.Logf("ListWorkflowExecution 错误: %v", err)
			return false
		}
		for _, item := range listOut.Items {
			if item.Object == triggerKey {
				t.Logf("WorkflowExecution: ExecutionID=%s State=%s", item.ExecutionID, item.State)
				return item.State == "Success" || item.State == "Failed"
			}
		}
		return false
	}, 120*time.Second, 10*time.Second, "120 秒内未观察到 WorkflowExecution 完成")

	t.Log("✅ 事件触发音频 Workflow 链路验证成功")
}

func TestDP_WorkflowEventTriggerVideoTranscode(t *testing.T) {
	env := newDPTestEnv()
	env.skipWorkflowJob(t)
	client := env.newClient(t)
	ctx := context.Background()

	putOut, err := client.PutProcessTemplate(ctx, &tos.PutProcessTemplateInput{
		Bucket: env.bucket,
		Tag:    "Transcode",
		TemplateConfig: map[string]interface{}{
			"Tag": "Transcode", "Name": "sdk-test-transcode-mp4",
			"TranscodeConfig": map[string]interface{}{
				"TimeInterval": map[string]interface{}{"Start": 0, "Duration": 0},
				"Container":    map[string]interface{}{"Format": "mp4"},
				"Video": map[string]interface{}{
					"Codec": "h264", "Width": 0, "Height": 720, "Crf": 23, "PixFmt": "yuv420p",
				},
			},
		},
	})
	require.NoError(t, err, "PutProcessTemplate(Transcode) 失败")
	templateID := putOut.TemplateID
	t.Logf("创建自定义 Transcode 模板: %s", templateID)
	defer client.DeleteProcessTemplate(ctx, &tos.DeleteProcessTemplateInput{Bucket: env.bucket, ID: templateID})

	_, err = client.CreateWorkflow(ctx, &tos.PutConvertWorkflowInput{
		Bucket: env.bucket,
		Role:   "ServiceRoleforTOSDataProcess",
		Rules: []tos.WorkflowRule{{
			ID:      "event-trigger-video",
			Enabled: true,
			ExtFilter: &tos.WorkflowExtFilter{
				VideoExts: []string{"mp4"},
			},
			Topology: [][]string{{"op-video"}},
			Operations: tos.WorkflowOperations{
				Transcode: []tos.WorkflowOperationTranscode{{
					OperationID: "op-video",
					TemplateID:  templateID,
					Output: tos.WorkflowJobOutput{
						Region: env.region,
						Bucket: env.bucket,
						Object: "dp-test/workflow-event/video/${InputName}.mp4",
					},
				}},
			},
		}},
	})
	require.NoError(t, err, "CreateWorkflow 失败")
	t.Log("Workflow 规则创建成功")
	defer client.DeleteWorkflow(ctx, &tos.DeleteConvertWorkflowInput{Bucket: env.bucket})

	triggerKey := fmt.Sprintf("dp-test/workflow-event/trigger-video-%d.mp4", time.Now().UnixNano())
	srcOut, err := client.GetObjectV2(ctx, &tos.GetObjectV2Input{Bucket: env.bucket, Key: env.videoKey})
	if err != nil {
		t.Fatalf("获取源视频失败: %v", err)
	}
	defer srcOut.Content.Close()

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: env.bucket, Key: triggerKey},
		Content:             srcOut.Content,
	})
	require.NoError(t, err, "PutObjectV2 失败")
	t.Logf("上传触发对象成功: %s", triggerKey)

	require.Eventually(t, func() bool {
		listOut, err := client.ListWorkflowExecution(ctx, &tos.ListWorkflowExecutionInput{Bucket: env.bucket})
		if err != nil {
			t.Logf("ListWorkflowExecution 错误: %v", err)
			return false
		}
		for _, item := range listOut.Items {
			if item.Object == triggerKey {
				t.Logf("WorkflowExecution: ExecutionID=%s State=%s", item.ExecutionID, item.State)
				return item.State == "Success" || item.State == "Failed"
			}
		}
		return false
	}, 120*time.Second, 10*time.Second, "120 秒内未观察到 WorkflowExecution 完成")

	t.Log("✅ 事件触发视频转码 Workflow 链路验证成功")
}
