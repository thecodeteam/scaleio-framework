package rpm

import (
	log "github.com/Sirupsen/logrus"

	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//MdmRpmMgr implementation for MdmRpmMgr
type MdmRpmMgr struct {
	*mgr.MdmManager
}

//EnvironmentSetup for setting up the environment for ScaleIO
func (mdm *MdmRpmMgr) EnvironmentSetup(state *types.ScaleIOFramework) (bool, error) {
	log.Infoln("EnvironmentSetup ENTER")

	//TODO

	log.Infoln("EnvironmentSetup LEAVE")

	return false, nil
}

//NewMdmRpmMgr generates a MdmRpmMgr object
func NewMdmRpmMgr(state *types.ScaleIOFramework) MdmRpmMgr {
	myMdmMgr := &mgr.MdmManager{}
	myMdmRpmMgr := MdmRpmMgr{myMdmMgr}

	myMdmRpmMgr.MdmManager.RexrayInstallCheck = rexrayInstallCheck
	myMdmRpmMgr.MdmManager.DvdcliInstallCheck = dvdcliInstallCheck

	//TODO pending

	return myMdmRpmMgr
}
