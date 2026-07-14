package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

const (
	objectPrefix          = "sdk-post-image-demo"
	c2paAgentName         = "Volcengine_Ark_CN"
	c2paAgentVersion      = "1.0.0"
	c2paModelName         = "post-image-process-sdk-demo"
	aigcContentProducer   = "001191441900MA7LWBFQ1510000"
	aigcContentPropagator = "001191441900MA7LWBFQ1520000"
)

type config struct {
	endpoint  string
	region    string
	accessKey string
	secretKey string
	bucket    string
	c2paAppID string
	proxy     string
	keep      bool
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	clientOptions := []tos.ClientOption{
		tos.WithRegion(cfg.region),
		tos.WithCredentials(tos.NewStaticCredentials(cfg.accessKey, cfg.secretKey)),
	}
	if cfg.proxy != "" {
		proxyURL, err := url.Parse(cfg.proxy)
		if err != nil {
			return fmt.Errorf("parse TOS_PROXY: %w", err)
		}
		clientOptions = append(clientOptions, tos.WithProxyFunc(http.ProxyURL(proxyURL)))
	}
	client, err := tos.NewClientV2(cfg.endpoint, clientOptions...)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	suffix := time.Now().UTC().Format("20060102T150405000000000")
	sourceKey := fmt.Sprintf("%s/%s-source.jpg", objectPrefix, suffix)
	watermarkKey := fmt.Sprintf("%s/%s-watermark.png", objectPrefix, suffix)
	resultKey := fmt.Sprintf("%s/%s-result.jpg", objectPrefix, suffix)
	businessID := makeBusinessID(resultKey)
	if !cfg.keep {
		defer cleanup(ctx, client, cfg.bucket, sourceKey, watermarkKey, resultKey)
	}

	watermark, err := makeWatermarkPNG()
	if err != nil {
		return err
	}
	watermarkRequestID, err := uploadObject(ctx, client, cfg.bucket, watermarkKey, "image/png", watermark)
	if err != nil {
		return fmt.Errorf("upload watermark image: %w", err)
	}
	fmt.Printf("watermark object: %s/%s (request_id=%s, size=%d)\n",
		cfg.bucket, watermarkKey, watermarkRequestID, len(watermark))

	source, err := makeSourceJPEG()
	if err != nil {
		return err
	}
	putRequestID, err := uploadObject(ctx, client, cfg.bucket, sourceKey, "image/jpeg", source)
	if err != nil {
		return fmt.Errorf("upload source image: %w", err)
	}
	fmt.Printf("source object: %s/%s (request_id=%s, size=%d)\n", cfg.bucket, sourceKey, putRequestID, len(source))

	process, err := buildPostProcess(ctx, cfg, watermarkKey, resultKey, businessID)
	if err != nil {
		return err
	}
	postOutput, err := client.PostDataProcess(ctx, &tos.PostDPInput{
		Bucket:      cfg.bucket,
		Key:         sourceKey,
		PostProcess: process,
	})
	if err != nil {
		return fmt.Errorf("post image process: %w", err)
	}
	if err := validatePostOutput(postOutput, cfg.bucket, resultKey); err != nil {
		return err
	}
	fmt.Printf("post image SaveAs: PASS (request_id=%s, status=%s, size=%d)\n",
		postOutput.RequestID, postOutput.ImageProcessOutput.Status, postOutput.ImageProcessOutput.FileSize)

	headOutput, err := client.HeadObjectV2(ctx, &tos.HeadObjectV2Input{Bucket: cfg.bucket, Key: resultKey})
	if err != nil {
		return fmt.Errorf("head result object: %w", err)
	}
	if headOutput.ContentLength <= 0 {
		return fmt.Errorf("result object is empty")
	}
	fmt.Printf("result object: %s/%s (request_id=%s, size=%d)\n",
		cfg.bucket, resultKey, headOutput.RequestID, headOutput.ContentLength)

	aigcBody, aigcRequestID, err := getImageMetadata(ctx, client, cfg.bucket, resultKey, enum.ImageOperationGetAIGCMetadata)
	if err != nil {
		return fmt.Errorf("get AIGC metadata: %w", err)
	}
	if err := validateAIGCMetadata(aigcBody, businessID); err != nil {
		return err
	}
	fmt.Printf("AIGC metadata: PASS (request_id=%s)\n", aigcRequestID)

	c2paBody, c2paRequestID, err := getImageMetadata(ctx, client, cfg.bucket, resultKey, enum.ImageOperationGetC2PAMetadata)
	if err != nil {
		return fmt.Errorf("get C2PA metadata: %w", err)
	}
	if err := validateC2PAMetadata(c2paBody, makeBlindWatermarkLogID(businessID)); err != nil {
		return err
	}
	fmt.Printf("C2PA metadata: PASS (request_id=%s)\n", c2paRequestID)
	fmt.Println("post image SaveAs demo: PASS")
	if cfg.keep {
		fmt.Println("cleanup skipped because TOS_DEMO_KEEP_OBJECTS is enabled")
	}
	return nil
}

func loadConfig() (config, error) {
	cfg := config{
		endpoint:  os.Getenv("TOS_ENDPOINT"),
		region:    os.Getenv("TOS_REGION"),
		accessKey: os.Getenv("TOS_ACCESS_KEY"),
		secretKey: os.Getenv("TOS_SECRET_KEY"),
		bucket:    os.Getenv("TOS_BUCKET"),
		c2paAppID: os.Getenv("TOS_C2PA_APP_ID"),
		proxy:     os.Getenv("TOS_PROXY"),
		keep:      strings.EqualFold(os.Getenv("TOS_DEMO_KEEP_OBJECTS"), "true"),
	}
	required := []struct {
		name  string
		value string
	}{
		{"TOS_ENDPOINT", cfg.endpoint},
		{"TOS_REGION", cfg.region},
		{"TOS_ACCESS_KEY", cfg.accessKey},
		{"TOS_SECRET_KEY", cfg.secretKey},
		{"TOS_BUCKET", cfg.bucket},
		{"TOS_C2PA_APP_ID", cfg.c2paAppID},
	}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			return config{}, fmt.Errorf("environment variable %s is required", item.name)
		}
	}
	return cfg, nil
}

func buildPostProcess(ctx context.Context, cfg config, watermarkKey, resultKey, businessID string) (string, error) {
	version, level := 2, 3
	transparency, offset := 60, 74
	logID := makeBlindWatermarkLogID(businessID)
	now := time.Now().UTC().Format(time.RFC3339)
	manifest := map[string]interface{}{
		"claim_generator": c2paAgentName + "/" + c2paAgentVersion,
		"assertions": []map[string]interface{}{{
			"label": "c2pa.actions.v2",
			"data": map[string]interface{}{
				"actions": []map[string]interface{}{{
					"action":            "c2pa.created",
					"when":              now,
					"digitalSourceType": "http://cv.iptc.org/newscodes/digitalsourcetype/trainedAlgorithmicMedia",
					"softwareAgent": map[string]string{
						"name": c2paAgentName, "version": c2paAgentVersion,
					},
					"parameters": map[string]string{
						"time": now, "name": c2paAgentName, "model_name": c2paModelName, "log_id": logID,
					},
				}},
			},
		}},
	}
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("marshal C2PA manifest: %w", err)
	}

	// C2PA 必须最后执行，避免签名后再修改图片导致签名失效。
	process, err := tos.PostDataProcessHelper(ctx, tos.PostDataProcessParams{
		PostProcessType: enum.PostProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{
			{
				Operation: enum.ImageOperationWatermark,
				WatermarkParams: &tos.ImageWatermarkParams{
					Image: encode(watermarkKey + "?x-tos-process=image/resize,p_140"),
					T:     &transparency,
					X:     &offset,
					Y:     &offset,
				},
			},
			{
				Operation: enum.ImageOperationBlindWatermark,
				BlindWatermarkParams: &tos.ImageBlindWatermarkParams{
					Text: logID, Version: &version, Level: &level,
				},
			},
			{
				Operation: enum.ImageOperationSetAIGCMetadata,
				AIGCMetadataParams: &tos.ImageAIGCMetadataParams{
					Label:             encode("1"),
					ContentProducer:   encode(aigcContentProducer),
					ProduceID:         encode(businessID),
					ContentPropagator: encode(aigcContentPropagator),
					PropagateID:       encode(businessID),
					ReservedCode1:     encode(businessID),
					ReservedCode2:     encode(businessID),
				},
			},
			{
				Operation: enum.ImageOperationSetC2PAMetadata,
				SetC2PAMetadataParams: &tos.ImageSetC2PAMetadataParams{
					AppID: cfg.c2paAppID, Manifest: encode(string(manifestJSON)),
				},
			},
		},
		SaveAsParams: &tos.SaveAsParams{SaveBucket: cfg.bucket, SaveObject: resultKey},
	})
	if err != nil {
		return "", fmt.Errorf("build post image process: %w", err)
	}
	return process, nil
}

func validatePostOutput(output *tos.PostDPOutput, bucket, object string) error {
	if output == nil || output.ImageProcessOutput == nil {
		return fmt.Errorf("post image response does not contain ImageProcessOutput")
	}
	if output.VideoProcessOutput != nil {
		return fmt.Errorf("post image response unexpectedly contains VideoProcessOutput")
	}
	imageOutput := output.ImageProcessOutput
	if imageOutput.Bucket != bucket || imageOutput.Object != object {
		return fmt.Errorf("unexpected SaveAs target: got %s/%s, want %s/%s",
			imageOutput.Bucket, imageOutput.Object, bucket, object)
	}
	if !strings.EqualFold(imageOutput.Status, "OK") {
		return fmt.Errorf("unexpected SaveAs status: %s", imageOutput.Status)
	}
	if imageOutput.FileSize <= 0 {
		return fmt.Errorf("SaveAs response has invalid file size: %d", imageOutput.FileSize)
	}
	return nil
}

func getImageMetadata(ctx context.Context, client *tos.ClientV2, bucket, object string, operation enum.ImageOperation) ([]byte, string, error) {
	process, err := tos.GetDataProcessHelper(ctx, tos.GetDataProcessParams{
		GetProcessType:     enum.GetProcessTypeImage,
		ImageProcessParams: []tos.ImageProcessParams{{Operation: operation}},
	})
	if err != nil {
		return nil, "", err
	}
	output, err := client.GetDataProcess(ctx, &tos.GetObjectV2Input{Bucket: bucket, Key: object, Process: process})
	if err != nil {
		return nil, "", err
	}
	if output.Content == nil {
		return nil, output.RequestID, fmt.Errorf("metadata response body is nil")
	}
	defer output.Content.Close()
	body, err := ioutil.ReadAll(output.Content)
	if err != nil {
		return nil, output.RequestID, err
	}
	if output.StatusCode != http.StatusOK {
		return nil, output.RequestID, fmt.Errorf("unexpected status code: %d", output.StatusCode)
	}
	return body, output.RequestID, nil
}

func validateAIGCMetadata(body []byte, businessID string) error {
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("AIGC metadata is not valid JSON: %w", err)
	}
	metadata := response
	if nested, ok := response["AIGC"].(map[string]interface{}); ok {
		metadata = nested
	}
	want := map[string]string{
		"Label":             "1",
		"ContentProducer":   aigcContentProducer,
		"ProduceID":         businessID,
		"ContentPropagator": aigcContentPropagator,
		"PropagateID":       businessID,
		"ReservedCode1":     businessID,
		"ReservedCode2":     businessID,
	}
	for key, value := range want {
		if metadata[key] != value {
			return fmt.Errorf("unexpected AIGC metadata %s: got %v, want %s; response=%s",
				key, metadata[key], value, body)
		}
	}
	return nil
}

func validateC2PAMetadata(body []byte, logID string) error {
	var response interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("C2PA metadata is not valid JSON: %w", err)
	}
	text := string(body)
	for _, expected := range []string{"c2pa.actions.v2", c2paAgentName, c2paModelName, logID} {
		if !strings.Contains(text, expected) {
			return fmt.Errorf("C2PA metadata does not contain %q", expected)
		}
	}
	return nil
}

func uploadObject(ctx context.Context, client *tos.ClientV2, bucket, key, contentType string, content []byte) (string, error) {
	output, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:        bucket,
			Key:           key,
			ContentLength: int64(len(content)),
			ContentType:   contentType,
		},
		Content: bytes.NewReader(content),
	})
	if err != nil {
		return "", err
	}
	return output.RequestID, nil
}

func makeSourceJPEG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 601, 601))
	for y := 0; y < 601; y++ {
		for x := 0; x < 601; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(32 + x*191/600),
				G: uint8(48 + y*175/600),
				B: uint8(224 - (x+y)*128/1200),
				A: 255,
			})
		}
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("encode source JPEG: %w", err)
	}
	return output.Bytes(), nil
}

func makeWatermarkPNG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 280, 96))
	for y := 0; y < 96; y++ {
		for x := 0; x < 280; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8(32 + x*160/279),
				G: uint8(160 + y*64/95),
				B: 240,
				A: 220,
			})
		}
	}
	var output bytes.Buffer
	if err := png.Encode(&output, img); err != nil {
		return nil, fmt.Errorf("encode watermark PNG: %w", err)
	}
	return output.Bytes(), nil
}

func makeBusinessID(object string) string {
	sum := sha256.Sum256([]byte(object))
	return "02" + hex.EncodeToString(sum[:27])
}

func makeBlindWatermarkLogID(businessID string) string {
	sum := sha256.Sum256([]byte(businessID))
	payload := make([]byte, 18)
	copy(payload[:4], []byte{0x01, 0x32, 0x00, 0x04})
	copy(payload[4:14], sum[:10])
	return base64.RawURLEncoding.EncodeToString(payload)
}

func cleanup(ctx context.Context, client *tos.ClientV2, bucket string, objects ...string) {
	for _, object := range objects {
		if _, err := client.DeleteObjectV2(ctx, &tos.DeleteObjectV2Input{Bucket: bucket, Key: object}); err != nil {
			fmt.Fprintf(os.Stderr, "cleanup warning for %s/%s: %v\n", bucket, object, err)
		}
	}
}

func encode(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}
