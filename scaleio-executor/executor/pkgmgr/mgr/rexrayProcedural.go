package basemgr

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	rexrayBintrayRootURI = "https://dl.bintray.com/emccode/rexray/"
)

func getRexrayVersionFromBintray(state *types.ScaleIOFramework) (string, error) {
	url := rexrayBintrayRootURI + state.Rexray.Branch
	version, err := xplatform.GetInstance().Inst.GetVersionFromBintray(url)
	return version, err
}

func getRexrayVersionToInstall(state *types.ScaleIOFramework) (string, error) {
	if state.Rexray.Version == "latest" {
		version, err := getRexrayVersionFromBintray(state)
		return version, err
	}

	return state.Rexray.Version, nil
}

func doesSciniExistInRexrayInitD() (bool, error) {
	log.Debugln("doesSciniExistInRexrayInitD LEAVE")

	file, err := os.Open("/etc/init.d/rexray")
	if err != nil {
		log.Debugln("Failed on file Open:", err)
		log.Debugln("doesSciniExistInRexrayInitD LEAVE")
		return false, err
	}
	defer file.Close()

	r, err := regexp.Compile("scini")
	if err != nil {
		log.Debugln("regexp is invalid")
		log.Debugln("doesSciniExistInRexrayInitD LEAVE")
		return false, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		log.Debugln("Line:", line)
		if len(line) == 0 {
			continue
		}

		strings := r.FindStringSubmatch(line)
		if strings != nil || len(strings) == 1 {
			log.Debugln("Match found:", line)
			log.Debugln("doesSciniExistInRexrayInitD LEAVE")
			return true, nil
		}
	}

	log.Debugln("Scini is not configured in the rexray init.d")
	log.Debugln("doesSciniExistInRexrayInitD LEAVE")
	return false, nil
}

//RexraySetup procedure for setting up REX-Ray
func (nm *NodeManager) RexraySetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("BaseManager::RexraySetup ENTER")

	//REX-Ray Install
	rrVer, rrVerErr := getRexrayVersionToInstall(state)
	rrInst, rrInstErr := xplatform.GetInstance().Inst.GetInstalledVersion(types.RexRayPackageName, false)
	log.Debugln("rrVer:", rrVer)
	log.Debugln("rrVerErr:", rrVerErr)
	log.Debugln("rrInst:", rrInst)
	log.Debugln("rrInstErr:", rrInstErr)

	if rrVerErr != nil || rrInstErr != nil || rrVer != rrInst {
		gateway, err := common.GetGatewayAddress(state)
		if err != nil {
			log.Errorln("Unable to find the Gateway IP Address")
			log.Infoln("BaseManager::RexraySetup LEAVE")
			return false, err
		}

		//REX-Ray Install
		rexrayInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/rexray/install " +
			"| INSECURE=1 sh -"
		if strings.Compare(state.Rexray.Version, "latest") != 0 {
			rexrayInstallCmdline = "curl -ksSL https://dl.bintray.com/emccode/rexray/install | INSECURE=1 sh -s -- " +
				state.Rexray.Branch + " " + state.Rexray.Version
		} else if strings.Compare(state.Rexray.Branch, "stable") != 0 {
			rexrayInstallCmdline = "curl -ksSL https://dl.bintray.com/emccode/rexray/install | INSECURE=1 sh -s -- " +
				state.Rexray.Branch
		}

		err = xplatform.GetInstance().Run.Command(rexrayInstallCmdline, nm.RexrayInstallCheck, "")
		if err != nil {
			log.Errorln("Install REX-Ray Failed:", err)
			log.Infoln("BaseManager::RexraySetup LEAVE")
			return false, err
		}

		systemIdenifier := "systemName: " + state.ScaleIO.ClusterName
		if state.ScaleIO.ClusterID != "" {
			systemIdenifier = "systemId: " + state.ScaleIO.ClusterID
		}

		rexrayConfig := `rexray:
  logLevel: debug
libstorage:
  integration:
    volume:
      operations:
        mount:
          preempt: true
        unmount:
          ignoreUsedCount: true
  service: scaleio
  server:
    services:
      scaleio:
        driver: scaleio
        scaleio:
          endpoint: https://{IP_ADDRESS}/api
          insecure: true
          thinOrThick: ThinProvisioned
          userName: admin
          password: {PASSWORD}
          {SYSTEMIDENTIFIER}
          protectionDomainName: {PROTECTIONDOMAINNAME}
          storagePoolName: {STORAGEPOOLNAME}`

		rexrayConfig = strings.Replace(rexrayConfig, "{IP_ADDRESS}", gateway, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{PASSWORD}",
			state.ScaleIO.AdminPassword, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{SYSTEMIDENTIFIER}",
			systemIdenifier, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{PROTECTIONDOMAINNAME}",
			state.ScaleIO.ProtectionDomain, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{STORAGEPOOLNAME}",
			state.ScaleIO.StoragePool, -1)

		file, err := os.OpenFile("/etc/rexray/config.yml",
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			log.Errorln("Writing Config File Failed:", err)
			log.Infoln("BaseManager::RexraySetup LEAVE")
			return false, err
		}

		file.WriteString(rexrayConfig)
		file.Close()

		log.Debugln("Write Config File:")
		log.Debugln(rexrayConfig)

		//TODO need to change dependency based on init system
		found, errInitd := doesSciniExistInRexrayInitD()
		if errInitd == nil {
			if !found {
				log.Debugln("Modify REX-Ray init.d to add Scini dependency")

				writeSciniCmdline := "sed -i 's/\\/usr\\/bin\\/rexray start/if \\[ -e \\/etc\\/init.d\\/scini \\]\\; then \\/etc\\/init.d\\/scini start; fi\\n    \\/usr\\/bin\\/rexray start/' /etc/init.d/rexray"
				output, errScini := xplatform.GetInstance().Run.CommandOutput(writeSciniCmdline)
				if errScini != nil || len(output) > 0 {
					log.Errorln("Failed to add Scini dependency:", errScini)
					log.Infoln("BaseManager::RexraySetup LEAVE")
					return false, errScini
				}
			} else {
				log.Debugln("Scini has already been configured as a dependency on REX-Ray initd")
			}
		} else {
			log.Infoln("doesSciniExistInRexrayInitD failed:", errInitd)
			log.Infoln("BaseManager::RexraySetup LEAVE")
			return false, errInitd
		}
	} else {
		log.Infoln(types.RexRayPackageName, "is already installed")
	}

	if rrInst == "" && rrInstErr == nil {
		log.Debugln("No previous install of", types.RexRayPackageName,
			"exists. Reboot required!")
		log.Infoln("BaseManager::RexraySetup LEAVE")
		return true, nil
	}

	log.Infoln("BaseManager::RexraySetup Succeeded")
	log.Infoln("BaseManager::RexraySetup LEAVE")
	return false, nil
}

/*
func rexrayServerSetup() error {
	log.Infoln("RexrayServerSetup ENTER")

	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		log.Infoln("RexrayServerSetup LEAVE")
		return err
	}

	//REX-Ray Install
	rexrayInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/rexray/install | INSECURE=1 sh -"
	err = xplatform.GetInstance().Run.Command(rexrayInstallCmdline, rexrayInstallCheck, "")
	if err != nil {
		log.Errorln("Install REX-Ray Failed:", err)
		log.Infoln("RexrayServerSetup LEAVE")
		return err
	}

	rexrayConfig := `rexray:
  logLevel: debug
libstorage:
  host: tcp://127.0.0.1:7979
  embedded: true
  integration:
    volume:
      operations:
        mount:
          preempt: true
        unmount:
          ignoreUsedCount: true
  service: scaleio
  server:
    endpoints:
      public:
        address: tcp://:7979
    services:
      scaleio:
        driver: scaleio
        scaleio:
          endpoint: https://{IP_ADDRESS}/api
          insecure: true
          userName: admin
          password: {PASSWORD}
          systemName: {SYSTEMNAME}
          protectionDomainName: {PROTECTIONDOMAINNAME}
          storagePoolName: {STORAGEPOOLNAME}`

	gatewayIP := pri.IPAddress
	log.Infoln("Gateway IP to Use:", gatewayIP)
	if len(state.ScaleIO.LbGateway) > 0 {
		gatewayIP = state.ScaleIO.LbGateway
		log.Infoln("LbGateway Set. Using IP:", gatewayIP)
	}

	rexrayConfig = strings.Replace(rexrayConfig, "{IP_ADDRESS}", gatewayIP, -1)
	rexrayConfig = strings.Replace(rexrayConfig, "{PASSWORD}",
		state.ScaleIO.AdminPassword, -1)
	rexrayConfig = strings.Replace(rexrayConfig, "{SYSTEMNAME}",
		state.ScaleIO.ClusterName, -1)
	rexrayConfig = strings.Replace(rexrayConfig, "{PROTECTIONDOMAINNAME}",
		state.ScaleIO.ProtectionDomain, -1)
	rexrayConfig = strings.Replace(rexrayConfig, "{STORAGEPOOLNAME}",
		state.ScaleIO.StoragePool, -1)

	file, err := os.OpenFile("/etc/rexray/config.yml",
		os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		log.Errorln("Writing Config File Failed:", err)
		log.Infoln("RexrayServerSetup LEAVE")
		return err
	}

	file.WriteString(rexrayConfig)
	file.Close()

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStopCmdline := "rexray service stop -l debug"
	err = xplatform.GetInstance().Run.CommandEx(rexrayStopCmdline, rexrayStopCheck, "", 20)
	if err != nil {
		log.Warnln("REX-Ray stop Failed:", err)
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStartCmdline := "rexray service start -l debug"
	err = xplatform.GetInstance().Run.CommandEx(rexrayStartCmdline, rexrayStartCheck, "", 20)
	if err != nil {
		log.Errorln("REX-Ray start Failed:", err)
		log.Infoln("RexrayServerSetup LEAVE")
		return err
	}

	log.Infoln("RexrayServerSetup Succeeded")
	log.Infoln("RexrayServerSetup LEAVE")
	return nil
}

func rexrayClientSetup(state *types.ScaleIOFramework) error {
	log.Infoln("RexrayClientSetup ENTER")

	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		log.Infoln("RexrayClientSetup LEAVE")
		return err
	}

	//REX-Ray Install
	rexrayInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/rexray/install | INSECURE=1 sh -"
	err = xplatform.GetInstance().Run.Command(rexrayInstallCmdline, rexrayInstallCheck, "")
	if err != nil {
		log.Errorln("Install REX-Ray Failed:", err)
		log.Infoln("RexrayClientSetup LEAVE")
		return err
	}

	rexrayConfig := `rexray:
  logLevel: debug
libstorage:
  host:    tcp://{IP_ADDRESS}:7979
  service: scaleio`

	gatewayIP := pri.IPAddress
	log.Infoln("Gateway IP to Use:", gatewayIP)
	if len(state.ScaleIO.LbGateway) > 0 {
		gatewayIP = state.ScaleIO.LbGateway
		log.Infoln("LbGateway Set. Using IP:", gatewayIP)
	}

	rexrayConfig = strings.Replace(rexrayConfig, "{IP_ADDRESS}", gatewayIP, -1)

	file, err := os.OpenFile("/etc/rexray/config.yml",
		os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		log.Errorln("Writing Config File Failed:", err)
		log.Infoln("RexrayClientSetup LEAVE")
		return err
	}

	file.WriteString(rexrayConfig)
	file.Close()

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStopCmdline := "rexray service stop -l debug"
	err = xplatform.GetInstance().Run.CommandEx(rexrayStopCmdline, rexrayStopCheck, "", 20)
	if err != nil {
		log.Warnln("REX-Ray stop Failed:", err)
	}

	time.Sleep(time.Duration(common.DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStartCmdline := "rexray service start -l debug"
	err = xplatform.GetInstance().Run.CommandEx(rexrayStartCmdline, rexrayStartCheck, "", 20)
	if err != nil {
		log.Errorln("REX-Ray start Failed:", err)
		log.Infoln("RexrayClientSetup LEAVE")
		return err
	}

	log.Infoln("RexrayClientSetup Succeeded")
	log.Infoln("RexrayClientSetup LEAVE")
	return nil
}
*/
