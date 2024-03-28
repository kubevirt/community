# SIG Storage

## Scope

SIG Storage mainly responsible for the code that bridges the gap between Kubernetes storage primitives (PersistentVolumeClaims, VolumeSnapshots, CSI drivers, etc) and Virtual Machine storage requirements (bootable/mountable VM disk images, disk expansion, hotplug support, etc).

### In Scope

The main areas:
* APIs
  * VirtualMachineSnapshot
  * VirtualMachineRestore
  * VirtualMachineExport
* KubeVirt volumes
  * ContainerDisks
  * PersistentVolumeClaims
  * DataVolumes
  * EphemralDisks
  * EmptyDisks
  * HostDisks
  * ConfigMaps
  * Secrets
  * MemoryDumps
* Virtualization tools compatibility
  * Hotplug Volumes
  * virtio-fs
  * SCSI persistent reservation
  * libguestfs tools
* KubeVirt client commands
  * virtctl image-upload
  * virtctl guestfs
  * virtctl vmexport
  * virtctl memory-dump

Additional **secondary** responsibilities are:
* Fixing storage related functional tests and flakiness

### Out of scope
The SIG is NOT responsible for:
* Testing and fixing every CSI storage provider integrating with KubeVirt.

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs

* Welcome new contributors
* Resolve conflicts

## Meeting Mechanics

* Where: Zoom meeting ID: 97050528531
  * Link to join is [here](https://zoom.us/j/97050528531)

* When: Starting Monday, 5 Dec 2022 @ 14:00 CET/CEST (08:00 EST/EDT) and repeating every two weeks.  See the [KubeVirt calendar here](https://calendar.google.com/calendar/embed?src=kubevirt@cncf.io)

* [Meeting minutes](https://docs.google.com/document/d/1mqJMjzT1biCpImEvi76DCMZxv-DwxGYLiPRLcR6CWpE/edit) are sent to the kubevirt-dev mailing list/group:
  * mail-to: kubevirt-dev@googlegroups.com
  * Subject: [SIG-storage] Meeting Notes <DATE>

## Agenda
* Discuss storage topics important to community members.
* Raise awareness of issues and pull requests that affect storage.
