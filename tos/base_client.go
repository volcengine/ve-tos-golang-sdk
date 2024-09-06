package tos

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

func newBaseClient(c *Client) *baseClient {
	return &baseClient{Client: c}
}

type baseClient struct {
	*Client
}

func (cli *baseClient) PutObjectTagging(ctx context.Context, input *PutObjectTaggingInput, option ...Option) (*PutObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	data, contentMD5, err := marshalInput("PutObjectTaggingInput", putObjectTaggingInput{
		TagSet: input.TagSet,
	})
	if err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
		WithQuery("tagging", "").
		WithParams(*input).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutObjectTaggingOutput{RequestInfo: res.RequestInfo()}
	output.VersionID = res.Header.Get(HeaderVersionID)
	return &output, nil
}

func (cli *baseClient) GetObjectTagging(ctx context.Context, input *GetObjectTaggingInput, option ...Option) (*GetObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
		WithQuery("tagging", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetObjectTaggingOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	output.VersionID = res.Header.Get(HeaderVersionID)
	return &output, nil
}

func (cli *baseClient) DeleteObjectTagging(ctx context.Context, input *DeleteObjectTaggingInput, option ...Option) (*DeleteObjectTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
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

func (cli *baseClient) RestoreObject(ctx context.Context, input *RestoreObjectInput, option ...Option) (*RestoreObjectOutput, error) {

	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("RestoreObjectInput", restoreObjectInput{
		Days:                 input.Days,
		RestoreJobParameters: input.RestoreJobParameters,
	})
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
		WithParams(*input).
		WithQuery("restore", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(data), cli.roundTripper(http.StatusOK, http.StatusAccepted))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := RestoreObjectOutput{RequestInfo: res.RequestInfo()}
	return &output, nil

}

func (cli *baseClient) PutSymlink(ctx context.Context, input *PutSymlinkInput, option ...Option) (*PutSymlinkOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
		WithQuery("symlink", "").
		WithHeader(HeaderSymlinkTarget, url.QueryEscape(input.SymlinkTargetKey)).
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutSymlinkOutput{RequestInfo: res.RequestInfo()}
	output.VersionID = res.Header.Get(HeaderVersionID)
	return &output, nil
}

func (cli *baseClient) GetSymlink(ctx context.Context, input *GetSymlinkInput, option ...Option) (*GetSymlinkOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key, option...).
		WithQuery("symlink", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := &GetSymlinkOutput{RequestInfo: res.RequestInfo()}
	output.VersionID = res.Header.Get(HeaderVersionID)
	output.SymlinkTargetKey, _ = url.QueryUnescape(res.Header.Get(HeaderSymlinkTarget))
	output.SymlinkTargetBucket = res.Header.Get(HeaderSymlinkTargetBucket)
	lastModified, _ := time.ParseInLocation(http.TimeFormat, res.Header.Get(HeaderLastModified), time.UTC)
	output.LastModified = lastModified
	return output, nil
}

func (cli *baseClient) GetBucketACL(ctx context.Context, input *GetBucketACLInput, option ...Option) (*GetBucketACLOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("acl", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetBucketACLOutput{RequestInfo: res.RequestInfo()}
	marshalRes := bucketACL{}
	if err = marshalOutput(res, &marshalRes); err != nil {
		return nil, err
	}
	output.Grants = marshalRes.GrantList
	output.Owner = marshalRes.Owner
	output.BucketAclDelivered = marshalRes.BucketAclDelivered
	return &output, nil
}

func (cli *baseClient) PutBucketACL(ctx context.Context, input *PutBucketACLInput, option ...Option) (*PutBucketACLOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	reqBuilder := cli.newBuilder(input.Bucket, "").
		WithQuery("acl", "").
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		WithParams(*input)
	var reqData io.Reader

	if (input.Owner.ID != "" && len(input.Grants) != 0) || input.BucketAclDelivered {
		data, contentMD5, err := marshalInput("PutBucketACLInput", bucketACL{
			Owner:              input.Owner,
			GrantList:          input.Grants,
			BucketAclDelivered: input.BucketAclDelivered,
		})
		if err != nil {
			return nil, err
		}
		_ = reqBuilder.WithHeader(HeaderContentMD5, contentMD5)
		reqData = bytes.NewReader(data)
	}

	res, err := reqBuilder.Request(ctx, http.MethodPut, reqData, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutBucketACLOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}
