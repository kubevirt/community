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
* Creating and Starting a New VM shoul
* starting and old VM that was not running
* restoring an old VM
* restoring a new VM


## Repos
Kubevirt/kubevirt

# Design

This section provides a range of alternative solutions for implementing a persistent, universally unique firmware UUID for VMIs.
Option 1 is preferable as it avoids new API fields, ensures backward compatibility, and has minimal drawbacks.

### 1. Persist UUID in the VM's Spec Field

**Description:**

if the FW ID is not specified in `vm.spec.template.spec.firmware.uuid` it shall be generated and stored in `vm.spec.template.spec.firmware.uuid`.

To prevent disruptions caused by introducing a new generation logic for firmware UUIDs, we propose a multiphase approach:
* For new VMs, which are started for the first time, add the firmware UUID to the VM template spec if it is not specified already. Consider this a required field which if not specified, is auto-filled.
  For old VMs, which are currently running, the VM controller reads the VMI/domain firmware and fills the VM template field.
  For old VMs, which are stopped and now started, perform the legacy calculation (which used the name as the seed) to fill the VM template field.
* use alerts / VM conditions in order to notify the users to restart VMs that do not have firmware UUID in their spec.
* After a while, let the community know (i.e. via a mail to mailing list) that in the next version firmware UUID must reside in spec.
* Switch to using vm.metadata.uid as the new firmware UUID (unless a UUID is already provided in the VM's spec, therefore keeping backward compatibility).

This phased approach ensures a smooth transition for users while maintaining backward compatibility and minimizing disruptions to workloads.

**Pros:**

- **Simple Implementation:** Requires minimal changes to existing logic.
- **Backward Compatibility:** Preserves existing UUIDs by storing them in the spec.
- **No New API Fields:** Avoids adding new fields to the API.

**Cons:**

- **Spec Misuse:** Spec fields should not store system-generated data.
- **Low Relevance:** Users may not see the UUID as part of the desired state.

---

### 2. Persist Firmware UUID in VM Status Field

**Description:**

Introduce a new field, vm.status.firmwareUUID, to store the firmware UUID of a VMI.
The VM controller stores the firmware UUID in `vm.status.firmwareUUID`. This ensures the UUID remains consistent across VM restarts.

- If `vm.spec.template.spec.firmware.uuid` is set, the controller propagates it to `vmi.spec.firmware.uuid`.
- If not, the controller:
  - For new VMs: will Generate a UUID using `vm.metadata.uid`.
  - For old VMs: will use `vm.status.firmwareUUID` for restarts (based on the legacy calculation using the vm.name as seed).
- During sync, the controller updates `vm.status.firmwareUUID` using `vmi.spec.firmware.uuid`.

this approach will also be multiphase same as the one above.

**Pros:**

- **Backward Compatibility:** Works with existing VMs without disruption.
- **GitOps Friendly:** Keeps the spec clean and uses status for system data.

**Cons:**

- **API Expansion:** Adds a new status field, requiring long-term support.

---

### 3. Upgrade-Specific Persistence of Firmware UUID

**Description:** Before upgrading KubeVirt, persist the `Firmware.UUID` of existing VMs in `vm.spec`.
After the upgrade, any VM without `Firmware.UUID` is considered new.
For these new VMs, use `vm.metadata.uid` as the firmware UUID if `vm.spec.template.spec.Firmware.UUID` is not defined.

**Pros:**
- Upgrade-specific code can be removed after two or three releases.
- Maintains a straightforward logic post-upgrade, with minimal changes required.
- Limits disturbance to GitOps workflows to pre-existing VMs that had potentially buggy UUID behavior.

**Cons:**
- Requires additional code (likely in `virt-operator`) to handle the upgrade-specific persistence.
- Code should be added to the computation of the "restart required" condition, so that it is not raised if vm.Template.Firmware.UUID equals vmi.Firmware.UUID.

--- 

## API Examples
**Example uuid persistence within VM Status**

```yaml
apiVersion: kubevirt.io/v1alpha3
kind: VirtualMachine
metadata:
  name: mytestvm
status:
   conditions:
      - lastProbeTime: "2024-11-06T01:12:29Z"
        lastTransitionTime: "2024-11-06T01:12:29Z"
        message: Guest VM is not reported as running
        reason: GuestNotRunning
        status: "False"
        type: Ready
   created: true
   runStrategy: Once
   firmwareUUID: "123e4567-e89b-12d3-a456-426614174000"
```

## Scalability
The proposed changes have no anticipated impact on scalability capabilities of the KubeVirt framework

## Update/Rollback Compatibility

### What Happens to the Firmware UUID During a Restore?
* The firmware UUID is part of the VM's spec, or dynamically generated by the VM controller if it is not defined.
* Snapshots created before implementing the firmware UUID persistence mechanism will not include this field.
* As a result, during a restore, a new UUID may be generated if the vm.spec.template.spec.firmware.uuid field is absent.

**Note:** as long as we are focusing on just the persistence phase, we ensure that existing behavior remains unchanged, avoiding restore issues while laying a foundation for future UUID management enhancements.

## Functional Testing Approach

### Existing Tests
- Verify that a newly created VMI has a unique firmware UUID assigned.
- Ensure the UUID persists across VMI restarts.

### Scenarios Needing Coverage
- **Existing VMs (Pre-Fix):**  
  Test scenarios where VMs created before the fix are started after the fix to ensure they maintain the same firmware UUID as before the update.
  - Validate that UUIDs are preserved without any changes when these VMs begin running.
  - Confirm that UUID persistence remains consistent across restarts post-update.

# Implementation Phases
1. **Alpha Phase**
  - Persist firmware UUID in the `VM` template spec field.
  - Implement logic for:
    - **New VMs:** Auto-generate and store UUID if not provided.
    - **Running Old VMs:** Retrieve UUID from `VMI`/domain and persist in spec.
    - **Stopped Old VMs:** Use legacy calculation to populate UUID when started.
  - Cover with unit tests and basic e2e tests.

2. **Beta Phase**
  - Ensure seamless upgrade compatibility for pre-existing VMs.
  - Expand e2e tests for migrations, backups, and restores.
  - Communicate changes in release notes and documentation.

3. **GA Phase**
  - Finalize implementation and remove legacy UUID logic.
  - Validate with performance and compatibility tests.

4. **Feature Gate**
  - **No FG protection planned**—behavior will be default upon implementation.  