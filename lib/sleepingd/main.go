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
		EnsureListeningTimeout: 5 * time.Second,
	}
	lock := sync.Mutex{}
	dms := NewDeadMansSwitch(time.Duration(opts.TimeoutSeconds) * time.Second)
	go func() {
		for {
			<-dms.ExpireCh
			lock.Lock()
			defer lock.Unlock()
			Must(proc.EnsureStopped())
			Must(proc.EnsureNotListening(opts.CommandPort))
		}
	}()
	activityCallback := func() {
		lock.Lock()
		defer lock.Unlock()
		Must(proc.EnsureStarted())
		Must(proc.EnsureListening(opts.CommandPort))
		dms.Delay()
	}
	proxy, err := NewProxy(&ProxyOptions{
		Protocol:              "tcp",
		ListenAddr:            fmt.Sprintf("%s:%d", opts.ListenHost, opts.ListenPort),
		UpstreamAddr:          fmt.Sprintf("127.0.0.1:%d", opts.CommandPort),
		NewConnectionCallback: activityCallback,
		DataCallback:          dms.Delay,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "sleepingd: listening on %s:%d, proxying to 127.0.0.1:%d with %s command line: %s\n", opts.ListenHost, opts.ListenPort, opts.CommandPort, shell, opts.Command)
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)
	interrupt := <-interruptCh
	LogError(proxy.Close())
	LogError(proc.EnsureStopped())
	os.Exit(128 + int(interrupt.(syscall.Signal)))
	panic("unreachable code")
}
