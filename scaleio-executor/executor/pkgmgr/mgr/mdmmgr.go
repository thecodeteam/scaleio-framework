package mgr

//MdmManager implementation for MDM Package Manager
type MdmManager struct {
	NodeManager

	//ScaleIO node
	MdmPackageName         string
	MdmPackageDownload     string
	MdmInstallCmd          string
	MdmInstallCheck        string
	LiaPackageName         string
	LiaPackageDownload     string
	LiaInstallCmd          string
	LiaInstallCheck        string
	GatewayPackageName     string
	GatewayPackageDownload string
	GatewayInstallCmd      string
	GatewayInstallCheck    string
}
