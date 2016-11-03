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

//ScaleioSecondaryMdmNode implementation for ScaleIO Secondary MDM Node
type ScaleioSecondaryMdmNode struct {
	common.ScaleioNode
	PkgMgr mgr.IMdmMgr
}

//NewSec generates a Secondary MDM Node object
func NewSec(state *types.ScaleIOFramework) *ScaleioSecondaryMdmNode {
	myNode := &ScaleioSecondaryMdmNode{}

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
func (ssmn *ScaleioSecondaryMdmNode) RunStateUnknown() {
	reboot, err := ssmn.PkgMgr.EnvironmentSetup(ssmn.State)
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

	common.WaitForCleanPrereqsReboot(ssmn)

	errState = ssmn.UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		if ssmn.State.Debug {
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
func (ssmn *ScaleioSecondaryMdmNode) RunStatePrerequisitesInstalled() {
	common.WaitForPrereqsFinish(ssmn)
	err := ssmn.PkgMgr.ManagementSetup(ssmn.State, true)
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

	err = ssmn.PkgMgr.NodeSetup(ssmn.State)
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
	common.WaitForBaseFinish(ssmn)

	errState := ssmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (ssmn *ScaleioSecondaryMdmNode) RunStateInitializeCluster() {
	common.WaitForClusterInstallFinish(ssmn)
	reboot, err := ssmn.PkgMgr.GatewaySetup(ssmn.State)
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
	common.WaitForClusterInitializeFinish(ssmn)
	reboot, err := ssmn.PkgMgr.RexraySetup(ssmn.State)
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

	err = ssmn.PkgMgr.SetupIsolator(ssmn.State)
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

	common.WaitForCleanInstallReboot(ssmn)

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

		if ssmn.State.Debug {
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

	//TODO temporary until libkv
	ssmn.LeaveMarkerFileForConfigured()

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
			err := ssmn.PkgMgr.AddSdsNodesToCluster(ssmn.State, true)
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
