package basemgr

//MdmManager implementation for MDM Package Manager
type MdmManager struct {
	BaseManager

	//ScaleIO node
	MdmPackageName         string
	MdmPackageDownload     string
	MdmInstallCmd          string
	MdmInstallCheck        string
	LiaPackageName         string
	LiaPackageDownload     string
	LiaInstallCmd          string
	LiaInstallCheck        string
	LiaRestartCheck        string
	GatewayPackageName     string
	GatewayPackageDownload string
	GatewayInstallCmd      string
	GatewayInstallCheck    string
	GatewayRestartCheck    string
}
