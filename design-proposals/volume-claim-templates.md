# Overview
This proposal is about having a better integration with kubernetes storage APIs.  
With recent support of kubernetes [volume populators](https://kubernetes.io/blog/2022/05/16/volume-populators-beta/) we have implemented [CDI volume populators](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/cdi-populators.md) which allows population of PVC the kubernetes way without the need for DataVolumes.  
Along major benefits of using this populators we loss one of DataVolumes greatest achievements - generating a PVC with the optimal properties fitting the user's storage.  
This design should help us gain the best of all worlds.  

## Motivation
Harnessing k8s and CDI new volume populators API along with our knowledge of storage capabilities will allow the user to enjoy synchronization with k8s and better compatibility with backup and restore solutions as well as the smart creation of PVCs that best suits the user's storage.

## Goals
- Provide a method for creating VMs with PVCs that suits best the user's storage while using CDI volume populators.

## Non Goals

## User Stories
* As a user creating VMs, I want to be able to create a PVC within the VM spec while getting the best PVC spec that is also compatible with backup and restore solutions.

## Repos
* **CDI**: Provide API to that receives a PVC spec and in case of missing properties such as VolumeMode and AccessModes fills the optimal values according to the storage class.
* **CDI**: Update DataImportCron API to create/update VolumeCloneSources that can be consumed as sources for PVC handled by CDI volume populators. VolumeCloneSource will need to be able to be updated with different PVC source.
* **Kubevirt**: Provide API to create a VM with minimal volume information and receive a VM with best suited populated PVC.
* **Kubevirt**: Provide virtctl command to generate and create PVCs.
* **SSP**: Manage setting up VolumeCloneSources for commonly supported operating systems updated by DataImportCrons.
* **Common Templates**: Transition to use VolumeClaimTemplates in VM specs and use CDI's VolumeCloneSource as the source in the template.
* **OCP UI**: Provide API to create a CDI Volume populator CR that can be used by the volumeClaim. In case of golden image source need to ensure VolumeCloneSource is linked properly to a PVC in order to determine whether the boot image for a common template is available.

# Design

A new field will be added to the VM object called **VolumeClaimTemplates**. This field will provide the user a way to define and create a PVC owned by the VM while providing partial/full PVC spec embedded in the VM spec. In case of partial spec we will make sure the preferred VolumeMode and AccessModes combination according to the StorageProfile associated with provided/default StorageClass are used.  
This field should replace the DataVolumeTemplates field. Instead of providing a template for a DataVolume which eventually creates one, we will provide a PVC template and create it. DataVolumes and DataVolumeTemplates should be deprecated eventually.  
A new API will be created in the CDI API library which will receive a partial PVC template and using existing CDI logic will return a complete PVC spec.  
The provided PVC template in the VolumeClaimTemplate should have a source that can be handled by the new existing CDI volume populators API (DataSourceRef with CDI CR source) or a standalone k8s way (user costume populator, empty disk etc..)

A virtctl command called `create-pvc` can be added. Given minimal flags such as name, size and source this command can generate and create a PVC. The command will use the CDI API similarly as will be done in the VM controller.

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
With datavolumestemplate one of the downsides was that we couldn't use generateName of the datavolume name.
It will be cool if we could use it for the PVC name. Issue: we need to mention the PVC name in the vm.spec.template.volumes part when writing the VM spec. Any idea how to go past that? update the PVC name there after the PVC is created sounds like not the best solution. WDYT?

### Virtctl create-pvc command

Example for virtctl command uses:
```bash
$ virtctl create-pvc --generate-name=my-pvc- --size=1Gi
```
This command will create a PVC with a generated name with "my-pvc-" prefix. The PVC size will be 1Gi, it will have empty disk image and AccessModes and VolumeMode optimal for the defined default storage class.

```bash
$ virtctl create-pvc --name=my-pvc- --volume-import-source=my-import-source --size=1Gi --storage-class=rook-ceph-block
```
This command will create a PVC named "my-pvc" with AccessModes and VolumeMode optimal for rook-ceph-block. The PVC will be populated by the data source provided in VolumeImportSource CR called "my-import-source".

```bash
$ virtctl create-pvc --name=my-pvc --volume-clone-source=fedora
```
This will create a PVC my-pvc with DataSourceRef of VolumeCloneSource called my-source. CDI clone procedure doesn't require size, VolumeMode, AccessModes those will all be filled automatically.

## Scalability
I don't see any scalability issues.

## Update/Rollback Compatibility
Currently we will still support dataVolumeTemplates and dataVolumes but we do plan to eventually deprecate them.  
We should update our docs and UI to use the new API and prevent from using or mentioning dataVolumes.

## Functional Testing Approach
Modify most tests that use dataVolumeTemplates to use volumeClaimTemplates.  
Make sure this change integrates well with hotplugs, snapshot and restore and other storage related features.

# Implementation Phases
Currently Planned work:
- Create a function in CDI API library which uses CDI existing code to generate an optimal PVC spec. The function will receive a partial PVC spec and will fill in (if missing) information such as VolumeMode and AccessModes according to [Storage Profile](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/storageprofile.md)
- Add Kubevirt new API of VolumeClaimTemplates which will be used to mention PVC source and basic information with end result of creating a PVC that will have the optimal properties and data.
- New virtctl command for creating PVCs.

The rest of the work related to DIC and UI will be planned later

