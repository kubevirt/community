# Overview

[Instancetypes and preferences](https://kubevirt.io/user-guide/virtual_machines/instancetypes/) allow a users to pick a set of predefined resource, performance and other runtime preferences for their `VirtualMachine` without having to worry about the complexity of KubeVirt's API.

This design document aims to outline a simple way this feature could be further extended by allowing users to opt into having a default instance type and/or preference inferred from a given `Volume` associated with their `VirtualMachine`.

## Motivation

The inference of a default instance type and preference from a `Volume` associated with the `VirtualMachine` further reduces the choices a user has to make in order to get a running `VirtualMachineInstance` to a singular choice about which runnable image they would like to boot.

## Goals

* Allow users to opt into having a default instance type and preference inferred from a given `Volume` associated with their `VirtualMachine`.

## Non Goals

* Automatic annotation of a `Volume` or underlying type (such as a `PVC`) is not covered as part of this work but is something that will be investigated in the future.

## User Stories

* As an admin, user or third party creator of a `Volume`,  I want to provide annotations that inform the user of the default instance type or set of preferences a `VirtualMachine` should use when booting from the `Volume`.

* As a user, I want to opt into allowing KubeVirt to infer these defaults from a specific `Volume` associated with my `VirtualMachine`.

## Repos

* [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design

The `InstancetypeMatcher` and `PreferenceMatcher` types will be extended to include a new `inferFromVolume` attribute that will control this behaviour.

When provided if the type of `Volume` referred to by this attribute is supported it will be checked for the following annotations:

* `instancetype.kubevirt.io/defaultInstancetype`
* `instancetype.kubevirt.io/defaultInstancetypeKind` (Defaults to `VirtualMachineClusterInstancetype`)
* `instancetype.kubevirt.io/defaultPreference`
* `instancetype.kubevirt.io/defaultPreferenceKind` (Defaults to `VirtualMachineClusterPreference`)

If found these annotations will be used to populate the `InstancetypeMatcher` and `PreferenceMatcher` of the `VirtualMachine` by the mutation webhook with the existing checks of the validation webhook then used to ensure these resources exist and apply without conflicts to the `VirtualMachine` and eventual `VirtualMachineInstance` at runtime. The `inferFromVolume` attribute will be removed as part of this process as this lookup is only preformed once during the initial creation of a `VirtualMachine`.

Initial support will be introduced for the following `Volume` types:

* `PVC`
* `DataVolume`
  * Pre-created with a `PVC` or `SourceRef` source
  * As a `DataVolumeTemplate` on the `VirtualMachine` again with a `PVC` or `SourceRef` source
## API Examples

The following examples are made possible by the following prerequisite commands being run:

```yaml
$ wget https://github.com/cirros-dev/cirros/releases/download/0.5.2/cirros-0.5.2-x86_64-disk.img
[..]
$ ./cluster-up/virtctl.sh image-upload pvc cirros --size=1Gi --image-path=./cirros-0.5.2-x86_64-disk.img
[..]
$ ./cluster-up/kubectl.sh kustomize https://github.com/kubevirt/common-instancetypes.git | ./cluster-up/kubectl.sh apply -f -
[..]
$ ./cluster-up/kubectl.sh annotate pvc/cirros instancetype.kubevirt.io/defaultInstancetype=server.tiny instancetype.kubevirt.io/defaultPreference=cirros
```

### Instance type and preference inference directly from an existing PVC

```yaml
$ cat <<EOF | ./cluster-up/kubectl.sh apply -f -
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    inferFromVolume: cirros-disk
  preference:
    inferFromVolume: cirros-disk
  running: true
  template:
    spec:
      domain:
        devices: {}
      volumes:
      - persistentVolumeClaim:
          claimName: cirros
        name: cirros-disk
EOF
[..]
$ ./cluster-up/kubectl.sh get vms/cirros -o json | jq '.spec.instancetype, .spec.preference'
selecting docker as container runtime
{
  "kind": "virtualmachineclusterinstancetype",
  "name": "server.tiny",
  "revisionName": "cirros-server.tiny-ef0cbfb6-b48c-4e9f-aa7a-a06878b42503-1"
}
{
  "kind": "virtualmachineclusterpreference",
  "name": "cirros",
  "revisionName": "cirros-cirros-5bddae5d-47f8-433b-afa2-d4f846ef1830-1"
}
```

### Instance type and preference inference indirectly through DataVolume and PVC

```yaml
$ cat <<EOF | ./cluster-up/kubectl.sh apply -f -
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataVolume
metadata:
  name: cirros-dv
spec:
  pvc:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: rook-ceph-block
  source:
    pvc:
      name: cirros
      namespace: default
EOF
[..]
$ cat <<EOF | ./cluster-up/kubectl.sh apply -f -
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    inferFromVolume: cirros-disk
  preference:
    inferFromVolume: cirros-disk
  running: true
  template:
    spec:
      domain:
        devices: {}
      volumes:
      - dataVolume:
          name: cirros-dv
        name: cirros-disk
EOF
$ ./cluster-up/kubectl.sh get vms/cirros -o json | jq '.spec.instancetype, .spec.preference'
selecting docker as container runtime
{
  "kind": "virtualmachineclusterinstancetype",
  "name": "server.tiny",
  "revisionName": "cirros-server.tiny-ef0cbfb6-b48c-4e9f-aa7a-a06878b42503-1"
}
{
  "kind": "virtualmachineclusterpreference",
  "name": "cirros",
  "revisionName": "cirros-cirros-5bddae5d-47f8-433b-afa2-d4f846ef1830-1"
}
```

### Instance type and preference inference indirectly through DataVolume, DataVolumeTemplate and PVC

```yaml
$ cat <<EOF | ./cluster-up/kubectl.sh apply -f -
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    inferFromVolume: cirros-disk
  preference:
    inferFromVolume: cirros-disk
  running: true
  dataVolumeTemplates:
    - metadata:
        name: cirros-dv
      spec:
        pvc:
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
          storageClassName: rook-ceph-block
        source:
          pvc:
            name: cirros
            namespace: default
  template:
    spec:
      domain:
        devices: {}
      volumes:
      - dataVolume:
          name: cirros-dv
        name: cirros-disk
EOF
[..]
$ ./cluster-up/kubectl.sh get vms/cirros -o json | jq '.spec.instancetype, .spec.preference'
selecting docker as container runtime
{
  "kind": "virtualmachineclusterinstancetype",
  "name": "server.tiny",
  "revisionName": "cirros-server.tiny-ef0cbfb6-b48c-4e9f-aa7a-a06878b42503-1"
}
{
  "kind": "virtualmachineclusterpreference",
  "name": "cirros",
  "revisionName": "cirros-cirros-5bddae5d-47f8-433b-afa2-d4f846ef1830-1"
}

```

### Instance type and preference inference indirectly through DataVolume, DataVolumeTemplate, DataSource and PVC

```yaml
$ cat <<EOF | ./cluster-up/kubectl.sh apply -f -
---
apiVersion: cdi.kubevirt.io/v1beta1
kind: DataSource
metadata:
  name: cirros
spec:
  source:
    pvc:
      name: cirros
      namespace: default
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    inferFromVolume: cirros-disk
  preference:
    inferFromVolume: cirros-disk
  running: false
  dataVolumeTemplates:
    - metadata:
        name: cirros-dv
      spec:
        pvc:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
          storageClassName: local
        sourceRef:
          kind: DataSource
          name: cirros
          namespace: default
  template:
    spec:
      domain:
        devices: {}
      volumes:
        - dataVolume:
            name: cirros-dv
          name: cirros-disk
EOF
[..]
./cluster-up/kubectl.sh get vms/cirros -o json | jq '.spec.instancetype, .spec.preference'
selecting docker as container runtime
{
  "kind": "virtualmachineclusterinstancetype",
  "name": "server.tiny",
  "revisionName": "cirros-server.tiny-76454433-3d82-43df-a7e5-586e48c71f68-1"
}
{
  "kind": "virtualmachineclusterpreference",
  "name": "cirros",
  "revisionName": "cirros-cirros-85823ddc-9e8c-4d23-a94c-143571b5489c-1"
}
```

