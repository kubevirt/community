# Overview
This proposal introduces a mechanism to persist the firmware UUID of a Virtual Machine in KubeVirt.
By storing the firmware UUID, we ensure that it remains consistent across VM restarts.
which is crucial for applications and services that rely on the UUID for identification or licensing purposes.

## Motivation
* Duplicate UUIDs cause burden with third-party automation frameworks such as Gitops tools, making it difficult to maintain unique VM UUID. 
  Such tools have to manually assign unique UUIDs to VMs, thus also manage them externally. (issue)
* The non-persistent nature of UUIDs may trigger redundant cloud-init setup upon VM restore. (issue)
* The non-persistent nature of UUIDs may trigger redundant license activation events during VM lifetime i.e. Windows key activation re-occur when power off and power on the VM.

## Goals
* Improve the uniqueness of the VMI firmware UUID to become independent of the VM name and namespace.
* Maintain the UUID persistence over the VM lifecycle.
* Maintain backward compatibility
* Maintain UUID persistence with VM backup/restore.

## Non Goals


## Definition of Users
VM owners: who require consistent firmware UUIDs for their applications.

## User Stories
### Supported cases:
* Creating and Starting a New VM
* starting and old VM that was not running


## Repos
Kubevirt/kubevirt


## Proposed Solution

### Description
The firmware UUID persistence mechanism will operate as follows:

1. **New VMs**:
    - If the firmware UUID is not explicitly defined in `vm.spec.template.spec.firmware.uuid`, the mutator webhook will automatically set the firmware UUID to the value of `vm.metadata.uid`.

2. **Old VMs**:
    - For VMs created before the upgrade, the VM controller will patch the `vm.spec.template.spec.firmware.uuid` field with a value calculated using the legacy logic (based on the VM name and namespace). This ensures backward compatibility.

3. **Mitigation**:
    - Backups created before the implementation of this mechanism will result in VMs receiving a new UUID upon restore, as the firmware UUID field will be absent in older backups.
    - Users must be aware of this limitation and plan accordingly.

### Workflow
1. **Mutator Webhook**:
    - During the creation of a new VM, if the `vm.spec.template.spec.firmware.uuid` field is not set, the webhook will set it to `vm.metadata.uid`.

2. **Controller Logic**:
    - For old VMs:
        - If the firmware UUID field is absent, the controller will calculate the UUID using the legacy logic and patch the VM template spec.
        - This ensures backward compatibility without disrupting existing workflows.

3. **Backup and Restore**:
    - Backups created before this mechanism will not include the firmware UUID, leading to new UUIDs being generated upon restore.
    - Users must be aware of this limitation and plan accordingly.

--- 


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
   Verify that the firmware UUID is set to `vm.metadata.uid` when not explicitly defined.


# Implementation Phases

1. **Alpha Phase**
    - Persist firmware UUID in the `VM` template spec field.
    - Implement logic for:
        - **New VMs**: Auto-generate and store UUID using `vm.metadata.uid` if not explicitly provided.
        - **Old VMs**: Patch the `vm.spec.template.spec.firmware.uuid` field using the legacy calculation.
    - Cover implementation with unit tests.

2. **Beta Phase**
    - Ensure seamless upgrade compatibility for pre-existing VMs.
    - Communicate changes through release notes and detailed documentation.

3. **GA Phase**
    - Consider deprecating the legacy logic in the future and communicate timelines accordingly.

4. **Feature Gate**
    - **No Feature Gate Protection**: The behavior will be enabled by default upon implementation.  
