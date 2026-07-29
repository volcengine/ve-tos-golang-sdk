package tos

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type workflowRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f workflowRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestWorkflowRouteBuilderUsesBucketStyle(t *testing.T) {
	cli := &ClientV2{
		Client: Client{
			scheme:     "https",
			host:       "tos.example.com",
			urlMode:    urlModeDefault,
			recognizer: EmptyContentTypeRecognizer{},
		},
	}

	req := cli.newBuilder("my-bucket", "").
		WithQuery("workflow_execution", "").
		WithQuery("id", "exec-id").
		Build(http.MethodGet, nil)

	if req.Host != "my-bucket.tos.example.com" {
		t.Fatalf("unexpected host: got %q", req.Host)
	}
	if req.Path != "/" {
		t.Fatalf("unexpected path: got %q", req.Path)
	}
	if got := req.Query.Get("id"); got != "exec-id" {
		t.Fatalf("unexpected execution id query: got %q", got)
	}
	if _, ok := req.Query["workflow_execution"]; !ok {
		t.Fatal("expected workflow_execution query marker")
	}
}

func TestWorkflowConfigJSONShape(t *testing.T) {
	rule := WorkflowRule{
		ID:      "rule-1",
		Enabled: true,
		ExtFilter: &WorkflowExtFilter{
			VideoExts: []string{"mp4"},
		},
		Topology: [][]string{{"op-1"}},
		Operations: WorkflowOperations{
			Transcode: []WorkflowOperationTranscode{{
				OperationID:         "op-1",
				TemplateID:          "tpl-1",
				WatermarkTemplateID: []string{"wm-tpl-1", "wm-tpl-2"},
				Output: WorkflowJobOutput{
					Bucket: "target-bucket",
					Object: "out/${Number}.mp4",
				},
			}},
		},
	}

	data, _, err := marshalInput("workflow-rule", struct {
		Role  string         `json:"Role,omitempty"`
		Rules []WorkflowRule `json:"Rules,omitempty"`
	}{
		Role:  "trn:role",
		Rules: []WorkflowRule{rule},
	})
	if err != nil {
		t.Fatalf("marshal workflow config failed: %v", err)
	}

	got := string(data)
	if !containsAll(got,
		`"Role":"trn:role"`,
		`"ID":"rule-1"`,
		`"Enabled":true`,
		`"VideoExts":["mp4"]`,
		`"Topology":[["op-1"]]`,
		`"TemplateID":"tpl-1"`,
		`"WatermarkTemplateID":["wm-tpl-1","wm-tpl-2"]`,
	) {
		t.Fatalf("unexpected workflow config json: %s", got)
	}
}

func TestWorkflowQueryJobsRouteShape(t *testing.T) {
	cli := &ClientV2{
		Client: Client{
			scheme:     "https",
			host:       "tos.example.com",
			urlMode:    urlModeDefault,
			recognizer: EmptyContentTypeRecognizer{},
		},
	}

	req := cli.newBuilder("my-bucket", "").
		WithQuery("job_type", string(WorkflowJobTypeDocConvert)).
		WithQuery("page_size", "20").
		WithQuery("page_token", "token-1").
		WithQuery("start_time", "100").
		WithQuery("end_time", "200").
		Build(http.MethodGet, nil)

	if req.Host != "my-bucket.tos.example.com" {
		t.Fatalf("unexpected host: got %q", req.Host)
	}
	if req.Path != "/" {
		t.Fatalf("unexpected path: got %q", req.Path)
	}
	if got := req.Query.Get("job_type"); got != string(WorkflowJobTypeDocConvert) {
		t.Fatalf("unexpected job type: %q", got)
	}
	if got := req.Query.Get("page_size"); got != "20" {
		t.Fatalf("unexpected page_size: %q", got)
	}
	if got := req.Query.Get("page_token"); got != "token-1" {
		t.Fatalf("unexpected page_token: %q", got)
	}
}

func TestDocWorkflowHelpers(t *testing.T) {
	pdfInput, err := NewDocConvertToPDFWorkflowJobInput(CreateDocConvertToPDFJobParams{
		Bucket:       "my-bucket",
		Region:       "cn-beijing",
		SourceKey:    "input.docx",
		SourceType:   WorkflowDocType("docx"),
		OutputBucket: "dst-bucket",
		OutputObject: "output.pdf",
	})
	if err != nil {
		t.Fatalf("NewDocConvertToPDFWorkflowJobInput error: %v", err)
	}
	if pdfInput.JobType != WorkflowJobTypeDocConvert {
		t.Fatalf("unexpected pdf job type: %q", pdfInput.JobType)
	}
	if pdfInput.Bucket != "my-bucket" {
		t.Fatalf("unexpected pdf bucket: %q", pdfInput.Bucket)
	}
	pdfDetail, ok := pdfInput.JobDetail.(*WorkflowDocConvertProcessInput)
	if !ok {
		t.Fatalf("unexpected pdf job detail type: %T", pdfInput.JobDetail)
	}
	if pdfDetail.DocConvertConfig.TgtType != WorkflowDocTypePDF {
		t.Fatalf("unexpected pdf tgt type: %q", pdfDetail.DocConvertConfig.TgtType)
	}
	if pdfDetail.Output.Region != "cn-beijing" || pdfDetail.Output.Bucket != "dst-bucket" {
		t.Fatalf("unexpected pdf output: %+v", pdfDetail.Output)
	}

	startPage := 1
	endPage := 3
	imageInput, err := NewDocConvertToImageWorkflowJobInput(CreateDocConvertToImageJobParams{
		Bucket:       "my-bucket",
		Region:       "cn-beijing",
		SourceKey:    "input.pptx",
		SourceType:   WorkflowDocType("pptx"),
		OutputBucket: "dst-bucket",
		OutputObject: "page-{Page}.jpg",
		StartPage:    &startPage,
		EndPage:      &endPage,
	})
	if err != nil {
		t.Fatalf("NewDocConvertToImageWorkflowJobInput error: %v", err)
	}
	imageDetail := imageInput.JobDetail.(*WorkflowDocConvertProcessInput)
	if imageDetail.DocConvertConfig.TgtType != WorkflowDocTypeJPG {
		t.Fatalf("unexpected image tgt type: %q", imageDetail.DocConvertConfig.TgtType)
	}
	if imageDetail.DocConvertConfig.StartPage == nil || *imageDetail.DocConvertConfig.StartPage != 1 || imageDetail.DocConvertConfig.EndPage == nil || *imageDetail.DocConvertConfig.EndPage != 3 {
		t.Fatalf("unexpected page range: start=%v end=%v", imageDetail.DocConvertConfig.StartPage, imageDetail.DocConvertConfig.EndPage)
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestWorkflowMediaJobBuilders(t *testing.T) {
	audioInput, err := NewAudioConvertWorkflowJobInput(CreateAudioConvertWorkflowJobParams{
		Bucket:       "media-bucket",
		SourceKey:    "test.mp3",
		Region:       "cn-beijing",
		OutputObject: "audio/output.wav",
		ConvertParams: &AudioConvertParams{
			ContainerFormat: "wav",
		},
	})
	if err != nil {
		t.Fatalf("NewAudioConvertWorkflowJobInput error: %v", err)
	}
	if audioInput.JobType != WorkflowJobTypeAudioConvert {
		t.Fatalf("unexpected audio job type: %q", audioInput.JobType)
	}
	audioDetail, ok := audioInput.JobDetail.(*WorkflowAudioJobDetail)
	if !ok {
		t.Fatalf("unexpected audio job detail type: %T", audioInput.JobDetail)
	}
	if audioDetail.Tag != string(WorkflowJobTypeAudioConvert) {
		t.Fatalf("unexpected audio tag: %q", audioDetail.Tag)
	}
	if audioDetail.Input.Object != "test.mp3" || audioDetail.Output.Bucket != "media-bucket" || audioDetail.Output.Object != "audio/output.wav" {
		t.Fatalf("unexpected audio workflow detail: %+v", audioDetail)
	}

	audioConcatInput, err := NewAudioConcatWorkflowJobInput(CreateAudioConcatWorkflowJobParams{
		Bucket:       "media-bucket",
		SourceKey:    "test.mp3",
		Region:       "cn-beijing",
		OutputObject: "audio/concat.mp3",
		ConcatParams: &AudioConcatParams{
			ContainerFormat: "mp3",
			PreFragments:    []AudioConcatFragment{{Object: "pre-1.mp3"}},
		},
	})
	if err != nil {
		t.Fatalf("NewAudioConcatWorkflowJobInput error: %v", err)
	}
	audioConcatDetail := audioConcatInput.JobDetail.(*WorkflowAudioJobDetail)
	if audioConcatDetail.Tag != string(WorkflowJobTypeAudioConcat) {
		t.Fatalf("unexpected audio concat tag: %q", audioConcatDetail.Tag)
	}

	width := 640
	height := 360
	videoInput, err := NewVideoTranscodeWorkflowJobInput(CreateVideoTranscodeWorkflowJobParams{
		Bucket:       "media-bucket",
		SourceKey:    "test.mp4",
		Region:       "cn-beijing",
		OutputObject: "video/output.mp4",
		TranscodeConfig: &VideoTranscodeConfig{
			Transcode: &VideoTranscodeDetail{
				Container: &Container{Format: "mp4"},
				Video:     &VideoConfig{Codec: "h264", Width: &width, Height: &height},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewVideoTranscodeWorkflowJobInput error: %v", err)
	}
	videoDetail, ok := videoInput.JobDetail.(*WorkflowVideoJobDetail)
	if !ok {
		t.Fatalf("unexpected video job detail type: %T", videoInput.JobDetail)
	}
	if videoInput.JobType != WorkflowJobTypeTranscode || videoDetail.Tag != string(WorkflowJobTypeTranscode) {
		t.Fatalf("unexpected video transcode metadata: jobType=%q tag=%q", videoInput.JobType, videoDetail.Tag)
	}
	if videoDetail.TranscodeConfig == nil || videoDetail.TranscodeConfig.Transcode == nil || videoDetail.Output.Object != "video/output.mp4" {
		t.Fatalf("unexpected video transcode detail: %+v", videoDetail)
	}

	remuxInput, err := NewVideoRemuxWorkflowJobInput(CreateVideoRemuxWorkflowJobParams{
		Bucket:       "media-bucket",
		SourceKey:    "test.mp4",
		Region:       "cn-beijing",
		OutputObject: "video/remux.mp4",
		RemuxConfig: &WorkflowVideoRemuxConfig{
			Format: "mp4",
		},
	})
	if err != nil {
		t.Fatalf("NewVideoRemuxWorkflowJobInput error: %v", err)
	}
	remuxDetail := remuxInput.JobDetail.(*WorkflowVideoJobDetail)
	if remuxDetail.Tag != string(WorkflowJobTypeRemux) || remuxDetail.RemuxConfig == nil || remuxDetail.RemuxConfig.Format != "mp4" {
		t.Fatalf("unexpected remux detail: %+v", remuxDetail)
	}

	speechInput, err := NewSpeechRecognitionWorkflowJobInput(CreateSpeechRecognitionWorkflowJobParams{
		Bucket:       "media-bucket",
		SourceKey:    "speech.wav",
		Region:       "cn-beijing",
		OutputObject: "speech/result.json",
		SpeechRecognitionConfig: &SpeechRecognitionJobConfig{
			Language:     "zh",
			OutputFormat: "json",
		},
	})
	if err != nil {
		t.Fatalf("NewSpeechRecognitionWorkflowJobInput error: %v", err)
	}
	speechDetail := speechInput.JobDetail.(*WorkflowVideoJobDetail)
	if speechInput.JobType != WorkflowJobTypeSpeechRecognition ||
		speechDetail.Tag != string(WorkflowJobTypeSpeechRecognition) ||
		speechDetail.SpeechRecognitionConfig == nil ||
		speechDetail.SpeechRecognitionConfig.Language != "zh" ||
		speechDetail.Output.Object != "speech/result.json" {
		t.Fatalf("unexpected speech recognition detail: %+v", speechDetail)
	}

	compressInput, err := NewFileCompressWorkflowJobInput(CreateFileCompressWorkflowJobParams{
		Bucket:       "file-bucket",
		Region:       "cn-beijing",
		OutputObject: "output.zip",
		SourceKeys:   []string{"doc1.txt", "doc2.txt"},
		Format:       "zip",
	})
	if err != nil {
		t.Fatalf("NewFileCompressWorkflowJobInput error: %v", err)
	}
	if compressInput.JobType != WorkflowJobTypeFileCompress {
		t.Fatalf("unexpected compress job type: %q", compressInput.JobType)
	}
	compressDetail, ok := compressInput.JobDetail.(*WorkflowFileJobDetail)
	if !ok {
		t.Fatalf("unexpected compress job detail type: %T", compressInput.JobDetail)
	}
	if compressDetail.FileCompressConfig == nil || compressDetail.FileCompressConfig.Format != "zip" {
		t.Fatalf("unexpected compress config: %+v", compressDetail.FileCompressConfig)
	}
	if compressDetail.Output.Bucket != "file-bucket" || compressDetail.Output.Object != "output.zip" {
		t.Fatalf("unexpected compress output: %+v", compressDetail.Output)
	}

	uncompressInput, err := NewFileUncompressWorkflowJobInput(CreateFileUncompressWorkflowJobParams{
		Bucket:    "file-bucket",
		Region:    "cn-beijing",
		SourceKey: "archive.zip",
		Prefix:    "uncompressed/",
	})
	if err != nil {
		t.Fatalf("NewFileUncompressWorkflowJobInput error: %v", err)
	}
	if uncompressInput.JobType != WorkflowJobTypeFileUncompress {
		t.Fatalf("unexpected uncompress job type: %q", uncompressInput.JobType)
	}
	uncompressDetail, ok := uncompressInput.JobDetail.(*WorkflowFileJobDetail)
	if !ok {
		t.Fatalf("unexpected uncompress job detail type: %T", uncompressInput.JobDetail)
	}
	if uncompressDetail.FileUncompressConfig == nil || uncompressDetail.FileUncompressConfig.Prefix != "uncompressed/" {
		t.Fatalf("unexpected uncompress config: %+v", uncompressDetail.FileUncompressConfig)
	}
}

func TestCreateWorkflowJobUsesPublicEndpointWithMediaJobs(t *testing.T) {
	cli := &ClientV2{
		Client: Client{
			scheme:     "https",
			host:       "tos.example.com",
			urlMode:    urlModeDefault,
			recognizer: EmptyContentTypeRecognizer{},
		},
	}
	cli.SetHTTPTransport(workflowRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.Host != "my-bucket.tos.example.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if req.URL.Path != "/" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		if _, ok := req.URL.Query()["media_jobs"]; !ok {
			t.Fatal("expected media_jobs query parameter")
		}
		if req.URL.Query().Get("job_type") != string(WorkflowJobTypeDocConvert) {
			t.Fatalf("unexpected job_type: %s", req.URL.Query().Get("job_type"))
		}
		if req.Header.Get(HeaderContentType) != "application/json" {
			t.Fatalf("unexpected content-type: %s", req.Header.Get(HeaderContentType))
		}
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if strings.Contains(string(body), "BucketName") || strings.Contains(string(body), "BucketID") {
			t.Fatalf("body should only contain JobDetail, got: %s", string(body))
		}
		if !strings.Contains(string(body), `"foo"`) {
			t.Fatalf("unexpected request body: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderRequestID: []string{"req-1"}},
			Body:       ioutil.NopCloser(strings.NewReader(`{"Code":"OK","Message":"Success","JobId":"job-1"}`)),
		}, nil
	}))

	out, err := cli.CreateWorkflowJob(context.Background(), &CreateWorkflowJobInput{
		Bucket:    "my-bucket",
		JobType:   WorkflowJobTypeDocConvert,
		JobDetail: map[string]string{"foo": "bar"},
	})
	if err != nil {
		t.Fatalf("CreateWorkflowJob error: %v", err)
	}
	if out.JobID != "job-1" {
		t.Fatalf("unexpected job id: %s", out.JobID)
	}
}

func TestQueryWorkflowJobsUsesPublicEndpoint(t *testing.T) {
	cli := &ClientV2{
		Client: Client{
			scheme:     "https",
			host:       "tos.example.com",
			urlMode:    urlModeDefault,
			recognizer: EmptyContentTypeRecognizer{},
		},
	}
	cli.SetHTTPTransport(workflowRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.Host != "my-bucket.tos.example.com" {
			t.Fatalf("unexpected host: %s", req.URL.Host)
		}
		if req.URL.Path != "/" {
			t.Fatalf("unexpected path: %s", req.URL.Path)
		}
		if req.URL.Query().Get("job_type") != string(WorkflowJobTypeTranscode) {
			t.Fatalf("unexpected job_type: %s", req.URL.Query().Get("job_type"))
		}
		if req.URL.Query().Get("job_id") != "job-2" {
			t.Fatalf("unexpected job_id: %s", req.URL.Query().Get("job_id"))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderRequestID: []string{"req-2"}},
			Body:       ioutil.NopCloser(strings.NewReader(`{"JobType":"Transcode","Items":[],"NextPageToken":"next-1"}`)),
		}, nil
	}))

	out, err := cli.QueryWorkflowJobs(context.Background(), &QueryWorkflowJobsInput{
		Bucket:  "my-bucket",
		JobType: string(WorkflowJobTypeTranscode),
		JobID:   "job-2",
	})
	if err != nil {
		t.Fatalf("QueryWorkflowJobs error: %v", err)
	}
	if out.JobType != string(WorkflowJobTypeTranscode) || out.NextPageToken != "next-1" {
		t.Fatalf("unexpected query output: %+v", out)
	}
}
