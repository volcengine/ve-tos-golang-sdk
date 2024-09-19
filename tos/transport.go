package tos

import (
	"context"
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"
)

const DefaultHighLatencyLogThreshold = 100 // KB

type TransportConfig struct {

	// MaxIdleConns same as http.Transport MaxIdleConns. Default is 1024.
	MaxIdleConns int

	// MaxIdleConnsPerHost same as http.Transport MaxIdleConnsPerHost. Default is 1024.
	MaxIdleConnsPerHost int

	// MaxConnsPerHost same as http.Transport MaxConnsPerHost. Default is no limit.
	MaxConnsPerHost int

	// RequestTimeout same as http.Client Timeout
	// Deprecated: use ReadTimeout or WriteTimeout instead
	RequestTimeout time.Duration

	// DialTimeout same as net.Dialer Timeout
	DialTimeout time.Duration

	// KeepAlive same as net.Dialer KeepAlive
	KeepAlive time.Duration

	// IdleConnTimeout same as http.Transport IdleConnTimeout
	IdleConnTimeout time.Duration

	// TLSHandshakeTimeout same as http.Transport TLSHandshakeTimeout
	TLSHandshakeTimeout time.Duration

	// ResponseHeaderTimeout same as http.Transport ResponseHeaderTimeout
	ResponseHeaderTimeout time.Duration

	// ExpectContinueTimeout same as http.Transport ExpectContinueTimeout
	ExpectContinueTimeout time.Duration

	// ReadTimeout see net.Conn SetReadDeadline
	ReadTimeout time.Duration

	// WriteTimeout set net.Conn SetWriteDeadline
	WriteTimeout time.Duration

	// InsecureSkipVerify set tls.Config InsecureSkipVerify
	InsecureSkipVerify bool

	// DNSCacheTime Set Dns Cache Time.
	DNSCacheTime time.Duration

	// Proxy Set http proxy for http client.
	Proxy *Proxy

	//  When HighLatencyLogThreshold is greater than 0, it indicates the activation of high-latency logging, unit: KB.
	HighLatencyLogThreshold *int
}

type Transport interface {
	RoundTrip(context.Context, *Request) (*Response, error)
}

type DefaultTransport struct {
	client                  http.Client
	logger                  Logger
	resolver                *resolver
	highLatencyLogThreshold int
}

func (d *DefaultTransport) WithDefaultTransportLogger(logger Logger) {
	d.logger = logger
}

// NewDefaultTransport create a DefaultTransport with config
func NewDefaultTransport(config *TransportConfig) *DefaultTransport {
	var r *resolver

	if config.DNSCacheTime >= time.Minute {
		r = newResolver(config.DNSCacheTime)
	}

	transport := &http.Transport{
		DialContext: (&TimeoutDialer{
			Dialer: net.Dialer{
				Timeout:   config.DialTimeout,
				KeepAlive: config.KeepAlive,
			},
			resolver:     r,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ResponseHeaderTimeout: config.ResponseHeaderTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		DisableCompression:    true,
		// #nosec G402
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify},
	}

	if config.Proxy != nil && config.Proxy.proxyHost != "" {
		transport.Proxy = http.ProxyURL(config.Proxy.Url())
	}

	highLatencyLogThreshold := DefaultHighLatencyLogThreshold

	if config.HighLatencyLogThreshold != nil {
		highLatencyLogThreshold = *config.HighLatencyLogThreshold
	}

	return &DefaultTransport{
		client: http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: transport,
		},
		resolver:                r,
		highLatencyLogThreshold: highLatencyLogThreshold,
	}
}

// newDefaultTranposrtWithHTTPTransport
func newDefaultTranposrtWithHTTPTransport(transport http.RoundTripper) *DefaultTransport {
	return &DefaultTransport{
		client: http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: transport,
		},
	}
}

// NewDefaultTransportWithClient crate a DefaultTransport with a http.Client
func NewDefaultTransportWithClient(client http.Client) *DefaultTransport {
	return &DefaultTransport{client: client}
}

type slowDetectWrap struct {
	Size int
	Body io.Reader
}

func isSlow(totalSize, highLatencyLogThreshold int, cost time.Duration) bool {
	if highLatencyLogThreshold <= 0 {
		return false
	}
	if cost > time.Millisecond*500 && int(float64(totalSize)/(float64(cost)/float64(time.Second))) < highLatencyLogThreshold*1024 {
		return true
	}
	return false
}

func (s *slowDetectWrap) Close() error {
	if closer, ok := s.Body.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func (s *slowDetectWrap) Read(p []byte) (n int, err error) {
	n, err = s.Body.Read(p)
	if n > 0 {
		s.Size += n
	}
	return n, err
}

func (dt *DefaultTransport) RoundTrip(ctx context.Context, req *Request) (*Response, error) {
	hr, err := http.NewRequestWithContext(ctx, req.Method, req.URL(), req.Content)
	if err != nil {
		return nil, newTosClientError(err.Error(), err)
	}

	if req.ContentLength != nil {
		hr.ContentLength = *req.ContentLength
	}

	for key, values := range req.Header {
		hr.Header[key] = values
	}

	var accessLog *accessLogRequest
	if dt.logger != nil || req.enableSlowLog {
		var trace *httptrace.ClientTrace
		trace, accessLog = getClientTrace(GetUnixTimeMs())
		ctx = httptrace.WithClientTrace(ctx, trace)
		hr = hr.WithContext(ctx)
	}
	var wrap *slowDetectWrap
	start := time.Now()
	if req.enableSlowLog && hr.Body != nil {
		wrap = &slowDetectWrap{Body: hr.Body}
		hr.Body = wrap
	}

	res, err := dt.client.Do(hr)

	if req.enableSlowLog && hr.Body != nil && wrap != nil && isSlow(wrap.Size, dt.highLatencyLogThreshold, time.Since(start)) {
		accessLog.printSlowLog(dt.logger, hr, res, start, err)
	} else if accessLog != nil {
		accessLog.printAccessLog(dt.logger, hr, res, start, err)
	}

	if err != nil {
		return nil, newTosClientError(err.Error(), err).withUrl(req.URL())
	}

	return &Response{
		StatusCode:    res.StatusCode,
		ContentLength: res.ContentLength,
		Header:        res.Header,
		Body:          res.Body,
		RequestUrl:    req.URL(),
	}, nil
}

type TimeoutConn struct {
	net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	zero         time.Time
}

func NewTimeoutConn(conn net.Conn, readTimeout, writeTimeout time.Duration) *TimeoutConn {
	return &TimeoutConn{
		Conn:         conn,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}
}

func (tc *TimeoutConn) Read(b []byte) (n int, err error) {
	timeout := tc.readTimeout > 0
	if timeout {
		_ = tc.SetReadDeadline(time.Now().Add(tc.readTimeout))
	}

	n, err = tc.Conn.Read(b)
	if timeout {
		_ = tc.SetReadDeadline(time.Now().Add(tc.readTimeout * 5))
	}
	return n, err
}

func (tc *TimeoutConn) Write(b []byte) (n int, err error) {
	timeout := tc.writeTimeout > 0
	if timeout {
		_ = tc.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
	}

	n, err = tc.Conn.Write(b)
	if tc.readTimeout > 0 {
		_ = tc.SetReadDeadline(time.Now().Add(tc.readTimeout * 5))
	}
	return n, err
}

type TimeoutDialer struct {
	net.Dialer
	resolver     *resolver
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (d *TimeoutDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.resolver != nil {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ipList, err := d.resolver.GetIpList(ctx, host)
		if err != nil {
			return nil, err
		}

		// 随机打乱 IP List
		rand.Shuffle(len(ipList), func(i, j int) {
			ipList[i], ipList[j] = ipList[j], ipList[i]
		})

		for _, ip := range ipList {
			dialAddress := net.JoinHostPort(ip, port)
			conn, err := d.Dialer.DialContext(ctx, network, dialAddress)
			if err == nil {
				return NewTimeoutConn(conn, d.ReadTimeout, d.WriteTimeout), nil
			} else {
				d.resolver.Remove(host, ip)
			}
		}
	}

	conn, err := d.Dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return NewTimeoutConn(conn, d.ReadTimeout, d.WriteTimeout), nil
}
