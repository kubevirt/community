# Overview
This design proposal extends KubeVirt APIs to enable the storage migration between PVCs.

## Motivation
In some circumstances, the guest owner may want to have the option of switching the storage provider or class for a group of volumes used by an active VM. Currently, the VM must be turned off and recreated, causing service disruptions and downtime.


## Goals
Design and develop an API to handle live storage migration between PVCs.

 ## Non  Goals
*  This feature is not directly aimed at live migrating the VM using RWO volumes. Hence, the API covers and refers only to the volume migration.
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
This proposal introduces a new CRD called `VolumeMigration`

## Volume migration custom resource

```golang
type VolumeMigration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              VolumeMigrationSpec `json:"spec" valid:"required"`
	Status VolumeMigrationStatus `json:"status,omitempty"`
}

type VolumeMigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items []VolumeMigration `json:"items"`
}

type SourceReclaimPolicy string

const (
	SourceReclaimPolicyDelete SourceReclaimPolicy = "Delete"
	SourceReclaimPolicyRetain SourceReclaimPolicy = "Retain"
)

type MigratedVolume struct {
	SourceClaim              string                   `json:"sourceClaim" valid:"required"`
	DestinationClaim         string                   `json:"destinationClaim" valid:"required"`
	SourceReclaimPolicy SourceReclaimPolicy `json:"sourceReclaimPolicy,omitempty"`
}

type VolumeMigrationSpec struct {
	MigratedVolume []MigratedVolume `json:"migratedVolume,omitempty"`
}

const (
	ReasonRejectHotplugVolumes           = "Hotplug volumes aren't supported to be migrated yet"
	ReasonRejectShareableVolumes         = "Shareable disks aren't supported to be migrated"
	ReasonRejectFilesystemVolumes        = "Filesystem volumes aren't supported to be migrated"
	ReasonRejectLUNVolumes               = "LUN disks aren't supported to be migrated yet"
	ReasonRejectedMultipleVMIs           = "Volumes need to belong to the same VMI"
	ReasonRejectedPending                = "Volumes need to belong to a running VMI"
	ReasonRejectedMultipleVMIsAndPending = "Volumes need to belong to the same running VMI"
)

type VolumeMigrationPhase string

const (
	VolumeMigrationPhasePending    VolumeMigrationPhase = "Pending"
	VolumeMigrationPhaseScheduling VolumeMigrationPhase = "Scheduling"
	VolumeMigrationPhaseRunning    VolumeMigrationPhase = "Running"
	VolumeMigrationPhaseSucceeded  VolumeMigrationPhase = "Succeeded"
	VolumeMigrationPhaseFailed     VolumeMigrationPhase = "Failed"
	VolumeMigrationPhaseUnknown    VolumeMigrationPhase = "Unknown"
)

type MigratedVolumeValidation string

const (
	MigratedVolumeValidationValid    MigratedVolumeValidation = "Valid"
	MigratedVolumeValidationPending  MigratedVolumeValidation = "Pending"
	MigratedVolumeValidationRejected MigratedVolumeValidation = "Rejected"
)

type VolumeMigrationPhaseTransitionTimestamp struct {
	Phase                    VolumeMigrationPhase `json:"phase,omitempty"`
	PhaseTransitionTimestamp metav1.Time          `json:"phaseTransitionTimestamp,omitempty"`
}

type VolumeMigrationState struct {
	MigratedVolume `json:",inline"`
	Validation     MigratedVolumeValidation `json:"validation,omitempty"`
	Reason         *string                  `json:"reason,omitempty"`
}

type VolumeMigrationStatus struct {
	VolumeMigrationStates       []VolumeMigrationState                    `json:"volumeMigrationStates,omitempty"`
	VirtualMachineInstanceName  *string                                   `json:"virtualMachineInstanceName,omitempty"`
	VirtualMachineMigrationName *string                                   `json:"virtualMachineMigrationName,omitempty"`
	StartTimestamp              *metav1.Time                              `json:"startTimestamp,omitempty"`
	EndTimestamp                *metav1.Time                              `json:"endTimestamp,omitempty"`
	Phase                       VolumeMigrationPhase                      `json:"phase,omitempty"`
	PhaseTransitionTimestamps   []VolumeMigrationPhaseTransitionTimestamp `json:"phaseTransitionTimestamps,omitempty"`
}
```

Users may define more than one migrated volume, and those volumes may be associated with the same VMI, a different VMI, or none at all. The volumes defined on the same VMI will be categorized and grouped by the volume migration controller.

At the moment, PVCs can only be migrated using this API if they belong to a single running virtual machine. Consequently, all volume migration objects that have migrated volumes that are associated with different VMIs will be rejected by the controller. The object is in the `Pending` state if none of the migrated volumes have been assigned to a VMI, since the VMI might be created at a later time.

The `VolumeMigration` expects a list of source and destination PVCs and the destination PVCs need to be created either manually by the user or other tools.

According to the `SourceReclaimPolicy`, KubeVirt will determine how to handle the source PVCs. The source PVC will be automatically deleted under the `Delete` policy, but it will remain untouched under the `Retain` policy.
If the user doesn't specify any reclaim policy, the `Retain` policy will be set as default.

The core idea is that it is KubeVirt responsibility to handle the lifecycle of the source PVCs as it is the only component able to orchestrate and set the migration as successful.

Certain types of volumes and disks aren't suitable to be migrated or the feature isn't supported yet. Those migrated volumes will be marked as rejected.
If there are other valid volumes belonging to the same VMI as the rejected ones, the VM migration won't be created for those volumes as well. Those volumes will be listed as `Valid` under the `VolumeMigrationStates` field.

The current types of rejected volumes are:
  * Hotpluggable volumes, this case will be considered in a later iteration.
  * Filesystem volumes, currently virtiofs doesn't support live-migration.
  * Shareable volumes, if a disk is shared between multiple VMs, it cannot be safely copied.
  * LUNs disks, originally the disk might support SCSI protocol but the destination PVC class does not. This case will be considered in a later iteration.

The `VolumeMigrationStatus` reports the phase of the migration:
  * `Pending` if there are no VMI associated to the migrated volumes
  * `Scheduling` if the VMI migration object has been created but the migration hasn't started yet
  * `Running` once the VMI migration starts
  * `Succeeded` if the VMI migration has succeeded
  * `Failed` if the VMI migration has failed or there are rejected volumes

## Design details and motivation

### Motivation on the approaches

There a multiple ways to perform the storage migration using libvirt. It is possible to simply copy the storage to the new destination and then swap the original disk with the new one. This approach is the most efficient as the overhead is caused only by the storage copy.

A different approach is to simultaneously move the storage and live-migrate the guest. When the destination storage is not available on the host where the VM is executing, this solution is required. For this case, costs are greater since there is storage as well as live VM migration overhead.

As KubeVirt is designed today and due to Kubernetes design principles of volume pod immutability, it isn't possible to hotplug and unplug new volumes while the pod is still running. KubeVirt already supports volume hotplug by attaching a new volume to a dummy pod and then bind mounting the storage in the `virt-launcher` pod. Sadly, this method fails to address the issue of removing the old PVCs that were expressly declared in the virt-launcher pod.
Declaring all volumes hotpluggable could be a potential remedy, but we prefer to avoid it because it has been shown to be fragile for some storage providers and opaque to Kubernetes.

The only workable and generic solution for this restriction appears to be to live migrate the VM with storage. Live migration involves the creation of a new virt-launcher pod by KubeVirt that has the destination PVCs attached.

This migrating to a new storage provider is an exceptional operation, and very resource intensive. If in the future, the k8s community will consider the volume mutability in the pod spec, then the lighter approach could be taken into consideration.

### Design

This feature adds an additional controller for the CRD `Volume Migration` to virt-controller. Once the VolumeMigration object is created then KubeVirt will update the VMI status with the source and destination volume pairs.

Afterwards, the controller will create a VirtualMachineInstanceMigration to trigger the VM live migration.

If there are some migrated volumes in the VMI status, the VM migration controller will replace the destination volumes by creating the target virt-launcher pod with the new volumes and virt-launcher will select the source volumes to be migrated.

Once the migration completes, the controller will update the migrated volumes first in the VMI spec. This will trigger the update of the volumes in the VM spec if it exists.

DataVolumes are taken into account in the current proposal. DataVolumes are used to import the image during VM provisioning, but they may essentially be thought of as PVCs for the rest of the VM life cycle. According to this proposal, KubeVirt will replace the source DataVolumes with the destination PVCs.

If the VM makes use of datavolume templates, then the template associated with the source volume will be removed when the VM specification is updated.

The VM will continue owning the source datavolumes that have a `Retain` source policy applied to them. Therefore, those DVs and PVCs are also gathered as garbage after the VM is removed.

A schematic representation of the migration flow is shown below.
![](volume-migration.png)

### Limitations

Upon completion of the storage migration, the VM specification is dynamically updated by the VM controller. The VM update may deviate from the original declaration and result in an inconsistency with the definition in the repository if the VM was generated using [GitOps](https://kubevirt.io/user-guide/operations/gitops/) or a comparable declarative technique.

Overlaying tools are responsible to monitor and update the VM declaration with the change caused by this API. To prevent the old version of the VM from being redeployed, this update must be considered before the repository containing the VM declaration is synchronized again.

### API examples

Example of VolumeMigration flow

Create a VM with source storage and the destination PVC:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: dest-pvc
spec:
  accessModes:
    - ReadWriteOnce
  volumeMode: Filesystem
  resources:
    requests:
      storage: 1Gi
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-alpine-datavolume
  name: vm-alpine-datavolume
spec:
  dataVolumeTemplates:
  - metadata:
      creationTimestamp: null
      name: source-pvc
    spec:
      pvc:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: local
      source:
        registry:
          url: docker://registry:5000/kubevirt/alpine-container-disk-demo:devel
  running: true
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-alpine-datavolume
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumedisk1
          interfaces:
          - masquerade: {}
            name: default
        resources:
          requests:
            memory: 128Mi
      networks:
      - name: default
        pod: {}
      terminationGracePeriodSeconds: 0
      volumes:
      - dataVolume:
          name: source-pvc
        name: datavolumedisk1

```

Trigger the migration with:
```yaml
apiVersion: storage.kubevirt.io/v1alpha1
kind: VolumeMigration
metadata:
  name: vol-mig
spec:
  migrationVolumes:
  - sourceClaim: source-pvc
    destinationClaim: dest-pvc
    sourceReclaimPolicy: Delete
```

Once the migration succeed:

Changes on the VMI:
```yaml

apiVersion: v1
items:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachineInstance
[...]
    volumes:
    - name: datavolumedisk1
      persistentVolumeClaim:
        claimName: dest-pvc
 status:
    migrationState:
      completed: true
      endTimestamp: "2023-08-17T08:14:45Z"
```

On the VM:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
[...]
      volumes:
      - name: datavolumedisk1
        persistentVolumeClaim:
          claimName: dest-pvc

```

#### Examples for the Volume Migration status

1. If the migrated volumes belong to multiple VMIs:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm1
spec:
  template:
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumedisk
      volumes:
      - dataVolume:
          name: src-pvc1
        name: datavolumedisk
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm2
spec:
  template:
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            name: datavolumedisk
      volumes:
      - dataVolume:
          name: src-pvc2
        name: datavolumedisk
---
apiVersion: storage.kubevirt.io/v1alpha1
kind: VolumeMigration
metadata:
  name: vol-mig
spec:
  migratedVolume:
  - sourceClaim: src-pvc1
    destinationClaim: dest-pvc1
  - sourceClaim: src-pvc2
    destinationClaim: dest-pvc2
status:
  volumeMigrationStates:
  - migratedVolume:
      sourceClaim: src-pvc1
      destinationClaim: dest-pvc1
      sourceReclaimPolicy: "Delete"
    validation: "Rejected"
    reason: "Migrated volumes need to belong to the same VMI"
  - migratedVolume:
      sourceClaim: src-pvc2
      destinationClaim: dest-pvc2
      sourceReclaimPolicy: "Delete"
    validation: "Rejected"
    reason: "Migrated volumes need to belong to the same VMI"
  phase: Failed
  phaseTransitionTimestamps:
  - phase: Failed
    phaseTransitionTimestamp: "2024-01-12T08:54:11Z"
```

2. Pending volumes if there are no VMIs:
```yaml
apiVersion: storage.kubevirt.io/v1alpha1
kind: VolumeMigration
metadata:
  name: vol-mig
spec:
  migratedVolume:
  - sourceClaim: src-pvc1
    destinationClaim: dest-pvc1
  - sourceClaim: src-pvc2
    destinationClaim: dest-pvc2
status:
  volumeMigrationStates:
  - migratedVolume:
      sourceClaim: src-pvc1
      destinationClaim: dest-pvc1
      sourceReclaimPolicy: "Delete"
    validation: "Pending"
    reason: "No VMI associated to the volume"
  - migratedVolume:
      sourceClaim: src-pvc2
      destinationClaim: dest-pvc2
      sourceReclaimPolicy: "Delete"
    validation: "Pending"
    reason: "No VMI associated to the volume"
  phase: Pending
  phaseTransitionTimestamps:
  - phase: Pending
    phaseTransitionTimestamp: "2024-01-12T08:54:11Z"
```

3. Some of the volumes are rejected and some are valid:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm
spec:
  template:
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: virtio
            shareable: true
            name: datavolumedisk1
          - disk:
              bus: virtio
            name: datavolumedisk2
          filesystems:
          - name:datavolumedisk3

      volumes:
      - dataVolume:
          name: src-pvc1
        name: datavolumedisk1
      - dataVolume:
          name: src-pvc2
        name: datavolumedisk2
      - dataVolume:
          name: src-pvc3
        name: datavolumedisk3
---
apiVersion: storage.kubevirt.io/v1alpha1
kind: VolumeMigration
metadata:
  name: vol-mig
spec:
  migratedVolume:
  - sourceClaim: src-pvc1
    destinationClaim: dest-pvc1
  - sourceClaim: src-pvc2
    destinationClaim: dest-pvc2
  - sourceClaim: src-pvc3
    destinationClaim: dest-pvc3
status:
  volumeMigrationStates:
  - migratedVolume:
      sourceClaim: src-pvc1
      destinationClaim: dest-pvc1
      sourceReclaimPolicy: "Delete"
    validation: "Rejected"
    reason: "Shareable disks aren't supported to be migrated"
  - migratedVolume:
      sourceClaim: src-pvc2
      destinationClaim: dest-pvc2
      sourceReclaimPolicy: "Delete"
    validation: "Valid"
  - migratedVolume:
      sourceClaim: src-pvc3
      destinationClaim: dest-pvc3
      sourceReclaimPolicy: "Delete"
    validation: "Rejected"
    reason: "Filesystem volumes aren't supported to be migrated"
  phase: Failed
  phaseTransitionTimestamps:
  - phase: Failed
    phaseTransitionTimestamp: "2024-01-12T08:54:11Z"
```

## Scalability

This proposal shouldn't have any impact on scalability.

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

# Implementation Phases
1. Introduction of the `VolumeMigration` object and storage migration controller for disks (no lun)
2. Add support for hotplugged volumes
3. Add support for LUNs disks
