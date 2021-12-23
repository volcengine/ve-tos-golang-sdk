package tos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
)

type CreateMultipartUploadOutput struct {
	RequestInfo          `json:"-"`
	Bucket               string `json:"Bucket,omitempty"`
	Key                  string `json:"Key,omitempty"`
	UploadID             string `json:"UploadId,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

type multipartUpload struct {
	Bucket   string `json:"Bucket,omitempty"`
	Key      string `json:"Key,omitempty"`
	UploadID string `json:"UploadId,omitempty"`
}

// CreateMultipartUpload create a multipart upload operation
//   objectKey: the name of object
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
func (bkt *Bucket) CreateMultipartUpload(ctx context.Context, objectKey string, options ...Option) (*CreateMultipartUploadOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithQuery("uploads", "").
		Request(ctx, http.MethodPost, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var upload multipartUpload
	if err = marshalOutput(res.RequestInfo().RequestID, res.Body, &upload); err != nil {
		return nil, err
	}

	return &CreateMultipartUploadOutput{
		RequestInfo:          res.RequestInfo(),
		Bucket:               upload.Bucket,
		Key:                  upload.Key,
		UploadID:             upload.UploadID,
		SSECustomerAlgorithm: res.Header.Get(HeaderSSECustomerAlgorithm),
		SSECustomerKeyMD5:    res.Header.Get(HeaderSSECustomerKeyMD5),
	}, nil
}

type UploadPartInput struct {
	Key        string    `json:"Key,omitempty"`
	UploadID   string    `json:"UploadId,omitempty"`
	PartSize   int64     `json:"PartSize,omitempty"`
	PartNumber int       `json:"PartNumber,omitempty"`
	Content    io.Reader `json:"-"`
}

type UploadPartOutput struct {
	RequestInfo          `json:"-"`
	PartNumber           int    `json:"PartNumber,omitempty"`
	ETag                 string `json:"ETag,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

func (up *UploadPartOutput) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: up.PartNumber, ETag: up.ETag}
}

// UploadPart upload a part for a multipart upload operation
// input: the parameters, some fields is required, e.g., Key, UploadID, PartSize, PartNumber and PartNumber
func (bkt *Bucket) UploadPart(ctx context.Context, input *UploadPartInput, options ...Option) (*UploadPartOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("uploadId", input.UploadID).
		WithQuery("partNumber", strconv.Itoa(input.PartNumber)).
		Request(ctx, http.MethodPut, io.LimitReader(input.Content, input.PartSize), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &UploadPartOutput{
		RequestInfo:          res.RequestInfo(),
		PartNumber:           input.PartNumber,
		ETag:                 res.Header.Get(HeaderETag),
		SSECustomerAlgorithm: res.Header.Get(HeaderSSECustomerAlgorithm),
		SSECustomerKeyMD5:    res.Header.Get(HeaderSSECustomerKeyMD5),
	}, nil
}

type CompleteMultipartUploadOutput struct {
	RequestInfo `json:"-"`
	VersionID   string `json:"VersionId,omitempty"`
}

type uploadedPart struct {
	PartNumber int    `json:"PartNumber"`
	ETag       string `json:"ETag"`
}

type uploadedParts []uploadedPart

func (p uploadedParts) Less(i, j int) bool { return p[i].PartNumber < p[j].PartNumber }
func (p uploadedParts) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p uploadedParts) Len() int           { return len(p) }

type completeMultipartUploadInput struct {
	Parts uploadedParts `json:"Parts"`
}

type MultipartUploadedPart interface {
	uploadedPart() uploadedPart
}

type CompleteMultipartUploadInput struct {
	Key           string                  `json:"Key,omitempty"`
	UploadID      string                  `json:"UploadId,omitempty"`
	UploadedParts []MultipartUploadedPart `json:"UploadedParts,omitempty"`
}

// CompleteMultipartUpload complete a multipart upload operation
//   input: input.Key the object name,
//     input.UploadID the uploadID got from CreateMultipartUpload
//     input.UploadedParts upload part output got from UploadPart or UploadPartCopy
func (bkt *Bucket) CompleteMultipartUpload(ctx context.Context, input *CompleteMultipartUploadInput, options ...Option) (*CompleteMultipartUploadOutput, error) {
	multipart := completeMultipartUploadInput{Parts: make(uploadedParts, 0, len(input.UploadedParts))}
	for _, p := range input.UploadedParts {
		multipart.Parts = append(multipart.Parts, p.uploadedPart())
	}

	sort.Sort(multipart.Parts)
	data, err := json.Marshal(&multipart)
	if err != nil {
		return nil, fmt.Errorf("tos: marshal uploadParts err: %s", err.Error())
	}

	res, err := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("uploadId", input.UploadID).
		Request(ctx, http.MethodPost, bytes.NewReader(data), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &CompleteMultipartUploadOutput{
		RequestInfo: res.RequestInfo(),
		VersionID:   res.Header.Get(HeaderVersionID),
	}, nil
}

type AbortMultipartUploadInput struct {
	Key      string `json:"Key,omitempty"`
	UploadID string `json:"UploadId,omitempty"`
}

type AbortMultipartUploadOutput struct {
	RequestInfo `json:"-"`
}

// AbortMultipartUpload abort a multipart upload operation
func (bkt *Bucket) AbortMultipartUpload(ctx context.Context, input *AbortMultipartUploadInput, options ...Option) (*AbortMultipartUploadOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("uploadId", input.UploadID).
		Request(ctx, http.MethodDelete, nil, bkt.client.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &AbortMultipartUploadOutput{RequestInfo: res.RequestInfo()}, nil
}

type Owner struct {
	ID          string `json:"ID,omitempty"`
	DisplayName string `json:"DisplayName,omitempty"`
}

type UploadedPart struct {
	PartNumber   int32  `json:"PartNumber,omitempty"`   // Part编号
	LastModified string `json:"LastModified,omitempty"` // 最后一次修改时间
	ETag         string `json:"ETag,omitempty"`         // ETag
	Size         int64  `json:"Size,omitempty"`         // Part大小
}

type ListUploadedPartsInput struct {
	Key              string `json:"Key,omitempty"`
	UploadID         string `json:"UploadId,omitempty"`
	MaxParts         int    `json:"MaxParts,omitempty"`             // 最大Part个数
	PartNumberMarker int    `json:"NextPartNumberMarker,omitempty"` // 起始Part的位置
}

type ListUploadedPartsOutput struct {
	RequestInfo          `json:"-"`
	Bucket               string         `json:"Bucket,omitempty"`               // Bucket名称
	Key                  string         `json:"Key,omitempty"`                  // Object名称
	UploadID             string         `json:"UploadId,omitempty"`             // 上传ID
	PartNumberMarker     int            `json:"PartNumberMarker,omitempty"`     // 当前页起始位置
	NextPartNumberMarker int            `json:"NextPartNumberMarker,omitempty"` // 下一个Part的位置
	MaxParts             int            `json:"MaxParts,omitempty"`             // 最大Part个数
	IsTruncated          bool           `json:"IsTruncated,omitempty"`          // 是否完全上传完成
	StorageClass         string         `json:"StorageClass,omitempty"`         // 存储类型
	Owner                Owner          `json:"Owner,omitempty"`                // 属主
	UploadedParts        []UploadedPart `json:"Parts,omitempty"`                // 已完成的Part
}

// ListUploadedParts List Uploaded Parts
//   objectKey: the object name
//   input: key, uploadID and other parameters
func (bkt *Bucket) ListUploadedParts(ctx context.Context, input *ListUploadedPartsInput, options ...Option) (*ListUploadedPartsOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("uploadId", input.UploadID).
		WithQuery("max-parts", strconv.Itoa(input.MaxParts)).
		WithQuery("part-number-marker", strconv.Itoa(input.PartNumberMarker)).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListUploadedPartsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

type UploadInfo struct {
	Key          string `json:"Key,omitempty"`
	UploadId     string `json:"UploadId,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Initiated    string `json:"Initiated,omitempty"`
}

type UploadCommonPrefix struct {
	Prefix string `json:"Prefix"`
}

type ListMultipartUploadsOutput struct {
	RequestInfo        `json:"-"`
	Bucket             string               `json:"Bucket,omitempty"`
	KeyMarker          string               `json:"KeyMarker,omitempty"`
	UploadIdMarker     string               `json:"UploadIdMarker,omitempty"`
	NextKeyMarker      string               `json:"NextKeyMarker,omitempty"`
	NextUploadIdMarker string               `json:"NextUploadIdMarker,omitempty"`
	Delimiter          string               `json:"Delimiter,omitempty"`
	Prefix             string               `json:"Prefix,omitempty"`
	MaxUploads         int32                `json:"MaxUploads,omitempty"`
	IsTruncated        bool                 `json:"IsTruncated,omitempty"`
	Upload             []UploadInfo         `json:"Uploads,omitempty"`
	CommonPrefixes     []UploadCommonPrefix `json:"CommonPrefixes,omitempty"`
}

type ListMultipartUploadsInput struct {
	Prefix         string `json:"Prefix,omitempty"`
	Delimiter      string `json:"Delimiter,omitempty"`
	KeyMarker      string `json:"KeyMarker,omitempty"`
	UploadIDMarker string `json:"UploadIdMarker,omitempty"`
	MaxUploads     int    `json:"MaxUploads,omitempty"`
}

// ListMultipartUploads list multipart uploads
func (bkt *Bucket) ListMultipartUploads(ctx context.Context, input *ListMultipartUploadsInput, options ...Option) (*ListMultipartUploadsOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("uploads", "").
		WithQuery("prefix", input.Prefix).
		WithQuery("delimiter", input.Delimiter).
		WithQuery("key-marker", input.KeyMarker).
		WithQuery("upload-id-marker", input.UploadIDMarker).
		WithQuery("max-uploads", strconv.Itoa(input.MaxUploads)).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListMultipartUploadsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
