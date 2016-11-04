package common

import (
	"github.com/codedellemc/scaleio-framework/scaleio-executor/config"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//IScaleioNode is the interface for implementing a ScaleIO node
type IScaleioNode interface {
	SetConfig(cfg *config.Config)
	SetRetrieveState(getstate RetrieveState)
	GetConfig() *config.Config
	GetSelfNode() *types.ScaleIONode
	UpdateScaleIOState() *types.ScaleIOFramework
	LeaveMarkerFileForConfigured()
	UpdateNodeState(nodeState int) error
	UpdatePingNode() error

	RunStateUnknown()
	RunStateCleanPrereqsReboot()
	RunStatePrerequisitesInstalled()
	RunStateBasePackagedInstalled()
	RunStateInitializeCluster()
	RunStateInstallRexRay()
	RunStateCleanInstallReboot()
	RunStateSystemReboot()
	RunStateFinishInstall()
	RunStateUpgradeCluster()
	RunStateFatalInstall()
}
