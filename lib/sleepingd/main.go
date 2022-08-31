package sleepingd

import (
	"fmt"
	"io"
	"net"

	"gopkg.in/validator.v2"
)

type Options struct {
	Command         string `validate:"nonzero"`
	TimeoutSeconds  int    `validate:"min=1"`
	CommandPort     int    `validate:"min=1"`
	ListenPort      int    `validate:"min=1"`
	ListenHost      string `validate:"nonzero"`
	HealthcheckPath string
}

func Main(opts *Options) error {
	if err := validator.Validate(opts); err != nil {
		return fmt.Errorf("internal logic error: failed struct validation: %v", err)
	}
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", opts.ListenHost, opts.ListenPort))
	if err != nil {
		return err
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go func(c net.Conn) {
			_, _ = io.Copy(c, c)
			_ = c.Close()
		}(conn)
	}
}
