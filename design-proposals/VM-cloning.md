# Overview
VM Cloning is the action of creating a new VM from an existing one that has the exact same settings. This includes
VM disks, network definition, CPU topology and so on.

## Motivation
The ability to clone a VM is a generic virtualization operation which may be useful for several use-cases.

For example, a system admin / user might want to base a new VM on an existing one. In such a case we can clone a VM and
then change it according to our needs. In addition, a cluster admin might want to try different configurations
(for example, to check whether other configuration provide better performance for his workload), in this case he can
clone the existing VM, test it and decide whether to move the new settings.

Naively copy/pasting a VM spec would not work in most cases therefore this serves as an automation for cloning safely.

## Goals
Have a mechanism to clone VMs. Original and cloned VM would have the same:
* Disks
* Network configuration
* Volumes
* CPU topology

Also, provide an ability to fine-tune what's being cloned. For example, it is possible to clone or skip
annotations / labels.  

## Non Goals
In first implementation iteration the following limitations exist:
* Only offline cloning will be supported, meaning cloning a VM when it's stopped.
* Cloned VM would be in the same namespace as the original one.
* Host disks would not be supported.

## Definition of Users
Everyone who creates a VM might create a VM clone.

## User Stories
* As a user / admin, I want to be able to create a new VM that is based on an existing VM.
* As a user / admin, I want to create a clone to a VM using virtctl.

## Repos
Kubevirt/kubevirt

# Design
Design can be inspired by the way VMIs are being migrated. In order to migrate a VMI, a [VirtualMachineInstanceMigration](https://kubevirt.io/user-guide/operations/live_migration/#initiate-live-migration)
CRD is posted into the cluster. This CRD is very thin and basically triggers a migration with basic parameters. Another
example is VM snapshots / restores.

For VM cloning a similar approach can be taken: a new CRD would be introduced that would provide some basic parameters like:
* The VM that needs to be cloned
* A new name for the VM (could be optional - a new name will be generated)
* New MAC address (optional: if not provided MAC would be erased)
* Which labels / annotations etc to clone or skip.

More parameters could be added in the future, like the new namespace for the VM, and more.

Also, in the future it's possible to expand cloning beyond VMs. For example, we can clone to / from
VMIs, snapshots, VMPools, etc.

## API Examples
As said above, API can be similar to [VirtualMachineInstanceMigration](https://kubevirt.io/user-guide/operations/live_migration/#initiate-live-migration)
and [VirtualMachineSnapshot](https://kubevirt.io/api-reference/master/definitions.html#_v1alpha1_virtualmachinesnapshot).

Here's a sketch for the new CRD:
```yaml
kind: VirtualMachineClone
  spec:
    source:
      # More sources may be supported in the future, e.g. VirtualMachineInstance or VirtualMachineSnapshot
      kind: VirtualMachine
      name: my-other-cool-vm
    
    target:
      # Again, more types will maybe be supported in the future
      kind: VirtualMachine
      name: my-target-vm  # Optional - a new name can be generated, e.g. my-cool-vm-jfh54b
    
    # All of the fields below are optional. Default is to clone everything.
    # In the future we can expend this to clone more aspects
    Annotations:
      cloneAllByDefault: true
      exclude:
      - "key-to-exclude"
      - "another-key"
      include:  # This is valuable when "cloneAllByDefault" is set to false
      - "key-to-include"
    Labels:
      # This is an example for how the default setting looks like
      cloneAllByDefault: true
    Disks:
      # Same as above
    Network:
      # Same as above

    # This is the place to specify certain fine-tunings. In the future this can allow choosing
    # under-the-hood optimizations / algorithms etc.
    newMacAddress: my-new-mac-address  # Optional - new mac address can be generated automatically
```

In the future, we may want to support configuring new namespace for the cloned VM. However, this requires
copying config-maps / secrets to another namespace which could be a potential risk.

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
New API so should not affect updates / rollbacks.

## Implementation details / challenges (in short)
Implementation utilizes VM Snapshots / VM Restores. As a start, a cloning operation would simply wrap snapshotting
and restoring a VM. In other words, if a VirtualMachineClone object that asks to clone VM1 into VM2, the following 
will happen under the hood:

* A VirtualMachineSnapshot will be created for VM1
* VM2 would be created (VM2 doesn't need to start/boot)
* A VirtualMachineRestore will be created, asking to restore VM1's snapshot into VM2
* The snapshot object would be deleted

Many optimizations and tweaks can and should pop up in the future to both allow fine-tuning and to introduce
runtime / storage optimizations.

Another note: we should bear in mind that a VM that boots for the first time usually does some special operations
like defining ssh hostname, MAC address, etc. Since the GUI performs cloning already we can look on their implementation
and learn from it.

## Functional Testing Approach
Functional tests can:
* Clone a VM and see it's successful
* Ensure that an illegal clone is not possible
  * Maybe check that MAC address is not the same as before? or does not exist in the cluster?
* Ensure that cloned VM has same spec as original VM in terms of number of disks, volumes, etc.

# Implementation Phases
As laid out in [this comment](https://github.com/kubevirt/community/pull/159#pullrequestreview-880329021),
the implementation phases are:

1) **Extend snapshot functionality to restore to new VM**
    * Focus entirely on making the SnapshotRestore object capable of restoring a snapshot to a new VM. This would only
    involve using the existing snapshot apis and extending that functionality.
    * The end result here is someone could technically clone a VM by creating a snapshot and restoring to a new VM.

2) **Introduce Clone API to coordinate the workflow of a snapshot/restore to new VM**
    * The new Clone API would be a wrapper around the existing snapshot/restore logic and coordinate this workflow
    in a declarative way. This would allow someone to declare they want to Clone a VM using a single object
    (VirtualMachineClone) and the clone controller would coordinate creating a snapshot and restoring the snapshot
    to the target VM name.
    * The end result here is someone could post a VirtualMachineClone object targeting one source VM and get a new target VM.

3) **Enhancements**
    * Today, snapshot/restore requires storage provisioners that support volume snapshots. We could
      technically support offline snapshots using PVC cloning in the event that a storage provider doesn't support
      snapshot. By enhancing the snapshot/restore logic, the VirtualMachineClone logic would naturally get enhanced
      as well since the clone logic is built on snapshot/restore.
    * supporting Online Clones. Since we support online snapshots in some scenarios, we could extend Clone logic to
      be able to create clones of VMs which are online.
    * Introduce a `virtctl` cloning command