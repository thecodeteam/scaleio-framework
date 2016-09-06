package core

import (
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/exec"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/native/installers/deb"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//constants for verifying that the command lines executed properly
const (
	rexrayBintrayRootURI = "https://dl.bintray.com/emccode/rexray/"
	rexrayInstallCheck   = "rexray has been installed to"
	rexrayStopCheck      = "SUCCESS!|os: process already finished|REX-Ray is already stopped"
	rexrayStartCheck     = "SUCCESS!|REX-Ray already running at"
	rexrayRunningCheck   = "REX-Ray is running at PID"
	rexrayEnableCheck    = "Adding system startup for"
)

func getRexrayVersionToInstall(state *types.ScaleIOFramework) (string, error) {
	if state.Rexray.Version == "latest" {
		version, err := installers.GetRexrayVersionFromBintray(state)
		return version, err
	}

	return state.Rexray.Version, nil
}

func rexraySetup(state *types.ScaleIOFramework) error {
	log.Infoln("RexraySetup ENTER")

	//REX-Ray Install
	rrVer, rrVerErr := getRexrayVersionToInstall(state)
	rrInst, rrInstErr := deb.GetInstalledVersion(types.RexRayPackageName, false)
	rrInst = installers.CorrectVersionFromDeb(rrInst)
	log.Debugln("rrVer:", rrVer)
	log.Debugln("rrVerErr:", rrVerErr)
	log.Debugln("rrInst:", rrInst)
	log.Debugln("rrInstErr:", rrInstErr)

	if rrVerErr != nil || rrInstErr != nil || rrVer != rrInst {
		gateway, err := getGatewayAddress(state)
		if err != nil {
			log.Errorln("Unable to find the Gateway IP Address")
			log.Infoln("RexraySetup LEAVE")
			return err
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

		err = exec.RunCommand(rexrayInstallCmdline, rexrayInstallCheck, "")
		if err != nil {
			log.Errorln("Install REX-Ray Failed:", err)
			log.Infoln("RexraySetup LEAVE")
			return err
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
          userName: admin
          password: {PASSWORD}
          systemName: {SYSTEMNAME}
          protectionDomainName: {PROTECTIONDOMAINNAME}
          storagePoolName: {STORAGEPOOLNAME}`

		rexrayConfig = strings.Replace(rexrayConfig, "{IP_ADDRESS}", gateway, -1)
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
			log.Infoln("RexraySetup LEAVE")
			return err
		}

		file.WriteString(rexrayConfig)
		file.Close()

		time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

		rexrayStopCmdline := "service rexray stop"
		err = exec.RunCommand(rexrayStopCmdline, rexrayStopCheck, "")
		if err != nil {
			log.Warnln("REX-Ray stop Failed:", err)
		}

		time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

		rexrayStartCmdline := "service rexray start"
		err = exec.RunCommand(rexrayStartCmdline, rexrayStartCheck, "")
		if err != nil {
			log.Errorln("REX-Ray start Failed:", err)
			log.Infoln("RexrayClientSetup LEAVE")
			return err
		}

		rexrayEnableCmdline := "update-rc.d rexray enable"
		err = exec.RunCommand(rexrayEnableCmdline, rexrayEnableCheck, "")
		if err != nil {
			log.Errorln("REX-Ray enable Failed:", err)
			log.Infoln("RexrayClientSetup LEAVE")
			return err
		}
	} else {
		log.Infoln(types.RexRayPackageName, "is already installed")

		rexrayStartCmdline := "service rexray start"
		err := exec.RunCommand(rexrayStartCmdline, rexrayStartCheck, "")
		if err != nil {
			log.Warnln("REX-Ray start Failed:", err)
		}
	}

	log.Infoln("RexraySetup Succeeded")
	log.Infoln("RexraySetup LEAVE")
	return nil
}

/*
func rexrayServerSetup(state *types.ScaleIOFramework) error {
	log.Infoln("RexrayServerSetup ENTER")

	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		log.Infoln("RexrayServerSetup LEAVE")
		return err
	}

	//REX-Ray Install
	rexrayInstallCmdline := "curl -ksSL https://dl.bintray.com/emccode/rexray/install | INSECURE=1 sh -"
	err = exec.RunCommand(rexrayInstallCmdline, rexrayInstallCheck, "")
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

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStopCmdline := "rexray service stop -l debug"
	err = exec.RunCommandEx(rexrayStopCmdline, rexrayStopCheck, "", 20)
	if err != nil {
		log.Warnln("REX-Ray stop Failed:", err)
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStartCmdline := "rexray service start -l debug"
	err = exec.RunCommandEx(rexrayStartCmdline, rexrayStartCheck, "", 20)
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
	err = exec.RunCommand(rexrayInstallCmdline, rexrayInstallCheck, "")
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

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStopCmdline := "rexray service stop -l debug"
	err = exec.RunCommandEx(rexrayStopCmdline, rexrayStopCheck, "", 20)
	if err != nil {
		log.Warnln("REX-Ray stop Failed:", err)
	}

	time.Sleep(time.Duration(DelayBetweenCommandsInSeconds) * time.Second)

	rexrayStartCmdline := "rexray service start -l debug"
	err = exec.RunCommandEx(rexrayStartCmdline, rexrayStartCheck, "", 20)
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
