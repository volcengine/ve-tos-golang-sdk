package tos

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Bucket struct {
	name       string
	client     *Client
	baseClient *baseClient
}

// GetObject get data and metadata of an object
//
//	objectKey: the name of object
//	options: WithVersionID which version of this object
//	  WithRange the range of content,
//	  WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//	  WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
//
// Deprecated: use GetObject of ClientV2 instead
func (bkt *Bucket) GetObject(ctx context.Context, objectKey string, options ...Option) (*GetObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}
	rb := bkt.client.newBuilder(bkt.name, objectKey, options...)
	res, err := rb.WithRetry(nil, StatusCodeClassifier{}).Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	output := GetObjectOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: rb.Header.Get(HeaderContentRange),
		Content:      res.Body,
	}
	output.ObjectMeta.fromResponse(res, bkt.client.disableEncodingMeta)
	return &output, nil
}

func (cli *ClientV2) copyToFile(fileName string, reader io.Reader) error {
	fd, err := os.OpenFile(filepath.Clean(fileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, DefaultFilePerm)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = io.Copy(fd, reader)
	if err != nil {
		return err
	}
	return nil
}

// GetObjectToFile get object and write it to file
func (cli *ClientV2) GetObjectToFile(ctx context.Context, input *GetObjectToFileInput) (*GetObjectToFileOutput, error) {

	err := checkAndCreateDir(input.FilePath)
	if err != nil {
		return nil, InvalidFilePath.withCause(err)
	}

	tempFilePath := input.FilePath + TempFileSuffix
	start := time.Now()
	get, err := cli.GetObjectV2(ctx, &input.GetObjectV2Input)

	if err != nil {
		return nil, err
	}
	defer get.Content.Close()
	err = cli.copyToFile(tempFilePath, get.Content)
	if err != nil {
		return nil, newTosClientError("GetObject to File error", err)
	}

	cli.printGetObjectSlowLog(&get.RequestID, int(get.ContentLength), get.StatusCode, start)
	err = os.Rename(tempFilePath, input.FilePath)
	if err != nil {
		return nil, err
	}
	return &GetObjectToFileOutput{get.GetObjectBasicOutput}, nil
}

func (cli *ClientV2) printGetObjectSlowLog(requestId *string, contentLength int, statusCode int, start time.Time) {
	if cli.config.TransportConfig.HighLatencyLogThreshold != nil && isSlow(contentLength, *cli.config.TransportConfig.HighLatencyLogThreshold, time.Since(start)) {
		logger := cli.logger
		if logger == nil {
			logger = stdlog
		}

		prefix := buildSlowPrefix(requestId)
		logger.Warn(fmt.Sprintf("%s Response StatusCode:%d, Cost:%d ms", prefix, statusCode, time.Since(start).Milliseconds()))
	}
}

// GetObjectV2 get data and metadata of an object
func (cli *ClientV2) GetObjectV2(ctx context.Context, input *GetObjectV2Input) (*GetObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithParams(*input).WithRetry(nil, StatusCodeClassifier{})
	isRange := false
	if input.Range != "" {
		rb.WithHeader(HeaderRange, input.Range)
		isRange = true
	} else if input.RangeEnd != 0 || input.RangeStart != 0 {
		if input.RangeEnd < input.RangeStart {
			return nil, errors.New("tos: invalid range")
		}
		// set rb.Range will change expected code
		rb.Range = &Range{Start: input.RangeStart, End: input.RangeEnd}
		rb.WithHeader(HeaderRange, rb.Range.String())
		isRange = true
	}
	if isRange && input.ResponseContentEncoding == "" && !cli.disableTrailerHeader {
		rb.WithHeader(HeaderTosTrailer, "x-tos-hash-range-crc64ecma")
		rb.WithHeader(HeaderAcceptEncoding, "tos-raw-trailer")
	}
	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripperWithSlowLog(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	basic := GetObjectBasicOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: res.Header.Get(HeaderContentRange),
	}
	basic.ObjectMetaV2.fromResponseV2(res, cli.disableEncodingMeta)
	var serverCrc uint64
	var checker hash.Hash64
	// 200 为完整请求
	if res.StatusCode == http.StatusOK && cli.enableCRC {
		serverCrc = basic.HashCrc64ecma
		checker = NewCRC(DefaultCrcTable(), 0)
	}
	body := res.Body
	if isRange && res.Header.Get(HeaderRawContentLength) != "" && res.Header.Get(HeaderContentEncoding) != "" && !cli.disableTrailerHeader {
		body = newChunkReader(body, basic.ContentLength)
	}
	output := GetObjectV2Output{
		GetObjectBasicOutput: basic,
		Content:              wrapReader(body, basic.ContentLength, input.DataTransferListener, input.RateLimiter, &crcChecker{checker: checker, serverCrc: serverCrc}),
	}
	return &output, nil
}

// HeadObject get metadata of an object
//
//	objectKey: the name of object
//	options: WithVersionID which version of this object
//	  WithRange the range of content,
//	  WithIfModifiedSince return if the object modified after the given date, otherwise return status code 304
//	  WithIfUnmodifiedSince, WithIfMatch, WithIfNoneMatch set If-Unmodified-Since, If-Match and If-None-Match
//
// Deprecated: use HeadObject of ClientV2 instead
func (bkt *Bucket) HeadObject(ctx context.Context, objectKey string, options ...Option) (*HeadObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	rb := bkt.client.newBuilder(bkt.name, objectKey, options...)

	res, err := rb.WithRetry(nil, StatusCodeClassifier{}).Request(ctx, http.MethodHead, nil, bkt.client.roundTripper(expectedCode(rb)))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := HeadObjectOutput{
		RequestInfo:  res.RequestInfo(),
		ContentRange: rb.Header.Get(HeaderContentRange),
	}
	output.ObjectMeta.fromResponse(res, bkt.client.disableEncodingMeta)
	if symlinkTargetSize := output.Header.Get(HeaderSymlinkTargetSize); symlinkTargetSize != "" {
		output.SymlinkTargetSize, _ = strconv.ParseInt(symlinkTargetSize, 10, 64)
	}
	return &output, nil
}

// HeadObjectV2 get metadata of an object
func (cli *ClientV2) HeadObjectV2(ctx context.Context, input *HeadObjectV2Input) (*HeadObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}

	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
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
	output.ObjectMetaV2.fromResponseV2(res, cli.disableEncodingMeta)
	if symlinkTargetSize := output.Header.Get(HeaderSymlinkTargetSize); symlinkTargetSize != "" {
		output.SymlinkTargetSize, _ = strconv.ParseInt(symlinkTargetSize, 10, 64)
	}
	return &output, nil
}

func expectedCode(rb *requestBuilder) int {
	okCode := http.StatusOK
	if rb.Header.Get(HeaderRange) != "" || rb.Query.Get(QueryPartNumber) != "" {
		okCode = http.StatusPartialContent
	}
	return okCode
}

// DeleteObject delete an object
//
//	objectKey: the name of object
//	options: WithVersionID which version of this object will be deleted
//
// Deprecated: use DeleteObject of ClientV2 instead
func (bkt *Bucket) DeleteObject(ctx context.Context, objectKey string, options ...Option) (*DeleteObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).WithRetry(nil, StatusCodeClassifier{}).
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
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	reqBuilder := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{})
	if input.Recursive {
		reqBuilder.WithQuery(QueryRecursive, strconv.FormatBool(input.Recursive))
	}
	res, err := reqBuilder.
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
			VersionID:    res.Header.Get(HeaderVersionID),
			TrashPath:    res.Header.Get(HeaderTrashPath),
		}}, nil
}

// DeleteMultiObjects delete multi-objects
//
//	input: the objects will be deleted
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
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(in), bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := DeleteMultiObjectsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// DeleteMultiObjects delete multi-objects
func (cli *ClientV2) DeleteMultiObjects(ctx context.Context, input *DeleteMultiObjectsInput) (*DeleteMultiObjectsOutput, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	if len(input.Objects) == 0 {
		return nil, InvlidDeleteMultiObjectsLength
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
	rb := cli.newBuilder(input.Bucket, "").
		WithQuery("delete", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{})
	if input.Recursive {
		rb.WithQuery(QueryRecursive, strconv.FormatBool(input.Recursive))
	}
	res, err := rb.
		Request(ctx, http.MethodPost, bytes.NewReader(in), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteMultiObjectsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// PutObject put an object
//
//	objectKey: the name of object
//	content: the content of object
//	options: WithContentType set Content-Type,
//	  WithContentDisposition set Content-Disposition,
//	  WithContentLanguage set Content-Language,
//	  WithContentEncoding set Content-Encoding,
//	  WithCacheControl set Cache-Control,
//	  WithExpires set Expires,
//	  WithMeta set meta header(s),
//	  WithContentSHA256 set Content-Sha256,
//	  WithContentMD5 set Content-MD5
//	  WithExpires set Expires,
//	  WithServerSideEncryptionCustomer set server side encryption options
//	  WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl
//
// NOTICE: only content with a known length is supported now,
//
//	e.g, bytes.Buffer, bytes.Reader, strings.Reader, os.File, io.LimitedReader, net.Buffers.
//	if the parameter content(an io.Reader) is not one of these,
//	please use io.LimitReader(reader, length) to wrap this reader or use the WithContentLength option.
//
// Deprecated: use PutObjectV2 of ClientV2 instead
func (bkt *Bucket) PutObject(ctx context.Context, objectKey string, content io.Reader, options ...Option) (*PutObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}
	var (
		onRetry    func(req *Request) error = nil
		classifier classifier
	)
	classifier = NoRetryClassifier{}
	if seeker, ok := content.(io.Seeker); ok {
		start, err := seeker.Seek(0, io.SeekCurrent)
		if err == nil {
			onRetry = func(req *Request) error {
				// PutObject/UploadPart can be treated as an idempotent semantics if the request message body
				// supports a reset operation. e.g. the request message body is a string,
				// a local file handle, binary data in memory
				if seeker, ok := req.Content.(io.Seeker); ok {
					_, err := seeker.Seek(start, io.SeekStart)
					if err != nil {
						return err
					}
				} else {
					return newTosClientError("Io Reader not support retry", nil)
				}
				return nil
			}
			classifier = StatusCodeClassifier{}
		}
	}
	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithRetry(onRetry, classifier).
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

func skipEscape(i byte) bool {
	return (i >= 'A' && i <= 'Z') || (i >= 'a' && i <= 'z') || (i >= '0' && i <= '9') ||
		i == '-' ||
		i == '.' ||
		i == '_' ||
		i == '~'
}

func skipContentDispositionEscape(i byte) bool {
	return skipEscape(i) ||
		i == ' ' ||
		i == '\'' ||
		i == '"'
}

func escapeHeader(s string, skip func(i byte) bool) string {
	var buf bytes.Buffer
	for i := 0; i < len(s); i++ {
		c := s[i]
		if skip(c) {
			buf.WriteByte(c)
		} else {
			fmt.Fprintf(&buf, "%%%02X", c)
		}
	}
	return buf.String()
}

func existChinese(s string) bool {
	r := []rune(s)

	for i := 0; i < len(r); i++ {
		if r[i] >= 0x4E00 && r[i] <= 0x9FA5 {
			return true
		}
	}
	return false
}

// url-encode Chinese characters only
func headerEncode(s string) string {
	return escapeHeader(s, skipEscape)
}

func checkCrc64(res *Response, checker hash.Hash64) error {
	if res.Header.Get(HeaderHashCrc64ecma) == "" || checker == nil {
		return nil
	}
	crc64, err := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	if err != nil {
		return &TosServerError{
			TosError:    newTosErr("tos: server returned invalid crc", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
			RequestInfo: res.RequestInfo(),
		}
	}
	if checker.Sum64() != crc64 {
		return &TosServerError{
			TosError:    newTosErr(fmt.Sprintf("tos: crc64 check failed, expected:%d, in fact:%d", crc64, checker.Sum64()), res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
			RequestInfo: res.RequestInfo(),
		}
	}
	return nil
}

type crcChecker struct {
	checker   hash.Hash64
	serverCrc uint64
}

type nopCloser struct {
	base io.Reader
}

func wrapCloser(reader io.Reader) io.ReadCloser {
	return &nopCloser{base: reader}
}

func (n2 nopCloser) Seek(offset int64, whence int) (int64, error) {
	seeker, ok := n2.base.(io.Seeker)
	if !ok {
		return 0, NotSupportSeek
	}
	return seeker.Seek(offset, whence)
}

func (n2 nopCloser) Read(p []byte) (n int, err error) {
	return n2.base.Read(p)
}

func (n2 nopCloser) Close() error {
	return nil
}

// wrapReader wrap reader with some extension function.
// If reader can be interpreted as io.ReadCloser, use itself as base ReadCloser, else wrap it a NopCloser.
func wrapReader(reader io.Reader, totalBytes int64, listener DataTransferListener, limiter RateLimiter, crcChecker *crcChecker) io.ReadCloser {
	var wrapped io.ReadCloser
	// get base ReadCloser
	if rc, ok := reader.(io.ReadCloser); ok {
		wrapped = rc
	} else {
		wrapped = wrapCloser(reader)
	}
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
	if crcChecker != nil && crcChecker.checker != nil {
		wrapped = &readCloserWithCRC{
			serverCrc: crcChecker.serverCrc,
			checker:   crcChecker.checker,
			base:      wrapped,
		}
	}
	return wrapped
}

func (cli *ClientV2) setExpectHeader(rb *requestBuilder, contentLength int64) {
	if cli.except100ContinueThreshold > 0 && (contentLength < 0 || contentLength > cli.except100ContinueThreshold) {
		rb.WithHeader(HeaderExpect, "100-continue")
	}
}

// PutObjectV2 put an object
func (cli *ClientV2) PutObjectV2(ctx context.Context, input *PutObjectV2Input) (*PutObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if err := isValidSSECAlgorithm(input.SSECAlgorithm); len(input.SSECAlgorithm) != 0 && err != nil {
		return nil, err
	}

	if err := isValidACL(input.ACL); len(input.ACL) != 0 && err != nil {
		return nil, err
	}

	if err := isValidStorageClass(input.StorageClass); len(input.StorageClass) != 0 && err != nil {
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

	var (
		onRetry    func(req *Request) error = nil
		classifier classifier
	)

	if content != nil {
		if _, ok := content.(*os.File); ok {
			content = wrapCloser(content)
		}
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, &crcChecker{checker: checker})
	}

	classifier = NoRetryClassifier{}
	if seeker, ok := content.(io.Seeker); ok {
		start, err := seeker.Seek(0, io.SeekCurrent)
		if err == nil {
			onRetry = func(req *Request) error {
				// PutObject/UploadPart can be treated as an idempotent semantics if the request message body
				// supports a reset operation. e.g. the request message body is a string,
				// a local file handle, binary data in memory
				if seeker, ok := req.Content.(io.Seeker); ok {
					_, err := seeker.Seek(start, io.SeekStart)
					if err != nil {
						return err
					}
				} else {
					return newTosClientError("Io Reader not support retry", nil)
				}
				return nil
			}
			classifier = StatusCodeClassifier{}
		}
	}

	rb := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithContentLength(contentLength).
		WithParams(*input).
		WithEnableTrailer(input.ContentMD5 == "" && !cli.disableTrailerHeader).
		WithRetry(onRetry, classifier)

	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPut, content, cli.roundTripperWithSlowLog(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if err = checkCrc64(res, checker); err != nil {
		return nil, err
	}
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	callbackResult := ""
	if input.Callback != "" && res.Body != nil {
		callbackRes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, &TosServerError{
				TosError:    newTosErr(fmt.Sprintf("tos: read callback result err:%s", err.Error()), res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
				RequestInfo: res.RequestInfo(),
			}
		}
		if len(callbackRes) > 0 {
			callbackResult = string(callbackRes)
		}
	}

	return &PutObjectV2Output{
		RequestInfo:               res.RequestInfo(),
		ETag:                      res.Header.Get(HeaderETag),
		SSECAlgorithm:             res.Header.Get(HeaderSSECustomerAlgorithm),
		SSECKeyMD5:                res.Header.Get(HeaderSSECustomerKeyMD5),
		VersionID:                 res.Header.Get(HeaderVersionID),
		ServerSideEncryption:      res.Header.Get(HeaderServerSideEncryption),
		ServerSideEncryptionKeyID: res.Header.Get(HeaderServerSideEncryptionKmsKeyID),
		CallbackResult:            callbackResult,
		HashCrc64ecma:             crc64,
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
//
//	objectKey: the name of object
//	content: the content of object
//	offset: append position, equals to the current object-size
//	options: WithContentType set Content-Type,
//	  WithContentDisposition set Content-Disposition,
//	  WithContentLanguage set Content-Language,
//	  WithContentEncoding set Content-Encoding,
//	  WithCacheControl set Cache-Control,
//	  WithExpires set Expires,
//	  WithMeta set meta header(s),
//	  WithACL WithACLGrantFullControl WithACLGrantRead WithACLGrantReadAcp WithACLGrantWrite WithACLGrantWriteAcp set object acl
//	above options only take effect when offset parameter is 0.
//	  WithContentSHA256 set Content-Sha256,
//	  WithContentMD5 set Content-MD5.
//
// NOTICE: only content with a known length is supported now,
//
//	e.g, bytes.Buffer, bytes.Reader, strings.Reader, os.File, io.LimitedReader, net.Buffers.
//	if the parameter content(an io.Reader) is not one of these,
//	please use io.LimitReader(reader, length) to wrap this reader or use the WithContentLength option.
//
// Deprecated: use AppendObject of ClientV2 instead
func (bkt *Bucket) AppendObject(ctx context.Context, objectKey string, content io.Reader, offset int64, options ...Option) (*AppendObjectOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithQuery("append", "").
		WithQuery("offset", strconv.FormatInt(offset, 10)).
		WithRetry(nil, NoRetryClassifier{}).
		Request(ctx, http.MethodPost, content, bkt.client.roundTripperWithSlowLog(http.StatusOK))
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

func (bkt *Bucket) PutObjectTagging(ctx context.Context, input *PutObjectTaggingInput, option ...Option) (*PutObjectTaggingOutput, error) {
	return bkt.baseClient.PutObjectTagging(ctx, input, option...)
}

func (bkt *Bucket) GetObjectTagging(ctx context.Context, input *GetObjectTaggingInput, option ...Option) (*GetObjectTaggingOutput, error) {
	return bkt.baseClient.GetObjectTagging(ctx, input, option...)
}

func (bkt *Bucket) DeleteObjectTagging(ctx context.Context, input *DeleteObjectTaggingInput, option ...Option) (*DeleteObjectTaggingOutput, error) {
	return bkt.baseClient.DeleteObjectTagging(ctx, input, option...)
}

func (bkt *Bucket) RestoreObject(ctx context.Context, input *RestoreObjectInput, option ...Option) (*RestoreObjectOutput, error) {
	return bkt.baseClient.RestoreObject(ctx, input, option...)
}

func (bkt *Bucket) PutSymlink(ctx context.Context, input *PutSymlinkInput, option ...Option) (*PutSymlinkOutput, error) {
	if input.Bucket == "" {
		input.Bucket = bkt.name
	}
	return bkt.baseClient.PutSymlink(ctx, input, option...)
}

func (bkt *Bucket) GetSymlink(ctx context.Context, input *GetSymlinkInput, option ...Option) (*GetSymlinkOutput, error) {
	if input.Bucket == "" {
		input.Bucket = bkt.name
	}
	return bkt.baseClient.GetSymlink(ctx, input, option...)
}

func (cli *ClientV2) hnsAppendObject(ctx context.Context, input *AppendObjectV2Input) (*AppendObjectV2Output, error) {
	resp, err := cli.modifyObjectWithInitCrc64(ctx, input)
	if err != nil {
		return nil, err
	}
	return &AppendObjectV2Output{
		RequestInfo:      resp.RequestInfo,
		NextAppendOffset: resp.NextModifyOffset,
		HashCrc64ecma:    resp.HashCrc64ecma,
	}, nil
}

func (cli *ClientV2) fnsAppendObject(ctx context.Context, input *AppendObjectV2Input) (*AppendObjectV2Output, error) {
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
		content = wrapReader(content, contentLength, input.DataTransferListener, input.RateLimiter, &crcChecker{checker: checker})
	}
	rb := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("append", "").
		WithParams(*input).
		WithContentLength(contentLength).
		WithRetry(nil, NoRetryClassifier{})
	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPost, content, cli.roundTripperWithSlowLog(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	nextOffset := res.Header.Get(HeaderNextAppendOffset)
	appendOffset, err := strconv.ParseInt(nextOffset, 10, 64)
	if err != nil {
		return nil, &TosServerError{
			TosError:    newTosErr(fmt.Sprintf("tos: server return unexpected Next-Append-Offset header %q", nextOffset), res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
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

// AppendObjectV2 append content at the tail of an appendable object
// HNS bucket append requires the object to already exist
func (cli *ClientV2) AppendObjectV2(ctx context.Context, input *AppendObjectV2Input) (*AppendObjectV2Output, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	bucketType, err := cli.getBucketType(ctx, input.Bucket)
	if err != nil {
		return nil, err
	}
	if bucketType == enum.BucketTypeHNS {
		return cli.hnsAppendObject(ctx, input)
	} else {
		return cli.fnsAppendObject(ctx, input)
	}
}

// SetObjectMeta overwrites metadata of the object
//
//	objectKey: the name of object
//	options: WithContentType set Content-Type,
//	  WithContentDisposition set Content-Disposition,
//	  WithContentLanguage set Content-Language,
//	  WithContentEncoding set Content-Encoding,
//	  WithCacheControl set Cache-Control,
//	  WithExpires set Expires,
//	  WithMeta set meta header(s),
//	  WithVersionID which version of this object will be set
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
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodPost, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &SetObjectMetaOutput{RequestInfo: res.RequestInfo()}, nil
}

// SetObjectMeta overwrites metadata of the object
func (cli *ClientV2) SetObjectMeta(ctx context.Context, input *SetObjectMetaInput) (*SetObjectMetaOutput, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
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
// Deprecated: use ListObjectsV2 of ClientV2 instead
func (bkt *Bucket) ListObjects(ctx context.Context, input *ListObjectsInput, options ...Option) (*ListObjectsOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("prefix", input.Prefix).
		WithQuery("delimiter", input.Delimiter).
		WithQuery("marker", input.Marker).
		WithQuery("max-keys", strconv.Itoa(input.MaxKeys)).
		WithQuery("encoding-type", input.EncodingType).
		WithQuery("fetch-meta", strconv.FormatBool(input.FetchMeta)).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	internalOutput := &listObjectsOutput{}
	if err = marshalOutput(res, &internalOutput); err != nil {
		return nil, err
	}
	output := ListObjectsOutput{
		RequestInfo:    res.RequestInfo(),
		Name:           internalOutput.Name,
		Prefix:         internalOutput.Prefix,
		Marker:         internalOutput.Marker,
		MaxKeys:        internalOutput.MaxKeys,
		NextMarker:     internalOutput.NextMarker,
		Delimiter:      internalOutput.Delimiter,
		IsTruncated:    internalOutput.IsTruncated,
		EncodingType:   internalOutput.EncodingType,
		CommonPrefixes: internalOutput.CommonPrefixes,
		Contents:       nil,
	}
	contents := make([]ListedObject, 0, len(internalOutput.Contents))
	for _, content := range internalOutput.Contents {

		var hashCrc uint64
		if len(content.HashCrc64ecma) == 0 {
			hashCrc = 0
		} else {
			hashCrc, err = strconv.ParseUint(content.HashCrc64ecma, 10, 64)
			if err != nil {
				return nil, &TosServerError{
					TosError:    newTosErr("tos: server returned invalid HashCrc64Ecma", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
					RequestInfo: RequestInfo{RequestID: res.RequestInfo().RequestID},
				}
			}
		}
		contents = append(contents, ListedObject{
			Key:           content.Key,
			LastModified:  content.LastModified,
			ETag:          content.ETag,
			Size:          content.Size,
			Owner:         content.Owner,
			StorageClass:  content.StorageClass,
			Type:          content.Type,
			Meta:          parseUserMetaData(content.Meta),
			HashCrc64ecma: hashCrc,
			ObjectType:    content.Type,
		})
	}
	output.Contents = contents
	return &output, nil
}

// ListObjectsV2 list objects of a bucket
// Deprecated: use ListObjectsType2 of ClientV2 instead
func (cli *ClientV2) ListObjectsV2(ctx context.Context, input *ListObjectsV2Input) (*ListObjectsV2Output, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithParams(*input).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	temp := listObjectsV2Output{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &temp); err != nil {
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
					TosError:    newTosErr("tos: server returned invalid HashCrc64Ecma", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
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
			Meta:          parseUserMetaData(object.Meta),
			ObjectType:    object.ObjectType,
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

func (cli *ClientV2) listObjectsType2(ctx context.Context, input *ListObjectsType2Input) (*ListObjectsType2Output, error) {
	res, err := cli.newBuilder(input.Bucket, "").
		WithParams(*input).
		WithQuery("list-type", "2").
		WithQuery("fetch-owner", "true").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	temp := listObjectsType2Output{
		RequestInfo: res.RequestInfo(),
	}
	if err = marshalOutput(res, &temp); err != nil {
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
					TosError:    newTosErr("tos: server returned invalid HashCrc64Ecma", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
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
			HashCrc64ecma: hashCrc,
			Meta:          parseUserMetaData(object.Meta),
			ObjectType:    object.ObjectType,
		})
	}
	output := ListObjectsType2Output{
		RequestInfo:           temp.RequestInfo,
		Name:                  temp.Name,
		ContinuationToken:     temp.ContinuationToken,
		Prefix:                temp.Prefix,
		MaxKeys:               temp.MaxKeys,
		KeyCount:              temp.KeyCount,
		Delimiter:             temp.Delimiter,
		IsTruncated:           temp.IsTruncated,
		EncodingType:          temp.EncodingType,
		CommonPrefixes:        temp.CommonPrefixes,
		NextContinuationToken: temp.NextContinuationToken,
		Contents:              contents,
	}
	return &output, nil
}

func (cli *ClientV2) ListObjectsType2(ctx context.Context, input *ListObjectsType2Input) (*ListObjectsType2Output, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	copyInput := *input
	input = &copyInput
	if input.MaxKeys == 0 {
		input.MaxKeys = DefaultListMaxKeys
	}
	if input.ListOnlyOnce {
		return cli.listObjectsType2(ctx, input)
	}
	var output *ListObjectsType2Output
	for {
		res, err := cli.listObjectsType2(ctx, input)
		if err != nil {
			return nil, err
		}
		if output == nil {
			output = res
		} else {
			output.KeyCount += res.KeyCount
			output.IsTruncated = res.IsTruncated
			output.NextContinuationToken = res.NextContinuationToken
			output.Contents = append(output.Contents, res.Contents...)
			output.CommonPrefixes = append(output.CommonPrefixes, res.CommonPrefixes...)
		}
		if !res.IsTruncated || len(res.Contents)+len(output.CommonPrefixes) >= input.MaxKeys {
			break
		}
		input.ContinuationToken = res.NextContinuationToken
		input.MaxKeys = input.MaxKeys - res.KeyCount
	}

	return output, nil
}

// ListObjectVersions list multi-version objects of a bucket
//
// Deprecated: use ListObjectV2Versions of ClientV2 instead
func (bkt *Bucket) ListObjectVersions(ctx context.Context, input *ListObjectVersionsInput, options ...Option) (*ListObjectVersionsOutput, error) {
	res, err := bkt.client.newBuilder(bkt.name, "", options...).
		WithQuery("prefix", input.Prefix).
		WithQuery("delimiter", input.Delimiter).
		WithQuery("key-marker", input.KeyMarker).
		WithQuery("version-id-marker", input.VersionIDMarker).
		WithQuery("max-keys", strconv.Itoa(input.MaxKeys)).
		WithQuery("encoding-type", input.EncodingType).
		WithQuery("fetch-meta", strconv.FormatBool(input.FetchMeta)).
		WithQuery("versions", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	interOutput := listObjectVersionsOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &interOutput); err != nil {
		return nil, err
	}
	output := ListObjectVersionsOutput{
		RequestInfo:         interOutput.RequestInfo,
		Name:                interOutput.Name,
		Prefix:              interOutput.Prefix,
		KeyMarker:           interOutput.KeyMarker,
		VersionIDMarker:     interOutput.VersionIDMarker,
		Delimiter:           interOutput.Delimiter,
		EncodingType:        interOutput.EncodingType,
		MaxKeys:             interOutput.MaxKeys,
		NextKeyMarker:       interOutput.NextKeyMarker,
		NextVersionIDMarker: interOutput.NextVersionIDMarker,
		IsTruncated:         interOutput.IsTruncated,
		CommonPrefixes:      interOutput.CommonPrefixes,
		DeleteMarkers:       interOutput.DeleteMarkers,
	}

	contents := make([]ListedObjectVersion, 0, len(interOutput.Versions))
	for _, content := range interOutput.Versions {
		var hashCrc uint64
		if len(content.HashCrc64ecma) == 0 {
			hashCrc = 0
		} else {
			hashCrc, err = strconv.ParseUint(content.HashCrc64ecma, 10, 64)
			if err != nil {
				return nil, &TosServerError{
					TosError:    newTosErr("tos: server returned invalid HashCrc64Ecma", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
					RequestInfo: RequestInfo{RequestID: res.RequestInfo().RequestID},
				}
			}
		}
		contents = append(contents, ListedObjectVersion{
			Key:           content.Key,
			IsLatest:      content.IsLatest,
			LastModified:  content.LastModified,
			ETag:          content.ETag,
			Size:          content.Size,
			Owner:         content.Owner,
			StorageClass:  content.StorageClass,
			Type:          content.Type,
			VersionID:     content.VersionID,
			Meta:          parseUserMetaData(content.Meta),
			ObjectType:    content.Type,
			HashCrc64ecma: hashCrc,
		})
	}
	output.Versions = contents

	return &output, nil
}

// ListObjectVersionsV2 list multi-version objects of a bucket
func (cli *ClientV2) ListObjectVersionsV2(
	ctx context.Context,
	input *ListObjectVersionsV2Input) (*ListObjectVersionsV2Output, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	res, err := cli.newBuilder(input.Bucket, "").
		WithParams(*input).
		WithQuery("versions", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	temp := listObjectVersionsV2Output{RequestInfo: res.RequestInfo()}

	if err = marshalOutput(res, &temp); err != nil {
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
					TosError:    newTosErr("tos: server returned invalid HashCrc64Ecma", res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
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
			Meta:          parseUserMetaData(version.Meta),
			ObjectType:    version.ObjectType,
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

func (cli *ClientV2) RestoreObject(ctx context.Context, input *RestoreObjectInput) (*RestoreObjectOutput, error) {
	return cli.baseClient.RestoreObject(ctx, input)
}

func (cli *ClientV2) GetFileStatus(ctx context.Context, input *GetFileStatusInput) (*GetFileStatusOutput, error) {
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	bucketType, err := cli.getBucketType(ctx, input.Bucket)
	if err != nil {
		return nil, err
	}
	if bucketType == enum.BucketTypeHNS {
		resp, err := cli.HeadObjectV2(ctx, &HeadObjectV2Input{
			Bucket: input.Bucket,
			Key:    input.Key,
		})
		if err != nil {
			return nil, err
		}
		return &GetFileStatusOutput{
			RequestInfo:  resp.RequestInfo,
			Key:          input.Key,
			Size:         resp.ContentLength,
			LastModified: resp.LastModified,
			Crc32:        resp.Header.Get(HeaderHashCrc32C),
			Crc64:        strconv.FormatUint(resp.HashCrc64ecma, 10),
			Etag:         resp.ETag,
		}, nil
	}
	res, err := cli.newBuilder(input.Bucket, input.Key).
		SetGeneric(input.GenericInput).
		WithQuery("stat", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetFileStatusOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil

}

func (cli *ClientV2) modifyObjectWithInitCrc64(ctx context.Context, input *AppendObjectV2Input) (*modifyObjectOutput, error) {
	if err := isValidNames(input.Bucket, input.Key, cli.isCustomDomain); err != nil {
		return nil, err
	}
	var (
		content       = input.Content
		contentLength = input.ContentLength
		checker       hash.Hash64
	)

	if contentLength <= 0 {
		contentLength = tryResolveLength(content)
	}

	if input.Offset == 0 && contentLength >= 0 {
		headResp, err := cli.HeadObjectV2(ctx, &HeadObjectV2Input{Bucket: input.Bucket, Key: input.Key})
		if err == nil && headResp.ContentLength > 0 {
			return nil, newTosClientError("tos: The object offset of this modify not matched.", nil)
		}
		if err != nil {
			serr, ok := err.(*TosServerError)
			if !ok {
				return nil, err
			}
			if serr.StatusCode == http.StatusNotFound && serr.RequestInfo.EcCode == "0017-00000003" {
				putResp, err := cli.PutObjectV2(ctx, &PutObjectV2Input{
					PutObjectBasicInput: PutObjectBasicInput{
						Bucket:                  input.Bucket,
						Key:                     input.Key,
						ContentLength:           contentLength,
						ContentMD5:              input.ContentMD5,
						ContentSHA256:           input.ContentSHA256,
						CacheControl:            input.CacheControl,
						ContentDisposition:      input.ContentDisposition,
						ContentEncoding:         input.ContentEncoding,
						ContentLanguage:         input.ContentLanguage,
						ContentType:             input.ContentType,
						Expires:                 input.Expires,
						ACL:                     input.ACL,
						GrantFullControl:        input.GrantFullControl,
						GrantRead:               input.GrantRead,
						GrantReadAcp:            input.GrantReadAcp,
						GrantWriteAcp:           input.GrantWriteAcp,
						WebsiteRedirectLocation: input.WebsiteRedirectLocation,
						StorageClass:            input.StorageClass,
						TrafficLimit:            input.TrafficLimit,
						DataTransferListener:    input.DataTransferListener,
						RateLimiter:             input.RateLimiter,
						Meta:                    input.Meta,
						ForbidOverwrite:         true,
					},
					Content: input.Content,
				})

				if err != nil {
					return nil, err
				}

				return &modifyObjectOutput{
					RequestInfo:      putResp.RequestInfo,
					NextModifyOffset: contentLength,
					HashCrc64ecma:    putResp.HashCrc64ecma,
				}, nil
			}
		}
	}

	modifyInput := &modifyObjectInput{
		Bucket:               input.Bucket,
		Key:                  input.Key,
		Offset:               input.Offset,
		Content:              content,
		ContentLength:        contentLength,
		DataTransferListener: input.DataTransferListener,
		RateLimiter:          input.RateLimiter,
		TrafficLimit:         input.TrafficLimit,
	}

	if cli.enableCRC {
		checker = NewCRC(DefaultCrcTable(), input.PreHashCrc64ecma)
	}
	if content != nil {
		content = wrapReader(content, contentLength, modifyInput.DataTransferListener, modifyInput.RateLimiter, &crcChecker{checker: checker})
	}
	rb := cli.newBuilder(modifyInput.Bucket, modifyInput.Key).
		WithQuery("modify", "").
		WithParams(*modifyInput).
		WithContentLength(contentLength).
		WithRetry(nil, NoRetryClassifier{})
	cli.setExpectHeader(rb, contentLength)
	res, err := rb.Request(ctx, http.MethodPost, content, cli.roundTripperWithSlowLog(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	nextOffset := res.Header.Get(HeaderNextModifyOffset)
	nextModifyOffset, err := strconv.ParseInt(nextOffset, 10, 64)
	if err != nil {
		return nil, &TosServerError{
			TosError:    newTosErr(fmt.Sprintf("tos: server return unexpected Next-Modify-Offset header %q", nextOffset), res.RequestUrl, res.RequestInfo().EcCode, res.RequestInfo().RequestID),
			RequestInfo: res.RequestInfo(),
		}
	}
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	if err = checkCrc64(res, checker); err != nil {
		return nil, err
	}
	return &modifyObjectOutput{
		RequestInfo:      res.RequestInfo(),
		NextModifyOffset: nextModifyOffset,
		HashCrc64ecma:    crc64,
	}, nil
}

func (cli *ClientV2) PutSymlink(ctx context.Context, input *PutSymlinkInput) (*PutSymlinkOutput, error) {
	return cli.baseClient.PutSymlink(ctx, input)
}

func (cli *ClientV2) GetSymlink(ctx context.Context, input *GetSymlinkInput) (*GetSymlinkOutput, error) {
	return cli.baseClient.GetSymlink(ctx, input)
}
