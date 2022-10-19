package cases

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func killNicely(t *testing.T, proc *os.Process) {
	assert.NoError(t, proc.Signal(syscall.SIGTERM))
	done := make(chan struct{})
	go func() {
		_, err := proc.Wait()
		assert.NoError(t, err)
		done <- struct{}{}
	}()
	select {
	case <-done:
		return
	case <-time.NewTimer(1 * time.Second).C:
		proc.Kill()
	}
	<-done
}

func Test_Basics(t *testing.T) {
	sb := exec.Command("sleepingd")
	sb.Env = append(
		os.Environ(),
		"SLEEPING_BEAUTY_COMMAND=python3 -u -m http.server -b 127.0.0.1 -d / 6666",
		"SLEEPING_BEAUTY_TIMEOUT_SECONDS=1",
		"SLEEPING_BEAUTY_COMMAND_PORT=6666",
		"SLEEPING_BEAUTY_LISTEN_PORT=4444",
	)
	sbStdout := bytes.Buffer{}
	sb.Stdout = &sbStdout
	sbStderr := bytes.Buffer{}
	sb.Stderr = &sbStderr
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		curl := exec.Command("curl", "-m5", "-sS", "http://127.0.0.1:4444")
		curlStdout := bytes.Buffer{}
		curl.Stdout = &curlStdout
		curlStderr := bytes.Buffer{}
		curl.Stderr = &curlStderr
		assert.NoError(t, curl.Run(), "stderr: %s", curlStderr.String())
		assert.Contains(t, curlStdout.String(), "Directory listing")
		time.Sleep(2 * time.Second)
	}
	numProcStarts := strings.Count(sbStdout.String(), "Serving HTTP")
	assert.Equal(t, 3, numProcStarts)
	numProcRequests := strings.Count(sbStderr.String(), "GET / HTTP/1.1")
	assert.Equal(t, 3, numProcRequests)
	numDaemonStarts := strings.Count(sbStderr.String(), "starting subprocess")
	assert.Equal(t, 3, numDaemonStarts)
	numDaemonStops := strings.Count(sbStderr.String(), "stopping subprocess")
	assert.Equal(t, 3, numDaemonStops)
	assert.Contains(t, sbStderr.String(), "listening on 0.0.0.0:4444, proxying to 127.0.0.1:6666 with /bin/bash command line: python3 -u -m http.server -b 127.0.0.1 -d / 6666")
}

func Test_Keepalive(t *testing.T) {
	// This is a regression test. In previous versions of
	// sleepingd, if the server returned a 'connection:
	// keep-alive' header, then the client would naturally hold
	// its end of the connection open for a few seconds. But this
	// meant that the server connection would be EOF'd before the
	// client closed its side, which triggered a race condition in
	// sleepingd that occurred with that ordering.
	//
	// To ensure that we get a keep-alive response, we use
	// gunicorn with >1 threads.
	sb := exec.Command("sleepingd")
	sb.Env = append(
		os.Environ(),
		"SLEEPING_BEAUTY_COMMAND=python3 -u -m gunicorn --chdir ../resources app:app -b 127.0.0.1:6666 --threads 2",
		"SLEEPING_BEAUTY_TIMEOUT_SECONDS=2",
		"SLEEPING_BEAUTY_COMMAND_PORT=6666",
		"SLEEPING_BEAUTY_LISTEN_PORT=4444",
	)
	sbOutput := bytes.Buffer{}
	sb.Stdout = &sbOutput
	sb.Stderr = &sbOutput
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	curl := exec.Command("curl", "-m5", "-sS", "http://127.0.0.1:4444/about")
	curlStdout := bytes.Buffer{}
	curl.Stdout = &curlStdout
	curlStderr := bytes.Buffer{}
	curl.Stderr = &curlStderr
	assert.NoError(t, curl.Run(), "stderr: %s", curlStderr.String())
	assert.Contains(t, curlStdout.String(), "About this application")
	// Make sure there is no spurious error like this:
	// 'read tcp 127.0.0.1:39068->127.0.0.1:5001: use of closed network connection'
	assert.NotContains(t, sbOutput.String(), "fatal")
	assert.NotContains(t, sbOutput.String(), "error")
	// Make sure that we are actually testing what we want to
	// test, namely by checking for the keep-alive header on the
	// upstream server.
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get("http://127.0.0.1:6666/about")
	assert.NoError(t, err)
	assert.Equal(t, "keep-alive", res.Header.Get("connection"))
}

func Test_ConcurrentRequests(t *testing.T) {
	sb := exec.Command("sleepingd")
	sb.Env = append(
		os.Environ(),
		"SLEEPING_BEAUTY_COMMAND=python3 -u -m gunicorn --chdir ../resources app:app -b 127.0.0.1:6666 --threads 2",
		"SLEEPING_BEAUTY_TIMEOUT_SECONDS=2",
		"SLEEPING_BEAUTY_COMMAND_PORT=6666",
		"SLEEPING_BEAUTY_LISTEN_PORT=4444",
	)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	sbOutput := bytes.Buffer{}
	sb.Stdout = &sbOutput
	sb.Stderr = &sbOutput
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			res, err := client.Get("http://127.0.0.1:4444/about")
			assert.NoError(t, err)
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), "About this application")
			wg.Done()
		}()
	}
	wg.Wait()
	assert.NotContains(t, sbOutput.String(), "fatal")
	assert.NotContains(t, sbOutput.String(), "error")
}

func Test_Load(t *testing.T) {
	sb := exec.Command("sleepingd")
	sb.Env = append(
		os.Environ(),
		"SLEEPING_BEAUTY_COMMAND=python3 -u -m gunicorn --chdir ../resources app:app -b 127.0.0.1:6666 --threads 2",
		"SLEEPING_BEAUTY_TIMEOUT_SECONDS=2",
		"SLEEPING_BEAUTY_COMMAND_PORT=6666",
		"SLEEPING_BEAUTY_LISTEN_PORT=4444",
	)
	sbOutput := bytes.Buffer{}
	sb.Stdout = &sbOutput
	sb.Stderr = &sbOutput
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	k6 := exec.Command("k6", "run", "../resources/loadtest.js")
	k6Output := bytes.Buffer{}
	k6.Stdout = &k6Output
	k6.Stderr = &k6Output
	err := k6.Run()
	assert.NoError(t, err, "output:", k6Output.String())
	assert.NotContains(t, sbOutput.String(), "fatal")
	assert.NotContains(t, sbOutput.String(), "error")
}
