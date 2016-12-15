package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/common"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func setNodeState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.UpdateNode{
		Acknowledged: false,
		ExecutorID:   "",
		State:        types.StateUnknown,
		KeyValue:     make(map[string]string),
	}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := common.FindScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}

	server.Lock()
	//set state on object...
	node.State = state.State
	node.LastContact = time.Now().Unix()

	//save state in metadata...
	err = server.Store.SetNodeInfo(node.Hostname, node.Persona, node.State)
	server.Unlock()

	if err != nil {
		http.Error(w, "SetNodeInfo Err: "+err.Error(), http.StatusBadRequest)
		return
	}

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}

func setNodeDevices(w http.ResponseWriter, r *http.Request, server *RestServer) {
	log.Debugln("setNodeDevices ENTER")

	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.UpdateDevices{
		Acknowledged: false,
		ExecutorID:   "",
		Devices:      make([]string, 0),
		KeyValue:     make(map[string]string),
	}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := common.FindScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}

	server.Lock()
	//set last contact time on object...
	node.LastContact = time.Now().Unix()

	//Set advertised!
	node.Advertised = true

	for _, device := range state.Devices {
		log.Debugln("Device:", device)
		if node.ProvidesDomains == nil {
			node.ProvidesDomains = make(map[string]*types.ProtectionDomain)
		}
		if node.ProvidesDomains[server.Config.ProtectionDomain] == nil {
			log.Debugln("ProvidesDomains is nil")
			node.ProvidesDomains[server.Config.ProtectionDomain] = &types.ProtectionDomain{
				Name: server.Config.ProtectionDomain,
			}
		}
		pd := node.ProvidesDomains[server.Config.ProtectionDomain]
		if pd.Pools == nil {
			pd.Pools = make(map[string]*types.StoragePool)
		}
		if pd.Pools[server.Config.StoragePool] == nil {
			log.Debugln("Pools is nil")
			pd.Pools[server.Config.StoragePool] = &types.StoragePool{
				Name: server.Config.StoragePool,
			}
		}
		sp := pd.Pools[server.Config.StoragePool]

		if sp.Devices == nil {
			log.Debugln("Devices is nil")
			sp.Devices = make([]string, 0)
		}
		log.Debugln("Add device")
		sp.Devices = append(sp.Devices, device)
	}
	server.Unlock()

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}

func setNodePing(w http.ResponseWriter, r *http.Request, server *RestServer) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.PingNode{
		Acknowledged: false,
		ExecutorID:   "",
		KeyValue:     make(map[string]string),
	}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := common.FindScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}

	server.Lock()
	node.LastContact = time.Now().Unix()
	server.Unlock()

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}
