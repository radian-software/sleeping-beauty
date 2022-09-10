package sleepingd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type SubprocessManager struct {
	Command                []string
	TerminationGracePeriod time.Duration
	cmd                    *exec.Cmd
}

func (sm *SubprocessManager) EnsureStopped() error {
	if sm.cmd == nil {
		return nil // already stopped
	}
	fmt.Fprintf(os.Stderr, "sleepingd: stopping subprocess\n")
	_ = sm.cmd.Process.Signal(syscall.SIGTERM)
	waitCh := make(chan error)
	go func() {
		waitCh <- sm.cmd.Wait()
	}()
	select {
	case err := <-waitCh:
		if _, ok := err.(*exec.ExitError); err == nil || ok {
			sm.cmd = nil
			return nil
		}
		return err
	case <-time.NewTimer(sm.TerminationGracePeriod).C:
		_ = sm.cmd.Process.Kill()
	}
	select {
	case err := <-waitCh:
		if _, ok := err.(*exec.ExitError); err == nil || ok {
			sm.cmd = nil
			return nil
		}
		return err
	case <-time.NewTimer(1 * time.Second).C:
		return fmt.Errorf("failed to kill pid %d", sm.cmd.Process.Pid)
	}
}

func (sm *SubprocessManager) EnsureStarted() error {
	if sm.cmd != nil {
		return nil // already started
	}
	fmt.Fprintf(os.Stderr, "sleepingd: starting subprocess\n")
	sm.cmd = exec.Command(sm.Command[0], sm.Command[1:]...)
	sm.cmd.Stdout = os.Stdout
	sm.cmd.Stderr = os.Stderr
	return sm.cmd.Start()
}
