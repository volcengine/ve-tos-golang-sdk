package tos

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"sync/atomic"
	"time"
)

type accessLogRequest struct {
	clientDnsCost                int64
	clientDialCost               int64
	clientTlsHandShakeCost       int64
	clientSendHeadersAndBodyCost int64
	clientWaitResponseCost       int64
	clientSendRequestCost        int64
	actionStartMs                int64
}

func newAccessLogRequest(actionStartMs int64) *accessLogRequest {
	return &accessLogRequest{
		clientDnsCost:                -1,
		clientDialCost:               -1,
		clientTlsHandShakeCost:       -1,
		clientSendHeadersAndBodyCost: -1,
		clientWaitResponseCost:       -1,
		clientSendRequestCost:        -1,
		actionStartMs:                actionStartMs,
	}
}
func (r *accessLogRequest) printAccessLog(logger Logger, req *http.Request, response *http.Response, start time.Time, err error) {
	if logger == nil {
		return
	}
	atomic.CompareAndSwapInt64(&r.clientSendRequestCost, -1, GetUnixTimeMs()-r.actionStartMs)
	var requestId *string
	if response != nil {
		requestId = StringPtr(response.Header.Get(HeaderRequestID))
	}
	prefix := buildPrefix(requestId)
	if req != nil {
		logger.Debug(fmt.Sprintf("%s method: %s, host: %s, request uri: %s, dns cost: %d ms, dial cost: %d ms, tls handshake cost: %d ms, send headers and body cost: %d ms, wait response cost: %d ms, request cost: %d ms",
			prefix, req.Method, req.URL.Host, req.URL.EscapedPath(), r.clientDnsCost, r.clientDialCost, r.clientTlsHandShakeCost,
			r.clientSendHeadersAndBodyCost, r.clientWaitResponseCost, r.clientSendRequestCost))
	} else {
		logger.Debug(fmt.Sprintf("%s dns cost: %d ms, dial cost: %d ms, tls handshake cost: %d ms, send headers and body cost: %d ms, wait response cost: %d ms, request cost: %d ms",
			prefix, r.clientDnsCost, r.clientDialCost, r.clientTlsHandShakeCost, r.clientSendHeadersAndBodyCost, r.clientWaitResponseCost, r.clientSendRequestCost))
	}
	if err != nil {
		logger.Info(fmt.Sprintf("[tos]  http error:%s, Cost:%d ms.", err.Error(), time.Since(start).Milliseconds()))
	} else {
		logger.Info(fmt.Sprintf("%s Response StatusCode:%d, Cost:%d ms", prefix, response.StatusCode, time.Since(start).Milliseconds()))
	}
}

func (r *accessLogRequest) printSlowLog(logger Logger, req *http.Request, response *http.Response, start time.Time, err error) {
	if logger == nil {
		logger = stdlog
	}
	atomic.CompareAndSwapInt64(&r.clientSendRequestCost, -1, GetUnixTimeMs()-r.actionStartMs)
	var requestId *string
	if response != nil {
		requestId = StringPtr(response.Header.Get(HeaderRequestID))
	}
	prefix := buildSlowPrefix(requestId)
	if req != nil {
		logger.Warn(fmt.Sprintf("%s, method: %s, host: %s, request uri: %s, dns cost: %d ms, dial cost: %d ms, tls handshake cost: %d ms, send headers and body cost: %d ms, wait response cost: %d ms, request cost: %d ms",
			prefix, req.Method, req.URL.Host, req.URL.EscapedPath(), r.clientDnsCost, r.clientDialCost, r.clientTlsHandShakeCost,
			r.clientSendHeadersAndBodyCost, r.clientWaitResponseCost, r.clientSendRequestCost))
	} else {
		logger.Warn(fmt.Sprintf("%s, dns cost: %d ms, dial cost: %d ms, tls handshake cost: %d ms, send headers and body cost: %d ms, wait response cost: %d ms, request cost: %d ms",
			prefix, r.clientDnsCost, r.clientDialCost, r.clientTlsHandShakeCost, r.clientSendHeadersAndBodyCost, r.clientWaitResponseCost, r.clientSendRequestCost))
	}
	// GET 请求整体耗时需要在读流后打印
	if req.Method == http.MethodGet {
		return
	}
	if err != nil {
		logger.Warn(fmt.Sprintf("%s http error:%s, Cost:%d ms.", prefix, err.Error(), time.Since(start).Milliseconds()))
		return
	}
	logger.Warn(fmt.Sprintf("%s Response StatusCode:%d, Cost:%d ms", prefix, response.StatusCode, time.Since(start).Milliseconds()))
}
func buildSlowPrefix(requestId *string) string {
	prefix := "[tos slow]"

	if requestId != nil {
		prefix = fmt.Sprintf("[tos slow log] requestId: %s", *requestId)
	}
	return prefix
}

func buildPrefix(requestId *string) string {
	prefix := "[tos]"

	if requestId != nil {
		prefix = fmt.Sprintf("[tos] requestId: %s", *requestId)
	}
	return prefix
}

func getClientTrace(actionStartMs int64) (*httptrace.ClientTrace, *accessLogRequest) {
	var dnsStart int64
	var dialStart int64
	var tlsHandShakeStart int64
	var sendHeadersAndBodyStart int64
	var waitResponseStart int64
	r := newAccessLogRequest(actionStartMs)
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			atomic.StoreInt64(&r.clientWaitResponseCost, GetUnixTimeMs()-atomic.LoadInt64(&waitResponseStart))
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			atomic.StoreInt64(&dnsStart, GetUnixTimeMs())
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			atomic.StoreInt64(&r.clientDnsCost, GetUnixTimeMs()-atomic.LoadInt64(&dnsStart))
		},
		ConnectStart: func(network, addr string) {
			atomic.StoreInt64(&dialStart, GetUnixTimeMs())
		},
		ConnectDone: func(network, addr string, err error) {
			now := GetUnixTimeMs()
			atomic.StoreInt64(&sendHeadersAndBodyStart, now)
			atomic.StoreInt64(&r.clientDialCost, now-atomic.LoadInt64(&dialStart))
		},
		TLSHandshakeStart: func() {
			atomic.StoreInt64(&tlsHandShakeStart, GetUnixTimeMs())
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			now := GetUnixTimeMs()
			atomic.StoreInt64(&sendHeadersAndBodyStart, now)
			atomic.StoreInt64(&r.clientTlsHandShakeCost, now-atomic.LoadInt64(&tlsHandShakeStart))
		},

		GotConn: func(httptrace.GotConnInfo) {
			atomic.StoreInt64(&sendHeadersAndBodyStart, GetUnixTimeMs())
		},

		WroteRequest: func(info httptrace.WroteRequestInfo) {
			atomic.StoreInt64(&waitResponseStart, GetUnixTimeMs())
			atomic.StoreInt64(&r.clientSendHeadersAndBodyCost, atomic.LoadInt64(&waitResponseStart)-atomic.LoadInt64(&sendHeadersAndBodyStart))
		},
	}
	return trace, r
}
