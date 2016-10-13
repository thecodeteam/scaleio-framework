package scaleionodes

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	basenode "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioPrimaryMdmNode implementation for ScaleIO Primary MDM Node
type ScaleioPrimaryMdmNode struct {
	basenode.MdmScaleioNode
}

//NewPri generates a Primary MDM Node object
func NewPri() *ScaleioPrimaryMdmNode {
	myNode := &ScaleioPrimaryMdmNode{}
	return myNode
}

//RunStateUnknown default action for StateUnknown
func (spmn *ScaleioPrimaryMdmNode) RunStateUnknown() {
	reboot, err := spmn.EnvironmentSetup()
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	spmn.State = common.WaitForCleanPrereqsReboot(spmn.GetState)

	errState = spmn.UpdateNodeState(types.StatePrerequisitesInstalled)
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
func (spmn *ScaleioPrimaryMdmNode) RunStatePrerequisitesInstalled() {
	spmn.State = common.WaitForPrereqsFinish(spmn.GetState)
	err := spmn.ManagementSetup(true)
	if err != nil {
		log.Errorln("ManagementSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = spmn.NodeSetup()
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateBasePackagedInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateBasePackagedInstalled")
	}
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (spmn *ScaleioPrimaryMdmNode) RunStateBasePackagedInstalled() {
	spmn.State = common.WaitForBaseFinish(spmn.GetState)
	err := spmn.CreateCluster()
	if err != nil {
		log.Errorln("CreateCluster Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (spmn *ScaleioPrimaryMdmNode) RunStateInitializeCluster() {
	spmn.State = common.WaitForClusterInstallFinish(spmn.GetState)
	err := spmn.InitializeCluster()
	if err != nil {
		log.Errorln("InitializeCluster Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	reboot, err := spmn.GatewaySetup()
	if err != nil {
		log.Errorln("GatewaySetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}
	spmn.RebootRequired = reboot

	errState := spmn.UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (spmn *ScaleioPrimaryMdmNode) RunStateInstallRexRay() {
	spmn.State = common.WaitForClusterInitializeFinish(spmn.GetState)
	reboot, err := spmn.RexraySetup()
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = spmn.SetupIsolator()
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	spmn.State = common.WaitForCleanInstallReboot(spmn.GetState)

	//requires a reboot?
	if spmn.RebootRequired || reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", spmn.RebootRequired)
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

		errState = spmn.UpdateNodeState(types.StateSystemReboot)
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

		errState = spmn.UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (spmn *ScaleioPrimaryMdmNode) RunStateSystemReboot() {
	errState := spmn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateFinishInstall default action for StateFinishInstall
func (spmn *ScaleioPrimaryMdmNode) RunStateFinishInstall() {
	log.Debugln("In StateFinishInstall. Wait for", common.PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(common.PollForChangesInSeconds) * time.Second)

	if spmn.State.DemoMode {
		log.Infoln("DemoMode = TRUE. Leaving marker file for previously configured")
		spmn.LeaveMarkerFileForConfigured()
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

	//This is the checkForNewDataNodesToAdd(). Other functionality TBD.
	//TODO replace this at some point with API calls instead of CLI
	err := spmn.AddSdsNodesToCluster(true)
	if err != nil {
		log.Errorln("Failed to add node to ScaleIO cluster:", err)
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (spmn *ScaleioPrimaryMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
