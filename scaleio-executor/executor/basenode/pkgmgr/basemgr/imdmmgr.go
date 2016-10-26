package basemgr

import (
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//IMdmMgr abstracts the MDM Package Manager on the platform
type IMdmMgr interface {
	IBaseMgr

	ManagementSetup(state *types.ScaleIOFramework, isPriOrSec bool) error
	CreateCluster(state *types.ScaleIOFramework) error
	AddSdsNodesToCluster(state *types.ScaleIOFramework, needsLogin bool) error
	InitializeCluster(state *types.ScaleIOFramework) error
	GatewaySetup(state *types.ScaleIOFramework) (bool, error)
}
