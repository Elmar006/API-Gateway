// Package proxy provides utilities for building reverse proxy handlers.
package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// Config holds settings for the reverse proxy behavior.
type Config struct {
	FlushInterval   time.Duration
	DialTimeout     time.Duration
	ResponseTimeout time.Duration
	EnableWebSocket bool
}

func DefaultConfig() Config {
	return Config{
		FlushInterval:   50 * time.Millisecond,
		DialTimeout:     10 * time.Second,
		ResponseTimeout: 0,
		EnableWebSocket: true,
	}
}

// ReverseProxyBuilder constructs httputil.ReverseProxy instances for target backends.
type ReverseProxyBuilder struct {
	targetURL *url.URL
	config    Config
}

func NewReverseProxyBuilder(targetURL string, config Config) (*ReverseProxyBuilder, error) {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid target URL: %w", err)
	}
	return &ReverseProxyBuilder{
		targetURL: parsed,
		config:    config,
	}, nil
}

func (b *ReverseProxyBuilder) Build() *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(b.targetURL)
	proxy.Transport = b.createTransport()
	proxy.FlushInterval = b.config.FlushInterval

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		b.modifyRequest(req)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("Bad Gateway"))
	}

	if b.config.EnableWebSocket {
		b.enableWebSocket(proxy)
	}

	return proxy
}

func (b *ReverseProxyBuilder) createTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   b.config.DialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: b.config.ResponseTimeout,
	}
}

func (b *ReverseProxyBuilder) modifyRequest(req *http.Request) {
	req.Host = b.targetURL.Host

	if req.Header.Get("X-Forwarded-For") == "" {
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	}
	if req.Header.Get("X-Real-IP") == "" {
		req.Header.Set("X-Real-IP", req.RemoteAddr)
	}

	req.Header.Del("Connection")
}

func (b *ReverseProxyBuilder) enableWebSocket(proxy *httputil.ReverseProxy) {
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if strings.ToLower(req.Header.Get("Connection")) == "upgrade" &&
			strings.ToLower(req.Header.Get("Upgrade")) == "websocket" {
			req.Header.Set("Connection", "Upgrade")
			req.Header.Set("Upgrade", "websocket")
		}
	}
}
