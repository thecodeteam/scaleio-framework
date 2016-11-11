package common

import "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"

//IScaleioNode is the interface for implementing a ScaleIO node
type IScaleioNode interface {
	GetSelfNode() *types.ScaleIONode
	UpdateScaleIOState() *types.ScaleIOFramework
	UpdateNodeState(nodeState int) error
	UpdateDevices() error
	UpdatePingNode() error

	RunStateUnknown()
	RunStateCleanPrereqsReboot()
	RunStatePrerequisitesInstalled()
	RunStateBasePackagedInstalled()
	RunStateInitializeCluster()
	RunStateAddResourcesToScaleIO()
	RunStateInstallRexRay()
	RunStateCleanInstallReboot()
	RunStateSystemReboot()
	RunStateFinishInstall()
	RunStateUpgradeCluster()
	RunStateFatalInstall()
}
