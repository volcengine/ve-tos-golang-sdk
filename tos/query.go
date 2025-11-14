package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (cli *ClientV2) SimpleQuery(ctx context.Context, input *SimpleQueryInput) (*SimpleQueryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	data, contentMD5, err := marshalInput("SimpleQuery", input)
	if err != nil {
		return nil, err
	}
	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		WithQuery("mode", "SimpleQuery").
		WithHeader(HeaderContentMD5, contentMD5).
		RequestControl(ctx, http.MethodPost, bytes.NewReader(data), "/datasetquery", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := SimpleQueryOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) SemanticQuery(ctx context.Context, input *SemanticQueryInput) (*SemanticQueryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	data, contentMD5, err := marshalInput("SemanticQuery", input)
	if err != nil {
		return nil, err
	}
	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		WithQuery("mode", "SemanticQuery").
		WithHeader(HeaderContentMD5, contentMD5).
		RequestControl(ctx, http.MethodPost, bytes.NewReader(data), "/datasetquery", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := SemanticQueryOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
