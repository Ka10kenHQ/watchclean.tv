package proxy

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Proxy struct {
	Addr  string
	LastError time.Time
}

type Pool struct {
	proxies []*Proxy
	mu      sync.Mutex
	rand    *rand.Rand
}

func NewPool(proxyAddrs []string) *Pool {
	proxies := make([]*Proxy, len(proxyAddrs))
	for i, p := range proxyAddrs {
		proxies[i] = &Proxy { Addr: p}
	}

	return &Pool {
		proxies: proxies,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *Pool) GetRandom() (*Proxy, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.proxies) == 0 {
		return nil, errors.New("no proxies available")
	}

	idx := p.rand.Intn(len(p.proxies))
	return p.proxies[idx], nil
}

func (p *Pool) MakeError(proxy *Proxy) {
	p.mu.Lock()
	defer p.mu.Unlock()

	proxy.LastError = time.Now()
}
