package scaleionodes

import (
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioDataNode implementation for ScaleIO Fake Node
type ScaleioDataNode struct {
	common.ScaleioNode
	pkgMgr mgr.INodeMgr
}

//NewData generates a Data Node object
func NewData() *ScaleioDataNode {
	myNode := &ScaleioDataNode{}
	return myNode
}

//RunStateUnknown default action for StateUnknown
func (sdn *ScaleioDataNode) RunStateUnknown() {
	reboot, err := sdn.EnvironmentSetup()
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := sdn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := sdn.UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	sdn.State = common.WaitForCleanPrereqsReboot(sdn.GetState)

	errState = sdn.UpdateNodeState(types.StatePrerequisitesInstalled)
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
func (sdn *ScaleioDataNode) RunStatePrerequisitesInstalled() {
	err := sdn.NodeSetup()
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := sdn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := sdn.UpdateNodeState(types.StateInstallRexRay)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInstallRexRay")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (sdn *ScaleioDataNode) RunStateInstallRexRay() {
	if sdn.State.ScaleIO.Preconfig.PreConfigEnabled {
		log.Debugln("Pre-Config is enabled skipping wait for Cluster Initialization")
	} else {
		//we need to wait because without the gateway, the rexray service restart
		//will fail
		sdn.State = common.WaitForClusterInitializeFinish(sdn.GetState)
	}

	reboot, err := sdn.RexraySetup()
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := sdn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = sdn.SetupIsolator()
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := sdn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := sdn.UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	sdn.State = common.WaitForCleanInstallReboot(sdn.GetState)

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("reboot:", reboot)

		time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)

		errState = sdn.UpdateNodeState(types.StateSystemReboot)
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

		errState = sdn.UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (sdn *ScaleioDataNode) RunStateSystemReboot() {
	errState := sdn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (sdn *ScaleioDataNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}
