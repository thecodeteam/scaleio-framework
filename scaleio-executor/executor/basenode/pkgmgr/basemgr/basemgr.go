package basemgr

import "errors"

var (
	//ErrBaseUnimplemented failed because function is unimplemented in the base class
	ErrBaseUnimplemented = errors.New("Function is unimplemented in the base class")
)

//BaseManager implementation for base Package Manager
type BaseManager struct {
	//ScaleIO node

	//Rexray
	RexrayInstallCheck string

	//Isolator
	DvdcliInstallCheck string
}
