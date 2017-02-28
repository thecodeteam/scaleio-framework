package common

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	//DelayBetweenCommandsInSeconds just like it says
	DelayBetweenCommandsInSeconds = 5

	//DelayIfInstalledInSeconds just like it says
	DelayIfInstalledInSeconds = 5

	//DelayOnVolumeCreateInSeconds just like it says
	DelayOnVolumeCreateInSeconds = 20

	//DelayForRebootInSeconds the amount of time to wait for the reboot
	DelayForRebootInSeconds = 5

	//WaitForRebootInSeconds the amount of time to wait for the reboot
	WaitForRebootInSeconds = 120

	//OfflineTimeForMdmNodesInSeconds is the max time an MDM node can be offline
	OfflineTimeForMdmNodesInSeconds = 180

	//PollStatusInSeconds the amount of time to wait before updating state
	PollStatusInSeconds = 5

	//PollAfterFatalInSeconds the amount of time to wait before checking
	PollAfterFatalInSeconds = 3600

	//PollForChangesInSeconds cluster working. check for changes.
	PollForChangesInSeconds = 30
)

//RetrieveState is a call back to retrieve an update of the state
type RetrieveState func() (*types.ScaleIOFramework, error)

func waitForRunState(state *types.ScaleIOFramework, runState int, allNodes bool) bool {
	for _, node := range state.ScaleIO.Nodes {
		switch node.Persona {
		case types.PersonaMdmPrimary:
			if node.State < runState {
				return false
			}
		case types.PersonaMdmSecondary:
			if node.State < runState {
				return false
			}
		case types.PersonaTb:
			if node.State < runState {
				return false
			}
		case types.PersonaNode:
			if allNodes && node.State < runState {
				return false
			}
		}
	}

	return true
}

//WaitForStableState waits until a state can successfully be retrieved
func WaitForStableState(getstate RetrieveState) *types.ScaleIOFramework {
	var err error
	var state *types.ScaleIOFramework
	for {
		state, err = getstate()
		if err == nil {
			log.Debugln("waitForState BREAK")
			break
		}
		log.Debugln("Waiting for", PollStatusInSeconds, "seconds")
		time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
	}
	return state
}

func waitForState(sio IScaleioNode, nodeState int, allNodes bool) *types.ScaleIOFramework {
	var state *types.ScaleIOFramework
	for {
		state = sio.UpdateScaleIOState()
		if waitForRunState(state, nodeState, allNodes) {
			log.Debugln("Achieve state", nodeState, "among nodes")
			break
		}
		log.Debugln("Waiting for", PollStatusInSeconds, "seconds")
		time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
	}
	return state
}

//WaitForPrereqsFinish waits until all prereqs have been installed
func WaitForPrereqsFinish(sio IScaleioNode) {
	waitForState(sio, types.StatePrerequisitesInstalled, false)
}

//WaitForCleanPrereqsReboot waits until all systems are ready to reboot
//after prereq install
func WaitForCleanPrereqsReboot(sio IScaleioNode) {
	waitForState(sio, types.StateCleanPrereqsReboot, true)
}

//WaitForBaseFinish waits until base ScaleIO components have been installed
func WaitForBaseFinish(sio IScaleioNode) {
	waitForState(sio, types.StateBasePackagedInstalled, false)
}

//WaitForClusterInstallFinish waits for the cluster to be created
func WaitForClusterInstallFinish(sio IScaleioNode) {
	waitForState(sio, types.StateInitializeCluster, false)
}

//WaitForClusterInitializeFinish wait until the cluster has been initialized
func WaitForClusterInitializeFinish(sio IScaleioNode) {
	waitForState(sio, types.StateInstallRexRay, false)
}

//WaitForCleanInstallReboot waits for reboot of nodes after install
func WaitForCleanInstallReboot(sio IScaleioNode) {
	waitForState(sio, types.StateCleanInstallReboot, true)
}
