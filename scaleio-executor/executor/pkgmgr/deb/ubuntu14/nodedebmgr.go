package ubuntu14

import (
	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//NodeDebUbuntu14Mgr implementation for NodeDebUbuntu14Mgr
type NodeDebUbuntu14Mgr struct {
	*mgr.NodeManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (dpm *NodeDebUbuntu14Mgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
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

//NewNodeDebUbuntu14Mgr generates a NodeDebMgr object
func NewNodeDebUbuntu14Mgr(state *types.ScaleIOFramework) *NodeDebUbuntu14Mgr {
	myNodeMgr := &mgr.NodeManager{}
	myNodeDebUbuntu14Mgr := &NodeDebUbuntu14Mgr{myNodeMgr}

	//ScaleIO node
	myNodeDebUbuntu14Mgr.NodeManager.SdsPackageName = types.Ubuntu14SdsPackageName
	myNodeDebUbuntu14Mgr.NodeManager.SdsPackageDownload = state.ScaleIO.Ubuntu14.Sds
	myNodeDebUbuntu14Mgr.NodeManager.SdsInstallCmd = "dpkg -i {LocalSds}"
	myNodeDebUbuntu14Mgr.NodeManager.SdsInstallCheck = sdsInstallCheck
	myNodeDebUbuntu14Mgr.NodeManager.SdcPackageName = types.Ubuntu14SdcPackageName
	myNodeDebUbuntu14Mgr.NodeManager.SdcPackageDownload = state.ScaleIO.Ubuntu14.Sdc
	myNodeDebUbuntu14Mgr.NodeManager.SdcInstallCmd = "MDM_IP={MdmPair} dpkg -i {LocalSdc}"
	myNodeDebUbuntu14Mgr.NodeManager.SdcInstallCheck = sdcInstallCheck

	//REX-Ray
	myNodeDebUbuntu14Mgr.NodeManager.RexrayInstallCheck = rexrayInstallCheck

	//Isolator
	myNodeDebUbuntu14Mgr.NodeManager.DvdcliInstallCheck = dvdcliInstallCheck

	return myNodeDebUbuntu14Mgr
}
