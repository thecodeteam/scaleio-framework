package rpm

import (
	log "github.com/Sirupsen/logrus"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//NodeRpmMgr implementation for NodeRpmMgr
type NodeRpmMgr struct {
	*mgr.NodeManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (mdm *NodeRpmMgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("EnvironmentSetup ENTER")

	//TODO

	log.Infoln("EnvironmentSetup LEAVE")

	return false, nil
}

//NewNodeRpmMgr generates a NodeRpmMgr object
func NewNodeRpmMgr(state *types.ScaleIOFramework) NodeRpmMgr {
	myNodeMgr := &mgr.NodeManager{}
	myNodeRpmMgr := NodeRpmMgr{myNodeMgr}

	myNodeRpmMgr.NodeManager.RexrayInstallCheck = rexrayInstallCheck
	myNodeRpmMgr.NodeManager.DvdcliInstallCheck = dvdcliInstallCheck

	//TODO pending

	return myNodeRpmMgr
}
