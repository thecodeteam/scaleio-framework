package server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func findScaleIONodeByExecutorID(nodes types.ScaleIONodes, executorID string) *types.ScaleIONode {
	log.Debugln("findScaleIONodeByExecutorID ENTER")
	for i := 0; i < len(nodes); i++ {
		node := nodes[i]
		if node.ExecutorID == executorID {
			log.Debugln("Node Found:", node.ExecutorID)
			log.Debugln("findScaleIONodeByExecutorID LEAVE")
			return node
		}
	}
	log.Debugln("Node NOT Found")
	log.Debugln("findScaleIONodeByExecutorID LEAVE")
	return nil
}

func setNodeState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.UpdateNode{}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := findScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}
	node.State = state.State
	node.LastContact = time.Now().Unix()

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}

func setNodeAdded(w http.ResponseWriter, r *http.Request, server *RestServer) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.AddNode{}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := findScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}
	node.InCluster = true
	node.LastContact = time.Now().Unix()

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

	state := &types.PingNode{}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	node := findScaleIONodeByExecutorID(server.State.ScaleIO.Nodes, state.ExecutorID)
	if node == nil {
		http.Error(w, "Unable to find the Executor", http.StatusBadRequest)
		return
	}
	node.LastContact = time.Now().Unix()

	//acknowledged the state change
	state.Acknowledged = true

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
	}
}
