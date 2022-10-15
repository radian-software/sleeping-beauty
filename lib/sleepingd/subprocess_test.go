package sleepingd

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SubprocessManager(t *testing.T) {
	sm := &SubprocessManager{
		Command:                []string{"sleep", "86400"},
		TerminationGracePeriod: 100 * time.Millisecond,
	}
	assert.Nil(t, sm.cmd)
	err := sm.EnsureStarted()
	assert.NoError(t, err)
	assert.NotNil(t, sm.cmd)
	proc, err := os.FindProcess(sm.cmd.Process.Pid)
	assert.NoError(t, err)
	err = proc.Signal(syscall.Signal(0))
	assert.NoError(t, err) // assert process is running
	err = sm.EnsureStopped()
	assert.NoError(t, err)
	assert.Nil(t, sm.cmd)
	err = proc.Signal(syscall.Signal(0))
	assert.Error(t, err) // assert process is no longer running
	err = sm.EnsureStarted()
	assert.NoError(t, err)
	assert.NotNil(t, sm.cmd.Process)
	proc, err = os.FindProcess(sm.cmd.Process.Pid)
	assert.NoError(t, err)
	err = proc.Signal(syscall.Signal(0))
	assert.NoError(t, err) // assert process is running again
	// Have to tear down all subprocesses at end of test or Go
	// will wait forever.
	err = sm.EnsureStopped()
	assert.NoError(t, err)
}

func assertPortBound(t *testing.T, port int, shouldBeBound bool) {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if shouldBeBound && err != nil {
		assert.Fail(t, fmt.Sprintf("port %d should be bound but is not: %s", port, err.Error()))
	} else if !shouldBeBound && err == nil {
		assert.Fail(t, fmt.Sprintf("port %d should not be bound but is", port))
	}
	if err == nil {
		err = conn.Close()
		assert.NoError(t, err)
	}
}

func Test_SubprocessManagerListen(t *testing.T) {
	sm := &SubprocessManager{
		Command:                []string{"nc", "-lk", "127.0.0.1", "7000"},
		TerminationGracePeriod: 100 * time.Millisecond,
		EnsureListeningTimeout: 100 * time.Millisecond,
	}
	assertPortBound(t, 7000, false)
	err := sm.EnsureStarted()
	assert.NoError(t, err)
	err = sm.EnsureListening(7000)
	assert.NoError(t, err)
	assertPortBound(t, 7000, true)
	err = sm.EnsureStopped()
	assert.NoError(t, err)
	assertPortBound(t, 7000, false)
}

// https://blog.phusion.nl/2015/01/20/docker-and-the-pid-1-zombie-reaping-problem/
const handleSignalsCorrectly = `trap 'kill $(jobs -p) &>/dev/null' EXIT; `

func Test_SubprocessManagerListenWait(t *testing.T) {
	sm := &SubprocessManager{
		Command:                []string{"bash", "-c", handleSignalsCorrectly + "sleep 0.1 && nc -lk 127.0.0.1 7000"},
		TerminationGracePeriod: 100 * time.Millisecond,
		EnsureListeningTimeout: 200 * time.Millisecond,
	}
	assertPortBound(t, 7000, false)
	err := sm.EnsureStarted()
	assert.NoError(t, err)
	assertPortBound(t, 7000, false) // process is started but not listening yet
	err = sm.EnsureListening(7000)
	assert.NoError(t, err)
	assertPortBound(t, 7000, true)
	err = sm.EnsureStopped()
	assert.NoError(t, err)
	err = sm.EnsureNotListening(7000)
	assert.NoError(t, err)
	assertPortBound(t, 7000, false)
}

func Test_SubprocessManagerListenTimeout(t *testing.T) {
	sm := &SubprocessManager{
		Command:                []string{"bash", "-c", handleSignalsCorrectly + "sleep 0.3 && nc -lk 127.0.0.1 7000"},
		TerminationGracePeriod: 100 * time.Millisecond,
		EnsureListeningTimeout: 200 * time.Millisecond,
	}
	assertPortBound(t, 7000, false)
	err := sm.EnsureStarted()
	assert.NoError(t, err)
	assertPortBound(t, 7000, false) // process is started but not listening yet
	err = sm.EnsureListening(7000)
	assert.Error(t, err) // should time out
	err = sm.EnsureListening(7000)
	assert.NoError(t, err) // second attempt should pass
	assertPortBound(t, 7000, true)
	err = sm.EnsureStopped()
	assert.NoError(t, err)
	err = sm.EnsureNotListening(7000)
	assert.NoError(t, err)
	assertPortBound(t, 7000, false)
}
