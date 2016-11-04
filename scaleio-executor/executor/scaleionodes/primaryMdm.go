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

//ScaleioPrimaryMdmNode implementation for ScaleIO Primary MDM Node
type ScaleioPrimaryMdmNode struct {
	common.ScaleioNode
	PkgMgr mgr.IMdmMgr
}

//NewPri generates a Primary MDM Node object
func NewPri(state *types.ScaleIOFramework) *ScaleioPrimaryMdmNode {
	myNode := &ScaleioPrimaryMdmNode{}

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
func (spmn *ScaleioPrimaryMdmNode) RunStateUnknown() {
	reboot, err := spmn.PkgMgr.EnvironmentSetup(spmn.State)
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

	common.WaitForCleanPrereqsReboot(spmn)

	errState = spmn.UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		if spmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := spmn.Config.ParseIPFromRestURI()

			if err1 == nil && err2 == nil && ip1 == ip2 {
				log.Infoln("Delay reboot host running the Scheduler")
				time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)
			}

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
func (spmn *ScaleioPrimaryMdmNode) RunStatePrerequisitesInstalled() {
	common.WaitForPrereqsFinish(spmn)
	err := spmn.PkgMgr.ManagementSetup(spmn.State, true)
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

	err = spmn.PkgMgr.NodeSetup(spmn.State)
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
	common.WaitForBaseFinish(spmn)
	err := spmn.PkgMgr.CreateCluster(spmn.State)
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
	common.WaitForClusterInstallFinish(spmn)
	err := spmn.PkgMgr.InitializeCluster(spmn.State)
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

	reboot, err := spmn.PkgMgr.GatewaySetup(spmn.State)
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
	common.WaitForClusterInitializeFinish(spmn)
	reboot, err := spmn.PkgMgr.RexraySetup(spmn.State)
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

	err = spmn.PkgMgr.SetupIsolator(spmn.State)
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

	common.WaitForCleanInstallReboot(spmn)

	//requires a reboot?
	if spmn.RebootRequired || reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", spmn.RebootRequired)
		log.Debugln("reboot:", reboot)

		errState = spmn.UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		if spmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := spmn.Config.ParseIPFromRestURI()

			if err1 == nil && err2 == nil && ip1 == ip2 {
				log.Infoln("Delay reboot host running the Scheduler")
				time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)
			}

			rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
			if rebootErr != nil {
				log.Errorln("Install Kernel Failed:", rebootErr)
			}

			time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
		}
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

	//TODO temporary until libkv
	spmn.LeaveMarkerFileForConfigured()

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
	err := spmn.PkgMgr.AddSdsNodesToCluster(spmn.State, true)
	if err != nil {
		log.Errorln("Failed to add node to ScaleIO cluster:", err)
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (spmn *ScaleioPrimaryMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
