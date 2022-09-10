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
	data, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Equal(t, "failed to dial upstream 127.0.0.1:7000: dial tcp 127.0.0.1:7000: connect: connection refused\n", string(data))
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

func Test_Proxy_NewConnectionChannel(t *testing.T) {
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
}
