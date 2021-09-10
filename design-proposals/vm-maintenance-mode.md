# Overview
Add a new "Maintenance" state for the VMI/VM that represents a special state where the VM doesn't run and we can safely access the VM disks.

The "maintenance" pods are pods created to perform disk operations (for example to modify the disk with libguestfs-tools) by KubeVirt. The maintenance pods must have the same scheduling constraints as the VM in order to not alter the scheduling of the VM.

We can place the VM in this state by setting the `RunStrategy = Halted` and a VMI by setting `StartStrategy = Maintenance`. If we set the `StartStrategy` to `Maintenance` while creating the VMI then KubeVirt will simply create the VMI object without starting the virt-launcher pod. If the annotation `kubevirt.io/leave-launcher-pod-after-qemu-exit: true` introduced by https://github.com/kubevirt/kubevirt/pull/6040 is set then we cannot set the `StartStrategy = Maintenance`. If the `StartStrategy` to `Maintenance` while the VMI is running then the virt-lancher pod will be terminated.

Additionally, we also need to guarantee that there are not multiple maintance pods that can access the VM disks at the same time. In order to guarantee this, KubeVirt needs to monitor those pods. One concrete examples, are the pods created by `virtctl guestfs` command. The virt-controller can put the label `kubevirt.io/maintenance: <pod-name>`with the name of the maitenance pod that is accessing the VM disks on the VM/VMI. Once this pod is in Terminating, Failed, Completed or doesn't exit anymore, we can remove the label. If a second maintenance pods wants to access the VM disks virt-controller needs to be able to set the label on the VMI/VM with the name of the second pod. If the label is already set, this will fail and we know that there is another pod that is accessing the VM disks. A label cannot be changed if it isn’t explicitly overwritten, and the set of the label fails if it is already present. This guarantees that the VM/VMI can be locked for a single maintenance pod at the same time.


A VM is in Maintenance state if `RunStrategy = Halted` and it has the label `kubevirt.io/maintenance: <pod-name>`. A VMI is in Maintenance phase if `StartStrategy = Maintenance` and it has the label `kubevirt.io/maintenance: <pod-name>`.

This approch works based on VM/VMI definition. This implies that we cannot access disks separtely as we are basically locking the VM/VMI object. An alternative approach based on single PVC has been proposed [here](https://docs.google.com/document/d/1XIEVdszVBBihfkuQ7I6YPCtmKMHctdlDOE8Dup-G06g/edit?usp=sharing).

## What is missing
How do we request the maintenance pods? Should they be create on the server (KubeVirt) side or client?

## Motivation
The main motivation for this proposal is to put the VM in this state in which we can safely perform disk operations using the maintenance pod. For example, for interactive operations (e.g virtctl guestfs) or batch processing jobs (e.g tekton task). We can access the volumes used by the VM from its specification and attach these volumes to the maintenance pod for disk inspection and manipulation. 

## Goals
The user can put the VM in this mode and safely access disks without the danger of corrupting the data.

## Non Goals
  + This proposal does not include any shared storage locking (such as Sanlock).  It is opt-in locking of the VM workload.
  + This proposal does not protect from pods that directly use the PVC and are not controlled by KubeVirt


## User Stories
  + As a user creating VMs, I want to be able to safely access the disks attached to the VM for disk inspection and manipulation with `virtctl guestfs` command
  + As storage administrator, I want to be able to create periodic or automatic batch processing job on disks of a VM and safely access the disks
    + Sparsify disks
  + As VM administrator, I want to be able to create automatic batch processing job on disks of a VM and safely access the disks
    + Restoring passwords
    + Creating new users
    + Installing missing packages

## Repos
kubevirt/kubevirt

# Design
Add new phase for VM/VMI that represents the state when the disks owned by the VM are access for maintenance and the VM cannot be started if it is in Maintenance mode without explicitly remove it from this state.

## API Examples
Example of VMI with StartStrategy = Maintenance:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  labels:
    special: vmi
  name: vmi
spec:
  startStrategy: Maintenance
````

Example with virtctl:
```bash
$ kubectl apply -f examples/vm-cirros.yaml
$ kubectl get vm
NAME        AGE   STATUS    READY
vm-cirros   3s    Stopped   False
$ virtctl guestfs --vm vm-cirros
[...]

$ kubectl get vm
NAME        AGE   STATUS        READY
vm-cirros   3s    Maintenance   False
$ virtctl start vm-cirros
Failed to start vm-cirros as it is in maintenance mode
# Exit from guestfs pod remove vm-cirros from the Maintenance status
$ kubectl get vm
NAME        AGE   STATUS    READY
vm-cirros   3s    Stopped   False
$ virtctl start vm-cirros
VM vm-cirros was scheduled to start
```

## Functional Testing Approach
Verify that the maintenance pod could be created only during the Maintenance state
 
