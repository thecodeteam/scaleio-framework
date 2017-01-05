# Installation

How do I install this?

---

## Overview
ScaleIO-Framework is written in Go, so there are typically no dependencies that
must be installed alongside its single binary file. The Framework is deployed
via [Marathon](https://mesosphere.github.io/marathon/).

## Requirements
- Ubuntu 14.04 or CentOS7/RHEL7
- Since Ubuntu 14.04 support for ScaleIO is limited, this framework currently only supports ScaleIO version 2.0.1-2072 and kernel linux-image-4.4.0-38-generic. You can run `apt-get -y update && apt-get -y install linux-image-4.4.0-38-generic` for install that version.
- An **existing** 3-node or greater ScaleIO cluster using version 2.0.1-2072 must be running/provided. Primary, Secondary, and TieBreaker MDM are required for a minimal 3-node cluster.
- The ScaleIO cluster must have a Protection Domain and Storage Pool present which is capable of provisioning volumes.
- This Framework is implemented using HTTP APIs provided by Apache Mesos. This requires an Apache Mesos cluster running version 1.0 or higher with a compatible version of Marathon installed.

## Deployment Strategies
Before you deploy the Framework, you might want to think about what you want your
final fully deployed ScaleIO cluster to look like. Please take a look at the
[Deployment Strategies](/user-guide/deployment-strategies.md) section to
determine the best method for deploying your desired configuration.

## Deploying the latest version

The easiest way to deploy the latest version of the ScaleIO Framework is to
`curl` the [JSON file](sioframework-latest.json) below representing a task to
Marathon.

Before issuing the `curl` command, there are a couple of placeholder variables
you need to replace with real values. Those placeholders are enclosed in brackets.

```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0-rc2/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.3.0-rc2/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR ANY MESOS MASTER]:5050 -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS]",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```

Once those values have been replaced, issues the `curl` command like so:
```
curl -k -XPOST -d @sioframework-latest.json -H "Content-Type: application/json" [MARATHON IP ADDRESS]:8080/v2/apps
```

## Deploying a specific release

To deploy a specific version of the ScaleIO Framework we use a similar
`curl` command with a slightly modified version of the [JSON file](sioframework-v020.json)
to Marathon.

Before issuing the `curl` command, we need to fill in the placeholder variables
like before.

```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.2.0/scaleio-scheduler",
    "https://github.com/codedellemc/scaleio-framework/releases/download/v0.2.0/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR ANY MESOS MASTER]:5050 -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS]",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```

Once those values have been replaced, issues the `curl` command like so:
```
curl -k -XPOST -d @sioframework-v020.json -H "Content-Type: application/json" [MARATHON IP ADDRESS]:8080/v2/apps
```

## Build and install from source

The `ScaleIO-Framework` is also fairly simple to build from source. For more
information, please visit the [build-reference](/developer-guide/build-reference.md)
for more details.

To deploy a build from source, you will need access to an HTTP server in which
you can house the scaleio-scheduler and scaleio-executor for Marathon to download
from. After you have placed your custom build binaries, you can make the following
modification to the JSON file to deploy the Framework (replace `your.domain/path/to/your`
with the http location of your binaries):

```
{
  "id": "scaleio-scheduler",
  "uris": [
    "https://your.domain/path/to/your/scaleio-scheduler",
    "https://your.domain/path/to/your/scaleio-executor"
  ],
  "cmd": "chmod u+x scaleio-scheduler && ./scaleio-scheduler -loglevel=debug -rest.port=$PORT -uri=[IP ADDRESS FOR ANY MESOS MASTER]:5050 -scaleio.password=[SCALEIO GATEWAY PASSWORD] -scaleio.protectiondomain=[PROTECTION DOMAIN NAME] -scaleio.storagepool=[STORAGE POOL NAME] -scaleio.preconfig.primary=[MASTER MDM IP ADDRESS] -scaleio.preconfig.secondary=[SLAVE MDM IP ADDRESS] -scaleio.preconfig.tiebreaker=[TIEBREAKER MDM IP ADDRESS] -scaleio.preconfig.gateway=[GATEWAY IP ADDRESS]",
  "mem": 32,
  "cpus": 0.2,
  "instances": 1,
  "constraints": [
    ["hostname", "UNIQUE"]
  ]
}
```
