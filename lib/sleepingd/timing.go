package sleepingd

import "time"

// DeadMansSwitch is a struct used for getting automatically notified
// some time after a process stops sending events. See
// GetDeadMansSwitch for more explanation of usage.
type DeadMansSwitch struct {
	delayCh  chan<- struct{}
	ExpireCh <-chan struct{}
}

// NewDeadMansSwitch returns a DeadMansSwitch struct. After getting
// the struct back, invoke Delay to start the timer. Then ExpireCh
// will receive an event automatically after the provided timeout.
// However, at any time you can invoke the Delay method again to
// re-set this timeout to be from the current time rather than from
// when the NewDeadMansSwitch was returned. In other words, this lets
// you get notified automatically some time after a process stops
// sending events. If you invoke Delay again after the ExpireCh
// receives an event, another event will be scheduled for the future.
func NewDeadMansSwitch(timeout time.Duration) *DeadMansSwitch {
	delayCh := make(chan struct{})
	expireCh := make(chan struct{})
	dms := DeadMansSwitch{
		delayCh:  delayCh,
		ExpireCh: expireCh,
	}
	var timer *time.Timer
	go func() {
		<-delayCh
		timer = time.NewTimer(timeout)
		for {
			select {
			case <-delayCh:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
			case <-timer.C:
				expireCh <- struct{}{}
			}
		}
	}()
	return &dms
}

func (dms *DeadMansSwitch) Delay() {
	dms.delayCh <- struct{}{}
}
