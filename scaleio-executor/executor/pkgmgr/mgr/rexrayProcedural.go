package mgr

import (
	"os"
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

const (
	initUnknown = iota
	initSystemD
	initUpdateRcD
	initChkConfig
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

//RexraySetup procedure for setting up REX-Ray
func (nm *NodeManager) RexraySetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("RexraySetup ENTER")

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
			log.Infoln("RexraySetup LEAVE")
			return false, err
		}

		//REX-Ray configuration
		err = os.MkdirAll("/etc/rexray", os.ModeDir)
		if err != nil {
			log.Infoln("Failed to create directory (/etc/rexray) with err:", err)
			log.Infoln("RexraySetup LEAVE")
			return false, err
		}

		systemIdenifier := "systemName: " + state.ScaleIO.ClusterName
		if state.ScaleIO.ClusterID != "" {
			systemIdenifier = "systemId: " + state.ScaleIO.ClusterID
		}

		rexrayConfig := `rexray:
  logLevel: debug
  modules:
    default-docker:
      type: docker
      libstorage:
        service: scaleio 
      host: unix:///run/docker/plugins/docker.sock
    mesos:
      type: docker
      libstorage:
        service: scaleio
      host: unix:///run/docker/plugins/mesos.sock
      libstorage:
        integration:
          volume:
            operations:
              unmount:
                ignoreUsedCount: true
libstorage:
  service: scaleio
  integration:
    volume:
      operations:
        mount:
          preempt: true
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
			log.Infoln("RexraySetup LEAVE")
			return false, err
		}

		file.WriteString(rexrayConfig)
		file.Close()

		log.Debugln("Write Config File:")
		log.Debugln(rexrayConfig)

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
			log.Infoln("RexraySetup LEAVE")
			return false, err
		}

		err = xplatform.GetInstance().Init.AddDependentService("rexray", "scini")
		if err != nil {
			log.Infoln("AddDependentService scini<-rexray failed:", err)
			log.Infoln("fixRexrayDependencyToScini LEAVE")
			return false, err
		}
	} else {
		log.Infoln(types.RexRayPackageName, "is already installed")
	}

	if rrInst == "" {
		log.Infoln("No previous install of", types.RexRayPackageName,
			"exists. Reboot required!")
		log.Infoln("RexraySetup LEAVE")
		return true, nil
	}

	log.Infoln("Previous install of", types.RexRayPackageName,
		"exists. No reboot required.")
	log.Infoln("RexraySetup LEAVE")
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
