package config

import (
	"os"
	"os/user"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
)

func env(key, defaultValue string) (value string) {
	if value = os.Getenv(key); value == "" {
		value = defaultValue
	}
	return
}

func envInt(key, defaultValue string) int {
	value, err := strconv.Atoi(env(key, defaultValue))
	if err != nil {
		panic(err.Error())
	}
	return value
}

func envBool(key, defaultValue string) bool {
	return env(key, defaultValue) == "true"
}

func envDuration(key, defaultValue string) time.Duration {
	value, err := time.ParseDuration(env(key, defaultValue))
	if err != nil {
		panic(err.Error())
	}
	return value
}

func envFloat(key, defaultValue string) float64 {
	value, err := strconv.ParseFloat(env(key, defaultValue), 64)
	if err != nil {
		panic(err.Error())
	}
	return value
}

func mesosUser() string {
	u, err := user.Current()
	if err != nil {
		log.Warnln("Unable to determine user")
		return "root"
	}

	log.Debugln("User:", u.Username)
	return u.Username
}

func mesosHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorln("Unable to determine Hostname")
		return "UNKNOWN"
	}
	log.Debugln("Hostname:", hostname)
	return hostname
}
