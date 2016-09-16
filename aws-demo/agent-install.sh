#!/bin/bash

#install mesos
# This script is an installation aid for installing Mesos on a pre-existing
# AWS Ubuntu instance.
# A t2.small is the recommended minimum instance type (2 CPU+2GB memory).
# This is a aid for developers using an AWS environment for testing.
# This script is not used during a build, or in production deployments.

# Make sure only root can run our script
if [[ $EUID -ne 0 ]]; then
   echo "This script must be run as root" 1>&2
   exit 1
fi

# Validate parameters
if [ $# -ne 2 ]; then
    echo $0: usage: mesos-install master-hostname my-hostname
    echo example: mesos-install ec2-WW-XX-YY-ZZ.us-west-2.compute.amazonaws.com ec2-WW-XX-YY-ZZ.us-west-2.compute.amazonaws.com
    exit 1
fi

mymasterhostname=$1
mymasterip=$(hostname -i)
myhostname=$2
myip=$(hostname -i)

dnslookup1=$(dig +short $mymasterhostname)
if [ "$dnslookup1" != "$mymasterip" ]; then
    echo hostname $1 is invalid - it does not resolve to the master IP
    exit 1
fi
dnslookup2=$(dig +short $myhostname)
if [ "$dnslookup2" != "$myip" ]; then
    echo hostname $2 is invalid - it does not resolve to the ip of this host
    exit 1
fi

hostnamectl set-hostname $myhostname

# Add Mesosphere repositories
apt-key adv --keyserver keyserver.ubuntu.com --recv E56151BF
DISTRO=$(lsb_release -is | tr '[:upper:]' '[:lower:]')
CODENAME=$(lsb_release -cs)
echo "deb http://repos.mesosphere.io/${DISTRO} ${CODENAME} main" | sudo tee /etc/apt/sources.list.d/mesosphere.list

# mesos and marathon versions
mesosver=1.0.0-2.0.89.ubuntu1404

# Update the kernel (required for ubuntu ScaleIO only)
apt-get -y install linux-image-4.2.0-30-generic

# Add Java repo
add-apt-repository -y ppa:webupd8team/java
apt-get -y update

# Install Oracle Java - note this will trigger a license "popup"
apt-get -y install oracle-java8-installer
apt-get install oracle-java8-set-default

apt-get install mesos=$mesosver
apt-mark hold mesos

# Write Mesos master configuration
echo "zk://${mymasterip}:2181/mesos" > /etc/mesos/zk

# Write Mesos agent configuration
echo "$myhostname" > /etc/mesos-slave/hostname
echo "$myip" > /etc/mesos-slave/ip
echo "mesos" > /etc/mesos-slave/containerizers
echo "5mins" > /etc/mesos-slave/executor_registration_timeout
echo "/tmp/mesos" > /etc/mesos-slave/work_dir

service zookeeper stop
echo manual | sudo tee /etc/init/zookeeper.override

service mesos-master stop
echo manual | sudo tee /etc/init/mesos-master.override

service mesos-slave start

reboot

exit 0
