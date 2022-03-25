# Overview
Today, anything that happens on the host side of VMIs (i.e. virt-launcher) is transient.  
Any change a VMI makes outside of its own file-system is lost after a shutdown.  
Such changes include EFI firmware settings and information stored inside the TPM.  
This design provides a way to persist small VM-specific files used by virt-launcher.

## Motivation
Though we now support exposing an emulated TPM device to guests (vTPM), using the device in any meaningful way would require persisting data across reboots.  
As we do not currently have a way to do that, all data stored by the VMI into the vTPM is lost after a shutdown (a soft reboot keeps the same virt-launcher, avoiding losses).

## Goals
- Being able to persist modifications of various emulated devices across reboots for a given VM
- Enable live migration of VMs that use backend storage
- Avoid consuming large amounts of storage
- Prevent virt-launcher for VM X from seeing data that belongs to VM Y

## Non Goals
- Persisting everything. Some information that would usually persist on a non-ephemeral host just don't matter to us.
- Supporting user modifications of the backend storage. Users whishing to inject their own files should do so at their own risk.

## Definition of Users
This feature is intended for end-users who wish to preserve the state of some emulated device across reboots.

## User Stories
As a VM user, I want to be able to store the encryption key of my VMs partitions inside the vTPM (example: BitLocker)

## Repos
kubevirt/kubevirt

# Design
The storage class to be used for the PVC would be defined in the KubeVirt CR. If none is defined, VMs that need persistent storage would fail to start.  
The storage used for this would be 1 PVC per VM, and would be created on first start of a VM that needs backend storage.  
The PVC would be RWX to ensure a VM can be migrated, but also just started on a node, then stopped and re-started on another node.  
The PVC would be added to the virt-launcher compute container as a volume, but never assigned to the user's VM.  
The PVC would be mounted in a directory at the root of the filesystem. Appropriate symbolic links will be created before starting the libvirt domain.  
The PVC would be owned-by the VM, and therefore will get deleted when its associated VM is deleted.

For now, the only 2 candidates features for backend storage are:
- The persistent TPM state, which takes about 6KiB when freshly created. We have yet to find its maximum size.
- The OVMF vars, which can take up to 528KiB
I propose finding the biggest size these can take, multiply the sum by 2 and round up to the nearest MiB to allow for future use-cases.
The total will likely add up to 2MiB, 3MiB tops.

## API Examples
Currently, the TPM device API doesn't have any option.  
We will add a `persistent` boolean option that will default to false.  
When set to true, the state of the TPM will be persisted.

## Scalability
Every VM will take a couple more MiBs of storage at most, which shouldn't be a concern, even on cluster running a large number of VMs.  

It's important to note however that the CSI should be carefully chosen and support small files.  
Some CSIs may allocate up to a minimum of 1GiB per volume, which would be quite problematic here, or at least a big waste of space.

## Update/Rollback Compatibility
Since backend storage will only be used to implement new VMI APIs, existing VMIs just won't be affected.  
Existing VMs will however be able to be modified to add backend-storage-related functionalities, but those will only take effect after a reboot.
Therefore, updating from a KubeVirt version that doesn't support backend storage to one that does is not a concern, including for running VMIs.

Additionally, migrations will require shared storage support for all the involved components, such as swtpm and parts of libvirt.

## Functional Testing Approach
- Create a VM with a persistent TPM
- Start the VM
- Seal a specific string against a set of PCRs
- Stop the VM
- Start the VM
- Migrate the VM if the cluster supports migration
- Unseal the string and ensure it is correct
- Stop and delete the VM
- Ensure the backend storage PVC gets deleted along with the VM

# Implementation Phases
- Automatically create the PVC on VM creation, with the VM object set as its owner
- Mount the volume in the place(s) where the TPM state file(s) belong
- Add an option to the TPM device API to enable persistence
- Write tests that include reboot, migration and auto-deletion

# Concern
If a TPM-enabled disk encryption software like BitLocker uses PCRs that currently change on hard reboots, there will be more work needed.  
That could happen if for a example a randomly generated serial number is involved. In that case, we will have to find a way to also persist such data.
