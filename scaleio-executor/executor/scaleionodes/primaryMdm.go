package scaleionodes

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
	xplatformsys "github.com/dvonthenen/goxplatform/sys"

	config "github.com/codedellemc/scaleio-framework/scaleio-executor/config"
	common "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/common"
	ubuntu14 "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/deb/ubuntu14"
	mgr "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/mgr"
	rhel7 "github.com/codedellemc/scaleio-framework/scaleio-executor/executor/pkgmgr/rpm/rhel7"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleioPrimaryMdmNode implementation for ScaleIO Primary MDM Node
type ScaleioPrimaryMdmNode struct {
	common.ScaleioNode
	PkgMgr mgr.IMdmMgr
}

//NewPri generates a Primary MDM Node object
func NewPri(state *types.ScaleIOFramework, cfg *config.Config, getstate common.RetrieveState) *ScaleioPrimaryMdmNode {
	myNode := &ScaleioPrimaryMdmNode{}
	myNode.Config = cfg
	myNode.GetState = getstate
	myNode.RebootRequired = false

	var pkgmgr mgr.IMdmMgr
	switch xplatform.GetInstance().Sys.GetOsType() {
	case xplatformsys.OsRhel:
		log.Infoln("Is RHEL7")
		pkgmgr = rhel7.NewMdmRpmRhel7Mgr(state)
	case xplatformsys.OsUbuntu:
		log.Infoln("Is Ubuntu14")
		pkgmgr = ubuntu14.NewMdmDebUbuntu14Mgr(state)
	}
	myNode.PkgMgr = pkgmgr

	return myNode
}

//RunStateUnknown default action for StateUnknown
func (spmn *ScaleioPrimaryMdmNode) RunStateUnknown() {
	reboot, err := spmn.PkgMgr.EnvironmentSetup(spmn.State)
	if err != nil {
		log.Errorln("EnvironmentSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateCleanPrereqsReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanPrereqsReboot")
	}

	common.WaitForCleanPrereqsReboot(spmn)

	errState = spmn.UpdateNodeState(types.StatePrerequisitesInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StatePrerequisitesInstalled")
	}

	//requires a reboot?
	if reboot {
		log.Infoln("Reboot required before StatePrerequisitesInstalled!")

		if spmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := spmn.Config.ParseIPFromRestURI()

			if err1 == nil && err2 == nil && ip1 == ip2 {
				log.Infoln("Delay reboot host running the Scheduler")
				time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)
			}

			rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
			if rebootErr != nil {
				log.Errorln("Install Kernel Failed:", rebootErr)
			}

			time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
		}
	} else {
		log.Infoln("No need to reboot while installing prerequisites")
	}
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (spmn *ScaleioPrimaryMdmNode) RunStatePrerequisitesInstalled() {
	common.WaitForPrereqsFinish(spmn)
	err := spmn.PkgMgr.ManagementSetup(spmn.State, true)
	if err != nil {
		log.Errorln("ManagementSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = spmn.PkgMgr.NodeSetup(spmn.State)
	if err != nil {
		log.Errorln("NodeSetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = spmn.UpdateDevices()
	if err != nil {
		log.Errorln("UpdateDevices Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateBasePackagedInstalled)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateBasePackagedInstalled")
	}
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (spmn *ScaleioPrimaryMdmNode) RunStateBasePackagedInstalled() {
	common.WaitForBaseFinish(spmn)
	err := spmn.PkgMgr.CreateCluster(spmn.State)
	if err != nil {
		log.Errorln("CreateCluster Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	err = spmn.UpdateCluster()
	if err != nil {
		log.Errorln("UpdateCluster Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateInitializeCluster)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateInitializeCluster")
	}
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (spmn *ScaleioPrimaryMdmNode) RunStateInitializeCluster() {
	common.WaitForClusterInstallFinish(spmn)
	reboot, err := spmn.PkgMgr.GatewaySetup(spmn.State)
	if err != nil {
		log.Errorln("GatewaySetup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}
	spmn.RebootRequired = spmn.RebootRequired || reboot

	errState := spmn.UpdateNodeState(types.StateAddResourcesToScaleIO)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateAddResourcesToScaleIO")
	}
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (spmn *ScaleioPrimaryMdmNode) RunStateInstallRexRay() {
	reboot, err := spmn.PkgMgr.RexraySetup(spmn.State, spmn.Config.ExecutorID)
	if err != nil {
		log.Errorln("REX-Ray setup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}
	spmn.RebootRequired = spmn.RebootRequired || reboot

	err = spmn.PkgMgr.SetupIsolator(spmn.State)
	if err != nil {
		log.Errorln("Mesos Isolator setup Failed:", err)
		errState := spmn.UpdateNodeState(types.StateFatalInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFatalInstall")
		}
		return
	}

	errState := spmn.UpdateNodeState(types.StateCleanInstallReboot)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateCleanInstallReboot")
	}

	common.WaitForCleanInstallReboot(spmn)

	//requires a reboot?
	if spmn.RebootRequired {
		log.Infoln("Reboot required before StateFinishInstall!")
		log.Debugln("rebootRequired:", spmn.RebootRequired)

		errState := spmn.UpdateNodeState(types.StateSystemReboot)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateSystemReboot")
		}

		if spmn.State.Debug {
			log.Infoln("Skipping the reboot since Debug is TRUE")
		} else {
			ip1, err1 := xplatform.GetInstance().Nw.AutoDiscoverIP()
			ip2, err2 := spmn.Config.ParseIPFromRestURI()

			if err1 == nil && err2 == nil && ip1 == ip2 {
				log.Infoln("Delay reboot host running the Scheduler")
				time.Sleep(time.Duration(common.DelayForRebootInSeconds) * time.Second)
			}

			rebootErr := xplatform.GetInstance().Run.Command(common.RebootCmdline, common.RebootCheck, "")
			if rebootErr != nil {
				log.Errorln("Install Kernel Failed:", rebootErr)
			}

			time.Sleep(time.Duration(common.WaitForRebootInSeconds) * time.Second)
		}
	} else {
		log.Infoln("No need to reboot while installing REX-Ray")

		errState := spmn.UpdateNodeState(types.StateFinishInstall)
		if errState != nil {
			log.Errorln("Failed to signal state change:", errState)
		} else {
			log.Debugln("Signaled StateFinishInstall")
		}
	}
}

//RunStateSystemReboot default action for StateSystemReboot
func (spmn *ScaleioPrimaryMdmNode) RunStateSystemReboot() {
	errState := spmn.UpdateNodeState(types.StateFinishInstall)
	if errState != nil {
		log.Errorln("Failed to signal state change:", errState)
	} else {
		log.Debugln("Signaled StateFinishInstall")
	}
}

//RunStateFinishInstall default action for StateFinishInstall
func (spmn *ScaleioPrimaryMdmNode) RunStateFinishInstall() {
	node := spmn.GetSelfNode()
	if !node.Imperative && !node.Advertised {
		err := spmn.UpdateDevices()
		if err == nil {
			log.Infoln("UpdateDevices() Succcedeed. Devices advertised!")
		} else {
			log.Errorln("UpdateDevices() Failed. Err:", err)
		}
	}

	log.Debugln("In StateFinishInstall. Wait for", common.PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(common.PollForChangesInSeconds) * time.Second)

	//TODO eventual plan for MDM node behavior
	/*
		if clusterStatusBad then
			doClusterRemediate()
		else if upgrade then
			_ = waitForClusterUpgrade(spmn.UpdateScaleIOState())
			doUpgrade()
	*/
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (spmn *ScaleioPrimaryMdmNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	//TODO process the upgrade here
}

//UpdateCluster this function tells the scheduler that ScaleIO has been configured
func (spmn *ScaleioPrimaryMdmNode) UpdateCluster() error {
	log.Debugln("UpdateCluster ENTER")

	url := spmn.State.SchedulerAddress + "/api/state"

	state := &types.UpdateCluster{
		Acknowledged: false,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Errorln("Failed to marshall state object:", err)
		log.Debugln("UpdateCluster LEAVE")
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	if err != nil {
		log.Errorln("Failed to create new HTTP request:", err)
		log.Debugln("UpdateCluster LEAVE")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to make HTTP call:", err)
		log.Debugln("UpdateCluster LEAVE")
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Failed to read the HTTP Body:", err)
		log.Debugln("UpdateCluster LEAVE")
		return err
	}

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.UpdateCluster
	err = json.Unmarshal(body, &newstate)
	if err != nil {
		log.Errorln("Failed to unmarshal the UpdateState object:", err)
		log.Debugln("UpdateCluster LEAVE")
		return err
	}

	log.Debugln("Acknowledged:", newstate.Acknowledged)

	if !newstate.Acknowledged {
		log.Errorln("Failed to receive an acknowledgement")
		log.Debugln("UpdateCluster LEAVE")
		return common.ErrStateChangeNotAcknowledged
	}

	log.Errorln("UpdateCluster Succeeded")
	log.Debugln("UpdateCluster LEAVE")
	return nil
}
