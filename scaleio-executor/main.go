package main

import (
	"flag"
	"math/rand"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	config "github.com/codedellemc/scaleio-framework/scaleio-executor/config"
	executor "github.com/codedellemc/scaleio-framework/scaleio-executor/executor"
)

// ----------------------- func init() ------------------------- //

func init() {
	rand.Seed(time.Now().UnixNano())
	log.SetOutput(os.Stdout)
	log.Info("Initializing the ScaleIO Executor...")
}

func main() {
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		log.Debugln(pair[0], "=", pair[1])
	}

	cfg := config.NewConfig()
	fs := flag.NewFlagSet("executor", flag.ExitOnError)
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

	log.Infoln("ExecutorID:", cfg.ExecutorID)
	log.Infoln("FrameworkID:", cfg.FrameworkID)
	log.Infoln("MesosAgent:", cfg.MesosAgent)
	log.Infoln("RestURI:", cfg.SchedulerURI)

	exec := executor.NewScaleIOExecutor(cfg)
	log.Debugln("Starting ScaleIO Executor")
	<-exec.Start()
	log.Info("Executor terminating")
}
