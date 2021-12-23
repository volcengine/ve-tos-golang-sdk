package tos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type CopyObjectOutput struct {
	RequestInfo     `json:"-"`
	VersionID       string `json:"VersionId,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"`
	ETag            string `json:"ETag,omitempty"`
	LastModified    string `json:"LastModified,omitempty"`
}

// CopyObject copy an object
//   srcObjectKey: the source object name
//   dstObjectKey: the destination object name. srcObjectKey and dstObjectKey belongs to the same bucket.
//   options: WithVersionID the version id of source object,
//     WithMetadataDirective copy source object metadata or replace with new object metadata,
//     WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl,
//     WithCopySourceIfMatch WithCopySourceIfNoneMatch WithCopySourceIfModifiedSince WithCopySourceIfUnmodifiedSince set copy conditions
//   if CopyObject called with WithMetadataDirective(tos.MetadataDirectiveReplace), these options can be used:
//     WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
func (bkt *Bucket) CopyObject(ctx context.Context, srcObjectKey, dstObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidKey(dstObjectKey, srcObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, bkt.name, dstObjectKey, bkt.name, srcObjectKey, options...)
}

// CopyObjectTo copy an object
//   dstBucket: the destination bucket
//   dstObjectKey: the destination object name
//   srcObjectKey: the source object name
//   options: WithVersionID the version id of source object,
//     WithMetadataDirective copy source object metadata or replace with new object metadata.
//     WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl,
//     WithCopySourceIfMatch WithCopySourceIfNoneMatch WithCopySourceIfModifiedSince WithCopySourceIfUnmodifiedSince set copy conditions
//   if CopyObjectTo called with WithMetadataDirective(tos.MetadataDirectiveReplace), these options can be used:
//     WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
func (bkt *Bucket) CopyObjectTo(ctx context.Context, dstBucket, dstObjectKey, srcObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidNames(dstBucket, dstObjectKey, srcObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, dstBucket, dstObjectKey, bkt.name, srcObjectKey, options...)
}

// CopyObjectFrom copy an object
//   srcBucket: the srcBucket bucket
//   srcObjectKey: the source object name
//   dstObjectKey: the destination object name
//   options: WithVersionID the version id of source object,
//     WithMetadataDirective copy source object metadata or replace with new object metadata
//     WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl,
//     WithCopySourceIfMatch WithCopySourceIfNoneMatch WithCopySourceIfModifiedSince WithCopySourceIfUnmodifiedSince set copy conditions
//   if CopyObjectFrom called with WithMetadataDirective(tos.MetadataDirectiveReplace), these options can be used:
//     WithContentType set Content-Type,
//     WithContentDisposition set Content-Disposition,
//     WithContentLanguage set Content-Language,
//     WithContentEncoding set Content-Encoding,
//     WithCacheControl set Cache-Control,
//     WithExpires set Expires,
//     WithMeta set meta header(s),
func (bkt *Bucket) CopyObjectFrom(ctx context.Context, srcBucket, srcObjectKey, dstObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidNames(srcBucket, srcObjectKey, dstObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, bkt.name, dstObjectKey, srcBucket, srcObjectKey, options...)
}

func (cli *Client) copyObject(ctx context.Context, dstBucket, dstObject string, srcBucket, srcObject string, options ...Option) (*CopyObjectOutput, error) {
	res, err := cli.newBuilder(dstBucket, dstObject, options...).
		RequestWithCopySource(ctx, http.MethodPut, srcBucket, srcObject, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	out := CopyObjectOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(out.RequestID, res.Body, &out); err != nil {
		return nil, err
	}
	out.VersionID = res.Header.Get(HeaderVersionID)
	out.SourceVersionID = res.Header.Get(HeaderCopySourceVersionID)
	return &out, nil
}

type UploadPartCopyInput struct {
	UploadID        string `json:"UploadId,omitempty"`
	DestinationKey  string `json:"DestinationKey,omitempty"`
	SourceBucket    string `json:"SourceBucket,omitempty"`
	SourceKey       string `json:"SourceKey,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"` // optional
	StartOffset     *int64 `json:"StartOffset,omitempty"`     // optional
	PartSize        *int64 `json:"PartSize,omitempty"`        // optional
	PartNumber      int    `json:"PartNumber,omitempty"`
}

type UploadPartCopyOutput struct {
	RequestInfo     `json:"-"`
	VersionID       string `json:"VersionId,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"`
	PartNumber      int    `json:"PartNumber,omitempty"`
	ETag            string `json:"ETag,omitempty"`
	LastModified    string `json:"LastModified,omitempty"`
}

func (up *UploadPartCopyOutput) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: up.PartNumber, ETag: up.ETag}
}

type uploadPartCopyOutput struct {
	ETag         string `json:"ETag,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
}

func copyRange(startOffset, partSize *int64) string {
	cr := ""
	if startOffset != nil {
		if partSize != nil {
			cr = fmt.Sprintf("bytes=%d-%d", *startOffset, *startOffset+*partSize-1)
		} else {
			cr = fmt.Sprintf("bytes=%d-", *startOffset)
		}
	} else if partSize != nil {
		cr = fmt.Sprintf("bytes=0-%d", *partSize-1)
	}
	return cr
}

// UploadPartCopy copy a part of object as a part of a multipart upload operation
//   input: uploadID, DestinationKey, SourceBucket, SourceKey and other parameters,
//   options: WithCopySourceIfMatch WithCopySourceIfNoneMatch WithCopySourceIfModifiedSince WithCopySourceIfUnmodifiedSince set copy conditions
func (bkt *Bucket) UploadPartCopy(ctx context.Context, input *UploadPartCopyInput, options ...Option) (*UploadPartCopyOutput, error) {
	if err := isValidNames(input.SourceBucket, input.DestinationKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.DestinationKey, options...).
		WithQuery("partNumber", strconv.Itoa(input.PartNumber)).
		WithQuery("uploadId", input.UploadID).
		WithQuery("versionId", input.SourceVersionID).
		WithHeader(HeaderCopySourceRange, copyRange(input.StartOffset, input.PartSize)).
		RequestWithCopySource(ctx, http.MethodPut, input.SourceBucket, input.SourceKey, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var out uploadPartCopyOutput
	if err = marshalOutput(res.RequestInfo().RequestID, res.Body, &out); err != nil {
		return nil, err
	}

	return &UploadPartCopyOutput{
		RequestInfo:     res.RequestInfo(),
		VersionID:       res.Header.Get(HeaderVersionID),
		SourceVersionID: res.Header.Get(HeaderCopySourceVersionID),
		PartNumber:      input.PartNumber,
		ETag:            out.ETag,
		LastModified:    out.LastModified,
	}, nil
}

func copySource(bucket, object, versionID string) string {
	if len(versionID) == 0 {
		return "/" + bucket + "/" + url.QueryEscape(object)
	}
	return "/" + bucket + "/" + url.QueryEscape(object) + "?versionId=" + versionID
}
