package sleepingd

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Proxy(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!\n"))
	})
	server := http.Server{
		Addr:    "127.0.0.1:7000",
		Handler: mux,
	}
	go assert.NoError(t, server.ListenAndServe())
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
