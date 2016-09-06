package server

import (
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/native"
)

func downloadExecutor(w http.ResponseWriter, r *http.Request, server *RestServer) {
	path := server.Config.AltExecutorPath
	if len(path) == 0 {
		log.Debugln("AltExecutorPath = \"\" BEGIN")
		pathTmp, err := util.GetFullPath()
		if err != nil {
			http.Error(w, "Unable to determine executor location", http.StatusNotFound)
			return
		}
		path = util.AppendSlash(pathTmp) + "scaleio-executor"
		log.Debugln("Path:", path)
		log.Debugln("AltExecutorPath = \"\" END")
	}
	log.Infoln("Path:", path)
	_, err := os.Stat(path)
	if err != nil {
		http.Error(w, "File does not exist", http.StatusNotFound)
		log.Errorln("Executor does not exist:", path)
	}
	log.Debugln("Path:", path)
	http.ServeFile(w, r, path)
}
