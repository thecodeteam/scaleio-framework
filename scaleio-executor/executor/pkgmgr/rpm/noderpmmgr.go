package rpm

import (
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//NodeRpmMgr implementation for NodeRpmMgr
type NodeRpmMgr struct {
	*mgr.NodeManager
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
