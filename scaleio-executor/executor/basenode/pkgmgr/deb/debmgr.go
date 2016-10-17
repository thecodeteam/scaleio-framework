package deb

import (
	basemgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode/pkgmgr/basemgr"
)

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//DebPkgMgr implementation for DebPkgMgr
type DebPkgMgr struct {
	*basemgr.BaseManager
}

//NewDebPkgMgr generates a DebPkgMgr object
func NewDebPkgMgr() *DebPkgMgr {
	myBaseMgr := &basemgr.BaseManager{}
	myDebPkgMgr := &DebPkgMgr{myBaseMgr}
	myDebPkgMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myDebPkgMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck
	return myDebPkgMgr
}
