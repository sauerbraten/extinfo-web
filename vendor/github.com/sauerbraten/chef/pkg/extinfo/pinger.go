package extinfo

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type packet struct {
	addr *net.UDPAddr
	buf  []byte
}

type request struct {
	packet
	resp     chan []byte
	done     chan struct{} // closed by send() caller after it got all expected packets
	deadline time.Time
}

type Pinger struct {
	c *net.UDPConn

	requests  chan request       // from send() callers to UDP socket
	responses chan packet        // from UDP socket to send() callers
	pending   map[string]request // addr -> waiting requests

	cachingUDPResolver
}

func NewPinger(laddr string) (*Pinger, error) {
	addr, err := net.ResolveUDPAddr("udp4", laddr)
	if err != nil {
		return nil, err
	}

	c, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	p := &Pinger{
		c:         c,
		requests:  make(chan request),
		responses: make(chan packet),
		pending:   map[string]request{},

		cachingUDPResolver: newCacheingUDPResolver(1 * time.Hour),
	}

	go p.handleIncoming()
	go p.run()

	return p, nil
}

func (p *Pinger) handleIncoming() {
	inbuf := [1024]byte{}
	for {
		n, raddr, err := p.c.ReadFromUDP(inbuf[:])
		if err != nil {
			panic(err)
		}
		resp := make([]byte, n)
		copy(resp, inbuf[:])
		p.responses <- packet{raddr, resp}
	}
}

func (p *Pinger) run() {
	cleanup := time.NewTicker(500 * time.Millisecond)

	for {
		select {
		case req := <-p.requests: // send() was called

			p.c.SetWriteDeadline(req.deadline)
			_, err := p.c.WriteToUDP(req.buf, req.addr)
			if err != nil {
				// log.Printf("sending to %s failed: %v\n", req.addr, err)
				close(req.resp)
				break
			}
			p.pending[req.addr.String()] = req

		case resp := <-p.responses: // something arrived on the UDP socket

			addr := resp.addr.String()
			req, ok := p.pending[addr]
			if !ok || time.Now().After(req.deadline) {
				// all cleanup happens below
				break
			}

			select {
			case <-time.After(time.Until(req.deadline)):
				// log.Printf("discarding response packet %v from %s for slow receiver\n", resp.buf, addr)
			case <-req.done:
				// log.Printf("discarding extraneous response packet %v from %s (in response to %v)\n", resp.buf, addr, req.buf)
			case req.resp <- resp.buf:
				// keep channel open since response can span multiple packages
			}

		case <-cleanup.C:

			var tbd []string // to be deleted
			for addr, req := range p.pending {
				if time.Now().After(req.deadline) {
					tbd = append(tbd, addr)
				}
			}
			for _, addr := range tbd {
				close(p.pending[addr].resp)
				delete(p.pending, addr)
			}
		}
	}
}

func (p *Pinger) send(host string, port int, buf []byte, timeout time.Duration) (<-chan []byte, func(), error) {
	addr, err := p.resolve(host, port)
	if err != nil {
		return nil, func() {}, fmt.Errorf("resolving %s:%d: %w", host, port+1, err)
	}

	req := request{
		packet:   packet{addr, buf},
		resp:     make(chan []byte, 10),
		done:     make(chan struct{}),
		deadline: time.Now().Add(timeout),
	}

	p.requests <- req

	return req.resp, func() { close(req.done) }, nil
}

func (p *Pinger) expectSinglePacket(host string, port int, req []byte, timeout time.Duration) ([]byte, error) {
	c, done, err := p.send(host, port, req, timeout)
	if err != nil {
		return nil, err
	}
	defer done()

	resp, ok := <-c
	if !ok {
		return nil, fmt.Errorf("receiving response from %s:%d timed out", host, port)
	}
	return resp, nil
}

type cachingUDPResolver struct {
	m          sync.RWMutex
	addrs      map[string]*net.UDPAddr
	staleAfter map[string]time.Time
	lifetime   time.Duration
}

func newCacheingUDPResolver(lifetime time.Duration) cachingUDPResolver {
	return cachingUDPResolver{
		addrs:      map[string]*net.UDPAddr{},
		staleAfter: map[string]time.Time{},
		lifetime:   lifetime,
	}
}

func (c *cachingUDPResolver) resolve(host string, port int) (*net.UDPAddr, error) {
	_addr := fmt.Sprintf("%s:%d", host, port+1)

	c.m.RLock()
	addr, ok := c.addrs[_addr]
	stale := time.Now().After(c.staleAfter[_addr])
	c.m.RUnlock()

	if !ok || stale {
		var err error
		addr, err = net.ResolveUDPAddr("udp4", _addr)
		if err != nil {
			return nil, err
		}

		c.m.Lock()
		defer c.m.Unlock()
		c.addrs[_addr] = addr
		c.staleAfter[_addr] = time.Now().Add(c.lifetime)
	}

	return addr, nil
}
