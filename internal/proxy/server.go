package proxy

import (
	"context"
	"io"
	"log"
	"maps"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Server struct {
	Addr string
    Pool   []string // list of upstream proxies
}

func (s *Server) selectProxy() string {
    if len(s.Pool) == 0 {
        return "" // no upstream proxy
    }
    return s.Pool[rand.Intn(len(s.Pool))]
}

func NewServer(addr string) *Server {
	return &Server { Addr: addr }
}

func (s *Server) Start() error {
	server := &http.Server {
		Addr: s.Addr,
		Handler: http.HandlerFunc(s.handle),
		ReadTimeout: 10 * time.Second,
	}

	log.Println("Proxy listenning on", s.Addr);
	return server.ListenAndServe()
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		s.tunnelHTTPS(w, r);
		return
	}
	s.forwardHTTP(w, r)
}

func (s *Server) tunnelHTTPS(w http.ResponseWriter, r *http.Request) {
	dest, err := net.DialTimeout("tcp", r.Host, 10*time.Second);
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	hj, _ := w.(http.Hijacker)
	client, _, _ := hj.Hijack()

	client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	go io.Copy(dest, client)
	go io.Copy(client, dest)
}

func (s *Server) forwardHTTP(w http.ResponseWriter, r *http.Request) {
    proxyURL := s.selectProxy()

	resolver := &net.Resolver{
        PreferGo: true,
        Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
            d := net.Dialer{Timeout: 2 * time.Second}
            return d.DialContext(ctx, network, "1.1.1.1:53") // Cloudflare DNS
        },
    }

	dialer := &net.Dialer{
		Resolver: resolver,
		Timeout:  10 * time.Second,
	}

	transport := &http.Transport{
		DialContext: dialer.DialContext,
	}

    if proxyURL != "" {
        url, _ := url.Parse(proxyURL)
        transport.Proxy = http.ProxyURL(url)
    }

	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	maps.Copy(dst, src)
}
