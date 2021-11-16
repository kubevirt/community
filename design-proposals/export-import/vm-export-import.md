# Overview
Proposal for exporting/importing Kubevirt Virtual Machine data from/to a Kubernetes cluster.

## Motivation
The Containerized Data Importer (CDI) project is mostly focused on getting Virtual Machine data into a cluster. The DataVolume API declares import and upload operations that populate and initialize individual Persistent Volume Claims (PVCs) with external data.  Of course that data is very likely to change once the PVCs are mounted and consumed by VirtualMachines (VMs).  But there is no standard way to get that valuable data out of the cluster.  Maybe an external backup would be useful for Disaster Recovery purposes?  Or for migrating VMs between clusters?

Once there is a standard API for exporting all persistent data for a VM, it follows that a new type of import operation is required, one that operates on a higher level than the existing CDI primitatives.  It should be able to parse an exported arhive and restore the Virtual Machine configuration and data.

## Goals
* Define Virtuam Machine Archive file format
* Define VirtualMachineExport API
* Introduce virtctl export|import vm commands

## Non Goals
* Implementation details of internal components (servers/proxies/controllers/etc)

## User Stories
* As a KubeVirt user, I would like to backup the current state of a Virtual Machine by exporting a Virtual Machine Archive
* As a KubeVirt user, I would like to backup the state of a Virtual Machine Snapshot by exporting a Virtual Machine Archive
* As a KubeVirt user, I would like to create a new Virtual Machine from a Virtual Machine Archive

## Repos
* [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

## Virtual Machine Archive File Format

A Vritual Machine Archive is a xz compressed tar archive with the following structure

```
.
├── manifest
├── resources
│   ├── pvc-block-pvc.yaml
│   ├── pvc-filesystem-pvc.yaml
│   ├── pvc-virtiofs-pvc.yaml
│   └── vm.yaml
└── volumes
    ├── block-pvc
    │   └── disk.img
    ├── filesystem-pvc
    │   └── disk.img
    └── virtiofs-pvc
        ├── file1
        ├── file2
        └── file3
```

### Manifest

The `manifest` file lists all files (excluding itself) in the archive and their SHA-1 hashes.  It can be used to ensure that corrumted data is not imported.

### Resources

The `resources` path contains copies of all the Kubernetes resources associaterd to the Virtual Machine.

### Volumes

The `volumes` path contains an entry for each exported volume.

## VirtualMachineExport API

The VirtualMachineExport API allows users to download a Virtual Machine Archive from a URL.

### User Flow

1.  Create a VirtualMachineExport resource
2.  Wait for the `Ready` condition of the VirtualMachineExport to be `True`
3.  Download the Virtual Machine Archive by getting the URL in `status.url`

### Export a VirtualMachine

``` yaml
apiVersion: export.kubevirt.io/v1alpha1
kind: VirtualMachineExport
metadata:
    name: export-vm1
spec:
    source:
        apiGroup: kubevirt.io
        kind: VirtualMachine
        name: vm1
```

### Export a VirtualMachine

``` yaml
apiVersion: export.kubevirt.io/v1alpha1
kind: VirtualMachineExport
metadata:
    name: export-snap1
spec:
    source:
        apiGroup: snapshot.kubevirt.io
        kind: VirtualMachineSnapshot
        name: snap1
```

### VirtualMachineExport Status

``` yaml
apiVersion: export.kubevirt.io/v1alpha1
kind: VirtualMachineExport
metadata:
    namespace: ns1
    name: export-vm1
spec:
    ...
status:
    phase: Ready # Pending|Ready|Terminated
    token: 8fmf94kf
    url: https://virt-import.kubevirt.my-cluster/ns1/export-vm1/download?token=8fmf94kf
    cert: <base64 encoded cert>
    clusterURL: https://virt-import.kubevirt/ns1/export-vm1/download?token=8fmf94kf
    clusterCert: <base64 encoded cert>
    conditions:
    - type: Ready
      status: True
      ...
      reason: Ready to export Virtual Machine
```

* Cluster must have Ingress/Route support for `url` and `cert` to be set

## virtctl export vm|vmsnapshot

virtctl will be extended to include an `export` command that does everything in the [User Flow](#user-flow) section with the addition of deleting the VirtualMachineExport when complete.

```
virtctl export vm vm1 vm1.tar.gz
```

```
virtctl export vmsnapshot snap1 snap1.tar.gz
```

## virtctl import vm

virtctl will be extended to include an `import` command that will create a new VirtualMachine based on the the contents of Virtual Machine Archive.

```
virtctl import vm import1 vm1.tar.gz
```

### Import Implementation

`virtctl import` will do the following on behalf of the user:

1. Validate target Virtual Machine name does not exist
2. Validate all Persistent Volume Claim (PVC) Storage Classes exist, if not require `--storage-class-overide` args for each
3. Generate DataVolumeTemplate with unique name for each PVC with [upload](https://github.com/kubevirt/containerized-data-importer-api/blob/main/pkg/apis/core/v1beta1/types.go#L139-L140) source
4. Create new VirtualMachine resource with `spec.dataVolumeTemplates` populated from DataVolumeTemplates above
5. Upload volume data from archive to target DataVolumes

## Scalability

Network bandwith is expected to be the biggest bottleneck for both export and import.  Given the current design, this is mostly and issue for export.  If there is a network disconnect while downloading an archive, the user will have to start again from the beginning.

When importing, it is possible to upload to multiple target volumes in parallel and maximize throughput.

# Implementation Phases

1.  VirtualMachineExport API and all supporting infrastructure to support VirtualMachineSnapshot source
2.  Import Virtual Machine Archive
3.  Extend VirtualMachineExport API to support VirtualMachine source
