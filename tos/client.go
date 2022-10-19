package tos

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const (
	signPolicyDate                  = "x-tos-date"
	signPolicyCredential            = "x-tos-credential"
	signPolicyAlgorithm             = "x-tos-algorithm"
	signPolicySecurityToken         = "x-tos-security-token"
	signPolicyExpiration            = "expiration"
	signConditionContentLengthRange = "content-length-range"
	signConditionBucket             = "bucket"
	signConditionKey                = "key"
	signConditions                  = "conditions"
	maxPreSignExpires               = 604800 // 7 day
	defaultSignExpires              = 3600   // 1 hour
)

// Client TOS Client
// use NewClient to create a new Client
//
// example:
//   client, err := NewClient(endpoint, WithCredentials(credentials), WithRegion(region))
//   if err != nil {
//      // ...
//   }
//   // do something
//
// if you only access the public bucket:
//   client, err := NewClient(endpoint)
//   // do something
//
// Deprecated: use ClientV2 instead
type Client struct {
	scheme       string
	host         string
	urlMode      urlMode
	userAgent    string
	credentials  Credentials // nullable
	signer       Signer      // nullable
	transport    Transport
	recognizer   ContentTypeRecognizer
	config       Config
	retry        *retryer
	dnsCacheTime time.Duration // milliseconds
	enableCRC    bool
	logger       Logger
}

// ClientV2 TOS ClientV2
// use NewClientV2 to create a new ClientV2
//
// example:
//   client, err := NewClientV2(endpoint, WithCredentials(credentials), WithRegion(region))
//   if err != nil {
//      // ...
//   }
//   // do something
//
// if you only access the public bucket:
//   client, err := NewClientV2(endpoint)
//   // do something
//
type ClientV2 struct {
	Client
}

func (cli *ClientV2) Close() {
	if t, ok := cli.transport.(*DefaultTransport); ok {
		if h, ok := t.client.Transport.(*http.Transport); ok {
			h.CloseIdleConnections()
		}
	}
}

func (cli *ClientV2) SetHTTPTransport(transport http.RoundTripper) {
	cli.transport = newDefaultTranposrtWithHTTPTransport(transport)
}

type ClientOption func(*Client)

// WithCredentials set Credentials
//
// see StaticCredentials, WithoutSecretKeyCredentials and FederationCredentials
func WithCredentials(credentials Credentials) ClientOption {
	return func(client *Client) {
		client.credentials = credentials
	}
}

// WithEnableVerifySSL set whether a client verifies the server's certificate chain and host name.
func WithEnableVerifySSL(enable bool) ClientOption {
	skip := !enable
	return func(client *Client) {
		client.config.TransportConfig.InsecureSkipVerify = skip
	}
}

// WithRequestTimeout set timeout for single http request
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.ResponseHeaderTimeout = timeout

	}
}

//
// WithLogger sets the tos sdk logger
//
func WithLogger(logger Logger) ClientOption {
	return func(client *Client) {
		client.logger = logger
	}
}

// WithConnectionTimeout set timeout for constructing connection
func WithConnectionTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.DialTimeout = timeout
	}
}

// WithProxy set http Proxy for tos client
func WithProxy(proxy *Proxy) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.Proxy = proxy
	}
}

// WithMaxConnections set maximum number of http connections
func WithMaxConnections(max int) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.MaxIdleConns = max
		client.config.TransportConfig.MaxIdleConnsPerHost = max
		client.config.TransportConfig.MaxConnsPerHost = max
	}

}

// WithIdleConnTimeout set max idle time of a http connection
func WithIdleConnTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.IdleConnTimeout = timeout
	}
}

// WithUserAgentSuffix set suffix of user-agent
func WithUserAgentSuffix(suffix string) ClientOption {
	return func(client *Client) {
		client.userAgent = strings.Join([]string{client.userAgent, suffix}, " ")
	}
}

// WithDNSCacheTime set dnsCacheTime in Minute
func WithDNSCacheTime(dnsCacheTime int) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.DNSCacheTime = time.Minute * time.Duration(dnsCacheTime)
	}
}

// WithEnableCRC set if check crc after uploading object.
// Checking crc is enabled by default.
func WithEnableCRC(enableCRC bool) ClientOption {
	return func(client *Client) {
		client.enableCRC = enableCRC
	}
}

// // WithMaxRetryCount set MaxRetryCount
func WithMaxRetryCount(retryCount int) ClientOption {
	return func(client *Client) {
		if client.retry != nil {
			client.retry.SetBackoff(exponentialBackoff(retryCount, DefaultRetryBackoffBase))
		}
	}
}

// WithTransport set Transport
//
// Deprecated: this function is Deprecated.
// If you want to set http.Transport use WithHTTPTransport instead
func WithTransport(transport Transport) ClientOption {
	return func(client *Client) {
		client.transport = transport
	}
}

// WithHTTPTransport set Transport of http.Client
func WithHTTPTransport(transport http.RoundTripper) ClientOption {
	return func(client *Client) {
		client.transport = newDefaultTranposrtWithHTTPTransport(transport)
	}
}

// WithTransportConfig set TransportConfig
func WithTransportConfig(config *TransportConfig) ClientOption {
	return func(client *Client) {
		// client.config never be nil
		client.config.TransportConfig = *config
	}
}

// WithSocketTimeout set read-write timeout
func WithSocketTimeout(readTimeout, writeTimeout time.Duration) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.ReadTimeout = readTimeout
		client.config.TransportConfig.WriteTimeout = writeTimeout
	}
}

// WithRegion set region
func WithRegion(region string) ClientOption {
	return func(client *Client) {
		// client.config never be nil
		client.config.Region = region
		if endpoint, ok := SupportedRegion()[region]; ok {
			if len(client.config.Endpoint) == 0 {
				client.config.Endpoint = endpoint
			}
		}
	}
}

// WithSigner for self-defined Signer
func WithSigner(signer Signer) ClientOption {
	return func(client *Client) {
		client.signer = signer
	}
}

// WithPathAccessMode url mode is path model or default mode
//
// Deprecated: This option is deprecated. Setting PathAccessMode will be ignored silently.
func WithPathAccessMode(pathAccessMode bool) ClientOption {
	return func(client *Client) {
	}
}

// WithAutoRecognizeContentType set to recognize Content-Type or not, the default is enabled.
func WithAutoRecognizeContentType(enable bool) ClientOption {
	return func(client *Client) {
		if enable {
			client.recognizer = ExtensionBasedContentTypeRecognizer{}
		} else {
			client.recognizer = EmptyContentTypeRecognizer{}
		}
	}
}

// WithContentTypeRecognizer set ContentTypeRecognizer to recognize Content-Type,
// the default is ExtensionBasedContentTypeRecognizer
func WithContentTypeRecognizer(recognizer ContentTypeRecognizer) ClientOption {
	return func(client *Client) {
		client.recognizer = recognizer
	}
}

func schemeHost(endpoint string) (scheme string, host string, urlMode urlMode) {
	if strings.HasPrefix(endpoint, "https://") {
		scheme = "https"
		host = endpoint[len("https://"):]
	} else if strings.HasPrefix(endpoint, "http://") {
		scheme = "http"
		host = endpoint[len("http://"):]
	} else {
		scheme = "https"
		host = endpoint
	}
	urlMode = urlModeDefault
	hostWithoutPort, _, _ := net.SplitHostPort(host)
	if net.ParseIP(host) != nil || net.ParseIP(hostWithoutPort) != nil {
		urlMode = urlModePath
	}

	return scheme, host, urlMode
}

func initClient(client *Client, endpoint string, options ...ClientOption) error {
	client.config.Endpoint = endpoint
	for _, option := range options {
		option(client)
	}
	client.scheme, client.host, client.urlMode = schemeHost(client.config.Endpoint)
	if client.transport == nil {
		transport := NewDefaultTransport(&client.config.TransportConfig)
		transport.WithDefaultTransportLogger(client.logger)
		client.transport = transport
	}
	if cred := client.credentials; cred != nil && client.signer == nil {
		if len(client.config.Region) == 0 {
			if region, ok := SupportedEndpoint()[client.host]; ok {
				client.config.Region = region
			} else {
				return newTosClientError("tos: missing Region option", nil)
			}
		}
		signer := NewSignV4(cred, client.config.Region)
		signer.WithSignLogger(client.logger)
		client.signer = signer
	}
	return nil
}

// NewClient create a new Tos Client
//   endpoint: access endpoint
//   options: WithCredentials set Credentials
//     WithRegion set region, this is required if WithCredentials is used
//     WithSocketTimeout set read-write timeout
//     WithTransportConfig set TransportConfig
//     WithTransport set self-defined Transport
func NewClient(endpoint string, options ...ClientOption) (*Client, error) {
	client := Client{
		recognizer: ExtensionBasedContentTypeRecognizer{},
		config:     defaultConfig(),
		userAgent:  fmt.Sprintf("tos-go-sdk/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version()),
		retry:      newRetryer([]time.Duration{}),
	}
	client.retry.SetJitter(0.25)
	err := initClient(&client, endpoint, options...)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

// NewClientV2 create a new Tos ClientV2
//   endpoint: access endpoint
//   options: WithCredentials set Credentials
//     WithRegion set region, this is required if WithCredentials is used.
//     If Region is supported and the Endpoint parameter is not set, the Endpoint will be resolved automatically
//     WithSocketTimeout set read-write timeout
//     WithTransportConfig set TransportConfig
//     WithTransport set self-defined Transport
//     WithLogger set self-defined Logger
//     WithEnableCRC set CRC switch.
//     WithMaxRetryCount  set Max Retry Count
func NewClientV2(endpoint string, options ...ClientOption) (*ClientV2, error) {
	client := ClientV2{
		Client: Client{
			recognizer: ExtensionBasedContentTypeRecognizer{},
			config:     defaultConfig(),
			retry:      newRetryer([]time.Duration{}),
			userAgent:  fmt.Sprintf("tos-go-sdk/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version()),
			enableCRC:  true,
		},
	}
	client.retry.SetJitter(0.25)
	err := initClient(&client.Client, endpoint, options...)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func (cli *Client) newBuilder(bucket, object string, options ...Option) *requestBuilder {
	rb := &requestBuilder{
		Signer:     cli.signer,
		Scheme:     cli.scheme,
		Host:       cli.host,
		Bucket:     bucket,
		Object:     object,
		URLMode:    cli.urlMode,
		Query:      make(url.Values),
		Header:     make(http.Header),
		OnRetry:    func(req *Request) error { return nil },
		Classifier: StatusCodeClassifier{},
	}
	rb.Header.Set(HeaderUserAgent, cli.userAgent)
	if typ := cli.recognizer.ContentType(object); len(typ) > 0 {
		rb.Header.Set(HeaderContentType, typ)
	}
	for _, option := range options {
		option(rb)
	}
	rb.Retry = cli.retry
	return rb
}

func (cli *Client) roundTrip(ctx context.Context, req *Request, expectedCode int, expectedCodes ...int) (*Response, error) {
	res, err := cli.transport.RoundTrip(ctx, req)
	if err != nil {
		return nil, err
	}
	readBody := req.Method != http.MethodHead
	if err = checkError(res, readBody, expectedCode, expectedCodes...); err != nil {
		return nil, err
	}
	return res, nil
}

func (cli *Client) roundTripper(expectedCode int) roundTripper {
	return func(ctx context.Context, req *Request) (*Response, error) {
		start := time.Now()
		resp, err := cli.roundTrip(ctx, req, expectedCode)
		if cli.logger != nil {
			if err != nil {
				cli.logger.Info(fmt.Sprintf("[tos] http error:%s.", err.Error()))
			} else {
				cli.logger.Info(fmt.Sprintf("[tos] Response StatusCode:%d, RequestId:%s, Cost:%d ms", resp.StatusCode, resp.RequestInfo().RequestID, time.Since(start).Milliseconds()))
			}
		}
		return resp, err
	}
}

// PreSignedURL return pre-signed url
//   httpMethod: HTTP method, {
//     PutObject: http.MethodPut
//     GetObject: http.MethodGet
//     HeadObject: http.MethodHead
//     DeleteObject: http.MethodDelete
//   },
//   bucket: the bucket name
//   objectKey: the object name
//   ttl: the time-to-live of signed URL
//   options: WithVersionID the version id of the object
//  Deprecated: use PreSignedURL of ClientV2 instead
func (cli *Client) PreSignedURL(httpMethod string, bucket, objectKey string, ttl time.Duration, options ...Option) (string, error) {
	if err := isValidNames(bucket, objectKey); err != nil {
		return "", err
	}
	return cli.newBuilder(bucket, objectKey, options...).
		PreSignedURL(httpMethod, ttl)
}

// PreSignedURL return pre-signed url
func (cli *ClientV2) PreSignedURL(input *PreSignedURLInput) (*PreSignedURLOutput, error) {
	if err := IsValidBucketName(input.Bucket); err != nil {
		return nil, err
	}
	rb := cli.newBuilder(input.Bucket, input.Key)

	if input.AlternativeEndpoint != "" {
		_, host, _ := schemeHost(input.AlternativeEndpoint)
		rb.Host = host
	}
	for k, v := range input.Header {
		rb.WithHeader(k, v)
	}
	for k, v := range input.Query {
		rb.WithQuery(k, v)
	}
	if input.Expires == 0 {
		input.Expires = defaultPreSignedURLExpires
	}

	if input.Expires > maxPreSignedURLExpires {
		return nil, InvalidPreSignedURLExpires
	}

	signedURL, err := rb.PreSignedURL(string(input.HTTPMethod), time.Second*time.Duration(input.Expires))
	if err != nil {
		return nil, err
	}
	signed := make(map[string]string)
	for k := range rb.Header {
		signed[k] = rb.Header.Get(k)
	}
	output := &PreSignedURLOutput{
		SignedUrl:    signedURL,
		SignedHeader: signed,
	}
	return output, nil
}

func (cli *ClientV2) PreSignedPostSignature(ctx context.Context, input *PreSingedPostSignatureInput) (*PreSingedPostSignatureOutput, error) {
	algorithm := signPrefix
	postPolicy := make(map[string]interface{})
	cred := cli.credentials.Credential()
	region := cli.config.Region
	date := UTCNow()
	if input.Expires == 0 {
		input.Expires = defaultSignExpires
	}
	if input.Expires > maxPreSignedURLExpires {
		return nil, InvalidPreSignedURLExpires
	}

	postPolicy[signPolicyExpiration] = date.Add(time.Second * time.Duration(input.Expires)).Format(serverTimeFormat)

	cond := make([]interface{}, 0)
	credential := fmt.Sprintf("%s/%s/%s/tos/request", cred.AccessKeyID, date.Format(yyMMdd), region)
	cond = append(cond, map[string]string{signPolicyAlgorithm: algorithm})
	cond = append(cond, map[string]string{signPolicyCredential: credential})
	cond = append(cond, map[string]string{signPolicyDate: date.Format(iso8601Layout)})

	if cred.SecurityToken != "" {
		cond = append(cond, map[string]string{signPolicySecurityToken: cred.SecurityToken})
	}

	if input.Bucket != "" {
		cond = append(cond, map[string]string{signConditionBucket: input.Bucket})
	}

	if input.Key != "" {
		cond = append(cond, map[string]string{signConditionKey: input.Key})
	}

	for _, condition := range input.Conditions {
		if condition.Operator != nil {
			cond = append(cond, []string{*condition.Operator, "$" + condition.Key, condition.Value})
		} else {
			cond = append(cond, map[string]string{condition.Key: condition.Value})

		}
	}
	if input.ContentLengthRange != nil {
		cond = append(cond, []interface{}{signConditionContentLengthRange, input.ContentLengthRange.RangeStart, input.ContentLengthRange.RangeEnd})
	}
	postPolicy[signConditions] = cond
	originPolicy, err := json.Marshal(postPolicy)
	if err != nil {
		return nil, InvalidMarshal
	}
	signK := SigningKey(&SigningKeyInfo{
		Date:       date.Format(yyMMdd),
		Region:     region,
		Credential: &cred,
	})
	policy := base64.StdEncoding.EncodeToString(originPolicy)
	return &PreSingedPostSignatureOutput{
		OriginPolicy: string(originPolicy),
		Policy:       policy,
		Algorithm:    signPrefix,
		Credential:   credential,
		Date:         date.Format(iso8601Layout),
		Signature:    hex.EncodeToString(hmacSHA256(signK, []byte(policy))),
	}, nil

}

func (cli *ClientV2) FetchObjectV2(ctx context.Context, input *FetchObjectInputV2) (*FetchObjectOutputV2, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}
	if err := isValidStorageClass(input.StorageClass); len(input.StorageClass) > 0 && err != nil {
		return nil, InvalidStorageClass
	}

	if err := isValidACL(input.ACL); len(input.ACL) > 0 && err != nil {
		return nil, InvalidACL
	}

	data, contentMD5, err := marshalInput("FetchObjectInputV2", &fetchObjectInput{
		URL:           input.URL,
		IgnoreSameKey: input.IgnoreSameKey,
		ContentMD5:    input.HexMD5,
	})
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, input.Key).
		WithQuery("fetch", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithParams(*input).
		WithRetry(nil, ServerErrorClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := FetchObjectOutputV2{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}

	output.VersionID = res.Header.Get(HeaderVersionID)
	output.SSECAlgorithm = res.Header.Get(HeaderSSECustomerAlgorithm)
	output.SSECKeyMD5 = res.Header.Get(HeaderCopySourceSSECKeyMD5)
	return &output, nil
}

func (cli *ClientV2) PutFetchTaskV2(ctx context.Context, input *PutFetchTaskInputV2) (*PutFetchTaskOutputV2, error) {

	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	if err := isValidStorageClass(input.StorageClass); len(input.StorageClass) > 0 && err != nil {
		return nil, err
	}

	if err := isValidACL(input.ACL); len(input.ACL) > 0 && err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutFetchTaskInputV2", putFetchTaskV2Input{
		URL:           input.URL,
		IgnoreSameKey: input.IgnoreSameKey,
		HexMD5:        input.HexMD5,
		Object:        input.Key,
	})
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		WithQuery("fetchTask", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithParams(*input).
		WithRetry(nil, ServerErrorClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutFetchTaskOutputV2{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(output.RequestID, res.Body, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
