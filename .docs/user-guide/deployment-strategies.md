# Deployment Strategies

What deployment strategies does the Framework offer?

---

## Overview
This section describes the deployment options that are available via the
Framework. Depending on what your end game is, you might prefer to one method
over the other. This guide will help you decide which option will best suit
your needs.

## Single Global Pool

I call this the "Just Make It Work" option. This is by far the easiest option to
get working because it requires zero work, but the issue is that unless you give
some thought in what resources are attached to what nodes, it may be difficult
to get this method of deployment to adhere to [best practices](h15148-emc-scaleio-deployment-guide.pdf)
for deploying ScaleIO.

When using this option, the Framework will create a single protection domain named
**default** and a single storage pool named **default** (unless you override the
values for -scaleio.protectiondomain and -scaleio.storagepool) out of any free
device (ie does not have a filesystem) attached to every Mesos Agent node that
is available.

It feels natural to just attach a disk (whether your on bare metal, in a virtualized
environment, or in AWS, etc) to every Mesos Agent node in your Mesos cluster and
call it a day. That configuration will give you a functioning ScaleIO cluster, but
it may not follow best practices. For a demo or test environment, this is a great
and simple configuration to deploy however. An example of how you might get this
option to work in a production environment is to earmark certain nodes to have
disks attached to certain Mesos Agent nodes (provides storage. these are the sds's)
and have other Mesos Agent nodes without any disks (consumes storage. these are
the sdc's).

## Imperative Deployment

I call this the "Deploy It Exactly Like This" option. Using
[Mesos Agent attributes](http://mesos.apache.org/documentation/latest/attributes-resources/),
you can declare a precise configuration for every Mesos Agent node in your
configuration.

To declare what disks contributes to what StoragePools and ProtectionDomains,
the following Mesos Agent nodes attributes must be defined: scaleio-sds-domains,
scaleio-sds-[domain], and scaleio-sds-[pool]. In the example below, the Mesos
attributes create: 1) a protection domain named `mydomain` with storage pool named
`mypool` consisting of the disk attached at `/dev/xvdf`, and 2) a second protection
domain named `myotherdomain` with storage pools named `saltwaterpool` consisting of the
disks attached at `/dev/xvdg` and `/dev/xvdh` and `freshwaterpool` consisting of the
disk attached at `/dev/xvdi`.

```
# cat /etc/mesos-slave/attributes/scaleio-sds-domains
mydomain,myotherdomain

#cat /etc/mesos-slave/attributes/scaleio-sds-mydomain
mypool

# cat /etc/mesos-slave/attributes/scaleio-sds-myotherdomain
saltwaterpool,freshwaterpool

# cat /etc/mesos-slave/attributes/scaleio-sds-mypool
/dev/xvdf

# cat /etc/mesos-slave/attributes/scaleio-sds-saltwaterpool
/dev/xvdg,/dev/xvdh

# cat /etc/mesos-slave/attributes/scaleio-sds-freshwaterpool
/dev/xvdi
```

To declare what Mesos Agent nodes consume storage out of which
ProtectionDomain/StoragePool combination, the following Mesos Agent nodes
attributes must be defined: scaleio-sdc-domains and scaleio-sdc-[domain]. In the
example below, the Mesos Agent node will consume/provision/etc volumes from the
protection domain named `mydomain` and storage pool named `mypool`:

```
# cat /etc/mesos-slave/attributes/scaleio-sdc-domains
mydomain

# cat /etc/mesos-slave/attributes/scaleio-sdc-mydomain1
mypool
```

There is a limitation currently in [rexray](https://github.com/codedellemc/rexray)
such that you can only consume/provision/etc a single ProtectionDomain/StoragePool
combination. If no sdc attributes are defined, the first attribute pair from the
sds attributes will be used. Else the default values based on the Configuration
flags will be used (-scaleio.protectiondomain and
-scaleio.storagepool).

## Declarative Deployment

Coming Soon.
