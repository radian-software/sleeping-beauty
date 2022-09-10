package sleepingd

import (
	"fmt"
	"os"
)

func Log(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "sleepingd: "+format, args...)
}

func LogError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "sleepingd: error: %s", err.Error())
	}
}

func Must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "sleepingd: fatal: %s", err.Error())
		os.Exit(1)
	}
}
