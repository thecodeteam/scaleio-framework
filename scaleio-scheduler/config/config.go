package config

import (
	"flag"
	"strconv"
)

//consts exported out of package
const (
	//VersionInt in INT form
	VersionInt = 1

	//VersionStr in string form
	VersionStr = "0.1.0-rc1"

	//CPUPerMdmExecutor num of CPUs to MDM executor
	CPUPerMdmExecutor = 1.5

	//CPUPerNonExecutor num of CPUs to Non-MDM executor
	CPUPerNonExecutor = 1.0

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
	debMdm = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-mdm-2.0-5014.0.Ubuntu.14.04.x86_64.deb"
	debSds = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-sds-2.0-5014.0.Ubuntu.14.04.x86_64.deb"
	debSdc = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-sdc-2.0-5014.0.Ubuntu.14.04.x86_64.deb"
	debLia = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-lia-2.0-5014.0.Ubuntu.14.04.x86_64.deb"
	debGw  = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/emc-scaleio-gateway_2.0-5014.0_amd64.deb"
	rpmMdm = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-mdm-2.0-6035.0.el7.x86_64.rpm"
	rpmSds = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-sds-2.0-6035.0.el7.x86_64.rpm"
	rpmSdc = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-sdc-2.0-6035.0.el7.x86_64.rpm"
	rpmLia = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-lia-2.0-6035.0.el7.x86_64.rpm"
	rpmGw  = "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/EMC-ScaleIO-gateway-2.0-6035.0.x86_64.rpm"
	isoBin = "https://github.com/emccode/mesos-module-dvdi/releases/download/v0.4.5/libmesos_dvdi_isolator-1.0.0.so"
)

//Config is the representation of the config
type Config struct {
	LogLevel     string
	DemoMode     bool
	Experimental bool

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

	ClusterName          string
	ClusterID            string
	ProtectionDomain     string
	StoragePool          string
	AdminPassword        string
	PrimaryMdmAddress    string
	SecondaryMdmAddress  string
	TieBreakerMdmAddress string
	GatewayAddress       string
	BlockDevice          string
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
	fs.BoolVar(&cfg.DemoMode, "demomode", cfg.DemoMode,
		"Sets the application to demo mode")
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
	fs.Float64Var(&cfg.ExecutorCPUFactor, "executor.cpufactor", cfg.ExecutorCPUFactor,
		"Fudge factor for effective CPU available. This allows overhead/reserve.")
	fs.Float64Var(&cfg.ExecutorMdmMemory, "executor.memory.mdm", cfg.ExecutorMdmMemory,
		"Memory resources (MB) to consume per-executor on MDM nodes")
	fs.Float64Var(&cfg.ExecutorNonMemory, "executor.memory.non", cfg.ExecutorNonMemory,
		"Memory resources (MB) to consume per-executor on Non-MDM nodes")
	fs.Float64Var(&cfg.ExecutorMemoryFactor, "executor.memoryfactor", cfg.ExecutorMemoryFactor,
		"Fudge factor for effective memory available. This allows overhead/reserve.")
	fs.StringVar(&cfg.User, "user", cfg.User, "The User account the framework is running under")
	fs.StringVar(&cfg.Hostname, "hostname", cfg.Hostname, "The Hostname where the framework runs")
	fs.StringVar(&cfg.Role, "role", cfg.Role, "Framework role to register with the Mesos master")

	fs.StringVar(&cfg.ClusterName, "scaleio.clustername", cfg.ClusterName, "ScaleIO Cluster Name")
	fs.StringVar(&cfg.ClusterID, "scaleio.clusterid", cfg.ClusterID, "ScaleIO Cluster ID")
	fs.StringVar(&cfg.ProtectionDomain, "scaleio.protectiondomain", cfg.ProtectionDomain,
		"ScaleIO Protection Domain Name")
	fs.StringVar(&cfg.StoragePool, "scaleio.storagepool", cfg.StoragePool,
		"ScaleIO StoragePool Name")
	fs.StringVar(&cfg.AdminPassword, "scaleio.password", cfg.AdminPassword,
		"ScaleIO Admin Password")
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
	fs.StringVar(&cfg.BlockDevice, "scaleio.device", cfg.BlockDevice,
		"Specifies which device to use")
	fs.StringVar(&cfg.DebMdm, "scaleio.deb.mdm", cfg.DebMdm, "ScaleIO MDM Package")
	fs.StringVar(&cfg.DebSds, "scaleio.deb.sds", cfg.DebSds, "ScaleIO SDS Package")
	fs.StringVar(&cfg.DebSdc, "scaleio.deb.sdc", cfg.DebSdc, "ScaleIO SDC Package")
	fs.StringVar(&cfg.DebLia, "scaleio.deb.lia", cfg.DebLia, "ScaleIO LIA Package")
	fs.StringVar(&cfg.DebGw, "scaleio.deb.gw", cfg.DebGw, "ScaleIO Gateway Package")
	fs.StringVar(&cfg.RpmMdm, "scaleio.rpm.mdm", cfg.RpmMdm, "ScaleIO MDM Package")
	fs.StringVar(&cfg.RpmSds, "scaleio.rpm.sds", cfg.RpmSds, "ScaleIO SDS Package")
	fs.StringVar(&cfg.RpmSdc, "scaleio.rpm.sdc", cfg.RpmSdc, "ScaleIO SDC Package")
	fs.StringVar(&cfg.RpmLia, "scaleio.rpm.lia", cfg.RpmLia, "ScaleIO LIA Package")
	fs.StringVar(&cfg.RpmGw, "scaleio.rpm.gw", cfg.RpmGw, "ScaleIO Gateway Package")
}

//NewConfig creates a new Config object
func NewConfig() *Config {
	ip, err := autoDiscoverIP()
	if err != nil {
		ip = "127.0.0.1"
	}

	return &Config{
		LogLevel:             env("LOG_LEVEL", "info"),
		DemoMode:             envBool("DEMO_MODE", "false"),
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
		ClusterName:          env("CLUSTER_NAME", "scaleio"),
		ClusterID:            env("CLUSTER_ID", ""),
		ProtectionDomain:     env("PROTECTION_DOMAIN", "default"),
		StoragePool:          env("STORAGE_POOL", "default"),
		AdminPassword:        env("ADMIN_PASSWORD", "Scaleio123"),
		PrimaryMdmAddress:    env("PRIMARY_MDM_ADDRESS", ""),
		SecondaryMdmAddress:  env("SECONDARY_MDM_ADDRESS", ""),
		TieBreakerMdmAddress: env("TIEBREAKER_MDM_ADDRESS", ""),
		GatewayAddress:       env("GATEWAY_ADDRESS", ""),
		BlockDevice:          env("BLOCK_DEVICE", "/dev/xvdf"),
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
