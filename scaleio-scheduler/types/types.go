package types

const (
	//DebMdmPackageName MDM package name for DEB
	DebMdmPackageName = "emc-scaleio-mdm"

	//DebSdsPackageName SDS package name for DEB
	DebSdsPackageName = "emc-scaleio-sds"

	//DebSdcPackageName SDC package name for DEB
	DebSdcPackageName = "emc-scaleio-sdc"

	//DebLiaPackageName LIA package name for DEB
	DebLiaPackageName = "emc-scaleio-lia"

	//DebGwPackageName GW package name for DEB
	DebGwPackageName = "emc-scaleio-gateway"

	//RexRayPackageName rexray package name for DEB
	RexRayPackageName = "rexray"

	//DvdcliPackageName DVDCLI package name for DEB
	DvdcliPackageName = "dvdcli"
)

const (
	//PersonaMdmPrimary is the first MDM
	PersonaMdmPrimary = 1

	//PersonaMdmSecondary is the second MDM
	PersonaMdmSecondary = 2

	//PersonaTb is the tie breaker
	PersonaTb = 3

	//PersonaNode is just a normal data node
	PersonaNode = 4
)

const (
	//StateUnknown will start with a fresh installation (or upgrade)
	StateUnknown = 0

	//StateCleanPrereqsReboot after kernel version has been updated
	StateCleanPrereqsReboot = 1

	//StatePrerequisitesInstalled the prerequisite packages are installed
	StatePrerequisitesInstalled = 2

	//StateBasePackagedInstalled the base ScaleIO packages are installed
	StateBasePackagedInstalled = 3

	//StateInitializeCluster the cluster is setup, now initial the cluster
	StateInitializeCluster = 4

	//StateInstallRexRay install rexray
	StateInstallRexRay = 5

	//StateCleanInstallReboot after installing all components
	StateCleanInstallReboot = 6

	//StateFinishInstall the agent node installation is complete
	StateFinishInstall = 1024

	//StateUpgradeCluster start the upgrade process
	StateUpgradeCluster = 2048

	//StateFatalInstall the agent node installation had a fatal error
	//manual intervention is required for now
	StateFatalInstall = 4096
)

//Version describes the version of the REST API
type Version struct {
	VersionInt int               `json:"versionint"`
	VersionStr string            `json:"versionstr"`
	BuildStr   string            `json:"buildstr,omitempty"`
	KeyValue   map[string]string `json:"keyvalue,omitempty"`
}

//DebPackages describes the download URIs for Ubuntu install packages
type DebPackages struct {
	DebMdm string `json:"debmdm"`
	DebSds string `json:"debsds"`
	DebSdc string `json:"debsdc"`
	DebLia string `json:"deblia"`
	DebGw  string `json:"debgw"`
}

//RpmPackages describes the download URIs for CentOS install packages
type RpmPackages struct {
	RpmMdm string `json:"rpmmdm"`
	RpmSds string `json:"rpmsds"`
	RpmSdc string `json:"rpmsdc"`
	RpmLia string `json:"rpmlia"`
	RpmGw  string `json:"rpmgw"`
}

//ScaleIONode node definition
type ScaleIONode struct {
	AgentID     string            `json:"name"`
	TaskID      string            `json:"taskid"`
	ExecutorID  string            `json:"executorid"`
	OfferID     string            `json:"offerid"`
	IPAddress   string            `json:"ipaddress"`
	Hostname    string            `json:"hostname"`
	Index       int               `json:"index"`
	Persona     int               `json:"persona"`
	State       int               `json:"state"`
	InCluster   bool              `json:"incluster"`
	LastContact int64             `json:"lastcontact"`
	KeyValue    map[string]string `json:"keyvalue,omitempty"`
}

//ScaleIONodes collection of ScaleIONode
type ScaleIONodes []*ScaleIONode

//ScaleIOPreConfig information to attach Data nodes to an existing
//ScaleIO cluster
type ScaleIOPreConfig struct {
	PreConfigEnabled     bool   `json:"preconfigenabled"`
	PrimaryMdmAddress    string `json:"preconfigprimdm"`
	SecondaryMdmAddress  string `json:"preconfigsecmdm"`
	TieBreakerMdmAddress string `json:"preconfigtbmdm"`
	GatewayAddress       string `json:"preconfiggateway"`
}

//ScaleIOConfig describes the configuration for this cluster
type ScaleIOConfig struct {
	ClusterID        string            `json:"clusterid"`
	ClusterName      string            `json:"clustername"`
	LbGateway        string            `json:"lbgateway"`
	ProtectionDomain string            `json:"protectiondomain"` //optional. Default: pd
	StoragePool      string            `json:"storagepool"`      //optional. Default: sp
	AdminPassword    string            `json:"adminpassword"`    //optional. Default: Scaleio123
	BlockDevice      string            `json:"blockdevice"`      //optional. Default: /dev/xvdf
	KeyValue         map[string]string `json:"keyvalue,omitempty"`
	Nodes            ScaleIONodes
	Preconfig        ScaleIOPreConfig
	Deb              DebPackages
	Rpm              RpmPackages
}

//IsolatorConfig describes the configuration for the mesos isolator
//on these ScaleIO nodes
type IsolatorConfig struct {
	Binary string `json:"isolatorbinary"`
}

//RexrayConfig describes the configuration for REX-Ray on these ScaleIO nodes
type RexrayConfig struct {
	Branch  string `json:"rexraybranch"`
	Version string `json:"rexrayversion"`
}

//ScaleIOFramework describes the overall framework state
type ScaleIOFramework struct {
	SchedulerAddress string            `json:"scheduleraddress"`
	LogLevel         string            `json:"loglevel"`
	DemoMode         bool              `json:"demomode"`
	Experimental     bool              `json:"experimental"`
	KeyValue         map[string]string `json:"keyvalue,omitempty"`
	ScaleIO          ScaleIOConfig
	Rexray           RexrayConfig
	Isolator         IsolatorConfig
}

//UpdateNode describes an executor going through a state change
type UpdateNode struct {
	Acknowledged bool              `json:"acknowledged"`
	ExecutorID   string            `json:"executorid"`
	State        int               `json:"state"`
	KeyValue     map[string]string `json:"keyvalue,omitempty"`
}

//AddNode describes an executor being added to the ScaleIO cluster
type AddNode struct {
	Acknowledged bool              `json:"acknowledged"`
	ExecutorID   string            `json:"executorid"`
	KeyValue     map[string]string `json:"keyvalue,omitempty"`
}

//PingNode describes a "I am still here" update
type PingNode struct {
	Acknowledged bool              `json:"acknowledged"`
	ExecutorID   string            `json:"executorid"`
	KeyValue     map[string]string `json:"keyvalue,omitempty"`
}
