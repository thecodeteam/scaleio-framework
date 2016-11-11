package executor

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"

	config "github.com/codedellemc/scaleio-framework/scaleio-executor/config"
	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	scaleionodes "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/scaleionodes"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrFoundSelfFailed found myself failed
	ErrFoundSelfFailed = errors.New("Failed to locate self node")
)

func whichNode(cfg *config.Config, getstate common.RetrieveState) (common.IScaleioNode, error) {
	log.Infoln("WhichNode ENTER")

	log.Infoln("ScaleIO Executor Retrieve State from Scheduler")
	state := common.WaitForStableState(getstate)

	log.Infoln("Find Self Node")
	node := common.GetSelfNode(state, cfg.ExecutorID)
	if node == nil {
		log.Infoln("GetSelfNode Failed")
		log.Infoln("WhichNode LEAVE")
		return nil, ErrFoundSelfFailed
	}

	var sionode common.IScaleioNode

	switch node.Persona {
	case types.PersonaMdmPrimary:
		log.Infoln("Is Primary")
		sionode = scaleionodes.NewPri(state, cfg, getstate)
	case types.PersonaMdmSecondary:
		log.Infoln("Is Secondary")
		sionode = scaleionodes.NewSec(state, cfg, getstate)
	case types.PersonaTb:
		log.Infoln("Is TieBreaker")
		sionode = scaleionodes.NewTb(state, cfg, getstate)
	case types.PersonaNode:
		log.Infoln("Is DataNode")
		sionode = scaleionodes.NewData(state, cfg, getstate)
	}

	log.Infoln("WhichNode Succeeded")
	log.Infoln("WhichNode LEAVE")
	return sionode, nil
}

//RunExecutor starts the executor
func RunExecutor(cfg *config.Config, getstate common.RetrieveState) error {
	log.Infoln("RunExecutor ENTER")
	log.Infoln("executorID:", cfg.ExecutorID)

	node, err := whichNode(cfg, getstate)
	if err != nil {
		log.Errorln("Unable to find Self in node list")
		log.Infoln("RunExecutor LEAVE")
		return ErrFoundSelfFailed
	}

	for {
		node.UpdateScaleIOState()

		self := node.GetSelfNode()
		if self == nil {
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

		switch self.State {
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

		case types.StateAddResourcesToScaleIO:
			node.RunStateAddResourcesToScaleIO()

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
