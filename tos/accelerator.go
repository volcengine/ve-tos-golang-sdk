package tos

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
)

func (cli *ClientV2) PutAccelerator(ctx context.Context, input *PutAcceleratorInput) (*PutAcceleratorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	data, contentMD5, err := marshalInput("PutAccelerator", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithQuery("name", input.AcceleratorName).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodPost, bytes.NewReader(data), "/accelerator", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutAcceleratorOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) GetAccelerator(ctx context.Context, input *GetAcceleratorInput) (*GetAcceleratorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithQuery("id", input.AcceleratorID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accelerator", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetAcceleratorOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) ListAccelerator(ctx context.Context, input *ListAcceleratorInput) (*ListAcceleratorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	rb := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput)
	if input.MaxResults != 0 {
		rb.WithQuery("maxResults", strconv.Itoa(input.MaxResults))
	}
	if input.NextToken != "" {
		rb.WithQuery("nextToken", input.NextToken)
	}

	res, err := rb.WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accelerator", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListAcceleratorOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) DeleteAccelerator(ctx context.Context, input *DeleteAcceleratorInput) (*DeleteAcceleratorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithQuery("id", input.AcceleratorID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodDelete, nil, "/accelerator", cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := DeleteAcceleratorOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}
