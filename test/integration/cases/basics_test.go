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
		"SLEEPING_BEAUTY_COMMAND=python3 -m http.server -b 127.0.0.1 -d / 6666",
		"SLEEPING_BEAUTY_TIMEOUT_SECONDS=1",
		"SLEEPING_BEAUTY_COMMAND_PORT=6666",
		"SLEEPING_BEAUTY_LISTEN_PORT=4444",
	)
	sbStdout := bytes.Buffer{}
	sb.Stdout = &sbStdout
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		curl := exec.Command("curl", "-sS", "http://localhost:4444")
		curlStdout := bytes.Buffer{}
		curl.Stdout = &curlStdout
		curlStderr := bytes.Buffer{}
		curl.Stderr = &curlStderr
		assert.NoError(t, curl.Run(), "stderr: %s", curlStderr.String())
		assert.Contains(t, curlStdout.String(), "Directory listing")
		time.Sleep(2 * time.Second)
	}
	killNicely(t, sb.Process)
	numStarts := strings.Count(sbStdout.String(), "Serving HTTP")
	assert.Equal(t, 3, numStarts)
}
