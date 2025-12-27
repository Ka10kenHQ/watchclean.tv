package proxy

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"
)


type ProxyClient struct {
	pool *Pool
}

func NewProxyClient(pool *Pool) *ProxyClient {
	return &ProxyClient {pool: pool}
}

func (pc *ProxyClient) Get() (*http.Client, *Proxy, error) {
	proxy, err := pc.pool.GetRandom()
	if err != nil {
		return nil, nil, err
	}

	proxyUrl, err := url.Parse(proxy.Addr)
	if err != nil {
		pc.pool.MakeError(proxy)
		return nil, nil, err
	}

	// Cloudflare DNS
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 2 * time.Second}
			return d.DialContext(ctx, network, "1.1.1.1:53")
		},
	}

	dialer := &net.Dialer{
		Resolver: resolver,
		Timeout:  10 * time.Second,
	}

	transport := &http.Transport{
		Proxy:       http.ProxyURL(proxyUrl),
		DialContext: dialer.DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return client, proxy, nil
}

func (pc *ProxyClient) MarkError(p *Proxy) {
	pc.pool.MakeError(p)
}
