# ScaleIO-Framework for Apache Mesos

![ScaleIO-Framework](images/logo.png)

## Overview
The `ScaleIO Framework` deploys Dell EMC ScaleIO as a simplified task in Apache
Mesos. All the required components to consume and provision storage volumes from
a ScaleIO cluster are automatically deployed and configured on the Mesos Agents.
This creates an automated mechanism to have a fully configured and reliable
persistent storage solution for containers running on Apache Mesos.

## Key Features
- Installs required components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- On-boards new Agent nodes with *"native"* access to ScaleIO volumes
- All Agent nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data using [REX-Ray](https://github.com/codedellemc/rexray) as a container runtime volume driver
- Additional storage can be added to the ScaleIO cluster to expand capacity

## What it does
Container runtime schedulers need to be integrated with every aspect of available
hardware resources, including persistent storage. When requesting resources for
an application the scheduler gets offers for CPU, RAM and disk.

To be able to offer persistent storage in a scalable way, the ScaleIO Framework
installs and configures all necessary ScaleIO, a Software-based Storage Platform,
components along with all the "glue" to connect Mesos and ScaleIO to service
applications requiring persistent storage.

## Framework workflow

What does the ScaleIO Framework really do under the covers? Up to this point,
its been stated that the Framework automates the lifecycle of ScaleIO and any
related components required to provision external persistent storage in a "run
it and forget it" fashion, but what does that really mean?

The ScaleIO Framework performs the following steps on deployment. It installs
and configures:

1. Any dependencies required for ScaleIO to run. This is done via apt-get or yum.
2. The ScaleIO SDS (or Server) package. This is the service that takes designated disks (physical or virtual) and contributes them to the ScaleIO cluster.
3. The ScaleIO SDC (or Client) package. This is the service that provides access to ScaleIO volumes created within the ScaleIO cluster.
4. [REX-Ray](https://github.com/codedellemc/rexray) which provides Mesos the ability to provision external storage for tasks that are backed by Docker containers.
5. [mesos-module-dvdi](https://github.com/emccode/mesos-module-dvdi) and [DVDCLI](https://github.com/emccode/dvdcli) which provides Mesos the ability to provision external storage for tasks that using the Mesos Universal Containerizer. This includes any configuration required on the Mesos Agent nodes.

## Installation and Configuration
Installation is fairly simple as it only requires a simple `curl` command to
deploy the Framework. There are some prerequisites that are required however. You
can find more information on the [Installation Page](/user-guide/installation.md).
There are various [configuration options](/user-guide/configuration.md) available
as well that will allow you to tune and customize your deployment of ScaleIO.

## Hello ScaleIO-Framework
Looking for a video demo or to get some hands on experience with demo environment?
Take a look in the [Users Guide](/user-guide/demo.md) section for instructions
to quickly prop up an environment quickly.

## Getting Help
Having issues? No worries, let's figure it out together.

### GitHub and Slack
If a little extra help is needed, please don't hesitate to use
[GitHub issues](https://github.com/codedellemc/scaleio-framework/issues) or join
the active conversation on the
[{code} Community Slack Team](http://community.codedellemc.com/) in the #mesos
channel
