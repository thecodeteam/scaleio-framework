package basemgr

import "errors"

var (
	//ErrBaseUnimplemented failed because function is unimplemented in the base class
	ErrBaseUnimplemented = errors.New("Function is unimplemented in the base class")

	//ErrStateChangeNotAcknowledged failed to find MDM Pair
	ErrStateChangeNotAcknowledged = errors.New("The node state change was not acknowledged")
)

//BaseManager implementation for Base Package Manager
type BaseManager struct {
	//ScaleIO node
	SdsPackageName     string
	SdsPackageDownload string
	SdsInstallCmd      string
	SdsInstallCheck    string
	SdcPackageName     string
	SdcPackageDownload string
	SdcInstallCmd      string
	SdcInstallCheck    string

	//Rexray
	RexrayInstallCheck string

	//Isolator
	DvdcliInstallCheck string
}
