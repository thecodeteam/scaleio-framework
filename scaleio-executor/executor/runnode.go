package executor

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	scaleionodes "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/scaleionodes"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrFoundSelfFailed found myself failed
	ErrFoundSelfFailed = errors.New("Failed to locate self node")
)

//TODO temporary until libkv
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
	log.Debugln("Node has not been previously been configured")
	return false
}

func whichNode(executorID string, getstate common.RetrieveState) (common.IScaleioNode, error) {
	log.Infoln("WhichNode ENTER")

	var sionode common.IScaleioNode

	if nodePreviouslyConfigured() {
		log.Infoln("nodePreviouslyConfigured is TRUE. Launching FakeNode.")
		sionode = scaleionodes.NewFake()
	} else {
		log.Infoln("ScaleIO Executor Retrieve State from Scheduler")
		state := common.WaitForStableState(getstate)

		log.Infoln("Find Self Node")
		node := common.GetSelfNode(executorID, state)
		if node == nil {
			log.Infoln("GetSelfNode Failed")
			log.Infoln("WhichNode LEAVE")
			return nil, ErrFoundSelfFailed
		}

		switch node.Persona {
		case types.PersonaMdmPrimary:
			log.Infoln("Is Primary")
			sionode = scaleionodes.NewPri(state)
		case types.PersonaMdmSecondary:
			log.Infoln("Is Secondary")
			sionode = scaleionodes.NewSec(state)
		case types.PersonaTb:
			log.Infoln("Is TieBreaker")
			sionode = scaleionodes.NewTb(state)
		case types.PersonaNode:
			log.Infoln("Is DataNode")
			sionode = scaleionodes.NewData(state)
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
func RunExecutor(executorID string, getstate common.RetrieveState) error {
	log.Infoln("RunExecutor ENTER")
	log.Infoln("executorID:", executorID)

	node, err := whichNode(executorID, getstate)
	if err != nil {
		log.Errorln("Unable to find Self in node list")
		log.Infoln("RunExecutor LEAVE")
		return ErrFoundSelfFailed
	}

	for {
		node.UpdateScaleIOState()

		if node.GetSelfNode() == nil {
			log.Errorln("Unable to find Self in node list")
			errState := node.UpdateNodeState(types.StateFatalInstall)
			if errState != nil {
				log.Errorln("Failed to signal state change:", errState)
			} else {
				log.Debugln("Signaled StateFatalInstall")
			}
			time.Sleep(time.Duration(common.PollAfterFatalInSeconds) * time.Second)
			continue
		}

		switch node.GetSelfNode().State {
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
