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
* Maintain backward compatibility

## Non Goals
* Maintain UUID persistence with VM backup/restore.
* Change UUID of currently existing VMs

## Definition of Users
VM owners: who require consistent firmware UUIDs for their applications.
cluster-admin: to ensure VMs have universally unique firmware IDs

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

3. **Mitigation**:
    - Backups created before the implementation of this mechanism will result in VMs receiving a new UUID upon restore, as the firmware UUID field will be absent in older backups.
    - Users must be aware of this limitation and plan accordingly.

### Workflow
1. **Mutator Webhook**:
    - During the creation of a new VM, if the `vm.spec.template.spec.firmware.uuid` field is not set, the webhook will set it to a random uuid based on the package `"github.com/google/uuid"`.

2. **Controller Logic**:
    - For existing VMs:
        - If the firmware UUID field is absent, the controller will calculate the UUID using the legacy logic and patch the VM template spec.
        - This ensures backward compatibility without disrupting existing workflows.

3. **Backup and Restore**:
    - Backups created before this mechanism will not include the firmware UUID, leading to new UUIDs being generated upon restore.
    - Users must be aware of this limitation and plan accordingly.

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
- **Main Idea**: The UUID would be generated during the VM's first boot but based on both the name and namespace, resulting in a breaking change. Users would be warned over several versions, and tooling would be provided to add the current firmware UUID to the spec to avoid breaking workloads.
- **Pros**:
   - Avoids abusing the spec and keeps it clean.
   - Simple to implement.
- **Cons**:
   - Requires user intervention to avoid breaking existing workloads.

---

### 3. Upgrade-Specific Persistence of Firmware UUID
- **Main Idea**: Just before KubeVirt is upgraded, persist `Firmware.UUID` of existing VMs in `vm.spec`. From that point on, any VM without `Firmware.UUID` is considered new. For new VMs, use `vm.metadata.uid` if `vm.spec.template.spec.Firmware.UUID` is not defined.
- **Pros**:
   - The upgrade-specific code can be removed after two or three releases.
   - Simple logic for handling UUIDs post-upgrade.
   - Limited disruption to GitOps, affecting only pre-existing (and somewhat buggy) VMs.
- **Cons**:
   - Requires additional upgrade-specific code (possibly in `virt-operator`) to handle the persistence.

---

### 4. Dual UUID Logic with Upgrade Annotation
- **Main Idea**: During an upgrade, KubeVirt would use two UUID generation logics:
   1. Fresh clusters would use the non-buggy `vm.metadata.uid`.
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
Backups created before implementing the persistent firmware UUID mechanism will not include the firmware UUID in the VM's spec.
As a result, restoring such backups will generate a new UUID for the VM.
This change may lead to compatibility issues for workloads or systems that rely on consistent UUIDs, such as licensing servers or configuration management systems.
Users are advised to take this into consideration and plan backup and restore operations accordingly. 

## Testing

### Existing Tests
- Verify that a newly created VMI has a unique firmware UUID assigned.
- Ensure the UUID persists across VMI restarts.

### Scenarios Needing Coverage

1. **Old VM Patching**:  
   Validate that old VMs without a firmware UUID receive one calculated using the legacy logic.

2. **New VM Creation**:  
   Verify that the firmware UUID is set to via `"github.com/google/uuid"` when not explicitly defined.


# Implementation Phases

1. **Alpha Phase**
    - Persist firmware UUID in the `VM` template spec field.
    - Implement logic for:
        - **New VMs**: Auto-generate and store random UUID if not explicitly provided.
        - **Existing VMs**: Patch the `vm.spec.template.spec.firmware.uuid` field using the legacy calculation.
    - Cover implementation with unit tests.

2. **Beta Phase**
    - Ensure seamless upgrade compatibility for pre-existing VMs.
    - Communicate changes through release notes and detailed documentation.

3. **GA Phase**
    - in two versions we can assume that all clusters are already upgraded and the old code can be dropped.

4. **Feature Gate**
    - **No Feature Gate Protection**: The behavior will be enabled by default upon implementation.  
