package scaleionodes

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioSecondaryMdmNode implementation for ScaleIO Secondary MDM Node
type ScaleioSecondaryMdmNode struct {
	common.ScaleioNode
	PkgMgr mgr.IMdmMgr
}

//NewSec generates a Secondary MDM Node object
func NewSec() *ScaleioSecondaryMdmNode {
	myNode := &ScaleioSecondaryMdmNode{}

	var pkgmgr mgr.IPkgMgr
	switch xplatform.GetInstance().Sys.GetOsType() {
	case xplatformsys.OsRhel:
		pkgmgr = rpmmgr.NewMdmRpmMgr()
	case xplatformsys.OsUbuntu:
		pkgmgr = debmgr.NewMdmDebMgr()
	}
	myNode.PkgMgr = pkgmgr

	return myNode
}

//RunStateUnknown default action for StateUnknown
func (ssmn *ScaleioSecondaryMdmNode) RunStateUnknown() {
	reboot, err := ssmn.EnvironmentSetup()
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := ssmn.UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	ssmn.State = common.WaitForCleanPrereqsReboot(ssmn.GetState)

	errState = ssmn.UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

		rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing prerequisites")
	}
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (ssmn *ScaleioSecondaryMdmNode) RunStatePrerequisitesInstalled() {
	ssmn.State = common.WaitForPrereqsFinish(ssmn.GetState)
	err := ssmn.ManagementSetup(true)
	if err != nil {
		log.Errorln("ManagementSetup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = ssmn.NodeSetup()
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := ssmn.UpdateNodeState(types.StateBasePackagedInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateBasePackagedInstalled")
	}
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (ssmn *ScaleioSecondaryMdmNode) RunStateBasePackagedInstalled() {
	ssmn.State = common.WaitForBaseFinish(ssmn.GetState)

	errState := ssmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (ssmn *ScaleioSecondaryMdmNode) RunStateInitializeCluster() {
	ssmn.State = common.WaitForClusterInstallFinish(ssmn.GetState)
	reboot, err := ssmn.GatewaySetup()
	if err != nil {
		log.Errorln("GatewaySetup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}
	ssmn.RebootRequired = reboot

	errState := ssmn.UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (ssmn *ScaleioSecondaryMdmNode) RunStateInstallRexRay() {
	ssmn.State = common.WaitForClusterInitializeFinish(ssmn.GetState)
	reboot, err := ssmn.RexraySetup()
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = ssmn.SetupIsolator()
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := ssmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := ssmn.UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	ssmn.State = common.WaitForCleanInstallReboot(ssmn.GetState)

	//requires a reboot?
	if ssmn.RebootRequired || reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", ssmn.RebootRequired)
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

		errState = ssmn.UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing REX-Ray")

		errState = ssmn.UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (ssmn *ScaleioSecondaryMdmNode) RunStateSystemReboot() {
	errState := ssmn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateFinishInstall default action for StateFinishInstall
func (ssmn *ScaleioSecondaryMdmNode) RunStateFinishInstall() {
	log.Debugln("In StateFinishInstall. Wait for", common.PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(common.PollForChangesInSeconds) * time.Second)

	if ssmn.State.DemoMode {
		log.Infoln("DemoMode = TRUE. Leaving marker file for previously configured")
		ssmn.LeaveMarkerFileForConfigured()
	}

	//TODO eventual plan for MDM node behavior
	/*
		if clusterStatusBad then
			doClusterRemediate()
		else if upgrade then
			_ = common.WaitForClusterUpgrade(spmn.UpdateScaleIOState())
			doUpgrade()
		else
			checkForNewDataNodesToAdd()
	*/

	//TODO replace this at some point with API calls instead of CLI
	pri, errPri := common.GetPrimaryMdmNode(ssmn.State)

	if errPri != nil {
		log.Errorln("Unable to find the Primary MDM Node. Retry again later.", errPri)
	} else {
		if (pri.LastContact + common.OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() {
			//This is the checkForNewDataNodesToAdd(). Other functionality TBD.
			err := ssmn.AddSdsNodesToCluster(true)
			if err != nil {
				log.Errorln("Failed to add node to ScaleIO cluster:", err)
			}
		}
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (ssmn *ScaleioSecondaryMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
