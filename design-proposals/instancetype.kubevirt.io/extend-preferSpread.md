# Overview

`preferredCPUTopology: preferSpread` was introduced to the preference CRD and API with KubeVirt v1.1. It allows for vCPUs provided by an instance type to be spread across guest visible sockets and cores using a configurable ratio that defaults to 2.

For example:

```yaml

$ ./cluster-up/kubectl.sh apply -f -<<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: spread
spec:
  cpu:
    guest: 4
  memory:
    guest: 1Gi
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachinePreference
metadata:
  name: spread
spec:
  cpu:
    preferredCPUTopology: preferSpread
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-spread
spec:
  instancetype:
    name: spread
    kind: VirtualMachineInstancetype
  preference:
    name: spread
    kind: VirtualMachinePreference
  runStrategy: Always
  template:
    spec:
      domain:
        devices: {}
      volumes:
        - containerDisk:
            image: quay.io/containerdisks/fedora:39
          name: containerdisk
EOF
[..]
$ ./cluster-up/kubectl.sh get vmi/vm-spread -o json | jq .spec.domain.cpu
{
  "cores": 2,
  "model": "host-model",
  "sockets": 2,
  "threads": 1
}
```

This design proposal aims to extend this implementation by introducing new configurables within a preference to control how vCPUs are spread with the goal of allowing vCPUs to also be spread to threads.

## Motivation

The current implementation of the `preferSpread` `preferredCPUTopology` is limited to spreading vCPUs across sockets and cores of the guest CPU topology. While a useful starting point it would be beneficial to also allow these vCPUs to be spread across all parts of the guest visible CPU topology including threads in order to mimic SMT within the guest.

This is particularly useful for performance related workloads such as DPDK etc where we would like to make use of SMT on the host in order to map through sibling threads from pCPUs to guest thread siblings. Note that this use case requires `dedicatedCPUPlacement` and `guestMappingPassthrough` to be useful so the CPUManager can allocate complete pCPUs (via the `full-pcpus-only` option) and for these pCPU siblings to be mapped correctly as vCPU thread siblings to the guest.

## Goals

* Allow for vCPUs to be spread across all attributes of the guest visible CPU topology

## User Stories

* As an admin or VM owner I would like to provide preferences that spread vCPUs provided by an instance type over the guest visible sockets, cores and threads

* As a VM owner I would like to consume preferences that that spread vCPUs provided by an instance type over the guest visible sockets, cores and threads in order to mimic SMT within the guest OS

## Repos

* kubevirt/kubevirt

* kubevirt/common-instancetypes

# Design

### `spec.cpu.spreadOptions`

A new `spreadOptions` struct will be introduced to hold new configurables detailing how vCPUs should be spread over the guest visible CPU topology.

```go
type spreadAcross string

type SpreadOptions struct {
	// Across optionally defines how to spread vCPUs across the guest visible topology.
	// Default: SocketsCores
	//
	//+optional
	Across *spreadAcross `json:"across,omitempty"`

	// Ratio optionally defines the ratio to spread vCPUs across the guest visible topology.
	// Default: 2
	//
	//+optional
	Ratio *uint32 `json:"ratio,omitempty"`
}
```

The `ratio` option is a replacement for the original `spec.preferSpreadSocketToCoreRatio` configurable. The original being deprecated and discussed later in this proposal.

A new `across` option will be introduced that will map to the following constants and behaviour.

### `spreadAcross` `SocketsCores` (default)

This option will spread vCPUs between sockets and cores. Given the current behaviour of `PreferSpread` this will remain the default.

```go
const SpreadAcrossSocketsCores SpreadAcross = "SocketsCores"
```

### `spreadAcross` `CoresThreads`

This option will spread vCPUs between cores and threads with a max of 2 threads per core. This essentially ignores the ratio and always applies a ratio of 2.

```go
const SpreadAcrossCoresThreads SpreadAcross = "CoresThreads"
```

### `spreadAcross` `SocketsCoresThreads`

This option will spread vCPUs between all elements of the CPU topology with a max of 2 threads per core.

```go
const spreadAcrossSocketsCoresThreads SpreadAcross = "SocketsCoresThreads"
```

### `spec.preferSpreadSocketToCoreRatio` deprecation

This original configurable for the `preferSpread` `preferredCPUTopology` was added directly to the core preference `spec`.

This was a mistake and given the new nested `spec.cpu.spreadOptions` struct we can now deprecate the original field and move users over to `spec.cpu.spreadOptions.ratio`.

### Provide shorter preferredCPUTopology constants (Optional)

While working on this area it would be useful to also cleanup the API a little with shorter `preferredCPUTopology` constants.

`PreferSockets` -> `Sockets`
`PreferCores`   -> `Cores`
`PreferThreads` -> `Threads`
`PreferSpread`  -> `Spread`

This will allow the originals to be deprecated for removal in a future `v1beta2` or eventual `v1` release of the API.

## API Examples

### `SocketsCores` (default)

```yaml
$ ./cluster-up/kubectl.sh apply -f -<<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: spread
spec:
  cpu:
    guest: 8
  memory:
    guest: 1Gi
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachinePreference
metadata:
  name: spread
spec:
  cpu:
    preferredCPUTopology: Spread
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-spread
spec:
  instancetype:
    name: spread
    kind: VirtualMachineInstancetype
  preference:
    name: spread
    kind: VirtualMachinePreference
  runStrategy: Always
  template:
    spec:
      domain:
        devices: {}
      volumes:
        - containerDisk:
            image: quay.io/containerdisks/fedora:39
          name: containerdisk
EOF
[..]
$ ./cluster-up/kubectl.sh get vmi/vm-spread -o json | jq .spec.domain.cpu
{
  "cores": 4,
  "sockets": 2,
  "threads": 1
}
```

### `CoresThreads`

```yaml
$ ./cluster-up/kubectl.sh apply -f -<<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: spread
spec:
  cpu:
    guest: 8
  memory:
    guest: 1Gi
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachinePreference
metadata:
  name: spread
spec:
  cpu:
    preferredCPUTopology: Spread
    spreadOptions:
      across: CoresThreads
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-spread
spec:
  instancetype:
    name: spread
    kind: VirtualMachineInstancetype
  preference:
    name: spread
    kind: VirtualMachinePreference
  runStrategy: Always
  template:
    spec:
      domain:
        devices: {}
      volumes:
        - containerDisk:
            image: quay.io/containerdisks/fedora:39
          name: containerdisk
EOF
[..]
$ ./cluster-up/kubectl.sh get vmi/vm-spread -o json | jq .spec.domain.cpu
{
  "cores": 4,
  "sockets": 1,
  "threads": 2
}
```

### `SocketsCoresThreads`

```yaml
$ ./cluster-up/kubectl.sh apply -f -<<EOF
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachineInstancetype
metadata:
  name: spread
spec:
  cpu:
    guest: 16
  memory:
    guest: 1Gi
---
apiVersion: instancetype.kubevirt.io/v1beta1
kind: VirtualMachinePreference
metadata:
  name: spread
spec:
  cpu:
    preferredCPUTopology: Spread
    spreadOptions:
      across: SocketsCoresThreads
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-spread
spec:
  instancetype:
    name: spread
    kind: VirtualMachineInstancetype
  preference:
    name: spread
    kind: VirtualMachinePreference
  runStrategy: Always
  template:
    spec:
      domain:
        devices: {}
      volumes:
        - containerDisk:
            image: quay.io/containerdisks/fedora:39
          name: containerdisk
EOF
[..]
$ ./cluster-up/kubectl.sh get vmi/vm-spread -o json | jq .spec.domain.cpu
{
  "cores": 4,
  "sockets": 2,
  "threads": 2
}
```

## Functional Testing Approach

* Extend existing unit test coverage to ensure the expected topologies are generated within the VirtualMachineInstance.
* Unclear if any additional functional test coverage is required outside of this.

# Implementation Phases

TBD
