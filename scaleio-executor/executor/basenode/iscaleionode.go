package basenode

import "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"

//IScaleioNode is the interface for implementing a ScaleIO node
type IScaleioNode interface {
	SetExecutorID(ID string)
	SetRetrieveState(getstate RetrieveState)
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
