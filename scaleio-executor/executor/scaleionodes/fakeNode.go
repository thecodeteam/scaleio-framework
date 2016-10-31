package scaleionodes

import (
	log "github.com/Sirupsen/logrus"

	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioFakeNode implementation for ScaleIO Fake Node
type ScaleioFakeNode struct {
	common.ScaleioNode
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
