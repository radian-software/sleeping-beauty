package sleepingd

import (
	"fmt"
	"os"
)

func Log(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "sleepingd: "+format+"\n", args...)
}

func LogError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "sleepingd: error: %s\n", err.Error())
	}
}

func Must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "sleepingd: fatal: %s\n", err.Error())
		os.Exit(1)
	}
}
