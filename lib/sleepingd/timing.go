package sleepingd

import "time"

// DeadMansSwitch is a struct used for getting automatically notified
// some time after a process stops sending events. See
// GetDeadMansSwitch for more explanation of usage.
type DeadMansSwitch struct {
	DelayCh  chan<- struct{}
	ExpireCh <-chan struct{}
	Expired  bool
}

// NewDeadMansSwitch returns a DeadMansSwitch struct. This struct
// contains two channels. The ExpireCh will receive an event
// automatically after the provided timeout. However, at any time you
// can send an event to the DelayCh to re-set this timeout to be from
// the current time rather than from when the NewDeadMansSwitch was
// returned. In other words, this lets you get notified automatically
// some time after a process stops sending events. To help in avoiding
// race conditions, the field Expired is set from false to true
// immediately *before* an event is sent to the ExpireCh.
func NewDeadMansSwitch(timeout time.Duration) *DeadMansSwitch {
	delayCh := make(chan struct{})
	expireCh := make(chan struct{})
	timer := time.NewTimer(timeout)
	dms := DeadMansSwitch{
		DelayCh:  delayCh,
		ExpireCh: expireCh,
		Expired:  false,
	}
	go func() {
		for {
			select {
			case <-delayCh:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(timeout)
				continue
			case <-timer.C:
				dms.Expired = true
				expireCh <- struct{}{}
				break
			}
		}
	}()
	return &dms
}
