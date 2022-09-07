package sleepingd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetDeadMansSwitch(t *testing.T) {
	type deadMansSwitchTestPhase struct {
		Name      string
		Duration  time.Duration
		SendDelay bool
	}
	tests := []struct {
		Description        string
		Timeout            time.Duration
		Phases             []deadMansSwitchTestPhase
		ShouldExpireDuring string
	}{
		{
			Description: "should fire if no delays triggered",
			Timeout:     100 * time.Millisecond,
			Phases: []deadMansSwitchTestPhase{
				{
					Name:     "wait",
					Duration: 200 * time.Millisecond,
				},
			},
			ShouldExpireDuring: "wait",
		},
		{
			Description: "should fire later if delays triggered",
			Timeout:     200 * time.Millisecond,
			Phases: []deadMansSwitchTestPhase{
				{
					Name:     "wait 1",
					Duration: 100 * time.Millisecond,
				},
				{
					Name:      "delay 1",
					SendDelay: true,
				},
				{
					Name:     "wait 2",
					Duration: 100 * time.Millisecond,
				},
				{
					Name:      "delay 2",
					SendDelay: true,
				},
				{
					Name:     "wait 3",
					Duration: 150 * time.Millisecond,
				},
				{
					Name:     "wait 4",
					Duration: 150 * time.Millisecond,
				},
			},
			ShouldExpireDuring: "wait 3",
		},
		{
			Description: "should fire if not delayed enough",
			Timeout:     200 * time.Millisecond,
			Phases: []deadMansSwitchTestPhase{
				{
					Name:     "short wait",
					Duration: 100 * time.Millisecond,
				},
				{
					Name:      "delay",
					SendDelay: true,
				},
				{
					Name:     "long wait part 1",
					Duration: 150 * time.Millisecond,
				},
				{
					Name:     "long wait part 2",
					Duration: 150 * time.Millisecond,
				},
			},
			ShouldExpireDuring: "long wait part 2",
		},
		{
			Description: "later delays should replace earlier ones",
			Timeout:     200 * time.Millisecond,
			Phases: []deadMansSwitchTestPhase{
				{
					Name:     "short wait",
					Duration: 100 * time.Millisecond,
				},
				{
					Name:      "delay 1",
					SendDelay: true,
				},
				{
					Name:     "another short wait",
					Duration: 100 * time.Millisecond,
				},
				{
					Name:      "delay 2",
					SendDelay: true,
				},
				{
					Name:      "delay 3",
					SendDelay: true,
				},
				{
					Name:     "long wait part 1",
					Duration: 150 * time.Millisecond,
				},
				{
					Name:     "long wait part 2",
					Duration: 150 * time.Millisecond,
				},
				{
					Name:     "long wait part 3",
					Duration: 150 * time.Millisecond,
				},
			},
			ShouldExpireDuring: "long wait part 2",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.Description, func(t *testing.T) {
			assert.NotZero(t, test.Description, "bad test")
			assert.NotZero(t, test.ShouldExpireDuring, "bad test")
			assert.NotZero(t, test.Timeout, "bad test")
			t.Parallel()
			s := GetDeadMansSwitch(test.Timeout)
			for _, phase := range test.Phases {
				assert.NotZero(t, phase.Name, "bad test")
				if phase.SendDelay {
					s.DelayCh <- struct{}{}
					continue
				}
				if phase.Duration > 0 {
					select {
					case <-time.NewTimer(phase.Duration).C:
						// Proceed to next phase
					case <-s.ExpireCh:
						assert.Equal(t, test.ShouldExpireDuring, phase.Name)
					}
					continue
				}
				assert.Fail(t, "bad test")
			}
			assert.Fail(t, "dead man's switch never fired")
		})
	}
}
