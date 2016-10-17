package rpm

import (
	basemgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode/pkgmgr/basemgr"
)

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//RpmPkgMgr implementation for RpmPkgMgr
type RpmPkgMgr struct {
	*basemgr.BaseManager
}

//NewRpmPkgMgr generates a RpmPkgMgr object
func NewRpmPkgMgr() *RpmPkgMgr {
	myBaseMgr := &basemgr.BaseManager{}
	myRpmPkgMgr := &RpmPkgMgr{myBaseMgr}
	myRpmPkgMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myRpmPkgMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck
	return myRpmPkgMgr
}
