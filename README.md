# ScaleIO Framework for Apache Mesos

![logo](img/logo.png)

The ScaleIO Framework deploys Dell EMC ScaleIO as a simplified task in Apache Mesos. All the required components to consume and provision storage volumes from a ScaleIO cluster are automatically deployed and configured on the Mesos Agents. This creates an automated mechanism to have a fully configured and reliable persistent storage solution for containers running on Apache Mesos.

## Key Features
- Installs required components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- On-boards new Agent nodes with *"native"* access to ScaleIO volumes
- All Agent nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data using [REX-Ray](https://github.com/emccode/rexray) as a container runtime volume driver
- Additional storage can be added to the ScaleIO cluster to expand capacity

## What it does
Container runtime schedulers need to be integrated with every aspect of available hardware resources, including persistent storage. When requesting resources for an application the scheduler gets offers for CPU, RAM and disk.

To be able to offer persistent storage in a scalable way, the ScaleIO Framework installs and configures all necessary ScaleIO, a Software-based Storage Platform, components along with all the "glue" to connect Mesos and ScaleIO to service applications requiring persistent storage.

## Documentation [![Docs](https://readthedocs.org/projects/scaleio-framework/badge/?version=stable)](http://scaleio-framework.readthedocs.org/en/stable/)
You will find complete documentation for ScaleIO-Framework at [scaleio-framework.readthedocs.org](http://scaleio-framework.readthedocs.org/en/stable/), including
[licensing](http://scaleio-framework.readthedocs.org/en/stable/about/license/) and
[support](http://scaleio-framework.readthedocs.org/en/stable/#getting-help) information.
Documentation provided at RTD is based on the latest stable build. The `/.docs`
directory in this repo will refer to the latest or specific commit.

## Road map / TBDs
The current release highlights the capabilities of combining Software Defined Storage with a platform that offers 2-layer scheduling. Subsequent versions will add significantly more features.

- Add CoreOS support
- Add ability to provision an entire ScaleIO cluster and include the MDM management nodes from initialization
- Allow more customization of the ScaleIO deployment
- Manage entire life cycle (upgrades, maintenance, etc) of all nodes in the ScaleIO cluster
- Manages health of a ScaleIO cluster by monitoring for critical events (Performance, Almost Full, etc)

## Support
Please file bugs and issues on the Github issues page for this project. This is to help keep track and document everything related to this repo. For general discussions and further support,  join the [{code} by Dell EMC Community](http://community.codedellemc.com/) slack team. The code and documentation are released with no warranties or SLAs and are intended to be supported through a community driven process.
