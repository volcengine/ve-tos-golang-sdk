package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (cli *ClientV2) PutBucketNotification(ctx context.Context, input *PutBucketNotificationInput) (*PutBucketNotificationOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutBucketNotification", putBucketNotificationInput{CloudFunctionConfigurations: input.CloudFunctionConfigurations, RocketMQConfigurations: input.RocketMQConfigurations})
	if err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("notification", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutBucketNotificationOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetBucketNotification(ctx context.Context, input *GetBucketNotificationInput) (*GetBucketNotificationOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("notification", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetBucketNotificationOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) PutBucketNotificationType2(ctx context.Context, input *PutBucketNotificationType2Input) (*PutBucketNotificationType2Output, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutBucketNotificationType2", putBucketNotificationType2Input{Rules: input.Rules, Version: input.Version})
	if err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("notification_v2", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutBucketNotificationType2Output{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetBucketNotificationType2(ctx context.Context, input *GetBucketNotificationType2Input) (*GetBucketNotificationType2Output, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("notification_v2", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetBucketNotificationType2Output{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
