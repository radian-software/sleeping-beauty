package sleepingd

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
