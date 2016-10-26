package rpm

import (
	basemgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode/pkgmgr/basemgr"
)

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//RpmMgr implementation for RpmPkgMgr
type RpmMgr struct {
	*basemgr.BaseManager
}

//NewRpmMgr generates a RpmMgr object
func NewRpmMgr() RpmMgr {
	myBaseMgr := &basemgr.BaseManager{}
	myRpmMgr := RpmMgr{myBaseMgr}
	myRpmMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myRpmMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck
	return myRpmMgr
}
