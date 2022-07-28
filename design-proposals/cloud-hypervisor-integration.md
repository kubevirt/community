# Overview

[Cloud Hypervisor](https://github.com/cloud-hypervisor/cloud-hypervisor) is a
Virtual Machine Monitor meant for running modern Cloud workloads. It is written
in Rust and free of legacy devices to provide a smaller attack surface to the
guest, which makes it a more secure alternative when it comes to running virtual
machines.

This document describes a design proposal for integrating Cloud Hypervisor with
KubeVirt, providing KubeVirt's users the possibility to rely on Cloud Hypervisor
to create virtual machines as an alternative to the default libvirt/QEMU.

## Motivation

Since Cloud Hypervisor aims at running virtual machines more securely, it is
important to offer KubeVirt's user this choice.

Cloud Hypervisor has been designed for Cloud workloads, which makes it perfectly
fit for the Cloud Native ecosystem, and that is the reason why it is already
integrated as part of the Kata Containers project.

To extend its overall Cloud Native support, it seems logical to integrate it
with KubeVirt.

One other reason for going through this effort is to identify if the abstraction
layers are correctly defined to support another VMM. This will help improve the
existing code by defining cleaner interfaces if needed.

## Goals

Provide users a way to choose Cloud Hypervisor over libvirt/QEMU to run their
virtual machines.

## Non Goals

Support all features available through KubeVirt.

Since Cloud Hypervisor has a much narrower scope than libvirt/QEMU, it doesn't
support as many features. Therefore, we can only expect a subset of KubeVirt's
features to be supported by Cloud Hypervisor.

## Definition of Users

This feature is directed at KubeVirt's users who want to run virtual machine
more securely by choosing Cloud Hypervisor over libvirt/QEMU.

## User Stories

A user recently tried Cloud Hypervisor and wants to use it for running virtual
machines on his Kubernetes/KubeVirt cluster.

## Repos

- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

Looking at KubeVirt's architecture, each `virt-launcher` instance manages a
single pod. This is the abstraction layer we need to replace so that instead of
using `libvirt` to spawn QEMU's virtual machines, it will run and communicate
with Cloud Hypervisor directly.

A new launcher `ch-launcher` will be created so that it fully replaces the
existing `virt-launcher` component when needed.

## API Examples

Example of how a user could request Cloud Hypervisor as the underlying
hypervisor through the VMI spec:

```yaml
spec:
  hypervisor: cloud-hypervisor
```

- Introduction of a new field `Hypervisor` in `VirtualMachineInstanceSpec`
- By default if no `Hypervisor` is provided, it would default to `libvirt`.
- The two acceptable entries would be either `cloud-hypervisor` or `libvirt`.

The `virt-operator` can inform all other components about the hypervisor type
based on the information from the VMI spec. A different `virt-launcher` image
would be picked instead of the default one, so that it contains `ch-launcher`.

## Features

### Supported features

Here is a list of features expected to be available with Cloud Hypervisor.

#### Lifecycle

- Create a VM
- Start a VM
- Pause/Resume a VM
- Snapshot/Restore a VM
- Stop a VM

#### Virtual Hardware

- OVMF support (EFI)
- CPU topology
- CPU model is exclusively the equivalent of host for QEMU (no emulation of specific CPU model)
- RNG using virtio-rng
- Only headless VMs as we don't have graphics or video device emulation
- CPU constraints + hotplug
- Memory constraints + hotplug
- Hugepages

#### NUMA

- Host NUMA to select specific host CPUs and make sure memory is allocated on expected NUMA node
- Guest NUMA to expose any NUMA configuration to the guest

#### Disks and Volumes

- Disk support with virtio-block
- Filesystem support with virtio-fs
- Volume hotplug
- Offline snapshot
- Online snapshot (will come later with the guest agent support for freezing the filesystem)

#### Network

- Support based on virtio-net or vhost-user-net
- Support for tap and macvtap

#### Host Device Assignment

- VFIO supported for passing through PCI devices

#### Accessing Virtual Machines

- Serial port (0x3f8) and virtio-console are supported
- Create a PTY so that external process can later connect to it
- Support for SSH as it directly depends on virtio-net support

#### Confidential Computing

- Support for SGX
- Experimental support for TDX

#### Architectures

- x86_64
- Aarch64

#### Migration

This should be supported eventually but it still requires some assessment of
how to achieve it. Therefore we might not see this feature being supported for
some time.

### Unsupported features

Here is the list of what will be missing compared to what libvirt/QEMU supports:

#### Virtual Hardware

- No CPU model emulation
- No way to pick a type of clock
- No way to pick a type of timer
- No support for emulated video and graphics devices
- No way to pick between different features like `acpi`, `apic`. We can select
  `hyperv` though, which enables KVM Hyper-V enlightments
- No support for emulated input device

#### Disks and Volumes

- No support for cdrom, floppy disk or luns
- No support for resizable disk

#### Network

- No support for emulated NICs such as e1000, e1000e, ... (which means no SLIRP)

#### Accessing Virtual Machines

- No support for VNC

### Guest Agent

Features related to the ability of running a dedicated agent in the guest have
not been tested yet. The existing QEMU agent must be evaluated to see if it
could work and be reused directly with Cloud Hypervisor. If that's not the case,
an agent program would have to be developed for operations like `GuestPing`,
`ListInterfaces`, ...

## Update/Rollback Compatibility

This new feature should not impact updates moving forward since it doesn't
remove anything.

## Functional Testing Approach

Create an additional CI entry to run Cloud Hypervisor dedicated testing. And of
course the set of tests that will be run would be a subset of what is already
available.

## Proof of Concept

As a reference, a PoC can be found through the following
[pull request](https://github.com/kubevirt/kubevirt/pull/8056).

It modifies the existing `virt-launcher` component so that it manages Cloud
Hypervisor VMs instead of libvirt ones.
    
It adds support for the following features:
- containerDisk
  Since there's no support for compressed QCOW2 in Cloud Hypervisor,
  I've simply converted the image to a RAW version. That means we
  don't get the COW benefit but it works fine.
- emptyDisk
  I've added a way to create a RAW image instead of QCOW2 since Cloud
  Hypervisor doesn't support compressed QCOW2 images.
- cloudInitNoCloud
  Pretty straightforward, I reused most of the code provided by the
  repository
- Console
  I had to run two extra go routines to redirect input/output between
  the PTY device that is created by Cloud Hypervisor and the socket
  located at /var/run/kubevirt-private/<pod-UID>/virt-serial0 that is
  expected by virt-handler
- Network
  Added support for both bridge and masquerade modes. This is done
  through the existing code, with minimal changes as I used the
  api.Domain reference that is being modified to retrieve both TAP
  interface name and expected MAC address
- Kernel boot + initramfs
  This is "supposedly" working but when I used vmi-kernel-boot example
  I ended up running into some issues because the kernel binary
  vmlinuz is not a PVH ELF header. I didnt' spend some time creating a
  dedicated docker image containing the right type of kernel binary
  but I expect this to work as long as the user provides a proper
  image
- VM lifecycle
  - Sync VMI creates and boots the VM based on the configuration that
    has been generated from the VirtualMachineInstanceSpec. The support
    for updating the VM and especially hotplugging devices hasn't been
    implemented through this POC.
  - Pausing and resuming the VM is supported through virtctl
  - Stopping and deleting the VM is also supported through kubectl
    delete.
- Lifecycle events
  Listen to the events reported by Cloud Hypervisor through the
  event-monitor socket, and transform them into domain events, setting
  the appropriate status and reason for a state change

It has been tested with the following VMI examples:
- examples/vmi-fedora
- examples/vmi-masquerade

Note the Bazel workspace had to be updated so that the virt-launcher
container image would be generated with both `CLOUDHV.fd` firmware and
the Cloud Hypervisor binary.

# Implementation Phases

## Create `ch-launcher` binary

Create a minimal `ch-launcher` binary based off the `virt-launcher` one, just
enough to launch Cloud Hypervisor and connect to it, but with the domain manager
implementation providing empty shells.

## Create a new image

The first thing to do is to update the Bazel workspace to be able to generate a
new `ch-launcher` image dedicated for Cloud Hypervisor. This image should
contain what is needed to start a Cloud Hypervisor virtual machine, that is
the `cloud-hypervisor` binary pulled from the Cloud Hypervisor release, and the
associated OVMF firmware called `CLOUDHV.fd`. It must also contain the
`ch-launcher` binary instead of the `virt-launcher` one.

## Update VMI specification

Add a new field `Hypervisor` to the `VirtualMachineInstanceSpec` structure to
carry information about which hypervisor should be used.

Update all the components where it's assumed to always rely on `virt-launcher`
image so that it is dynamically chosen based on the `Hypervisor` value.

## Implement basic features

At this point we must extend the minimal `ch-launcher` implementation to end up
with a functional implementation so that some testing can be performed.

## Add a new CI worker

Define a new entry in the CI to perform the testing of KubeVirt with Cloud
Hypervisor. The amount of tests that can be run will be directly dependent on
the amount of features supported by this first version of `ch-launcher`.

## Enable new features one by one

At this point, it makes sense to submit one pull request per new feature that
we want to support as part of the Cloud Hypervisor integration effort.