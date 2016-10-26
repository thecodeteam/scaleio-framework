package deb

import (
	basemgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/basenode/pkgmgr/basemgr"
)

const (
	rexrayInstallCheck = "rexray has been installed to"
	dvdcliInstallCheck = "dvdcli has been installed to"
)

//DebMgr implementation for DebPkgMgr
type DebMgr struct {
	*basemgr.BaseManager
}

//NewDebMgr generates a DebMgr object
func NewDebMgr() DebMgr {
	myBaseMgr := &basemgr.BaseManager{}
	myDebPkgMgr := DebMgr{myBaseMgr}
	myDebPkgMgr.BaseManager.RexrayInstallCheck = rexrayInstallCheck
	myDebPkgMgr.BaseManager.DvdcliInstallCheck = dvdcliInstallCheck
	return myDebPkgMgr
}
