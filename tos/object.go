package tos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Bucket struct {
	name   string
	client *Client
}

// GetObject get data and metadata of an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object
//    WithRange the range of content,
//    WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//    WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
//
// Deprecated: use GetObject of ClientV2 instead
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

// GetObjectToFile get object and write it to file
func (cli *ClientV2) GetObjectToFile(ctx context.Context, input *GetObjectToFileInput) (*GetObjectToFileOutput, error) {
	tempFilePath := input.FilePath + TempFileSuffix
	fd, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, DefaultFilePerm)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	get, err := cli.GetObjectV2(ctx, &input.GetObjectV2Input)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fd, get.Content)
	if err != nil {
		return nil, err
	}
	err = os.Rename(tempFilePath, input.FilePath)
	if err != nil {
		return nil, err
	}
	return &GetObjectToFileOutput{get.GetObjectBasicOutput}, nil
}

// GetObjectV2 get data and metadata of an object
func (cli *ClientV2) GetObjectV2(ctx context.Context, input *GetObjectV2Input) (*GetObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		WithParams(*input)
	if input.RangeEnd != 0 || input.RangeStart != 0 {
		if input.RangeEnd < input.RangeStart {
			return nil, errors.New("tos: invalid range")
		}
		// set rb.Range will change expected code
		rb.Range = &Range{Start: input.RangeStart, End: input.RangeEnd}
		rb.WithHeader(HeaderRange, rb.Range.String())
	}
	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	basic := GetObjectBasicOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: res.Header.Get(HeaderContentRange),
	}
	basic.ObjectMetaV2.fromResponseV2(res)
	output := GetObjectV2Output{
		GetObjectBasicOutput: basic,
		Content:              wrapReader(res.Body, res.ContentLength, input.DataTransferListener, input.RateLimiter, nil),
	}
	return &output, nil
}

// HeadObject get metadata of an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object
//    WithRange the range of content,
//    WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//    WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
//
// Deprecated: use HeadObject of ClientV2 instead
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

// HeadObjectV2 get metadata of an object
func (cli *ClientV2) HeadObjectV2(ctx context.Context, input *HeadObjectV2Input) (*HeadObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}

	rb := cli.newBuilder(input.Bucket, input.Key).
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{})
	res, err := rb.Request(ctx, http.MethodHead, nil, cli.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := HeadObjectV2Output{
		RequestInfo: res.RequestInfo(),
	}
	output.ObjectMetaV2.fromResponseV2(res)
	return &output, nil
}

func expectedCode(rb *requestBuilder) int {
	okCode := http.StatusOK
	if rb.Range != nil {
		okCode = http.StatusPartialContent
	}
	return okCode
}

// DeleteObject delete an object
//  objectKey: the name of object
//  options: WithVersionID which version of this object will be deleted
//
// Deprecated: use DeleteObject of ClientV2 instead
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

// DeleteObjectV2 delete an object
func (cli *ClientV2) DeleteObjectV2(ctx context.Context, input *DeleteObjectV2Input) (*DeleteObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	deleteMarker, _ := strconv.ParseBool(res.Header.Get(HeaderDeleteMarker))
	return &DeleteObjectV2Output{
		DeleteObjectOutput{
			RequestInfo:  res.RequestInfo(),
			DeleteMarker: deleteMarker,
			VersionID:    res.Header.Get(HeaderVersionID)}}, nil
}

// DeleteMultiObjects delete multi-objects
//   input: the objects will be deleted
//
// Deprecated: use DeleteMultiObjects of ClientV2 instead
func (bkt *Bucket) DeleteMultiObjects(ctx context.Context, input *DeleteMultiObjectsInput, options ...Option) (*DeleteMultiObjectsOutput, error) {
	for _, object := range input.Objects {
		if err := isValidKey(object.Key); err != nil {
			return nil, err
		}
	}

	in, contentMD5, err := marshalInput("DeleteMultiObjectsInput", deleteMultiObjectsInput{
		Objects: input.Objects,
		Quiet:   input.Quiet,
	})
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

// DeleteMultiObjects delete multi-objects
func (cli *ClientV2) DeleteMultiObjects(ctx context.Context, input *DeleteMultiObjectsInput) (*DeleteMultiObjectsOutput, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	for _, object := range input.Objects {
		if err := isValidKey(object.Key); err != nil {
			return nil, err
		}
	}
	in, contentMD5, err := marshalInput("DeleteMultiObjectsInput", deleteMultiObjectsInput{
		Objects: input.Objects,
		Quiet:   input.Quiet,
	})
	if err != nil {
		return nil, err
	}
	// POST method, don't retry
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("delete", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(nil, ServerErrorClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(in), cli.roundTripper(http.StatusOK))
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
//
// Deprecated: use PutObject of ClientV2 instead
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

// url-encode Chinese characters only
func urlEncodeChinese(s string) string {
	var res string
	r := []rune(s)
	for i := 0; i < len(r); i++ {
		if r[i] >= 0x4E00 && r[i] <= 0x9FA5 {
			res += url.QueryEscape(string(r[i]))
		} else {
			res += string(r[i])
		}
	}
	return res
}

func checkCrc64(res *Response, checker hash.Hash64) error {
	if res.Header.Get(HeaderHashCrc64ecma) == "" || checker == nil {
		return nil
	}
	crc64, err := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	if err != nil {
		return &TosServerError{
			TosError:    TosError{"tos: server returned invalid crc"},
			RequestInfo: res.RequestInfo(),
		}
	}
	if checker.Sum64() != crc64 {
		return &TosServerError{
			TosError:    TosError{Message: fmt.Sprintf("tos: crc64 check failed, expected:%d, in fact:%d", crc64, checker.Sum64())},
			RequestInfo: res.RequestInfo(),
		}
	}
	return nil
}

// wrapReader wrap reader with some extension function.
// If reader can be interpreted as io.ReadCloser, use itself as base ReadCloser, else wrap it a NopCloser.
func wrapReader(reader io.Reader, totalBytes int64, listener DataTransferListener, limiter RateLimiter, checker hash.Hash64) io.ReadCloser {
	var wrapped io.ReadCloser
	// get base ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		wrapped = rc
	}
	wrapped = ioutil.NopCloser(reader)
	// wrap with listener
	if listener != nil {
		wrapped = &readCloserWithListener{
			listener: listener,
			base:     wrapped,
			consumed: 0,
			total:    totalBytes,
		}
	}
	// wrap with limiter
	if limiter != nil {
		wrapped = &ReadCloserWithLimiter{
			limiter: limiter,
			base:    wrapped,
		}
	}
	// wrap with crc64 checker
	if checker != nil {
		wrapped = &readCloserWithCRC{
			checker: checker,
			base:    wrapped,
		}
	}
	return wrapped
}

// PutObjectV2 put an object
func (cli *ClientV2) PutObjectV2(ctx context.Context, input *PutObjectV2Input) (*PutObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}
	var (
		checker       hash.Hash64
		content       = input.Content
		contentLength = input.ContentLength
	)
	if cli.enableCRC {
		checker = NewCRC(DefaultCrcTable(), 0)
	}
	if contentLength <= 0 {
		contentLength = tryResolveLength(content)
	}
	if content != nil {
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, checker)
	}
	var (
		onRetry    func(req *Request) = nil
		classifier classifier
	)
	classifier = StatusCodeClassifier{}
	if seeker, ok := content.(io.Seeker); ok {
		start, err := seeker.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, err
		}
		onRetry = func(req *Request) {
			// PutObject/UploadPart can be treated as an idempotent semantics if the request message body
			// supports a reset operation. e.g. the request message body is a string,
			// a local file handle, binary data in memory
			if seeker, ok := req.Content.(io.Seeker); ok {
				seeker.Seek(start, io.SeekStart)
			}
		}
	}
	if onRetry == nil {
		classifier = ServerErrorClassifier{}
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		WithContentLength(contentLength).
		WithParams(*input).
		WithRetry(onRetry, classifier)
	res, err := rb.Request(ctx, http.MethodPut, content, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if err = checkCrc64(res, checker); err != nil {
		return nil, err
	}
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	return &PutObjectV2Output{
		RequestInfo:   res.RequestInfo(),
		ETag:          res.Header.Get(HeaderETag),
		SSECAlgorithm: res.Header.Get(HeaderSSECustomerAlgorithm),
		SSECKeyMD5:    res.Header.Get(HeaderSSECustomerKeyMD5),
		VersionID:     res.Header.Get(HeaderVersionID),
		HashCrc64ecma: crc64,
	}, nil
}

// PutObjectFromFile put an object from file
func (cli *ClientV2) PutObjectFromFile(ctx context.Context, input *PutObjectFromFileInput) (*PutObjectFromFileOutput, error) {
	file, err := os.Open(input.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	putOutput, err := cli.PutObjectV2(ctx, &PutObjectV2Input{
		PutObjectBasicInput: input.PutObjectBasicInput,
		Content:             file,
	})
	if err != nil {
		return nil, err
	}
	return &PutObjectFromFileOutput{*putOutput}, err
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
//
// Deprecated: use AppendObject of ClientV2 instead
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
		return nil, fmt.Errorf("tos: server return unexpected Next-Append-Offset header %q", nextOffset)
	}
	return &AppendObjectOutput{
		RequestInfo:      res.RequestInfo(),
		ETag:             res.Header.Get(HeaderETag),
		NextAppendOffset: appendOffset,
	}, nil
}

// AppendObjectV2 append content at the tail of an appendable object
func (cli *ClientV2) AppendObjectV2(ctx context.Context, input *AppendObjectV2Input) (*AppendObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}
	var (
		checker       hash.Hash64
		content       = input.Content
		contentLength = input.ContentLength
	)
	if contentLength <= 0 {
		contentLength = tryResolveLength(content)
	}
	if cli.enableCRC {
		checker = NewCRC(DefaultCrcTable(), input.PreHashCrc64ecma)
	}
	if content != nil {
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, checker)
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("append", "").
		WithParams(*input).
		WithContentLength(contentLength).
		WithRetry(nil, NoRetryClassifier{}).
		Request(ctx, http.MethodPost, content, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	nextOffset := res.Header.Get(HeaderNextAppendOffset)
	appendOffset, err := strconv.ParseInt(nextOffset, 10, 64)
	if err != nil {
		return nil, &TosServerError{
			TosError:    TosError{fmt.Sprintf("tos: server return unexpected Next-Append-Offset header %q", nextOffset)},
			RequestInfo: res.RequestInfo(),
		}
	}
	if err = checkCrc64(res, checker); err != nil {
		return nil, err
	}
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	return &AppendObjectV2Output{
		RequestInfo:      res.RequestInfo(),
		VersionID:        res.Header.Get(HeaderVersionID),
		NextAppendOffset: appendOffset,
		HashCrc64ecma:    crc64,
	}, nil
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
//
// Deprecated: use SetObjectMeta of ClientV2 instead
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

// SetObjectMeta overwrites metadata of the object
func (cli *ClientV2) SetObjectMeta(ctx context.Context, input *SetObjectMetaInput) (*SetObjectMetaOutput, error) {
	if err := isValidNames(input.Bucket, input.Key); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("metadata", "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodPost, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &SetObjectMetaOutput{RequestInfo: res.RequestInfo()}, nil
}

// ListObjects list objects of a bucket
//
// Deprecated: use ListObjects of ClientV2 instead
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

// ListObjectsV2 list objects of a bucket
func (cli *ClientV2) ListObjectsV2(ctx context.Context, input *ListObjectsV2Input) (*ListObjectsV2Output, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithParams(*input).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	temp := listObjectsV2Output{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(temp.RequestID, res.Body, &temp); err != nil {
		return nil, err
	}
	contents := make([]ListedObjectV2, 0, len(temp.Contents))
	for _, object := range temp.Contents {
		var hashCrc uint64
		if len(object.HashCrc64ecma) == 0 {
			hashCrc = 0
		} else {
			hashCrc, err = strconv.ParseUint(object.HashCrc64ecma, 10, 64)
			if err != nil {
				return nil, &TosServerError{
					TosError:    TosError{Message: "tos: server returned invalid HashCrc64Ecma"},
					RequestInfo: RequestInfo{RequestID: temp.RequestID},
				}
			}
		}
		contents = append(contents, ListedObjectV2{
			Key:           object.Key,
			LastModified:  object.LastModified,
			ETag:          object.ETag,
			Size:          object.Size,
			Owner:         object.Owner,
			StorageClass:  object.StorageClass,
			HashCrc64ecma: uint64(hashCrc),
		})
	}
	output := ListObjectsV2Output{
		RequestInfo:    temp.RequestInfo,
		Name:           temp.Name,
		Prefix:         temp.Prefix,
		Marker:         temp.Marker,
		MaxKeys:        temp.MaxKeys,
		NextMarker:     temp.NextMarker,
		Delimiter:      temp.Delimiter,
		IsTruncated:    temp.IsTruncated,
		EncodingType:   temp.EncodingType,
		CommonPrefixes: temp.CommonPrefixes,
		Contents:       contents,
	}
	return &output, nil
}

// ListObjectVersions list multi-version objects of a bucket
//
// Deprecated: use ListObjectV2Versions of ClientV2 instead
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

// ListObjectVersionsV2 list multi-version objects of a bucket
func (cli *ClientV2) ListObjectVersionsV2(
	ctx context.Context,
	input *ListObjectVersionsV2Input) (*ListObjectVersionsV2Output, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("versions", "").
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	temp := listObjectVersionsV2Output{RequestInfo: res.RequestInfo()}

	if err = marshalOutput(temp.RequestID, res.Body, &temp); err != nil {
		return nil, err
	}
	versions := make([]ListedObjectVersionV2, 0, len(temp.Versions))
	for _, version := range temp.Versions {
		var hashCrc uint64
		if len(version.HashCrc64ecma) == 0 {
			hashCrc = 0
		} else {
			hashCrc, err = strconv.ParseUint(version.HashCrc64ecma, 10, 64)
			if err != nil {
				return nil, &TosServerError{
					TosError:    TosError{Message: "tos: server returned invalid HashCrc64Ecma"},
					RequestInfo: RequestInfo{RequestID: temp.RequestID},
				}
			}
		}
		versions = append(versions, ListedObjectVersionV2{
			Key:           version.Key,
			LastModified:  version.LastModified,
			ETag:          version.ETag,
			IsLatest:      version.IsLatest,
			Size:          version.Size,
			Owner:         version.Owner,
			StorageClass:  version.StorageClass,
			VersionID:     version.VersionID,
			HashCrc64ecma: hashCrc,
		})
	}
	output := ListObjectVersionsV2Output{
		RequestInfo:         temp.RequestInfo,
		Name:                temp.Name,
		Prefix:              temp.Prefix,
		KeyMarker:           temp.KeyMarker,
		VersionIDMarker:     temp.VersionIDMarker,
		Delimiter:           temp.Delimiter,
		EncodingType:        temp.EncodingType,
		MaxKeys:             temp.MaxKeys,
		NextKeyMarker:       temp.NextKeyMarker,
		NextVersionIDMarker: temp.NextVersionIDMarker,
		IsTruncated:         temp.IsTruncated,
		CommonPrefixes:      temp.CommonPrefixes,
		DeleteMarkers:       temp.DeleteMarkers,
		Versions:            versions,
	}

	return &output, nil
}
