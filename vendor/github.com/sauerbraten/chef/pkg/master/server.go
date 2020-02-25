package master

import (
	"bufio"
	"net"
	"strings"
	"time"
)

type ServerList map[string]*net.UDPAddr

type Server struct {
	addr    string
	timeout time.Duration
	cache   ServerList // to re-use resolved UDP addresses
}

func New(addr string, timeout time.Duration) *Server {
	return &Server{
		addr:    addr,
		timeout: timeout,
		cache:   ServerList{},
	}
}

func (s *Server) ServerList() (servers ServerList, err error) {
	conn, err := net.DialTimeout("tcp", s.addr, s.timeout)
	if err != nil {
		return
	}
	defer conn.Close()

	in := bufio.NewScanner(conn)
	out := bufio.NewWriter(conn)

	// request list

	_, err = out.WriteString("list\n")
	if err != nil {
		return
	}

	err = out.Flush()
	if err != nil {
		return
	}

	// receive list

	servers = ServerList{}

	for in.Scan() {
		msg := in.Text()
		if msg == "\x00" {
			// end of list
			continue
		}

		msg = strings.TrimPrefix(msg, "addserver ")
		msg = strings.TrimSpace(msg)

		// 12.23.34.45 28785 → 12.23.34.45:28785
		msg = strings.Replace(msg, " ", ":", -1)

		addr, ok := s.cache[msg]
		if !ok {
			addr, err = net.ResolveUDPAddr("udp", msg)
			if err != nil {
				return
			}
			s.cache[msg] = addr // cache resolved address
		}

		servers[addr.String()] = addr
	}

	err = in.Err()

	return
}
