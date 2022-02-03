# Overview
VM Cloning is the action of creating a new VM from an existing one that has the exact same setting. This includes
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
CRD is posted into the cluster. This CRD is very thin and basically triggers a migration with basic parameters.

For VM cloning a similar approach can be taken: a new CRD would be introduced that would provide some basic parameters like:
* The VM that needs to be cloned
* A new name for the VM (could be optional - a new name will be generated)
* New MAC address (optional: if not provided MAC would be erased)
* Whether to clone or skip labels / annotations etc.

More parameters could be added in the future, like the new namespace for the VM, and more.

## API Examples
As said above, API can be similar to [VirtualMachineInstanceMigration](https://kubevirt.io/user-guide/operations/live_migration/#initiate-live-migration).

Here's a sketch for the new CRD:
```yaml
kind: VirtualMachineClone
  spec:
    source:
      # More sources may be supported in the future, e.g. VirtualMachinInstance or VirtualMachineSnapshot
      kind: VirtualMachine
      name: my-other-cool-vm
    
    targetVirtualMachineName: my-cool-clone   # Optional - a new name can be generated, e.g. my-cool-vm-jfh54b 
    newMacAddress: my-new-mac-address         # Optional - new mac address can be generated automatically
    
    # All of the fields below are optional and defaults to true
    cloneAnnotations: true
    cloneLabels: true
    cloneDisks: true
    cloneNetwork: true
```

In the future, we may want to support configuring new namespace for the cloned VM. However, this requires
copying config-maps / secrets to another namespace which could be a potential risk.

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
New API so should not affect updates / rollbacks.

## Implementation details / challenges (in short)
Most of the implementation can be fairly simple, but there are a couple of things that better be in mind

* Storage:
    * Things are pretty easy thanks to CDI's [smart-cloning](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/smart-clone.md).
With smart cloning we can simply reference the existing DVs to smart-clone them. Smart cloning is supported only for
offline usage, therefore it is a current limitation for VM cloning.
      * CDI would have to be installed on the cluster in order to clone disks and volumes.
    *  Everything that is backed by a container-disk can be copied very easily, that includes CDRoms, ephemeral disks,
  etc.
    * Since secrets and config-maps are namespace scoped, if the VM is cloned to the same namespace it's no problem
  to reference the same secret / config-maps in the new VM. This ia a reason for limiting the clone to be at the same
      namespace.
    * Host disks will not be supported.
      
* Network:
  * The only important thing is the MAC address which needs to either be deleted or changed (a MAC address would be
    generated for a new VM that does not specify it).
    
* Metadata:
  * Obviously, name should be changed
  * Also: timestamps, UIDs and similar fields should be deleted
  * Regarding annotations / labels - I think it's better to keep them as-is by default as it's difficult to say which
    one are important to the user. However, this can be configurable.
    
< Please feel free to provide feedback on any of these > 

## Functional Testing Approach
Functional tests can:
* Clone a VM and see it's successful
* Ensuring that an illegal clone is not possible
  * Maybe check that MAC address is not the same as before? or does not exist in the cluster?
* Ensuring that cloned VM has same spec as original VM in terms of number of disks, volumes, etc.

# Implementation Phases
First implementation cycle, as stated above, should allow cloning offline and to the same namespace only.
