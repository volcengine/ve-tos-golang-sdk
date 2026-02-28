package tos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type TosVectorsClient struct {
	scheme                       string
	host                         string
	userAgent                    string
	credentials                  Credentials // nullable
	signer                       Signer      // nullable
	transport                    Transport
	config                       Config
	retry                        *retryer
	logger                       Logger
	except100ContinueThreshold   int64
	userAgentProductName         string
	userAgentSoftName            string
	userAgentSoftVersion         string
	userAgentCustomizedKeyValues map[string]string
}

type TosVectorsClientOption func(*TosVectorsClient)

func (cli *TosVectorsClient) Close() {
	if t, ok := cli.transport.(*DefaultTransport); ok {
		if h, ok := t.client.Transport.(*http.Transport); ok {
			h.CloseIdleConnections()
		}
		if t.resolver != nil {
			t.resolver.Close()
		}
	}

	// stop background credentials refresh if present
	if pc, ok := cli.credentials.(*providerBackedCredentials); ok {
		pc.Stop()
	}

}

func (cli *TosVectorsClient) SetHTTPTransport(transport http.RoundTripper) {
	cli.transport = newDefaultTranposrtWithHTTPTransport(transport)
}

func WithVectorsCredentials(credentials Credentials) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.credentials = credentials
	}
}

// WithEnableVerifySSL set whether a client verifies the server's certificate chain and host name.
func WithVectorsEnableVerifySSL(enable bool) TosVectorsClientOption {
	skip := !enable
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.InsecureSkipVerify = skip
	}
}

// WithRequestTimeout set timeout for single http request
func WithVectorsRequestTimeout(timeout time.Duration) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.ResponseHeaderTimeout = timeout

	}
}

// WithLogger sets the tos sdk logger
func WithVectorsLogger(logger Logger) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.logger = logger
	}
}

// WithConnectionTimeout set timeout for constructing connection
func WithVectorsConnectionTimeout(timeout time.Duration) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.DialTimeout = timeout
	}
}

// WithExcept100ContinueThreshold set threshold for 100 continue
func WithVectorsExcept100ContinueThreshold(threshold int64) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.except100ContinueThreshold = threshold
	}
}

// WithProxy set http Proxy for tos vectors client
func WithVectorsProxy(proxy *Proxy) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.Proxy = proxy
	}
}

func WithVectorsProxyFunc(proxyFunc func(*http.Request) (*url.URL, error)) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.ProxyFunc = proxyFunc
	}
}

// WithMaxConnections set max connections for http client
func WithVectorsMaxConnections(max int) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.MaxIdleConns = max
		client.config.TransportConfig.MaxIdleConnsPerHost = max
		client.config.TransportConfig.MaxConnsPerHost = max
	}
}

// WithIdleConnTimeout set max idle time of a http connection
func WithVectorsIdleConnTimeout(timeout time.Duration) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.IdleConnTimeout = timeout
	}
}

// WithUserAgentSuffix set suffix of user-agent
func WithVectorsUserAgentSuffix(suffix string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.userAgent = strings.Join([]string{client.userAgent, suffix}, " ")
	}
}

// WithDNSCacheTime set dnsCacheTime in Minute
func WithVectorsDNSCacheTime(dnsCacheTime int) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.DNSCacheTime = time.Minute * time.Duration(dnsCacheTime)
	}
}

// // WithMaxRetryCount set MaxRetryCount
func WithVectorsMaxRetryCount(retryCount int) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		if client.retry != nil {
			client.retry.SetBackoff(exponentialBackoff(retryCount, DefaultRetryBackoffBase))
		}
	}
}

// WithHTTPTransport set Transport of http.Client
func WithVectorsHTTPTransport(transport http.RoundTripper) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.transport = newDefaultTranposrtWithHTTPTransport(transport)
	}
}

// WithTransportConfig set TransportConfig
func WithVectorsTransportConfig(config *TransportConfig) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		// client.config never be nil
		client.config.TransportConfig = *config
	}
}

// WithRegion set region
func WithVectorsRegion(region string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		// client.config never be nil
		client.config.Region = region
	}
}

func WithVectorsEndpoint(endpoint string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.Endpoint = endpoint
	}
}

// WithSigner for self-defined Signer
func WithVectorsSigner(signer Signer) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.signer = signer
	}
}

func WithVectorsUserAgentProductName(userAgentProductName string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.userAgentProductName = userAgentProductName
	}
}

func WithVectorsUserAgentSoftName(userAgentSoftName string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.userAgentSoftName = userAgentSoftName
	}
}

func WithVectorsUserAgentSoftVersion(userAgentSoftVersion string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.userAgentSoftVersion = userAgentSoftVersion
	}
}

func WithVectorsUserAgentCustomizedKeyValues(userAgentCustomizedKeyValues map[string]string) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.userAgentCustomizedKeyValues = userAgentCustomizedKeyValues
	}
}

// With the HighLatencyLogThreshold set and assuming the unit is in kilobytes (KB)
func WithVectorsHighLatencyLogThreshold(highLatencyLogThreshold int) TosVectorsClientOption {
	return func(client *TosVectorsClient) {
		client.config.TransportConfig.HighLatencyLogThreshold = &highLatencyLogThreshold
	}
}

func NewTosVectorsClient(options ...TosVectorsClientOption) (*TosVectorsClient, error) {
	client := TosVectorsClient{
		config:                     defaultConfig(),
		retry:                      newRetryer(exponentialBackoff(DefaultRetryTime, DefaultRetryBackoffBase)),
		userAgent:                  fmt.Sprintf("ve-tos-go-sdk/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version()),
		except100ContinueThreshold: enum.DefaultExcept100ContinueThreshold,
	}

	client.retry.SetJitter(0.25)
	err := initVectorsClient(&client, options...)
	if err != nil {
		return nil, err
	}

	return &client, nil
}

func initVectorsClient(client *TosVectorsClient, options ...TosVectorsClientOption) error {
	for _, option := range options {
		option(client)
	}
	client.scheme, client.host, _ = schemeHost(client.config.Endpoint)
	if len(client.config.Region) == 0 {
		if region, ok := SupportedEndpoint()[client.host]; ok {
			client.config.Region = region
		}
	}

	if client.transport == nil {
		transport := NewDefaultTransport(&client.config.TransportConfig)
		transport.WithDefaultTransportLogger(client.logger)
		client.transport = transport
	}

	if cred := client.credentials; cred != nil && client.signer == nil {
		if len(client.config.Region) == 0 {
			return newTosClientError("tos: missing Region option", nil)
		}
		signer := NewSignV4(cred, client.config.Region)
		signer.WithServiceName(VectorServiceName)
		signer.WithSigningKey(vectorsSigningKey)
		signer.WithSignLogger(client.logger)
		client.signer = signer
	}
	return nil
}

func (cli *TosVectorsClient) newBuilder(accountId string, bucket string, options ...Option) *requestBuilder {
	var host string
	if len(accountId) > 0 && len(bucket) > 0 {
		host = bucket + "-" + accountId + "." + cli.host
	} else {
		host = cli.host
	}
	rb := &requestBuilder{
		Signer:     cli.signer,
		Scheme:     cli.scheme,
		Host:       host,
		AccountID:  accountId,
		Query:      make(url.Values),
		Header:     make(http.Header),
		OnRetry:    func(req *Request) error { return nil },
		Classifier: StatusCodeClassifier{},
	}
	rb.Header.Set(HeaderUserAgent, cli.userAgent)
	for _, option := range options {
		option(rb)
	}
	rb.Retry = cli.retry
	return rb
}

func (cli *TosVectorsClient) roundTrip(ctx context.Context, req *Request, expectedCode int, expectedCodes ...int) (*Response, error) {
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

func (cli *TosVectorsClient) roundTripper(expectedCode int, expectedCodes ...int) roundTripper {
	return func(ctx context.Context, req *Request) (*Response, error) {
		resp, err := cli.roundTrip(ctx, req, expectedCode, expectedCodes...)
		return resp, err
	}
}
