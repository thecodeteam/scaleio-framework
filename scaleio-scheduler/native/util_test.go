package util

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	assert "github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)
	//log.Infoln("Start tests")
	m.Run()
}

func TestGetFilename(t *testing.T) {
	URI := "http://127.0.0.1:8080/dir/myfile.deb"
	file := GetFilenameFromURIOrFullPath(URI)
	assert.Equal(t, file, "myfile.deb")
}

func TestFilenameOnly(t *testing.T) {
	URI := "myfile.deb"
	file := GetFilenameFromURIOrFullPath(URI)
	assert.Equal(t, file, "myfile.deb")
}

func TestPathFromFullPath(t *testing.T) {
	path := "/tmp/dir/myfile.deb"
	dir := GetPathFileFullFilename(path)
	assert.Equal(t, dir, "/tmp/dir")
}
