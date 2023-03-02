# Live resize vCPU and Memory

Author: Andrei Kvapil \<kvapss@gmail.com\>

## Overview

Increase the amount of vCPU and RAM available for the guest OS on the fly.

## Motivation

This is a standard task for any virtualization system.  
Therefore, virtual machines should support hotplugging the CPU and RAM and be scalable vertically without rebooting.

## Goals

- Ability to resize the virtual machine without rebooting on the fly.

## Non Goals

- This feature must be supported by the guest OS.
- Live hotplugging works only when the maximum limit for CPU and RAM is specified. Thus a user should know about this opportunity to specify the correct values beforehand.

## Definition of Users

- Users who are running virtual machines with sensitive to downtime applications (almost all legacy, databases, etc.)

## User Stories

* As a user / admin, I want to have an opportunity for updating CPU/RAM resources on the fly without restarting the virtual machine.

## Repos

- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

**API changes**

Since the [`InPlacePodVerticalScaling`](https://github.com/kubernetes/kubernetes/pull/102884) feature has been merged into Kubernetes now we can rely on it to perform live VM resize.

Two new fields should be added to the domain spec. They must be set during the VMI creation and must not be changed for VMI that have already been created.

- `cpu.maxSockets` - Maximum number of VCPUs which can be hotplugged.

  The example below generates the following domain configuration:
  ```xml
  <vcpu placement='static' current='4'>16</vcpu>
  ```

  - The `current` value is equal to the requested cpu resources from domain spec.
  - The number of sockets is equal to `cpu.maxSockets` or the `current` value if `cpu.maxSockets` is not specified.

- `memory.max` - The maximum amount of memory that can be allocated for the VM.

  The example below generates the following domain configuration:
  ```xml
  <currentMemory unit='GiB'>2</currentMemory>
  <memory unit='GiB'>50</memory>
  ```

  - The `currentMemory` is equal to the requested memory resources or to the `memory.guest` if specified.
  - The `memory` is equal to `memory.max` if specified or to the `currentMemory` if `memory.max` is not specified.

The VMI spec should become mutable for the following fields:

- `domain.cpu.sockets` (only if `maxSockets` specified, the value must not be greater than `maxSockets`)
- `domain.memory.guest` (can be scaled up and scaled down with no troubles)
- `resources.requests.cpu`
- `resources.requests.memory`
- `resources.limits.cpu`
- `resources.limits.memory`


The memory ballooning device should be enabled for this feature to work.
*(it was disabled in v0.59 due to uselessness, the related PR: https://github.com/kubevirt/kubevirt/pull/1367)*

**Performing live-resize**

The virt-launcher's `SyncVMI` procedure should include the following steps:

- ```shell
  virsh setvcpus 1 <num>
  ```
  
  to change the number of vCPUs
  
- ```shell
  virsh setmem 1 --size <size>
  ```
  
  to change the size of the current memory

## API Examples

```diff
 apiVersion: kubevirt.io/v1
 kind: VirtualMachineInstance
 metadata:
   name: vm100
   namespace: default
 spec:
   domain:
     cpu:
       sockets: 4
+      maxSockets: 16
     memory:
       guest: 5Gi
+      max: 50Gi
     devices:
       disks:
       - disk:
           bus: virtio
         name: containerdisk
     resources:
       requests:
         cpu: "4"
         memory: 2Gi
   volumes:
   - name: containerdisk
     containerDisk:
       image: kubevirt/fedora-cloud-container-disk-demo:latest
```

## Scalability

I don't see any scalability issues.

## Update/Rollback Compatibility

**Update:**

- This configuration will work only for newly created VMIs.

**Rollback:**

- The rollback should not affect existing VMs, because KubeVirt has no opportunity to reconfigure resources on-the-fly and appropriate procedure for this.
  The already created VMs will simply continue to exist with the specified parameters.

## Functional Testing Approach

- Create a VM with `cpu.maxSockets` and `memory.max`.
- Wait until it starts.
- Check number of cores and the amount of memory inside the VMs.
- Resize the VMI.
- Check number of cores and the amount of memory inside the VMs.
- Destroy the VMI.

# Implementation Phases

**Phase 1:**
- Introduce the `LiveResize` feature gate
- Enable memory ballooning device back
- Update the VMI validation webhook for the specified fields
- Implement live resizing procedure for the virt-launcher

**Phase 2:**
- Update the `VirtualMachine` controller to synchronize requested resources from the VMI template into the existing `VirtualMachineInstance` resource.

**Phase 3:**
- `VirtualMachineInstance` controller should take into account amount of resources on node and live-migrate the virtual machine before the expansion in case of lack of resources on the current node.
