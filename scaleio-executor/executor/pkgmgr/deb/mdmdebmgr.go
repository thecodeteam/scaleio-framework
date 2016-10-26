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
	mdmInstallCheck          = "mdm start/running"
	sdsInstallCheck          = "sds start/running"
	sdcInstallCheck          = "Success configuring module"
	clusterConfigCheck       = "Mode: 3_node"
	createClusterCheck       = "Successfully created the MDM Cluster"
	loggedInCheck            = "Logged in"
	setPasswordCheck         = "Password changed successfully"
	addMdmToClusterCheck     = "Successfully added a standby MDM"
	changeClusterModeCheck   = "Successfully switched the cluster mode"
	clusterNotInitialedCheck = "Query-all-SDS returned 0 SDS nodes"
	liaInstallCheck          = "lia start/running"
	liaRestartCheck          = liaInstallCheck
	gatewayInstallCheck      = "The EMC ScaleIO Gateway is running"
	gatewayRestartCheck      = "scaleio-gateway start/running"
	clusterRenameCheck       = "Successfully renamed system to"
	addProtectionDomainCheck = "Successfully created protection domain"
	addStoragePoolCheck      = "Successfully created a storage pool"
	addSdsCheck              = "Successfully created SDS"
	addVolumeCheck           = "Successfully created volume of size"

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
func NewMdmDebMgr() MdmDebMgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmDebMgr := MdmDebMgr{myMdmMgr}

	myMdmDebMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myMdmDebMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck

	/*
		mdmCmdline := "MDM_ROLE_IS_MANAGER=" + strPriOrSec + " dpkg -i " + localMdm
		sdsCmdline := "dpkg -i " + localSds
		sdcCmdline := "MDM_IP=" + mdmPair + " dpkg -i " + localSdc
		liaCmdline := "TOKEN=" + state.ScaleIO.AdminPassword + " dpkg -i " + localLia
		gwCmdline := "GATEWAY_ADMIN_PASSWORD=" + state.ScaleIO.AdminPassword + " dpkg -i " + localGw
	*/

	return myMdmDebMgr
}
