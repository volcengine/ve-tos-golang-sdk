package tos

import (
	"bytes"
	"context"
	"net/http"
)

const (
	FetchTaskStateFailed  = "Failed"
	FetchTaskStateSucceed = "Succeed"
	FetchTaskStateExpired = "Expired"
	FetchTaskStateRunning = "Running"
)

type FetchObjectInput struct {
	URL           string `json:"URL,omitempty"`           // required
	Key           string `json:"Key,omitempty"`           // required
	IgnoreSameKey bool   `json:"IgnoreSameKey,omitempty"` // optional, default value is false
	ContentMD5    string `json:"ContentMD5,omitempty"`    // hex-encoded md5, optional
}

type FetchObjectOutput struct {
	RequestInfo `json:"-"`
	VersionID   string `json:"VersionId,omitempty"` // may be empty
	ETag        string `json:"ETag,omitempty"`
}

type fetchObjectInput struct {
	URL           string `json:"URL,omitempty"`           // required
	IgnoreSameKey bool   `json:"IgnoreSameKey,omitempty"` // optional, default value is false
	ContentMD5    string `json:"ContentMD5,omitempty"`    // base64-encoded md5, optional
}

// FetchObject fetch an object from specified URL
// options:
//    WithMeta set meta header(s)
//    WithServerSideEncryptionCustomer set server side encryption options
//    WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl
// Calling FetchObject will block util fetch operation is finished
func (bkt *Bucket) FetchObject(ctx context.Context, input *FetchObjectInput, options ...Option) (*FetchObjectOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("FetchObjectInput", &fetchObjectInput{
		URL:           input.URL,
		IgnoreSameKey: input.IgnoreSameKey,
		ContentMD5:    input.ContentMD5,
	})
	if err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("fetch", "").
		WithHeader(HeaderContentMD5, contentMD5).
		Request(ctx, http.MethodPost, bytes.NewReader(data), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}

	out := FetchObjectOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(out.RequestID, res.Body, &out); err != nil {
		return nil, err
	}

	out.VersionID = res.Header.Get(HeaderVersionID)
	return &out, nil
}

type PutFetchTaskInput struct {
	URL           string `json:"URL,omitempty"`           // required
	Key           string `json:"Object,omitempty"`        // required
	IgnoreSameKey bool   `json:"IgnoreSameKey,omitempty"` // optional, default value is false
	ContentMD5    string `json:"ContentMD5,omitempty"`    // hex-encoded md5, optional
}

type PutFetchTaskOutput struct {
	RequestInfo `json:"-"`
	TaskID      string `json:"TaskId,omitempty"`
}

func (bkt *Bucket) PutFetchTask(ctx context.Context, input *FetchObjectInput, options ...Option) (*PutFetchTaskOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutFetchTaskInput", input)
	if err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("fetchTask", "").
		WithHeader(HeaderContentMD5, contentMD5).
		Request(ctx, http.MethodPost, bytes.NewReader(data), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}

	out := PutFetchTaskOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(out.RequestID, res.Body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type GetFetchTaskInput struct {
	TaskID string `json:"taskID,omitempty"`
}

type GetFetchTaskOutput struct {
	RequestInfo `json:"-"`
	State       string `json:"State,omitempty"`
	// Cause       string `json:"Cause,omitempty"`
}

func (bkt *Bucket) GetFetchTask(ctx context.Context, input *GetFetchTaskInput, options ...Option) (*GetFetchTaskOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("fetchTask", "").
		WithQuery("taskId", input.TaskID).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))

	out := GetFetchTaskOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(out.RequestID, res.Body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
