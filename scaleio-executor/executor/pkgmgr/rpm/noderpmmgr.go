package rpm

import mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//NodeRpmMgr implementation for NodeRpmMgr
type NodeRpmMgr struct {
	*mgr.NodeManager
}

//NewNodeRpmMgr generates a NodeRpmMgr object
func NewNodeRpmMgr() NodeRpmMgr {
	myNodeMgr := &mgr.NodeManager{}
	myNodeRpmMgr := NodeRpmMgr{myNodeMgr}

	myNodeRpmMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myNodeRpmMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck

	//TODO pending

	return myNodeRpmMgr
}
