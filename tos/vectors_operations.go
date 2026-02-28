package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (client *TosVectorsClient) PutVectors(ctx context.Context, input *PutVectorsInput) (*PutVectorsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tosvectors: missing IndexName", nil)
	}

	data, _, err := marshalInput("PutVectorsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/PutVectors", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &PutVectorsOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) GetVectors(ctx context.Context, input *GetVectorsInput) (*GetVectorsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tosvectors: missing IndexName", nil)
	}

	data, _, err := marshalInput("GetVectorsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/GetVectors", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := &GetVectorsOutput{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, output); err != nil {
		return nil, err
	}

	return output, nil
}

func (client *TosVectorsClient) DeleteVectors(ctx context.Context, input *DeleteVectorsInput) (*DeleteVectorsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tos: missing IndexName", nil)
	}

	data, _, err := marshalInput("DeleteVectorsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/DeleteVectors", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &DeleteVectorsOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) QueryVectors(ctx context.Context, input *QueryVectorsInput) (*QueryVectorsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tosvectors: missing IndexName", nil)
	}

	data, _, err := marshalInput("QueryVectorsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/QueryVectors", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := QueryVectorsOutput{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (client *TosVectorsClient) ListVectors(ctx context.Context, input *ListVectorsInput) (*ListVectorsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tosvectors: missing IndexName", nil)
	}

	data, _, err := marshalInput("ListVectorsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/ListVectors", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListVectorsOutput{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
