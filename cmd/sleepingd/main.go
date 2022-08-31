package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/radian-software/sleeping-beauty/lib/sleepingd"
)

type envConfig struct {
	Command         string `env:"SLEEPING_BEAUTY_COMMAND,notEmpty"`
	TimeoutSeconds  int    `env:"SLEEPING_BEAUTY_TIMEOUT_SECONDS,required"`
	CommandPort     int    `env:"SLEEPING_BEAUTY_COMMAND_PORT,required"`
	ListenPort      int    `env:"SLEEPING_BEAUTY_LISTEN_PORT,required"`
	ListenHost      string `env:"SLEEPING_BEAUTY_LISTEN_HOST,notEmpty" envDefault:"0.0.0.0"`
	HealthcheckPath string `env:"SLEEPING_BEAUTY_HEALTHCHECK_PATH"`
}

func mainE() error {
	envCfg := envConfig{}
	if err := env.Parse(&envCfg); err != nil {
		return err
	}
	if envCfg.TimeoutSeconds <= 0 {
		return fmt.Errorf("invalid timeout: %d", envCfg.TimeoutSeconds)
	}
	if envCfg.CommandPort <= 0 {
		return fmt.Errorf("invalid port: %d", envCfg.CommandPort)
	}
	if envCfg.ListenPort <= 0 {
		return fmt.Errorf("invalid port: %d", envCfg.ListenPort)
	}
	return sleepingd.Main(&sleepingd.Options{
		Command:         envCfg.Command,
		TimeoutSeconds:  envCfg.TimeoutSeconds,
		CommandPort:     envCfg.CommandPort,
		ListenPort:      envCfg.ListenPort,
		ListenHost:      envCfg.ListenHost,
		HealthcheckPath: envCfg.HealthcheckPath,
	})
}

func main() {
	if err := mainE(); err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
