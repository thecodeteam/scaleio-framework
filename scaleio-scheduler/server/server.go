package server

import (
	"fmt"
	"net/http"
	"strconv"

	negroni "github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//RestServer representation for a REST API server
type RestServer struct {
	Config *config.Config
	Server *negroni.Negroni
	State  *types.ScaleIOFramework
}

//NewRestServer generates a new REST API server
func NewRestServer(cfg *config.Config) *RestServer {
	preconfig := cfg.PrimaryMdmAddress != "" && cfg.SecondaryMdmAddress != "" &&
		cfg.TieBreakerMdmAddress != ""

	scaleio := &types.ScaleIOFramework{
		SchedulerAddress: fmt.Sprintf("http://%s:%d", cfg.RestAddress, cfg.RestPort),
		LogLevel:         cfg.LogLevel,
		DemoMode:         cfg.DemoMode,
		Debug:            cfg.Debug,
		Experimental:     cfg.Experimental,
		ScaleIO: types.ScaleIOConfig{
			ClusterID:        cfg.ClusterID,
			ClusterName:      cfg.ClusterName,
			LbGateway:        cfg.LbGateway,
			ProtectionDomain: cfg.ProtectionDomain,
			StoragePool:      cfg.StoragePool,
			AdminPassword:    cfg.AdminPassword,
			BlockDevice:      cfg.BlockDevice,
			Preconfig: types.ScaleIOPreConfig{
				PreConfigEnabled:     preconfig,
				PrimaryMdmAddress:    cfg.PrimaryMdmAddress,
				SecondaryMdmAddress:  cfg.SecondaryMdmAddress,
				TieBreakerMdmAddress: cfg.TieBreakerMdmAddress,
				GatewayAddress:       cfg.GatewayAddress,
			},
			Ubuntu14: types.Ubuntu14Packages{
				Mdm: cfg.DebMdm,
				Sds: cfg.DebSds,
				Sdc: cfg.DebSdc,
				Lia: cfg.DebLia,
				Gw:  cfg.DebGw,
			},
			Rhel7: types.Rhel7Packages{
				Mdm: cfg.RpmMdm,
				Sds: cfg.RpmSds,
				Sdc: cfg.RpmSdc,
				Lia: cfg.RpmLia,
				Gw:  cfg.RpmGw,
			},
		},
		Rexray: types.RexrayConfig{
			Branch:  cfg.RexrayBranch,
			Version: cfg.RexrayVersion,
		},
		Isolator: types.IsolatorConfig{
			Binary: cfg.IsolatorBinary,
		},
	}

	restServer := &RestServer{cfg, nil, scaleio}

	mux := mux.NewRouter()
	mux.HandleFunc("/scaleio-executor", func(w http.ResponseWriter, r *http.Request) {
		downloadExecutor(w, r, restServer)
	}).Methods("GET")
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		getVersion(w, r, restServer)
	}).Methods("GET")
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		getState(w, r, restServer)
	}).Methods("GET")
	mux.HandleFunc("/api/node/state", func(w http.ResponseWriter, r *http.Request) {
		setNodeState(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/node/cluster", func(w http.ResponseWriter, r *http.Request) {
		setNodeAdded(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/node/ping", func(w http.ResponseWriter, r *http.Request) {
		setNodePing(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/ui/state", func(w http.ResponseWriter, r *http.Request) {
		displayState(w, r, restServer)
	}).Methods("GET")
	//TODO delete this below when a real UI is embedded
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		displayState(w, r, restServer)
	}).Methods("GET")
	server := negroni.Classic()
	server.UseHandler(mux)

	//Run is a blocking call for Negroni... so go routine it
	go func() {
		server.Run(cfg.RestAddress + ":" + strconv.Itoa(cfg.RestPort))
	}()

	restServer.Server = server

	return restServer
}
