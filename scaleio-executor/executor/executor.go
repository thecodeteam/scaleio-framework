package executor

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	xplatform "github.com/dvonthenen/goxplatform"
	jsonpb "github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/codedellemc/scaleio-framework/scaleio-executor/client"
	"github.com/codedellemc/scaleio-framework/scaleio-executor/config"
	exec "github.com/codedellemc/scaleio-framework/scaleio-executor/mesos/exec"
	mesos "github.com/codedellemc/scaleio-framework/scaleio-executor/mesos/v1"
	types "github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

const (
	numberOfEventFailuresBeforeExiting = 10
)

//ScaleIOExecutor is the representation for an ScaleIO Executor process
type ScaleIOExecutor struct {
	Config *config.Config

	ExecutorID  *mesos.ExecutorID
	FrameworkID *mesos.FrameworkID

	Client *client.Client

	Events   chan *exec.Event
	DoneChan chan struct{}

	ObservedFailures int
}

//NewScaleIOExecutor creates a ScaleIO executor object
func NewScaleIOExecutor(cfg *config.Config) *ScaleIOExecutor {
	return &ScaleIOExecutor{
		Config:           cfg,
		ExecutorID:       &mesos.ExecutorID{Value: proto.String(cfg.ExecutorID)},
		FrameworkID:      &mesos.FrameworkID{Value: proto.String(cfg.FrameworkID)},
		Client:           client.New(cfg.MesosAgent, "/api/v1/executor"),
		Events:           make(chan *exec.Event),
		DoneChan:         make(chan struct{}),
		ObservedFailures: 0,
	}
}

//Start kicks off the executor workflow
func (e *ScaleIOExecutor) Start() <-chan struct{} {
	if err := e.subscribe(); err != nil {
		log.Errorln(err)
	}
	go e.handleEvents()
	return e.DoneChan
}

func (e *ScaleIOExecutor) stop() {
	close(e.Events)
}

func (e *ScaleIOExecutor) retrieveState() (*types.ScaleIOFramework, error) {
	log.Debugln("RetrieveState ENTER")
	url := e.Config.SchedulerURI + "/api/state"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorln("Error is HTTP NewRequest:", err)
		log.Debugln("RetrieveState LEAVE")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Error is HTTP Do:", err)
		log.Debugln("RetrieveState LEAVE")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1048576))
	if err != nil {
		log.Errorln("Error is IO ReadAll:", err)
		log.Debugln("RetrieveState LEAVE")
		return nil, err
	}

	log.Debugln("Body: ", string(body))

	var state types.ScaleIOFramework
	err = json.Unmarshal(body, &state)
	if err != nil {
		log.Errorln("Error is IO ReadAll:", err)
		log.Debugln("RetrieveState LEAVE")
		return nil, err
	}

	log.Debugln("RetrieveState Succeeded")
	log.Debugln("RetrieveState LEAVE")

	return &state, nil
}

func (e *ScaleIOExecutor) send(call *exec.Call) (*http.Response, error) {
	marshaler := jsonpb.Marshaler{
		EnumsAsInts:  true,
		EmitDefaults: false,
		Indent:       "  ",
		OrigName:     false,
	}
	strJSON, errJSON := marshaler.MarshalToString(call)
	if errJSON == nil {
		log.Debugln("JSON:\n", strJSON)
	} else {
		log.Debugln("Unable to marshal to JSON:", errJSON)
	}

	payload, err := proto.Marshal(call)
	if err != nil {
		log.Errorln("Failed to Marshal Protobuf:", err)
		return nil, err
	}

	resp, err := e.Client.Send(payload)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		msg := fmt.Sprint("StatusCode is not equal to StatusOK:", resp.StatusCode)
		log.Errorln(msg)
		return nil, errors.New(msg)
	}

	log.Infoln("StatusCode: StatusOK")
	return resp, nil
}

func (e *ScaleIOExecutor) subscribe() error {
	call := &exec.Call{
		FrameworkId: e.FrameworkID,
		ExecutorId:  e.ExecutorID,
		Type:        exec.Call_SUBSCRIBE.Enum(),
		Subscribe: &exec.Call_Subscribe{
			UnacknowledgedTasks:   []*mesos.TaskInfo{},
			UnacknowledgedUpdates: []*exec.Call_Update{},
		},
	}

	resp, err := e.send(call)
	if resp != nil {
		go e.qEvents(resp)
	}
	return err
}

func (e *ScaleIOExecutor) qEvents(resp *http.Response) {
	log.Debugln("qEvents ENTER")
	defer func() {
		resp.Body.Close()
		close(e.Events)
	}()

	dec := json.NewDecoder(resp.Body)
	for {
		event := new(exec.Event)
		if event == nil {
			log.Errorln("Event is nil")
		}
		log.Debugln("Waiting for Event")
		err := dec.Decode(event)
		log.Debugln("Received for Event")
		if err != nil {
			e.ObservedFailures = e.ObservedFailures + 1
			if e.ObservedFailures > numberOfEventFailuresBeforeExiting {
				log.Fatalln("Received too many invalid events. Lost communication to agent.")
				return
			}
			if err == io.EOF {
				log.Debugln("err == io.EOF")
				log.Debugln("qEvents LEAVE")
				return
			}
			log.Warnln("Error decoding event. Skip event. Err:", err)
			continue
		}

		//TODO fix this at some point. This is due to the RecordIO Format
		// RecordIO = <Message LENGTH>\n<Message of Size=LENGTH>
		log.Debugln("Adding Event:", event.String())
		e.ObservedFailures = 0
		e.Events <- event
	}
}

func (e *ScaleIOExecutor) handleEvents() {
	defer close(e.DoneChan)
	for ev := range e.Events {
		switch ev.GetType() {

		case exec.Event_SUBSCRIBED:
			sub := ev.GetSubscribed()
			log.Infoln("[EVENT] Executor SUBSCRIBED with id:", sub.GetExecutorInfo().GetExecutorId())
			log.Debugln("AgentInfo:", sub.GetAgentInfo().String())

		case exec.Event_LAUNCH:
			task := ev.GetLaunch().GetTask()
			log.Infoln("[EVENT] LAUNCH:", task.GetTaskId().GetValue())
			log.Debugln("Task:", task.String())

			err := e.sendUpdate(task, mesos.TaskState_TASK_RUNNING.Enum())
			if err != nil {
				log.Errorln("Failed while sending update:", err)
			}

			go func() {
				//TODO reminder not to rely on node.LastContact value until we add in pings
				errNode := RunExecutor(e.Config, e.retrieveState)
				if errNode != nil {
					myErr := e.sendUpdate(task, mesos.TaskState_TASK_ERROR.Enum())
					if myErr != nil {
						log.Errorln("Failed while sending update:", myErr)
					}
				}
			}()

		case exec.Event_ACKNOWLEDGED:
			log.Infoln("[EVENT] Received ACKNOWLEDGED:", ev.GetAcknowledged().String())

		case exec.Event_MESSAGE:
			log.Infoln("[EVENT] Received MESSAGE:", ev.GetMessage().String())

		case exec.Event_KILL:
			log.Infoln("[EVENT] Received KILL")

		case exec.Event_SHUTDOWN:
			log.Infoln("[EVENT] Received SHUTDOWN")
			e.stop()

		case exec.Event_ERROR:
			log.Infoln("[EVENT] Received ERROR")
			err := ev.GetError().GetMessage()
			log.Infoln(err)
		}
	}
}

func (e *ScaleIOExecutor) sendUpdate(task *mesos.TaskInfo, state *mesos.TaskState) error {
	log.Debugln("sendUpdate ENTER")
	log.Debugln("FrameworkID:", e.FrameworkID.String())
	log.Debugln("ExecutorID:", e.ExecutorID.String())
	log.Debugln("TaskId:", task.GetTaskId().String())
	log.Debugln("State:", state.String())

	call := &exec.Call{
		Type:        exec.Call_UPDATE.Enum(),
		FrameworkId: e.FrameworkID,
		ExecutorId:  e.ExecutorID,
		Update: &exec.Call_Update{
			Status: &mesos.TaskStatus{
				TaskId:     task.TaskId,
				ExecutorId: e.ExecutorID,
				State:      state,
				Source:     mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
				Uuid:       []byte(xplatform.GetInstance().Sys.GetUUID()),
			},
		},
	}

	log.Debugln("Call:", call.String())

	_, err := e.send(call)
	return err
}
