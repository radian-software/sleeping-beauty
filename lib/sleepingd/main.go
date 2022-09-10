package sleepingd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/riywo/loginshell"
	"gopkg.in/validator.v2"
)

type Options struct {
	Command        string `validate:"nonzero"`
	TimeoutSeconds int    `validate:"min=1"`
	CommandPort    int    `validate:"min=1"`
	ListenPort     int    `validate:"min=1"`
	ListenHost     string `validate:"nonzero"`
}

func Main(opts *Options) error {
	if err := validator.Validate(opts); err != nil {
		return fmt.Errorf("internal logic error: failed struct validation: %v", err)
	}
	shell, err := loginshell.Shell()
	if err != nil {
		return err
	}
	proc := &SubprocessManager{
		Command:                []string{shell, "-c", opts.Command},
		TerminationGracePeriod: 5 * time.Second,
	}
	dms := (*DeadMansSwitch)(nil)
	dmsLock := sync.Mutex{}
	newConnCallback := func() {
		dmsLock.Lock()
		defer dmsLock.Unlock()
		proc.EnsureStarted()
		if dms == nil || dms.Expired {
			// If dms is nil then it means no timer is
			// running currently. If dms is not nil but
			// dms.Expired is true, then it means that an
			// event is about to be (or already has been)
			// sent to the ExpireCh, but since we acquired
			// the lock, we know that the process has not
			// been stopped from the goroutine below.
			// Therefore, we can set dms.Expired back to
			// false to cancel the process stopping,
			// before setting a new timer.
			if dms.Expired {
				dms.Expired = false
			}
			dms = NewDeadMansSwitch(time.Duration(opts.TimeoutSeconds) * time.Second)
			go func() {
				<-dms.ExpireCh
				dmsLock.Lock()
				defer dmsLock.Unlock()
				// Since ExpireCh is messaged after
				// dms.Expired is set, this condition
				// should normally always be true. It
				// will be canceled in the case that
				// dms.Expired was set back to false
				// by the code in the parent goroutine
				// above.
				if dms.Expired {
					proc.EnsureStopped()
					dms = nil
				}
			}()
		} else {
			dms.DelayCh <- struct{}{}
		}
	}
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:              "tcp",
		ListenAddr:            fmt.Sprintf("%s:%d", opts.ListenHost, opts.ListenPort),
		UpstreamAddr:          fmt.Sprintf("127.0.0.1:%d", opts.CommandPort),
		NewConnectionCallback: newConnCallback,
	})
	if err != nil {
		return err
	}
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)
	interrupt := <-interruptCh
	_ = proxy.Close()
	proc.EnsureStopped()
	os.Exit(128 + int(interrupt.(syscall.Signal)))
	panic("unreachable code")
}
