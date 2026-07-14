package tos

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// GetDataProcess 通过 GET 请求对 TOS 对象进行数据处理（图片/视频/文档/HLS/文件/点云）。
// 与 GetObjectV2 平级，独立构建请求，仅复用 SDK 底层 requestBuilder 基础设施。
// 使用 WithParams 自动序列化 GetObjectV2Input 的 location tag 字段，再手动覆盖 SaveBucket/SaveObject 为 base64 编码。
func (cli *ClientV2) GetDataProcess(ctx context.Context, input *GetObjectV2Input) (*GetObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}

	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithParams(*input)

	// WithParams 设置的 SaveBucket/SaveObject 是原始值，需手动覆盖为 base64 编码才能满足现在 TOS 网关要求
	if input.SaveBucket != "" {
		//用WithQuery来覆盖Params
		rb.WithQuery("x-tos-save-bucket", base64.URLEncoding.EncodeToString([]byte(input.SaveBucket)))
	}
	if input.SaveObject != "" {
		//用WithQuery来覆盖Params
		rb.WithQuery("x-tos-save-object", base64.URLEncoding.EncodeToString([]byte(input.SaveObject)))
	}

	// 不支持 SaveAs 的 process 类型，静默忽略 SaveBucket/SaveObject 以保证兼容性
	if !getProcessSupportsSaveAs(input.Process) {
		rb.Query.Del("x-tos-save-bucket")
		rb.Query.Del("x-tos-save-object")
	}

	// Deprecated: 此处为了保证平滑升级，保留文档处理兼容逻辑，与 GetObjectV2 保持一致
	if input.Process == "doc-preview" {
		if input.StartPage != nil {
			rb.WithQuery("start-page", strconv.Itoa(*input.StartPage))
		}
		if input.EndPage != nil {
			rb.WithQuery("end-page", strconv.Itoa(*input.EndPage))
		}
		if input.ImageMode != nil {
			rb.WithQuery("image-mode", strconv.Itoa(int(*input.ImageMode)))
		}
	}

	// --- 1. Range 请求处理：优先使用原始 Range 字符串，其次用 RangeStart/RangeEnd 构造 ---
	isRange := false
	if input.Range != "" {
		rb.WithHeader(HeaderRange, input.Range)
		isRange = true
	} else if input.RangeEnd != 0 || input.RangeStart != 0 {
		if input.RangeEnd < input.RangeStart {
			return nil, fmt.Errorf("tos: invalid range")
		}
		rb.Range = &Range{Start: input.RangeStart, End: input.RangeEnd}
		rb.WithHeader(HeaderRange, rb.Range.String())
		isRange = true
	}

	// Range 请求需要 trailer 机制传递分片 CRC64 校验值
	if isRange && !cli.disableTrailerHeader {
		rb.WithHeader(HeaderTosTrailer, "x-tos-hash-range-crc64ecma")
		rb.WithHeader(HeaderAcceptEncoding, "tos-raw-trailer")
	}

	// --- 2. 发送请求，允许 2xx 状态码（Range 请求返回 206） ---
	res, err := rb.WithRetry(nil, StatusCodeClassifier{}).Request(ctx, http.MethodGet, nil, cli.roundTripperWithSlowLog(expectedCode(rb)))
	if err != nil {
		return nil, err
	}

	// --- 3. 解析响应元数据 ---
	basic := GetObjectBasicOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: res.Header.Get(HeaderContentRange),
	}
	basic.ObjectMetaV2.fromResponseV2(res, cli.disableEncodingMeta)

	// --- 4. CRC64 校验：仅完整请求（200）才做，Range 请求（206）不校验 ---
	var serverCrc uint64
	var checker hash.Hash64
	if res.StatusCode == http.StatusOK && cli.enableCRC {
		serverCrc = basic.HashCrc64ecma
		checker = NewCRC(DefaultCrcTable(), 0)
	}

	// --- 5. 响应体处理：Range 请求的 trailer 模式需用 chunkReader 解包 ---
	body := res.Body
	if isRange && res.Header.Get(HeaderRawContentLength) != "" && res.Header.Get(HeaderContentEncoding) != "" && !cli.disableTrailerHeader {
		body = newChunkReader(body, basic.ContentLength)
	}

	// --- 6. 组装输出：wrapReader 负责限速、进度回调、CRC64 校验 ---
	output := GetObjectV2Output{
		GetObjectBasicOutput: GetObjectBasicOutput{
			RequestInfo:  basic.RequestInfo,
			ContentRange: basic.ContentRange,
			ObjectMetaV2: basic.ObjectMetaV2,
		},
		Content: wrapReader(body, basic.ContentLength, input.DataTransferListener, input.RateLimiter, &crcChecker{checker: checker, serverCrc: serverCrc}),
	}
	return &output, nil
}

// PutDataProcess 通过 PUT 请求在上传对象的同时进行数据处理（图片/视频/文档）。
// 与 PutObjectV2 平级，独立构建请求，仅复用 SDK 底层 requestBuilder 基础设施。
func (cli *ClientV2) PutDataProcess(ctx context.Context, input *PutObjectV2Input) (*PutObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}

	// --- 1. CRC64 校验初始化与 Content-Length 解析 ---
	var (
		checker       hash.Hash64
		content       = input.Content
		contentLength = input.ContentLength
	)
	if cli.enableCRC {
		checker = NewCRC(DefaultCrcTable(), 0)
	}
	if contentLength <= 0 {
		contentLength = tryResolveLength(content)
	}

	// --- 2. 重试策略：根据 Content 的能力选择重试方式 ---
	// Retryable 接口 → Reset 重试；io.Seeker → Seek 回起点重试；其余不重试
	var (
		onRetry    func(req *Request) error = nil
		classifier classifier
	)

	if content != nil {
		// *os.File 需要 wrapCloser 保证关闭时只关闭包装层而不关闭底层文件句柄
		if _, ok := content.(*os.File); ok {
			content = wrapCloser(content)
		}
		// wrapReader 注入限速、进度回调和 CRC64 校验
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, &crcChecker{checker: checker})
	}

	classifier = NoRetryClassifier{}

	if _, ok := input.Content.(Retryable); ok {
		onRetry = func(req *Request) error {
			retryableReader := req.Content.(Retryable)
			return retryableReader.Reset()
		}
		classifier = StatusCodeClassifier{}
	} else if seeker, ok := input.Content.(io.Seeker); ok {
		start, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		onRetry = func(req *Request) error {
			if seeker, ok := req.Content.(io.Seeker); ok {
				_, err := seeker.Seek(start, io.SeekStart)
				if err != nil {
					return err
				}
			} else {
				return newTosClientError("Io Reader not support retry", nil)
			}
			return nil
		}
		classifier = StatusCodeClassifier{}
	}

	// --- 3. 构建请求：WithParams 序列化 location tag 字段，setPutProcessHeaders 映射数据处理 header ---
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithContentLength(contentLength).
		WithEnableTrailer(input.ContentMD5 == "" && !cli.disableTrailerHeader).
		WithRetry(onRetry, classifier).
		WithParams(*input)

	// 根据 ProcessType + Process 自动映射出对应的请求 header
	setPutProcessHeaders(rb, input.ProcessType, input.Process)

	// 大文件上传使用 Expect: 100-continue 避免服务端拒绝后仍传输完整 body
	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPut, content, cli.roundTripperWithSlowLog(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	// --- 4. 校验响应 CRC64 与服务端返回值是否一致 ---
	if err = checkCrc64(res, checker); err != nil {
		return nil, err
	}
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)

	return &PutObjectV2Output{
		RequestInfo:   res.RequestInfo(),
		ETag:          res.Header.Get(HeaderETag),
		HashCrc64ecma: crc64,
		VersionID:     res.Header.Get(HeaderVersionID),
	}, nil
}

// parsePostProcessRequest 从请求体识别处理类型，并保留兼容旧版视频 JSON 的行为。
func parsePostProcessRequest(body string) (enum.PostProcessType, string) {
	trimmed := strings.TrimSpace(body)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var legacy struct {
			CanonicalURI string `json:"CanonicalUri"`
		}
		if err := json.Unmarshal([]byte(trimmed), &legacy); err == nil {
			return enum.PostProcessTypeVideo, legacy.CanonicalURI
		}
		return enum.PostProcessTypeVideo, ""
	}

	processURI := strings.TrimPrefix(trimmed, "x-tos-post-process=")
	switch {
	case strings.HasPrefix(processURI, "image/"):
		return enum.PostProcessTypeImage, processURI
	case strings.HasPrefix(processURI, "video/"):
		return enum.PostProcessTypeVideo, processURI
	case strings.HasPrefix(processURI, "doc-preview"):
		return enum.PostProcessTypeDoc, processURI
	default:
		return "", processURI
	}
}

// parsePostImageOutput 根据请求是否包含 SaveAs 区分结构化响应和由调用方负责关闭的响应流。
func parsePostImageOutput(res *Response, output *PostDPOutput, saveAs bool) (*PostDPOutput, error) {
	rawContentType := res.Header.Get(HeaderContentType)
	imageOutput := &ImageProcessOutput{}
	output.ImageProcessOutput = imageOutput
	if !saveAs {
		imageOutput.Content = res.Body
		imageOutput.ContentType = rawContentType
		imageOutput.ContentLength = res.ContentLength
		return output, nil
	}

	defer res.Close()
	if err := marshalOutput(res, imageOutput); err != nil {
		return nil, err
	}
	return output, nil
}

func usesGenericVideoSaveAs(processURI string) bool {
	if strings.HasPrefix(processURI, "video/remux") {
		return true
	}
	return strings.HasPrefix(processURI, "video/convert") && !strings.Contains(processURI, ",f_pcm")
}

// parsePostVideoOutput 根据协议版本和请求操作区分旧版视频响应与新版 Convert/Remux SaveAs 响应。
func parsePostVideoOutput(res *Response, output *PostDPOutput, processURI string, legacyJSON bool) (*PostDPOutput, error) {
	defer res.Close()

	// 旧版 JSON 同步 Transcode 的 bucket/object/object_size/status 历史上会反序列化到
	// PcmDataProcessOutput。即使 JSON 中带 CanonicalUri，也必须保留该公开字段行为。
	if !legacyJSON && usesGenericVideoSaveAs(processURI) {
		var response struct {
			Bucket     string                      `json:"bucket"`
			Object     string                      `json:"object"`
			FileSize   int64                       `json:"fileSize,string"`
			ObjectSize int64                       `json:"object_size,string"`
			Status     enum.VideoDataProcessStatus `json:"status"`
		}
		if err := marshalOutput(res, &response); err != nil {
			return nil, err
		}
		size := response.ObjectSize
		if size == 0 {
			size = response.FileSize
		}
		output.VideoProcessOutput = &VideoProcessOutput{
			SaveAsBucket:     response.Bucket,
			SaveAsObject:     response.Object,
			SaveAsObjectSize: size,
			SaveAsStatus:     response.Status,
		}
		return output, nil
	}

	if err := marshalOutput(res, output); err != nil {
		return nil, err
	}
	return output, nil
}

// parsePostUntypedOutput 校验尚无类型化响应的 Post 操作，避免将 Doc 等响应误判成视频。
func parsePostUntypedOutput(res *Response, output *PostDPOutput) (*PostDPOutput, error) {
	defer res.Close()
	var raw json.RawMessage
	if err := marshalOutput(res, &raw); err != nil {
		return nil, err
	}
	return output, nil
}

// PostDataProcess 通过 POST 请求对 TOS 对象进行同步数据处理。
// PostProcess 字段支持 Post 处理流水线字符串或旧版 JSON，SDK 会自动补全
// "x-tos-post-process=" 前缀。
func (cli *ClientV2) PostDataProcess(ctx context.Context, input *PostDPInput) (*PostDPOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	// --- 自动补全 post-process 前缀：用户传入裸 process 字符串时补 "x-tos-post-process="，JSON 格式不补 ---
	postProcessBody := strings.TrimSpace(input.PostProcess)
	if postProcessBody != "" &&
		!strings.HasPrefix(postProcessBody, "x-tos-post-process=") &&
		!strings.HasPrefix(postProcessBody, "{") &&
		!strings.HasPrefix(postProcessBody, "[") {
		postProcessBody = "x-tos-post-process=" + postProcessBody
	}

	// x-tos-post-process 作为 query key 标记（值为空），实际参数通过 body 传递。
	// 新版 Post 视频协议使用 text/plain；保留旧版 JSON body 的 application/json 兼容行为。
	legacyJSON := strings.HasPrefix(postProcessBody, "{") || strings.HasPrefix(postProcessBody, "[")
	contentType := "text/plain"
	if legacyJSON {
		contentType = "application/json"
	}
	processType, processURI := parsePostProcessRequest(postProcessBody)

	res, err := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("x-tos-post-process", "").
		WithHeader(HeaderContentType, contentType).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader([]byte(postProcessBody)), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}

	output := &PostDPOutput{
		RequestInfo: res.RequestInfo(),
	}
	switch processType {
	case enum.PostProcessTypeImage:
		return parsePostImageOutput(res, output, strings.Contains(processURI, "&x-tos-save-object="))
	case enum.PostProcessTypeVideo:
		return parsePostVideoOutput(res, output, processURI, legacyJSON)
	default:
		return parsePostUntypedOutput(res, output)
	}
}

// PostDataProcessAsync 提交异步数据处理任务。新调用应设置 JobType + JobBody，
// SDK 将结构化 JobBody 序列化为 JSON；PostDataProcessAsyncHelper 返回的 JSON 字符串
// 也可以直接作为 JobBody 传入。请求通过 job_type + media_jobs/file_jobs query 提交。
// Key + PostProcess 的 x-tos-async-process 模式仅为旧版兼容路径。
func (cli *ClientV2) PostDataProcessAsync(ctx context.Context, input *PostDPAsyncInput) (*PostDPAsyncOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.JobBody != nil && input.JobType == "" {
		return nil, fmt.Errorf("tos: JobType is required when JobBody is set")
	}

	if input.JobType != "" {
		// --- Job 模式：JobBody 序列化为 JSON，通过 job_type + media_jobs/file_jobs query 提交 ---
		if input.JobBody == nil {
			return nil, fmt.Errorf("tos: JobBody is required when JobType is set")
		}
		data, err := marshalPostDataProcessAsyncJobBody(input.JobBody)
		if err != nil {
			return nil, err
		}
		// jobsQueryParam 根据 JobType 返回 "media_jobs" 或 "file_jobs"
		queryKey := input.JobType.jobsQueryParam()
		res, err := cli.newBuilder(input.Bucket, "").
			SetGeneric(input.GenericInput).
			WithQuery("job_type", string(input.JobType)).
			WithQuery(queryKey, "").
			WithHeader(HeaderContentType, "application/json").
			WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
			Request(ctx, http.MethodPost, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
		if err != nil {
			return nil, err
		}
		defer res.Close()

		output := PostDPAsyncOutput{
			RequestInfo: res.RequestInfo(),
		}
		if err = marshalOutput(res, &output); err != nil {
			return nil, err
		}
		return &output, nil
	}

	// Deprecated: 旧版 async-process 模式，query-string process 作为 body 提交。
	if input.Key == "" {
		return nil, fmt.Errorf("tos: Key is required for async-process mode")
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("x-tos-async-process", "").
		WithRetry(OnRetryFromStart, StatusCodeClassifier{})

	// 自动补全 "x-tos-async-process=" 前缀，JSON 格式不补
	postProcessBody := input.PostProcess
	var bodyReader io.Reader
	if postProcessBody != "" {
		if !strings.HasPrefix(postProcessBody, "{") &&
			!strings.HasPrefix(postProcessBody, "[") &&
			!strings.HasPrefix(postProcessBody, "x-tos-async-process=") {
			postProcessBody = "x-tos-async-process=" + postProcessBody
		}
		bodyReader = bytes.NewReader([]byte(postProcessBody))
	}

	res, err := rb.Request(ctx, http.MethodPost, bodyReader, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PostDPAsyncOutput{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func marshalPostDataProcessAsyncJobBody(jobBody interface{}) ([]byte, error) {
	var raw []byte
	switch body := jobBody.(type) {
	case string:
		raw = []byte(body)
	case []byte:
		raw = body
	case json.RawMessage:
		raw = body
	default:
		data, _, err := marshalInput("PostDataProcessAsyncInput.JobBody", jobBody)
		return data, err
	}

	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' || !json.Valid(raw) {
		return nil, fmt.Errorf("tos: JobBody must contain a valid JSON object")
	}
	return raw, nil
}

// GetDPAsyncResult 查询异步数据处理任务结果，支持两种模式：
// 1. Job 模式：设置 JobType + Bucket + JobId，通过 job_type + job_id query 查询，返回 ProcessJobResult。
// 2. async-process 模式：设置 JobId，通过 x-tos-async-result + job-id query 查询，返回 DPAsyncResult。
func (cli *ClientV2) GetDPAsyncResult(ctx context.Context, input *GetDPAsyncResultInput) (*GetDPAsyncResultOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if input.JobId == "" {
		return nil, InvalidMarshal
	}

	if input.JobType != "" {
		// --- Job 模式：通过 job_type + job_id query 查询，反序列化为 ProcessJobResult ---
		if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
			return nil, err
		}
		res, err := cli.newBuilder(input.Bucket, "").
			SetGeneric(input.GenericInput).
			WithQuery("job_type", string(input.JobType)).
			WithQuery("job_id", input.JobId).
			WithRetry(nil, StatusCodeClassifier{}).
			Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
		if err != nil {
			return nil, err
		}
		defer res.Close()

		var jobResult ProcessJobResult
		if err = marshalOutput(res, &jobResult); err != nil {
			return nil, err
		}
		output := GetDPAsyncResultOutput{
			RequestInfo: res.RequestInfo(),
			JobResult:   &jobResult,
		}
		return &output, nil
	}

	// --- async-process 模式：通过 x-tos-async-result + job-id query 查询，反序列化为 DPAsyncResult ---
	res, err := cli.newBuilder("", "").
		SetGeneric(input.GenericInput).
		WithQuery("x-tos-async-result", "").
		WithQuery("job-id", input.JobId).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetDPAsyncResultOutput{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// getProcessSupportsSaveAs 判断 process 字符串对应的操作是否支持 SaveAs。
// inspect、info 类只读操作和 hls/ts 不支持另存为，后端会报错；
// 其余 image/video/doc 预览/转码等操作支持。
func getProcessSupportsSaveAs(process string) bool {
	if process == "" {
		return false
	}
	// 不支持 SaveAs 的精确匹配，使用 enum 拼接避免 magic string
	switch process {
	case string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationInspect),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationInfo),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationAverageHue),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationGetAIGCMetadata),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationGetC2PAMetadata),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationAiTag),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationEmbedding),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationUnderstanding),
		string(enum.GetProcessTypeImage) + "/" + string(enum.ImageOperationOCR),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationInfo),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationPm3u8),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationAIGCMetadata),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationC2PAMetadata),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationEmbedding),
		string(enum.GetProcessTypeVideo) + "/" + string(enum.VideoOperationUnderstanding):
		return false
	}
	// hls/ts 前缀匹配
	if strings.HasPrefix(process, string(enum.GetProcessTypeHls)+"/ts") {
		return false
	}
	// file/ 和 pointcloud/ 不支持 SaveAs
	if strings.HasPrefix(process, string(enum.GetProcessTypeFile)+"/") ||
		strings.HasPrefix(process, string(enum.GetProcessTypePointCloud)+"/") {
		return false
	}
	return true
}

// setPutProcessHeaders 根据 ProcessType + Process 自动填对应的请求 header。
// Image: JSON 值 → X-Tos-Image-Operations，process string → X-Tos-Process
// Video: base64(JSON) → X-Tos-Video-Operations
// Doc: process string → X-Tos-Process
func setPutProcessHeaders(rb *requestBuilder, processType enum.PutProcessType, process string) {
	if processType == "" || process == "" {
		return
	}
	switch processType {
	case enum.PutProcessTypeImage:
		// 这里因为图片的Put请求可以走JSON 所以单独特判一下 process字符串 走不同的请求头
		if strings.HasPrefix(process, "{") {
			rb.WithHeader(HeaderImageOperations, process)
		} else {
			rb.WithHeader(HeaderProcess, process)
		}
	case enum.PutProcessTypeVideo:
		rb.WithHeader(HeaderVideoOperations, base64.StdEncoding.EncodeToString([]byte(process)))
	case enum.PutProcessTypeDoc:
		rb.WithHeader(HeaderProcess, process)
	}
}
