# Overview
Virtiofs allows sharing Secrets, ConfigMaps, ServiceAccounts and DownwardAPI (a.k.a., ConfigVolumes) in a dynamic way,
i.e., any change in those objects are reflected in the VirtualMachine (VM) without restarting the VM. Moreover, PVCs may
also be shared using virtiofs allowing multiple readers and writers in the shared PVCs.

## Motivation
Currently, the usage of feature gates are mixed with the feature control. The feature gate `ExperimentalVirtiofsSupport`
controls whether virtiofsd runs as rootful container inside the launcher pod or not. The feature gate ExperimentalVirtiofsSupport
controls whether virtiofs can be used, and it automatically runs as rootful container inside the launcher pod when
needed. There is no way to only run virtiofs for use cases which doesn't require rootful virtiofsd or the other way
around.

In order to GA virtiofs, VMs should be able to live migrate while sharing data with virtiofs. The virtiofsd’s live
migration support will distinguish between different use cases:  sharing ConfigVolumes and PVCs. 

In order to fully support sharing ConfigVolumes and PVCs while allowing live migration, deep changes are required in
KubeVirt, QEMU, libvirt and virtiofsd projects. However, those changes are complementary, i.e., the changes required to
support sharing ConfigVolumes while allowing a VM to live migrate are the needed and preliminary steps to support
sharing PVCs. Therefore, this distinction will create a staggered and granular support giving users time to test the
ConfigVolumes sharing support and allowing the involved developers to get feedback before next steps to support sharing
PVCs.

## Goals
- Align virtiofs feature gates to just control the feature enablement.
- Deprecate current virtiofs feature gate.
- Protect the sharing functionalities: ConfigVolumes and PVCs, with feature gates.

## Non Goals
- Support live migration while sharing data with virtiofs.

## Definition of Users
A VM owner with permissions to read ConfigVolumes and/or read and/or write PVCs.

## User Stories
As a VM owner, I want to share a ConfigVolume in a dynamic way. For this purpose,
I need to define a ConfigVolume or use an existing one, enable the feature gate `EnableVirtioFsConfigVolumes` define a
volume in `spec.volumes` and define the sharing device in `spec.domain.devices.filesystems` as `virtiofs: {}` in the VM
manifest.

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
As stated before, `ExperimentalVirtiofsSupport` will be marked as deprecated. However, in order to do not disrupt current
systems relying on the `ExperimentalVirtiofsSupport` feature gate, it will allow users to share ConfigVolumes and PVCs
until KubeVirt version 1.6. In versions equal or greater than 1.6, the feature gate
`ExperimentalVirtiofsSupport` will have no effect. Moreover, the use of the deprecated `ExperimentalVirtiofsSupport`
will trigger a warning advising the user to move to the new feature gates.

## Functional Testing Approach
Current functional tests will be adjusted to those new feature gates. 

# Implementation Phases
- Drop usage of `ExperimentalVirtiofsSupport` 
- Add the feature gate `EnableVirtioFsConfigVolumes`.
- Add the feature gate and `EnableVirtioFsPVC`.