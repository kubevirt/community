# Overview
This proposal is about adding supporting DRA (dynamic resource allocation) in KubeVirt.
DRA allows vendors fine-grained control of devices.  The device-plugin model will continue
to exist in Kubernetes, but DRA will offer vendors more control over the device topology.

## Motivation
DRA adoption is important for KubeVirt so that vendors can expect the same
control of their devices using Virtual Machines and Containers.

## Goals
- Align on how KubeVirt will consume external and in-tree DRA drivers
- Align on what drivers KubeVirt will support in tree

## Non Goals
- Replace device-plugin support in KubeVirt

## Definition of Users
A user is a person that wants to attach a device to a VM

## User Stories
- As a user, I would like to use my GPU dra driver with KubeVirt
- As a user, I would like to use KubeVirt's default driver

## Repos
kubevirt/kubevirt

# Design

## API Examples

### VM API with PassThrough GPU

```
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
name: gpu.resource.nvidia.com
driverName: gpu.resource.nvidia.com
---
apiVersion: rtx4090.gpu.resource.nvidia.com/v1
kind: ClaimParameters
name: rtx4090-claim-parameters
spec:
  driver: vfio
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
  name: rtx4090-claim-template
spec:
  spec:
    resourceClassName: gpu.resource.nvidia.com
    parametersRef:
      apiGroup: rtx4090.gpu.resource.nvidia.com/v1
      kind: ClaimParameters
      name: rtx4090-claim-parameters
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-cirros
    spec:
      resourceClaims:
      - name: rtx4090
        source:
          resourceClaimTemplateName: rtx4090-claim-template
–--
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-cirros
spec:
  containers:
  - name: virt-launcher
    image: virt-launcher
    resources:
      claims:
      - name: rtx4090
  resourceClaims:
  - name: rtx4090
    source:
      resourceClaimTemplateName: rtx4090-claim-template
```

### VM API with vGPU

```
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
name: gpu.resource.nvidia.com
driverName: gpu.resource.nvidia.com
---
apiVersion: a100.gpu.resource.nvidia.com/v1
kind: ClaimParameters
name: a100-40C-claim-parameters
spec:
  driver: nvidia
  profile: A100DX-40C # Maximum 2 40C vGPUs per GPU
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
  name: a100-40C-claim-template
spec:
  spec:
    resourceClassName: gpu.resource.nvidia.com
    parametersRef:
      apiGroup: a100.gpu.resource.nvidia.com/v1
      kind: ClaimParameters
      name: a100-40C-claim-parameters
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-cirros
    spec:
      resourceClaims:
      - name: a100-40C
        source:
          resourceClaimTemplateName: a100-40C-claim-template
–--
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-cirros
spec:
  containers:
  - name: virt-launcher
    image: virt-launcher
    resources:
      claims:
      - name: a100-40C
  resourceClaims:
  - name: a100-40C
    source:
      resourceClaimTemplateName: a100-40C-claim-template
```

# References

- Structured parameters
https://github.com/kubernetes/kubernetes/pull/123516
- Structured parameters KEP
https://github.com/kubernetes/enhancements/issues/4381
- DRA
https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/
