package core

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-executor/native"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/exec"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers/deb"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	isolatorInstallDir = "/usr/lib"
	slaveRestartCheck  = "mesos-slave start/running, process"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

var (
	//ErrParseIsolatorVersionFailed failed to parse version from filename
	ErrParseIsolatorVersionFailed = errors.New("Failed to parse version from filename")

	//ErrIsolatorNotInstalled failed to parse version from filename
	ErrIsolatorNotInstalled = errors.New("The Mesos Module Isolator is not installed")

	//ErrIsolatorNameInvalid failed to parse version from filename
	ErrIsolatorNameInvalid = errors.New("The Mesos Module Isolator name is invalid")
)

func parseIsolatorVersionFromFilename(filename string) (string, error) {
	log.Debugln("parseIsolatorVersionFromFilename ENTER")
	log.Debugln("filename:", filename)

	r, err := regexp.Compile(".*([0-9]+\\.[0-9]+\\.[0-9]+)\\.so")
	if err != nil {
		log.Debugln("regexp is invalid")
		log.Debugln("parseIsolatorVersionFromFilename LEAVE")
		return "", err
	}
	strings := r.FindStringSubmatch(filename)
	if strings == nil || len(strings) < 2 {
		log.Debugln("Unable to find version from string")
		log.Debugln("parseIsolatorVersionFromFilename LEAVE")
		return "", ErrParseIsolatorVersionFailed
	}

	version := strings[1]

	log.Debugln("Found:", version)
	log.Debugln("parseIsolatorVersionFromFilename LEAVE")

	return version, nil
}

func findIsolatorVersionOnFilesystem() (string, error) {
	log.Debugln("findIsolatorOnFilesystem ENTER")

	items, err := ioutil.ReadDir(isolatorInstallDir)
	if err != nil {
		log.Debugln("ReadDir Failed:", err)
		log.Debugln("findIsolatorOnFilesystem LEAVE")
		return "", err
	}

	for _, item := range items {
		log.Debugln("Item:", item.Name())
		if item.IsDir() {
			log.Debugln("Is Dir")
			continue
		}

		if !strings.Contains(item.Name(), "libmesos_dvdi_isolator") {
			log.Debugln("Is not the Mesos Isolator file")
			continue
		}

		version, err := parseIsolatorVersionFromFilename(item.Name())
		if err != nil {
			log.Debugln("ReadDir Failed:", err)
			log.Debugln("findIsolatorOnFilesystem LEAVE")
			return "", ErrIsolatorNameInvalid
		}

		log.Debugln("findIsolatorOnFilesystem LEAVE")
		return version, nil
	}

	log.Debugln("Unable to find isolator installation")
	log.Debugln("findIsolatorOnFilesystem LEAVE")
	return "", ErrIsolatorNotInstalled
}

func setupIsolator(state *types.ScaleIOFramework) error {
	log.Infoln("SetupIsolator ENTER")

	//Mesos Isolator Install
	isoVer, isoVerErr := parseIsolatorVersionFromFilename(state.Isolator.Binary)
	isoInst, isoInstErr := findIsolatorVersionOnFilesystem()
	log.Debugln("isoVer:", isoVer)
	log.Debugln("isoVerEr:", isoVerErr)
	log.Debugln("isoInst:", isoInst)
	log.Debugln("isoInstErr:", isoInstErr)

	if isoVerErr != nil || isoInstErr != nil || isoVer != isoInst {
		log.Infoln("Installing Mesos Isolator")

		localIsolator, err := installers.DownloadPackage(state.Isolator.Binary)
		if err != nil {
			log.Errorln("Error downloading Isolator package:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		//Copy File
		dstFullPath := isolatorInstallDir + "/" + native.GetFilenameFromURIOrFullPath(localIsolator)
		err = native.FileCopy(localIsolator, dstFullPath)
		if err != nil {
			log.Errorln("Failed to Copy isolator to Dst:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		//Create the mesos-slave isolation file
		isolationFileContents := "com_emccode_mesos_DockerVolumeDriverIsolator"

		isolationFile, err := os.OpenFile("/etc/mesos-slave/isolation",
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Errorln("Writing Isolation File Failed:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		isolationFile.WriteString(isolationFileContents)
		isolationFile.Close()

		//Create the mesos-slave modules file
		modulesFileContents := "file:///usr/lib/dvdi-mod.json"

		modulesFile, err := os.OpenFile("/etc/mesos-slave/modules",
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Errorln("Writing Modules File Failed:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		modulesFile.WriteString(modulesFileContents)
		modulesFile.Close()

		//Create the dvdi-mod.json file
		dvdimodFileContents := `{
  "libraries": [
    {
      "file": "/usr/lib/libmesos_dvdi_isolator-1.0.0.so",
      "modules": [
        {
          "name": "com_emccode_mesos_DockerVolumeDriverIsolator",
          "parameters": [
            {
              "key": "isolator_command",
              "value": "/emc/dvdi_isolator"
            }
          ]
        }
      ]
    }
  ]
}`

		dvdimodFile, err := os.OpenFile("/usr/lib/dvdi-mod.json",
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Errorln("Writing dvdi-mod.json File Failed:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		dvdimodFile.WriteString(dvdimodFileContents)
		dvdimodFile.Close()

		time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

		slaveRestartCmdline := "service mesos-slave restart"
		err = exec.RunCommand(slaveRestartCmdline, slaveRestartCheck, "")
		if err != nil {
			log.Errorln("Mesos Slave restart Failed:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}
	} else {
		log.Infoln("Mesos Module Isolator is already installed")
	}

	//DVDCLI install
	dcVer, dcVerErr := installers.GetDvdliVersionFromBintray()
	dcInst, dcInstErr := deb.GetInstalledVersion(types.DvdcliPackageName, false)
	dcInst = installers.CorrectVersionFromDeb(dcInst)
	log.Debugln("dcVer:", dcVer)
	log.Debugln("dcVerErr:", dcVerErr)
	log.Debugln("dcInst:", dcInst)
	log.Debugln("dcInstErr:", dcInstErr)

	if dcVerErr != nil || dcInstErr != nil || dcVer != dcInst {
		dvdcliInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/dvdcli/install " +
			"| INSECURE=1 sh -"
		err := exec.RunCommand(dvdcliInstallCmdline, dvdcliInstallCheck, "")
		if err != nil {
			log.Errorln("Install DVDCLI Failed:", err)
			log.Infoln("RexraySetup LEAVE")
			return err
		}
	} else {
		log.Infoln("DVDCLI is already installed")
	}

	log.Infoln("SetupIsolator Succeeded")
	log.Infoln("SetupIsolator LEAVE")
	return nil
}
