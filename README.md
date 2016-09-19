# ScaleIO Framework

Today software storage platforms are managed through a combination of manual/automated installs and runbooks to help with those “Day 2” maintenance operations typically done by hand. With this [Apache Mesos  Framework](http://mesos.apache.org/documentation/latest/architecture/), deploying ScaleIO is as simple as launching any other task in Mesos. All software needed is rolled out and configured without any manual intervention and within a couple of minutes, [ScaleIO](https://www.emc.com/storage/scaleio/index.htm) is ready to provision volumes for all your container needs.

## Key Features
- Installs all components on existing Mesos Agents to consume and provision ScaleIO storage volumes
- Onboards new Agent nodes with *"native"* access to ScaleIO volumes
- All Agents nodes are configured to be highly available so failed applications can be restarted on other Agent nodes while preserving their data
- Additional storage can be added to the ScaleIO cluster to expand capacity

## Requirements
- Supports Ubuntu 14.04 only (additional platforms to be made available in the future)
- Since Ubuntu support for ScaleIO is limited, this framework currently only supports ScaleIO version 2.0-5014.0.
- An existing 3-node ScaleIO cluster version 2.0-5014.0 must already running/provided (Primary, Secondary, TieBreaker MDM are configured and running)
- The ScaleIO cluster must already have a Protection Domain and Storage Pool present which is capable of provisioning volumes from.
- This Framework is implemented on the HTTP APIs provided by Apache Mesos. This requires an Apache Mesos cluster running version 1.0 or higher.

**IMPORTANT NOTE:** In order to avoid the Mesos Agent nodes from rebooting, it is highly recommended that the Agent Nodes have kernel version 4.2.0-30 installed prior to launching the scheduler. You can do this by running the following command prior to bringing up the Mesos Agent service ```mesos-slave```: ```apt-get -y install linux-image-4.2.0-30-generic```

## Getting Started Quickly (For Demo or Test)

If you are looking to get a demo or test environment up quickly, an [AWS CloudFormation template](https://github.com/codedellemc/scaleio-framework/raw/master/aws-demo/ScaleIO_Mesos_Testing_Cluster.json) is provided that installs a working ScaleIO cluster and Apache Mesos cluster up on Amazon. This template currently works in the **US-West-1 (aka N.California) region only**.

When you deploy this template, it uses six 't2.medium' instances, which, in the N.California region, cost $0.068 per hour to run, so the AWS EC2 compute usage for this cluster should run you about $9.78/day to keep running. The template provisions nine EBS volumes in total - six for the operating systems, and three 100-gigabyte volumes for the ScaleIO storage. Pricing on EBS volumes is a little harder to figure out, but it should be negligible.

The password for the ScaleIO admin is 'F00barbaz'. Other places a password might be used, same thing. The ScaleIO nodes are Redhat 7.X instances and the Mesos nodes are Ubuntu 14.04 instances and as such, the usernames used to log into those systems via ssh are ec2-user and ubuntu, respectively.

#### AWS Web GUI

Using the AWS web gui, in the services selection window:

- Click 'Cloudformation' under 'Management Tools'
- Click 'Create Stack', then 'Choose a Template'
- Click 'Upload file to S3', and upload the [.json](https://github.com/codedellemc/scaleio-framework/raw/master/aws-demo/ScaleIO_Mesos_Testing_Cluster.json) file from this repo
- Give the stack a name (like MesosFrameworkDemo or something)
- Select a keypair that exists in the N.California region!
- Click next, add tags if you want, or don't, then click next
- Review your settings and click 'Create' to create the stack

The stack will take approximately two minutes to build, and then the nodes should be available for ssh login.

#### Verify the ScaleIO Configuration

It is important that you determine the Master, Slave and TieBreaker MDM (or Metadata Manager) nodes as this information will be needed to launch the framework. I have noticed that ScaleIONode2 (with the Private IP address of 10.0.0.12) is typically the Primary MDM node. Log into that system by ssh'ing into that instance using the Public DNS or IP, then run the following commands:

Log into ScaleIO: ```scli --login --username admin --password F00barbaz```

Verify the MDM nodes: ```scli --query_cluster```

Sample output will look like this:
<pre>
Cluster:
    Mode: 3_node, State: Normal, Active: 3/3, Replicas: 2/2
Master MDM:
    Name: Manager2, ID: 0x1ed68652078a0ab1
        IPs: 10.0.0.12, Management IPs: 10.0.0.12, Port: 9011
        Version: 2.0.5014
Slave MDMs:
    Name: Manager1, ID: 0x44691e69695396d0
        IPs: 10.0.0.11, Management IPs: 10.0.0.11, Port: 9011
        Status: Normal, Version: 2.0.5014
Tie-Breakers:
    Name: Tie-Breaker1, ID: 0x569bc3812558b2d2
        IPs: 10.0.0.13, Port: 9011
        Status: Normal, Version: 2.0.5014
</pre>

#### Launch the Framework in this Demo/Test Environment

To help speed up getting started, a [scaleio.json](https://github.com/codedellemc/scaleio-framework/raw/master/aws-demo/scaleio.json) file has been placed in the aws-demo directory on GitHub. Open that file and correctly update the internal IP addresses of the Master, Slave, and TieBreaker MDM nodes. Typically, but not always, the defaults that are already contained within that file are the correct values:

<pre>
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc2/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc2/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=10.0.0.21:5050 -scaleio.password=F00barbaz -scaleio.protectiondomain=default -scaleio.storagepool=default -scaleio.preconfig.primary=10.0.0.12 -scaleio.preconfig.secondary=10.0.0.11 -scaleio.preconfig.tiebreaker=10.0.0.13 -scaleio.preconfig.gateway=10.0.0.11 -executor.memory.non=256 -executor.cpu.non=0.5",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
</pre>

Once the values are correct in the JSON file, you can then cURL the JSON to Marathon by running the following command:
<pre>
curl -k -XPOST -d @scaleio.json -H "Content-Type: application/json" [MESOS MASTER PUBLIC DNS/IP ADDRESS]:8080/v2/apps
</pre>

You can see the status of the ScaleIO framework by viewing the status in the (for now, a minimal) UI. To do that, perform the following steps:

- Open up the Marathon UI by opening the following URL: http://[MESOS MASTER PUBLIC DNS/IP ADDRESS]:8080
- Click the scaleio-scheduler in the Marathon UI
- The Private IP Address for the scheduler is listed, substitute that Private IP with the Agent's Public IP Address and keep the existing port values
- You should see a list of Mesos Agent nodes with the current status of ScaleIO rollout

## Launching the Framework on your Existing ScaleIO Install
If you are not running [MesosDNS](https://github.com/mesosphere/mesos-dns) or some other service discovery application in your Mesos cluster, you can create the following JSON to curl to Marathon:
<pre>
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc2/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.1.0-rc2/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR MESOS MASTER LEADER]:5050 -scaleio.clustername=[SCALEIO NAME] -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS] -executor.memory.non=256 -executor.cpu.non=0.5",
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

## Full Documentation
Continue reading the full documentation at [TBD](https://github.com/codedellemc/scaleio-framework).

## Roadmap / TBDs
- Add CentOS/RHEL support
- Add CoreOS support
- Add for the ability to provision the entire ScaleIO cluster include the MDM management nodes from scratch
- Allow for more customization of the ScaleIO rollout
- Manage the entire lifecycle (upgrades, maintenance, etc) of all nodes in the ScaleIO cluster automatically
- Manages the health of a ScaleIO cluster by monitoring for critical events (Performance, Almost Full, etc)
- TBD

## Status
This first release highlights the capabilities of combining Software Defined Storage together with a Scheduling platform that offers 2 layer scheduling. Subsequent versions will add significantly more features towards making this framework open up new use-cases.
