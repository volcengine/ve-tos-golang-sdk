package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (cli *ClientV2) PutObjectTagging(ctx context.Context, input *PutObjectTaggingInput) (*PutObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	data, contentMD5, err := marshalInput("PutObjectTaggingInput", putObjectTaggingInput{
		TagSet: input.TagSet,
	})
	if err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("tagging", "").
		WithParams(*input).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutObjectTaggingOutput{RequestInfo: res.RequestInfo()}
	output.VersionID = res.Header.Get(HeaderVersionID)
	return &output, nil
}

func (cli *ClientV2) GetObjectTagging(ctx context.Context, input *GetObjectTaggingInput) (*GetObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("tagging", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetObjectTaggingOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	output.VersionID = res.Header.Get(HeaderVersionID)
	return &output, nil
}

func (cli *ClientV2) DeleteObjectTagging(ctx context.Context, input *DeleteObjectTaggingInput) (*DeleteObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("tagging", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteObjectTaggingOutput{RequestInfo: res.RequestInfo()}
	output.VersionID = res.Header.Get(HeaderVersionID)

	return &output, nil

}
