package sleepingd

import (
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
