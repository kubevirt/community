# Overview
This proposal is about having a better integration with kubernetes storage APIs.  
With recent support of kubernetes [volume populators](https://kubernetes.io/blog/2022/05/16/volume-populators-beta/) we have implemented [CDI volume populators](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/cdi-populators.md) which allows population of PVC the kubernetes way without the need for DataVolumes.  
Along major benefits of using this populators we loss one of DataVolumes greatest achievements - generating a PVC with the optimal properties fitting the user's storage according to [Storage Profile](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/storageprofile.md).  
This design should help us gain the best of all worlds.

## Motivation
DataVolumes were our only way of having a populated PVC with a VM disk image. While solving this crucial problem for VM's in kubernetes, Datavolumes certainly have downsides, mostly with regard to integrating with backup/Disaster Recovery solutions. Now that kubernetes has a general solution to the specific problem that DataVolumes solve we can prevent from using Datavolumes.  
Harnessing k8s and CDI new volume populators API along with our knowledge of storage capabilities will allow the user to enjoy synchronization with k8s and better compatibility with GitOps, Backup and Restore, and Disaster Recovery workflows as well as the smart creation of PVCs that best suits the user's storage.  

## Goals
- Provide a method for creating VMs with PVCs that suits best the user's storage while using CDI volume populators.

## Non Goals

## User Stories
* As a user creating VMs, I want to be able to create a PVC within the VM spec while getting the best PVC spec that is also compatible with backup and restore solutions.

## Repos
* **CDI**: Provide API to that receives a PVC spec and in case of missing properties such as VolumeMode and AccessModes fills the optimal values according to the storage class.
* **CDI**: Update DataImportCron API to create/update VolumeCloneSources that can be consumed as sources for PVC handled by CDI volume populators. VolumeCloneSource will need to be able to be updated with different PVC source. Need to wait for [k8s cross namespace support](https://kubernetes.io/blog/2023/01/02/cross-namespace-data-sources-alpha/) and update clone populator accordingly to allow clone from golden images namespace.
* **Kubevirt**: Provide API to create a VM with minimal volume information and receive a VM with best suited populated PVC. Adjust InstanceTypes and preferences API to correctly create VM.
* **Kubevirt**: (optional) Provide virtctl command to generate and create PVCs.
* **Kubevirt**: Extend virtctl VM create command to use the new API.
* **SSP**: Manage setting up VolumeCloneSources for commonly supported operating systems updated by DataImportCrons.

*>Note:* Any repo or component that creates VMs (UI, virtctl, tekton pipelines, etc) may want to adjust and prefer using this new API in the VM spec. It should be noted that this can be done incrementally since we do not propose to eliminate the dataVolumeTemplates section right away.

# Design

A new field will be added to the VM object called **VolumeClaimTemplates**. This field will allow the user a way to define and create a PVC owned by the VM while providing partial/full PVC spec embedded in the VM spec. In case of providing full PVC spec with source independent of CDI, no CDI dependency will be required and this operation will not require CDI in the cluster. In case of partial spec we will use CDI API library to fill out these possible fields: VolumeMode, AccessModes, Size(in case of clone) and virtualization optimal storage class (planned work - [design doc PR](https://github.com/kubevirt/community/pull/233)).  
As described a new API will be created in the CDI API library. This API will receive a partial PVC template and will return a complete PVC spec. VolumeMode and AccessModes will be filled according to storageProfile matching the PVC storageClass. In case of empty storageClass it will rely on either the default StorageClass or if defined `virtualization optimal storage class`(WIP as mentioned). If the of the source of the PVC is VolumeCloneSource Size is also optional and can be filled automatically.  
The provided PVC template in the VolumeClaimTemplate should have a source that can either be handled by the new CDI volume populators API (DataSourceRef with CDI CR source), other costume populator source (DataSourceRef with costume CR source with appropriate controller available) or be left with no source resulting in empty disk.  
This field will complement the DataVolumeTemplates field. Instead of providing a template for a DataVolume which eventually creates one, we will provide a PVC template and create it. DataVolumes and DataVolumeTemplates should be deprecated eventually.  
Existing `vm create` virtctl command will be adjusted to create a VM with this new field.  

A complementary API which doesn't require VM is an optional virtctl command called `create-pvc`. Similarly to the VolumeClaimTemplate the user can provide minimal flags such as name and source and a PVC will be generated and created. Given all the required fields and either no source or costume DataSourceRef the command will not rely on CDI, otherwise CDI will be required and will use the CDI API library as described before.  

The road not taken:  
Another solution for an API that doesn't require neither VM nor virtctl command was suggested to be used as the underling API for PVC creation:  
Adding a PVC mutating webhook. This webhook happens after k8s mutating webhook and can complete all the mentioned fields same as described in the CDI API library changes. Both virtctl and the VM controller can simply apply the given partial PVC spec and it will be mutated in this webhook filled with our smart storage knowledge.  
This gives an advantage we can't receive only with using the CDI API library: the option of creating a PVC without Kubevirt present (either controllers nor virtctl).  
There are some complications with this idea:  
- General hook on PVCs - can be solved by setting a label on the PVC and use `ObjectSelector = MatchLabels` in the webhook, so only CDI-labeled PVCs are handled by the webhook (requires a user that wants to use the webhook to know and put this label on the PVC)  
- Default Filesystem volume mode - the webhook is done after k8s mutating webhook so VolumeMode will never be received empty, the webhook will have to know if Filesystem was the user requirement or we should override this variable if we think it should. Can be solved with some annotation the user adds to mention Filesystem is his choice - but its not so elegant and yet another requirement the user needs to know.  
- Downtimes - the need to make sure the webhook was applied - best case if the webhook was not applied the PVC creation fails since it might have missing required fields (example: AccessMode). Worst case PVC was created without the webhook applied and we might start populating a PVC without our desired changes (even if we can know that the PVC was not rendered by the webhook by some stamped annotation we will still need to delete it and create a new one)  
- K8s updates in their PVC mutating webhook that might impact ours. The webhook Will need to be adjusted according to k8s changes.  

After some deliberation taking into account the pros and cons I decided not to rely on the webhook from the application. I leave the option of implementing this webhook to be used locally only with applying a standalone PVC yaml.  

## API Examples
### VM with VolumeClaimTemplates, VolumeImportSource as volume source

This example will create a VolumeImportSource along side with a VM. Once the VM is created we will generate a PVC spec with the optimal VolumeMode and AccessModes. This PVC will be created and populated with the provided image in the VolumeImportSource url.
```yaml
apiVersion: cdi.kubevirt.io/v1beta1
kind: VolumeImportSource
metadata:
  name: my-import-source
spec:
  source:
      http:
         url: "https://download.cirros-cloud.net/0.4.0/cirros-0.4.0-x86_64-disk.img"
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  running: true
  volumeClaimTemplates:
  - metadata:
      name: cirros-pvc
    spec:
      dataSourceRef:
        apiGroup: cdi.kubevirt.io
        kind: VolumeImportSource
        name: my-import-source
      resources:
        requests:
          storage: 1Gi
  template:
    spec:
      domain:
        devices: {}
      volumes:
      - persistentVolumeClaim:
          name: cirros-pvc
        name: cirros-disk
```

#### Open question:
With dataVolumeTemplates one of the downsides was that we couldn't use generateName for the datavolume name.  
It will be cool if we could use it for the PVC name. Issue: we need to mention the PVC name in the vm.spec.template.volumes part when writing the VM spec. Any idea how to go past that? Update the PVC name there after the PVC is created sounds like not the best solution. WDYT?  

**Suggested solution:**  
PVC name in VolumeClaimTemplates will be unique in the VM spec and will be mentioned with the same name in the volumes list in the VM spec.  
When creating the PVC the VM name will be added as a prefix for the PVC name, in case the VM name was generated then we will get a unique generated PVC name, resulting in the possibility of applying the VM yaml with the VolumeClaimTemplates multiple times getting different unique VM and PVC names.  
For example:  
For VM with `generateName:my-vm`, the VM generated name will look something like `my-vm12345`. Having in VolumeClaimTemplates a PVC with `name:my-pvc` we will created a PVC with the name: `my-vm12345-my-pvc`.  
Regarding how will it look like in the VMI: before creating the VMI we will update the VMI volumes list to have the concatenated PVC name, similar to how we update the VMI name to equal to the generated VM name.  
Example to how it will all look like:  
VM yaml:  
```yaml
kind: VirtualMachine
metadata:
  generateName: testvm-
...
spec:
  volumeClaimTemplates:
  - metadata:
      name: pvc1
...
  template:
    spec:
      volumes:
      - persistentVolumeClaim
          name:  pvc1
        name: volume1
...
```

Created PVC name: testvm-qwert-pvc1  

Translates to VMI yaml:  
```yaml
kind: VirtualMachineInstance
metadata:
  Name: testvm-qwert
...
  spec:
    volumes:
    - persistentVolumeClaim
         name:  testvm-qwert-pvc1
       name: volume1
...
```

### Virtctl create-pvc command

Example for virtctl command uses:

Command to create a PVC with a generated name with "my-pvc-" prefix. The PVC size will be 1Gi, it will have empty disk image and AccessModes and VolumeMode optimal for the defined default storage class.
```bash
$ virtctl create-pvc --generate-name=my-pvc- --size=1Gi
```

Command to create a PVC named "my-pvc" with AccessModes and VolumeMode optimal for rook-ceph-block. The PVC will be populated by the data source provided in VolumeImportSource CR called "my-import-source".
```bash
$ virtctl create-pvc --name=my-pvc- --data-source-ref-kind=VolumeImportSource --data-source-ref-name=my-import-source --size=1Gi --storage-class=rook-ceph-block
```

Command to create a PVC my-pvc with DataSourceRef of VolumeCloneSource called my-source. CDI clone procedure doesn't require size, VolumeMode, AccessModes those will all be filled automatically.
```bash
$ virtctl create-pvc --name=my-pvc --data-source-ref-kind=VolumeCloneSource --data-source-ref-name=fedora
```

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
Currently we will still support dataVolumeTemplates and dataVolumes but we do plan to eventually deprecate them.  
We should update our docs and UI to use the new API and prevent from using or mentioning dataVolumes.

## Functional Testing Approach
Modify most tests that use dataVolumeTemplates to use volumeClaimTemplates.  
Make sure this change integrates well with hotplugs, snapshot and restore and other storage related features.  
Test in tier-2 the integration with GitOps, backup and restore and DR solutions.  

# Implementation Phases
Currently Planned work:
- Create a function in CDI API library which uses CDI existing code to generate an optimal PVC spec. The function will receive a partial PVC spec and will fill in (if missing) information such as VolumeMode and AccessModes according to [Storage Profile](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/storageprofile.md)
- Add Kubevirt new API of VolumeClaimTemplates which will be used to mention PVC source and basic information with end result of creating a PVC that will have the optimal properties and data.
- Add to Kubevirt `vm create` virtctl command the use of the new VolumeClaimTemplates
- Adjust InstanceTypes and Preferences with the new API.
- (optional) New virtctl command for creating PVCs.

The rest of the work related to DIC and UI will be planned later.  
Mentioning again: the related DIC work depends on [k8s cross namespace](https://kubernetes.io/blog/2023/01/02/cross-namespace-data-sources-alpha/) moving to beta and updating CDI clone populator accordingly.  
The change in API in VolumeClaimTemplates is that `Namespace` field will be able to be specified in the dataSourceRef.

