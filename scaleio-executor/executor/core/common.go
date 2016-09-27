package core

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	rebootCmdline = "shutdown -r 1"
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

func leaveMarkerFileForConfigured(node *types.ScaleIONode) {
	err := os.MkdirAll("/etc/scaleio-framework", 0644)
	if err != nil {
		log.Errorln("Unable to mkdir:", err)
	}

	data := []byte(personaToString(node.Persona))
	err = ioutil.WriteFile("/etc/scaleio-framework/state", data, 0644)
	if err != nil {
		log.Errorln("Unable to write to marker file:", err)
	}
}

//WhichNode execute commands based on the persona
func WhichNode(executorID string, getstate retrievestate) error {
	log.Infoln("WhichNode ENTER")

	if nodePreviouslyConfigured() {
		log.Infoln("nodePreviouslyConfigured is TRUE. Launching FakeNode.")
		fakeNode(executorID, getstate)
		log.Infoln("WhichNode LEAVE")
		return nil
	}

	log.Infoln("ScaleIO Executor Retrieve State from Scheduler")
	state := waitForStableState(getstate)

	log.Infoln("Find Self Node")
	node := getSelfNode(executorID, state)
	if node == nil {
		log.Infoln("GetSelfNode Failed")
		log.Infoln("WhichNode LEAVE")
		return ErrFoundSelfFailed
	}

	switch node.Persona {
	case types.PersonaMdmPrimary:
		primaryMDM(executorID, getstate)
	case types.PersonaMdmSecondary:
		secondaryMDM(executorID, getstate)
	case types.PersonaTb:
		tieBreaker(executorID, getstate)
	case types.PersonaNode:
		dataNode(executorID, getstate)
	}

	log.Infoln("WhichNode LEAVE")
	return nil
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
