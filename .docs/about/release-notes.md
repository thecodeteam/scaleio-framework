# Release Notes

---

## Upgrading

To upgrade the ScaleIO-Framework to the latest version, simply redeploy the Framework using the Marathon JSON based on the release you want to target. For example, if you are currently running version 0.1.0 and want to upgrade to 0.2.0, upgrading would simply consist of curl'ing the following JSON to Marathon:

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

Use the following REST API to determine the currently installed version of the ScaleIO-Framework:

```
GET [SCHEDULER IP/FQDN]:[Marathon Dynamic Port]/version

{
  "versionint": 1,
  "versionstr": "0.2.0"
}
```

## Version 0.2.0 (2016/11/09)
ScaleIO Framework 0.2.0 introduces RHEL7/CentOS7 support and also refreshes the version of ScaleIO to versio 2.0.1 which is the latest as of writting this release..

### New Features
- Addressed Issue [#65](https://github.com/codedellemc/scaleio-framework/issues/65): RHEL7 and CentOS7 Support. Supports ScaleIO 2.0.1
- Addressed Issue [#91](https://github.com/codedellemc/scaleio-framework/issues/91): Updated Ubuntu14 to support ScaleIO 2.0.1. The CloudFormation template in the demo folder has also been updated to handle ScaleIO 2.0.1.
- Fixed Issue [#93](https://github.com/codedellemc/scaleio-framework/pull/93): The REX-ray configuration file that is created follows the suggested best practices.
- Added an intelligent reboot feature which will fix a reboot timing issue when the ScaleIO node that is running the scheduler is rebooted before other nodes have had the opportunity to contact it for the current state. I have not seen this happen yet, but there was certainly the possibility. That has been resolved now.

### Enhancements
- Massive restructuring to the executor. This was largely in part due to time to market release of 0.1.0. With addition of RHEL7/CentOS7 support, the project needed to be restructured to support multiple platforms in a maintainable fashion.
- Removed the following flags. This is largely in part to due differences with both DEB and RPM package managers (in command and operational behavior) between versions of platforms (ie RHEL6 vs RHEL7).
  - scaleio.deb.mdm
  - scaleio.deb.sds
  - scaleio.deb.sdc
  - scaleio.deb.lia
  - scaleio.deb.gw
  - scaleio.rpm.mdm
  - scaleio.rpm.sds
  - scaleio.rpm.sdc
  - scaleio.rpm.lia
  - scaleio.rpm.gw
- Added platform specific flags for the ScaleIO packages. This is largely in part due to each platform having a different DEB or RPM between platform versions. Added the following flags:
  - scaleio.ubuntu14.mdm
  - scaleio.ubuntu14.sds
  - scaleio.ubuntu14.sdc
  - scaleio.ubuntu14.lia
  - scaleio.ubuntu14.gw
  - scaleio.rhel7.mdm
  - scaleio.rhel7.sds
  - scaleio.rhel7.sdc
  - scaleio.rhel7.lia
  - scaleio.rhel7.gw
- Renamed the following 3 flags to match the CPU flags.
  - executor.memory.mdm -> executor.mem.mdm
  - executor.memory.no -> executor.mem.non
  - executor.memoryfactor -> executor.memfactor
- Added a new flag "Debug" to help with debugging the scheduler and executor. Among some of the things the debug flag does is prevent the reboot of the Mesos Agent node.
- Fixed Issues [#94](https://github.com/codedellemc/scaleio-framework/issues/94) and [#72]( https://github.com/codedellemc/scaleio-framework/issues/72): Documentation related changes.

### Bug Fixes
- Supports Mesos Master leader changes. Implements the Mesos Master redirect functionality to connect to a different master.
- Fixed an issue that sometimes caused REX-Ray not to start on reboot. Placed additional dependencies on ScaleIO scini driver.

## Version 0.1.0 (2016/09/28)

Initial Release
