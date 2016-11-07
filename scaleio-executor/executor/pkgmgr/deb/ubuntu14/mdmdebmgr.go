package ubuntu14

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//MdmDebMgr implementation for MdmDebMgr
type MdmDebUbuntu14Mgr struct {
	*mgr.MdmManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (mdm *MdmDebUbuntu14Mgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
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

	kernelErr := xplatform.GetInstance().Inst.IsInstalled("linux-image-4.4.0-38-generic")
	if kernelErr != nil {
		log.Infoln("Installing linux-image-4.4.0-38-generic")

		kernelCmdline := "apt-get -y install linux-image-4.4.0-38-generic"
		err := xplatform.GetInstance().Run.Command(kernelCmdline, genericInstallCheck, "")
		if err != nil {
			log.Errorln("Install Kernel Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("linux-image-4.4.0-38-generic is already installed")
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

//NewMdmDebUbuntu14Mgr generates a MdmDebMgr object
func NewMdmDebUbuntu14Mgr(state *types.ScaleIOFramework) *MdmDebUbuntu14Mgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmDebUbuntu14Mgr := &MdmDebUbuntu14Mgr{myMdmMgr}

	//ScaleIO node
	myMdmDebUbuntu14Mgr.MdmManager.SdsPackageName = types.Ubuntu14SdsPackageName
	myMdmDebUbuntu14Mgr.MdmManager.SdsPackageDownload = state.ScaleIO.Ubuntu14.Sds
	myMdmDebUbuntu14Mgr.MdmManager.SdsInstallCmd = "dpkg -i {LocalSds}"
	myMdmDebUbuntu14Mgr.MdmManager.SdsInstallCheck = sdsInstallCheck
	myMdmDebUbuntu14Mgr.MdmManager.SdcPackageName = types.Ubuntu14SdcPackageName
	myMdmDebUbuntu14Mgr.MdmManager.SdcPackageDownload = state.ScaleIO.Ubuntu14.Sdc
	myMdmDebUbuntu14Mgr.MdmManager.SdcInstallCmd = "MDM_IP={MdmPair} dpkg -i {LocalSdc}"
	myMdmDebUbuntu14Mgr.MdmManager.SdcInstallCheck = sdcInstallCheck
	myMdmDebUbuntu14Mgr.MdmManager.MdmPackageName = types.Ubuntu14MdmPackageName
	myMdmDebUbuntu14Mgr.MdmManager.MdmPackageDownload = state.ScaleIO.Ubuntu14.Mdm
	myMdmDebUbuntu14Mgr.MdmManager.MdmInstallCmd = "MDM_ROLE_IS_MANAGER={PriOrSec} dpkg -i {LocalMdm}"
	myMdmDebUbuntu14Mgr.MdmManager.MdmInstallCheck = mdmInstallCheck
	myMdmDebUbuntu14Mgr.MdmManager.LiaPackageName = types.Ubuntu14LiaPackageName
	myMdmDebUbuntu14Mgr.MdmManager.LiaPackageDownload = state.ScaleIO.Ubuntu14.Lia
	myMdmDebUbuntu14Mgr.MdmManager.LiaInstallCmd = "TOKEN=" + state.ScaleIO.AdminPassword + " dpkg -i {LocalLia}"
	myMdmDebUbuntu14Mgr.MdmManager.LiaInstallCheck = liaInstallCheck
	myMdmDebUbuntu14Mgr.MdmManager.GatewayPackageName = types.Ubuntu14GwPackageName
	myMdmDebUbuntu14Mgr.MdmManager.GatewayPackageDownload = state.ScaleIO.Ubuntu14.Gw
	myMdmDebUbuntu14Mgr.MdmManager.GatewayInstallCmd = "GATEWAY_ADMIN_PASSWORD=" + state.ScaleIO.AdminPassword + " dpkg -i {LocalGw}"
	myMdmDebUbuntu14Mgr.MdmManager.GatewayInstallCheck = gatewayInstallCheck

	//REX-Ray
	myMdmDebUbuntu14Mgr.MdmManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myMdmDebUbuntu14Mgr.MdmManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myMdmDebUbuntu14Mgr
}
