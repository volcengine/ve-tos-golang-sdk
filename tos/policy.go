package tos

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

type BucketPolicy struct {
	Policy string `json:"Policy,omitempty"`
}

type GetBucketPolicyOutput struct {
	RequestInfo `json:"-"`
	Policy      string `json:"Policy,omitempty"`
}

type PutBucketPolicyOutput struct {
	RequestInfo `json:"-"`
}

type DeleteBucketPolicyOutput struct {
	RequestInfo `json:"-"`
}

// GetBucketPolicy get bucket access policy
func (cli *Client) GetBucketPolicy(ctx context.Context, bucket string) (*GetBucketPolicyOutput, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(bucket, "").
		WithQuery("policy", "").
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &GetBucketPolicyOutput{
		RequestInfo: res.RequestInfo(),
		Policy:      string(data),
	}, nil
}

// PutBucketPolicy set bucket access policy
func (cli *Client) PutBucketPolicy(ctx context.Context, bucket string, policy *BucketPolicy) (*PutBucketPolicyOutput, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(bucket, "").
		WithQuery("policy", "").
		Request(ctx, http.MethodPut, strings.NewReader(policy.Policy), cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &PutBucketPolicyOutput{RequestInfo: res.RequestInfo()}, nil
}

// DeleteBucketPolicy delete bucket access policy
func (cli *Client) DeleteBucketPolicy(ctx context.Context, bucket string) (*DeleteBucketPolicyOutput, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(bucket, "").
		WithQuery("policy", "").
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &DeleteBucketPolicyOutput{RequestInfo: res.RequestInfo()}, nil
}
