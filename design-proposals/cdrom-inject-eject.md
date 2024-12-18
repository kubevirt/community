# Overview
Proposal for for adding inject/eject CD-ROM support to KubeVirt

## Motivation
In a world where all of our knowledge has been collected, indexed, and can be served up via a network connection, the humble CD-ROM is still a "thing." And to the dedicated hangers-on of this once ubiquitous technology, KubeVirt's CD-ROM support is lacking in a very fundamental way. Today, a user must reboot a VM in order to change the media in a CD-ROM. That is just. not. right. Fortunately, the virtualization technology that underpins KubeVirt has had this functionality for a long time and we can easily plug it into the KubeVirt API to unleash the full power of the CD-ROM

## Goals
- Extend existing APIs to support inject/eject CD-ROM operations
- Extend virtctl to have cdrom specific commands
- Support for PersistentVolumeClaim/DataVolume based disk images

## Non Goals
- Support for any volume type other than DavaVolume/PersistentVolumeClaim

## Definition of Users
KubeVirt end user, may not be a namespace/cluster admin

## User Stories
- As a VM owner, I want to be able to attach a CD-ROM for the purposes of installing an operating system, then be able eject the CD and perform a guest-side reboot to use the operating system.
- As a VM Owner I want to attach an ISO/CD-ROM from my desktop to the VM without the need to reboot the VM
- As a VM Owner I want to attach a PVC which contains an ISO to my VM without the need to reboot the VM
- As a VM Owner I want to be able to "eject" (remove) the ISO/CD from my VM without the need to reboot the VM

## Repos
kubevirt/kubevirt

# Design
Like most new features in KubeVirt, the building blocks already exist, just have to wire everything together

## VirtualMachineInstance API
Currently, every disk in a KubeVirt VM must have a corresponding volume. That restriction will be lifted for CD-ROM disks, i.e., it will be possible to define a VMI that has no corresponding volume for a CD-ROM. For Example:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
...
spec:
  domain:
    devices:
      disks:
      - cdrom:
        bus: sata
      name: cdrom
      - disk:
        bus: virtio
      name: root
...
  volumes:
  - dataVolume:
      name: fedora-root
    name: root
  networks:
...
```

## Volume Hotplug Subresource API
This feature leverages the Volume Hotplug API in order to add/remove disk images in running Pods. Hence, the restriction on PersistentVolumeClaim and DataVolume volume types.

Hotplug operations are wrapped in the `VirtualMachineVolumeRequest` struct which looks like this:

```golang
type VirtualMachineVolumeRequest struct {
  // AddVolumeOptions when set indicates a volume should be added. The details
  // within this field specify how to add the volume
  AddVolumeOptions *AddVolumeOptions `json:"addVolumeOptions,omitempty" optional:"true"`
  // RemoveVolumeOptions when set indicates a volume should be removed. The details
  // within this field specify how to add the volume
  RemoveVolumeOptions *RemoveVolumeOptions `json:"removeVolumeOptions,omitempty" optional:"true"`
}
```

One of `AddVolumeOptions` or `RemoveVolumeOptions` is expected to be set

### AddVolumeOptions
`AddVolumeOptions` as currently defined:

```golang
type AddVolumeOptions struct {
  // Name represents the name that will be used to map the
  // disk to the corresponding volume. This overrides any name
  // set inside the Disk struct itself.
  Name string `json:"name"`
  // Disk represents the hotplug disk that will be plugged into the running VMI
  Disk *Disk `json:"disk"`
  // VolumeSource represents the source of the volume to map to the disk.
  VolumeSource *HotplugVolumeSource `json:"volumeSource"`
  DryRun []string `json:"dryRun,omitempty"`
}
```

Currently the `disk` field is required. That restriction will be removed. However, when the `addVolume` request is received, we will validate that the corresponding disk is in fact a CD-ROM. The following request will be valid for the VirtualMachine defined above:

```yaml
addVolumeOptions:
  name: cdrom
  volumeSource:
    persistentVolumeClaim:
       claimName: windows11-iso
       hotpluggable: true
```

### RemoveVolumeOptions
`RemoveVolumeOptions` as currently defined:

```golang
type RemoveVolumeOptions struct {
  // Name represents the name that maps to both the disk and volume that
  // should be removed
  Name string `json:"name"`
  DryRun []string `json:"dryRun,omitempty"`
}
```

Since we don't want to remove the CD-ROM when ejecting a disk, a new field will be added to `RemoveVolumeOptions` to specify whether to keep/remove a disk.

```golang
// DiskRetentionPolicy defines how removevolume subresource API
// requests should handle the corresponding disk
type DiskRetentionPolicy string

// DiskRetentionDelete specifies that removevolume should
// delete the matching disk in the VirtualMachineInstance spec
const DiskRetentionDelete = "delete"

// DiskRetentionKeep specifies that removevolume should
// keep the matching disk in the VirtualMachineInstance spec
const DiskRetentionKeep = "keep"
```

A `DiskRetentionPolicy` field will be added to `RemoveVolumeOptions`. It will be optional and if excluded, the default is `DiskRetentionDelete` in order to be compatible with old clients

This would be the API for ejecting a CD-ROM in the VirtualMachine defined above

```yaml
removeVolumeOptions:
  name: cdrom
  diskRetentionPolicy: keep
```

## virtctl
A couple new convenience commands will be added to virtctl

### virtctl cdrom inject
Inject a cdrom

```bash
virtctl cdrom intect myvm --volume-name=cdrom --claim-name=windows-iso
```

### virtctl cdrom eject
Eject a cdrom

```bash
virtctl cdrom eject myvm --volume-name=cdrom --claim-name=windows-iso
```

## Scalability
Nothing new here that should effect system scalability

## Update/Rollback Compatibility
API changes are backwards compatible
- One field is no longer required
- One field is new

Rollback would be a problem given the `VirtualMachineInstance` API change. Do we currently support rollback?

## Functional Testing Approach
- Inject a CD-ROM to a running VM
- Eject a CD-ROM from a running VM
- Boot with hotplug CD-ROM, eject, and inject a different CD-ROM

# Implementation Phases
This is a small change and will be implemented in one phase
