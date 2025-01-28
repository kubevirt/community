# Overview
Proposal for for adding inject/eject CD-ROM support to KubeVirt

## Motivation
In a world where all of our knowledge has been collected, indexed, and can be served up via a network connection, the humble CD-ROM is still a "thing." And to the dedicated hangers-on of this once ubiquitous technology, KubeVirt's CD-ROM support is lacking in a very fundamental way. Today, a user must reboot a VM in order to change the media in a CD-ROM. That is just. not. right. Fortunately, the virtualization technology that underpins KubeVirt has had this functionality for a long time and we can easily plug it into the KubeVirt API to unleash the full power of the CD-ROM

## Goals
- Extend existing APIs to support inject/eject CD-ROM operations
- Extend virtctl to have cdrom specific commands
- Support for PersistentVolumeClaim/DataVolume based disk images

## Non Goals
- Support for any [volume type](https://kubevirt.io/user-guide/storage/disks_and_volumes/#volumes) other than DavaVolume/PersistentVolumeClaim

## Definition of Users
KubeVirt end user, may not be a namespace/cluster admin

## User Stories
- As a VM owner, I want to be able to attach a CD-ROM for the purposes of installing an operating system, then be able eject the CD and perform a guest-side reboot to use the operating system
- As a VM Owner, I want to attach an ISO/CD-ROM from my desktop to the VM without the need to reboot the VM
- As a VM Owner, I want to attach a PVC which contains an ISO to my VM without the need to reboot the VM
- As a VM Owner, I want to be able to "eject" (remove) the ISO/CD from my VM without the need to reboot the VM

## Repos
kubevirt/kubevirt

# Design
Like most new features in KubeVirt, the building blocks already exist, just have to tweak a couple things and wire it all together

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

At present, the `disk` field is required. That restriction will be removed. However, when the `addVolume` request is received, we will validate that the corresponding disk is in fact a CD-ROM. The following request will be valid for the VirtualMachine defined above:

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
When KubeVirt hotplugs a disk, a Pod is created to mount the volume to the Node. It is possible that adding this feature will lead to more Pods in the cluster, which could cause Quota violations and increased resource consumption

## Update/Rollback Compatibility
API changes are backwards compatible
- One field is made optional
- One field is new

Rollback would be a problem given the `VirtualMachineInstance` API change. Do we currently support rollback?

## Functional Testing Approach
- Inject a CD-ROM to a running VM
- Eject a CD-ROM from a running VM
- Boot with hotplug CD-ROM, eject, and inject a different CD-ROM

# Implementation Phases
This is a small change and will be implemented in one phase

# Drawbacks

## Subresource API
This design leverages the existing Subresource Volume Hotplug API for VM/VMIs, which could be considered a relic of the past, a time when KubeVirt purposely did not support any "live" updates to a VM spec. Times have changed. And they have changed mostly because of the following limitations:

### Not GitOps compatible
Subresource APIs are not helpful to users that want to communicate all VM state changes in a declarative/GitOps compatible way. In fact, GitOps users have to be extra careful when using this feature to avoid conflicts with volume names/definitions. Unfortunately, KubeVirt does not support declarative volume updates for any disk type at this time.

### Complicated for users to invoke directly
kubectl does not give users an easy way to invoke a subresource API. Which means, as a practical matter, virtctl support is necessary and not just "nice to have."

# Alternatives

## Declatative API
Given the drawbacks of the subresource API, why not propose a declarative way to inject/eject CD-ROMs in this document? The advantages of a declarative API are obvious. It is "the Kubernetes way." Here's why:

### It's bigger than just CD-ROM
Declarative inject/eject CD-ROM falls under the umbrella of Declarative Volume Hotplug. CD-ROMs are just one of the four [disk types](https://kubevirt.io/user-guide/storage/disks_and_volumes/#disks) that KubeVirt supports. The work required for supporting one disk type is roughly equal to supporting all four, requires no special considerations for CD-ROMs, and is worthy it's own design proposal. Ultimately we should support both declarative and subresource. Supporting just subresource initially would allow us to release this feature sooner. But that is not the only reason to support subresource.

### VirtualMachineInstances
If declarative is the only way to go for inject/eject, `VirtualMachineInstances` will be neglected. They can stand on their own without a `VirtualMachine` owner and currently support Subresource Volume Hotplug. VMIs are also immutable to regular users and we have no intention of changing that.

### RBAC Limitations
Say you want to give a user permission to inject/eject CD-ROMs but not allow them to hotplug memory/cpu. You can do that with a subresource API. There is no straightforward way to do this with a declarative API. As more users come to KubeVirt from traditional virtualization environments with complex authorization requirements and no desire for GitOps, we may start looking at subresources a little more fondly.
