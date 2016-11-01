package rhel7

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//MdmRpmRhel7Mgr implementation for MdmRpmRhel7Mgr
type MdmRpmRhel7Mgr struct {
	*mgr.MdmManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (mdm *MdmRpmRhel7Mgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("EnvironmentSetup ENTER")

	aioErr := xplatform.GetInstance().Inst.IsInstalled("libaio")
	zipErr := xplatform.GetInstance().Inst.IsInstalled("zip")
	unzipErr := xplatform.GetInstance().Inst.IsInstalled("unzip")
	javaErr := xplatform.GetInstance().Inst.IsInstalled("java-1.8.0-openjdk")
	if aioErr != nil || zipErr != nil || unzipErr != nil || javaErr != nil {
		log.Infoln("Installing libaio1, zip, and java-1.8.0-openjdk")

		miscCmdline := "yum -y install zip unzip libaio java-1.8.0-openjdk"
		err := xplatform.GetInstance().Run.Command(miscCmdline, aiozipCheck, "")
		if err != nil {
			log.Errorln("Install Prerequisites Failed:", err)
			log.Infoln("EnvironmentSetup LEAVE")
			return false, err
		}
	} else {
		log.Infoln("libaio1, zip, and java-1.8.0-openjdk are already installed")
	}

	log.Infoln("EnvironmentSetup LEAVE")

	return false, nil
}

//NewMdmRpmRhel7Mgr generates a MdmRpmRhel7Mgr object
func NewMdmRpmRhel7Mgr(state *types.ScaleIOFramework) *MdmRpmRhel7Mgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmRpmRhel7Mgr := &MdmRpmRhel7Mgr{myMdmMgr}

	//ScaleIO node
	myMdmRpmRhel7Mgr.MdmManager.SdsPackageName = types.Rhel7SdsPackageName
	myMdmRpmRhel7Mgr.MdmManager.SdsPackageDownload = state.ScaleIO.Rhel7.Sds
	myMdmRpmRhel7Mgr.MdmManager.SdsInstallCmd = "rpm -Uvh {LocalSds}"
	myMdmRpmRhel7Mgr.MdmManager.SdsInstallCheck = sdsInstallCheck
	myMdmRpmRhel7Mgr.MdmManager.SdcPackageName = types.Rhel7SdcPackageName
	myMdmRpmRhel7Mgr.MdmManager.SdcPackageDownload = state.ScaleIO.Rhel7.Sdc
	myMdmRpmRhel7Mgr.MdmManager.SdcInstallCmd = "MDM_IP={MdmPair} rpm -Uvh {LocalSdc}"
	myMdmRpmRhel7Mgr.MdmManager.SdcInstallCheck = sdcInstallCheck
	myMdmRpmRhel7Mgr.MdmManager.MdmPackageName = types.Rhel7MdmPackageName
	myMdmRpmRhel7Mgr.MdmManager.MdmPackageDownload = state.ScaleIO.Rhel7.Mdm
	myMdmRpmRhel7Mgr.MdmManager.MdmInstallCmd = "MDM_ROLE_IS_MANAGER={PriOrSec} rpm -Uvh {LocalMdm}"
	myMdmRpmRhel7Mgr.MdmManager.MdmInstallCheck = mdmInstallCheck
	myMdmRpmRhel7Mgr.MdmManager.LiaPackageName = types.Rhel7LiaPackageName
	myMdmRpmRhel7Mgr.MdmManager.LiaPackageDownload = state.ScaleIO.Rhel7.Lia
	myMdmRpmRhel7Mgr.MdmManager.LiaInstallCmd = "TOKEN=" + state.ScaleIO.AdminPassword + " rpm -Uvh {LocalLia}"
	myMdmRpmRhel7Mgr.MdmManager.LiaInstallCheck = liaInstallCheck
	myMdmRpmRhel7Mgr.MdmManager.GatewayPackageName = types.Rhel7GwPackageName
	myMdmRpmRhel7Mgr.MdmManager.GatewayPackageDownload = state.ScaleIO.Rhel7.Gw
	myMdmRpmRhel7Mgr.MdmManager.GatewayInstallCmd = "GATEWAY_ADMIN_PASSWORD=" + state.ScaleIO.AdminPassword + " rpm -Uvh {LocalGw}"
	myMdmRpmRhel7Mgr.MdmManager.GatewayInstallCheck = gatewayInstallCheck

	//REX-Ray
	myMdmRpmRhel7Mgr.MdmManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myMdmRpmRhel7Mgr.MdmManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myMdmRpmRhel7Mgr
}
