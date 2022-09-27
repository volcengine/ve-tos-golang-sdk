package tos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

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
//
// Deprecated: use CopyObject of ClientV2 instead
func (bkt *Bucket) CopyObject(ctx context.Context, srcObjectKey, dstObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidKey(dstObjectKey, srcObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, bkt.name, dstObjectKey, bkt.name, srcObjectKey, options...)
}

// CopyObjectTo copy an object to target bucket
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
//
// Deprecated: use CopyObject of ClientV2 instead
func (bkt *Bucket) CopyObjectTo(ctx context.Context, dstBucket, dstObjectKey, srcObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidNames(dstBucket, dstObjectKey, srcObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, dstBucket, dstObjectKey, bkt.name, srcObjectKey, options...)
}

// CopyObjectFrom copy an object from target bucket
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
//
// Deprecated: use CopyObject of ClientV2 instead
func (bkt *Bucket) CopyObjectFrom(ctx context.Context, srcBucket, srcObjectKey, dstObjectKey string, options ...Option) (*CopyObjectOutput, error) {
	if err := isValidNames(srcBucket, srcObjectKey, dstObjectKey); err != nil {
		return nil, err
	}

	return bkt.client.copyObject(ctx, bkt.name, dstObjectKey, srcBucket, srcObjectKey, options...)
}

func (cli *Client) copyObject(ctx context.Context, dstBucket, dstObject string, srcBucket, srcObject string, options ...Option) (*CopyObjectOutput, error) {
	res, err := cli.newBuilder(dstBucket, dstObject, options...).
		WithCopySource(srcBucket, srcObject).
		Request(ctx, http.MethodPut, nil, cli.roundTripper(http.StatusOK))
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

// CopyObject copy an object
func (cli *ClientV2) CopyObject(ctx context.Context, input *CopyObjectInput) (*CopyObjectOutput, error) {
	if err := IsValidBucketName(input.SrcBucket); err != nil {
		return nil, err
	}
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	if err := isValidKey(input.Key, input.SrcKey); err != nil {
		return nil, err
	}
	if err := isValidMetadataDirective(input.MetadataDirective); len(input.MetadataDirective) != 0 && err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithParams(*input).
		WithCopySource(input.SrcBucket, input.SrcKey).
		WithRetry(nil, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, nil, cli.roundTripper(http.StatusOK))
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

func copySource(bucket, object, versionID string) string {
	if len(versionID) == 0 {
		return "/" + bucket + "/" + url.QueryEscape(object)
	}
	return "/" + bucket + "/" + url.QueryEscape(object) + "?versionId=" + versionID
}

func (up *UploadPartCopyOutput) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: up.PartNumber, ETag: up.ETag}
}

// UploadPartCopy copy a part of object as a part of a multipart upload operation
//   input: uploadID, DestinationKey, SourceBucket, SourceKey and other parameters,
//   options: WithCopySourceIfMatch WithCopySourceIfNoneMatch WithCopySourceIfModifiedSince WithCopySourceIfUnmodifiedSince set copy conditions
//
// Deprecated: use UploadPartCopy of ClientV2 instead
func (bkt *Bucket) UploadPartCopy(ctx context.Context, input *UploadPartCopyInput, options ...Option) (*UploadPartCopyOutput, error) {
	if err := isValidNames(input.SourceBucket, input.DestinationKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, input.DestinationKey, options...).
		WithQuery("partNumber", strconv.Itoa(input.PartNumber)).
		WithQuery("uploadId", input.UploadID).
		WithQuery("versionId", input.SourceVersionID).
		WithHeader(HeaderCopySourceRange, copyRange(input.StartOffset, input.PartSize)).
		WithCopySource(input.SourceBucket, input.SourceKey).
		Request(ctx, http.MethodPut, nil, bkt.client.roundTripper(http.StatusOK))
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

func copyRangeV2(start, end int64) string {
	cr := ""
	if start == 0 && end == 0 {
		return cr
	}
	if start > end {
		return cr
	}
	cr = fmt.Sprintf("bytes=%d-%d", start, end)
	return cr
}

// UploadPartCopyV2 copy a part of object as a part of a multipart upload operation
func (cli *ClientV2) UploadPartCopyV2(
	ctx context.Context,
	input *UploadPartCopyV2Input) (*UploadPartCopyV2Output, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	if err := IsValidBucketName(input.SrcBucket); err != nil {
		return nil, err
	}
	if err := isValidKey(input.SrcKey, input.Key); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithParams(*input).
		WithHeader(HeaderCopySourceRange, copyRangeV2(input.CopySourceRangeStart, input.CopySourceRangeEnd)).
		WithCopySource(input.SrcBucket, input.SrcKey).
		WithRetry(nil, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var out uploadPartCopyOutput
	if err = marshalOutput(res.RequestInfo().RequestID, res.Body, &out); err != nil {
		return nil, err
	}

	lastModified, _ := time.ParseInLocation(http.TimeFormat, res.Header.Get(HeaderLastModified), time.UTC)
	return &UploadPartCopyV2Output{
		RequestInfo:         res.RequestInfo(),
		CopySourceVersionID: res.Header.Get(HeaderCopySourceVersionID),
		PartNumber:          input.PartNumber,
		ETag:                out.ETag,
		LastModified:        lastModified,
	}, nil
}
