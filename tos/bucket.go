package tos

import (
	"context"
	"net/http"
)

// Bucket create a Bucket handle
func (cli *Client) Bucket(bucket string) (*Bucket, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}

	return &Bucket{name: bucket, client: cli}, nil
}

type CreateBucketInput struct {
	Bucket           string `json:"Bucket,omitempty"`           // required
	ACL              string `json:"ACL,omitempty"`              // optional
	GrantFullControl string `json:"GrantFullControl,omitempty"` // optional
	GrantRead        string `json:"GrantRead,omitempty"`        // optional
	GrantReadAcp     string `json:"GrantReadAcp,omitempty"`     // optional
	GrantWrite       string `json:"GrantWrite,omitempty"`       // optional
	GrantWriteAcp    string `json:"GrantWriteAcp,omitempty"`    // optional
}

type CreateBucketOutput struct {
	RequestInfo `json:"-"`
	Location    string `json:"Location,omitempty"`
}

// CreateBucket create a bucket
func (cli *Client) CreateBucket(ctx context.Context, input *CreateBucketInput) (*CreateBucketOutput, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		WithHeader(HeaderACL, input.ACL).
		WithHeader(HeaderGrantFullControl, input.GrantFullControl).
		WithHeader(HeaderGrantRead, input.GrantRead).
		WithHeader(HeaderGrantReadAcp, input.GrantReadAcp).
		WithHeader(HeaderGrantWrite, input.GrantWrite).
		WithHeader(HeaderGrantWriteAcp, input.GrantWriteAcp).
		Request(ctx, http.MethodPut, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &CreateBucketOutput{
		RequestInfo: res.RequestInfo(),
		Location:    res.Header.Get(HeaderLocation),
	}, nil
}

type HeadBucketOutput struct {
	RequestInfo `json:"-"`
	Region      string `json:"Region,omitempty"`
}

// HeadBucket get some info of a bucket
func (cli *Client) HeadBucket(ctx context.Context, bucket string) (*HeadBucketOutput, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(bucket, "").
		Request(ctx, http.MethodHead, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &HeadBucketOutput{
		RequestInfo: res.RequestInfo(),
		Region:      res.Header.Get(HeaderBucketRegion),
	}, nil
}

type DeleteBucketOutput struct {
	RequestInfo `json:"-"`
}

// DeleteBucket delete a bucket
func (cli *Client) DeleteBucket(ctx context.Context, bucket string) (*DeleteBucketOutput, error) {
	if err := IsValidBucketName(bucket); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(bucket, "").
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &DeleteBucketOutput{RequestInfo: res.RequestInfo()}, nil
}

type ListedBucket struct {
	CreationDate     string `json:"CreationDate,omitempty"`
	Name             string `json:"Name,omitempty"`
	Location         string `json:"Location,omitempty"`
	ExtranetEndpoint string `json:"ExtranetEndpoint,omitempty"`
	IntranetEndpoint string `json:"IntranetEndpoint,omitempty"`
}

type ListedOwner struct {
	ID string `json:"ID,omitempty"`
}

type ListBucketsOutput struct {
	RequestInfo `json:"-"`
	Buckets     []ListedBucket `json:"Buckets,omitempty"`
	Owner       ListedOwner    `json:"Owner,omitempty"`
}

type ListBucketsInput struct{}

// ListBuckets list the buckets that the AK can access
func (cli *Client) ListBuckets(ctx context.Context, _ *ListBucketsInput) (*ListBucketsOutput, error) {
	res, err := cli.newBuilder("", "").
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListBucketsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
