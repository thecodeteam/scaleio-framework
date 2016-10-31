package scaleionodes

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
	xplatformsys "github.com/dvonthenen/goxplatform/sys"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	ubuntu14 "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/deb/ubuntu14"
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	rhel7 "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/rpm/rhel7"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioTieBreakerMdmNode implementation for ScaleIO TieBreaker MDM Node
type ScaleioTieBreakerMdmNode struct {
	common.ScaleioNode
	PkgMgr mgr.IMdmMgr
}

//NewTb generates a TieBreaker MDM Node object
func NewTb(state *types.ScaleIOFramework) *ScaleioTieBreakerMdmNode {
	myNode := &ScaleioTieBreakerMdmNode{}

	var pkgmgr mgr.IMdmMgr
	switch xplatform.GetInstance().Sys.GetOsType() {
	case xplatformsys.OsRhel:
		pkgmgr = rhel7.NewMdmRpmRhel7Mgr(state)
	case xplatformsys.OsUbuntu:
		pkgmgr = ubuntu14.NewMdmDebUbuntu14Mgr(state)
	}
	myNode.PkgMgr = pkgmgr

	return myNode
}

//RunStateUnknown default action for StateUnknown
func (stbmn *ScaleioTieBreakerMdmNode) RunStateUnknown() {
	reboot, err := stbmn.PkgMgr.EnvironmentSetup(stbmn.State)
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := stbmn.UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	stbmn.State = common.WaitForCleanPrereqsReboot(stbmn.GetState)

	errState = stbmn.UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		if stbmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

			rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
			if rebootErr != nil {
				log.Errorln("Install Kernel Failed:", rebootErr)
			}

			time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
		}
	} else {
		log.Infoln("No need to reboot while installing prerequisites")
	}
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (stbmn *ScaleioTieBreakerMdmNode) RunStatePrerequisitesInstalled() {
	stbmn.State = common.WaitForPrereqsFinish(stbmn.GetState)
	err := stbmn.PkgMgr.ManagementSetup(stbmn.State, false)
	if err != nil {
		log.Errorln("ManagementSetup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = stbmn.PkgMgr.NodeSetup(stbmn.State)
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := stbmn.UpdateNodeState(types.StateBasePackagedInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateBasePackagedInstalled")
	}
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (stbmn *ScaleioTieBreakerMdmNode) RunStateBasePackagedInstalled() {
	stbmn.State = common.WaitForBaseFinish(stbmn.GetState)

	errState := stbmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInitializeCluster() {
	stbmn.State = common.WaitForClusterInstallFinish(stbmn.GetState)
	reboot, err := stbmn.PkgMgr.GatewaySetup(stbmn.State)
	if err != nil {
		log.Errorln("GatewaySetup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}
	stbmn.RebootRequired = reboot

	errState := stbmn.UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInstallRexRay() {
	stbmn.State = common.WaitForClusterInitializeFinish(stbmn.GetState)
	reboot, err := stbmn.PkgMgr.RexraySetup(stbmn.State)
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = stbmn.PkgMgr.SetupIsolator(stbmn.State)
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := stbmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := stbmn.UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	stbmn.State = common.WaitForCleanInstallReboot(stbmn.GetState)

	//requires a reboot?
	if stbmn.RebootRequired || reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", stbmn.RebootRequired)
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

		errState = stbmn.UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		if stbmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
			if rebootErr != nil {
				log.Errorln("Install Kernel Failed:", rebootErr)
			}

			time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
		}
	} else {
		log.Infoln("No need to reboot while installing REX-Ray")

		errState = stbmn.UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (stbmn *ScaleioTieBreakerMdmNode) RunStateSystemReboot() {
	errState := stbmn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateFinishInstall default action for StateFinishInstall
func (stbmn *ScaleioTieBreakerMdmNode) RunStateFinishInstall() {
	log.Debugln("In StateFinishInstall. Wait for", common.PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(common.PollForChangesInSeconds) * time.Second)

	if stbmn.State.DemoMode {
		log.Infoln("DemoMode = TRUE. Leaving marker file for previously configured")
		stbmn.LeaveMarkerFileForConfigured()
	}

	//TODO eventual plan for MDM node behavior
	/*
		if clusterStatusBad then
			doClusterRemediate()
		else if upgrade then
			_ = waitForClusterUpgrade(spmn.UpdateScaleIOState())
			doUpgrade()
		else
			checkForNewDataNodesToAdd()
	*/

	//TODO replace this at some point with API calls instead of CLI
	pri, errPri := common.GetPrimaryMdmNode(stbmn.State)
	sec, errSec := common.GetSecondaryMdmNode(stbmn.State)

	if errPri != nil {
		log.Errorln("Unable to find the Primary MDM Node. Retry again later.")
	} else if errSec != nil {
		log.Errorln("Unable to find the Secondary MDM Node. Retry again later.")
	} else {
		if (pri.LastContact+common.OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() &&
			(sec.LastContact+common.OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() {
			//This is the checkForNewDataNodesToAdd(). Other functionality TBD.
			err := stbmn.PkgMgr.AddSdsNodesToCluster(stbmn.State, true)
			if err != nil {
				log.Errorln("Failed to add node to ScaleIO cluster:", err)
			}
		}
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
