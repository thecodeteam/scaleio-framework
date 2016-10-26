package mgr

import (
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//INodeMgr abstracts the Package Manager for ScaleIO node on the platform
type INodeMgr interface {
	EnvironmentSetup(state *types.ScaleIOFramework) (bool, error)
	NodeSetup(state *types.ScaleIOFramework) error

	RexraySetup(state *types.ScaleIOFramework) (bool, error)
	SetupIsolator(state *types.ScaleIOFramework) error
}
