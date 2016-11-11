package common

import "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"

const (
	//PollStatusInSeconds the amount of time to wait before updating state
	PollStatusInSeconds = 30
)

//RetrieveState is a call back to retrieve an update of the state
type RetrieveState func() (*types.ScaleIOFramework, error)

//SyncRunState a function to waiting on a specific state
func SyncRunState(state *types.ScaleIOFramework, runState int, allNodes bool) bool {
	for _, node := range state.ScaleIO.Nodes {
		switch node.Persona {
		case types.PersonaMdmPrimary:
			if node.State < runState {
				return false
			}
		case types.PersonaMdmSecondary:
			if node.State < runState {
				return false
			}
		case types.PersonaTb:
			if node.State < runState {
				return false
			}
		case types.PersonaNode:
			if allNodes && node.State < runState {
				return false
			}
		}
	}

	return true
}
