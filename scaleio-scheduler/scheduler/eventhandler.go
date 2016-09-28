package scheduler

import (
	log "github.com/Sirupsen/logrus"

	"github.com/codedellemc/scaleio-framework/scaleio-scheduler/config"
	sched "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/sched"
	mesos "github.com/codedellemc/scaleio-framework/scaleio-scheduler/mesos/v1"
)

func (s *ScaleIOScheduler) subscribed(event *sched.Event) {
	sub := event.GetSubscribed()
	s.Framework.Id = sub.FrameworkId
	log.Infoln("[EVENT] Received SUBSCRIBED. FrameworkID:", sub.FrameworkId.GetValue())
}

func (s *ScaleIOScheduler) offers(event *sched.Event) {
	offers := event.GetOffers().GetOffers()
	log.Infoln("[EVENT] Received", len(offers), "OFFERS")

	//TODO do a better job at selecting the MDM nodes if not preconfigured
	//loop through offers and build "State" object.
	//TODO retrieve pri, sec, tb from zookeeper. Use libkv.
	for _, offer := range offers {
		if doesExecutorExistOnHost(offer) {
			log.Debugln("Skipping node as it already has an executor on it.")
			continue
		}

		node := findScaleIONodeByHostname(s.Server.State.ScaleIO.Nodes, offer.GetHostname())
		if node != nil {
			log.Errorln("Found existing node by Hostname:", offer.GetHostname())
			continue
		}

		ID := s.getNextNodeID()
		persona := s.getNodeType(ID)
		log.Debugln("Creating new metadata node. ID:", ID, "Persona:", persona)
		s.Server.State.ScaleIO.Nodes = append(s.Server.State.ScaleIO.Nodes,
			prepareScaleIONode(offer, persona, ID))
	}
	//TODO save pri, sec, tb to zookeeper. Use libkv.

	//setup executors
	for _, offer := range offers {
		log.Debugln("Offer:", offer)

		cpuResources := filterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "cpus"
		})
		cpus := 0.0
		for _, res := range cpuResources {
			cpus += res.GetScalar().GetValue()
		}

		memResources := filterResources(offer.Resources, func(res *mesos.Resource) bool {
			return res.GetName() == "mem"
		})
		mems := 0.0
		for _, res := range memResources {
			mems += res.GetScalar().GetValue()
		}

		log.Infoln("Received Offer <", offer.GetId().GetValue(), ",", offer.GetHostname(),
			"> with cpus=", cpus, " mem=", mems)

		// account for executor resources if there's an executor already running on the slave
		if doesExecutorExistOnHost(offer) {
			log.Debugln("Skipping agent as it already has an executor on it. Decline offer.")
			message := generateDeclineCall(s.Config, offer)
			s.send(message)
			continue
		}

		//find node based on state
		node := findScaleIONodeByHostname(s.Server.State.ScaleIO.Nodes, offer.GetHostname())
		if node == nil {
			log.Errorln("Unable to find node by Hostname:", offer.GetHostname())
			continue
		}

		//if node is an MDM node
		if IsNodeAnMDMNode(node) {
			if config.CPUPerMdmExecutor >= (cpus*s.Config.ExecutorCPUFactor) ||
				config.MemPerMdmExecutor >= (mems*s.Config.ExecutorMemoryFactor) {
				log.Warnln("Does not have enough resources to install ScaleIO (MDM) on node",
					offer.GetId().GetValue(), ",", offer.GetHostname())
				continue
			}
		} else {
			if config.CPUPerNonExecutor >= cpus || config.MemPerNonExecutor >= mems {
				log.Warnln("Does not have enough resources to install ScaleIO (Non-MDM) on node",
					offer.GetId().GetValue(), ",", offer.GetHostname())
				continue
			}
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
