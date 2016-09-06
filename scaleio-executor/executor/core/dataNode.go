package core

import (
	"time"

	log "github.com/Sirupsen/logrus"

	nodestate "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/node"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func dataNode(executorID string, getstate retrievestate) {
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
			}
			time.Sleep(time.Duration(PollAfterFatalInSeconds) * time.Second)
			continue
		}

		switch node.State {
		case types.StateUnknown:
			err := environmentSetup(state)
			if err != nil && err != ErrRebootRequired {
				log.Errorln("EnvironmentSetup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StatePrerequisitesInstalled)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

			//wait for the reboot
			if err == ErrRebootRequired {
				time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
			}

		case types.StatePrerequisitesInstalled:
			err := nodeSetup(state)
			if err != nil {
				log.Errorln("NodeSetup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateInstallRexRay)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

		case types.StateBasePackagedInstalled:
			log.Debugln("In StateBasePackagedInstalled. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateInitializeCluster:
			log.Debugln("In StateInitializeCluster. Do nothing.")
			time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)

		case types.StateInstallRexRay:
			if state.ScaleIO.Preconfig.PreConfigEnabled {
				log.Debugln("Pre-Config is enabled skipping wait for Cluster Initialization")
			} else {
				//we need to wait because without the gateway, the rexray service restart
				//will fail
				state = waitForClusterInitializeFinish(getstate)
			}

			err := rexraySetup(state)
			if err != nil {
				log.Errorln("REX-Ray setup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			err = setupIsolator(state)
			if err != nil {
				log.Errorln("Mesos Isolator setup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateFinishInstall)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

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
