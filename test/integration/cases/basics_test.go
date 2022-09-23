package cases

import (
	"io"
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
	sbPipe, err := sb.StdoutPipe()
	assert.NoError(t, err)
	assert.NoError(t, sb.Start())
	defer killNicely(t, sb.Process)
	time.Sleep(500 * time.Millisecond)
	for i := 0; i < 3; i++ {
		curl := exec.Command("curl", "http://localhost:4444")
		curlPipe, err := curl.StdoutPipe()
		assert.NoError(t, err)
		assert.NoError(t, curl.Run())
		curlStdout, err := io.ReadAll(curlPipe)
		assert.NoError(t, err)
		assert.Contains(t, string(curlStdout), "Directory listing")
		time.Sleep(2 * time.Second)
	}
	sbStdout, err := io.ReadAll(sbPipe)
	assert.NoError(t, err)
	numStarts := strings.Count(string(sbStdout), "Serving HTTP")
	assert.Equal(t, 3, numStarts)
}
