package sleepingd

import (
	"fmt"
	"io"
	"net"
)

type ProxyOptions struct {
	Protocol        string
	ListenAddr      string
	UpstreamAddr    string
	NewConnectionCh *chan<- struct{}
}

type Proxy struct {
	listener net.Listener
}

func NewProxy(opts *ProxyOptions) (*Proxy, error) {
	l, err := net.Listen(opts.Protocol, opts.ListenAddr)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			if opts.NewConnectionCh != nil {
				*opts.NewConnectionCh <- struct{}{}
			}
			go func(c net.Conn) {
				uc, err := net.Dial(opts.Protocol, opts.UpstreamAddr)
				if err != nil {
					_, _ = c.Write([]byte(fmt.Sprintf("failed to dial upstream %s: %s\n", opts.UpstreamAddr, err)))
					return
				}
				_, err = io.Copy(uc, c)
				if err != nil {
					_, _ = c.Write([]byte(fmt.Sprintf("failed to write to upstream %s: %s\n", opts.UpstreamAddr, err)))
				}
				written, err := io.Copy(c, uc)
				if err != nil && written == 0 {
					// If partial response, don't
					// write error message to stream
					_, _ = c.Write([]byte(fmt.Sprintf("got no response from upstream %s: %s\n", opts.UpstreamAddr, err)))
				}
				_ = uc.Close()
				_ = c.Close()
			}(conn)
		}
	}()
	return &Proxy{
		listener: l,
	}, nil
}

func (p *Proxy) Close() error {
	return p.listener.Close()
}
