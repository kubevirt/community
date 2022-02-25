# Overview
Proposal for exporting KubeVirt Virtual Machine volumes from a Kubernetes cluster

## Motivation
There are currently multiple ways to import/initialize Virtual Machine volumes.  But there is no easy and efficient method to get volume data out of the cluster.

## Goals
* Define VirtualMachineExport API
* Introduce virtctl vmexport command

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

In the future, the VirtualMachineExport type may be extended to include other artifacts like VirtualMachine yamls or a single .ova like file containing all persistent data and manifests for a single VM.

### User Flow

1.  Create a VirtualMachineExport resource
2.  Wait for the `Ready` condition of the VirtualMachineExport to be `True`
3.  Download a Virtual Machine volume in the desired format from a URL listed under `status.links.volumes`

### Export Formats

KubeVirt supports `kubevirt` and `archive` content types.  They are defined [here](https://github.com/kubevirt/containerized-data-importer-api/blob/v1.45.0/pkg/apis/core/v1beta1/types.go#L105-L113):

| Volume Content Type | Export Format |
| ------------------- | ------------- |
| kubevirt            | raw           |
| kubevirt            | gzip          |
| archive             | dir           |
| archive             | tar.gz        |

#### raw

The http endpoint supports random access to a raw disk image.

#### gzip

The http endpoint serves a `gzip` compressed raw disk image (no range support).

#### dir

The http endpoint is a golang [http.FileServer](https://pkg.go.dev/net/http#FileServer) serving up the root of a FileSystem PersistentVolumeClaim.  Getting the root URL will return a list of files/directories on the root.

#### tar.gz

The http endpoint serves a `tar.gz` streaming download (no range support).

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

### Export a PersistentVolumeClaim

``` yaml
apiVersion: export.kubevirt.io/v1alpha1
kind: VirtualMachineExport
metadata:
    name: export-snap1
spec:
    source:
        apiGroup: v1
        kind: PersistentVolumeClaim
        name: pvc1
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
            url: https://virt-export.kubevirt/ns1/export-vm1/volumes/kubevirt-volume/disk.img?token=8fmf94kf
          - format: gzip
            url: https://virt-export.kubevirt/ns1/export-vm1/volumes/kubevirt-volume/disk.img.gz?token=8fmf94kf
        - name: archive-volume
          - format: dir
            url: https://virt-export.kubevirt/ns1/export-vm1/volumes/archive-volume/dir?token=8fmf94kf
          - format: tar.gz
            url: https://virt-export.kubevirt/ns1/export-vm1/volumes/archive-volume/disk.tar.gz?token=8fmf94kf
      external:
        cert: <base64 encoded cert>
        volumes:
        - name: kubevirt-volume
          - format: raw
            url: https://virt-export.kubevirt.my-cluster/ns1/export-vm1/volumes/kubevirt-volume/disk.img?token=8fmf94kf
          - format: gzip
            url: https://virt-export.kubevirt.my-cluster/ns1/export-vm1/volumes/kubevirt-volume/disk.img.gz?token=8fmf94kf
        - name: archive-volume
          - format: dir
            url: https://virt-export.kubevirt.my-cluster/ns1/export-vm1/volumes/archive-volume/dir?token=8fmf94kf
          - format: tar.gz
            url: https://virt-export.kubevirt.my-cluster/ns1/export-vm1/volumes/archive-volume/disk.tar.gz?token=8fmf94kf
    conditions:
    - type: Ready
      status: True
      ...
      reason: Ready to export Virtual Machine
```

* Cluster must have Ingress/Route support for `external` urls and certificate to be set

## virtctl vmexport

virtctl will be extended to include an `vmexport` command that will create/delete a `VirtualMachineExport`

```
virtctl vmexport create --vm=vm1 vm1-export
```

```
virtctl vmexport create --snapshot=snap1 snap1-export
```

```
virtctl vmexport delete snap1-export
```

Once a `VirtualMachineExport` is created, `virtctl vmexport` can download volume archives

```
virtctl vmexport download vm1-export --volume=volume1 --output disk.img.gz
```

## Volume Migration Between Clusters

### Import

If cluster1 is accessible from cluster2

```bash
# set kubeconfig for cluster1
kubectl config set-cluster cluster1

# create VirtualMachineExport
virtctl vmexport create -n namespace1 --vm=vm1 vm1-export

# wait for VirtualMachineExport to be ready
kubectl wait vmexport -n namespace1 vm1-export --for condition=Ready

# get import URL
IMPORT_URL=$(kubectl get -n namespace1 vm1-export -o="jsonpath={.status.links.external.volumes ... .url}")

# set kubeconfig for cluster2
kubectl config set-cluster cluster2

cat <<EOF | kubectl apply -f -
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: migrated-volume
  namespace: namespace2
spec:
  source:
      http:
         url: "${IMPORT_URL}"
  pvc:
    storageClassName: rook-ceph-block
    volumeMode: Filesystem
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
EOF

# wait for import to complete
kubectl wait dv -n namespace2 migrated-volume --for condition=Ready
```

### Upload

If cluster1 is NOT accessible from cluster2

```bash
# set kubeconfig for cluster1
kubectl config set-cluster cluster1

# create VirtualMachineExport
virtctl vmexport create -n namespace1 --vm=vm1 vm1-export

# wait for VirtualMachineExport to be ready
kubectl wait vmexport -n namespace1 vm1-export --for condition=Ready

# download disk image
virtctl vmexport download -n namespace1 vm1-export --volume=volume1 --output disk.img.gz

# set kubeconfig for cluster2
kubectl config set-cluster cluster2

# upload to target
virtctl image-upload dv -n namespace2 migrated-volume  --size=1G --image-path=disk.img.gz
```

## Scalability

Network bandwith is expected to be the biggest bottleneck for export.  For this reason, traffic will not be going through `kube-apiserver`.  Rather, traffic will be routed through a proxy that is exported via `Ingress` or a `Route` in OpenShift.  If there is a network disconnect while downloading an archive, the user will have to start again from the beginning.  But if the `raw` or `dir` endpoints are used, a client will be able to continue from the last byte received.

# Implementation Phases

1.  VirtualMachineExport API and all supporting infrastructure to support VirtualMachineSnapshot source
2.  Extend VirtualMachineExport API to support VirtualMachine source
3.  Extend VirtualMachineExport API to support PersistentVolumeClaim source
