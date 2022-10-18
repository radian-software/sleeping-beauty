package cases

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
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
		curl := exec.Command("curl", "-sS", "http://127.0.0.1:4444")
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
	assert.Contains(t, sbStderr.String(), "listening on 0.0.0.0:4444, proxying to 127.0.0.1:6666 with /usr/bin/bash command line: python3 -u -m http.server -b 127.0.0.1 -d / 6666")
}
