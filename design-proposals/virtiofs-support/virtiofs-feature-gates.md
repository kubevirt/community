# Overview
Virtiofs allows sharing Secrets, ConfigMaps, ServiceAccounts and DownwardAPI (a.k.a., ConfigVolumes) in a dynamic way,
i.e., any change in those objects are reflected in the VirtualMachine (VM) without restarting the VM. Moreover, PVCs may
also be shared using virtiofs allowing multiple readers and writers in the shared PVCs.

## Motivation
Currently, the usage of feature gates are mixed with the feature control. The feature gate `ExperimentalVirtiofsSupport`
controls whether virtiofsd runs as rootful container inside the launcher pod or not. Moreover, there is no way to
disable it completely and this feature is not in GA.

In order to GA virtiofs, VMs should be able to live migrate while sharing data with virtiofs. The virtiofsd’s live
migration support will distinguish between different use cases:  sharing ConfigMaps, Secrets, DownwardAPI and
ServiceAccounts and sharing PVCs. 

This distinction will be useful because in order to fully support sharing ConfigVolumes and PVCs while allowing
live migration, deep changes are required in KubeVirt, QEMU, libvirt and virtiofsd projects. However, those
changes are complementary, i.e., the changes required to support sharing ConfigVolumes while allowing a VM to live
migrate are the needed and preliminary steps to support sharing PVCs. Therefore, this distinction will create a staggered
and granular support giving users time to test the ConfigVolumes sharing support and allowing the involved developers to
get feedback before next steps to support sharing PVCs.

## Goals
- Align virtiofs feature gates to just control the feature enablement.
- Protect the sharing functionalities: ConfigVolumes and PVCs, with feature gates.

## Non Goals
- Support live migration while sharing data with virtiofs.

## Definition of Users
A VM owner with permissions to read ConfigVolumes and/or read and/or write PVCs.

## User Stories
As a VM owner, I want to share Secrets, ConfigMaps, ServiceAccounts and/or DownwardAPI in a dynamic way. For this purpose,
I need to define a Secret, ConfigMap, ServiceAccount and/or DownwardAPI or use an existing one, enable the feature gate
`EnableVirtioFsConfigVolumes` define a volume in `spec.volumes` and define the sharing device
in `spec.domain.devices.filesystems` as `virtiofs: {}` in the VM manifest.

As a VM owner, I want to share PVCs with one or multiple VMs and be able to read and/or write them. For that purpose,
I need to define a PVC or use an existing one, enable the feature gate `EnableVirtioFsPVC`, define a volume in 
`spec.volumes`, and define the sharing device in `spec.domain.devices.fileystems` as `virtiofs: {}` in the VM manifest.

## Repos
Kubevirt/kubevirt

# Design
## KubeVirt Feature Gates
Two feature gates will be introduced as a means to granular support virtiofs functionalities:
- `EnableVirtioFsConfigVolumes` to allow sharing ConfigMaps, Secrets, DownwardAPI and ServiceAccount.
- `EnableVirtioFsPVC` to allow sharing PVCs.
The feature gates may start its path to GA once a VM will be able to live migrate while sharing data with virtiofs.
These feature gates will be removed once GA is reached. 

## VM API
From the VM API point of view, everything should keep as it is.

## Update/Rollback Compatibility
This is a breaking change, users using virtiofs in any way won’t be able to further use virtiofs without enabling the
proper feature gate/s.

## Functional Testing Approach
Current functional tests will be adjusted to those new feature gates. 

# Implementation Phases
- Drop usage of `ExperimentalVirtiofsSupport` 
- Add the feature gate `EnableVirtioFsConfigVolumes`.
- Add the feature gate and `EnableVirtioFsPVC`.

# Open Questions
- Currently, it is possible to run rootful virtiofs containers.
   - Should we support this use case at all?, since the possibility of running a rootful container will be dropped.
   - Should we drop this functionality entirely?
