package mgr

import (
	"bufio"
	"errors"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
	xplatformsys "github.com/dvonthenen/goxplatform/init"

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

var (
	//ErrAddDependencyFailed Failed to add the scini dependency to REX-Ray
	ErrAddDependencyFailed = errors.New("Failed to add the scini dependency to REX-Ray")
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

func fixSciniDepInRexrayInitD() error {
	log.Debugln("fixSciniDepInRexrayInitD ENTER")

	writeSciniCmdline := "sed -i 's/\\/usr\\/bin\\/rexray start/if \\[ -e \\/etc\\/init.d\\/scini \\]\\; then \\/etc\\/init.d\\/scini start; fi\\n    \\/usr\\/bin\\/rexray start/' /etc/init.d/rexray"
	output, errScini := xplatform.GetInstance().Run.CommandOutput(writeSciniCmdline)
	if errScini != nil {
		log.Errorln("Failed to add Scini dependency:", errScini)
		log.Debugln("fixSciniDepInRexrayInitD LEAVE")
		return errScini
	}
	if len(output) > 0 {
		log.Errorln("Output Error:", output)
		log.Debugln("fixSciniDepInRexrayInitD LEAVE")
		return ErrAddDependencyFailed
	}

	log.Debugln("Scini has been configured as a dependency on REX-Ray InitD")
	log.Debugln("fixSciniDepInRexrayInitD LEAVE")

	return nil
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

	log.Debugln("Scini is not configured in the rexray InitD")
	log.Debugln("doesSciniExistInRexrayInitD LEAVE")
	return false, nil
}

//RexraySetup procedure for setting up REX-Ray
func (nm *NodeManager) RexraySetup(state *types.ScaleIOFramework, executorID string) (bool, error) {
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

		self := common.GetSelfNode(state, executorID)
		if self == nil {
			log.Errorln("Unable to find self node")
			log.Infoln("RexraySetup LEAVE")
			return false, common.ErrFindNodeFailed
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

		//TODO change this when rexray supports multiple pd/sp
		protectionDomain := state.ScaleIO.ProtectionDomain
		storagePool := state.ScaleIO.StoragePool

		if len(self.ConsumesDomains) > 0 {
			for _, domain := range self.ConsumesDomains {
				protectionDomain = domain.Name
				for _, pool := range domain.Pools {
					storagePool = pool.Name
					break //only allow 1 pool
				}
				break //only allow 1 domain
			}
		} else if len(self.ProvidesDomains) > 0 {
			for _, domain := range self.ProvidesDomains {
				protectionDomain = domain.Name
				for _, pool := range domain.Pools {
					storagePool = pool.Name
					break //only allow 1 pool
				}
				break //only allow 1 domain
			}
		}
		log.Debugln("ProtectionDomain:", protectionDomain)
		log.Debugln("StoragePool:", storagePool)
		//TODO change this when rexray supports multiple pd/sp

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
    rexray:
      type: docker
      libstorage:
        service: scaleio
      host: unix:///run/docker/plugins/rexray.sock
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
		rexrayConfig = strings.Replace(rexrayConfig, "{PASSWORD}", state.ScaleIO.AdminPassword, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{SYSTEMIDENTIFIER}", systemIdenifier, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{PROTECTIONDOMAINNAME}", protectionDomain, -1)
		rexrayConfig = strings.Replace(rexrayConfig, "{STORAGEPOOLNAME}", storagePool, -1)

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

		//special case for ubuntu
		initType := xplatform.GetInstance().Init.GetInitSystemType()
		if initType == xplatformsys.InitUpdateRcD {
			found, errInitd := doesSciniExistInRexrayInitD()
			if errInitd != nil {
				log.Infoln("doesSciniExistInRexrayInitD failed:", errInitd)
				log.Infoln("fixRexrayDependencyToScini LEAVE")
				return false, errInitd
			}
			if !found {
				log.Debugln("Modify REX-Ray SystemD to add Scini dependency")

				errScini := fixSciniDepInRexrayInitD()
				if errScini != nil {
					log.Errorln("Failed to add Scini dependency:", errScini)
					log.Debugln("fixRexrayDependencyToScini LEAVE")
					return false, errScini
				}

				log.Debugln("Scini has been configured as a dependency on REX-Ray initd")
			} else {
				log.Debugln("Scini has already been configured as a dependency on REX-Ray initd")
			}
		}

		err = xplatform.GetInstance().Init.AddDependentService("rexray", "scini")
		if err != nil {
			log.Infoln("AddDependentService scini<-rexray failed:", err)
			log.Infoln("fixRexrayDependencyToScini LEAVE")
			return false, err
		}
	} else {
		log.Infoln(types.RexRayPackageName, "is already installed")
		time.Sleep(time.Duration(common.DelayIfInstalledInSeconds) * time.Second)
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
