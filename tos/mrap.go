package tos

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
)

func (cli *ClientV2) CreateMultiRegionAccessPoint(ctx context.Context, input *CreateMultiRegionAccessPointInput) (*CreateMultiRegionAccessPointOutput, error) {

	if input == nil {
		return nil, InputIsNilClientError
	}
	data, contentMD5, err := marshalInput("CreateMultiRegionAccessPoint", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithQuery("name", input.Name).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodPost, bytes.NewReader(data), "/mrap", cli.roundTripper(http.StatusOK))

	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := CreateMultiRegionAccessPointOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetMultiRegionAccessPoint(ctx context.Context, input *GetMultiRegionAccessPointInput) (*GetMultiRegionAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithQuery("name", input.Name).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/mrap", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetMultiRegionAccessPointOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) DeleteMultiRegionAccessPoint(ctx context.Context, input *DeleteMultiRegionAccessPointInput) (*DeleteMultiRegionAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithQuery("name", input.Name).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodDelete, nil, "/mrap", cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := DeleteMultiRegionAccessPointOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) ListMultiRegionAccessPoints(ctx context.Context, input *ListMultiRegionAccessPointsInput) (*ListMultiRegionAccessPointsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	rb := cli.newControlBuilder(input.AccountID)
	if input.NextToken != "" {
		rb.WithQuery("nextToken", input.NextToken)
	}
	if input.MaxResults != 0 {
		rb.WithQuery("maxResults", strconv.Itoa(input.MaxResults))
	}
	res, err := rb.WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/mrap", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListMultiRegionAccessPointsOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
