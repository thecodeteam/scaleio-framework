package basemgr

import (
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//IBaseMgr abstracts the Base Package Manager on the platform
type IBaseMgr interface {
	EnvironmentSetup(state *types.ScaleIOFramework) (bool, error)
	NodeSetup(state *types.ScaleIOFramework) error

	RexraySetup(state *types.ScaleIOFramework) (bool, error)
	SetupIsolator(state *types.ScaleIOFramework) error
}
