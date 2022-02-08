# Overview
Proposal for exporting KubeVirt Virtual Machine volumes from a Kubernetes cluster

## Motivation
There are currently multiple ways to import/initialize Virtual Machine volumes.  But there is no easy and efficient method to get volume data out of the cluster.

## Goals
* Define VirtualMachineExport API
* Introduce virtctl export command

## Non Goals
* Implementation details of internal components (servers/proxies/controllers/etc)

## User Stories
* As a KubeVirt user, I would like to export a Virtual Machine volume for local debugging/analysis.
* As a KubeVirt user, I would like to export a Virtual Machine volume as part of migrating a Virtual Machine to a different cluster.
* As a KubeVirt user, I would like to export a Virtual Machine volume to serve as an offsite backup.

## Repos
* [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

## VirtualMachineExport API

The VirtualMachineExport API allows users to download a Virtual Machine volumes from URLs.

### User Flow

1.  Create a VirtualMachineExport resource
2.  Wait for the `Ready` condition of the VirtualMachineExport to be `True`
3.  Download a Virtual Machine volume in the desired format listed under `status.links.volumes`

### Export Formats

| Volume Content Type | Export Format |
| ------------------- | ------------- |
| kubevirt            | raw           |
| kubevirt            | tar.gz        |
| archive             | dir           |
| archive             | tar.gz        |

#### raw

The http endpoint supports random access to a raw disk image.

#### tar.gz

The http endpoint is a link to a `tar.gz` streaming download.

#### dir

The http endpoint is a golang `http.FileServer` serving up the root of the filesystem.

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

### Export a VirtualMachineSnapshot

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
    links:
      internal:
        cert: <base64 encoded cert>
        volumes:
        - name: kubevirt-volume
          - format: raw
            url: https://virt-import.kubevirt/ns1/export-vm1/volumes/kubevirt-volume/raw?token=8fmf94kf
          - format: archive
            url: https://virt-import.kubevirt/ns1/export-vm1/volumes/kubevirt-volume/archive?token=8fmf94kf
        - name: archive-volume
          - format: dir
            url: https://virt-import.kubevirt/ns1/export-vm1/volumes/archive-volume/dir?token=8fmf94kf
          - format: archive
            url: https://virt-import.kubevirt/ns1/export-vm1/volumes/archive-volume/archive?token=8fmf94kf
      external:
        cert: <base64 encoded cert>
        volumes:
        - name: kubevirt-volume
          - format: raw
            url: https://virt-import.kubevirt.my-cluster/ns1/export-vm1/volumes/kubevirt-volume/raw?token=8fmf94kf
          - format: archive
            url: https://virt-import.kubevirt.my-cluster/ns1/export-vm1/volumes/kubevirt-volume/archive?token=8fmf94kf
        - name: archive-volume
          - format: dir
            url: https://virt-import.kubevirt.my-cluster/ns1/export-vm1/volumes/archive-volume/dir?token=8fmf94kf
          - format: archive
            url: https://virt-import.kubevirt.my-cluster/ns1/export-vm1/volumes/archive-volume/archive?token=8fmf94kf
    conditions:
    - type: Ready
      status: True
      ...
      reason: Ready to export Virtual Machine
```

* Cluster must have Ingress/Route support for `external` urls and certificate to be set

## virtctl export

virtctl will be extended to include an `export` command that will create/delete a `VirtualMachineExport`

```
virtctl export create --vm=vm1 vm1-export
```

```
virtctl export create --snapshot=snap1 snap1-export
```

```
virtctl export delete snap1-export
```

Once a `VirtualMachineExport` is created, `virtctl export` can download volume archives

```
virtctl export download vm1-export --volume=volume1 --output volume1.tar.gz
```

## Scalability

Network bandwith is expected to be the biggest bottleneck for export.  For this reason, traffic will not be going through `kube-apiserver`.  Rather, traffic will be routed through a proxy that is exported via `Ingres` or a `Route` in OpenShift.  If there is a network disconnect while downloading an archive, the user will have to start again from the beginning.  But if the `raw` or `dir` endpoints are used, a client will be able to continue from the last byte received.

# Implementation Phases

1.  VirtualMachineExport API and all supporting infrastructure to support VirtualMachineSnapshot source
2.  Extend VirtualMachineExport API to support VirtualMachine source
