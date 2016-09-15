# ScaleIO Framework

Today software storage platforms are managed through a combination of manual/automated installs and runbooks to help with those “Day 2” maintenance operations typically done by hand. With this  [Framework](http://mesos.apache.org/documentation/latest/architecture/), deploying ScaleIO is as simple as launching any other task in Mesos. All software needed is rolled out and configured without any manual intervention and within a couple of minutes, [ScaleIO](https://www.emc.com/storage/scaleio/index.htm) is ready to provision volumes for all your container needs.

## Key Features
- Installs all components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- Onboards new Agent nodes with *by default* access to ScaleIO volumes
- All Agents nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data
- Additional storage can be added to the ScaleIO cluster to expand capacity

## Roadmap / TBDs
- Add CentOS/RHEL support
- Add CoreOS support
- Add for the ability to provision the entire ScaleIO cluster include the MDM management nodes from scratch
- Allow for more customization of the ScaleIO rollout
- Manage the entire lifecycle (upgrades, maintenance, etc) of all nodes in the ScaleIO cluster automatically
- Manages the health of a ScaleIO cluster by monitoring for critical events (Performance, Almost Full, etc)
- TBD

## Requirements
- Supports only Ubuntu 14.04 (additional platforms to be made available in the future)
- Since Ubuntu support for ScaleIO is limited, currently only supports ScaleIO version 2.0-5014.0.
- An existing 3-node ScaleIO cluster version 2.0-5014.0 must already running/provided (Primary, Secondary, TieBreaker MDM are configured and running)
- The ScaleIO cluster must already have a Protection Domain and Storage Pool present which is capable of provisioning volumes from.
- This Framework is implemented on the HTTP APIs provided by Apache Mesos. This requires an Apache Mesos cluster running version 1.0 or higher.

**IMPORTANT NOTE:** In order to avoid the Mesos Agent nodes from rebooting, it is highly recommended that the Agent Nodes have kernel version 4.2.0-30 installed prior to launching the scheduler. You can do this by running the following command prior to bringing up the Mesos Agent service ```mesos-slave```:
<pre>
apt-get -y install linux-image-4.2.0-30-generic
</pre>

## Supported ScaleIO Configurations
There are two supported configurations for your preexisting ScaleIO cluster:
- The first is a minimal ScaleIO configuration of 3 nodes in which each nodes has minimally a 180GB disk attached to each MDM (Pri, Sec, Tiebreaker) node and those disks comprise the Protection Domain and Storage Pool. The Mesos Agent nodes that are brought online will then create/mount/unmount volumes that are provisioned from the MDM nodes.
- In this configuration, the 3 ScaleIO MDM nodes and a second group/pool of servers that contribute attached disks to the Protection Domain and Storage Pool are separate servers. In this scenario, the Mesos Agent nodes that are brought online will then create/mount/unmount volumes that are provisioned from this second group/pool of servers.

The limited configuration support is mainly due to lack of management capabilities in this initial version and the reduced scope for this first release. These limitations will be expanded in future versions.

## Status
This first release highlights the capabilities of combining Software Defined Storage together with a Scheduling platform that offers 2 layer scheduling. Subsequent versions will add significantly more features towards making this framework open up new use-cases.

## Full Documentation
Continue reading the full documentation at [TBD](https://github.com/codedellemc/scaleio-framework).

## Launching the Framework
If you are not running [MesosDNS](https://github.com/mesosphere/mesos-dns) or some other service discovery application in your Mesos cluster, you can create the following JSON to curl to Marathon:
<pre>
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR MESOS MASTER LEADER]:5050 -scaleio.preconfig.primary=[IP ADDRESS FOR PRIMARY MDM] -scaleio.preconfig.secondary=[IP ADDRESS FOR SECONDARY MDM] -scaleio.preconfig.tiebreaker=[IP ADDRESS FOR TIEBREAKER] -executor.memory.non=256 -executor.cpu.non=0.5",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
</pre>

If you are using a service discovery application like [MesosDNS](https://github.com/mesosphere/mesos-dns), you can replace the Mesos Master IP with mesos.leader as shown below. It is **highly recommended** that you run a service discovery application in the event that the Mesos Master Leader dies. In the previous JSON example, the IP address is hardcoded to only talk to that individual Mesos Master.
<pre>
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc1/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=leader.mesos:5050 -scaleio.preconfig.primary=[IP ADDRESS FOR PRIMARY MDM] -scaleio.preconfig.secondary=[IP ADDRESS FOR SECONDARY MDM] -scaleio.preconfig.tiebreaker=[IP ADDRESS FOR TIEBREAKER] -executor.memory.non=256 -executor.cpu.non=0.5",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
</pre>

You can then cURL the JSON to Marathon by running the following command:
<pre>
curl -k -XPOST -d @[SCALEIO JSON FILE] -H "Content-Type: application/json" [MARATHON IP ADDRESS]:8080/v2/apps
</pre>

Example:
<pre>
curl -k -XPOST -d @scaleio.json -H "Content-Type: application/json" 127.0.0.1:8080/v2/apps
</pre>
