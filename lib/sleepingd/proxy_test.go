package sleepingd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Proxy_HTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!\n"))
	})
	server := http.Server{
		Addr:    "127.0.0.1:7000",
		Handler: mux,
	}
	go server.ListenAndServe()
	defer server.Close()
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:     "tcp",
		ListenAddr:   "127.0.0.1:7001",
		UpstreamAddr: "127.0.0.1:7000",
	})
	assert.NoError(t, err)
	defer proxy.Close()
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get("http://127.0.0.1:7001")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!\n", string(body))
	_ = res.Body.Close()
	// Nothing should be running anymore
	time.Sleep(100 * time.Millisecond)
	assert.Zero(t, globalCopyCounter)
}

func Test_Proxy_NoUpstream(t *testing.T) {
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:     "tcp",
		ListenAddr:   "127.0.0.1:7001",
		UpstreamAddr: "127.0.0.1:7000",
	})
	assert.NoError(t, err)
	defer proxy.Close()
	conn, err := net.Dial("tcp", "127.0.0.1:7001")
	assert.NoError(t, err)
	// Note: this write will most likely succeed, since we read
	// the first few bytes into an in-memory buffer before lazily
	// trying to open a connection to the upstream. That is okay,
	// we don't attempt to test the results, but it is important
	// that we actually do attempt to write, otherwise the
	// connection will never close.
	_, _ = conn.Write([]byte("is anybody there?\n"))
	data, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Empty(t, data)
	// Nothing should be running anymore
	time.Sleep(100 * time.Millisecond)
	assert.Zero(t, globalCopyCounter)
}

func getEchoserver(t *testing.T, protocol string, addr string) net.Listener {
	l, err := net.Listen(protocol, addr)
	assert.NoError(t, err)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			go io.Copy(conn, conn)
		}
	}()
	return l
}

func Test_Proxy_NewConnectionCallback(t *testing.T) {
	echoserver := getEchoserver(t, "tcp", "127.0.0.1:7000")
	defer echoserver.Close()
	numConns := 0
	numConnsLock := &sync.Mutex{}
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:     "tcp",
		ListenAddr:   "127.0.0.1:7001",
		UpstreamAddr: "127.0.0.1:7000",
		NewConnectionCallback: func() {
			numConnsLock.Lock()
			defer numConnsLock.Unlock()
			numConns += 1
		},
	})
	assert.NoError(t, err)
	defer proxy.Close()
	for i := 0; i < 5; i++ {
		i := i
		go func() {
			message := fmt.Sprintf("message %d", i)
			conn, err := net.Dial("tcp", "127.0.0.1:7001")
			assert.NoError(t, err)
			conn.Write([]byte(message))
			data, err := io.ReadAll(conn)
			assert.NoError(t, err)
			assert.Equal(t, message, string(data))
		}()
	}
	time.Sleep(250 * time.Millisecond)
	assert.Equal(t, 5, numConns)
	// Nothing should be running anymore
	time.Sleep(100 * time.Millisecond)
	assert.Zero(t, globalCopyCounter)
}

func getDelayedCloser(t *testing.T, protocol string, addr string, delay time.Duration) net.Listener {
	l, err := net.Listen(protocol, addr)
	assert.NoError(t, err)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				c.Write([]byte("hi\n"))
				time.Sleep(delay)
				c.Write([]byte("bye\n"))
				err = c.Close()
				assert.NoError(t, err)
			}(conn)
		}
	}()
	return l
}

func Test_Proxy_UpstreamClose(t *testing.T) {
	// This test ensures that when the upstream server closes its
	// connection, that closure is propagated to clients.
	// Otherwise we get a bug where smart clients like web
	// browsers will not realize that the connection is closed
	// upstream, and will keep trying to send traffic to the
	// still-open connection, to no avail.
	closer := getDelayedCloser(t, "tcp", "127.0.0.1:7000", 200*time.Millisecond)
	defer closer.Close()
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:     "tcp",
		ListenAddr:   "127.0.0.1:7001",
		UpstreamAddr: "127.0.0.1:7000",
	})
	assert.NoError(t, err)
	defer proxy.Close()
	conn, err := net.Dial("tcp", "127.0.0.1:7001")
	assert.NoError(t, err)
	_, _ = conn.Write([]byte("is anybody there?\n"))
	closed := make(chan error)
	go func() {
		buf := make([]byte, 1024)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				assert.Equal(t, io.EOF, err)
				closed <- err
				return
			}
		}
	}()
	select {
	case <-closed:
		assert.Fail(t, "channel closed ahead of time")
	case <-time.NewTimer(100 * time.Millisecond).C:
		// proceed
	}
	select {
	case <-time.NewTimer(200 * time.Millisecond).C:
		assert.Fail(t, "channel not closed soon enough")
	case <-closed:
		// proceed
	}
	// Nothing should be running anymore
	time.Sleep(100 * time.Millisecond)
	assert.Zero(t, globalCopyCounter)
}

// Listen for connections on the given protocol and addr. Accept them,
// but if any data is sent on an inbound connection, fail the test.
func getBombServer(t *testing.T, protocol string, addr string) net.Listener {
	l, err := net.Listen(protocol, addr)
	assert.NoError(t, err)
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				nr, _ := c.Read(make([]byte, 1))
				assert.Zero(t, nr, "bomb server received some data")
			}(conn)
		}
	}()
	return l
}

// This is a regression test, to make sure that CopyWithActivity
// sessions don't pile up when clients open and close TCP connections
// without sending any data. It also covers making sure no traffic is
// actually proxied to the upstream if data is not sent.
func Test_Proxy_MemoryLeak(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!\n"))
	})
	server := http.Server{
		Addr:    "127.0.0.1:7000",
		Handler: mux,
	}
	go server.ListenAndServe()
	defer server.Close()
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:     "tcp",
		ListenAddr:   "127.0.0.1:7001",
		UpstreamAddr: "127.0.0.1:7000",
	})
	assert.NoError(t, err)
	defer proxy.Close()
	for i := 0; i < 5; i++ {
		go func() {
			// Open a TCP connection but then close it
			// without sending any data.
			conn, err := net.Dial("tcp", "127.0.0.1:7001")
			assert.NoError(t, err)
			err = conn.Close()
			assert.NoError(t, err)
		}()
	}
	// Nothing should be running anymore - this is the part of the
	// test that will fail if there is a memory leak caused by the
	// bug this regression test is intended to catch
	time.Sleep(100 * time.Millisecond)
	assert.Zero(t, globalCopyCounter)
}
