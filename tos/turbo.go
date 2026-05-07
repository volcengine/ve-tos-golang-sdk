package tos

import (
	"context"
	"io"
	"net/http"
	"strconv"
)

// OpenTurbo opens a turbo object for write or create.
func (cli *ClientV2) OpenTurbo(ctx context.Context, input *OpenTurboInput) (*OpenTurboOutput, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	contentLength := input.ContentLength
	if contentLength <= 0 && input.Content != nil {
		contentLength = tryResolveLength(input.Content)
	}
	var content io.Reader = input.Content
	if content != nil {
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, nil)
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("openturbo", "").
		WithQuery("mode", strconv.Itoa(int(input.Mode))).
		WithParams(*input)
	if contentLength >= 0 {
		rb.WithContentLength(contentLength)
	}
	rb.WithRetry(nil, NoRetryClassifier{})
	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPost, content, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	nextOffset := res.Header.Get(HeaderTurboOffset)
	turboOffset, _ := strconv.ParseInt(nextOffset, 10, 64)
	turboToken := res.Header.Get(HeaderTurboToken)
	opened := res.Header.Get(HeaderTurboOpenedCount)
	turboOpened, _ := strconv.ParseInt(opened, 10, 64)
	return &OpenTurboOutput{
		RequestInfo:     res.RequestInfo(),
		NextTurboOffset: turboOffset,
		TurboToken:      turboToken,
		OpenedCount:     turboOpened,
	}, nil
}

// AppendTurbo appends data to a turbo object.
func (cli *ClientV2) AppendTurbo(ctx context.Context, input *AppendTurboInput) (*AppendTurboOutput, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	contentLength := input.ContentLength
	if contentLength <= 0 && input.Content != nil {
		contentLength = tryResolveLength(input.Content)
	}
	var content io.Reader = input.Content
	if content != nil {
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, nil)
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("appendturbo", "").
		WithParams(*input)
	if contentLength >= 0 {
		rb.WithContentLength(contentLength)
	}
	rb.WithRetry(nil, NoRetryClassifier{})
	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPost, content, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	nextOffset := res.Header.Get(HeaderTurboOffset)
	turboOffset, _ := strconv.ParseInt(nextOffset, 10, 64)
	turboToken := res.Header.Get(HeaderTurboToken)
	return &AppendTurboOutput{
		RequestInfo:     res.RequestInfo(),
		NextTurboOffset: turboOffset,
		TurboToken:      turboToken,
	}, nil
}

// CloseTurbo closes a turbo object.
func (cli *ClientV2) CloseTurbo(ctx context.Context, input *CloseTurboInput) (*CloseTurboOutput, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("closeturbo", "").
		WithQuery("mode", strconv.Itoa(int(input.Mode))).
		WithParams(*input).
		WithRetry(nil, NoRetryClassifier{})
	res, err := rb.Request(ctx, http.MethodPost, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	opened := res.Header.Get(HeaderTurboOpenedCount)
	turboOpened, _ := strconv.ParseInt(opened, 10, 64)
	return &CloseTurboOutput{
		RequestInfo: res.RequestInfo(),
		OpenedCount: turboOpened,
	}, nil
}

// ListOpenedTurbo lists opened turbo objects in a bucket.
func (cli *ClientV2) ListOpenedTurbo(ctx context.Context, input *ListOpenedTurboInput) (*ListOpenedTurboOutput, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	rb := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("listopenedturbo", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{})
	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListOpenedTurboOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
