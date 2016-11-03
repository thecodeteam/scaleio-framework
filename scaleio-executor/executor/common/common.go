package common

import (
	"errors"
	"strconv"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	//RebootCmdline reboot now
	RebootCmdline = "shutdown -r now"

	//RebootCheck check for the reboot
	RebootCheck = "reboot in 1 minute"
)

var (
	//ErrInitialStateFailed failed to get the initial state
	ErrInitialStateFailed = errors.New("Failed to get the initial cluster state")

	//ErrFindNodeFailed failed to find MDM Pair
	ErrFindNodeFailed = errors.New("Failed to find the specific node")

	//ErrMdmPairFailed failed to find MDM Pair
	ErrMdmPairFailed = errors.New("Failed to Find MDM Pair")
)

//PersonaStringToID String -> PersonaID
func PersonaStringToID(persona string) int {
	switch persona {
	case "primary":
		return types.PersonaMdmPrimary
	case "secondary":
		return types.PersonaMdmSecondary
	case "tiebreaker":
		return types.PersonaTb
	case "data":
		return types.PersonaNode
	default:
		return types.PersonaUnknown
	}
}

//PersonaIDToString PersonaID -> String
func PersonaIDToString(persona int) string {
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

//GetSelfNode gets self
func GetSelfNode(executorID string, state *types.ScaleIOFramework) *types.ScaleIONode {
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

//GetPrimaryMdmNode gets the Primary node
func GetPrimaryMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaMdmPrimary)
}

//GetSecondaryMdmNode gets the Secondary node
func GetSecondaryMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaMdmSecondary)
}

//GetTiebreakerMdmNode gets the TB node
func GetTiebreakerMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaTb)
}

func getMdmNodes(state *types.ScaleIOFramework) ([]types.ScaleIONode, error) {
	log.Infoln("getMdmNodes ENTER")

	pri, err := GetPrimaryMdmNode(state)
	if err != nil {
		log.Infoln("Failed to find Primary MDM")
		log.Infoln("getMdmNodes LEAVE")
		return nil, ErrMdmPairFailed
	}

	sec, err := GetSecondaryMdmNode(state)
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

//CreateMdmPairString just like it sounds
func CreateMdmPairString(state *types.ScaleIOFramework) (string, error) {
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

//GenerateSdsName creates the SDS name for this given node
func GenerateSdsName(node *types.ScaleIONode) string {
	str := "sds" + strconv.Itoa(node.Index)
	return str
}

//GetGatewayAddress returns the ScaleIO gateway address
func GetGatewayAddress(state *types.ScaleIOFramework) (string, error) {
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

	pri, err := GetPrimaryMdmNode(state)
	if err != nil {
		log.Errorln("Unable to find the Primary MDM node")
		return "", err
	}

	log.Debugln("Determined Gateway:", pri.IPAddress)
	return pri.IPAddress, nil
}
