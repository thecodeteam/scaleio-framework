package common

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	scaleionodes "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/scaleionodes"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	rebootCmdline = "shutdown -r now"
)

var (
	//ErrInitialStateFailed failed to get the initial state
	ErrInitialStateFailed = errors.New("Failed to get the initial cluster state")

	//ErrFoundSelfFailed found myself failed
	ErrFoundSelfFailed = errors.New("Failed to locate self node")

	//ErrFindNodeFailed failed to find MDM Pair
	ErrFindNodeFailed = errors.New("Failed to find the specific node")

	//ErrMdmPairFailed failed to find MDM Pair
	ErrMdmPairFailed = errors.New("Failed to Find MDM Pair")
)

func personaToString(persona int) string {
	switch persona {
	case types.PersonaMdmPrimary:
		return "primary"
	case types.PersonaMdmSecondary:
		return "secondary"
	case types.PersonaTb:
		return "tiebreaker"
	case types.PersonaNode:
		return "data"
	default:
		return "unknown"
	}
}

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

func getNodeType(state *types.ScaleIOFramework, nodeType int) (*types.ScaleIONode, error) {
	for _, node := range state.ScaleIO.Nodes {
		if node.Persona == nodeType {
			log.Infoln("Found MDM Node:", node.ExecutorID, "Persona:", node.Persona)
			return node, nil
		}
	}
	return nil, ErrFindNodeFailed
}

func getPrimaryMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaMdmPrimary)
}

func getSecondaryMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaMdmSecondary)
}

func getTiebreakerMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaTb)
}

func getMdmNodes(state *types.ScaleIOFramework) ([]types.ScaleIONode, error) {
	log.Infoln("getMdmNodes ENTER")

	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Infoln("Failed to find Primary MDM")
		log.Infoln("getMdmNodes LEAVE")
		return nil, ErrMdmPairFailed
	}

	sec, err := getSecondaryMdmNode(state)
	if err != nil {
		log.Infoln("Failed to find Secondary MDM")
		log.Infoln("getMdmNodes LEAVE")
		return nil, ErrMdmPairFailed
	}

	var col []types.ScaleIONode
	col = append(col, *pri)
	col = append(col, *sec)

	log.Infoln("Found MDM Pair")
	log.Infoln("getMdmNodes LEAVE")
	return col, nil
}

func createMdmPairString(state *types.ScaleIOFramework) (string, error) {
	//use pre-configured MDM nodes
	if state.ScaleIO.Preconfig.PreConfigEnabled {
		str := state.ScaleIO.Preconfig.PrimaryMdmAddress + "," +
			state.ScaleIO.Preconfig.SecondaryMdmAddress
		log.Debugln("Using Pre-Configured MDMs:", str)
		return str, nil
	}

	log.Debugln("Creating MDMs based on Configuration")

	nodes, err := getMdmNodes(state)
	if err != nil {
		return "", err
	}

	str := ""
	for _, node := range nodes {
		if len(str) > 0 {
			str += ","
		}
		str += node.IPAddress
	}

	log.Debugln("IP String:", str)
	return str, nil
}

func generateSdsName(node *types.ScaleIONode) string {
	str := "sds" + strconv.Itoa(node.Index)
	return str
}

func getGatewayAddress(state *types.ScaleIOFramework) (string, error) {
	if state.ScaleIO.Preconfig.PreConfigEnabled &&
		state.ScaleIO.Preconfig.GatewayAddress != "" {
		log.Debugln("Using Pre-Configured Gateway:",
			state.ScaleIO.Preconfig.GatewayAddress)
		return state.ScaleIO.Preconfig.GatewayAddress, nil
	} else if state.ScaleIO.Preconfig.PreConfigEnabled {
		log.Debugln("Pre-Configured Gateway Not Set:",
			state.ScaleIO.Preconfig.PrimaryMdmAddress)
		return state.ScaleIO.Preconfig.PrimaryMdmAddress, nil
	} else if state.ScaleIO.LbGateway != "" {
		log.Debugln("Using Load Balancing Gateway:",
			state.ScaleIO.LbGateway)
		return state.ScaleIO.LbGateway, nil
	}

	pri, err := getPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		return "", err
	}

	log.Debugln("Determined Gateway:", pri.IPAddress)
	return pri.IPAddress, nil
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
