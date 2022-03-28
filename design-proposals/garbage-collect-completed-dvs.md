# Overview
DataVolumes are a useful construct for provisioning VM disks but once the PVC is provisioned and populated the DataVolume provides no value. This design presents an approach for garbage collecting completed DataVolumes. 

## Motivation
In the field, customers are confused by DVs after the fact. For example, when someone deletes a PVC that is owned by a DV they are confused when the PVC is re-created.

## Goals
DataVolumes will be garbage collected after their PVC is provisioned and populated.

## Non Goals
(limitations to the scope of the design)

## Definition of Users
(who is this feature set intended for)

## User Stories
* As a VM owner I use a standalone DataVolume to prepare a VM disk PVC for use by my VM. Once the DataVolume is complete, it has no use and can be confusing, so I want the system to garbage collect it. The PVC should still maintain an independent lifecycle from the VM.
* As a VM owner I add an entry to the dataVolumeTemplates section of my VM spec to have a disk prepared for my VM. Once the DataVolume is complete, it has no use and can be confusing, so I want the system to garbage collect it. I want kubevirt to understand that it doesn't need to create a new DataVolume when the underlying PVC exists. The PVC should still share the same lifecycle as the VM.
* As a VM owner I want to restore my backed up VM without needing to recreate the DataVolume.
* As a VM owner I want to replicate my workload to another cluster without needing to mutate the PVC and DV with special annotations in order for them to behave as expected in the new cluster.

## Repos
* **CDI**: Controllers (DataVolume/Import/Upload/Clone/DataImportCron/DataSource) will need adaptations:
** So far the controllers assumed the existance of DV-PVC pairs, which is no longer a valid assumption.
** DataImportCron controller currently re-creates the import DV if deleted, even if the PVC was already populated.
** Warm imports - verify DV is Succeded only after the last cycle
* **KubeVirt**: VM controller adaptations (see below)
* **Tekton tasks** - check how DVs are used to make sure GC by default wouldn't break them
* **UI**: check if DVs are referred after Succeded
* **HCO**: expose CDIConfig `dataVolumeTTLSeconds`

## Configuration
* CDIConfig - minimal time for DV garbage collection after completion (disabled by default, for opt-in)
  `dataVolumeTTLSeconds *int32`
* DataVolume annotation allowing its garbage collection, `true` by default for newly created DVs (via mutating webhook) when `dataVolumeTTLSeconds` is set. User can annotate older DVs to get them garbage collected too.
  `"cdi.kubevirt.io/storage.dv.deleteAfterCompletion": "true"`

  **NOTE**: Since this default for newly created DVs may prove to be risky, we will take a final decision on it later in the development, as it is agnostic to the design.
# Design
Garbage collection is dispatched whenever a DV phase is updated to `Succeeded`. It iterates all `Succeeded` DVs annotated with `deleteAfterCompletion: true`, garbage collects those which their `dataVolumeTTLSeconds` has passed, and returning with `RequeueAfter` if needed.

When a DV is garbage collected:
* PVC ownership by the DV is removed
* If DV has ownership (e.g. by VM, when created via DataVolumeTemplates), it is transferred to its PVC
* Any DV information that might be useful after it is garbage collected is to be annotated on the PVC (TBD)
* DV is deleted

If the PVC population failed, the DV remains the owner of the PVC.

## API Examples
(tangible API examples used for discussion)

## Scalability
(overview of how the design scales)

## Update/Rollback Compatibility
Garbage collection is opt-in in a way that existing DVs are not impacted after an update. We change the fundamental expectations that exist for how DataVolumes behave now, and a change in default behavior is a breakage in compatibility. Therefore, existing KubeVirt + CDI combinations will continue to work even after updating CDI, and new/updated installs can opt-in to the garbage collection behavior if they'd like. The opt-in by CDI Config `dataVolumeTTLSeconds` solves the OpenShift Virtualization use case where we have tight control over the KubeVirt + CDI combinations. We should handle a few odd conditions as a result of this change that could impact compatibility.

### DataVolume volume source
KubeVirt has [DataVolume volume source](https://github.com/kubevirt/kubevirt/blob/d76ed475eec302d6f863a80eeff70db26403ee5b/staging/src/kubevirt.io/api/core/v1/schema.go#L703) for VMIs. In reality, this value translates directly to a PVC name, so we'll just need to make sure there's no validation that is expecting the DV to still exist.

### virt-controller
The existing virt-controller logic re-creates DVs if they don't exist. It will be updated so if the PVC already exists, it won't create a DV. KubeVirt already assumes that if a PVC exists, and it is not owned by a DV, it means the PVC is populated. Note this change doesn't prevent someone from pairing a new version of CDI (which automatically garbage collects) with an old version of KubeVirt (which recreates after garbage collection). This may create controller spin where virt-controller creates and CDI deletes, which is solved by the DV annotation.

## Functional Testing Approach
CDI and KubeVirt functional tests will need adaptations as they currently assume existance of DV-PVC pair

Upgrade testing to a version that garbage collects - verify no PVCs were deleted

# Implementation Phases
(How/if this design will get broken up into multiple phases)
