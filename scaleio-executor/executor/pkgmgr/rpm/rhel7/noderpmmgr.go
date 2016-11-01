package rhel7

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//NodeRpmRhel7Mgr implementation for NodeRpmRhel7Mgr
type NodeRpmRhel7Mgr struct {
	*mgr.NodeManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (npm *NodeRpmRhel7Mgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
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

//NewNodeRpmRhel7Mgr generates a NodeRpmRhel7Mgr object
func NewNodeRpmRhel7Mgr(state *types.ScaleIOFramework) *NodeRpmRhel7Mgr {
	myNodeMgr := &mgr.NodeManager{}
	myNodeRpmRhel7Mgr := &NodeRpmRhel7Mgr{myNodeMgr}

	//ScaleIO node
	myNodeRpmRhel7Mgr.NodeManager.SdsPackageName = types.Rhel7SdsPackageName
	myNodeRpmRhel7Mgr.NodeManager.SdsPackageDownload = state.ScaleIO.Rhel7.Sds
	myNodeRpmRhel7Mgr.NodeManager.SdsInstallCmd = "rpm -Uvh {LocalSds}"
	myNodeRpmRhel7Mgr.NodeManager.SdsInstallCheck = sdsInstallCheck
	myNodeRpmRhel7Mgr.NodeManager.SdcPackageName = types.Rhel7SdcPackageName
	myNodeRpmRhel7Mgr.NodeManager.SdcPackageDownload = state.ScaleIO.Rhel7.Sdc
	myNodeRpmRhel7Mgr.NodeManager.SdcInstallCmd = "MDM_IP={MdmPair} rpm -Uvh {LocalSdc}"
	myNodeRpmRhel7Mgr.NodeManager.SdcInstallCheck = sdcInstallCheck

	//REX-Ray
	myNodeRpmRhel7Mgr.NodeManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myNodeRpmRhel7Mgr.NodeManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myNodeRpmRhel7Mgr
}
