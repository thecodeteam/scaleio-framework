package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func getVersion(w http.ResponseWriter, r *http.Request, server *RestServer) {
	ver := types.Version{
		VersionInt: config.VersionInt,
		VersionStr: config.VersionStr,
		BuildStr:   "",
	}

	response, err := json.MarshalIndent(ver, "", "  ")
	if err != nil {
		http.Error(w, "Unable to marshall the response", http.StatusBadRequest)
		return
	}

	log.Debugln("response:", string(response))
	fmt.Fprintf(w, string(response))
}
