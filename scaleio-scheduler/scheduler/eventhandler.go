package scheduler

import (
	log "github.com/Sirupsen/logrus"

	sched "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/sched"
	common "github.com/codedellemc/scaleio-framework/scaleio-scheduler/scheduler/common"
)

func (s *ScaleIOScheduler) subscribed(event *sched.Event) {
	sub := event.GetSubscribed()
	s.Framework.Id = sub.FrameworkId
	log.Infoln("[EVENT] Received SUBSCRIBED. FrameworkID:", sub.FrameworkId.GetValue())
}

func (s *ScaleIOScheduler) offers(event *sched.Event) {
	offers := event.GetOffers().GetOffers()
	log.Infoln("[EVENT] Received", len(offers), "OFFERS")

	err := s.performNodeSelection(offers)
	if err != nil {
		log.Errorln("Failed to determine ScaleIO configuration:", err)
	}

	//setup executors
	for _, offer := range offers {
		log.Debugln("Offer:", offer)

		// account for executor resources if there's an executor already running on the slave
		if doesExecutorExistOnHost(offer) {
			log.Debugln("Skipping agent as it already has an executor on it. Decline offer.")
			message := generateDeclineCall(s.Config, offer)
			s.send(message)
			continue
		}

		//find node based on state
		node := common.FindScaleIONodeByHostname(s.Server.State.ScaleIO.Nodes, offer.GetHostname())
		if node == nil {
			log.Errorln("Unable to find node by Hostname:", offer.GetHostname())
			message := generateDeclineCall(s.Config, offer)
			s.send(message)
			continue
		}

		//generate accept call to launch executor
		message := generateAcceptCall(s.Config, offer, node)
		s.send(message)
	}
}

func (s *ScaleIOScheduler) rescind(event *sched.Event) {
	rescind := event.GetRescind()
	log.Infoln("[EVENT] Received RESCIND:", rescind.String())
}

func (s *ScaleIOScheduler) update(event *sched.Event) {
	update := event.GetUpdate().GetStatus()
	log.Infoln("[EVENT] Received STATUS:", update.String())

	ackRequired := len(update.Uuid) > 0
	if ackRequired {
		message := generateAcknowledgeCall(s.Framework, update)
		s.send(message)
	} else {
		log.Infoln("Not sending ACK, update is not from slave")
	}
}

func (s *ScaleIOScheduler) message(event *sched.Event) {
	message := event.GetMessage()
	log.Infoln("[EVENT] Received MESSAGE:", message.String())
}

func (s *ScaleIOScheduler) failure(event *sched.Event) {
	log.Infoln("[EVENT] Received FAILURE")
	fail := event.GetFailure()
	if fail.ExecutorId != nil {
		log.Infoln(
			"Executor", fail.ExecutorId.GetValue(), "terminated",
			"with status", fail.GetStatus(),
			"on agent", fail.GetAgentId().GetValue(),
		)
	} else {
		if fail.GetAgentId() != nil {
			log.Infoln("Agent", fail.GetAgentId().GetValue(), "failed")
		}
	}
}

func (s *ScaleIOScheduler) error(event *sched.Event) {
	err := event.GetError().GetMessage()
	log.Infoln("[EVENT] Received ERROR:", err)
}

func (s *ScaleIOScheduler) heartbeat(event *sched.Event) {
	log.Infoln("[EVENT] Received HEARTBEAT")
}
