package installers

import (
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"

	util "github.com/codedellemc/scaleio-framework/scaleio-executor/native"
)

var (
	//ErrParseVersionFailed failed to parse version from filename
	ErrParseVersionFailed = errors.New("Failed to parse version from filename")
)

//DownloadPackage downloads a payload specified by the URI and
//returns the local path for where the bits land
func DownloadPackage(installPackageURI string) (string, error) {
	log.Infoln("downloadPackage ENTER")
	log.Infoln("installPackageURI=", installPackageURI)

	path, err := util.GetFullPath()
	if err != nil {
		log.Errorln("GetFullPath Failed:", err)
		log.Infoln("downloadPackage LEAVE")
		return "", err
	}

	filename := util.GetFilenameFromURIOrFullPath(installPackageURI)
	log.Infoln("Filename:", filename)

	fullpath := util.AppendSlash(path) + filename
	log.Infoln("Fullpath:", fullpath)

	//create a downloaded file
	output, err := os.Create(fullpath)
	if err != nil {
		log.Errorln("Create File Failed:", err)
		log.Infoln("downloadPackage LEAVE")
		return "", err
	}

	//get the "executor" file
	resp, err := http.Get(installPackageURI)
	if err != nil {
		log.Errorln("HTTP GET Failed:", err)
		log.Infoln("downloadPackage LEAVE")
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	if err != nil {
		log.Errorln("IO Copy Failed:", err)
		log.Infoln("downloadPackage LEAVE")
		return "", err
	}
	output.Close()

	log.Infoln("downloadPackage Succeeded:", fullpath)
	log.Infoln("downloadPackage LEAVE")
	return fullpath, nil
}

//ParseVersionFromFilename this parses the version string out of the
//DEBs filename
func ParseVersionFromFilename(filename string) (string, error) {
	log.Debugln("ParseVersionFromFilename ENTER")
	log.Debugln("filename:", filename)

	r, err := regexp.Compile(".*([0-9]+\\.[0-9]+[\\.\\-][0-9]+\\.[0-9]+).*")
	if err != nil {
		log.Debugln("regexp is invalid")
		log.Debugln("ParseVersionFromFilename LEAVE")
		return "", err
	}
	strings := r.FindStringSubmatch(filename)
	if strings == nil || len(strings) < 2 {
		log.Debugln("Unable to find version from string")
		log.Debugln("ParseVersionFromFilename LEAVE")
		return "", ErrParseVersionFailed
	}

	version := strings[1]

	log.Debugln("Found:", version)
	log.Debugln("ParseVersionFromFilename LEAVE")

	return version, nil
}

//IsVersionStringHigher checks to see if one version is higher than the current
func IsVersionStringHigher(existing string, comparing string) bool {
	log.Debugln("IsVersionStringHigher ENTER")
	log.Debugln("existing:", existing)
	log.Debugln("comparing:", comparing)

	arr1 := strings.Split(existing, ".")
	arr2 := strings.Split(comparing, ".")

	for i := 0; i < len(arr1); i++ {
		tok2, err2 := strconv.Atoi(arr2[i])
		if err2 != nil {
			continue
		}
		tok1, err1 := strconv.Atoi(arr1[i])
		if err1 != nil {
			continue
		}
		if tok2 > tok1 {
			log.Debugln("New Higher:", comparing, ">", existing)
			log.Debugln("IsVersionStringHigher LEAVE")
			return true
		}
	}

	if len(arr2) > len(arr1) {
		log.Debugln("New Higher:", comparing, ">", existing)
		log.Debugln("IsVersionStringHigher LEAVE")
		return true
	}

	log.Debugln("Is Lower")
	log.Debugln("IsVersionStringHigher LEAVE")
	return false
}
