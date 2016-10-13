package core

import (
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
	errState := UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}
