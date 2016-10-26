package rpm

import (
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//MdmRpmMgr implementation for MdmRpmMgr
type MdmRpmMgr struct {
	*mgr.MdmManager
}

//NewMdmRpmMgr generates a MdmRpmMgr object
func NewMdmRpmMgr(state *types.ScaleIOFramework) MdmRpmMgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmRpmMgr := MdmRpmMgr{myMdmMgr}

	myMdmRpmMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myMdmRpmMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck

	//TODO pending

	return myMdmRpmMgr
}
