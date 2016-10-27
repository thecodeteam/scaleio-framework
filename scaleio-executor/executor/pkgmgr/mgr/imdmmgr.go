package mgr

import (
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//IMdmMgr abstracts the Package Manager for MDM nodes on the platform
type IMdmMgr interface {
	INodeMgr

	ManagementSetup(state *types.ScaleIOFramework, isPriOrSec bool) error
	CreateCluster(state *types.ScaleIOFramework) error
	AddSdsNodesToCluster(state *types.ScaleIOFramework, needsLogin bool) error
	InitializeCluster(state *types.ScaleIOFramework) error
	GatewaySetup(state *types.ScaleIOFramework) (bool, error)
}
