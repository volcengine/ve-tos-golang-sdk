package tos

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

type postDPRoundTripperFunc func(*http.Request) (*http.Response, error)

func (f postDPRoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type postDPTrackedBody struct {
	*bytes.Reader
	closed bool
}

func (b *postDPTrackedBody) Close() error {
	b.closed = true
	return nil
}

func newPostDPTestClient(rt http.RoundTripper) *ClientV2 {
	cli := &ClientV2{Client: Client{
		scheme:     "https",
		host:       "tos.example.com",
		urlMode:    urlModeDefault,
		recognizer: EmptyContentTypeRecognizer{},
	}}
	cli.SetHTTPTransport(rt)
	return cli
}

func TestPostDataProcess_ImageSaveAsResponse(t *testing.T) {
	responseData := []byte(`{"bucket":"target-bucket","fileSize":"123","object":"target.jpg","status":"OK"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if _, ok := req.URL.Query()["x-tos-post-process"]; !ok {
			t.Fatal("missing x-tos-post-process query marker")
		}
		if contentType := req.Header.Get(HeaderContentType); contentType != "text/plain" {
			t.Fatalf("unexpected request Content-Type: %s", contentType)
		}
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		expected := "x-tos-post-process=image/resize,w_50&x-tos-save-object=dGFyZ2V0LmpwZw"
		if string(body) != expected {
			t.Fatalf("expected body %q, got %q", expected, body)
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{HeaderContentType: []string{"application/json; charset=utf-8"}},
			ContentLength: int64(len(responseData)),
			Body:          responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.jpg",
		PostProcess: "image/resize,w_50&x-tos-save-object=dGFyZ2V0LmpwZw",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput == nil || output.VideoProcessOutput != nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	imageOutput := output.ImageProcessOutput
	if imageOutput.Bucket != "target-bucket" || imageOutput.Object != "target.jpg" ||
		imageOutput.FileSize != 123 || imageOutput.Status != "OK" {
		t.Fatalf("unexpected image SaveAs output: %+v", imageOutput)
	}
	if imageOutput.Content != nil {
		t.Fatal("SaveAs response must not expose Content")
	}
	if !responseBody.closed {
		t.Fatal("JSON response body must be closed by SDK")
	}
}

func TestPostDataProcess_ImageContentResponse(t *testing.T) {
	imageData := []byte("jpeg-image-data")
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(imageData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{HeaderContentType: []string{"image/jpeg"}},
			ContentLength: int64(len(imageData)),
			Body:          responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.jpg", PostProcess: "image/resize,w_50",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput == nil || output.VideoProcessOutput != nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	imageOutput := output.ImageProcessOutput
	if imageOutput.ContentType != "image/jpeg" || imageOutput.ContentLength != int64(len(imageData)) {
		t.Fatalf("unexpected content metadata: %+v", imageOutput)
	}
	if responseBody.closed {
		t.Fatal("SDK closed image response before caller read it")
	}
	data, err := ioutil.ReadAll(imageOutput.Content)
	if err != nil || !bytes.Equal(data, imageData) {
		t.Fatalf("read image content: data=%q err=%v", data, err)
	}
	if err := imageOutput.Content.Close(); err != nil {
		t.Fatalf("close image content: %v", err)
	}
	if !responseBody.closed {
		t.Fatal("caller close did not close response body")
	}
	if strings.TrimSpace(output.RequestInfo.Header.Get(HeaderContentType)) != "image/jpeg" {
		t.Fatalf("unexpected response header: %v", output.RequestInfo.Header)
	}
}

func TestPostDataProcess_ImageJSONContentResponse(t *testing.T) {
	responseData := []byte(`{"Tags":[{"Name":"cat"}]}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Header:        http.Header{HeaderContentType: []string{"application/json; charset=utf-8"}},
			ContentLength: int64(len(responseData)),
			Body:          responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.jpg", PostProcess: "image/aitag",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput == nil || output.VideoProcessOutput != nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	imageOutput := output.ImageProcessOutput
	if imageOutput.ContentType != "application/json; charset=utf-8" || imageOutput.ContentLength != int64(len(responseData)) {
		t.Fatalf("unexpected content metadata: %+v", imageOutput)
	}
	if responseBody.closed {
		t.Fatal("SDK closed JSON image response before caller read it")
	}
	data, err := ioutil.ReadAll(imageOutput.Content)
	if err != nil || !bytes.Equal(data, responseData) {
		t.Fatalf("read JSON image content: data=%q err=%v", data, err)
	}
	if err := imageOutput.Content.Close(); err != nil {
		t.Fatalf("close JSON image content: %v", err)
	}
	if !responseBody.closed {
		t.Fatal("caller close did not close JSON image response body")
	}
}

func TestPostDataProcess_ImageInvalidSaveAsResponse(t *testing.T) {
	responseData := []byte(`{"bucket":"target-bucket","fileSize":"invalid","object":"target.jpg","status":"OK"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.jpg",
		PostProcess: "image/resize,w_50&x-tos-save-object=dGFyZ2V0LmpwZw",
	})
	if err == nil || output != nil {
		t.Fatalf("expected invalid fileSize error, output=%+v err=%v", output, err)
	}
	if _, ok := err.(*TosServerError); !ok {
		t.Fatalf("expected TosServerError, got %T: %v", err, err)
	}
	if !responseBody.closed {
		t.Fatal("invalid JSON response body must be closed by SDK")
	}
}

func TestPostDataProcess_VideoResponseUnchanged(t *testing.T) {
	responseData := []byte(`{"bucket":"target-bucket","object":"audio/output.pcm","object_size":"32","status":"OK"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.mp4", PostProcess: "video/convert,f_pcm",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput != nil || output.VideoProcessOutput == nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	videoOutput := output.VideoProcessOutput
	if videoOutput.PcmBucket != "target-bucket" || videoOutput.PcmObject != "audio/output.pcm" ||
		videoOutput.PcmObjectSize != 32 || videoOutput.PcmStatus != "OK" {
		t.Fatalf("unexpected video response: %+v", output.VideoProcessOutput)
	}
	if !responseBody.closed {
		t.Fatal("video response body must be closed by SDK")
	}
}

func TestPostDataProcess_VideoSnapshotsResponseUnchanged(t *testing.T) {
	responseData := []byte(`{"OutputBucket":"target-bucket","TotalFrameCount":2,"SuccFrameCount":2,"FailFrameCount":0}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.mp4",
		PostProcess: "video/snapshots,f_png,m_index,index_0|10|3000&x-tos-save-object=c25hcHNob3QucG5n",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput != nil || output.VideoProcessOutput == nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	videoOutput := output.VideoProcessOutput
	if videoOutput.OutputBucket != "target-bucket" || videoOutput.TotalFrameCount != 2 ||
		videoOutput.SuccFrameCount != 2 || videoOutput.FailFrameCount != 0 {
		t.Fatalf("unexpected snapshots response: %+v", videoOutput)
	}
	if !responseBody.closed {
		t.Fatal("snapshots response body must be closed")
	}
}

func TestPostDataProcess_VideoSaveAsResponse(t *testing.T) {
	tests := []struct {
		name     string
		process  string
		response string
		size     int64
	}{
		{
			name:     "convert",
			process:  "video/convert,f_mp4,vcodec_h264&x-tos-save-object=b3V0cHV0Lm1wNA",
			response: `{"bucket":"target-bucket","object":"output.mp4","object_size":"1024","status":"OK"}`,
			size:     1024,
		},
		{
			name:     "remux fileSize",
			process:  "video/remux,f_mp4&x-tos-save-object=cmVtdXgubXA0",
			response: `{"bucket":"target-bucket","object":"output.mp4","fileSize":"2048","status":"OK"}`,
			size:     2048,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseBody := &postDPTrackedBody{Reader: bytes.NewReader([]byte(tt.response))}
			cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{HeaderContentType: []string{"application/json"}},
					Body:       responseBody,
				}, nil
			}))

			output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
				Bucket: "source-bucket", Key: "source.mp4", PostProcess: tt.process,
			})
			if err != nil {
				t.Fatalf("PostDataProcess error: %v", err)
			}
			if output.ImageProcessOutput != nil || output.VideoProcessOutput == nil {
				t.Fatalf("unexpected typed output: %+v", output)
			}
			videoOutput := output.VideoProcessOutput
			if videoOutput.SaveAsBucket != "target-bucket" || videoOutput.SaveAsObject != "output.mp4" ||
				videoOutput.SaveAsObjectSize != tt.size || videoOutput.SaveAsStatus != "OK" {
				t.Fatalf("unexpected video SaveAs response: %+v", videoOutput)
			}
			if !responseBody.closed {
				t.Fatal("video SaveAs response body must be closed")
			}
		})
	}
}

func TestPostDataProcess_DocResponseDoesNotPopulateTypedOutput(t *testing.T) {
	responseData := []byte(`{"bucket":"target-bucket","object":"preview.png","status":"OK"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.docx", PostProcess: "doc-preview|sys/saveas,b_preview.png",
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput != nil || output.VideoProcessOutput != nil {
		t.Fatalf("doc response must not populate image or video output: %+v", output)
	}
	if !responseBody.closed {
		t.Fatal("doc response body must be closed")
	}
}

func TestPostDataProcess_LegacyJSONBodyWithLeadingWhitespace(t *testing.T) {
	legacyBody := "\n  {\"CanonicalUri\":\"video/convert\"}"
	responseData := []byte(`{"bucket":"target-bucket","object":"output.mp4","status":"OK"}`)
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(body) != strings.TrimSpace(legacyBody) {
			t.Fatalf("unexpected legacy JSON body: %q", body)
		}
		if contentType := req.Header.Get(HeaderContentType); contentType != "application/json" {
			t.Fatalf("unexpected Content-Type: %s", contentType)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       ioutil.NopCloser(bytes.NewReader(responseData)),
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.mp4", PostProcess: legacyBody,
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput != nil || output.VideoProcessOutput == nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	videoOutput := output.VideoProcessOutput
	if videoOutput.PcmBucket != "target-bucket" || videoOutput.PcmObject != "output.mp4" ||
		videoOutput.PcmStatus != "OK" {
		t.Fatalf("unexpected legacy video response: %+v", videoOutput)
	}
	if videoOutput.SaveAsBucket != "" || videoOutput.SaveAsObject != "" || videoOutput.SaveAsStatus != "" {
		t.Fatalf("legacy video response must not populate SaveAs fields: %+v", videoOutput)
	}
}

func TestPostDataProcess_LegacyVideoTranscodeResponsePreservesPCMFields(t *testing.T) {
	legacyBody := `{"Tag":"Transcode","Name":"legacy-audio-pcm","TranscodeConfig":{}}`
	responseData := []byte(`{"bucket":"target-bucket","object":"audio/output.pcm","object_size":"32","status":"OK"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(body) != legacyBody {
			t.Fatalf("unexpected legacy Transcode body: %q", body)
		}
		if contentType := req.Header.Get(HeaderContentType); contentType != "application/json" {
			t.Fatalf("unexpected Content-Type: %s", contentType)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcess(context.Background(), &PostDPInput{
		Bucket: "source-bucket", Key: "source.mp4", PostProcess: legacyBody,
	})
	if err != nil {
		t.Fatalf("PostDataProcess error: %v", err)
	}
	if output.ImageProcessOutput != nil || output.VideoProcessOutput == nil {
		t.Fatalf("unexpected typed output: %+v", output)
	}
	videoOutput := output.VideoProcessOutput
	if videoOutput.PcmBucket != "target-bucket" || videoOutput.PcmObject != "audio/output.pcm" ||
		videoOutput.PcmObjectSize != 32 || videoOutput.PcmStatus != "OK" {
		t.Fatalf("legacy Transcode response did not preserve PCM fields: %+v", videoOutput)
	}
	if videoOutput.SaveAsBucket != "" || videoOutput.SaveAsObject != "" ||
		videoOutput.SaveAsObjectSize != 0 || videoOutput.SaveAsStatus != "" {
		t.Fatalf("legacy Transcode response must not populate SaveAs fields: %+v", videoOutput)
	}
	if !responseBody.closed {
		t.Fatal("legacy Transcode response body must be closed")
	}
}

func TestPostDataProcessAsync_JobUsesJSON(t *testing.T) {
	jobBody, err := PostDataProcessAsyncHelper(context.Background(), PostDataProcessAsyncParams{
		JobType: ProcessJobTypeAudioConvert,
		JobBody: &AudioConvertJobBody{
			Input:              ProcessJobInput{Object: "audio/input.mp3"},
			Output:             ProcessJobOutput{Region: "cn-beijing", Bucket: "target-bucket", Object: "audio/output.wav"},
			AudioConvertConfig: AudioConvertJobConfig{ContainerFormat: "wav"},
		},
	})
	if err != nil {
		t.Fatalf("PostDataProcessAsyncHelper error: %v", err)
	}

	responseData := []byte(`{"Code":"OK","Message":"submitted","JobId":"job-123"}`)
	responseBody := &postDPTrackedBody{Reader: bytes.NewReader(responseData)}
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", req.Method)
		}
		if req.URL.Query().Get("job_type") != string(ProcessJobTypeAudioConvert) {
			t.Fatalf("unexpected job_type: %s", req.URL.Query().Get("job_type"))
		}
		if _, ok := req.URL.Query()["media_jobs"]; !ok {
			t.Fatal("missing media_jobs query marker")
		}
		if contentType := req.Header.Get(HeaderContentType); contentType != "application/json" {
			t.Fatalf("unexpected Content-Type: %s", contentType)
		}
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		expected := `{"Input":{"Object":"audio/input.mp3"},"Output":{"Region":"cn-beijing","Bucket":"target-bucket","Object":"audio/output.wav"},"AudioConvertConfig":{"ContainerFormat":"wav"}}`
		if string(body) != expected {
			t.Fatalf("unexpected async JSON body: got %s, want %s", body, expected)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{HeaderContentType: []string{"application/json"}},
			Body:       responseBody,
		}, nil
	}))

	output, err := cli.PostDataProcessAsync(context.Background(), &PostDPAsyncInput{
		Bucket:  "source-bucket",
		JobType: ProcessJobTypeAudioConvert,
		JobBody: jobBody,
	})
	if err != nil {
		t.Fatalf("PostDataProcessAsync error: %v", err)
	}
	if output.Code != "OK" || output.Message != "submitted" || output.JobId != "job-123" {
		t.Fatalf("unexpected async output: %+v", output)
	}
	if !responseBody.closed {
		t.Fatal("async JSON response body must be closed")
	}
}

func TestPostDataProcessAsync_JobBodyRequiresJobType(t *testing.T) {
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("request must not be sent without JobType")
		return nil, nil
	}))

	output, err := cli.PostDataProcessAsync(context.Background(), &PostDPAsyncInput{
		Bucket:  "source-bucket",
		JobBody: &AudioConvertJobBody{},
	})
	if err == nil || output != nil || !strings.Contains(err.Error(), "JobType is required") {
		t.Fatalf("expected JobType validation error, output=%+v err=%v", output, err)
	}
}

func TestPostDataProcessAsync_InvalidJSONJobBodyRejected(t *testing.T) {
	cli := newPostDPTestClient(postDPRoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		t.Fatal("request must not be sent with invalid JSON JobBody")
		return nil, nil
	}))

	output, err := cli.PostDataProcessAsync(context.Background(), &PostDPAsyncInput{
		Bucket:  "source-bucket",
		JobType: ProcessJobTypeAudioConvert,
		JobBody: `{"Input":`,
	})
	if err == nil || output != nil || !strings.Contains(err.Error(), "valid JSON object") {
		t.Fatalf("expected invalid JSON error, output=%+v err=%v", output, err)
	}
}
