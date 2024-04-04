# Overview

At present users of the `instancetype.kubevirt.io/v1beta1` API and CRDs are able to define a CPU topology for a `VirtualMachine` through a combination of vCPUs provided by the `guest` attribute of an instance type and an optional topology provided by the `preferredCPUTopology` attribute of a preference.

For example, :

```yaml
$ ./cluster-up/kubectl.sh apply -f - <<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: instancetype-example
spec:
  cpu:
    guest: 2
  memory:
    guest: 128Mi
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachinePreference
metadata:
  name: preference-example
spec:
  cpu:
    preferredCPUTopology: preferCores
EOF

$ ./cluster-up/virtctl.sh create vm \
  --instancetype virtualmachineinstancetype/instancetype-example \
  --preference virtualmachinepreference/preference-example \
  --volume-containerdisk src:registry:5000/kubevirt/alpine-container-disk-demo:devel,name:alpine-disk \
  --name example | ./cluster-up/kubectl.sh apply -f -

$ ./cluster-up/kubectl.sh get vmis/example -o json | jq .spec.domain.cpu
{
  "cores": 2,
  "model": "host-model",
  "sockets": 1,
  "threads": 1
}
```

This design proposal aims to extend the API to allow for an optional guest CPU topology to be provided by a referenced instance type instead of a generic number of vCPUs.

## Motivation

The initial design of instance types and preferences was heavily influenced by hyper scalers and other large scale IaaS projects. Such environments allow users to select a generic amount of vCPUs without specifying exactly how these are then mapped to the underlying guest. While a useful starting point KubeVirt is now attracting traditional virtualization users and use cases that demand both certainty around how these vCPUs are mapped to the guest while also demanding specific topologies not easily expressed by our current API.

## Goals

* Allow for a complete guest CPU topology to be defined by an instance type
* This should be optional and mutually exclusive with the original vCPU API

## User Stories

* As an Operator I would like to provide instance types with specific guest CPU topologies for users to create VirtualMachines
* As a VM owner I would like to use an instance type with a specific guest CPU topology when creating a VirtualMachine

## Repos

* kubevirt/kubevirt

# Design

## `spec.cpu.guest`

`spec.cpu.guest` will be marked as optional, this change will require a new `v1beta2` version of the `instancetype.kubevirt.io` API group to be created given the change in requirement of this field.

## `spec.cpu.topology`

A new optional `spec.cpu.topology` attribute will be introduced using the `CPUTopology` type defined within the `kubevirt.io/api/v1` API.

## virt-api

The validation webhook will require `spec.cpu.guest` or `spec.cpu.topology` to be defined by an instance type.

## pkg/instancetype

Application of the provided topology will occur at the same time as the original `guest` value prior to the submission of the `VirtualMachineInstance`.

## API Examples

```yaml
$ ./cluster-up/kubectl.sh apply -f - <<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: instancetype-example
spec:
  cpu:
    topology:
      sockets: 3
      cores: 2
      threads: 1
  memory:
    guest: 128Mi
EOF

$ ./cluster-up/virtctl.sh create vm \
  --instancetype virtualmachineinstancetype/instancetype-example \
  --volume-containerdisk src:registry:5000/kubevirt/alpine-container-disk-demo:devel,name:alpine-disk \
  --name example | ./cluster-up/kubectl.sh apply -f -

$ ./cluster-up/kubectl.sh get vmis/example -o json | jq .spec.domain.cpu
{
  "cores": 2,
  "model": "host-model",
  "sockets": 3,
  "threads": 1
}

```

## Functional Testing Approach

Existing functional tests should be extended to validate the use of the new `topology` attribute.

# Implementation Phases

* Introduce `instancetype.kubevirt.io/v1beta2`
* Mark `spec.cpu.guest` as optional
* Introduce `spec.cpu.topology`
