# Overview
Today's smart cloning methods use a 1:1 cloning method which creates a temporary snapshot of the source PVC for each intended clone.  
Some storage systems (possibly ceph) are designed to scale better with a 1:N method where a single snapshot of the source PVC is used to create multiple clones.  

This design presents a possible way of implementing a flow where a snapshot is created for a golden image and then subsequent CDI smart clones use that snapshot as the PVC DataSource.

## Motivation
Following a [PoC](https://github.com/kubevirt/containerized-data-importer/pull/2355), we have identified a performance improvement that reinforces the above claim for ceph,  
and would like to productize that PoC so other storage systems could capitalize on this.

## Goals
* StorageProfile cloneStrategy will be added to tip off CDI about which storage can capitalize on this type of cloning
* (?) DataVolume will provide an API to opt-in/out of cloning from a cached snapshot

## Non Goals
* Invalidation of source DataVolumes' cached snapshots as these are defined to be immutable by [design](https://github.com/kubevirt/community/blob/main/design-proposals/golden-image-delivery-and-update-pipeline.md?plain=1#L11)

## Definition of Users
This feature is intended for cluster admins who wish to provide boot sources for their users to create VMs from,  
in the most suiting manner for their storage.

## User Stories
* As a KubeVirt user, I want to store a base image as a snapshot and clone from the snapshot instead of a pvc because this scales better on the storage systems used.

## Repos
* **containerized-data-importer**: DataVolume, Snapshot, StorageProfile controllers

# Design
Today we already have a mechanism in the form storage profiles that detects the preferred cloning strategy (`copy/snapshot/csi-clone`).  
The proposition is to extend this mechanism with a new `cached-snapshot` strategy.  
Upon dealing with target DataVolumes that request to clone a certain existing volume, we would calculate whether we should use the new cloning strategy.  
If that is the case, we will lazily decide if we should create a snapshot of the source.

## Snapshot of source does not exist
We create a snapshot and continue the smart clone flow as usual.

## Snapshot of source exists
Proceed to create a PVC out of this snapshot.

## Controller details
Today the datavolume controller passes the responsibility of creating a snapshot to the smart clone controller by creating a VolumeSnapshot object.  
Since we won't be creating a snapshot object at all times, a change is needed.  
The proposed change is to have the smart clone controller watch certain DataVolume fields (annotation/API) and act on their appearance.

## DataVolume level API (?)
No final decision whether we want DataVolume API so would appreciate any input!

We could opt for a seamless implementation via StorageProfiles,
and not introduce DataVolume APIs.

However, if we do,
```yaml
kind: DataVolume
metadata:
  name: clone-from-snap-target
spec:
  source:
    snapshot:
      namespace: snap-ns
      name: snap-name
```
Or directly introduce with sourceRef
```yaml
kind: DataVolume
metadata:
  name: clone-from-snap-target
spec:
  sourceRef:
    kind: DataSource
    name: snap-source
```


## API Examples
(tangible API examples used for discussion)

## Scalability
During PoC testing we discovered an error that stops progress after 700~ clones regarding JWT tokens
```bash
error verifying token: square/go-jose/jwt: validation failed, token is expired (exp)
```

## Update/Rollback Compatibility
Clone operations on new installs will opt-in to this type of cloning if we know for fact it scales better on the storage provisioner  
(Following PoC, we know this holds for ceph rbd).

## Functional Testing Approach
E2E using the existing ceph provisioner installed in kubevirtci clusters,
which performs better with this approach to cloning.

# Implementation Phases
(How/if this design will get broken up into multiple phases)
