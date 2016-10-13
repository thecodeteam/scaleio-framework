package common

import (
	"bufio"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	isolatorInstallDir = "/usr/lib"
	slaveRestartCheck  = "mesos-slave start/running, process"
	dvdcliInstallCheck = "dvdcli has been installed to"

	dvdcliBintrayRootURI = "https://dl.bintray.com/emccode/dvdcli/stable/"
)

var (
	//ErrRegexMatchFailed failed to validate the regex against string
	ErrRegexMatchFailed = errors.New("Failed to validate the regex against string")

	//ErrIsolatorNotInstalled failed to parse version from filename
	ErrIsolatorNotInstalled = errors.New("The Mesos Module Isolator is not installed")

	//ErrIsolatorNameInvalid failed to parse version from filename
	ErrIsolatorNameInvalid = errors.New("The Mesos Module Isolator name is invalid")
)

func getDvdcliVersionFromBintray() (string, error) {
	version, err := xplatform.GetInstance().Inst.GetVersionFromBintray(dvdcliBintrayRootURI)
	return version, err
}

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
		return "", ErrRegexMatchFailed
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

func getMesosPropertyFileContents(fullfilename string) (string, error) {
	log.Debugln("getMesosPropertyFileContents LEAVE")
	log.Debugln("fullfilename:", fullfilename)

	file, err := os.Open(fullfilename)
	if err != nil {
		log.Debugln("Failed on file Open:", err)
		log.Debugln("getMesosPropertyFileContents LEAVE")
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		log.Debugln("Line:", line)
		if len(line) == 0 {
			continue
		}

		log.Debugln("Got existing properties:", line)
		log.Debugln("getMesosPropertyFileContents LEAVE")
		return line, nil
	}

	log.Debugln("File exists but is empty")
	log.Debugln("getMesosPropertyFileContents LEAVE")
	return "", nil
}

func doesLineExistInMesosPropertyFile(fullfilename string, needle string) error {
	log.Debugln("doesLineExistInMesosPropertyFile ENTER")
	log.Debugln("fullfilename:", fullfilename)
	log.Debugln("needle:", needle)

	line, err := getMesosPropertyFileContents(fullfilename)
	if err != nil {
		log.Debugln("Failed getMesosPropertyFileContents:", err)
		log.Debugln("doesLineExistInMesosPropertyFile LEAVE")
		return err
	}

	r, err := regexp.Compile(needle)
	if err != nil {
		log.Debugln("regexp is invalid")
		log.Debugln("doesLineExistInMesosPropertyFile LEAVE")
		return err
	}
	strings := r.FindStringSubmatch(line)
	if strings == nil || len(strings) != 1 {
		log.Debugln("Unable to find specified content in file")
		log.Debugln("doesLineExistInMesosPropertyFile LEAVE")
		return ErrRegexMatchFailed
	}

	found := strings[0]
	log.Debugln("Found:", found)

	log.Debugln("Property exists in Properties file")
	log.Debugln("doesLineExistInMesosPropertyFile LEAVE")
	return nil
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

		localIsolator, err := xplatform.GetInstance().Inst.DownloadPackage(state.Isolator.Binary)
		if err != nil {
			log.Errorln("Error downloading Isolator package:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		//Copy File
		dstFullPath := isolatorInstallDir + "/" + xplatform.GetInstance().Fs.GetFilenameFromURIOrFullPath(localIsolator)
		err = xplatform.GetInstance().Fs.FileCopy(localIsolator, dstFullPath)
		if err != nil {
			log.Errorln("Failed to Copy isolator to Dst:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		//Create the mesos-slave isolation file
		isolationFile := "/etc/mesos-slave/isolation"
		isolationFileContents := "com_emccode_mesos_DockerVolumeDriverIsolator"

		err = doesLineExistInMesosPropertyFile(isolationFile, isolationFileContents)
		if err != nil {
			contents, errProp := getMesosPropertyFileContents(isolationFile)
			if err != nil {
				log.Warnln("getMesosPropertyFileContents returned err:", errProp)
			}

			if len(contents) > 0 {
				isolationFileContents = isolationFileContents + "," + contents
				log.Infoln("Preserving existing contents:", isolationFileContents)
			}

			isolationFile, errWrite := os.OpenFile(isolationFile,
				os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
			if errWrite != nil {
				log.Errorln("Writing Isolation File Failed:", errWrite)
				log.Infoln("SetupIsolator LEAVE")
				return errWrite
			}

			isolationFile.WriteString(isolationFileContents)
			isolationFile.Close()
		}

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
      "file": "/usr/lib/libmesos_dvdi_isolator-{ISO_VERSION}.so",
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

		dvdimodFileContents = strings.Replace(dvdimodFileContents, "{ISO_VERSION}", isoVer, -1)

		dvdimodFile, err := os.OpenFile("/usr/lib/dvdi-mod.json",
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Errorln("Writing dvdi-mod.json File Failed:", err)
			log.Infoln("SetupIsolator LEAVE")
			return err
		}

		dvdimodFile.WriteString(dvdimodFileContents)
		dvdimodFile.Close()
	} else {
		log.Infoln("Mesos Module Isolator is already installed")
	}

	//DVDCLI install
	dcVer, dcVerErr := getDvdcliVersionFromBintray()
	dcInst, dcInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(types.DvdcliPackageName, false)
	dcInst = xplatform.GetInstance().Inst.CorrectVersionFromDeb(dcInst)
	log.Debugln("dcVer:", dcVer)
	log.Debugln("dcVerErr:", dcVerErr)
	log.Debugln("dcInst:", dcInst)
	log.Debugln("dcInstErr:", dcInstErr)

	if dcVerErr != nil || dcInstErr != nil || dcVer != dcInst {
		dvdcliInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/dvdcli/install " +
			"| INSECURE=1 sh -"
		err := xplatform.GetInstance().Run.Command(dvdcliInstallCmdline, dvdcliInstallCheck, "")
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
