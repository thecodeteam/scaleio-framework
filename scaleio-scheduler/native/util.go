package util

import (
	"net"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
)

//ParseIP creates an IP object from a string
func ParseIP(address string) net.IP {
	addr, err := net.LookupIP(address)
	if err != nil {
		log.Errorln("LookupIP:", err)
	}
	if len(addr) < 1 {
		log.Errorln("failed to parse IP from address", address)
	}
	return addr[0]
}

//GetFullExePath returns the fullpath of the executable including the executable
//name itself
func GetFullExePath() (string, error) {
	path, err := os.Readlink("/proc/self/exe")
	if err != nil {
		log.Errorln("Readlink failed:", err)
		return "", nil
	}
	log.Debugln("EXE path:", path)
	return path, nil
}

//GetFullPath returns the fullpath of the executable without the executable name
func GetFullPath() (string, error) {
	path, err := os.Readlink("/proc/self/exe")
	if err != nil {
		log.Errorln("Readlink failed:", err)
		return "", nil
	}
	log.Debugln("EXE path:", path)

	tmp := GetPathFileFullFilename(path)
	return tmp, nil
}

//GetPathFileFullFilename returns the parent folder name
func GetPathFileFullFilename(path string) string {
	log.Debugln("GetPathFileFullFilename ENTER")
	log.Debugln("path:", path)
	last := strings.LastIndex(path, "/")
	if last == -1 {
		log.Debugln("No slash. Return Path:", path)
		log.Debugln("GetPathFileFullFilename LEAVE")
		return path
	}
	tmp := path[0:last]
	log.Debugln("Final Path:", tmp)
	log.Debugln("GetPathFileFullFilename LEAVE")
	return tmp
}

//GetFilenameFromURIOrFullPath retrieves the filename from an URI
func GetFilenameFromURIOrFullPath(path string) string {
	log.Debugln("GetFilenameFromURI ENTER")
	log.Debugln("path:", path)

	last := strings.LastIndex(path, "/")
	if last == -1 {
		log.Debugln("No slash. Return Path:", path)
		log.Debugln("GetFilenameFromURI LEAVE")
		return path
	}
	pathTmp := path[last+1:]
	log.Debugln("Return Path:", pathTmp)
	log.Debugln("GetFilenameFromURI LEAVE")

	return pathTmp
}

//AppendSlash appends a slash to a path if one is needed
func AppendSlash(path string) string {
	log.Debugln("AppendSlash ENTER")
	log.Debugln("path:", path)
	if path[len(path)-1] != '/' {
		path += "/"
	}
	log.Debugln("Return Path:", path)
	log.Debugln("GetFilenameFromURI LEAVE")
	return path
}
