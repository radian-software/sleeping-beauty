package sleepingd

import (
	"fmt"

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
	fmt.Println("Hello, world!")
	return nil
}
