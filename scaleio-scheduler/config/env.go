package config

import (
	"net"
	"os"
	"os/user"
	"strconv"
	"strings"
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

func autoDiscoverIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Warnln("Failed to get Interfaces", err)
		return "", err
	}

	var ip string
	for _, i := range ifaces {
		if strings.Contains(i.Name, "lo") || strings.Contains(i.Name, "docker") {
			log.Debugln("Skipping interface:", i.Name)
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			log.Infoln("Failed to get IPs on Interface", err)
			continue
		}
		// handle err
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP.String()
				log.Debugln("IPNet:", ip)
			case *net.IPAddr:
				ip = v.IP.String()
				log.Debugln("IPAddr:", ip)
			}
			if len(ip) > 0 {
				break
			}
		}

		log.Infoln("IP Discovered:", ip)
		break
	}

	return ip, nil
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
