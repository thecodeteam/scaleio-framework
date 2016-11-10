# ScaleIO Framework for Apache Mesos

![logo](img/logo.png)

The ScaleIO Framework deploys Dell EMC ScaleIO as a simplified task in Apache Mesos. All the required components to consume and provision storage volumes from a ScaleIO cluster are automatically deployed and configured on the Mesos Agents. This creates an automated mechanism to have a fully configured and reliable persistent storage solution for containers running on Apache Mesos.

Test it out following the [Demo Guide](demo/README.md) using an AWS Cloud Formation Template and provided JSON files. Watch the [YouTube Demo Video](https://youtu.be/tt6qhEkeVOQ?list=PLbssOJyyvHuWiBQAg9EFWH570timj2fxt) to see it in action.

## Key Features
- Installs required components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- On-boards new Agent nodes with *"native"* access to ScaleIO volumes
- All Agent nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data using [REX-Ray](https://github.com/emccode/rexray) as a container runtime volume driver
- Additional storage can be added to the ScaleIO cluster to expand capacity

## Requirements
- Ubuntu 14.04 or CentOS7/RHEL7
- Since Ubuntu support for ScaleIO is limited, this framework currently only supports ScaleIO version 2.0.1-2072.
- An existing 3-node or greater ScaleIO cluster using version 2.0.1-2072 must be running/provided. Primary, Secondary, and TieBreaker MDM are required for a minimal 3-node cluster.
- The ScaleIO cluster must have a Protection Domain and Storage Pool present which is capable of provisioning volumes.
- This Framework is implemented using HTTP APIs provided by Apache Mesos. This requires an Apache Mesos cluster running version 1.0 or higher.

**IMPORTANT NOTE for Ubuntu 14.04:** In order to avoid the Mesos Agent nodes from rebooting, it is highly recommended that the Agent Nodes have kernel version 4.2.0-30 installed prior to launching the scheduler. You can do this by running the following command prior to bringing up the Mesos Agent service:
```
apt-get -y update && apt-get -y install linux-image-4.4.0-38-generic
```

## Launch the Framework on an Existing ScaleIO Cluster
No existing cluster? Follow the [Demo Guide](demo/README.md) using an AWS Cloud Formation Template and provided JSON files to get started.

If [MesosDNS](https://github.com/mesosphere/mesos-dns) or another service discovery application is not running in the Mesos cluster, create the following JSON to curl to Marathon:
```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.2.0/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.2.0/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR MESOS MASTER]:5050 -scaleio.clustername=[SCALEIO NAME] -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS]",
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

## Under the Covers
What does the ScaleIO Framework really do under the covers? Up to this point, its been stated that the Framework automates the lifecycle of ScaleIO and any related components required to provision external persistent storage in a "run it and forget it" fashion, but what does that really mean?

The ScaleIO Framework performs the following steps on deployment. It installs and configures:

1. Any dependencies required for ScaleIO to run. This is done via apt-get or yum.
2. The ScaleIO SDS (or Server) package. This is the service that takes designated disks (physical or virtual) and contributes them to the ScaleIO cluster.
3. The ScaleIO SDC (or Client) package. This is the service that provides access to ScaleIO volumes created within the ScaleIO cluster.
4. [REX-Ray](https://github.com/codedellemc/rexray) which provides Mesos the ability to provision external storage for tasks that are backed by Docker containers.
5. [mesos-module-dvdi](https://github.com/emccode/mesos-module-dvdi) and [DVDCLI](https://github.com/emccode/dvdcli) which provides Mesos the ability to provision external storage for tasks that using the Mesos Universal Containerizer. This includes any configuration required on the Mesos Agent nodes.

## Road map / TBDs
The current release highlights the capabilities of combining Software Defined Storage with a platform that offers 2-layer scheduling. Subsequent versions will add significantly more features.

- Add CoreOS support
- Add ability to provision an entire ScaleIO cluster and include the MDM management nodes from initialization
- Allow more customization of the ScaleIO deployment
- Manage entire life cycle (upgrades, maintenance, etc) of all nodes in the ScaleIO cluster
- Manages health of a ScaleIO cluster by monitoring for critical events (Performance, Almost Full, etc)

## Support
Please file bugs and issues on the Github issues page for this project. This is to help keep track and document everything related to this repo. For general discussions and further support,  join the [{code} by Dell EMC Community](http://community.codedellemc.com/) slack team. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
