package installers

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//CorrectVersionFromDeb formats the version string to match dpkg info
func CorrectVersionFromDeb(version string) string {
	if len(version) == 0 {
		return ""
	}
	arr := strings.Split(version, "-")
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}

//GetVersionFromBintray grabs the version from bintray
func GetVersionFromBintray(URI string) (string, error) {
	log.Debugln("getRexrayVersionFromBintray ENTER")

	req, err := http.NewRequest("GET", URI, nil)
	if err != nil {
		log.Errorln("HTTP GET Failed:", err)
		log.Debugln("getRexrayVersionFromBintray LEAVE")
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("HTTP Do:", err)
		log.Debugln("getRexrayVersionFromBintray LEAVE")
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("ReadAll Failed:", err)
		log.Debugln("getRexrayVersionFromBintray LEAVE")
		return "", err
	}

	log.Debugln("Body: ", string(body))

	r, err := regexp.Compile(".*>([0-9]+\\.[0-9]+\\.[0-9]+)/.*")
	if err != nil {
		log.Errorln("Rexexp Failed:", err)
		log.Debugln("getRexrayVersionFromBintray LEAVE")
		return "", err
	}

	first := true
	highest := ""
	scanner := bufio.NewScanner(strings.NewReader(string(body)))

	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println("Line:", line)

		strings := r.FindStringSubmatch(line)
		if strings == nil || len(strings) < 2 {
			log.Debugln("Line does contain a version string")
			continue
		}

		version := strings[1]
		log.Debugln("version:", version)

		if first {
			highest = version
			first = false
			log.Debugln("New highest:", highest)
			continue
		}

		fmt.Println("highest:", highest)
		fmt.Println("version:", version)

		if IsVersionStringHigher(highest, version) {
			log.Debugln("New highest:", highest)
			highest = version
		}
	}

	log.Debugln("highest:", highest)
	log.Debugln("getRexrayVersionFromBintray LEAVE")
	return highest, nil
}
