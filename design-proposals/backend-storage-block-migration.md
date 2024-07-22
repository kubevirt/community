# Overview
This proposal is about supporting the migration of all VMs that use a persistent feature, such as persistent vTPM.

To make this document as simple as possible, let's consider that storage in Kubernetes can either be filesystem (FS) or
block mode, and either ReadWriteOnce (RWO) or ReadWriteMany (RWX) access mode, with storage classes supporting any
number of combinations of both.

As of KubeVirt 1.3, Here is how each of the 4 combinations is supported by the backend-storage for persistent VM state:
- RWO FS: supported by backend-storage but makes VMs non-migratable
- RWX FS: preferred combination, supported by backend-storage and compatible with VM migrations
- RWO block: not supported yet, but PRs exist to implement it in a non-migratable way 
- RWX block: same as above, with migratability functional but considered unsafe, because it leads to 2 pods (source and target virt-launchers) mounting the same ext3 filesystem concurrently.

This is the best we can do with the current implementation, which involves mounting the same partition on migration source and target.
This proposal is about switching to a different model which would instead create a new backend-storage PVC on migration,
allowing all 4 cases above to have proper migration support.

## Motivation/Goals
- Users want to be able to use any storage class for backend-storage
- Users want all VMs with persistent features to be potentially migratable

## Definition of Users
Any VM owner that wants to use persistent features, such as TPM or EFI

## User Stories
As a user, I want to be able to seamlessly enable persistent features while keeping the migratability of my VMs.

## Repos
kubevirt/kubevirt

# Design
- When creating a new VM with persistent features, we use either (first match wins):
  - The storage class specified in the KubeVirt Custom Resource
  - The storage class marked as the default for virtualization purposes
  - The storage class marked as the default for the cluster
- All new persistent storage PVCs will be created with the name `persistent-state-for-<vm_name>-<random_string>` and the new annotation `persistent-state-for-<vm_name>=true`
- When starting an existing VM with persistent features, we will look for any PVC with the annotation, or any PVC with the legacy name, for backwards compatibility
- For volume/access modes:
  - When multiple volume modes are listed in the StorageProfile of the StorageClass, we will default to filesystem (to avoid partitioning complexity)
  - When multiple access modes are listed for the selected volume mode, we will arbitrarily default to RWO.
  - When CDI is not present, or the StorageClass has no StorageProfile, or the StorageProfile has no modes, we'll default to RWO FS.
- On migration:
  - No matter what volume/access mode the PVC uses, we will create a new empty PVC for the target with the name `persistent-state-for-<vm_name>-<random_string>` and no annotation
  - In the migration object status, we store the names of both the source and target PVCs (see API section below)
  - If a migration succeeds, we set the annotation `persistent-state-for-<vm_name>=true` on the new (target) PVC and delete the old (source) PVC, using the source PVC name from the migration object
  - If a migration fails, we delete the target backend-storage PVC that was just created, using the target PVC name from the migration object
  - In the unlifely event that a VM shuts down towards the very end of a migration, the migration finalizer will decide which PVC to keep and which one to get rid of

## API
The only API that's introduced is a couple status fields in the VirtualMachineMigration object:
- `SourcePersistentStatePVCName`
- `TargetPersistentStatePVCName`

## Scalability
No new scalability concern will be introduced

## Update/Rollback Compatibility
Since the name of the backend-storage PVC will change, we will keep fallback code to look for the legacy PVC.

## Functional Testing Approach
All existing backend-storage-related functional tests will still apply. More tests could be added to ensure all 4 combinations
of FS/block RWX/RWO work for backend-storage, if we think it's worth the additional stress on the e2e test lanes. 

# Implementation Phases
First phase is block support. That effort is well underway already.  
Second phase is changing the PVC handling and could be done as part of the first phase PR or as a separate one.  
Either way, it's important that both phases land in the same KubeVirt release.