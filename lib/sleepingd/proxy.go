package sleepingd

import (
	"fmt"
	"net"
)

// ProxyOptions is used to configure NewProxy, which see for
// documentation. All fields are required except
// NewConnectionCallback.
type ProxyOptions struct {
	// Protocol is either "tcp" or "udp"
	Protocol string
	// ListenAddr is the address the proxy server will listen for
	// incoming TCP/UDP traffic on, e.g. "127.0.0.1:80"
	ListenAddr string
	// UpstreamAddr is the upstream address the proxy will proxy
	// TCP/UDP traffic to, e.g. "127.0.0.1:8080"
	UpstreamAddr string
	// NewConnectionCallback is a function of no arguments,
	// optional. If provided, then it is called synchronously when
	// a new connection is accepted, but before the connection is
	// proxied to the upstream address. This could be used to
	// track metrics on incoming connections, or to ensure that
	// the upstream is available before traffic is proxied to it.
	NewConnectionCallback func()
	// DataCallback is a function of no arguments, optional. If
	// provided, then it is called synchronously when data is to
	// be copied either to or from the backend server. This could
	// be used to track metrics on network activity.
	DataCallback func()
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
			go func(c net.Conn) {
				if opts.NewConnectionCallback != nil {
					opts.NewConnectionCallback()
				}
				uc, err := net.Dial(opts.Protocol, opts.UpstreamAddr)
				if err != nil {
					_, _ = c.Write([]byte(fmt.Sprintf("failed to dial upstream %s: %s\n", opts.UpstreamAddr, err)))
					return
				}
				activityCh := make(chan struct{})
				go func() {
					for {
						<-activityCh
						if opts.DataCallback != nil {
							opts.DataCallback()
						}
					}
				}()
				go func() {
					_ = CopyWithActivity(c, uc, activityCh)
				}()
				_ = CopyWithActivity(uc, c, activityCh)
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
