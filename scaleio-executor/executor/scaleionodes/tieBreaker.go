package core

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	basenode "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioTieBreakerMdmNode implementation for ScaleIO TieBreaker MDM Node
type ScaleioTieBreakerMdmNode struct {
	basenode.MdmScaleioNode
}

//NewTb generates a TieBreaker MDM Node object
func NewTb() *ScaleioTieBreakerMdmNode {
	myNode := &ScaleioTieBreakerMdmNode{}
	return myNode
}

//RunStateUnknown default action for StateUnknown
func (stbmn *ScaleioTieBreakerMdmNode) RunStateUnknown(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	reboot, err := environmentSetup(state)
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	state = waitForCleanPrereqsReboot(spmn.UpdateScaleIOState())

	errState = UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		time.Sleep(time.Duration(DelayForRebootInSeconds) * time.Second)

		rebootErr := xplatform.GetInstance().Run.Command(rebootCmdline, rebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing prerequisites")
	}
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (stbmn *ScaleioTieBreakerMdmNode) RunStatePrerequisitesInstalled(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	state = waitForPrereqsFinish(spmn.UpdateScaleIOState())
	err := managementSetup(state, false)
	if err != nil {
		log.Errorln("ManagementSetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}

	err = nodeSetup(state)
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}

	errState := UpdateNodeState(types.StateBasePackagedInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateBasePackagedInstalled")
	}
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (stbmn *ScaleioTieBreakerMdmNode) RunStateBasePackagedInstalled(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	state = waitForBaseFinish(spmn.UpdateScaleIOState())

	errState := UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInitializeCluster(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	state = waitForClusterInstallFinish(spmn.UpdateScaleIOState())
	reboot, err := gatewaySetup(state)
	if err != nil {
		log.Errorln("GatewaySetup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		continue
	}
	stbmn.RebootRequired = reboot

	errState := UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInstallRexRay(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	state = waitForClusterInitializeFinish(spmn.UpdateScaleIOState())
	reboot, err := rexraySetup(state)
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = setupIsolator(state)
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	state = waitForCleanInstallReboot(spmn.UpdateScaleIOState())

	//requires a reboot?
	if rebootRequired || reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", rebootRequired)
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(DelayForRebootInSeconds) * time.Second)

		errState = UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		rebootErr := xplatform.GetInstance().Run.Command(rebootCmdline, rebootCheck, "")
		if rebootErr != nil {
			log.Errorln("Install Kernel Failed:", rebootErr)
		}

		time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
	} else {
		log.Infoln("No need to reboot while installing REX-Ray")

		errState = UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (stbmn *ScaleioTieBreakerMdmNode) RunStateSystemReboot(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	errState := UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateFinishInstall default action for StateFinishInstall
func (stbmn *ScaleioTieBreakerMdmNode) RunStateFinishInstall(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	log.Debugln("In StateFinishInstall. Wait for", PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(PollForChangesInSeconds) * time.Second)

	if state.DemoMode {
		log.Infoln("DemoMode = TRUE. Leaving marker file for previously configured")
		LeaveMarkerFileForConfigured(node)
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

	pri, errPri := getPrimaryMdmNode(state)
	sec, errSec := getSecondaryMdmNode(state)

	if errPri != nil {
		log.Errorln("Unable to find the Primary MDM Node. Retry again later.")
	} else if errSec != nil {
		log.Errorln("Unable to find the Secondary MDM Node. Retry again later.")
	} else {
		if (pri.LastContact+OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() &&
			(sec.LastContact+OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() {
			//This is the checkForNewDataNodesToAdd(). Other functionality TBD.
			err := addSdsNodesToCluster(state, true)
			if err != nil {
				log.Errorln("Failed to add node to ScaleIO cluster:", err)
			}
		}
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateUpgradeCluster(state *types.ScaleIOFramework, node *types.ScaleIONode) {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
