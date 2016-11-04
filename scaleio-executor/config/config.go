package config

import (
	"errors"
	"flag"

	xplatform "github.com/dvonthenen/goxplatform"
)

var (
	//ErrInvalidRestURI The REST URI provided is not valid
	ErrInvalidRestURI = errors.New("The REST URI provided is not valid")
)

//Config is the representation of the config
type Config struct {
	LogLevel     string
	SchedulerURI string
	MesosAgent   string
	FrameworkID  string
	ExecutorID   string
}

//AddFlags adds flags to the command line parsing
func (cfg *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel,
		"Set the logging level for the application")
	fs.StringVar(&cfg.SchedulerURI, "rest.uri", cfg.SchedulerURI,
		"Scheduler REST API URI")

	fs.StringVar(&cfg.MesosAgent, "mesos.agent", cfg.MesosAgent,
		"Mesos Agent address")
	fs.StringVar(&cfg.FrameworkID, "framework.id", cfg.FrameworkID,
		"Framework ID")
	fs.StringVar(&cfg.ExecutorID, "executor.id", cfg.ExecutorID,
		"Executor ID")
}

//NewConfig creates a new Config object
func NewConfig() *Config {
	return &Config{
		LogLevel:     env("LOG_LEVEL", "info"),
		SchedulerURI: env("SCHEDULER_URI", "127.0.0.1"),
		MesosAgent:   env("MESOS_AGENT_ENDPOINT", "127.0.0.1"),
		FrameworkID:  env("MESOS_FRAMEWORK_ID", ""),
		ExecutorID:   env("MESOS_EXECUTOR_ID", ""),
	}
}

//ParseIPFromRestURI returns the IP from the URI
func (cfg *Config) ParseIPFromRestURI() (string, error) {
	strings, err := xplatform.GetInstance().Str.RegexMatch(cfg.SchedulerURI, ".+//(.*):[0-9]+")
	if err != nil {
		return "", err
	}
	if len(strings) != 2 {
		return "", ErrInvalidRestURI
	}
	ip := strings[1]
	return ip, nil
}
