package core

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	//DelayBetweenCommandsInSeconds just like it says
	DelayBetweenCommandsInSeconds = 5

	//DelayOnVolumeCreateInSeconds just like it says
	DelayOnVolumeCreateInSeconds = 20

	//DelayForRebootInSeconds the amount of time to wait for the reboot
	WaitForRebootInSeconds = 5

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

type retrievestate func() (*types.ScaleIOFramework, error)

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

func waitForStableState(getstate retrievestate) *types.ScaleIOFramework {
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

func waitForState(getstate retrievestate, nodeState int, allNodes bool) *types.ScaleIOFramework {
	var err error
	var state *types.ScaleIOFramework
	for {
		state, err = getstate()
		if err != nil {
			log.Debugln("Unable to getstate. Waiting for", PollStatusInSeconds, "seconds")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
			continue
		}
		if waitForRunState(state, nodeState, allNodes) {
			log.Debugln("Achieve state", nodeState, "among management nodes")
			break
		}
		log.Debugln("Waiting for", PollStatusInSeconds, "seconds")
		time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
	}
	return state
}

func waitForPrereqsFinish(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StatePrerequisitesInstalled, false)
}

func waitForCleanPrereqsReboot(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StateCleanPrereqsReboot, true)
}

func waitForBaseFinish(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StateBasePackagedInstalled, false)
}

func waitForClusterInstallFinish(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StateInitializeCluster, false)
}

func waitForClusterInitializeFinish(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StateInstallRexRay, false)
}

func waitForCleanInstallReboot(getstate retrievestate) *types.ScaleIOFramework {
	return waitForState(getstate, types.StateCleanInstallReboot, true)
}
