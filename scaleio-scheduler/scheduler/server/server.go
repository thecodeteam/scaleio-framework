package server

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	negroni "github.com/codegangsta/negroni"
	"github.com/gorilla/mux"

	config "github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	common "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/common"
	kvstore "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/kvstore"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	rootKey = "scaleio-framework"
)

//RestServer representation for a REST API server
type RestServer struct {
	Config *config.Config
	Store  *kvstore.KvStore
	Server *negroni.Negroni
	State  *types.ScaleIOFramework
	Index  int

	sync.Mutex
}

//NewRestServer generates a new REST API server
func NewRestServer(cfg *config.Config, store *kvstore.KvStore) *RestServer {
	preconfig := cfg.PrimaryMdmAddress != "" && cfg.SecondaryMdmAddress != "" &&
		cfg.TieBreakerMdmAddress != ""

	scaleio := &types.ScaleIOFramework{
		SchedulerAddress: fmt.Sprintf("http://%s:%d", cfg.RestAddress, cfg.RestPort),
		LogLevel:         cfg.LogLevel,
		Debug:            cfg.Debug,
		Experimental:     cfg.Experimental,
		KeyValue:         make(map[string]string),
		ScaleIO: &types.ScaleIOConfig{
			Configured:           store.GetConfigured(),
			ClusterID:            cfg.ClusterID,
			ClusterName:          cfg.ClusterName,
			LbGateway:            cfg.LbGateway,
			ProtectionDomain:     cfg.ProtectionDomain,
			StoragePool:          cfg.StoragePool,
			AdminPassword:        cfg.AdminPassword,
			APIVersion:           cfg.APIVersion,
			FakeUsedData:         0,
			CapacityData:         0,
			UsedData:             0,
			AtLeastOneImperative: false,
			KeyValue:             make(map[string]string),
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

	restServer := &RestServer{
		Config: cfg,
		Store:  store,
		State:  scaleio,
		Index:  1,
	}

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
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		setState(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/node/state", func(w http.ResponseWriter, r *http.Request) {
		setNodeState(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/node/device", func(w http.ResponseWriter, r *http.Request) {
		setNodeDevices(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/node/ping", func(w http.ResponseWriter, r *http.Request) {
		setNodePing(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/api/fake", func(w http.ResponseWriter, r *http.Request) {
		setFakeData(w, r, restServer)
	}).Methods("POST")
	mux.HandleFunc("/ui", func(w http.ResponseWriter, r *http.Request) {
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

	//MonitorForState watch for state changes
	go func() {
		err := restServer.MonitorForState()
		if err != nil {
			log.Errorln("MonitorForState:", err)
		}
	}()

	return restServer
}

func cloneState(src *types.ScaleIOFramework) *types.ScaleIOFramework {
	dst := &types.ScaleIOFramework{}

	dst.Debug = src.Debug
	dst.Experimental = src.Experimental
	dst.LogLevel = src.LogLevel
	dst.Rexray.Branch = src.Rexray.Branch
	dst.Rexray.Version = src.Rexray.Version
	dst.SchedulerAddress = src.SchedulerAddress
	dst.Isolator.Binary = src.Isolator.Binary
	dst.KeyValue = make(map[string]string)
	for key, val := range src.KeyValue {
		dst.KeyValue[key] = val
	}

	dst.ScaleIO = &types.ScaleIOConfig{}
	dst.ScaleIO.AdminPassword = src.ScaleIO.AdminPassword
	dst.ScaleIO.APIVersion = src.ScaleIO.APIVersion
	dst.ScaleIO.ClusterID = src.ScaleIO.ClusterID
	dst.ScaleIO.ClusterName = src.ScaleIO.ClusterName
	dst.ScaleIO.Configured = src.ScaleIO.Configured
	dst.ScaleIO.LbGateway = src.ScaleIO.LbGateway
	dst.ScaleIO.Preconfig.GatewayAddress = src.ScaleIO.Preconfig.GatewayAddress
	dst.ScaleIO.Preconfig.PreConfigEnabled = src.ScaleIO.Preconfig.PreConfigEnabled
	dst.ScaleIO.Preconfig.PrimaryMdmAddress = src.ScaleIO.Preconfig.PrimaryMdmAddress
	dst.ScaleIO.Preconfig.SecondaryMdmAddress = src.ScaleIO.Preconfig.SecondaryMdmAddress
	dst.ScaleIO.Preconfig.TieBreakerMdmAddress = src.ScaleIO.Preconfig.TieBreakerMdmAddress
	dst.ScaleIO.ProtectionDomain = src.ScaleIO.ProtectionDomain
	dst.ScaleIO.FakeUsedData = src.ScaleIO.FakeUsedData
	dst.ScaleIO.CapacityData = src.ScaleIO.CapacityData
	dst.ScaleIO.UsedData = src.ScaleIO.UsedData
	dst.ScaleIO.AtLeastOneImperative = src.ScaleIO.AtLeastOneImperative
	dst.ScaleIO.Rhel7.Gw = src.ScaleIO.Rhel7.Gw
	dst.ScaleIO.Rhel7.Lia = src.ScaleIO.Rhel7.Lia
	dst.ScaleIO.Rhel7.Mdm = src.ScaleIO.Rhel7.Mdm
	dst.ScaleIO.Rhel7.Sdc = src.ScaleIO.Rhel7.Sdc
	dst.ScaleIO.Rhel7.Sds = src.ScaleIO.Rhel7.Sds
	dst.ScaleIO.StoragePool = src.ScaleIO.StoragePool
	dst.ScaleIO.Ubuntu14.Gw = src.ScaleIO.Ubuntu14.Gw
	dst.ScaleIO.Ubuntu14.Lia = src.ScaleIO.Ubuntu14.Lia
	dst.ScaleIO.Ubuntu14.Mdm = src.ScaleIO.Ubuntu14.Mdm
	dst.ScaleIO.Ubuntu14.Sdc = src.ScaleIO.Ubuntu14.Sdc
	dst.ScaleIO.Ubuntu14.Sds = src.ScaleIO.Ubuntu14.Sds
	dst.ScaleIO.KeyValue = make(map[string]string)
	dst.ScaleIO.Nodes = make([]*types.ScaleIONode, 0)
	for key, val := range src.ScaleIO.KeyValue {
		dst.ScaleIO.KeyValue[key] = val
	}

	for _, node := range src.ScaleIO.Nodes {
		dstNode := &types.ScaleIONode{
			AgentID:         node.AgentID,
			TaskID:          node.TaskID,
			ExecutorID:      node.ExecutorID,
			OfferID:         node.OfferID,
			IPAddress:       node.IPAddress,
			Hostname:        node.Hostname,
			Persona:         node.Persona,
			State:           node.State,
			LastContact:     node.LastContact,
			Imperative:      node.Imperative,
			Advertised:      node.Advertised,
			KeyValue:        make(map[string]string),
			ProvidesDomains: make(map[string]*types.ProtectionDomain),
			ConsumesDomains: make(map[string]*types.ProtectionDomain),
		}
		for key, val := range node.KeyValue {
			dstNode.KeyValue[key] = val
		}
		for keyDomain, pDomain := range node.ProvidesDomains {
			dstPDomain := &types.ProtectionDomain{
				Name:     pDomain.Name,
				KeyValue: make(map[string]string),
				Pools:    make(map[string]*types.StoragePool),
			}
			for key, val := range pDomain.KeyValue {
				dstPDomain.KeyValue[key] = val
			}
			for keyPool, pPool := range pDomain.Pools {
				dstPool := &types.StoragePool{
					Name:     pPool.Name,
					Devices:  make([]string, 0),
					KeyValue: make(map[string]string),
				}
				for _, device := range pPool.Devices {
					dstPool.Devices = append(dstPool.Devices, device)
				}
				for key, val := range pPool.KeyValue {
					dstPool.KeyValue[key] = val
				}
				dstPDomain.Pools[keyPool] = dstPool
			}
			dstNode.ProvidesDomains[keyDomain] = dstPDomain
		}
		for keyDomain, cDomain := range node.ConsumesDomains {
			dstCDomain := &types.ProtectionDomain{
				Name:     cDomain.Name,
				KeyValue: make(map[string]string),
				Pools:    make(map[string]*types.StoragePool),
			}
			for key, val := range cDomain.KeyValue {
				dstCDomain.KeyValue[key] = val
			}
			for keyPool, pPool := range cDomain.Pools {
				dstPool := &types.StoragePool{
					Name:     pPool.Name,
					Devices:  make([]string, 0),
					KeyValue: make(map[string]string),
				}
				for _, device := range pPool.Devices {
					dstPool.Devices = append(dstPool.Devices, device)
				}
				for key, val := range pPool.KeyValue {
					dstPool.KeyValue[key] = val
				}
				dstCDomain.Pools[keyPool] = dstPool
			}
			dstNode.ConsumesDomains[keyDomain] = dstCDomain
		}

		dst.ScaleIO.Nodes = append(dst.ScaleIO.Nodes, dstNode)
	}

	return dst
}

//MonitorForState monitors for changes in state
func (s *RestServer) MonitorForState() error {
	cnt := uint64(1)

	var err error
	for {
		time.Sleep(time.Duration(common.PollStatusInSeconds) * time.Second)

		//must make a copy of the state because these operations can take a long time
		s.Lock()
		copyState := cloneState(s.State)
		s.Unlock()

		if common.SyncRunState(copyState, types.StateAddResourcesToScaleIO, true) {
			log.Debugln("Calling addResourcesToScaleIO()...")
			err := s.addResourcesToScaleIO(copyState)
			if err != nil {
				log.Errorln("addResourcesToScaleIO err:", err)
			}
			s.updateNodeState(types.StateInstallRexRay)
		}
		//to add more else if { SyncRunState(otherState) }

		//if in AWS, check for full and expand if needed
		if !copyState.ScaleIO.AtLeastOneImperative && (cnt%uint64(s.Config.CheckFull)) == 0 &&
			len(s.Config.AccessKey) > 0 && len(s.Config.SecretKey) > 0 {
			log.Debugln("Calling checkForFull()...")
			pairDomainPool, errCheck := s.checkForFull(copyState)
			if errCheck != nil {
				log.Errorln("checkForFull err:", errCheck)
			} else if len(pairDomainPool.Pools) == 0 {
				log.Infoln("There are no StoragePools that need expanding.")
			} else {
				errExpand := s.expandPools(pairDomainPool)
				if errExpand != nil {
					log.Errorln("expandPools err:", errCheck)
				}
			}
		}
		//if in AWS, check for full and expand if needed

		cnt = cnt + 1
	}

	return err
}

func (s *RestServer) updateNodeState(state int) {
	s.Lock()

	for i := 0; i < len(s.State.ScaleIO.Nodes); i++ {
		if s.State.ScaleIO.Nodes[i].State > state { //only update state if less than current
			continue
		}
		s.State.ScaleIO.Nodes[i].State = state
	}

	s.Unlock()
}
