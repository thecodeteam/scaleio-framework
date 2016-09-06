package core

import (
	"time"

	log "github.com/Sirupsen/logrus"

	nodestate "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/node"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func tieBreaker(executorID string, getstate retrievestate) {
	log.Infoln("TieBreaker ENTER")

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
			state = waitForPrereqsFinish(getstate)
			err := managementSetup(state, false)
			if err != nil {
				log.Errorln("ManagementSetup Failed:", err)
				errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
					types.StateFatalInstall)
				if errState != nil {
					log.Errorln("Failed to signal state change:", errState)
				}
				continue
			}

			err = nodeSetup(state)
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
				types.StateBasePackagedInstalled)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

		case types.StateBasePackagedInstalled:
			state = waitForBaseFinish(getstate)

			errState := nodestate.UpdateNodeState(state.SchedulerAddress, node.ExecutorID,
				types.StateInitializeCluster)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			}

		case types.StateInitializeCluster:
			state = waitForClusterInstallFinish(getstate)
			err := gatewaySetup(state)
			if err != nil {
				log.Errorln("GatewaySetup Failed:", err)
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

		case types.StateInstallRexRay:
			state = waitForClusterInitializeFinish(getstate)
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

			if state.DemoMode {
				log.Infoln("DemoMode = TRUE. Leaving marker file for previously configured")
				leaveMarkerFileForConfigured(node)
			}

			//TODO eventual plan for MDM node behavior
			/*
				if clusterStatusBad then
					doClusterRemediate()
				else if upgrade then
					_ = waitForClusterUpgrade(getstate)
					doUpgrade()
				else
					checkForNewDataNodesToAdd()
			*/

			pri, errPri := getPrimaryMdmNode(state)
			sec, errSec := getSecondaryMdmNode(state)

			if errPri != nil {
				log.Errorln("Unable to find the Primary MDM Node. Retry again later.")
			} else if errSec != nil {
				log.Errorln("Unable to find the Secondary MDM Node. Retry again later.")
			} else {
				if (pri.LastContact+OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() &&
					(sec.LastContact+OfflineTimeForMdmNodesInSeconds) < time.Now().Unix() {
					//This is the checkForNewDataNodesToAdd(). Other functionality TBD.
					err := addSdsNodesToCluster(state, true)
					if err != nil {
						log.Errorln("Failed to add node to ScaleIO cluster:", err)
					}
				}
			}

		case types.StateUpgradeCluster:
			//TODO process the upgrade here

		case types.StateFatalInstall:
			log.Debugln("Node marked Fatal. Wait for", PollAfterFatalInSeconds, "seconds")
			time.Sleep(time.Duration(PollAfterFatalInSeconds) * time.Second)
		}
	}
}
