package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPostProcessMatchesBeijingShape(t *testing.T) {
	cfg := config{bucket: "xsh-test-11", c2paAppID: "test-c2pa-app"}
	watermarkKey := "sdk-post-image-demo/watermark.png"
	resultKey := "sdk-post-image-demo/result.jpg"
	businessID := makeBusinessID(resultKey)

	process, err := buildPostProcess(context.Background(), cfg, watermarkKey, resultKey, businessID)
	if err != nil {
		t.Fatalf("build post process: %v", err)
	}

	sections := strings.Split(process, "&")
	operations := strings.Split(sections[0], "/")
	if len(operations) != 5 {
		t.Fatalf("unexpected operations: %q", sections[0])
	}
	wantNames := []string{"image", "watermark", "blindwatermark", "setaigcmetadata", "setc2pametadata"}
	for i, want := range wantNames {
		if got := strings.SplitN(operations[i], ",", 2)[0]; got != want {
			t.Fatalf("operation %d: got %q, want %q", i, got, want)
		}
	}

	watermark := operations[1]
	for _, want := range []string{"t_60", "x_74", "y_74"} {
		if !strings.Contains(watermark, want) {
			t.Fatalf("watermark operation does not contain %q: %s", want, watermark)
		}
	}
	encodedImage := operationParam(watermark, "image")
	decodedImage, err := base64.RawURLEncoding.DecodeString(encodedImage)
	if err != nil {
		t.Fatalf("decode watermark image: %v", err)
	}
	if want := watermarkKey + "?x-tos-process=image/resize,p_140"; string(decodedImage) != want {
		t.Fatalf("watermark image: got %q, want %q", decodedImage, want)
	}

	blindWatermark := operations[2]
	for _, want := range []string{"text_" + makeBlindWatermarkLogID(businessID), "version_2", "level_3"} {
		if !strings.Contains(blindWatermark, want) {
			t.Fatalf("blind watermark operation does not contain %q: %s", want, blindWatermark)
		}
	}

	aigc := operations[3]
	for _, want := range []string{
		"Label_" + encode("1"),
		"ContentProducer_" + encode(aigcContentProducer),
		"ProduceID_" + encode(businessID),
		"ReservedCode2_" + encode(businessID),
	} {
		if !strings.Contains(aigc, want) {
			t.Fatalf("AIGC operation does not contain %q: %s", want, aigc)
		}
	}

	c2pa := operations[4]
	if !strings.Contains(c2pa, "AppID_"+cfg.c2paAppID) {
		t.Fatalf("C2PA operation does not contain AppID: %s", c2pa)
	}
	manifestBytes, err := base64.RawURLEncoding.DecodeString(operationParam(c2pa, "Manifest"))
	if err != nil {
		t.Fatalf("decode C2PA manifest: %v", err)
	}
	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("unmarshal C2PA manifest: %v", err)
	}
	manifestText := string(manifestBytes)
	for _, want := range []string{c2paAgentName, c2paAgentVersion, c2paModelName, makeBlindWatermarkLogID(businessID)} {
		if !strings.Contains(manifestText, want) {
			t.Fatalf("C2PA manifest does not contain %q: %s", want, manifestText)
		}
	}

	if len(sections) != 3 || sections[1] != "x-tos-save-object="+encode(resultKey) ||
		sections[2] != "x-tos-save-bucket="+encode(cfg.bucket) {
		t.Fatalf("unexpected SaveAs parameters: %q", process)
	}
}

func TestBlindWatermarkLogIDMatchesProductionShape(t *testing.T) {
	encoded := makeBlindWatermarkLogID("business-id")
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("decode blind watermark payload: %v", err)
	}
	if len(payload) != 18 {
		t.Fatalf("payload length: got %d, want 18", len(payload))
	}
	if got := payload[:4]; string(got) != string([]byte{0x01, 0x32, 0x00, 0x04}) {
		t.Fatalf("unexpected payload prefix: %x", got)
	}
	if got := payload[14:]; string(got) != string([]byte{0, 0, 0, 0}) {
		t.Fatalf("unexpected payload suffix: %x", got)
	}
}

func operationParam(operation, name string) string {
	prefix := name + "_"
	for _, part := range strings.Split(operation, ",") {
		if strings.HasPrefix(part, prefix) {
			return strings.TrimPrefix(part, prefix)
		}
	}
	return ""
}
