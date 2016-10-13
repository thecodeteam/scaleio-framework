package executor

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	scaleionodes "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/scaleionodes"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrFoundSelfFailed found myself failed
	ErrFoundSelfFailed = errors.New("Failed to locate self node")
)

func nodePreviouslyConfigured() bool {
	if _, err := os.Stat("/etc/scaleio-framework/state"); err == nil {
		b, errFile := ioutil.ReadFile("/etc/scaleio-framework/state")
		if errFile != nil {
			log.Errorln("Unable to open file:", errFile)
		} else {
			log.Infoln("Node is configured as", string(b), "MDM node")
		}
		return true
	}
	return false
}

func getSelfNode(executorID string, state *types.ScaleIOFramework) *types.ScaleIONode {
	log.Infoln("getSelfNode ENTER")
	for _, node := range state.ScaleIO.Nodes {
		if executorID == node.ExecutorID {
			log.Infoln("getSelfNode Found:", node.ExecutorID)
			log.Infoln("getSelfNode LEAVE")
			return node
		}
	}
	log.Infoln("getSelfNode NOT FOUND")
	log.Infoln("getSelfNode LEAVE")
	return nil
}

func whichNode(executorID string, getstate retrievestate) (*IScaleioNode, error) {
	log.Infoln("WhichNode ENTER")

	var sionode IScaleioNode

	if nodePreviouslyConfigured() {
		log.Infoln("nodePreviouslyConfigured is TRUE. Launching FakeNode.")
		node = scaleionodes.NewFake()
	} else {
		log.Infoln("ScaleIO Executor Retrieve State from Scheduler")
		state := WaitForStableState(getstate)

		log.Infoln("Find Self Node")
		node := getSelfNode(executorID, state)
		if node == nil {
			log.Infoln("GetSelfNode Failed")
			log.Infoln("WhichNode LEAVE")
			return nil, ErrFoundSelfFailed
		}

		switch node.Persona {
		case types.PersonaMdmPrimary:
			log.Infoln("Is Primary")
			sionode = scaleionodes.NewPri()
		case types.PersonaMdmSecondary:
			log.Infoln("Is Secondary")
			sionode = scaleionodes.NewSec()
		case types.PersonaTb:
			log.Infoln("Is TieBreaker")
			sionode = scaleionodes.NewTb()
		case types.PersonaNode:
			log.Infoln("Is DataNode")
			sionode = scaleionodes.NewData()
		}

		sionode.SetExecutorID(executorID)
		sionode.SetRetrieveState(getstate)
		sionode.UpdateScaleIOState()
	}

	log.Infoln("WhichNode Succeeded")
	log.Infoln("WhichNode LEAVE")
	return sionode, nil
}

//RunExecutor starts the executor
func RunExecutor(executorID string, getstate retrievestate) error {
	log.Infoln("RunExecutor ENTER")
	log.Infoln("executorID:", executorID)

	for {
		node, err := whichNode(executorID, getstate)
		if err != nil {
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
			node.RunStateUnknown()

		case types.StateCleanPrereqsReboot:
			node.RunStateCleanPrereqsReboot()

		case types.StatePrerequisitesInstalled:
			node.RunStatePrerequisitesInstalled()

		case types.StateBasePackagedInstalled:
			node.RunStateBasePackagedInstalled()

		case types.StateInitializeCluster:
			node.RunStateInitializeCluster()

		case types.StateInstallRexRay:
			node.RunStateInstallRexRay()

		case types.StateCleanInstallReboot:
			node.RunStateCleanInstallReboot()

		case types.StateSystemReboot:
			node.RunStateSystemReboot()

		case types.StateFinishInstall:
			node.RunStateFinishInstall()

		case types.StateUpgradeCluster:
			node.RunStateUpgradeCluster()

		case types.StateFatalInstall:
			node.RunStateFatalInstall()
		}
	}

	log.Infoln("RunExecutor Succeeded")
	log.Infoln("RunExecutor LEAVE")
	return nil
}
