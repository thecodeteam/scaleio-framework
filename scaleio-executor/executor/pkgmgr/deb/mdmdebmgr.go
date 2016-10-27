package deb

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
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
func NewMdmDebMgr(state *types.ScaleIOFramework) *MdmDebMgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmDebMgr := &MdmDebMgr{myMdmMgr}

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
	myMdmDebMgr.MdmManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myMdmDebMgr.MdmManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myMdmDebMgr
}
