package sleepingd

import (
	"sync"
	"time"
)

// DeadMansSwitch is a struct used for getting automatically notified
// some time after a process stops sending events. See
// GetDeadMansSwitch for more explanation of usage.
type DeadMansSwitch struct {
	timeout   time.Duration
	precision time.Duration
	callback  func()

	lock     *sync.Mutex
	lastPing time.Time
	active   bool
}

// NewDeadMansSwitch returns a DeadMansSwitch struct. After getting
// the struct back, invoke Ping to start the timer. Then the provided
// callback will be invoked (on a separate goroutine) after the
// provided timeout. However, at any time you can invoke the Ping
// method again to re-set this timeout to be from the current time
// rather than from when the NewDeadMansSwitch was returned. In other
// words, this lets you get notified automatically some time after a
// process stops sending events. If you invoke Ping again after the
// callback is already invoked, another event will be scheduled for
// the future. The callback is not necessarily invoked at the exact
// specified timeout, but can be at most precision later.
func NewDeadMansSwitch(timeout time.Duration, precision time.Duration, callback func()) *DeadMansSwitch {
	return &DeadMansSwitch{
		timeout:   timeout,
		precision: precision,
		callback:  callback,
		lock:      &sync.Mutex{},
		// value of lastPing is ignored when active is false
		lastPing: time.Now(),
		active:   false,
	}
}

// Reschedule the invocation of the DeadMansSwitch callback to happen
// later, or schedule it to happen again in the future.
func (dms *DeadMansSwitch) Ping() {
	dms.lock.Lock()
	dms.lastPing = time.Now()
	if !dms.active {
		time.AfterFunc(dms.precision, dms.check)
		dms.active = true
	}
	dms.lock.Unlock()
}

func (dms *DeadMansSwitch) check() {
	dms.lock.Lock()
	if dms.active && time.Now().Sub(dms.lastPing) >= dms.timeout {
		go dms.callback()
		dms.active = false
	} else {
		time.AfterFunc(dms.precision, dms.check)
	}
	dms.lock.Unlock()
}
