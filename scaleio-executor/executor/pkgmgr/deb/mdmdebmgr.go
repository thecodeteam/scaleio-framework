package deb

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	//Environment
	aiozipCheck                = "[0-9]+ upgraded|[0-9]+ newly"
	genericInstallCheck        = "1 upgraded|1 newly"
	requiredKernelVersionCheck = "4.2.0-30-generic"

	//ScaleIO node
	mdmInstallCheck     = "mdm start/running"
	sdsInstallCheck     = "sds start/running"
	sdcInstallCheck     = "Success configuring module"
	liaInstallCheck     = "lia start/running"
	liaRestartCheck     = liaInstallCheck
	gatewayInstallCheck = "The EMC ScaleIO Gateway is running"
	gatewayRestartCheck = "scaleio-gateway start/running"

	//REX-Ray
	rexrayInstallCheck = "rexray has been installed to"

	//Isolator
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//MdmDebMgr implementation for MdmDebMgr
type MdmDebMgr struct {
	*mgr.MdmManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (mdm *MdmDebMgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("EnvironmentSetup ENTER")

	aioErr := xplatform.GetInstance().Inst.IsInstalled("libaio1")
	zipErr := xplatform.GetInstance().Inst.IsInstalled("zip")
	if aioErr != nil || zipErr != nil {
		log.Infoln("Installing libaio1 and zip")

		miscCmdline := "apt-get -y install libaio1 zip"
		err := xplatform.GetInstance().Run.Command(miscCmdline, aiozipCheck, "")
		if err != nil {
			log.Errorln("Install Prerequisites Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("libaio1 and zip are already installed")
	}

	kernelErr := xplatform.GetInstance().Inst.IsInstalled("linux-image-4.2.0-30-generic")
	if kernelErr != nil {
		log.Infoln("Installing linux-image-4.2.0-30-generic")

		kernelCmdline := "apt-get -y install linux-image-4.2.0-30-generic"
		err := xplatform.GetInstance().Run.Command(kernelCmdline, genericInstallCheck, "")
		if err != nil {
			log.Errorln("Install Kernel Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("linux-image-4.2.0-30-generic is already installed")
	}

	//get running kernel version
	kernelVer, kernelVerErr := xplatform.GetInstance().Sys.GetRunningKernelVersion()
	if kernelVerErr != nil {
		log.Errorln("Kernel Version Check Failed:", kernelVerErr)
		log.Infoln("EnvironmentSetup LEAVE")
		return false, kernelVerErr
	}

	if kernelVer != requiredKernelVersionCheck {
		log.Errorln("Kernel is installed but not running. Reboot Required!")
		log.Infoln("EnvironmentSetup LEAVE")
		return true, nil
	}
	log.Infoln("Already running kernel version", requiredKernelVersionCheck)
	//get running kernel version

	log.Infoln("EnvironmentSetup Succeeded")
	log.Infoln("EnvironmentSetup LEAVE")
	return false, nil
}

//NewMdmDebMgr generates a MdmDebMgr object
func NewMdmDebMgr(state *types.ScaleIOFramework) MdmDebMgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmDebMgr := MdmDebMgr{myMdmMgr}

	//ScaleIO node
	myMdmDebMgr.MdmManager.SdsPackageName = types.DebSdsPackageName
	myMdmDebMgr.MdmManager.SdsPackageDownload = state.ScaleIO.Deb.DebSds
	myMdmDebMgr.MdmManager.SdsInstallCmd = "dpkg -i {LocalSds}"
	myMdmDebMgr.MdmManager.SdsInstallCheck = sdsInstallCheck
	myMdmDebMgr.MdmManager.SdcPackageName = types.DebSdcPackageName
	myMdmDebMgr.MdmManager.SdcPackageDownload = state.ScaleIO.Deb.DebSdc
	myMdmDebMgr.MdmManager.SdcInstallCmd = "MDM_IP={MdmPair} dpkg -i {LocalSdc}"
	myMdmDebMgr.MdmManager.SdcInstallCheck = sdcInstallCheck
	myMdmDebMgr.MdmManager.MdmPackageName = types.DebMdmPackageName
	myMdmDebMgr.MdmManager.MdmPackageDownload = state.ScaleIO.Deb.DebMdm
	myMdmDebMgr.MdmManager.MdmInstallCmd = "MDM_ROLE_IS_MANAGER={PriOrSec} dpkg -i {LocalMdm}"
	myMdmDebMgr.MdmManager.MdmInstallCheck = mdmInstallCheck
	myMdmDebMgr.MdmManager.LiaPackageName = types.DebLiaPackageName
	myMdmDebMgr.MdmManager.LiaPackageDownload = state.ScaleIO.Deb.DebLia
	myMdmDebMgr.MdmManager.LiaInstallCmd = "TOKEN=" + state.ScaleIO.AdminPassword + " dpkg -i {LocalLia}"
	myMdmDebMgr.MdmManager.LiaInstallCheck = liaInstallCheck
	myMdmDebMgr.MdmManager.LiaRestartCheck = liaRestartCheck
	myMdmDebMgr.MdmManager.GatewayPackageName = types.DebGwPackageName
	myMdmDebMgr.MdmManager.GatewayPackageDownload = state.ScaleIO.Deb.DebGw
	myMdmDebMgr.MdmManager.GatewayInstallCmd = "GATEWAY_ADMIN_PASSWORD=" + state.ScaleIO.AdminPassword + " dpkg -i {LocalGw}"
	myMdmDebMgr.MdmManager.GatewayInstallCheck = gatewayInstallCheck
	myMdmDebMgr.MdmManager.GatewayRestartCheck = gatewayRestartCheck

	//REX-Ray
	myMdmDebMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myMdmDebMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myMdmDebMgr
}
