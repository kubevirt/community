# Overview
This proposal introduces a mechanism to make the firmware UUID of a Virtual Machine in KubeVirt universally unique.
It ensures the UUID remains consistent across VM restarts, preserving stability and reliability.
This change addresses a bug in the current behavior, where the UUID is derived from the VM name,
leading to potential collisions and inconsistencies for workloads relying on a truly unique and stable UUID.

## Motivation
* Duplicate UUIDs cause burden with third-party automation frameworks such as Gitops tools, making it difficult to maintain unique VM UUID. 
  Such tools have to manually assign unique UUIDs to VMs, thus also manage them externally.
* The non-persistent nature of UUIDs may trigger redundant cloud-init setup upon VM restore.
* The non-persistent nature of UUIDs may trigger redundant license activation events during VM lifetime i.e. Windows key activation re-occur when power off and power on the VM.

## Goals
* Improve the uniqueness of the VMI firmware UUID to become independent of the VM name and namespace.
* Maintain the UUID persistence over the VM lifecycle.
* Maintain UUID persistence with VM backup/restore.
* Maintain backward compatibility

## Non Goals
* Change UUID of currently existing VMs

## Definition of Users
VM owners: who require consistent firmware UUIDs for their applications.
cluster-admins: to ensure VMs have universally unique firmware IDs

## User Stories

### Supported Cases

#### 1. Creating and Starting a New VM
**As a user**,  
I want to create and start a new VM in KubeVirt,  
**so that** it gets a unique, persistent firmware UUID for consistency across reboots, migrations, and clusters.

- **Scenario**: Creating VM `X` in namespace `Y` on Cluster `A` and `B` results in different UUIDs.
- **Expected**: UUIDs are unique universally to avoid conflicts.

---

#### 2. Starting a VM Created by an Older KubeVirt Version
**As a user**,  
I want to start a VM from an older KubeVirt version,  
**so that** it is assigned a persistent UUID if one wasn’t set before.

- **Scenario**: A VM without a UUID starts on an upgraded KubeVirt version.
- **Expected**: The old UUID is persisted in the VM spec to ensure consistency.

#### 3. restoring a VM
**As a user**,  
I want to restore my VM from vmSnapshot.

- **Scenario**: A VM restore resource is created.
- **Expected**: The old UUID is being re-calculated and persisted in the VM spec to ensure consistency.


## Repos
Kubevirt/kubevirt


## Proposed Solution

### Description
The firmware UUID persistence mechanism will operate as follows:

1. **New VMs**:
    - If the firmware UUID is not explicitly defined in `vm.spec.template.spec.firmware.uuid`, the mutator webhook will automatically set the firmware UUID on create operation.
    - We can use what k8s uses to generate uuids via `"github.com/google/uuid"` package.

2. **Existing VMs**:
    - For VMs created before the upgrade, the VM controller will patch the `vm.spec.template.spec.firmware.uuid` field,
   with a value calculated using the legacy logic (based on the VM name).
   This ensures backward compatibility.

3. **Restored VMs**:
    - Backups created before the implementation of this mechanism will result in VMs  having no uuid in the snapshot content resource.
    - the restore controller will detect those scenarios and will recalculate the uuid based on the vm name.

### Workflow
1. **Mutator Webhook**:
    - During the creation of a new VM, if the `vm.spec.template.spec.firmware.uuid` field is not set, the webhook will set it to a random uuid based on the package `"github.com/google/uuid"`.

2. **Validation Webhook**
    - The validation webhook will reject any VM creation or update with an empty UUID.

3. **VM Controller Logic**:
    - For existing VMs:
        - If the firmware UUID field is absent, the controller will calculate the UUID using the legacy logic and patch the VM template spec.
        - This ensures backward compatibility without disrupting existing workflows.

4. **Backup and Restore Controller Logic**:
    - during restore operation the controller will load the snapshot content into memory.
    - the controller will check if the vm resource contains old uuid.
    - if no uuid the controller will add the legacy uuid calculation result to the vm spec before creating/updating the restored VM.
    - this code will be removed when the main VM controller logic will be removed as well.


## Scenario Comparison Table

The following table compares different scenarios using the proposed approach and alternative solutions:

| #      | Scenario                                       | Proposed Design                                                                                                                      | Alternative Solution (Status Field)                                                                                                                           | Alternative Solution (UUID in Annotation)                                                                                                                        |
|--------|------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1      | Existing stopped VM with empty UUID            | VM controller patches VM template spec using legacy logic.                                                                           | Controller checks status field; if missing, uses legacy logic and persists UUID in status.                                                                    | Controller checks annotation; if missing, uses legacy logic and persists UUID in annotation.                                                                     |
| 2      | Existing stopped VM with UUID                  | No operation (UUID is already set and remains unchanged).                                                                            | No operation (UUID is already set and remains unchanged).                                                                                                     | No operation (UUID is already set and remains unchanged).                                                                                                        |
| 3      | Existing running VM with empty UUID            | VM controller patches VM template spec using legacy logic.                                                                           | The VMI FW ID will be updated according to the status field, if it's missing then we fall back to legacy calculation and then persist it in the status field. | The VMI FW ID will be updated according to a dedicated annotation, if it's missing then we fall back to legacy calculation and then persist it in the annotation |
| 4      | Restore from existing VM snapshot              | If a restored VM lacks a UUID, the restore controller injects UUID into the VM template spec using legacy logic **before creation**. | If a restored VM lacks a UUID, the restore controller checks `status`. If missing, legacy logic is used.                                                      | If a restored VM lacks a UUID, the restore controller checks `annotations`. If missing, legacy logic is used.                                                    |
| 5      | New stopped VM with empty UUID                 | Mutating webhook injects a new UUID into the VM template spec.                                                                       | Mutating webhook injects a new UUID into the VM template spec and persists it in status.                                                                      | Mutating webhook injects a new UUID into the VM template spec and persists it in an annotation.                                                                  |
| 6      | New stopped VM with UUID                       | No operation (UUID remains unchanged).                                                                                               | No operation (UUID remains unchanged).                                                                                                                        | No operation (UUID remains unchanged).                                                                                                                           |
| 7      | New running VM                                 | No operation (UUID remains unchanged because the mutation webhook already set it).                                                   | No operation (UUID remains unchanged because the mutation webhook already set it).                                                                            | No operation (UUID remains unchanged because the mutation webhook already set it).                                                                               |
| 8      | Restore from new VM snapshot                   | No operation (UUID is restored as part of the snapshot).                                                                             | No operation (UUID is restored from the spec, since the mutating webhook already set the UUID).                                                               | No operation (UUID is restored from the spec, since the mutating webhook already set the UUID).                                                                  |
| 9      | User is trying to modify UUID in spec          | Users can modify uuid in the spec, no extra validation logic needed.                                                                 | Users can modify uuid in the spec, no extra validation logic needed.                                                                                          | Users can modify uuid in the spec, no extra validation logic needed.                                                                                             |
| 10     | User is trying to delete UUID from spec        | **Validation webhook rejects deletion. UUID cannot be removed.**                                                                     | **Users can remove UUID from spec**                                                                                                                           | **Users can remove UUID from spec.**                                                                                                                             |
| 11     | New stand-alone VMI is created with UUID       | No operation (UUID is preserved as provided).                                                                                        | No operation (UUID is preserved as provided).                                                                                                                 | No operation (UUID is preserved as provided).                                                                                                                    |
| 12     | Stand-alone VMI with empty UUID                | Mutating webhook generates and assigns a new UUID.                                                                                   | Mutating webhook generates and assigns a new UUID.                                                                                                            | Mutating webhook generates and assigns a new UUID.                                                                                                               |
| **13** | **Backward Compatibility of Restore & Backup** | ✅ **Fully compatible** – Legacy restore and backup methods continue to work; UUID is injected into `spec`.                           | ❌ **Breaks for old snapshots** without UUID in spec. External restorers must fetch UUID from `status` or use legacy logic.                                    | ❌ **Breaks for old snapshots** without UUID in spec. External restorers must fetch UUID from `annotations` or use legacy logic.                                  |

--- 

## Previously Proposed Solutions

Here are a few alternatives that were previously considered:

### 1. Create a New Firmware UUID Field Under Status
- **Main Idea**: The UUID would be generated during the VM's first boot and saved to a dedicated firmware UUID field in the status.
- **Pros**:
   - Avoids abusing the spec.
   - Keeps the spec clean.
- **Cons**:
   - Adds a new API field, which is difficult to remove and could increase load, hurt performance, and make the API less compact.

---

### 2. Introduce a Breaking Change
- **Main Idea**: The UUID would be generated during the VM's creation using the `"github.com/google/uuid"` package, resulting in a breaking change.
  Users would be warned over several versions, and tooling would be provided to add the current firmware UUID to the spec to avoid breaking workloads.
- **Pros**:
   - Avoids abusing the spec and keeps it clean.
   - Simple to implement.
- **Cons**:
   - Requires user intervention to avoid breaking existing workloads.

---

### 3. Upgrade-Specific Persistence of Firmware UUID
- **Main Idea**: Just After KubeVirt is upgraded, persist `Firmware.UUID` of existing VMs in `vm.spec`.
From that point on, any VM without `Firmware.UUID` is considered new. For new VMs, use `"github.com/google/uuid"` package to generate new uuid if `vm.spec.template.spec.Firmware.UUID` is not defined.
- **Pros**:
   - The upgrade-specific code can be removed after two or three releases.
   - Simple logic for handling UUIDs post-upgrade.
   - Limited disruption to GitOps, affecting only pre-existing (and somewhat buggy) VMs.
- **Cons**:
   - Requires additional upgrade-specific code (possibly in `virt-operator`) to handle the persistence.

---

### 4. Dual UUID Logic with Upgrade Annotation
- **Main Idea**: During an upgrade, KubeVirt would use two UUID generation logics:
   1. Fresh clusters would use the non-buggy `"github.com/google/uuid"` package.
   2. Upgraded clusters would have the KubeVirt CR annotated with `kubevirt.io/use-buggy-uuid`. Clusters with this annotation would continue using the buggy UUID logic.
- **Pros**:
   - Prevents the bug from spreading to new clusters.
   - Buys time to think of a more robust solution.
- **Cons**:
   - Does not solve the problem for existing clusters.
   - Adds complexity by requiring dual logic for UUID generation.


## Scalability
The proposed changes have no anticipated impact on scalability capabilities of the KubeVirt framework

## Update/Rollback Compatibility
Since this design includes additional logic in the restore controller to preserve UUID compatibility during existing backups, it does not impact updates or rollbacks.

## Testing

### Existing Tests
- Verify that a newly created VMI has a unique firmware UUID assigned.
- Ensure the UUID persists across VMI restarts.

### Scenarios Needing Coverage

1. **Old VM Patching**:  
   Validate that old VMs both Running/Stopped without a firmware UUID receive one calculated using the legacy logic.

2. **New VM Creation**:  
   Verify that the firmware UUID is set to via `"github.com/google/uuid"` when not explicitly defined.


# Implementation Phases

1. **First Phase**
    - Persist firmware UUID in the `VM` template spec field.
    - Implement logic for:
        - **New VMs**: Auto-generate and store a random UUID in the mutation-webhook if not explicitly provided by the user.
        - **Existing VMs**: Patch the `vm.spec.template.spec.firmware.uuid` field using the legacy calculation.
        - **Restored VMs**: uuid will be added to the restored VMs that don't have desired uuid
    - Cover implementation with unit tests.
    - add e2e test for uuid persistence for restored vms

2. **Second Phase**
    - after some releases we can discuss when we should drop the old legacy calculation logic.

3. **Feature Gate**
    - **No Feature Gate Protection**: The behavior will be enabled by default upon implementation.  
