package tos

import (
	"bytes"
	"context"
	"net/http"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func (cli *ClientV2) PutBucketObjectLock(ctx context.Context, input *PutBucketObjectLockInput) (*PutBucketObjectLockOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutBucketObjectLock", input.Configuration)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("object-lock", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutBucketObjectLockOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetBucketObjectLock(ctx context.Context, input *GetBucketObjectLockInput) (*GetBucketObjectLockOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	req := cli.newBuilder(input.Bucket, "").
		WithQuery("object-lock", "").
		WithRetry(nil, StatusCodeClassifier{})

	res, err := req.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var conf ObjectLockConfiguration
	if err = marshalOutput(res, &conf); err != nil {
		return nil, err
	}
	output := GetBucketObjectLockOutput{RequestInfo: res.RequestInfo(), Configuration: conf}
	return &output, nil
}

type PutBucketObjectLockInput struct {
	Bucket        string
	Configuration ObjectLockConfiguration
}

type ObjectLockConfiguration struct {
	ObjectLockEnabled enum.StatusType `json:"ObjectLockEnabled"`
	Rule              *RetentionRule  `json:"Rule,omitempty"`
}

type RetentionRule struct {
	DefaultRetention DefaultRetention `json:"DefaultRetention"`
}

type DefaultRetention struct {
	Days  int64              `json:"Days,omitempty"`
	Mode  enum.RetentionMode `json:"Mode"`
	Years int64              `json:"Years,omitempty"`
}

type PutBucketObjectLockOutput struct {
	RequestInfo
}
type GetBucketObjectLockInput struct {
	Bucket string
}
type GetBucketObjectLockOutput struct {
	RequestInfo
	Configuration ObjectLockConfiguration
}
