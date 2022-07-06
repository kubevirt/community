# Overview
This feature enables support of live-migrate for VMs with pod network in bridge and macvtap modes.

## Motivation
Masquerade and slirp are not so performant compared to bridge and macvtap modes.
We want to use the less latency modes with the opportunity to live-migrate the virtual machines.

## Goals
Provide a live-migration feature for VMs running with pod netowork connected in bridged mode and macvtap mode.

## Non Goals
The live-migration of virtual machines between the pods with different MAC address will invoke NIC reattaching procedure. This might affect some applications which are binding to the specific interface inside the VM.

The live-migration of virtual machines between the pods with same MAC address will invoke the procedure to link down and up for the VM to renew DHCP lease with IP address and routes inside the VM. This is less destructive operation, but still may affect some workloads.

## Definition of Users
Everyone who use bridge to bind pod network may want to live-migrate created VMs.

## User Stories
* As a user / admin, I want to have an opportunity for live-migration of a VM with bridged pod-network.

## Repos
- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

To add two additional methods into live-migration procedure:
- If MAC address changed: detach / attach interface after live migration to have correct MAC address set
- If MAC address is not changed: link down / link up interface to force VM request new IP address and routes from DHCP

## API Examples
There are no API changes from the user side.

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
This change adds new logic also for multus connected networks.
Because, multus can also be used to bind standard CNIs (eg. flannel), which does allow to preserve IP over the nodes.
The live-migration of such VMs were not handle network reconfiguration before, now it will be handled by the same procedure.

## Functional Testing Approach
- Create two VMs: client and server in bridge mode
- Wait for lunch
- Get IP address for server VM
- Run ping from client to server
- live-migrate client VM
- Run ping from client to server

# Implementation Phases
Implementation is already prepared as single pull request:
https://github.com/kubevirt/kubevirt/pull/7768
