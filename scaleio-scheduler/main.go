package main

import (
	"flag"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler"
)

// ----------------------- func init() ------------------------- //

func init() {
	log.SetOutput(os.Stdout)
	log.Infoln("Initializing the ScaleIO Scheduler...")
}

func prerequisites(cfg *config.Config) bool {
	if !cfg.Experimental && (cfg.PrimaryMdmAddress == "" &&
		cfg.SecondaryMdmAddress == "" && cfg.TieBreakerMdmAddress == "") {
		log.Errorln("Primary, Secondary and TieBreaker MDM nodes must be pre-configured" +
			"in order for additional nodes to be added to the ScaleIO cluster.")
		return false
	}

	if (cfg.PrimaryMdmAddress != "" && cfg.SecondaryMdmAddress != "" &&
		cfg.TieBreakerMdmAddress != "") || (cfg.PrimaryMdmAddress == "" &&
		cfg.SecondaryMdmAddress == "" && cfg.TieBreakerMdmAddress == "") {
		return true
	}

	if cfg.PrimaryMdmAddress != "" {
		log.Errorln("A Pre-Configured Primary MDM IP Address was provided. " +
			"Using this option requires that a Pre-Configured Primary and TieBreaker MDM " +
			"Nodes exist and their IP addresses be provided via command line option.")
	} else if cfg.SecondaryMdmAddress != "" {
		log.Errorln("A Pre-Configured Secondary MDM IP Address was provided. " +
			"Using this option requires that a Pre-Configured Primary and TieBreaker MDM " +
			"Nodes exist and their IP addresses be provided via command line option.")
	} else if cfg.TieBreakerMdmAddress != "" {
		log.Errorln("A Pre-Configured TieBreaker MDM IP Address was provided. " +
			"Using this option requires that a Pre-Configured Primary and Secondary MDM " +
			"Nodes exist and their IP addresses be provided via command line option.")
	}
	return false
}

func main() {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		log.Debugln(pair[0], "=", pair[1])
	}

	cfg := config.NewConfig()
	fs := flag.NewFlagSet("scheduler", flag.ExitOnError)
	cfg.AddFlags(fs)
	fs.Parse(os.Args[1:])

	level, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnln("Invalid log level. Defaulting to info.")
		level = log.InfoLevel
	} else {
		log.Infoln("Set logging to", cfg.LogLevel)
	}
	log.SetLevel(level)

	if !prerequisites(cfg) {
		return //prerequisites not met
	}

	sched := scheduler.NewScaleIOScheduler(cfg)
	log.Debugln("Starting ScaleIO Scheduler")
	<-sched.Start()
	log.Debugln("Scheduler terminating")
}
