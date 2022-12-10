package sleepingd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_LazyConn_Basics(t *testing.T) {
	type lazyConnTestPhase struct {
		Name             string
		Duration         time.Duration
		TriggerRead      bool
		TriggerWrite     bool
		TriggerClose     bool
		ExpectInit       bool
		ExpectRead       bool
		ExpectWrite      bool
		ExpectReadError  bool
		ExpectWriteError bool
	}
	tests := []struct {
		Description string
		OpenOnRead  bool
		OpenOnWrite bool
		Phases      []lazyConnTestPhase
	}{
		{
			Description: "unusable lazy channel that blocks forever",
			OpenOnRead:  false,
			OpenOnWrite: false,
			Phases: []lazyConnTestPhase{
				{
					Name:        "read should block",
					Duration:    100 * time.Millisecond,
					TriggerRead: true,
				},
				{
					Name:         "write should block",
					Duration:     100 * time.Millisecond,
					TriggerWrite: true,
				},
				{
					Name:             "close should abort read and write",
					Duration:         100 * time.Millisecond,
					TriggerClose:     true,
					ExpectReadError:  true,
					ExpectWriteError: true,
				},
				{
					Name:            "read should error immediately now",
					Duration:        100 * time.Millisecond,
					TriggerRead:     true,
					ExpectReadError: true,
				},
				{
					Name:             "write should error immediately now",
					Duration:         100 * time.Millisecond,
					TriggerWrite:     true,
					ExpectWriteError: true,
				},
			},
		},
		{
			Description: "open-on-write channel that we write to first",
			OpenOnRead:  false,
			OpenOnWrite: true,
			Phases: []lazyConnTestPhase{
				{
					Name:         "write should work",
					Duration:     100 * time.Millisecond,
					TriggerWrite: true,
					ExpectInit:   true,
					ExpectWrite:  true,
				},
				{
					Name:        "read should work afterwards",
					Duration:    100 * time.Millisecond,
					TriggerRead: true,
					ExpectRead:  true,
				},
				{
					Name:         "close should not cause any errors",
					Duration:     100 * time.Millisecond,
					TriggerClose: true,
				},
				{
					Name:            "read should error immediately now",
					Duration:        100 * time.Millisecond,
					TriggerRead:     true,
					ExpectReadError: true,
				},
				{
					Name:             "write should error immediately now",
					Duration:         100 * time.Millisecond,
					TriggerWrite:     true,
					ExpectWriteError: true,
				},
			},
		},
		{
			Description: "open-on-write channel that we read from first",
			OpenOnRead:  false,
			OpenOnWrite: true,
			Phases: []lazyConnTestPhase{
				{
					Name:        "read should block when done first",
					Duration:    100 * time.Millisecond,
					TriggerRead: true,
				},
				{
					Name:         "write should work and unblock read",
					Duration:     100 * time.Millisecond,
					TriggerWrite: true,
					ExpectInit:   true,
					ExpectRead:   true,
					ExpectWrite:  true,
				},
				{
					Name:         "close should not cause any errors",
					Duration:     100 * time.Millisecond,
					TriggerClose: true,
				},
				{
					Name:            "read should error immediately now",
					Duration:        100 * time.Millisecond,
					TriggerRead:     true,
					ExpectReadError: true,
				},
				{
					Name:             "write should error immediately now",
					Duration:         100 * time.Millisecond,
					TriggerWrite:     true,
					ExpectWriteError: true,
				},
			},
		},
		{
			Description: "open-on-write channel that we close before writing",
			OpenOnRead:  false,
			OpenOnWrite: true,
			Phases: []lazyConnTestPhase{
				{
					Name:        "read should block when done first",
					Duration:    100 * time.Millisecond,
					TriggerRead: true,
				},
				{
					Name:            "closing channel should error pending read",
					Duration:        100 * time.Millisecond,
					TriggerClose:    true,
					ExpectReadError: true,
				},
				{
					Name:            "read should error immediately now",
					Duration:        100 * time.Millisecond,
					TriggerRead:     true,
					ExpectReadError: true,
				},
				{
					Name:             "write should error immediately now",
					Duration:         100 * time.Millisecond,
					TriggerWrite:     true,
					ExpectWriteError: true,
				},
			},
		},
	}
	for _, tt := range tests {
		require.NotZero(t, tt.Description, "bad test")
		// These bools control whether we set up mock
		// expectations for these methods on the underlying
		// mock connection object (*not* the lazy connection).
		willExpectInit := false
		willExpectRead := false
		willExpectWrite := false
		for _, phase := range tt.Phases {
			if phase.ExpectInit {
				willExpectInit = true
			}
			if phase.ExpectRead {
				willExpectRead = true
			}
			if phase.ExpectWrite {
				willExpectWrite = true
			}
		}
		t.Logf("start test %s", tt.Description)
		initChan := make(chan struct{})
		readChan := make(chan struct{})
		writeChan := make(chan struct{})
		readErrChan := make(chan struct{})
		writeErrChan := make(chan struct{})
		conn := newMockConn(t)
		if willExpectRead {
			conn.ReturnFromNextRead([]byte("this was read\n"), readChan)
		}
		if willExpectWrite {
			conn.ExpectInNextWrite([]byte("this was written\n"), writeChan)
		}
		// Lazy connection will always be closed, but
		// underlying connection will not be closed if it was
		// never initialized in the first place.
		if willExpectInit {
			conn.On("Close").Return(nil).Once()
		}
		lazy := NewLazyConn(func() (SimpleConn, error) {
			initChan <- struct{}{}
			return conn, nil
		}, tt.OpenOnRead, tt.OpenOnWrite)
		for _, phase := range tt.Phases {
			require.NotZero(t, phase.Name, "bad test")
			require.NotZero(t, phase.Duration, "bad test")
			require.False(t, phase.TriggerRead && phase.TriggerWrite, "can't read and write at same time")
			require.False(t, phase.TriggerRead && phase.TriggerClose, "can't read and close at same time")
			require.False(t, phase.TriggerWrite && phase.TriggerClose, "can't write and close at same time")
			t.Logf("start test %s phase %s", tt.Description, phase.Name)
			if phase.TriggerRead {
				go func() {
					buf := make([]byte, 1024)
					nr, err := lazy.Read(buf)
					if err != nil {
						readErrChan <- struct{}{}
						assert.Zero(t, nr)
					} else {
						assert.Equal(t, "this was read\n", string(buf[:nr]))
					}
				}()
			}
			if phase.TriggerWrite {
				go func() {
					buf := []byte("this was written\n")
					nw, err := lazy.Write(buf)
					if err != nil {
						writeErrChan <- struct{}{}
						assert.Zero(t, nw)
					} else {
						assert.Equal(t, len(buf), nw)
					}
				}()
			}
			if phase.TriggerClose {
				go func() {
					err := lazy.Close()
					assert.NoError(t, err)
				}()
			}
			wasInit := false
			wasRead := false
			wasWrite := false
			wasReadError := false
			wasWriteError := false
			timer := time.NewTimer(phase.Duration)
		loop:
			for {
				select {
				case <-timer.C:
					// Proceed to next phase
					break loop
				case <-initChan:
					wasInit = true
				case <-readChan:
					wasRead = true
				case <-writeChan:
					wasWrite = true
				case <-readErrChan:
					wasReadError = true
				case <-writeErrChan:
					wasWriteError = true
				}
			}
			assert.Equal(t, phase.ExpectInit, wasInit)
			assert.Equal(t, phase.ExpectRead, wasRead)
			assert.Equal(t, phase.ExpectWrite, wasWrite)
			assert.Equal(t, phase.ExpectReadError, wasReadError)
			assert.Equal(t, phase.ExpectWriteError, wasWriteError)
		}
		conn.AssertExpectations(t)
	}
}

func Test_CopyWithActivity(t *testing.T) {
	activityCh := make(chan struct{})
	activityCount := 0
	go func() {
		for {
			<-activityCh
			activityCount += 1
		}
	}()
	src := bytes.NewReader([]byte("Hello, world!"))
	dst := &bytes.Buffer{}
	err := CopyWithActivity(dst, src, activityCh)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, world!", dst.String())
	time.Sleep(250 * time.Millisecond)
	assert.Greater(t, activityCount, 0)
}
