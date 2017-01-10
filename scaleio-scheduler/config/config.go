package config

import (
	"flag"
	"strconv"

	xplatform "github.com/dvonthenen/goxplatform"
)

//consts exported out of package
const (
	//VersionInt in INT form
	VersionInt = 1

	//VersionStr in string form
	VersionStr = "0.3.0"

	//CPUPerMdmExecutor num of CPUs to MDM executor
	CPUPerMdmExecutor = 1.5

	//CPUPerNonExecutor num of CPUs to Non-MDM executor
	CPUPerNonExecutor = 0.5

	//MemPerMdmExecutor mem to MDM executor
	MemPerMdmExecutor = 3072

	//MemPerNonExecutor mem to Non-MDM executor
	MemPerNonExecutor = 512

	//DefaultRestPort rest port
	DefaultRestPort = 35000

	//RexrayRetry exponential backoff for retries
	RexrayRetry = 5

	//RexrayDelay a random delay to add to the exponential backoff
	RexrayDelay = 7
)

//consts internal
const (
	debMdm = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-mdm-2.0-10000.2072.Ubuntu.14.04.x86_64.deb"
	debSds = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-sds-2.0-10000.2072.Ubuntu.14.04.x86_64.deb"
	debSdc = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-sdc-2.0-10000.2072.Ubuntu.14.04.x86_64.deb"
	debLia = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-lia-2.0-10000.2072.Ubuntu.14.04.x86_64.deb"
	debGw  = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/emc-scaleio-gateway_2.0-10000.2072_amd64.deb"
	rpmMdm = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-mdm-2.0-10000.2072.el7.x86_64.rpm"
	rpmSds = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-sds-2.0-10000.2072.el7.x86_64.rpm"
	rpmSdc = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-sdc-2.0-10000.2072.el7.x86_64.rpm"
	rpmLia = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-lia-2.0-10000.2072.el7.x86_64.rpm"
	rpmGw  = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0/EMC-ScaleIO-gateway-2.0-10000.2072.x86_64.rpm"
	isoBin = "https://github.com/emccode/mesos-module-dvdi/releases/download/v0.4.6/libmesos_dvdi_isolator-1.0.1.so"
)

//Config is the representation of the config
type Config struct {
	LogLevel        string
	Debug           bool
	DeleteKeyValues bool
	DumpKeyValues   bool
	StoreAddKey     string
	StoreAddVal     string
	StoreDelKey     string
	Experimental    bool

	RexrayBranch  string
	RexrayVersion string

	IsolatorBinary string

	RestAddress          string
	RestPort             int
	MasterREST           string
	AltExecutorPath      string
	ExecutorMdmCPU       float64
	ExecutorNonCPU       float64
	ExecutorCPUFactor    float64
	ExecutorMdmMemory    float64
	ExecutorNonMemory    float64
	ExecutorMemoryFactor float64
	User                 string
	Hostname             string
	Role                 string
	Store                string
	StoreURI             string

	ClusterName          string
	ClusterID            string
	LbGateway            string
	ProtectionDomain     string
	StoragePool          string
	AdminPassword        string
	APIVersion           string
	PrimaryMdmAddress    string
	SecondaryMdmAddress  string
	TieBreakerMdmAddress string
	GatewayAddress       string
	DebMdm               string
	DebSds               string
	DebSdc               string
	DebLia               string
	DebGw                string
	RpmMdm               string
	RpmSds               string
	RpmSdc               string
	RpmLia               string
	RpmGw                string
}

//AddFlags adds flags to the command line parsing
func (cfg *Config) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&cfg.LogLevel, "loglevel", cfg.LogLevel,
		"Set the logging level for the application")
	fs.BoolVar(&cfg.Debug, "debug", cfg.Debug,
		"Debug mode prevents the reboot so the logs dont get cycled")
	fs.BoolVar(&cfg.DeleteKeyValues, "store.delete", cfg.DeleteKeyValues,
		"Helper function that deletes ScaleIO Framework Key/Value Store")
	fs.BoolVar(&cfg.DumpKeyValues, "store.dump", cfg.DumpKeyValues,
		"Helper function that dumps ScaleIO Framework Key/Value Store")
	fs.StringVar(&cfg.StoreAddKey, "store.add.key", cfg.StoreAddKey,
		"Modify a select store key")
	fs.StringVar(&cfg.StoreAddVal, "store.add.value", cfg.StoreAddVal,
		"Set the values to the key provided by store.key")
	fs.StringVar(&cfg.StoreDelKey, "store.del.key", cfg.StoreDelKey,
		"Delete a select store key")
	fs.BoolVar(&cfg.Experimental, "experimental", cfg.Experimental,
		"Sets the application to experimental mode")

	fs.StringVar(&cfg.RexrayBranch, "rexray.branch", cfg.RexrayBranch,
		"Which branch to grab the REX-Ray package from")
	fs.StringVar(&cfg.RexrayVersion, "rexray.version", cfg.RexrayVersion,
		"Which version to install from the provided branch")

	fs.StringVar(&cfg.IsolatorBinary, "isolator.binary", cfg.IsolatorBinary,
		"The URL for which Mesos Module DVDI to install")

	fs.StringVar(&cfg.RestAddress, "rest.address", cfg.RestAddress,
		"Mesos scheduler REST API address")
	fs.IntVar(&cfg.RestPort, "rest.port", cfg.RestPort, "Mesos scheduler REST API port")
	fs.StringVar(&cfg.MasterREST, "uri", cfg.MasterREST, "Mesos scheduler API URL")
	fs.StringVar(&cfg.AltExecutorPath, "executor.altpath", cfg.AltExecutorPath,
		"Provide an alternate path to the executor binary")
	fs.Float64Var(&cfg.ExecutorMdmCPU, "executor.cpu.mdm", cfg.ExecutorMdmCPU,
		"CPU resources to consume per-executor on MDM nodes")
	fs.Float64Var(&cfg.ExecutorNonCPU, "executor.cpu.non", cfg.ExecutorNonCPU,
		"CPU resources to consume per-executor on Non-MDM nodes")
	fs.Float64Var(&cfg.ExecutorCPUFactor, "executor.cpu.factor", cfg.ExecutorCPUFactor,
		"Fudge factor for effective CPU available. This allows overhead/reserve.")
	fs.Float64Var(&cfg.ExecutorMdmMemory, "executor.mem.mdm", cfg.ExecutorMdmMemory,
		"Memory resources (MB) to consume per-executor on MDM nodes")
	fs.Float64Var(&cfg.ExecutorNonMemory, "executor.mem.non", cfg.ExecutorNonMemory,
		"Memory resources (MB) to consume per-executor on Non-MDM nodes")
	fs.Float64Var(&cfg.ExecutorMemoryFactor, "executor.mem.factor", cfg.ExecutorMemoryFactor,
		"Fudge factor for effective memory available. This allows overhead/reserve.")
	fs.StringVar(&cfg.User, "user", cfg.User, "The User account the framework is running under")
	fs.StringVar(&cfg.Hostname, "hostname", cfg.Hostname, "The Hostname where the framework runs")
	fs.StringVar(&cfg.Role, "role", cfg.Role, "Framework role to register with the Mesos master")
	fs.StringVar(&cfg.Store, "store.type", cfg.Store, "The type of keyvalue store to use")
	fs.StringVar(&cfg.StoreURI, "store.uri", cfg.StoreURI, "Store URI to connect with")

	fs.StringVar(&cfg.ClusterName, "scaleio.clustername", cfg.ClusterName, "ScaleIO Cluster Name")
	fs.StringVar(&cfg.ClusterID, "scaleio.clusterid", cfg.ClusterID, "ScaleIO Cluster ID")
	fs.StringVar(&cfg.LbGateway, "scaleio.lbgateway", cfg.LbGateway, "Load Balanced IP/DNS Name")
	fs.StringVar(&cfg.ProtectionDomain, "scaleio.protectiondomain", cfg.ProtectionDomain,
		"ScaleIO Protection Domain Name")
	fs.StringVar(&cfg.StoragePool, "scaleio.storagepool", cfg.StoragePool,
		"ScaleIO StoragePool Name")
	fs.StringVar(&cfg.AdminPassword, "scaleio.password", cfg.AdminPassword,
		"ScaleIO Admin Password")
	fs.StringVar(&cfg.APIVersion, "scaleio.apiversion", cfg.APIVersion,
		"ScaleIO API Version")
	fs.StringVar(&cfg.PrimaryMdmAddress, "scaleio.preconfig.primary",
		cfg.PrimaryMdmAddress, "Pre-Configured Pri MDM Node. Requires Sec and TB "+
			"MDM nodes to be Pre-Configured")
	fs.StringVar(&cfg.SecondaryMdmAddress, "scaleio.preconfig.secondary",
		cfg.SecondaryMdmAddress, "Pre-Configured Sec MDM Node. Requires Pri and TB "+
			"MDM nodes to be Pre-Configured")
	fs.StringVar(&cfg.TieBreakerMdmAddress, "scaleio.preconfig.tiebreaker",
		cfg.TieBreakerMdmAddress, "Pre-Configured Tiebreaker MDM Node. Requires Pri "+
			"and Sec MDM nodes to be Pre-Configured")
	fs.StringVar(&cfg.GatewayAddress, "scaleio.preconfig.gateway",
		cfg.GatewayAddress, "Used to set a separated Gateway node. Otherwise, the "+
			"Primary MDM node is assumed to have the Gateway installed on it.")
	fs.StringVar(&cfg.DebMdm, "scaleio.ubuntu14.mdm", cfg.DebMdm, "ScaleIO MDM Package for Ubuntu 14.04")
	fs.StringVar(&cfg.DebSds, "scaleio.ubuntu14.sds", cfg.DebSds, "ScaleIO SDS Package for Ubuntu 14.04")
	fs.StringVar(&cfg.DebSdc, "scaleio.ubuntu14.sdc", cfg.DebSdc, "ScaleIO SDC Package for Ubuntu 14.04")
	fs.StringVar(&cfg.DebLia, "scaleio.ubuntu14.lia", cfg.DebLia, "ScaleIO LIA Package for Ubuntu 14.04")
	fs.StringVar(&cfg.DebGw, "scaleio.ubuntu14.gw", cfg.DebGw, "ScaleIO Gateway Package for Ubuntu 14.04")
	fs.StringVar(&cfg.RpmMdm, "scaleio.rhel7.mdm", cfg.RpmMdm, "ScaleIO MDM Package for RHEL7")
	fs.StringVar(&cfg.RpmSds, "scaleio.rhel7.sds", cfg.RpmSds, "ScaleIO SDS Package for RHEL7")
	fs.StringVar(&cfg.RpmSdc, "scaleio.rhel7.sdc", cfg.RpmSdc, "ScaleIO SDC Package for RHEL7")
	fs.StringVar(&cfg.RpmLia, "scaleio.rhel7.lia", cfg.RpmLia, "ScaleIO LIA Package for RHEL7")
	fs.StringVar(&cfg.RpmGw, "scaleio.rhel7.gw", cfg.RpmGw, "ScaleIO Gateway Package for RHEL7")
	fs.StringVar(&cfg.RpmMdm, "scaleio.centos7.mdm", cfg.RpmMdm, "ScaleIO MDM Package for CentOS7")
	fs.StringVar(&cfg.RpmSds, "scaleio.centos7.sds", cfg.RpmSds, "ScaleIO SDS Package for CentOS7")
	fs.StringVar(&cfg.RpmSdc, "scaleio.centos7.sdc", cfg.RpmSdc, "ScaleIO SDC Package for CentOS7")
	fs.StringVar(&cfg.RpmLia, "scaleio.centos7.lia", cfg.RpmLia, "ScaleIO LIA Package for CentOS7")
	fs.StringVar(&cfg.RpmGw, "scaleio.centos7.gw", cfg.RpmGw, "ScaleIO Gateway Package for CentOS7")
}

//NewConfig creates a new Config object
func NewConfig() *Config {
	ip, err := xplatform.GetInstance().Nw.AutoDiscoverIP()
	if err != nil {
		ip = "127.0.0.1"
	}

	return &Config{
		LogLevel:             env("LOG_LEVEL", "info"),
		Debug:                envBool("DEBUG", "false"),
		DeleteKeyValues:      envBool("DELETE_STORE", "false"),
		DumpKeyValues:        envBool("DUMP_STORE", "false"),
		StoreAddKey:          env("STORE_ADD_KEY", ""),
		StoreAddVal:          env("STORE_ADD_VAL", ""),
		StoreDelKey:          env("STORE_DEL_KEY", ""),
		Experimental:         envBool("EXPERIMENTAL", "false"),
		RexrayBranch:         env("REXRAY_BRANCH", "stable"),
		RexrayVersion:        env("REXRAY_VERSION", "latest"),
		IsolatorBinary:       env("ISOLATOR_BINARY", isoBin),
		RestAddress:          env("REST_ADDRESS", ip),
		RestPort:             envInt("REST_PORT", strconv.Itoa(DefaultRestPort)),
		MasterREST:           env("MESOS_MASTER_HTTP", "http://127.0.0.1:5050/api/v1/scheduler"),
		AltExecutorPath:      env("ALT_EXECUTOR_PATH", ""),
		ExecutorMdmCPU:       envFloat("EXECUTOR_MDM_CPU", strconv.FormatFloat(CPUPerMdmExecutor, 'f', -1, 64)),
		ExecutorNonCPU:       envFloat("EXECUTOR_NON_CPU", strconv.FormatFloat(CPUPerNonExecutor, 'f', -1, 64)),
		ExecutorCPUFactor:    envFloat("EXECUTOR_CPU_FACTOR", "1.0"),
		ExecutorMdmMemory:    envFloat("EXECUTOR_MDM_MEM", strconv.Itoa(MemPerMdmExecutor)),
		ExecutorNonMemory:    envFloat("EXECUTOR_NON_MEM", strconv.Itoa(MemPerNonExecutor)),
		ExecutorMemoryFactor: envFloat("EXECUTOR_MEMORY_FACTOR", "1.0"),
		User:                 env("USER", mesosUser()),
		Hostname:             env("HOSTNAME", mesosHostname()),
		Role:                 env("ROLE", "scaleio"),
		Store:                env("STORE_TYPE", "zk"),
		StoreURI:             env("STORE_URI", ""),
		ClusterName:          env("CLUSTER_NAME", "scaleio"),
		ClusterID:            env("CLUSTER_ID", ""),
		LbGateway:            env("LB_GATEWAY", ""),
		ProtectionDomain:     env("PROTECTION_DOMAIN", "default"),
		StoragePool:          env("STORAGE_POOL", "default"),
		AdminPassword:        env("ADMIN_PASSWORD", "Scaleio123"),
		APIVersion:           env("API_VERSION", "2.0"),
		PrimaryMdmAddress:    env("PRIMARY_MDM_ADDRESS", ""),
		SecondaryMdmAddress:  env("SECONDARY_MDM_ADDRESS", ""),
		TieBreakerMdmAddress: env("TIEBREAKER_MDM_ADDRESS", ""),
		GatewayAddress:       env("GATEWAY_ADDRESS", ""),
		DebMdm:               env("DEB_MDM", debMdm),
		DebSds:               env("DEB_SDS", debSds),
		DebSdc:               env("DEB_SDC", debSdc),
		DebLia:               env("DEB_LIA", debLia),
		DebGw:                env("DEB_GW", debGw),
		RpmMdm:               env("RPM_MDM", rpmMdm),
		RpmSds:               env("RPM_SDS", rpmSds),
		RpmSdc:               env("RPM_SDC", rpmSdc),
		RpmLia:               env("RPM_LIA", rpmLia),
		RpmGw:                env("RPM_GW", rpmGw),
	}
}
