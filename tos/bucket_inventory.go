package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (cli *ClientV2) GetBucketInventory(ctx context.Context, input *GetBucketInventoryInput) (*GetBucketInventoryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("inventory", "").
		WithQuery("id", input.ID).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetBucketInventoryOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}

	return &output, nil
}

func (cli *ClientV2) PutBucketInventory(ctx context.Context, input *PutBucketInventoryInput) (*PutBucketInventoryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	data, contentMD5, err := marshalInput("PutBucketInventory", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("inventory", "").
		WithQuery("id", input.ID).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutBucketInventoryOutput{RequestInfo: res.RequestInfo()}
	return &output, nil

}

func (cli *ClientV2) DeleteBucketInventory(ctx context.Context, input *DeleteBucketInventoryInput) (*DeleteBucketInventoryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("inventory", "").
		WithQuery("id", input.ID).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteBucketInventoryOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) ListBucketInventory(ctx context.Context, input *ListBucketInventoryInput) (*ListBucketInventoryOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("inventory", "").
		WithQuery("continuation-token", input.ContinuationToken).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListBucketInventoryOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
