# Configuration

What are the Framework's configuration options?

---

## Overview

This page reviews how to configure the ScaleIO-Framework. We will first cover the
command line option and delving into the details of more advanced settings.

## Basic Command Line Options

Basic command line options available...

`-loglevel=[debug|info|warn|error]`  
Optional: Set the logging level for the application. Default: info

`-debug=[true|false]`  
Optional: Debug mode prevents the reboot so the logs dont get cycled. Default: false

`-rexray.branch=[stable|unstable|staged]`  
Optional: Which branch to grab the REX-Ray package from. Default: stable

`-rexray.version=[specific version|latest]`  
Optional: Which version to install from the provided branch. You can provide a
specific version such as 0.6.3. The default is the last build. Default: latest

`-isolator.binary=[specific version]`  
Optional: The URL for which Mesos Module DVDI to install. Default: latest supported

`-rest.address=`  
Optional: *Highly recommend not to modify* Mesos scheduler REST API address.
Default: Dynamically calculated.

`-rest.port=<port>`
Required: *Highly recommend using the dynamic Marathon port* Mesos scheduler
REST API port.

`-uri=<mesos master uri>`
Required: Mesos scheduler API URL. This is the Mesos Master HTTP API endpoint.

`-executor.cpu.factor=[float value]`  
Optional: Fudge factor for effective CPU available. This allows overhead/reserve.
A value of 1.0 means none of the CPU has been reserved. Default: 1.0

`-executor.mem.factor=[float value]`  
Optional: Fudge factor for effective memory available. This allows overhead/reserve.
A value of 1.0 means none of the memory has been reserved. Default: 1.0

`-store.type=[consul|etcd|zk|boltdb]`  
Optional: The type of keyvalue store to use. Default: zk

`-store.uri=[uri to connect to the key store based on store.type]`  
Optional: Store URI to connect with. When the value is an "empty string", the
framework will dynamically determine the zookeeper endpoints.
Default: "empty string"

`-scaleio.clustername=[cluster name]`  
Optional: ScaleIO Cluster Name. Default: scaleio

`-scaleio.clusterid=`  
Optional: ScaleIO Cluster ID. The priority of this flag supersedes scaleio.clustername
if used. Default: "empty string"

`-scaleio.lbgateway=[uri for the scaleio gateway]`  
Optional: Great to use if your ScaleIO gateway is clustered and load balanced.
If the value is "empty string", it is assumed that the ScaleIO is on the primary
MDM node. Default: "empty string"

`-scaleio.protectiondomain=[default pd name]`  
Optional: The default ScaleIO Protection Domain Name used. Default: default

`-scaleio.storagepool=[default sp name]`  
Optional: The default ScaleIO StoragePool Name used. Default: default

`-scaleio.password=[password]`  
Required: ScaleIO Admin Password. This is used to install all ScaleIO packages
that require a ScaleIO admin password, the password to access the ScaleIO gateway,
and etc. Default: Scaleio123

`-scaleio.preconfig.primary=<fqdn/ip of pri MDM node>`  
Required: FQDN or IP of the pre-configured Pri MDM Node. Requires Sec and TB MDM
nodes to be pre-configured

`-scaleio.preconfig.secondary=<fqdn/ip of sec MDM node>`  
Required: FQDN or IP of the pre-configured Sec MDM Node. Requires Pri and TB MDM
nodes to be pre-configured

`-scaleio.preconfig.tiebreaker=<fqdn/ip of sec MDM node>`  
Required: FQDN or IP of the pre-configured Tiebreaker MDM Node. Requires Pri and
Sec MDM nodes to be pre-configured

`-scaleio.preconfig.gateway=[fqdn/ip of ScaleIO gateway node]`  
Optional: Used to set a separate ScaleIO Gateway node. If the value is "empty string",
the Primary MDM node is assumed to have the Gateway installed on it.
Default: "empty string"

`-scaleio.ubuntu14.mdm=[URL for DEB]`  
Optional: ScaleIO MDM Package for Ubuntu 14.04. Default: package tested in the release

`-scaleio.ubuntu14.sds=[URL for DEB]`  
Optional: ScaleIO SDS Package for Ubuntu 14.04. Default: package tested in the release

`-scaleio.ubuntu14.sdc=[URL for DEB]`  
Optional: ScaleIO SDC Package for Ubuntu 14.04. Default: package tested in the release

`-scaleio.ubuntu14.lia=[URL for DEB]`  
Optional: ScaleIO LIA Package for Ubuntu 14.04. Default: package tested in the release

`-scaleio.ubuntu14.gw=[URL for DEB]`  
Optional: ScaleIO Gateway Package for Ubuntu 14.04. Default: package tested in the release

`-scaleio.centos7.mdm=[URL for RPM]`  
`-scaleio.rhel7.mdm=[URL for RPM]`
Optional: ScaleIO MDM Package for RHEL7/CentOS7. Currently RHEL7 and CentOS7 are the same.
Default: package tested in the release

`-scaleio.centos7.sds=[URL for RPM]`
`-scaleio.rhel7.sds=[URL for RPM]`
Optional: ScaleIO SDS Package for RHEL7/CentOS7. Currently RHEL7 and CentOS7 are the same.
Default: package tested in the release

`-scaleio.centos7.sdc=[URL for RPM]`
`-scaleio.rhel7.sdc=[URL for RPM]`
Optional: ScaleIO SDC Package for RHEL7/CentOS7. Currently RHEL7 and CentOS7 are the same.
Default: package tested in the release

`-scaleio.centos7.lia=[URL for RPM]`
`-scaleio.rhel7.lia=[URL for RPM]`
Optional: ScaleIO LIA Package for RHEL7/CentOS7. Currently RHEL7 and CentOS7 are the same.
Default: package tested in the release

`-scaleio.centos7.gw=[URL for RPM]`
`-scaleio.rhel7.gw=[URL for RPM]`
Optional: ScaleIO Gateway Package for RHEL7/CentOS7. Currently RHEL7 and CentOS7 are the same.
Default: package tested in the release

## Advanced Command Line Options

Not going to lie... some of these command line option descriptions will be left
intentional vague so you don't shoot yourself in the foot. No point in hiding them
because you can just run a help on the command line. So in the interest in full
disclosure, here they are.

Most of these options are highly recommended *NOT* to modify and/or use.

`-experimental=[true|false]`  
Optional: Sets the application to experimental mode. Default: false

`-executor.altpath=[http path]`  
Optional: This is to override the default location of the scaleio-executor binary.
Default: "empty string"

`-executor.cpu.mdm=[float value]`  
Optional: CPU resources consumed by ScaleIO MDM software.
These are either the primary, secondary, or tiebreaker MDM nodes. Modifying
these values may not earmark actually CPU consumed by the ScaleIO software. Default: 1.5

`-executor.cpu.non=[float value]`  
Optional: CPU resources consumed by ScaleIO Non-MDM software.
These are data nodes only. Modifying these values may not earmark actually CPU
consumed by the ScaleIO software. Default: 0.5

`-executor.mem.mdm=[int value]`  
Optional: Memory resources (MB) consumed by ScaleIO MDM software.
These are either the primary, secondary, or tiebreaker MDM nodes. Modifying
these values may not earmark actually memory consumed by the ScaleIO software.
Default: 3072

`-executor.mem.non=[int value]`  
Optional: Memory resources (MB) consumed by ScaleIO MDM software.
These are data nodes only. Modifying these values may not earmark actually memory
consumed by the ScaleIO software. Default: 512

`-user=[user to run the process under]`  
Optional: The User account the framework is running under. Root is required in
order to run administrative functions like installation of packages. Default: root

`-hostname=[hostname where the scheduler is running]`  
Optional: The Hostname where the framework runs. This is only used for setting up
a mock scheduler. Default: dynamically discovered

`-role=[role name]`  
Optional: *Currently not supported* Framework role to register with the Mesos master.
This will allow for multiple ScaleIO clusters to run in the same Mesos cluster. Default: scaleio

`-scaleio.apiversion=[ScaleIO API Version]`  
Optional: ScaleIO API Version. Matches the API version of the ScaleIO software
the framework installs. Default: 2.0

`-store.delete=<true|false>`  
Optional: Helper function that deletes ScaleIO Framework Key/Value Store. Default: false

`-store.dump=<true|false>`  
Optional: Helper function that dumps ScaleIO Framework Key/Value Store. Default: false

`-store.add.key=<key to add to store>`  
Optional: Modify a select store key. Default: "empty string"

`-store.add.value=<value to add to the store>`  
Optional: Set the values to the key provided by store.key. Default: "empty string"

`-store.del.key=<key to delete>`  
Optional: Delete a select store key. Default: "empty string"
