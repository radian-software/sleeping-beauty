package sleepingd

import (
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
	// a new connection is accepted and some data has been
	// received from the client, but before the data is proxied to
	// the upstream address. This could be used to track metrics
	// on incoming connections, or to ensure that the upstream is
	// available before traffic is proxied to it.
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
				uc := NewLazyConn(func() (net.Conn, error) {
					if opts.NewConnectionCallback != nil {
						opts.NewConnectionCallback()
					}
					uc, err := net.Dial(opts.Protocol, opts.UpstreamAddr)
					if err != nil {
						LogError(err)
						return nil, err
					}
					return uc, nil
					// openOnRead: false
					// openOnWrite: true
					//
					// Only actually open the
					// connection once client
					// writes to it.
				}, false, true)
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
					// Copy request from client to
					// upstream server. Ignore
					// errors because they may
					// indicate that client
					// disconnected unexpectedly
					// which is not actionable on
					// our end.
					_ = CopyWithActivity(uc, c, activityCh)
				}()
				// Copy response from upstream server
				// to client. Ignore errors, as above.
				_ = CopyWithActivity(c, uc, activityCh)
				// Once the upstream server closes its
				// connection or is unable to send
				// further data, we should proactively
				// close both it and the client
				// connection, to indicate to the
				// sender that more data cannot be
				// sent on this connection. Otherwise
				// smart clients such as web browsers
				// may attempt to reuse it, and hang.
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
