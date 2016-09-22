# ScaleIO Framework for Apache Mesos

![logo](img/logo.png)

The ScaleIO Framework deploys Dell EMC ScaleIO as a simplified task in Apache Mesos. All the required components to consume and provision storage volumes from a ScaleIO cluster are automatically deployed and configured on the Mesos Agents. This creates an automated mechanism to have a fully configured and reliable persistent storage solution for containers running on Apache Mesos. 

Test it out following the [Demo Guide](demo/README.md) using an AWS Cloud Formation Template and provided JSON files.

## Key Features
- Installs required components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- On-boards new Agent nodes with *"native"* access to ScaleIO volumes
- All Agent nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data using [REX-Ray](https://github.com/emccode/rexray) as a container runtime volume driver
- Additional storage can be added to the ScaleIO cluster to expand capacity

## Requirements
- Ubuntu 14.04 only (additional platforms to be made available in  future releases)
- Since Ubuntu support for ScaleIO is limited, this framework currently only supports ScaleIO version 2.0-5014.0.
- An existing 3-node or greater ScaleIO cluster using version 2.0-5014.0 must be running/provided. Primary, Secondary, and TieBreaker MDM are required for a minimal 3-node cluster.
- The ScaleIO cluster must have a Protection Domain and Storage Pool present which is capable of provisioning volumes.
- This Framework is implemented using HTTP APIs provided by Apache Mesos. This requires an Apache Mesos cluster running version 1.0 or higher.

**IMPORTANT NOTE:** In order to avoid the Mesos Agent nodes from rebooting, it is highly recommended that the Agent Nodes have kernel version 4.2.0-30 installed prior to launching the scheduler. You can do this by running the following command prior to bringing up the Mesos Agent service:
```
apt-get -y install linux-image-4.2.0-30-generic
```

## Launch the Framework on an Existing ScaleIO Cluster
No existing cluster? Follow the [Demo Guide](demo/README.md) using an AWS Cloud Formation Template and provided JSON files to get started.

If [MesosDNS](https://github.com/mesosphere/mesos-dns) or another service discovery application is not running in the Mesos cluster, create the following JSON to curl to Marathon:
```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc3/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc3/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR MESOS MASTER LEADER]:5050 -scaleio.clustername=[SCALEIO NAME] -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS] -executor.memory.non=256 -executor.cpu.non=0.5",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```

cURL to Marathon:
```
curl -k -XPOST -d @[SCALEIO JSON FILE] -H "Content-Type: application/json" [MARATHON IP ADDRESS]:8080/v2/apps
```

Example:
```
curl -k -XPOST -d @scaleio.json -H "Content-Type: application/json" 127.0.0.1:8080/v2/apps
```

## Road map / TBDs
The current release highlights the capabilities of combining Software Defined Storage with a platform that offers 2-layer scheduling. Subsequent versions will add significantly more features.

- Add CentOS/RHEL support
- Add CoreOS support
- Add ability to provision an entire ScaleIO cluster and include the MDM management nodes from initialization
- Allow more customization of the ScaleIO deployment
- Manage entire life cycle (upgrades, maintenance, etc) of all nodes in the ScaleIO cluster
- Manages health of a ScaleIO cluster by monitoring for critical events (Performance, Almost Full, etc)

## Support
Please file bugs and issues on the Github issues page for this project. This is to help keep track and document everything related to this repo. For general discussions and further support,  join the [{code} by Dell EMC Community](http://community.codedellemc.com/) slack team. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
