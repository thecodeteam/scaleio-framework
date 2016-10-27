package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrStateChangeNotAcknowledged failed to find MDM Pair
	ErrStateChangeNotAcknowledged = errors.New("The node state change was not acknowledged")
)

//ScaleioNode implementation for ScaleIO Node
type ScaleioNode struct {
	ExecutorID     string
	RebootRequired bool
	Node           *types.ScaleIONode
	State          *types.ScaleIOFramework
	GetState       RetrieveState
}

//SetExecutorID sets the ExecutorID
func (bsn *ScaleioNode) SetExecutorID(ID string) {
	bsn.ExecutorID = ID
}

//SetRetrieveState sets the retrieve state function
func (bsn *ScaleioNode) SetRetrieveState(getstate RetrieveState) {
	bsn.GetState = getstate
}

//GetSelfNode get self node
func (bsn *ScaleioNode) GetSelfNode() *types.ScaleIONode {
	return bsn.Node
}

//UpdateScaleIOState updates the state of the framework
func (bsn *ScaleioNode) UpdateScaleIOState() *types.ScaleIOFramework {
	state, err := bsn.GetState()
	if err != nil {
		log.Warnln("getState() failed:", err)
	}
	bsn.State = state
	bsn.Node = GetSelfNode(bsn.ExecutorID, bsn.State)

	return bsn.State
}

func personaToString(persona int) string {
	switch persona {
	case types.PersonaMdmPrimary:
		return "primary"
	case types.PersonaMdmSecondary:
		return "secondary"
	case types.PersonaTb:
		return "tiebreaker"
	case types.PersonaNode:
		return "data"
	default:
		return "unknown"
	}
}

//LeaveMarkerFileForConfigured sets a marker file when in demo mode
func (bsn *ScaleioNode) LeaveMarkerFileForConfigured() {
	err := os.MkdirAll("/etc/scaleio-framework", 0644)
	if err != nil {
		log.Errorln("Unable to mkdir:", err)
	}

	data := []byte(personaToString(bsn.Node.Persona))
	err = ioutil.WriteFile("/etc/scaleio-framework/state", data, 0644)
	if err != nil {
		log.Errorln("Unable to write to marker file:", err)
	}
}

//UpdateNodeState this function tells the scheduler that the executor's state
//has changed
func (bsn *ScaleioNode) UpdateNodeState(nodeState int) error {
	log.Debugln("NotifyNodeState ENTER")
	log.Debugln("State:", nodeState)

	url := bsn.State.SchedulerAddress + "/api/node/state"

	state := &types.UpdateNode{
		Acknowledged: false,
		ExecutorID:   bsn.ExecutorID,
		State:        nodeState,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Errorln("Failed to marshall state object:", err)
		log.Debugln("NotifyNodeState LEAVE")
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	if err != nil {
		log.Errorln("Failed to create new HTTP request:", err)
		log.Debugln("NotifyNodeState LEAVE")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to make HTTP call:", err)
		log.Debugln("NotifyNodeState LEAVE")
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Failed to read the HTTP Body:", err)
		log.Debugln("NotifyNodeState LEAVE")
		return err
	}

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.UpdateNode
	err = json.Unmarshal(body, &newstate)
	if err != nil {
		log.Errorln("Failed to unmarshal the UpdateState object:", err)
		log.Debugln("NotifyNodeState LEAVE")
		return err
	}

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)
	log.Debugln("State:", newstate.State)

	if !newstate.Acknowledged {
		log.Errorln("Failed to receive an acknowledgement")
		log.Debugln("NotifyNodeState LEAVE")
		return ErrStateChangeNotAcknowledged
	}

	log.Errorln("NotifyNodeState Succeeded")
	log.Debugln("NotifyNodeState LEAVE")
	return nil
}

//UpdatePingNode this function tells the scheduler that "I am still here"
func (bsn *ScaleioNode) UpdatePingNode() error {
	log.Debugln("UpdatePingNode ENTER")

	url := bsn.State.SchedulerAddress + "/api/node/ping"

	state := &types.PingNode{
		Acknowledged: false,
		ExecutorID:   bsn.ExecutorID,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Errorln("Failed to marshall state object:", err)
		log.Debugln("UpdatePingNode LEAVE")
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	if err != nil {
		log.Errorln("Failed to create new HTTP request:", err)
		log.Debugln("UpdatePingNode LEAVE")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to make HTTP call:", err)
		log.Debugln("UpdatePingNode LEAVE")
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Failed to read the HTTP Body:", err)
		log.Debugln("UpdatePingNode LEAVE")
		return err
	}

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.PingNode
	err = json.Unmarshal(body, &newstate)
	if err != nil {
		log.Errorln("Failed to unmarshal the UpdateState object:", err)
		log.Debugln("UpdatePingNode LEAVE")
		return err
	}

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)

	if !newstate.Acknowledged {
		log.Errorln("Failed to receive an acknowledgement")
		log.Debugln("UpdatePingNode LEAVE")
		return ErrStateChangeNotAcknowledged
	}

	log.Debugln("UpdatePingNode Succeeded")
	log.Debugln("UpdatePingNode LEAVE")
	return nil
}

//RunStateUnknown default action for StateUnknown
func (bsn *ScaleioNode) RunStateUnknown() {
	log.Debugln("In StateUnknown. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateCleanPrereqsReboot default action for StateCleanPrereqsReboot
func (bsn *ScaleioNode) RunStateCleanPrereqsReboot() {
	log.Debugln("In StateCleanPrereqsReboot. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStatePrerequisitesInstalled default action for StatePrerequisitesInstalled
func (bsn *ScaleioNode) RunStatePrerequisitesInstalled() {
	log.Debugln("In StatePrerequisitesInstalled. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateBasePackagedInstalled default action for StateBasePackagedInstalled
func (bsn *ScaleioNode) RunStateBasePackagedInstalled() {
	log.Debugln("In StateBasePackagedInstalled. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateInitializeCluster default action for StateInitializeCluster
func (bsn *ScaleioNode) RunStateInitializeCluster() {
	log.Debugln("In StateInitializeCluster. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateInstallRexRay default action for StateInstallRexRay
func (bsn *ScaleioNode) RunStateInstallRexRay() {
	log.Debugln("In StateInstallRexRay. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateCleanInstallReboot default action for StateCleanInstallReboot
func (bsn *ScaleioNode) RunStateCleanInstallReboot() {
	log.Debugln("In StateCleanInstallReboot. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateSystemReboot default action for StateSystemReboot
func (bsn *ScaleioNode) RunStateSystemReboot() {
	log.Debugln("In StateSystemReboot. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateFinishInstall default action for StateFinishInstall
func (bsn *ScaleioNode) RunStateFinishInstall() {
	log.Debugln("In StateFinishInstall. Wait for", PollForChangesInSeconds,
		"seconds for changes in the cluster.")
	time.Sleep(time.Duration(PollForChangesInSeconds) * time.Second)
}

//RunStateUpgradeCluster default action for StateUpgradeCluster
func (bsn *ScaleioNode) RunStateUpgradeCluster() {
	log.Debugln("In StateUpgradeCluster. Do nothing.")
	time.Sleep(time.Duration(PollStatusInSeconds) * time.Second)
}

//RunStateFatalInstall default action for StateFatalInstall
func (bsn *ScaleioNode) RunStateFatalInstall() {
	log.Debugln("Node marked Fatal. Wait for", PollAfterFatalInSeconds, "seconds")
	time.Sleep(time.Duration(PollAfterFatalInSeconds) * time.Second)
}
