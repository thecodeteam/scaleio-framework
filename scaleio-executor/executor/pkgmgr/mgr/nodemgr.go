package mgr

import "errors"

var (
	//ErrBaseUnimplemented failed because function is unimplemented in the base class
	ErrBaseUnimplemented = errors.New("Function is unimplemented in the base class")
)

//NodeManager implementation for Package Manager for ScaleIO Nodes
type NodeManager struct {
	//ScaleIO node
	SdsPackageName     string
	SdsPackageDownload string
	SdsInstallCmd      string
	SdsInstallCheck    string
	SdcPackageName     string
	SdcPackageDownload string
	SdcInstallCmd      string
	SdcInstallCheck    string

	//REX-Ray
	RexrayInstallCheck string

	//Isolator
	DvdcliInstallCheck string
}
