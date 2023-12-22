package tos

import "time"

type Config struct {
	Endpoint        string
	Region          string
	TransportConfig TransportConfig
}

func defaultConfig() Config {
	return Config{
		TransportConfig: DefaultTransportConfig(),
	}
}

func DefaultTransportConfig() TransportConfig {
	highLatencyLogThreshold := DefaultHighLatencyLogThreshold
	return TransportConfig{
		MaxIdleConns:            1024,
		MaxIdleConnsPerHost:     1024,
		MaxConnsPerHost:         1024,
		DialTimeout:             10 * time.Second,
		KeepAlive:               30 * time.Second,
		IdleConnTimeout:         60 * time.Second,
		TLSHandshakeTimeout:     10 * time.Second,
		ResponseHeaderTimeout:   60 * time.Second,
		ExpectContinueTimeout:   3 * time.Second,
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		HighLatencyLogThreshold: &highLatencyLogThreshold,
	}
}
