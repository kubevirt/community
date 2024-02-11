# Overview

This KubeVirt design proposal discusses how KubeVirt can be used to create `libvirt` virtual machines that are backed by diverse hypervisor drivers, such as QEMU/KVM, Xen, VirtualBox, etc. The aim of this proposal is to enumerate the design and implementation choices for enabling this multi-hypervisor support in KubeVirt. 

## Motivation

Although KubeVirt currently relies on libvirt to create and manage virtual machine instances (VMIs), it relies specifically on the QEMU/KVM virtualization stack (VMM and hypervisor) to host the VMI. This limits KubeVirt from being used in settings where a different VMM or hypervisor is used. 

In fact, libvirt itself is flexible enough to support a diverse set of VMMs and hypervisors. The libvirt API delegates its implementation to one or more internal drivers, dependending on the [connection URI](https://libvirt.org/uri.html) passed when initializing the library. The list of currently supported hypervisor drivers in Libvirt are:
- [LXC - Linux Containers](https://libvirt.org/drvlxc.html)
- [OpenVZ](https://libvirt.org/drvopenvz.html)
- [QEMU/KVM/HVF](https://libvirt.org/drvqemu.html)
- [VirtualBox](https://libvirt.org/drvvbox.html)
- [VMware ESX](https://libvirt.org/drvesx.html)
- [VMware Workstation/Player](https://libvirt.org/drvvmware.html)
- [Xen](https://libvirt.org/drvxen.html)
- [Microsoft Hyper-V](https://libvirt.org/drvhyperv.html)
- [Virtuozzo](https://libvirt.org/drvvirtuozzo.html)
- [Bhyve - The BSD Hypervisor](https://libvirt.org/drvbhyve.html)
- [Cloud Hypervisor](https://libvirt.org/drvch.html)

There are several parts in the KubeVirt source-code that hard-code the use of the QEMU/KVM hypervisor driver, which prevents the creation of VMIs using another hypervisor driver. Therefore, KubeVirt needs to be updated to introduce the choice of the backend libvirt hypervisor driver to use for creating a given VMI. This would expand the set of scenarios in which KubeVirt can be used.

## Goals

KubeVirt should be able to offer a choice to its users over which libvirt hypervisor-driver they want to use to create their VMI.

## Non Goals

Support all features available in KubeVirt in all libvirt hypervisor drivers. As libvirt makes progress in bringing feature parity among its hypervisor drivers, KubeVirt will also enable more features in the evolving hypervisor drivers.

## Definition of Users

## User Stories

- A user trying to use KubeVirt on a cluster of machines that have a hypervisor-VMM pair that is not necessarily QEMU/KVM.
- A regular user of libvirt with any of its supported hypervisor drivers who now wants to leverage the orchestration capability provided by KubeVirt. But the user does not want to abandon their hypervisor driver of choice.
- A regular user of a hypervisor-VMM-specific orchestration capability to expand their use to a cluster with a diverse set of hypervisor-VMM pairs.

## Repos

- KubeVirt
- Libvirt

# Design

## Notes [remove this section]

- Nodes should advertise which hypervisor-driver they can support
- KubeVirt API should be extended to say which hypervisor-driver to use.
- Different virt-launcher image for each supported hypervisor-driver.
- Maintain a list of which VM/VMI features are supported by different hypervisor-drivers. If the KubeVirt API user requests a feature for a VMI that is not supported by the requested hypervisor-driver, then the request for creation of VMI should be rejected.
  - If the cluster of machines has diff machines running diff hypervisor-drivers, then the virt-controller should select a machine to host the VMI so that the VMI features are supported on that machine's hypervisor driver. 

## KubeVirt

## Libvirt

## API Examples

## Functional Testing Approach

# Implementation Phases


# References

1. [Cloud Hypervisor integration - Google Groups](https://groups.google.com/g/kubevirt-dev/c/Pt9CDYJOR2A)
2. [[RFC] Cloud Hypervisor integration POC](https://github.com/kubevirt/kubevirt/pull/8056)
3. [design-proposals: Cloud Hypervisor integration](https://github.com/kubevirt/community/pull/184)