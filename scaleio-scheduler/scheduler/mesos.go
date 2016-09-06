package scheduler

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/gogo/protobuf/proto"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	sched "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/sched"
	mesos "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/v1"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

func prepareExecutorInfo(cfg *config.Config, executorID string) *mesos.ExecutorInfo {
	schedulerURI := fmt.Sprintf("http://%s:%d", cfg.RestAddress, cfg.RestPort)
	log.Infoln("Scheduler URI:", schedulerURI)
	uri := fmt.Sprintf("%s/scaleio-executor", schedulerURI)
	log.Infoln("Executor URI:", uri)

	executorUris := []*mesos.CommandInfo_URI{}
	executorUris = append(executorUris, &mesos.CommandInfo_URI{Value: &uri, Executable: proto.Bool(true)})
	executorCommand := fmt.Sprintf(
		"chmod u+x scaleio-executor && ./scaleio-executor -loglevel=%s -rest.uri=%s",
		cfg.LogLevel, schedulerURI)

	// Create mesos scheduler driver.
	return &mesos.ExecutorInfo{
		Name:       proto.String("scaleio-executor"),
		ExecutorId: &mesos.ExecutorID{Value: proto.String(executorID)},
		Command: &mesos.CommandInfo{
			Value: proto.String(executorCommand), //command to run on agent
			Uris:  executorUris,                  //URI to download
		},
	}
}

func prepareFrameworkInfo(cfg *config.Config) *mesos.FrameworkInfo {
	// the framework
	fwinfo := &mesos.FrameworkInfo{
		User:     proto.String(cfg.User),
		Name:     proto.String("ScaleIO Framework"),
		Hostname: proto.String(cfg.Hostname),
	}

	return fwinfo
}

func generateAcknowledgeCall(ID *mesos.FrameworkInfo, update *mesos.TaskStatus) *sched.Call {
	message := &sched.Call{
		FrameworkId: ID.GetId(),
		Type:        sched.Call_ACKNOWLEDGE.Enum(),
		Acknowledge: &sched.Call_Acknowledge{
			AgentId: update.GetAgentId(),
			TaskId:  update.GetTaskId(),
			Uuid:    update.GetUuid(),
		},
	}

	return message
}

func generateAcceptCall(cfg *config.Config, offer *mesos.Offer, node *types.ScaleIONode) *sched.Call {
	//offer ids
	var offerIDs []*mesos.OfferID

	myID := &mesos.OfferID{
		Value: offer.GetId().Value,
	}
	offerIDs = append(offerIDs, myID)

	log.Infoln("OfferID:")
	log.Infoln(myID.String())

	//create task
	var tasks []*mesos.TaskInfo

	taskID := &mesos.TaskID{
		Value: proto.String(node.TaskID),
	}

	log.Infoln("TaskID:")
	log.Infoln(taskID.String())

	cpu := cfg.ExecutorNonCPU
	mem := cfg.ExecutorNonMemory
	if IsNodeAnMDMNode(node) {
		cpu = cfg.ExecutorMdmCPU
		mem = cfg.ExecutorMdmMemory
	}

	task := &mesos.TaskInfo{
		Name:     proto.String("task-" + node.TaskID),
		TaskId:   taskID,
		AgentId:  offer.GetAgentId(),
		Executor: prepareExecutorInfo(cfg, node.ExecutorID),
		Resources: []*mesos.Resource{
			&mesos.Resource{
				Name:   proto.String("cpus"),
				Type:   mesos.Value_SCALAR.Enum(),
				Scalar: &mesos.Value_Scalar{Value: proto.Float64(cpu)},
			},
			&mesos.Resource{
				Name:   proto.String("mem"),
				Type:   mesos.Value_SCALAR.Enum(),
				Scalar: &mesos.Value_Scalar{Value: proto.Float64(mem)},
			},
		},
	}

	tasks = append(tasks, task)

	log.Infoln("Task:")
	log.Infoln(task.String())

	//create operations
	var operations []*mesos.Offer_Operation

	operation := &mesos.Offer_Operation{
		Type: mesos.Offer_Operation_LAUNCH.Enum(),
		Launch: &mesos.Offer_Operation_Launch{
			TaskInfos: tasks,
		},
	}

	operations = append(operations, operation)

	log.Infoln("Operation:")
	log.Infoln(operation.String())

	//launch the task
	message := &sched.Call{
		FrameworkId: offer.GetFrameworkId(),
		Type:        sched.Call_ACCEPT.Enum(),
		Accept: &sched.Call_Accept{
			OfferIds:   offerIDs,
			Operations: operations,
			Filters:    &mesos.Filters{RefuseSeconds: proto.Float64(30)},
		},
	}

	log.Infoln("Call:")
	log.Infoln(message.String())

	return message
}

func generateDeclineCall(cfg *config.Config, offer *mesos.Offer) *sched.Call {
	//offer ids
	var offerIDs []*mesos.OfferID

	myID := &mesos.OfferID{
		Value: offer.GetId().Value,
	}
	offerIDs = append(offerIDs, myID)

	log.Infoln("OfferID:")
	log.Infoln(myID.String())

	//decline the offer
	message := &sched.Call{
		FrameworkId: offer.GetFrameworkId(),
		Type:        sched.Call_DECLINE.Enum(),
		Decline: &sched.Call_Decline{
			OfferIds: offerIDs,
			Filters:  &mesos.Filters{RefuseSeconds: proto.Float64(30)},
		},
	}

	log.Infoln("Call:")
	log.Infoln(message.String())

	return message
}

func filterResources(resources []*mesos.Resource, filter func(*mesos.Resource) bool) (result []*mesos.Resource) {
	for _, resource := range resources {
		if filter(resource) {
			result = append(result, resource)
		}
	}
	return result
}

func doesExecutorExistOnHost(offer *mesos.Offer) bool {
	// account for executor resources if there's an executor already running on the slave
	if len(offer.ExecutorIds) != 0 {
		log.Infoln("scaleio-executor already exists on host", offer.GetId().GetValue(),
			",", offer.GetHostname(), ",", offer.GetAgentId())
	}

	return len(offer.ExecutorIds) != 0
}
