package scaleionodes

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
	xplatformsys "github.com/dvonthenen/goxplatform/sys"

	config "github.com/codedellemc/scaleio-framework/scaleio-executor/config"
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
func NewTb(state *types.ScaleIOFramework, cfg *config.Config, getstate common.RetrieveState) *ScaleioTieBreakerMdmNode {
	myNode := &ScaleioTieBreakerMdmNode{}
	myNode.Config = cfg
	myNode.GetState = getstate
	myNode.RebootRequired = false

	var pkgmgr mgr.IMdmMgr
	switch xplatform.GetInstance().Sys.GetOsType() {
	case xplatformsys.OsRhel:
		log.Infoln("Is RHEL7")
		pkgmgr = rhel7.NewMdmRpmRhel7Mgr(state)
	case xplatformsys.OsUbuntu:
		log.Infoln("Is Ubuntu14")
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

	common.WaitForCleanPrereqsReboot(stbmn)

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
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := stbmn.Config.ParseIPFromRestURI()

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
func (stbmn *ScaleioTieBreakerMdmNode) RunStatePrerequisitesInstalled() {
	common.WaitForPrereqsFinish(stbmn)
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

	err = stbmn.UpdateDevices()
	if err != nil {
		log.Errorln("UpdateDevices Failed:", err)
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
	common.WaitForBaseFinish(stbmn)

	errState := stbmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInitializeCluster() {
	common.WaitForClusterInstallFinish(stbmn)
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
	stbmn.RebootRequired = stbmn.RebootRequired || reboot

	errState := stbmn.UpdateNodeState(types.StateAddResourcesToScaleIO)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateAddResourcesToScaleIO")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (stbmn *ScaleioTieBreakerMdmNode) RunStateInstallRexRay() {
	reboot, err := stbmn.PkgMgr.RexraySetup(stbmn.State, stbmn.Config.ExecutorID)
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
	stbmn.RebootRequired = stbmn.RebootRequired || reboot

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

	common.WaitForCleanInstallReboot(stbmn)

	//requires a reboot?
	if stbmn.RebootRequired {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", stbmn.RebootRequired)

		errState := stbmn.UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		if stbmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := stbmn.Config.ParseIPFromRestURI()

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

		errState := stbmn.UpdateNodeState(types.StateFinishInstall)
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
	node := stbmn.GetSelfNode()
	if !node.Imperative && !node.Advertised {
		err := stbmn.UpdateDevices()
		if err == nil {
			log.Infoln("UpdateDevices() Succcedeed. Devices advertised!")
		} else {
			log.Errorln("UpdateDevices() Failed. Err:", err)
		}
	}

	log.Debugln("In StateFinishInstall. Wait for", common.PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(common.PollForChangesInSeconds) * time.Second)

	//TODO eventual plan for MDM node behavior
	/*
		if clusterStatusBad then
			doClusterRemediate()
		else if upgrade then
			_ = waitForClusterUpgrade(spmn.UpdateScaleIOState())
			doUpgrade()
	*/
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (stbmn *ScaleioTieBreakerMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
