package config

import "flag"

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
