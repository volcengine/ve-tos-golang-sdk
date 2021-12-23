package tos

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Bucket struct {
	name   string
	client *Client
}

type GetObjectOutput struct {
	RequestInfo  `json:"-"`
	ContentRange string        `json:"ContentRange,omitempty"`
	Content      io.ReadCloser `json:"-"`
	ObjectMeta
}

// GetObject get data and metadata of an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object
//    WithRange the range of content,
//    WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//    WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
func (bkt *Bucket) GetObject(ctx context.Context, objectKey string, options ...Option) (*GetObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	rb := bkt.client.newBuilder(bkt.name, objectKey, options...)
	res, err := rb.Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}

	output := GetObjectOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: rb.Header.Get(HeaderContentRange),
		Content:      res.Body,
	}
	output.ObjectMeta.fromResponse(res)
	return &output, nil
}

type HeadObjectOutput struct {
	RequestInfo  `json:"-"`
	ContentRange string `json:"ContentRange,omitempty"`
	ObjectMeta
}

// HeadObject get data and metadata of an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object
//    WithRange the range of content,
//    WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//    WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
func (bkt *Bucket) HeadObject(ctx context.Context, objectKey string, options ...Option) (*HeadObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	rb := bkt.client.newBuilder(bkt.name, objectKey, options...)
	res, err := rb.Request(ctx, http.MethodHead, nil, bkt.client.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := HeadObjectOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: rb.Header.Get(HeaderContentRange),
	}
	output.ObjectMeta.fromResponse(res)
	return &output, nil
}

func expectedCode(rb *requestBuilder) int {
	okCode := http.StatusOK
	if rb.Range != nil {
		okCode = http.StatusPartialContent
	}
	return okCode
}

type DeleteObjectOutput struct {
	RequestInfo  `json:"-"`
	DeleteMarker bool   `json:"DeleteMarker,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

// DeleteObject delete an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object will be deleted
func (bkt *Bucket) DeleteObject(ctx context.Context, objectKey string, options ...Option) (*DeleteObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		Request(ctx, http.MethodDelete, nil, bkt.client.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	deleteMarker, _ := strconv.ParseBool(res.Header.Get(HeaderDeleteMarker))
	return &DeleteObjectOutput{
		RequestInfo:  res.RequestInfo(),
		DeleteMarker: deleteMarker,
		VersionID:    res.Header.Get(HeaderVersionID),
	}, nil
}

type ObjectTobeDeleted struct {
	Key       string `json:"Key,omitempty"`
	VersionID string `json:"VersionId,omitempty"`
}

type DeleteMultiObjectsInput struct {
	Objects []ObjectTobeDeleted `json:"Objects,omitempty"`
	Quiet   bool                `json:"Quiet,omitempty"`
}

type Deleted struct {
	Key                   string `json:"Key,omitempty"`
	VersionID             string `json:"VersionId,omitempty"`
	DeleteMarker          *bool  `json:"DeleteMarker,omitempty"`
	DeleteMarkerVersionID string `json:"DeleteMarkerVersionId,omitempty"`
}

type DeleteError struct {
	Code      string `json:"Code,omitempty"`
	Message   string `json:"Message,omitempty"`
	Key       string `json:"Key,omitempty"`
	VersionID string `json:"VersionId,omitempty"`
}

type DeleteMultiObjectsOutput struct {
	RequestInfo `json:"-"`
	Deleted     []Deleted     `json:"Deleted,omitempty"` // 删除成功的Object列表
	Error       []DeleteError `json:"Error,omitempty"`   // 删除失败的Object列表
}

// DeleteMultiObjects delete multi-objects
//   input: the objects will be deleted
func (bkt *Bucket) DeleteMultiObjects(ctx context.Context, input *DeleteMultiObjectsInput, options ...Option) (*DeleteMultiObjectsOutput, error) {
	in, contentMD5, err := marshalInput("DeleteMultiObjectsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithHeader(HeaderContentMD5, contentMD5).
		WithQuery("delete", "").
		Request(ctx, http.MethodPost, bytes.NewReader(in), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteMultiObjectsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type PutObjectOutput struct {
	RequestInfo          `json:"-"`
	ETag                 string `json:"ETag,omitempty"`
	VersionID            string `json:"VersionId,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

// PutObject put an object
//   objectKey: the name of object
//   content: the content of object
//   options: WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
//     WithContentSHA256 set Content-Sha256,
//     WithContentMD5 set Content-MD5
//     WithExpires set Expires,
//     WithServerSideEncryptionCustomer set server side encryption options
//     WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl
//
// NOTICE: only content with a known length is supported now,
//   e.g, bytes.Buffer, bytes.Reader, strings.Reader, os.File, io.LimitedReader, net.Buffers.
//   if the parameter content(an io.Reader) is not one of these,
//   please use io.LimitReader(reader, length) to wrap this reader or use the WithContentLength option.
func (bkt *Bucket) PutObject(ctx context.Context, objectKey string, content io.Reader, options ...Option) (*PutObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		Request(ctx, http.MethodPut, content, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &PutObjectOutput{
		RequestInfo:          res.RequestInfo(),
		ETag:                 res.Header.Get(HeaderETag),
		VersionID:            res.Header.Get(HeaderVersionID),
		SSECustomerAlgorithm: res.Header.Get(HeaderSSECustomerAlgorithm),
		SSECustomerKeyMD5:    res.Header.Get(HeaderSSECustomerKeyMD5),
	}, nil
}

type AppendObjectOutput struct {
	RequestInfo      `json:"-"`
	ETag             string `json:"ETag,omitempty"`
	NextAppendOffset int64  `json:"NextAppendOffset,omitempty"`
}

// AppendObject append content at the tail of an appendable object
//   objectKey: the name of object
//   content: the content of object
//   offset: append position, equals to the current object-size
//   options: WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
//     WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl
//   above options only take effect when offset parameter is 0.
//     WithContentSHA256 set Content-Sha256,
//     WithContentMD5 set Content-MD5.
//
// NOTICE: only content with a known length is supported now,
//   e.g, bytes.Buffer, bytes.Reader, strings.Reader, os.File, io.LimitedReader, net.Buffers.
//   if the parameter content(an io.Reader) is not one of these,
//   please use io.LimitReader(reader, length) to wrap this reader or use the WithContentLength option.
func (bkt *Bucket) AppendObject(ctx context.Context, objectKey string, content io.Reader, offset int64, options ...Option) (*AppendObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithQuery("append", "").
		WithQuery("offset", strconv.FormatInt(offset, 10)).
		Request(ctx, http.MethodPost, content, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	nextOffset := res.Header.Get(HeaderNextAppendOffset)
	appendOffset, err := strconv.ParseInt(nextOffset, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("tos: server return unexptected Next-Append-Offset header %q", nextOffset)
	}

	return &AppendObjectOutput{
		RequestInfo:      res.RequestInfo(),
		ETag:             res.Header.Get(HeaderETag),
		NextAppendOffset: appendOffset,
	}, nil
}

type SetObjectMetaOutput struct {
	RequestInfo `json:"-"`
}

// SetObjectMeta overwrites metadata of the object
//   objectKey: the name of object
//   options: WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
//     WithVersionID which version of this object will be set
//
// NOTICE: SetObjectMeta always overwrites all previous metadata
func (bkt *Bucket) SetObjectMeta(ctx context.Context, objectKey string, options ...Option) (*SetObjectMetaOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithQuery("metadata", "").
		Request(ctx, http.MethodPost, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &SetObjectMetaOutput{RequestInfo: res.RequestInfo()}, nil
}

type ListObjectsInput struct {
	Prefix       string `json:"Prefix,omitempty"`
	Delimiter    string `json:"Delimiter,omitempty"`
	Marker       string `json:"Marker,omitempty"`
	MaxKeys      int    `json:"MaxKeys,omitempty"`
	Reverse      bool   `json:"Reverse,omitempty"`
	EncodingType string `json:"EncodingType,omitempty"` // "" or "url"
}

type ListedObject struct {
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	ETag         string `json:"ETag,omitempty"`
	Size         int64  `json:"Size,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Type         string `json:"Type,omitempty"`
}

type ListedCommonPrefix struct {
	Prefix string `json:"Prefix,omitempty"`
}

type ListObjectsOutput struct {
	RequestInfo    `json:"-"`
	Name           string               `json:"Name,omitempty"` // bucket name
	Prefix         string               `json:"Prefix,omitempty"`
	Marker         string               `json:"Marker,omitempty"`
	MaxKeys        int64                `json:"MaxKeys,omitempty"`
	NextMarker     string               `json:"NextMarker,omitempty"`
	Delimiter      string               `json:"Delimiter,omitempty"`
	IsTruncated    bool                 `json:"IsTruncated,omitempty"`
	EncodingType   string               `json:"EncodingType,omitempty"`
	CommonPrefixes []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents       []ListedObject       `json:"Contents,omitempty"`
}

// ListObjects list objects of a bucket
func (bkt *Bucket) ListObjects(ctx context.Context, input *ListObjectsInput, options ...Option) (*ListObjectsOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("prefix", input.Prefix).
		WithQuery("delimiter", input.Delimiter).
		WithQuery("marker", input.Marker).
		WithQuery("max-keys", strconv.Itoa(input.MaxKeys)).
		WithQuery("reverse", strconv.FormatBool(input.Reverse)).
		WithQuery("encoding-type", input.EncodingType).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListObjectsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type ListObjectVersionsInput struct {
	Prefix          string `json:"Prefix,omitempty"`
	Delimiter       string `json:"Delimiter,omitempty"`
	KeyMarker       string `json:"KeyMarker,omitempty"`
	VersionIDMarker string `json:"VersionIdMarker,omitempty"`
	MaxKeys         int    `json:"MaxKeys,omitempty"`
	EncodingType    string `json:"EncodingType,omitempty"` // "" or "url"
}

type ListedObjectVersion struct {
	ETag         string `json:"ETag,omitempty"`
	IsLatest     bool   `json:"IsLatest,omitempty"`
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	Size         int64  `json:"Size,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Type         string `json:"Type,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

type ListedDeleteMarkerEntry struct {
	IsLatest     bool   `json:"IsLatest,omitempty"`
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

type ListObjectVersionsOutput struct {
	RequestInfo         `json:"-"`
	Name                string                    `json:"Name,omitempty"` // bucket name
	Prefix              string                    `json:"Prefix,omitempty"`
	KeyMarker           string                    `json:"KeyMarker,omitempty"`
	VersionIDMarker     string                    `json:"VersionIdMarker,omitempty"`
	Delimiter           string                    `json:"Delimiter,omitempty"`
	EncodingType        string                    `json:"EncodingType,omitempty"`
	MaxKeys             int64                     `json:"MaxKeys,omitempty"`
	NextKeyMarker       string                    `json:"NextKeyMarker,omitempty"`
	NextVersionIDMarker string                    `json:"NextVersionIdMarker,omitempty"`
	IsTruncated         bool                      `json:"IsTruncated,omitempty"`
	CommonPrefixes      []ListedCommonPrefix      `json:"CommonPrefixes,omitempty"`
	Versions            []ListedObjectVersion     `json:"Versions,omitempty"`
	DeleteMarkers       []ListedDeleteMarkerEntry `json:"DeleteMarkers,omitempty"`
}

// ListObjectVersions list multi-version objects of a bucket
func (bkt *Bucket) ListObjectVersions(ctx context.Context, input *ListObjectVersionsInput, options ...Option) (*ListObjectVersionsOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("prefix", input.Prefix).
		WithQuery("delimiter", input.Delimiter).
		WithQuery("key-marker", input.KeyMarker).
		WithQuery("max-keys", strconv.Itoa(input.MaxKeys)).
		WithQuery("encoding-type", input.EncodingType).
		WithQuery("versions", "").
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListObjectVersionsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

//type ListObjectsV2Input struct {
//	Prefix            string `json:"Prefix,omitempty"`
//	Delimiter         string `json:"Delimiter,omitempty"`
//	Marker            string `json:"Marker,omitempty"`
//	MaxKeys           int    `json:"MaxKeys,omitempty"`
//	FetchOwner        bool   `json:"FetchOwner,omitempty"`
//	ContinuationToken string `json:"ContinuationToken,omitempty"`
//	Reverse           bool   `json:"Reverse,omitempty"`
//	EncodingType      string `json:"EncodingType,omitempty"` // "" or "url"
//}
//
//type ListObjectsV2Output struct {
//	RequestInfo           `json:"-"`
//	Name                  string               `json:"Name,omitempty"` // bucket name
//	Prefix                string               `json:"Prefix,omitempty"`
//	KeyCount              int64                `json:"KeyCount,omitempty"`
//	MaxKeys               int64                `json:"MaxKeys,omitempty"`
//	IsTruncated           bool                 `json:"IsTruncated,omitempty"`
//	ContinuationToken     string               `json:"ContinuationToken,omitempty"`
//	Delimiter             string               `json:"Delimiter,omitempty"`
//	EncodingType          string               `json:"EncodingType,omitempty"`
//	StartAfter            string               `json:"StartAfter,omitempty"`
//	NextContinuationToken string               `json:"NextContinuationToken,omitempty"`
//	CommonPrefixes        []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
//	Contents              []ListedObject       `json:"Contents,omitempty"`
//}
//
//// ListObjectsV2 list objects of a bucket
//func (bkt *Bucket) ListObjectsV2(ctx context.Context, input *ListObjectsV2Input, options ...Option) (*ListObjectsV2Output, error) {
//	res, err := bkt.client.newBuilder(bkt.name, "", options...).
//		WithQuery("prefix", input.Prefix).
//		WithQuery("delimiter", input.Delimiter).
//		WithQuery("marker", input.Marker).
//		WithQuery("max-keys", strconv.Itoa(input.MaxKeys)).
//		WithQuery("continuation-token", input.ContinuationToken).
//		WithQuery("fetch-owner", strconv.FormatBool(input.FetchOwner)).
//		WithQuery("reverse", strconv.FormatBool(input.Reverse)).
//		WithQuery("encoding-type", input.EncodingType).
//		WithQuery("list-type", "2").
//		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
//	if err != nil {
//		return nil, err
//	}
//	defer res.Close()
//
//	output := &ListObjectsV2Output{RequestInfo: res.RequestInfo()}
//	if err = marshalOutput(res.Body, output); err != nil {
//		return nil, err
//	}
//	return output, nil
//}
