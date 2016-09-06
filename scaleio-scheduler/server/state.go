package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func displayState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	response := "<html><head><title>Output</title><meta http-equiv=\"refresh\" content=\"2\" /></head><body>"

	for _, node := range server.State.ScaleIO.Nodes {
		response += node.ExecutorID
		response += " = "
		switch node.State {
		case types.StateUnknown:
			response += "Installing Prerequisite Packages"

		case types.StatePrerequisitesInstalled:
			response += "Installing ScaleIO Packages"

		case types.StateBasePackagedInstalled:
			response += "Creating ScaleIO Cluster"

		case types.StateInitializeCluster:
			response += "Initializing ScaleIO"

		case types.StateInstallRexRay:
			response += "Installing REX-Ray"

		case types.StateFinishInstall:
			response += "ScaleIO Running"

		case types.StateFatalInstall:
			response += "Installation Failed"
		}
		response += "<br />"
	}

	response += "</body></html>"

	//log.Debugln("response:", string(response))
	fmt.Fprintf(w, string(response))
}

func getState(w http.ResponseWriter, r *http.Request, server *RestServer) {
	response, err := json.MarshalIndent(server.State, "", "  ")
	if err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	log.Debugln("response:", string(response))
	fmt.Fprintf(w, string(response))
}
