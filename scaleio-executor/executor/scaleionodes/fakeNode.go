package core

import (
	log "github.com/Sirupsen/logrus"

	basenode "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioFakeNode implementation for ScaleIO Fake Node
type ScaleioFakeNode struct {
	basenode.BaseScaleioNode
}

//NewFake generates a Fake Node object
func NewFake() *ScaleioFakeNode {
	myNode := &ScaleioFakeNode{}
	return myNode
}

//RunStateUnknown default action for StateUnknown
func (sfn *ScaleioFakeNode) RunStateUnknown() {
	errState := sfn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}
