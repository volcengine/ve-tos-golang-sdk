package tos

import (
	"bytes"
	"context"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"net/http"
)

type putBucketAccessMonitor struct {
	Status enum.StatusType `json:"Status"`
}

func (cli *ClientV2) PutBucketAccessMonitor(ctx context.Context, input *PutBucketAccessMonitorInput) (*PutBucketAccessMonitorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutBucketAccessMonitor", putBucketAccessMonitor{
		Status: input.Status,
	})
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("accessmonitor", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))

	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutBucketAccessMonitorOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetBucketAccessMonitor(ctx context.Context, input *GetBucketAccessMonitorInput) (*GetBucketAccessMonitorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("accessmonitor", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetBucketAccessMonitorOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
