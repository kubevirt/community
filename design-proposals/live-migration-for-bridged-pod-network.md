# Live migration for bridged pod network

Author: Andrei Kvapil \<kvapss@gmail.com\>

## Overview
This feature enables support of live-migrate for VMs with pod network in bridge mode.

## Motivation
Masquerade and slirp are not so performant compared to bridge mode.
We want to use the less latency modes with the opportunity to live-migrate the virtual machines.

## Goals
Provide a live-migration feature for VMs running with pod network connected in bridged mode.

## Non Goals
Live migration with network reconfiguration is not seamless, sometimes it can take some time for applying a new settings.

## Definition of Users
Everyone who use bridge to bind pod network may want to live-migrate created VMs.

## User Stories
* As a user / admin, I want to have an opportunity for live-migration of a VM with bridged pod-network.

## Repos
- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

To add a new feature gate `NetworkAwareLiveMigration`, which enables two additional methods in live-migration procedure:
- If MAC address changed: detach / attach interface after live migration to have correct MAC address set

  The live-migration of virtual machines between the pods with different MAC address will invoke NIC reattaching procedure. This may affect some applications which are binding to the specific interface inside the VM.

- If MAC address is not changed: link down / link up interface to force VM request new IP address and routes from DHCP

  The live-migration of virtual machines between the pods with same MAC address will invoke the procedure to link down and up for the VM to renew DHCP lease with IP address and routes inside the VM. This is less destructive operation, but still may affect some workloads listening on particular IP addresses.

## API Examples
There are no API changes from the user side.

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
This change adds new logic also for multus connected networks.
When using multiple NICs, there are some CNIs that didn't work with live-migration. Now these CNIs will work.

## Functional Testing Approach
- Create two VMs: client and server in bridge mode
- Wait for launch
- Get IP address for server VM
- Run ping from client to server
- live-migrate client VM
- Run ping from client to server

# Implementation Phases
Implementation is already prepared as single pull request:
https://github.com/kubevirt/kubevirt/pull/7768
