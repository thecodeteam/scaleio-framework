package basemgr

import (
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//IPkgMgr abstracts the Package Manager on the platform
type IPkgMgr interface {
	RexraySetup(state *types.ScaleIOFramework) (bool, error)
	SetupIsolator(state *types.ScaleIOFramework) error
}
