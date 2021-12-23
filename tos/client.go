package tos

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
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
type Client struct {
	scheme      string
	host        string
	urlMode     urlMode
	userAgent   string
	credentials Credentials // nullable
	signer      Signer      // nullable
	transport   Transport
	recognizer  ContentTypeRecognizer
	config      Config
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

// WithTransport set Transport
func WithTransport(transport Transport) ClientOption {
	return func(client *Client) {
		client.transport = transport
	}
}

// WithTransportConfig set TransportConfig
func WithTransportConfig(config *TransportConfig) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig = *config
	}
}

// WithReadWriteTimeout set read-write timeout
func WithReadWriteTimeout(readTimeout, writeTimeout time.Duration) ClientOption {
	return func(client *Client) {
		client.config.TransportConfig.ReadTimeout = readTimeout
		client.config.TransportConfig.WriteTimeout = writeTimeout
	}
}

// WithRegion set region
func WithRegion(region string) ClientOption {
	return func(client *Client) {
		client.config.Region = region
	}
}

// WithSigner for self-defined Signer
func WithSigner(signer Signer) ClientOption {
	return func(client *Client) {
		client.signer = signer
	}
}

// WithPathAccessMode url mode is path model or default mode
func WithPathAccessMode(pathAccessMode bool) ClientOption {
	return func(client *Client) {
		if pathAccessMode {
			client.urlMode = urlModePath
		} else {
			client.urlMode = urlModeDefault
		}
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
		scheme = "http"
		host = endpoint
	}

	urlMode = urlModeDefault
	if net.ParseIP(host) != nil {
		urlMode = urlModePath
	}

	return scheme, host, urlMode
}

// NewClient create a new client
//   endpoint: access endpoint
//   options: WithCredentials set Credentials
//     WithRegion set region, this is required if WithCredentials is used
//     WithReadWriteTimeout set read-write timeout
//     WithTransportConfig set TransportConfig
//     WithTransport set self-defined Transport
func NewClient(endpoint string, options ...ClientOption) (*Client, error) {
	client := Client{
		recognizer: ExtensionBasedContentTypeRecognizer{},
		config:     defaultConfig(),
	}
	client.config.Endpoint = endpoint
	client.scheme, client.host, client.urlMode = schemeHost(endpoint)

	for _, option := range options {
		option(&client)
	}

	if client.transport == nil {
		client.transport = NewDefaultTransport(&client.config.TransportConfig)
	}

	if cred := client.credentials; cred != nil && client.signer == nil {
		if len(client.config.Region) == 0 {
			return nil, errors.New("tos: missing Region option")
		}
		client.signer = NewSignV4(cred, client.config.Region)
	}

	if len(client.userAgent) == 0 {
		client.userAgent = fmt.Sprintf("tos-go-sdk/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
	}

	return &client, nil
}

func (cli *Client) newBuilder(bucket, object string, options ...Option) *requestBuilder {
	rb := &requestBuilder{
		Signer:  cli.signer,
		Scheme:  cli.scheme,
		Host:    cli.host,
		Bucket:  bucket,
		Object:  object,
		URLMode: cli.urlMode,
		Query:   make(url.Values),
		Header:  make(http.Header),
	}

	rb.Header.Set(HeaderUserAgent, cli.userAgent)
	if typ := cli.recognizer.ContentType(object); len(typ) > 0 {
		rb.Header.Set(HeaderContentType, typ)
	}

	for _, option := range options {
		option(rb)
	}
	return rb
}

func (cli *Client) roundTrip(ctx context.Context, req *Request, expectedCode int, expectedCodes ...int) (*Response, error) {
	res, err := cli.transport.RoundTrip(ctx, req)
	if err != nil {
		return nil, err
	}

	if err = checkError(res, expectedCode, expectedCodes...); err != nil {
		return nil, err
	}
	return res, nil
}

func (cli *Client) roundTripper(expectedCode int) roundTripper {
	return func(ctx context.Context, req *Request) (*Response, error) {
		return cli.roundTrip(ctx, req, expectedCode)
	}
}

// PreSignedURL create a pre-signed URL
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
func (cli *Client) PreSignedURL(httpMethod string, bucket, objectKey string, ttl time.Duration, options ...Option) (string, error) {
	return cli.newBuilder(bucket, objectKey, options...).
		PreSignedURL(httpMethod, ttl)
}
