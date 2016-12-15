package common

import (
	"errors"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrNodeNotFound The node was not found
	ErrNodeNotFound = errors.New("The node was not found")

	//ErrAttributeNotFound The attribute was not found
	ErrAttributeNotFound = errors.New("The attribute was not found")
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

//FindScaleIONodeByHostname Find ScaleIO node by Hostname
func FindScaleIONodeByHostname(nodes []*types.ScaleIONode, hostname string) *types.ScaleIONode {
	log.Debugln("FindScaleIONodeByHostname ENTER")
	log.Debugln("hostname:", hostname)

	if len(hostname) == 0 {
		log.Debugln("Hostname is empty. Return nil.")
		log.Debugln("FindScaleIONodeByHostname LEAVE")
		return nil
	}

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		log.Debugln(node.Hostname, "=", hostname, "?")
		if node.Hostname == hostname {
			log.Debugln("Node Found")
			log.Debugln("FindScaleIONodeByHostname LEAVE")
			return node
		}
	}

	log.Debugln("Node NOT Found")
	log.Debugln("FindScaleIONodeByHostname LEAVE")
	return nil
}

//FindScaleIONodeByExecutorID Get ScaleIO node by ExecutorID
func FindScaleIONodeByExecutorID(nodes []*types.ScaleIONode, executorID string) *types.ScaleIONode {
	log.Debugln("FindScaleIONodeByExecutorID ENTER")
	log.Debugln("executorID:", executorID)

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		log.Debugln(node.ExecutorID, "=", executorID, "?")
		if node.ExecutorID == executorID {
			log.Debugln("Node Found:", node.ExecutorID)
			log.Debugln("FindScaleIONodeByExecutorID LEAVE")
			return node
		}
	}

	log.Debugln("Node NOT Found")
	log.Debugln("FindScaleIONodeByExecutorID LEAVE")
	return nil
}

func getNodeType(state *types.ScaleIOFramework, nodeType int) (*types.ScaleIONode, error) {
	for _, node := range state.ScaleIO.Nodes {
		if node.Persona == nodeType {
			log.Infoln("Found MDM Node:", node.ExecutorID, "Persona:", node.Persona)
			return node, nil
		}
	}
	return nil, ErrNodeNotFound
}

//GetPrimaryMdmNode gets the Primary node
func GetPrimaryMdmNode(state *types.ScaleIOFramework) (*types.ScaleIONode, error) {
	return getNodeType(state, types.PersonaMdmPrimary)
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
