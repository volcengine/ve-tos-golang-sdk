package tos

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type urlMode int

const (
	// urlModePath url pattern is http(s)://{bucket}.domain/{object}
	urlModeDefault = 0
	// urlModePath url pattern is http(s)://domain/{bucket}/{object}
	urlModePath           = 1
	contentDispositionSep = ";"
)

type Request struct {
	Scheme        string
	Method        string
	Host          string
	Path          string
	ContentLength *int64
	Content       io.Reader
	Query         url.Values
	Header        http.Header
	enableSlowLog bool
	RequestDate   time.Time // 不为空时，代表本次请求 Header 中指定的 X-Tos-Date 头域（转换为 UTC 时间），包含签名时和发送时
	RequestHost   string    // 不为空时，代表本次请求 Header 中指定的 Host 头域，仅影响签名和发送请求时的 Host 头域，实际建立仍使用 Endpoint
	rawContent    io.Reader
	rawContentLen *int64
}

func (req *Request) URL() string {
	u := url.URL{
		Scheme:   req.Scheme,
		Host:     req.Host,
		Path:     req.Path,
		RawQuery: req.Query.Encode(),
	}
	return u.String()
}

func OnRetryFromStart(req *Request) error {
	if seek, ok := req.Content.(io.Seeker); ok {
		_, err := seek.Seek(0, io.SeekStart)
		return err
	}
	return nil
}

// Range represents a range of an object
type Range struct {
	Start int64
	End   int64
}

// HTTP Range header
func (hr *Range) String() string {
	return fmt.Sprintf("bytes=%d-%d", hr.Start, hr.End)
}

type CopySource struct {
	srcBucket    string
	srcObjectKey string
}

type requestBuilder struct {
	Signer              Signer
	Scheme              string
	Host                string
	Bucket              string
	Object              string
	URLMode             urlMode
	ContentLength       *int64
	Range               *Range
	Query               url.Values
	Header              http.Header
	Retry               *retryer
	OnRetry             func(req *Request) error
	Classifier          classifier
	CopySource          *CopySource
	IsCustomDomain      bool
	DisableEncodingMeta bool
	RequestDate         time.Time
	RequestHost         string
	AccountID           string
	enableTrailerHeader bool
	// CheckETag  bool
	// CheckCRC32 bool
}

func (rb *requestBuilder) WithEnableTrailer(enable bool) *requestBuilder {
	rb.enableTrailerHeader = enable
	return rb
}
func (rb *requestBuilder) WithRetry(onRetry func(req *Request) error, classifier classifier) *requestBuilder {
	if onRetry == nil {
		rb.OnRetry = func(req *Request) error { return nil }
	} else {
		rb.OnRetry = onRetry
	}
	if classifier == nil {
		rb.Classifier = NoRetryClassifier{}
	} else {
		rb.Classifier = classifier
	}
	return rb
}

func (rb *requestBuilder) WithCopySource(srcBucket, srcObjectKey string) *requestBuilder {
	rb.CopySource = &CopySource{
		srcBucket:    srcBucket,
		srcObjectKey: srcObjectKey,
	}
	return rb
}

func (rb *requestBuilder) WithQuery(key, value string) *requestBuilder {
	rb.Query.Add(key, value)
	return rb
}

func (rb *requestBuilder) WithHeader(key, value string) *requestBuilder {
	if len(value) > 0 {
		rb.Header.Set(key, value)
	}
	return rb
}

func convertToString(iface interface{}, tag *reflect.StructTag) string {
	// return empty string if value is zero except filed with "default" tag
	var result string
	switch v := iface.(type) {
	case string:
		result = v
	case int:
		if v != 0 {
			result = strconv.Itoa(v)
		} else {
			result = tag.Get("default")
		}
	case int64:
		if v != 0 {
			result = strconv.Itoa(int(v))
		} else {
			result = tag.Get("default")
		}
	case time.Time:
		if !v.IsZero() {
			result = v.Format(http.TimeFormat)
		}
	case bool:
		result = strconv.FormatBool(v)
	default:
		if reflect.TypeOf(iface).Kind() == reflect.String {
			result = reflect.ValueOf(iface).String()
		}
	}
	return result
}

func encodeContentDisposition(input string) string {
	metas := strings.Split(input, contentDispositionSep)
	res := make([]string, 0, len(metas))
	for _, meta := range metas {
		metaValues := strings.SplitN(meta, "=", 2)
		if len(metaValues) > 1 && strings.TrimSpace(strings.ToLower(metaValues[0])) == "filename" {
			value := escapeHeader(metaValues[1], skipContentDispositionEscape)
			res = append(res, metaValues[0]+"="+value)
		} else {
			res = append(res, meta)
		}
	}
	return strings.Join(res, contentDispositionSep)

}

// WithParams will set filed with tag "header" in input to rb.Header.
func (rb *requestBuilder) WithParams(input interface{}) *requestBuilder {

	t := reflect.TypeOf(input)
	v := reflect.ValueOf(input)
	for i := 0; i < v.NumField(); i++ {
		filed := t.Field(i)
		if filed.Type.Kind() == reflect.Struct {
			rb.WithParams(v.Field(i).Interface())
		}
		location := filed.Tag.Get("location")
		switch location {
		case "header":
			value := convertToString(v.Field(i).Interface(), &filed.Tag)
			if filed.Tag.Get("encodeChinese") == "true" && !rb.DisableEncodingMeta {
				if filed.Tag.Get("locationName") == HeaderContentDisposition {
					value = encodeContentDisposition(value)
				} else {
					value = headerEncode(value)
				}
			}
			rb.WithHeader(filed.Tag.Get("locationName"), value)
		case "headers":
			if headers, ok := v.Field(i).Interface().(map[string]string); ok {
				for k, v := range headers {
					key := k
					value := v
					if !rb.DisableEncodingMeta {
						key = headerEncode(k)
						value = headerEncode(v)
					}
					rb.Header.Set(HeaderMetaPrefix+key, value)
				}
				return rb
			}
		case "query":
			v := convertToString(v.Field(i).Interface(), &filed.Tag)
			if len(v) > 0 {
				rb.WithQuery(filed.Tag.Get("locationName"), v)
			}
		}
	}
	return rb
}

func (rb *requestBuilder) WithContentLength(length int64) *requestBuilder {
	rb.ContentLength = &length
	return rb
}

func (rb *requestBuilder) hostPath() (string, string) {

	if rb.IsCustomDomain {
		if len(rb.Object) > 0 {
			return rb.Host, "/" + rb.Object
		}
		return rb.Host, "/"
	}

	if rb.URLMode == urlModePath {
		if len(rb.Object) > 0 {
			return rb.Host, "/" + rb.Bucket + "/" + rb.Object
		}
		return rb.Host, "/" + rb.Bucket // rb.Bucket may be empty ""
	}
	// URLModeDefault
	if len(rb.Bucket) == 0 {
		return rb.Host, "/"
	}
	return rb.Bucket + "." + rb.Host, "/" + rb.Object
}

func (rb *requestBuilder) build(method string, content io.Reader) *Request {
	host, path := rb.hostPath()
	req := &Request{
		Scheme:      rb.Scheme,
		Method:      method,
		Host:        host,
		Path:        path,
		Content:     content,
		Query:       rb.Query,
		Header:      rb.Header,
		RequestHost: rb.RequestHost,
		RequestDate: rb.RequestDate,
		rawContent:  content,
	}

	if content != nil {
		if rb.ContentLength != nil {
			req.ContentLength = rb.ContentLength
		} else if length := tryResolveLength(content); length >= 0 {
			req.ContentLength = &length
		}
	}
	req.rawContentLen = req.ContentLength
	return req
}

func (rb *requestBuilder) buildSign(req *Request) {
	if rb.Signer != nil {
		signed := rb.Signer.SignHeader(req)
		for key, values := range signed {
			req.Header[key] = values
		}
	}
}
func (rb *requestBuilder) buildTrailers(req *Request) {
	ioCloser := req.Content.(io.ReadCloser)
	c := &readCloserWithCRC{checker: NewCRC(DefaultCrcTable(), 0), base: ioCloser}
	body := io.TeeReader(req.Content, c.checker)
	chunkReader := newTosChunkEncodingReader(*req.ContentLength, map[string]trailerValue{tosChecksumCrc64Header: c}, body)
	body = chunkReader.body
	if req.ContentLength != nil && *req.ContentLength != -1 {
		length := chunkReader.getLength()
		req.ContentLength = &length
	}

	for k, vals := range chunkReader.getHttpHeader() {
		if k == "Content-Encoding" {
			ce := req.Header.Values(k)
			ce = append(vals, ce...)
			req.Header[k] = []string{strings.Join(ce, ",")}
		} else {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}

	}
	req.Content = body
}

func (rb *requestBuilder) Build(method string, content io.Reader) *Request {
	req := rb.build(method, content)
	if rb.enableTrailerHeader {
		rb.buildTrailers(req)
	}

	if rb.CopySource != nil {
		versionID := req.Query.Get("versionId")
		req.Query.Del("versionId")
		req.Header.Add(HeaderCopySource, copySource(rb.CopySource.srcBucket, rb.CopySource.srcObjectKey, versionID))
	}
	rb.buildSign(req)
	return req
}

func (rb *requestBuilder) BuildControl(method string, content io.Reader) *Request {
	req := rb.build(method, content)
	if rb.CopySource != nil {
		versionID := req.Query.Get("versionId")
		req.Query.Del("versionId")
		req.Header.Add(HeaderCopySource, copySource(rb.CopySource.srcBucket, rb.CopySource.srcObjectKey, versionID))
	}
	if rb.Signer != nil {
		signed := rb.Signer.SignHeader(req)
		for key, values := range signed {
			req.Header[key] = values
		}
	}
	return req
}

type roundTripper func(ctx context.Context, req *Request) (*Response, error)

func (rb *requestBuilder) SetGeneric(input GenericInput) *requestBuilder {
	if !input.RequestDate.IsZero() {
		rb.RequestDate = input.RequestDate
	}
	if input.RequestHost != "" {
		rb.RequestHost = input.RequestHost

	}
	return rb
}

func (rb *requestBuilder) Request(ctx context.Context, method string,
	content io.Reader, roundTripper roundTripper) (*Response, error) {

	var (
		req *Request
		res *Response
		err error
	)
	retryAfterSec := int64(-1)

	req = rb.Build(method, content)

	if rb.Retry != nil {
		work := func(retryCount int) (retrySec int64, err error) {
			if retryCount > 0 {
				req.Header.Set("x-sdk-retry-count", "attempt="+strconv.Itoa(retryCount)+"; max="+strconv.Itoa(len(rb.Retry.backoff)))
				err = rb.OnRetry(req)
				if err != nil {
					return -1, err
				}
			}
			res, err = roundTripper(ctx, req)
			if res != nil {
				if retryAfter := res.Header.Get("Retry-After"); retryAfter != "" {
					retryAfterInt, ierr := strconv.ParseInt(retryAfter, 10, 64)
					if ierr == nil {
						retryAfterSec = retryAfterInt
					}
				}
			}
			return retryAfterSec, err
		}
		err = rb.Retry.Run(ctx, work, rb.Classifier)
		if err != nil {
			return nil, err
		}
		return res, err
	}
	res, err = roundTripper(ctx, req)
	return res, err
}

func (rb *requestBuilder) RequestControl(ctx context.Context, method string,
	content io.Reader, path string, roundTripper roundTripper) (*Response, error) {

	var (
		res *Response
		err error
	)
	retryAfterSec := int64(-1)

	req := &Request{
		Scheme:      rb.Scheme,
		Method:      method,
		Host:        rb.Host,
		Path:        path,
		Content:     content,
		Query:       rb.Query,
		Header:      rb.Header,
		RequestHost: rb.RequestHost,
		RequestDate: rb.RequestDate,
	}
	if rb.AccountID == "" {
		return nil, newTosClientError("Control endpoint account id is empty.", nil)
	}

	if rb.Host == "" {
		return nil, newTosClientError("Control endpoint host is empty.", nil)
	}
	req.Header.Set("x-tos-account-id", rb.AccountID)

	if rb.Signer != nil {
		signed := rb.Signer.SignHeader(req)
		for key, values := range signed {
			req.Header[key] = values
		}
	}

	if rb.Retry != nil {
		work := func(retryCount int) (retrySec int64, err error) {
			if retryCount > 0 {
				req.Header.Set("x-sdk-retry-count", "attempt="+strconv.Itoa(retryCount)+"; max="+strconv.Itoa(len(rb.Retry.backoff)))
				req.Content = req.rawContent
				req.ContentLength = req.rawContentLen
				req.Header.Del(v4Date)
				err = rb.OnRetry(req)
				if err != nil {
					return -1, err
				}
				rb.buildTrailers(req)
				rb.buildSign(req)
			}
			res, err = roundTripper(ctx, req)
			if res != nil {
				if retryAfter := res.Header.Get("Retry-After"); retryAfter != "" {
					retryAfterInt, ierr := strconv.ParseInt(retryAfter, 10, 64)
					if ierr == nil {
						retryAfterSec = retryAfterInt
					}
				}
			}
			return retryAfterSec, err
		}
		err = rb.Retry.Run(ctx, work, rb.Classifier)
		if err != nil {
			return nil, err
		}
		return res, err
	}
	res, err = roundTripper(ctx, req)
	return res, err
}

func (rb *requestBuilder) PreSignedURL(method string, ttl time.Duration) (string, error) {
	req := rb.build(method, nil)
	if rb.Signer == nil {
		return req.URL(), nil
	}

	query := rb.Signer.SignQuery(req, ttl)
	for k, v := range query {
		req.Query[k] = v
	}
	return req.URL(), nil
}

type RequestInfo struct {
	RequestID  string
	ID2        string
	StatusCode int
	EcCode     string
	Header     http.Header
}

type Response struct {
	StatusCode    int
	ContentLength int64
	Header        http.Header
	Body          io.ReadCloser
	RequestUrl    string
}

func (r *Response) RequestInfo() RequestInfo {
	return RequestInfo{
		RequestID:  r.Header.Get(HeaderRequestID),
		ID2:        r.Header.Get(HeaderID2),
		StatusCode: r.StatusCode,
		EcCode:     r.Header.Get(HeaderTOSEC),
		Header:     r.Header,
	}
}

func (r *Response) Close() error {
	if r.Body != nil {
		return r.Body.Close()
	}
	return nil
}

func marshalOutput(res *Response, output interface{}) error {
	// Although status code is ok, we need to check if response body is valid.
	// If response body is invalid, TosServerError should be raised. But we can't
	// unmarshal error from response body now.
	reader := res.Body
	requestID := res.RequestInfo().RequestID
	requestURL := res.RequestUrl
	ecCode := res.Header.Get(HeaderTOSEC)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return &TosServerError{
			TosError:    newTosErr("tos: unmarshal response body failed.", requestURL, ecCode, requestID),
			RequestInfo: RequestInfo{RequestID: requestID},
		}
	}
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return &TosServerError{
			TosError:    newTosErr("server returns empty result", requestURL, ecCode, requestID),
			RequestInfo: RequestInfo{RequestID: requestID},
		}
	}
	if err = json.Unmarshal(data, output); err != nil {
		return &TosServerError{
			TosError:    newTosErr(err.Error(), requestURL, ecCode, requestID),
			RequestInfo: RequestInfo{RequestID: requestID},
		}
	}
	return nil
}

func marshalInput(name string, input interface{}) ([]byte, string, error) {
	data, err := json.Marshal(input)
	if err != nil {
		return nil, "", InvalidMarshal
	}

	sum := md5.Sum(data)
	return data, base64.StdEncoding.EncodeToString(sum[:]), nil
}

func fileUnreadLength(file *os.File) (int64, error) {
	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	stat, err := file.Stat()
	if err != nil {
		return 0, err
	}

	size := stat.Size()
	if offset > size || offset < 0 {
		return 0, newTosClientError("tos: unexpected file size and(or) offset", nil)
	}

	return size - offset, nil
}

func tryResolveLength(reader io.Reader) int64 {
	switch v := reader.(type) {
	case *bytes.Buffer:
		return int64(v.Len())
	case *bytes.Reader:
		return int64(v.Len())
	case *strings.Reader:
		return int64(v.Len())
	case *os.File:
		length, err := fileUnreadLength(v)
		if err != nil {
			return -1
		}
		return length
	case *io.LimitedReader:
		return v.N
	case *net.Buffers:
		if v != nil {
			length := int64(0)
			for _, p := range *v {
				length += int64(len(p))
			}
			return length
		}
		return 0
	default:
		return -1
	}
}

func Int64(value int64) *int64 { return &value }
