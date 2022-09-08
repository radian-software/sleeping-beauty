package sleepingd

import (
	"fmt"
	"io"
	"net"
)

// ProxyOptions is used to configure NewProxy, which see for
// documentation. All fields are required except NewConnectionCh.
type ProxyOptions struct {
	// Protocol is either "tcp" or "udp"
	Protocol string
	// ListenAddr is the address the proxy server will listen for
	// incoming TCP/UDP traffic on, e.g. "127.0.0.1:80"
	ListenAddr string
	// UpstreamAddr is the upstream address the proxy will proxy
	// TCP/UDP traffic to, e.g. "127.0.0.1:8080"
	UpstreamAddr string
	// NewConnectionCh is an optional channel. If it is provided
	// then the channel will be sent an event every time there is
	// a new incoming TCP/UDP connection, it is the caller's
	// responsibility to make sure that events are read from this
	// channel in a timely manner
	NewConnectionCh *chan<- struct{}
}

// Proxy is a struct returned by NewProxy, that represents a running
// proxy server. It can be used to stop the server by calling Close.
type Proxy struct {
	listener net.Listener
}

// NewProxy creates and starts a TCP or UDP server that will
// transparently proxy TCP/UDP traffic to an upstream address, see
// ProxyOptions for the options. An instance of Proxy is returned
// which can be used to stop the server later.
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
				go func() {
					_, _ = io.Copy(uc, c)
				}()
				_, _ = io.Copy(c, uc)
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
