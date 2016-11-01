package mgr

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//MdmManager implementation for MDM Package Manager
type MdmManager struct {
	NodeManager

	//ScaleIO node
	MdmPackageName         string
	MdmPackageDownload     string
	MdmInstallCmd          string
	MdmInstallCheck        string
	LiaPackageName         string
	LiaPackageDownload     string
	LiaInstallCmd          string
	LiaInstallCheck        string
	GatewayPackageName     string
	GatewayPackageDownload string
	GatewayInstallCmd      string
	GatewayInstallCheck    string
}

//UpdateAddNode this function tells the scheduler that the executor's ScaleIO
//components has been to the cluster
func (msn *MdmManager) UpdateAddNode(schedulerAddress string, executorID string) error {
	log.Debugln("UpdateAddNode ENTER")
	log.Debugln("SchedulerAddress:", schedulerAddress)
	log.Debugln("ExecutorID:", executorID)

	url := schedulerAddress + "/api/node/cluster"

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
