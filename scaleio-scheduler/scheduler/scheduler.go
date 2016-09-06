package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	jsonpb "github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/client"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	sched "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/sched"
	mesos "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/v1"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/server"
	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/types"
)

//ScaleIOScheduler represents a Mesos scheduler
type ScaleIOScheduler struct {
	Config *config.Config

	Framework     *mesos.FrameworkInfo
	TasksLaunched int

	Server *server.RestServer
	Client *client.Client

	Events chan *sched.Event
	//Offers   chan *sched.Offer
	//Status   chan *sched.Status
	DoneChan chan struct{}
}

//NewScaleIOScheduler returns a pointer to new Scheduler
func NewScaleIOScheduler(cfg *config.Config) *ScaleIOScheduler {
	return &ScaleIOScheduler{
		Config:        cfg,
		Client:        client.New(cfg.MasterREST, "/api/v1/scheduler"),
		Server:        server.NewRestServer(cfg),
		Framework:     prepareFrameworkInfo(cfg),
		TasksLaunched: 1,
		Events:        make(chan *sched.Event),
		DoneChan:      make(chan struct{}),
	}
}

//Start starts the scheduler and subscribes to event stream
// returns a channel to wait for completion.
func (s *ScaleIOScheduler) Start() <-chan struct{} {
	if err := s.subscribe(); err != nil {
		log.Errorln("Failed to subscribe:", err)
	}
	go s.handleEvents()
	return s.DoneChan
}

func (s *ScaleIOScheduler) stop() {
	close(s.Events)
}

func (s *ScaleIOScheduler) send(call *sched.Call) (*http.Response, error) {
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

	resp, err := s.Client.Send(payload)
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

// Subscribe subscribes the scheduler to the Mesos cluster.
// It keeps the http connection opens with the Master to stream
// subsequent events.
func (s *ScaleIOScheduler) subscribe() error {
	call := &sched.Call{
		Type: sched.Call_SUBSCRIBE.Enum(),
		Subscribe: &sched.Call_Subscribe{
			FrameworkInfo: s.Framework,
		},
	}

	resp, err := s.send(call)
	if resp != nil {
		go s.qEvents(resp)
	}
	return err
}

func (s *ScaleIOScheduler) qEvents(resp *http.Response) {
	log.Debugln("qEvents ENTER")
	defer func() {
		resp.Body.Close()
		close(s.Events)
	}()

	dec := json.NewDecoder(resp.Body)
	for {
		event := new(sched.Event)
		if event == nil {
			log.Errorln("Event is nil")
		}
		log.Debugln("Waiting for Event")
		err := dec.Decode(event)
		log.Debugln("Received for Event")
		if err != nil {
			if err == io.EOF {
				log.Debugln("err == io.EOF")
				log.Debugln("qEvents LEAVE")
				return
			}

			//TODO fix this at some point. This is due to the RecordIO Format
			// RecordIO = <Message LENGTH>\n<Message of Size=LENGTH>
			log.Warnln("Unable to decode event. Skip event. Err:", err)
			continue
		}
		log.Debugln("Adding Event:", event.String())
		s.Events <- event
	}
}

func (s *ScaleIOScheduler) handleEvents() {
	defer close(s.DoneChan)
	for event := range s.Events {
		switch event.GetType() {

		case sched.Event_SUBSCRIBED:
			s.subscribed(event)

		case sched.Event_OFFERS:
			s.offers(event)

		case sched.Event_RESCIND:
			s.rescind(event)

		case sched.Event_UPDATE:
			s.update(event)

		case sched.Event_MESSAGE:
			s.message(event)

		case sched.Event_FAILURE:
			s.failure(event)

		case sched.Event_ERROR:
			s.error(event)

		case sched.Event_HEARTBEAT:
			s.heartbeat(event)
		}
	}
}

func (s *ScaleIOScheduler) getNextNodeID() int {
	ID := s.TasksLaunched
	s.TasksLaunched = s.TasksLaunched + 1
	return ID
}

func (s *ScaleIOScheduler) getNodeType(ID int) int {
	if s.Config.PrimaryMdmAddress != "" {
		log.Debugln("Pre-Configured MDM Nodes provided. All nodes are Data Nodes.")
		return types.PersonaNode
	}

	persona := types.PersonaNode

	switch ID {
	case 1:
		persona = types.PersonaMdmPrimary
	case 2:
		persona = types.PersonaMdmSecondary
	case 3:
		persona = types.PersonaTb
	}

	return persona
}
