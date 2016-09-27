package core

import (
	"time"

	log "github.com/Sirupsen/logrus"

	nodestate "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/node"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func fakeNode(executorID string, getstate retrievestate) {
	log.Infoln("DataNode ENTER")

	for {
		state := waitForStableState(getstate)
		node := getSelfNode(executorID, state)
		if node == nil {
			log.Errorln("Unable to find Self in node list")
			errState := nodestate.UpdateNodeState(state.SchedulerAddress, executorID,
				types.StateFatalInstall)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			} else {
				log.Debugln("Signaled StateFatalInstall")
			}
			time.Sleep(time.Duration(PollAfterFatalInSeconds) * time.Second)
			continue
		}

		switch node.State {
		case types.StateUnknown:
			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateFinishInstall)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			} else {
				log.Debugln("Signaled StateFinishInstall")
			}

		case types.StatePrerequisitesInstalled:
			log.Debugln("In StatePrerequisitesInstalled. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateBasePackagedInstalled:
			log.Debugln("In StateBasePackagedInstalled. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateInitializeCluster:
			log.Debugln("In StateInitializeCluster. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateInstallRexRay:
			log.Debugln("In StateInstallRexRay. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateFinishInstall:
			log.Debugln("In StateFinishInstall. Wait for", PollForChangesInSeconds,
				"seconds for changes in the cluster.")
			time.Sleep(time.Duration(PollForChangesInSeconds) * time.Second)

		case types.StateUpgradeCluster:
			//TODO process the upgrade here

		case types.StateFatalInstall:
			log.Debugln("Node marked Fatal. Wait for", PollAfterFatalInSeconds, "seconds")
			time.Sleep(time.Duration(PollAfterFatalInSeconds) * time.Second)
		}
	}
}
