package scheduler

import (
	"strconv"

	log "github.com/Sirupsen/logrus"

	mesos "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/v1"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func findScaleIONodeByAgentID(nodes types.ScaleIONodes, agentID string) *types.ScaleIONode {
	log.Debugln("findScaleIONodeByAgentID ENTER")
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		if node.AgentID == agentID {
			log.Debugln("Node Found:", node.AgentID)
			log.Debugln("findScaleIONodeByAgentID LEAVE")
			return node
		}
	}
	log.Debugln("Node NOT Found")
	log.Debugln("findScaleIONodeByAgentID LEAVE")
	return nil
}

//IsNodeAnMDMNode returns true is node is an MDM node
func IsNodeAnMDMNode(node *types.ScaleIONode) bool {
	isMDM := node.Persona == types.PersonaMdmPrimary ||
		node.Persona == types.PersonaMdmSecondary ||
		node.Persona == types.PersonaTb
	if isMDM {
		log.Debugln("Node is an MDM Node")
	} else {
		log.Debugln("Node is an Data Node")
	}
	return isMDM
}

func prepareScaleIONode(offer *mesos.Offer, persona int, ID int) *types.ScaleIONode {
	node := &types.ScaleIONode{
		AgentID:     offer.GetAgentId().GetValue(),
		TaskID:      "scaleio" + strconv.Itoa(ID),
		ExecutorID:  "executor-scaleio" + strconv.Itoa(ID),
		OfferID:     offer.GetId().GetValue(),
		IPAddress:   offer.GetUrl().GetAddress().GetIp(),
		Index:       ID,
		Persona:     persona,
		State:       types.StateUnknown,
		InCluster:   false,
		LastContact: 0,
	}

	return node
}
