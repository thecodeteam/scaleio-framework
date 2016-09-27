package core

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dvonthenen/scaleio-executor/native/exec"

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
			reboot, err := environmentSetup(state)
			if err != nil {
				log.Errorln("EnvironmentSetup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateCleanPrereqsReboot)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

			state = waitForCleanPrereqsReboot(getstate)

			errState = nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StatePrerequisitesInstalled)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

			//requires a reboot?
			if reboot {
				log.Infoln("Reboot required before StatePrerequisitesInstalled!")

				rebootErr := exec.RunCommand(rebootCmdline, rebootCheck, "")
				if rebootErr != nil {
					log.Errorln("Install Kernel Failed:", rebootErr)
				}

				time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
			} else {
				log.Infoln("No need to reboot while installing prerequisites")
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

			reboot, err := rexraySetup(state)
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
				types.StateCleanInstallReboot)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

			state = waitForCleanInstallReboot(getstate)

			errState = nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateFinishInstall)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

			//requires a reboot?
			if reboot {
				log.Infoln("Reboot required before StateFinishInstall!")
				log.Debugln("reboot:", reboot)

				rebootErr := exec.RunCommand(rebootCmdline, rebootCheck, "")
				if rebootErr != nil {
					log.Errorln("Install Kernel Failed:", rebootErr)
				}

				time.Sleep(time.Duration(WaitForRebootInSeconds) * time.Second)
			} else {
				log.Infoln("No need to reboot while installing REX-Ray")
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
