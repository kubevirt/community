# Overview
This design proposal extends KubeVirt APIs to enable the storage migration between PVCs.

## Motivation
In some circumstances, the guest owner may want to have the option of switching the storage provider or class for a group of volumes used by an active VM. Currently, the VM must be turned off and recreated, causing service disruptions and downtime.


## Goals
Design and develop an API to live update and migrate PVCs.
  * The volume migration needs to be declarative and [GitOps](https://kubevirt.io/user-guide/operations/gitops/) compatible
  * Volume migration implies an update of the volumes used by the VM. Hence, we want to trigger the storage migration on a VM update.
  * The VM spec updates needs to be triggered by the user and not to be implicitly updated by the KubeVirt controller

 ## Non  Goals
*  This feature is not directly aimed at live migrating the VM using RWO volumes. Users will therefore continue to run into the restriction that the VM cannot be live-migrated if it uses RWO volumes.
*  This design doesn't cover migration between storage classes, but rather implements basic APIs that can be used by overlying tooling.
*  This design doesn't include the creation of the destination PVCs, it expects the destination PVCs to exist.

 ## Definition of Users
This feature is for all KubeVirt users who want to move the PVCs assigned to a VM without workload interruptions.

## User Stories

  1. As a VM owner, the storage class of volumes used by a running VM is deprecated, and the volumes need to be moved to another storage class.
  2. As a VM owner, I want to use higher performance storage class that is now available in the cluster.
  3. As a VM owner, I see my non-resizable storage device filling up and I want to move my VM to a larger storage device without interruptions in order to give it more storage.
  4. As an infra admin, I need to replace the storage array due to a hardware refresh and I want to do this without stopping the VMs in order to make this infra operation transparent to VM owners.

## Repos

  - [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design

This proposal introduces a new field in the VM spec to represent the volume update strategy called `updateVolumesStrategy`. This value is a unique field per VM, it can takes the values:
 * `replacement` when the update of the volume aims to substitute the volume with another upon a VM restart
 * `migration` when the update of the volume wants to migrate a volume to another one performing a full-copy of the content

If no value is specified, the default behavior is `replacement`.

Possibly, in the future, we could add a strategy for hotplugged volumes when the update wants to immediately add a new volume to a running VM.

## API examples

#### VM Spec API update

Original VM spec
```yaml
  spec
  volumes:
  - name: vol1
    persistentVolumeClaim:
     claimName: src-pvc1
  - name: vol2
    persistentVolumeClaim:
     claimName: src-pvc2
  - name: vol3
     persistentVolumeClaim:
    claimName: src-pvc3
```

Update VM spec:
```yaml
spec:
  updateVolumesStrategy: migration
  volumes:
  - name: vol1
    persistentVolumeClaim:
     claimName: dst-pvc1
  - name: vol2
    persistentVolumeClaim:
     claimName: dst-pvc2
  - name: vol3
    persistentVolumeClaim:
      claimName: src-pvc3
```

In this, example the user is able to update and migrate the first and second volume while the third one remains untouched.

## Flow

The volume migration update strategy follow the same [update sequence](https://github.com/kubevirt/community/blob/main/design-proposals/cpu-hotplug.md#phase-2-triggering-a-live-migration-when-the-number-of-sockets-changes-to-adjust-the-virt-launcher-cpu-resources) proposed for memory and cpu hotplug.

Sequence of events if the `updateVolumesStrategy` is `migration`:

* the user updates the volumes on a vm.spec.template.spec.domain.volumes
* the VM controller determines the volumes which needs to be updated
* the VM controller validates if the volumes can be migrated
* the VM controller updates the volumes accordingly in the active vmi.Spec
* the VMI controller writes the required volumes to migrate in the vmi.Status and set the update condition in the VMI status

Example of the VMI status:
```yaml
  status:
    conditions:
    - lastProbeTime: null
      lastTransitionTime: "2024-04-09T07:51:30Z"
      message: migrate volumes
      status: "True"
      type: VolumesChange
[..]
    migratedVolumes:
    - destinationPVCInfo:
        claimName: dest-pvc-1
        volumeMode: Block
      sourcePVCInfo:
        claimName: src-pvc-1
        volumeMode: Filesystem
      volumeName: volume1
    - destinationPVCInfo:
        claimName: dest-pvc-2
        volumeMode: Block
      sourcePVCInfo:
        claimName: src-pvc-2
        volumeMode: Filesystem
      volumeName: volume2
```

* the VMI workload-update controller detects the update condition `VolumesChange` and triggers the VMI migration. When it creates the VMI migration object it also set the label `kubevirt.io/volume-update-migration: vm-name`.
* the Migration controller constructs a new target pod with the desired volumes
* virt-launcher updates the volumes to migrate in the migration request for LibVirt
* the VMI live migrates
* the VMI status is updated to reflect that the volume update action have completed and the update condition is removed

This feature depends on the `vmRolloutStrategy` which needs to be set to `LiveUpdate`. On the long term, this should become the default behavior.

The label `kubevirt.io/volume-update-in-progress` can be used to filter and identify the VM migrations triggered by a volume update.

The status of the volume update can be monitored by using the label and getting the corresponding VMI migration:
```console
$ kubectl get virtualmachineinstancemigrations -l kubevirt.io/volume-update-in-progress: vmi` --watch=true
NAME                           PHASE       VMI
kubevirt-workload-update-abcd  Running     vmi
```

The volume migration can be aborted by changing the destination migrated volumes back to the previous ones if the migration is still in progress.
If there is a volume update in progress and the volumes don't correspond to the old set, then the update is ignored and the migration is continued. This avoids to create a confusing state for the volumes of the VM during the migration.

Example of the VMI status:
```yaml
  status:
    conditions:
    - lastProbeTime: null
      lastTransitionTime: "2024-04-09T08:00:53Z"
      status: "False"
      type: VolumesChange
    - lastProbeTime: null
      lastTransitionTime: "2024-04-09T08:00:53Z"
      message: volume migration aborted because migrated volumes have changed
      status: "True"
      type: ChangeAbortion
```

If the volume migration has already completed then this is simply considered a new volume update.

## Design details and motivation

### Motivation on the approaches

There a multiple ways to perform the storage migration using libvirt. It is possible to simply copy the storage to the new destination and then swap the original disk with the new one. This approach is the most efficient as the overhead is caused only by the storage copy.

A different approach is to simultaneously move the storage and live-migrate the guest. When the destination storage is not available on the host where the VM is executing, this solution is required. For this case, costs are greater since there is storage as well as live VM migration overhead.

As KubeVirt is designed today and due to Kubernetes design principles of volume pod immutability, it isn't possible to hotplug and unplug new volumes while the pod is still running. KubeVirt already supports volume hotplug by attaching a new volume to a dummy pod and then bind mounting the storage in the `virt-launcher` pod. Sadly, this method fails to address the issue of removing the old PVCs that were expressly declared in the virt-launcher pod.
Declaring all volumes hotpluggable could be a potential remedy, but we prefer to avoid it because it has been shown to be fragile for some storage providers and opaque to Kubernetes.

The only workable and generic solution for this restriction appears to be to live migrate the VM with storage. Live migration involves the creation of a new virt-launcher pod by KubeVirt that has the destination PVCs attached.

This migrating to a new storage provider is an exceptional operation, and very resource intensive. If in the future, the k8s community will consider the volume mutability in the pod spec, then the lighter approach could be taken into consideration.

## Limitations

### Unsupported volumes

Certain types of volumes and disks aren't suitable to be migrated or the feature isn't supported yet. Those migrated volumes will be marked as rejected in the VMI status

The current types of rejected volumes are:
  * Hotpluggable volumes, this case will be considered in a later iteration.
  * Filesystem volumes, currently virtiofs doesn't support live-migration.
  * Shareable volumes, if a disk is shared between multiple VMs, it cannot be safely copied.
  * LUNs disks, originally the disk might support SCSI protocol but the destination PVC class does not. This case will be considered in a later iteration.

### Datavolumes

DataVolumes are used to import the image during VM provisioning, but they may essentially be thought of as PVCs for the rest of the VM life cycle. With this API, we don't support live updates for the migration strategy to datavolumes. Therefore, the users can use this update strategy only when they replace datavolumes with PVCs.

If the VM makes use of datavolume templates, then the template associated with the source volume needs to be removed by the user as well.

Example using datavolumes:

Original VM definition:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-dv
spec:
  dataVolumeTemplates:
  - metadata:
      name: src-pvc
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        volumeMode: Filesystem
        resources:
          requests:
            storage: 2Gi
        storageClassName: local
      source:
        registry:
          url: docker://registry:5000/kubevirt/alpine-container-disk-demo:devel
[..]
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumedisk1

      volumes:
      - dataVolume:
          name: src-pvc
        name: datavolumedisk
```

Valid updated VM spec:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-dv
spec:
[..]
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumedisk1

      volumes:
      - persistentVolumeClaim:
          name: dst-pvc
        name: datavolumedisk
```

The DV is not deleted when the datavolume template is removed from the VM specification. Datavolumes should only be deleted following a successful storage transfer; otherwise, users may inadvertently erase their data before fully copying it.

## Scalability

Given that a significant amount of data may be copied, this functionality may have an impact on network traffic. This might have a further effect on scalability.

## Update/Rollback Compatibility

Old KubeVirt versions won't be able to use this feature and the new added fields
in the API shouldn't interfere with previous versions.

## Functional Testing Approach

Extensively add functional tests for migrating VMs with non-shared storage:
- Migrating from a filesystem PVC to another filesystem PVC
- Migrating from a filesystem PVC to another filesystem PVC with different filesystem overhead on source/target filesystems
- Migrating from a block PVC to a block PVC
- Migrating from a block PVC to a filesystem PVC
- Migrating from a filesystem PVC to a block PVC
- Migrating from a PVC to a larger PVC


## Implementation phases

1. Introduction of the `updateVolumesStrategy` field
2. Add support for LUNs disks
3. Add support for hotplugged volumes
