package session

import (
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type Session struct {
	transport   tos.Transport
	credentials tos.Credentials
	region      string
}

type Option func(*Session)

// NewSession create tos.Client with some same options, example:
//
//  session := NewSession(
//		WithRegion(region),
//		WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
//	client, err := session.NewClient(endpoint)
func NewSession(options ...Option) *Session {
	var session Session
	for _, option := range options {
		option(&session)
	}

	if session.transport == nil {
		config := tos.DefaultTransportConfig()
		session.transport = tos.NewDefaultTransport(&config)
	}
	return &session
}

// WithCredentials set Credentials
func WithCredentials(credentials tos.Credentials) Option {
	return func(session *Session) {
		session.credentials = credentials
	}
}

// WithTransport set Transport
func WithTransport(transport tos.Transport) Option {
	return func(session *Session) {
		session.transport = transport
	}
}

// WithRegion set region
func WithRegion(region string) Option {
	return func(session *Session) {
		session.region = region
	}
}

// NewClient create tos.Client from a Session, example:
//
//   client, err := session.NewClient(endpoint)
//   // or
//   client, err := session.NewClient(endpoint, someSpecialOptions...)
func (ss *Session) NewClient(endpoint string, options ...tos.ClientOption) (*tos.Client, error) {
	var sessionOptions []tos.ClientOption
	if ss.transport != nil {
		sessionOptions = append(sessionOptions, tos.WithTransport(ss.transport))
	}

	if len(ss.region) > 0 {
		sessionOptions = append(sessionOptions, tos.WithRegion(ss.region))
	}

	if ss.credentials != nil {
		sessionOptions = append(sessionOptions, tos.WithCredentials(ss.credentials))
	}

	return tos.NewClient(endpoint, append(sessionOptions, options...)...)
}
