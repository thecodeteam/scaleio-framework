package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func displayState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	response := "<html><head><title>Output</title><meta http-equiv=\"refresh\" content=\"2\" /></head><body>"

	server.Lock()
	for _, node := range server.State.ScaleIO.Nodes {
		response += node.ExecutorID
		response += " = "
		switch node.State {
		case types.StateUnknown:
			response += "Installing Prerequisite Packages"

		case types.StateCleanPrereqsReboot:
			response += "Sync on Prerequisite Install"

		case types.StatePrerequisitesInstalled:
			response += "Installing ScaleIO Packages"

		case types.StateBasePackagedInstalled:
			response += "Creating ScaleIO Cluster"

		case types.StateInitializeCluster:
			response += "Initializing ScaleIO"

		case types.StateAddResourcesToScaleIO:
			response += "Adding resources to ScaleIO cluster"

		case types.StateInstallRexRay:
			response += "Installing REX-Ray"

		case types.StateCleanInstallReboot:
			response += "Sync Before for Reboot"

		case types.StateSystemReboot:
			response += "System is Rebooting"

		case types.StateFinishInstall:
			response += "ScaleIO Running"

		case types.StateFatalInstall:
			response += "Installation Failed"
		}
		response += "<br />"
	}
	server.Unlock()

	response += "</body></html>"

	//log.Debugln("response:", string(response))
	fmt.Fprintf(w, string(response))
}

func setState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		http.Error(w, "Unable to read the HTTP Body stream", http.StatusBadRequest)
		return
	}
	if err := r.Body.Close(); err != nil {
		log.Warnln("Unable to close the HTTP Body stream:", err)
	}

	state := &types.UpdateCluster{
		Acknowledged: false,
		KeyValue:     make(map[string]string),
	}
	if err := json.Unmarshal(body, &state); err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	//update the object
	server.Lock()
	server.State.ScaleIO.Configured = true
	server.Unlock()

	//update the store
	err = server.Store.SetConfigured()
	if err != nil {
		http.Error(w, "Failed to update the Cluster configured bit", http.StatusBadRequest)
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

func getState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	server.Lock()
	response, err := json.MarshalIndent(server.State, "", "  ")
	server.Unlock()

	if err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	log.Debugln("response:", string(response))
	fmt.Fprintf(w, string(response))
}
