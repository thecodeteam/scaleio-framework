package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

var (
	//ErrStateChangeNotAcknowledged failed to find MDM Pair
	ErrStateChangeNotAcknowledged = errors.New("The node state change was not acknowledged")
)

//UpdateNodeState this function tells the scheduler that the executor's state
//has changed
func UpdateNodeState(uri string, executorID string, nodeState int) error {
	log.Debugln("NotifyNodeState ENTER")
	log.Debugln("URI:", uri)
	log.Debugln("ExecutorID:", executorID)
	log.Debugln("State:", nodeState)

	url := uri + "/api/node/state"

	state := &types.UpdateNode{
		Acknowledged: false,
		ExecutorID:   executorID,
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

//UpdateAddNode this function tells the scheduler that the executor's ScaleIO
//components has been to the cluster
func UpdateAddNode(uri string, executorID string) error {
	log.Debugln("UpdateAddNode ENTER")
	log.Debugln("URI:", uri)
	log.Debugln("ExecutorID:", executorID)

	url := uri + "/api/node/cluster"

	state := &types.AddNode{
		Acknowledged: false,
		ExecutorID:   executorID,
	}

	response, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Errorln("Failed to marshall state object:", err)
		log.Debugln("UpdateAddNode LEAVE")
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(response))
	if err != nil {
		log.Errorln("Failed to create new HTTP request:", err)
		log.Debugln("UpdateAddNode LEAVE")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to make HTTP call:", err)
		log.Debugln("UpdateAddNode LEAVE")
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Failed to read the HTTP Body:", err)
		log.Debugln("UpdateAddNode LEAVE")
		return err
	}

	log.Debugln("response Status:", resp.Status)
	log.Debugln("response Headers:", resp.Header)
	log.Debugln("response Body:", string(body))

	var newstate types.AddNode
	err = json.Unmarshal(body, &newstate)
	if err != nil {
		log.Errorln("Failed to unmarshal the UpdateState object:", err)
		log.Debugln("UpdateAddNode LEAVE")
		return err
	}

	log.Debugln("Acknowledged:", newstate.Acknowledged)
	log.Debugln("ExecutorID:", newstate.ExecutorID)

	if !newstate.Acknowledged {
		log.Errorln("Failed to receive an acknowledgement")
		log.Debugln("UpdateAddNode LEAVE")
		return ErrStateChangeNotAcknowledged
	}

	log.Errorln("UpdateAddNode Succeeded")
	log.Debugln("UpdateAddNode LEAVE")
	return nil
}

//UpdatePingNode this function tells the scheduler that "I am still here"
func UpdatePingNode(uri string, executorID string) error {
	log.Debugln("UpdatePingNode ENTER")
	log.Debugln("URI:", uri)
	log.Debugln("ExecutorID:", executorID)

	url := uri + "/api/node/ping"

	state := &types.PingNode{
		Acknowledged: false,
		ExecutorID:   executorID,
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
