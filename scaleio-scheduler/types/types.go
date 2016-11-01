package types

const (
	//Ubuntu14MdmPackageName MDM package name for Ubuntu14
	Ubuntu14MdmPackageName = "emc-scaleio-mdm"

	//Ubuntu14SdsPackageName SDS package name for Ubuntu14
	Ubuntu14SdsPackageName = "emc-scaleio-sds"

	//Ubuntu14SdcPackageName SDC package name for Ubuntu14
	Ubuntu14SdcPackageName = "emc-scaleio-sdc"

	//Ubuntu14LiaPackageName LIA package name for Ubuntu14
	Ubuntu14LiaPackageName = "emc-scaleio-lia"

	//Ubuntu14GwPackageName GW package name for Ubuntu14
	Ubuntu14GwPackageName = "emc-scaleio-gateway"

	//Rhel7MdmPackageName MDM package name for RHEL7
	Rhel7MdmPackageName = "EMC-ScaleIO-mdm-2.0-10000.2072.el7.x86_64"

	//Rhel7SdsPackageName SDS package name for RHEL7
	Rhel7SdsPackageName = "EMC-ScaleIO-sds-2.0-10000.2072.el7.x86_64"

	//Rhel7SdcPackageName SDC package name for RHEL7
	Rhel7SdcPackageName = "EMC-ScaleIO-sdc-2.0-10000.2072.el7.x86_64"

	//Rhel7LiaPackageName LIA package name for RHEL7
	Rhel7LiaPackageName = "EMC-ScaleIO-lia-2.0-10000.2072.el7.x86_64"

	//Rhel7GwPackageName GW package name for RHEL7
	Rhel7GwPackageName = "EMC-ScaleIO-gateway-2.0-10000.2072.x86_64"

	//RexRayPackageName rexray package name
	RexRayPackageName = "rexray"

	//DvdcliPackageName DVDCLI package name
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

	//StateSystemReboot system is rebooting
	StateSystemReboot = 7

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

//Ubuntu14Packages describes the download URIs for Ubuntu install packages
type Ubuntu14Packages struct {
	Mdm string `json:"ubuntu14mdm"`
	Sds string `json:"ubuntu14sds"`
	Sdc string `json:"ubuntu14sdc"`
	Lia string `json:"ubuntu14lia"`
	Gw  string `json:"ubuntu14gw"`
}

//Rhel7Packages describes the download URIs for CentOS install packages
type Rhel7Packages struct {
	Mdm string `json:"rhel7mdm"`
	Sds string `json:"rhel7sds"`
	Sdc string `json:"rhel7sdc"`
	Lia string `json:"rhel7lia"`
	Gw  string `json:"rhel7gw"`
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
	Ubuntu14         Ubuntu14Packages
	Rhel7            Rhel7Packages
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
	Debug            bool              `json:"debug"`
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
