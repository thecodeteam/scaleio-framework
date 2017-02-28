# Experimental

Yea, yea, yea... I know. I am pushing the envelope on this one.

---

## Overview
This section list functionality that exists that is in the experimental stage. You
don't want to be using these features in production unless you like to live
dangerously. I am talking like riding in the middle of the
[El Toro Y](https://en.wikipedia.org/wiki/El_Toro_Y) during rush hour on a
unicycle without a helmet while holding a pair of scissors. Just put it this way,
you are probably not likely to survive. Use these features at your own risk!

## Provision the Entire ScaleIO Cluster

If you have read the documentation like I have, provision the entire ScaleIO
cluster can be quite complex. This Framework's intention has always been about
making that process and the monitoring of ScaleIO super easy. This is an early
look at that process which will provision the entire ScaleIO cluster from
scratch. That means starting from a bare Mesos cluster, the Framework will allow
you to stand up the 3 MDM (Primary, Secondary, and TieBreaker) nodes plus any
other data nodes without any existing ScaleIO infrastructure.

To use this functionality, you must have 3 Mesos Agent nodes with at least 4GB
of memory and 1.5 CPUs free and available for the MDM nodes. Currently, only the
3 MDM nodes configuration is supported. Rhe 5 MDM node configuration will be
added at a later date. That means the minimum ScaleIO configuration would 3 nodes
where each of the 3 Agents nodes needs to meet the memory and CPU prerequisites,
have at least 180GB hard disks on each Agent node, and ScaleIO volumes can be
provisioned and consumed on those 3 Agent nodes.

To launch and provision the entire ScaleIO cluster, you can use the Marathon API
and cURL the following JSON:

```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.1/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.1/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -experimental=true -rest.port=$PORT -uri=[IP ADDRESS FOR ANY MESOS MASTER]:5050",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```

As you can see, as we start turning over more of the operation of ScaleIO to the
Framework, launching the Framework becomes even easier.

Now the bad news, the main reason why you wouldn't want to use this option in
production is that this currently will not automatically place the MDM nodes
into protection domains or in locations on the Mesos cluster that when nodes fail
the ScaleIO MDM nodes can tolerate fault and continue to function without
interruption. The above example works for the Single Global Pool
[deployment option](/user-guide/deployment-strategies.md) and it *might* meet your
needs. There is functionality to use the Imperative Deployment where you can
declare before launching the Framework where you want the Primary, Secondary,
and TieBreaker nodes to get installed on; meaning you can place these nodes based
on a plan to handle fault tolerance. For example, each MDM node can be chosen to
get installed in their own unique racks. For more information on that option,
please contact us directly in the
[{code} community slack](http://community.codedellemc.com/).

## Self Remediating Functionality in the Cloud

This is a preview feature for the ScaleIO Framework which starts to pave the way
of the real value add for what this project really intends to provide. This
feature only currently supports AWS environments but can be expanded to support
additional clouds (like GCE, DigitalOcean, etc) in the future. When this
feature is enabled (disabled by default), the Framework will monitor the Storage
Pools to see if they are starting to fill up and reach its maximum capacity. This
can happen especially when you provision your ScaleIO volumes to be thin provisioned
(ie allow the volume to grow to its max capacity on demand).

When the Storage Pool starts to reach full, the Framework will detect the event
and create additional EBS volumes, attach them to the Mesos Agent nodes, and add
them to the Storage Pool thus expanding its maximum capacity. Features like this
will be added in the near future and this feature will continue to be refined
going forward.

To enable this feature in AWS clouds, you will need to provide an IAM access and
secret key that has the following
[permissions](http://libstorage.readthedocs.io/en/stable/user-guide/storage-providers/#aws-ebs)
in order to provision and use these EBS volumes. If we then wanted to provision
the entire ScaleIO cluster using the above feature and enable this Self
Remediating Cloud functionality, you can use the Marathon API and cURL the
following JSON:

```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.1/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.1/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -experimental=true -rest.port=$PORT -uri=[IP ADDRESS FOR ANY MESOS MASTER]:5050 -aws.accesskey=[ACCESS KEY] -aws.secretkey=[SECRET KEY]",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```
