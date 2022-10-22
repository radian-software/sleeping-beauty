package sleepingd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type SubprocessManager struct {
	Command                []string
	TerminationGracePeriod time.Duration
	EnsureListeningTimeout time.Duration
	cmd                    *exec.Cmd
	listening              bool
}

func (sm *SubprocessManager) EnsureStopped() error {
	if sm.cmd == nil {
		return nil // already stopped
	}
	fmt.Fprintf(os.Stderr, "sleepingd: stopping subprocess\n")
	_ = syscall.Kill(-sm.cmd.Process.Pid, syscall.SIGTERM)
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
		_ = syscall.Kill(-sm.cmd.Process.Pid, syscall.SIGKILL)
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
	sm.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	sm.cmd.Stdout = os.Stdout
	sm.cmd.Stderr = os.Stderr
	return sm.cmd.Start()
}

func (sm *SubprocessManager) EnsureListening(port int) error {
	if sm.listening {
		return nil // already listening
	}
	done := make(chan error)
	go func() {
		for {
			conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err == nil {
				_ = conn.Close()
				done <- nil
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	select {
	case err := <-done:
		sm.listening = true
		return err
	case <-time.NewTimer(sm.EnsureListeningTimeout).C:
		return fmt.Errorf("process did not start listening on port %d", port)
	}
}

func (sm *SubprocessManager) EnsureNotListening(port int) error {
	if !sm.listening {
		return nil // already not listening
	}
	done := make(chan error)
	go func() {
		for {
			_, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
			if err != nil {
				done <- nil
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	select {
	case err := <-done:
		sm.listening = false
		return err
	case <-time.NewTimer(sm.EnsureListeningTimeout).C:
		return fmt.Errorf("process did not stop listening on port %d", port)
	}
}
