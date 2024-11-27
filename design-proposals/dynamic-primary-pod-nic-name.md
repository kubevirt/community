Support Dynamic Primary Pod NIC Name
=

# Summary

This proposal proposes a mechanism to dynamically determine the primary pod interface for KubeVirt virtual machines (
VMs).
The goal is to move away from the hard coded `eth0` interface and allow flexible, environment-driven interface
selection.

# Overview

## Motivation

In KubeVirt, each VM runs as a Kubernetes pod, and these pods have network interfaces attached to different networks.
The default primary pod interface is often named `eth0`.
Currently, KubeVirt attaches the VM’s interface destined to the pod network to the pod’s `eth0` interface.
However, users may require custom network setups where the pod’s primary interface is not named `eth0`.
For example, a CNI that sets up two interfaces in one CNI ADD may be used:
[ovn-k](https://github.com/ovn-org/ovn-kubernetes) could be configured to set up `eth0` and another interface that will
be considered as the primary interface.
We wish to connect the VMI's primary interface to the pod's actual primary interface.

## Goals

- Allow the primary pod interface name to be dynamically determined based on the environment.
- Ensure backward compatibility, so existing VMs with `eth0` as the primary interface will not break.

## Non Goals

- Migration of an existing VM whose pod primary interface is `eth0` to a pod whose primary interface has a custom name
  and vice versa.
- Support Istio integration with Masquerade binding and a custom primary interface name.

## Definition of Users

- VM owner

## User Stories

As a VM owner:

- I want my VMs to be able to bind the interface destined to pod network, to the pod's primary interface, no matter its
  name.
- I want my VMs to be able to connect to the cluster's default network, as they do today.
- I want my VMs' interface status reporting to be accurate when the interface destined to pod network is bound to a pod
  interface with a custom name.
- I want to be able to use core bindings and network binding plugins when the primary pod interface has a custom name.
- I want to be able to migrate my VM when the pod's primary interface has a custom name.

## Repos

- https://github.com/kubevirt/kubevirt

# Design

## Assumptions

The KubeVirt components participating in this proposal are:

- VMI controller
- virt-handler
- virt-launcher

Currently, the VMI controller sets the mapping between logical network name to pod interface name, for example:
A VMI interface connected to a secondary network named `blue` will be mapped to a pod interface named `pod16477688c0e`.

The VMI's primary interface is always mapped to pod interface named `eth0`.

> [!NOTE]
> The VMI's primary interface can connect to one of the following mutually exclusive options:
> 1. [Pod network](https://kubevirt.io/api-reference/v1.3.1/definitions.html#_v1_podnetwork) - the cluster's default
     network
> 2. [Multus default network](https://kubevirt.io/api-reference/v1.3.1/definitions.html#_v1_multusnetwork) - an
     alternative to the cluster's default network

When performing their network setup logic, virt-handler and virt-launcher perform the same naming scheme calculation to
map the logical network name to pod interface name.

The following table describes which information is accessible to each relevant KubeVirt component:

| Component/has access to | VirtualMachineInstance | Pod   | Cluster-config |
|-------------------------|------------------------|-------|----------------|
| VMI controller          | True                   | True  | True           |
| virt-handler            | True                   | False | True           |
| virt-launcher           | True                   | False | False          |

Cluster wide config passes from virt-handler to virt-launcher via their dedicated gRPC channel.

The pod interfaces' naming scheme is currently independently calculated at each of the three components, based on input
from VMI.Spec.

The information about the pod's primary interface name should be propagated to all three components.

## Solution

### Inferring the primary interface name from Multus’ network-status annotation

When KubeVirt is used in a cluster that has Multus as its default CNI, Multus adds the
`k8s.v1.cni.cncf.io/network-status` annotation to the virt-launcher pod following its network attachment process.
This annotation includes information about all the network interfaces attached to the pod.

The VMI controller currently uses this annotation for:

- Updating the VMI status interfaces section with the
  `multus-status` [info-source](https://kubevirt.io/api-reference/v1.3.1/definitions.html#_v1_virtualmachineinstancenetworkinterface).
- Creating the `kubevirt.io/network-info` annotation, which is the data source for DownwardAPI when using SR-IOV NICs
  and certain network binding plugins.

The `k8s.v1.cni.cncf.io/network-status` annotation shall be parsed to infer the default (primary) network interface
instead of relying on a hard coded
interface name (like eth0).

> [!NOTE]
> [ Kubernetes Network Custom Resource Definition De-facto Standard](https://github.com/k8snetworkplumbingwg/multi-net-spec/tree/master/v1.3)
> states:
>
> "default" - This required key’s value (type boolean) shall indicate that this attachment is the result of the
> cluster-wide default network.
> Only one element in the Network Attachment Status Annotation list may have the "default" key set to true.
>
> "interface" - This optional key’s value (type string) shall contain the network interface name in the pod’s network
> namespace
> corresponding to the network attachment.

VMI controller will look for the first entry to have `"default": true` and a non-empty `interface` field.
If such entry exists - it would be considered as the pod's primary interface name and will override the default `eth0`.

Example 1:

When a CNI sets a custom name for the pod's primary interface we can expect the following:

```json
[
  {
    "name": "cluster-network",
    "interface": "custom-iface",
    "ips": [
      "10.128.0.4"
    ],
    "mac": "0a:58:0a:80:00:04",
    "default": true,
    "dns": {}
  },
  {
    "name": "cluster-network",
    "interface": "eth0",
    "ips": [
      "10.244.0.4"
    ],
    "mac": "0a:58:0a:f4:00:04",
    "dns": {}
  }
]
```

Since the first entry has both `"default": true` and `"interface": "custom-iface"`, KubeVirt will infer that the primary
interface name is `custom-iface`.

Example 2:

```json
[
  {
    "name": "k8s-pod-network",
    "ips": [
      "10.244.196.146",
      "fd10:244::c491"
    ],
    "default": true,
    "dns": {}
  },
  {
    "name": "meganet",
    "interface": "pod7e0055a6880",
    "mac": "8a:37:d9:e7:0f:18",
    "dns": {}
  }
]
```

> [!NOTE]
> The example above was taken from a [kubevirtci](https://github.com/kubevirt/kubevirtci) cluster
> using [Calico](https://www.tigera.io/project-calico/) as its default CNI.

Since the first network-status entry has `default:true` but does not have `interface:<value>`, KubeVirt will use `eth0`
as the primary interface name.

> [!NOTE]
> The `interface` field is not present in the example above since Calico hasn't reported its interface name as part of
> the CNI ADD results.

### Reporting Pod Interface Name Per Interface

Currently, the `VirtualMachineInstanceStatus` struct holds a slice of interface status objects:

```go
type VirtualMachineInstanceStatus struct {
...
Interfaces []VirtualMachineInstanceNetworkInterface `json:"interfaces,omitempty"`
...
}
```

A new field will be added to the `VirtualMachineInstanceNetworkInterface` struct which will be used to report
the pod interface name connected to the network the VM interface should connect to:

```go
type VirtualMachineInstanceNetworkInterface struct {
...
// PodInterfaceName represents the name of the pod network interface
PodInterfaceName string `json:"podInterfaceName,omitempty"`
...
}
```

The VMI controller will set this field's value from the following sources:

- The output of the existing naming scheme calculation based on VMI.Spec
- Multus' network-status annotation

virt-handler and virt-launcher will be adjusted to read this value instead of performing an independent naming scheme
calculation.

For newly created VMs - the VMI object will be handed over to the virt-handler after the `PodInterfaceName` fields will
be filled.
virt-handler in turn - will send virt-launcher the `SyncVMI` command with all the `PodInterfaceName` data present.

For existing VMs, the information required for network setup is already saved in the persistent cache.

In case the `PodInterfaceName` field will be empty, virt-handler and virt-launcher will fall back to the existing naming
scheme calculation based on the VMI.spec.

#### API Example

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: myvmi
spec:
  ...
status:
  ...
  interfaces:
      - infoSource: 'domain, guest-agent, multus-status'
        podInterfaceName: custom-iface
        interfaceName: eth0
        ipAddress: 10.131.0.29
        ipAddresses:
          - 10.131.0.29
        mac: '52:54:00:c0:35:78'
        name: passtnet
        queueCount: 1
  ...
```

#### Pros

- Manual configuration is not required
- Provides backward compatibility
- This solution has the potential to ease future changes in the name generation, as it is generated once at the
  controller and not at each component (which may cause compatibility issues when the system is upgraded).

#### Cons

- Adds API
- Exposes an implementation detail to VM owners which is probably not useful for them.

Alternative solutions could be found in [Appendix A](#appendix-a---alternative-solutions).

### Setting a constant tap name for the VM's primary interface

Currently, for bindings on which KubeVirt is responsible for creating a tap, the tap's name is 
derived from the pod interface's name.

virt-handler will create the tap associated with the VM's primary interface as `tap0`.
virt-launcher will consume `tap0` for the VM's primary interface when performing its network setup flow.

Doing so opens up an option to support a futuristic migration to a target pod with a different
primary interface name without modifying the domain XML.

## Network Binding Plugins

A virtual machine's NIC could be bound to the pod's primary NIC using
a [network binding plugin](https://kubevirt.io/user-guide/network/network_binding_plugins/).

### CNI

KubeVirt's VMI controller is responsible for specifying the
proper [NetworkSelectionElement](https://github.com/k8snetworkplumbingwg/network-attachment-definition-client/blob/506cfdac925790adf2f56f27740d2e87eaf2c83c/pkg/apis/k8s.cni.cncf.io/v1/types.go#L140)
objects that will cause Multus to invoke a network binding plugin's CNI.

Since this specification happens before the pod is created, it is not possible for KubeVirt to know in advance what will
the pod's primary NIC's name be.

For binding plugins authors, the following option could be implemented.

#### Additional CNI Configuration field

Binding plugins' authors could add a configuration field to specify a list of possible primary interface names:

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: my-nad
spec:
  config: '{
  "cniVersion": "1.0.0",
  "type": "my-cni"
  "potentialPrimaryInterfaceNames": "custom-iface,eth0",
}'
```

These will be specified under the plugin's NetworkAttachmentDefinition object.

The binding plugin's CNI could look up these names in order.

### Sidecar

Network binding plugins could use
a [sidecar container](https://github.com/kubevirt/kubevirt/blob/6aba1bc426d8b2f93270ba649ca33b9bbd925729/docs/network/network-binding-plugin.md?plain=1#L177)
in order to:

- Mutate the domain xml that is passed to libvirt's daemon (virtqemud).
- Run arbitrary code in the context of the virt-launcher pod.

Since protocol between the virt-launcher and the sidecar provides the VirtualMachineInstanceObject, the sidecar could
infer what is the pod's primary interface by examining it.

In that case, a change to the network binding plugins' developer guide will be required.

## Scalability

The proposed solution does not affect scalability.

## Update/Rollback Compatibility

Existing VMs should continue working after an upgrade.
VMs running on nodes with old virt-handler and virt-launcher should continue working as expected.
This is because old virt-handler and virt-launcher are not aware of the new field that reports the pod interface name.

When performing migration in order to upgrade the virt-launcher pod, there won't be a change to the pod interface name.
Changing the primary pod interface name on migration is a non-goal.

## Functional Testing Approach

In order to test this functionality under kubevirt/kubevirt, a new dummy CNI would be used.
This dummy CNI will create a primary pod interface with a custom name.

Two scenarios will be added to KubeVirt's network e2e tests with the special CNI:

- VM Creation
- VM Migration

Additional, more extensive tests will be performed outside kubevirt/kubevirt scope.

# Implementation Phases

- Map the affected areas.
- Complete the design proposal
- Expose new API on the VMI object.
- Consume the pod interface name instead of the hard-coded `eth0` for the VMI primary interface

# Appendix A - Alternative Solutions

<!-- TOC -->

* [Inferring the primary interface name from Multus’ network-status annotation](#inferring-the-primary-interface-name-from-multus-network-status-annotation-1)
    * [Alternative 1 - Report Primary Pod Interface Name Per VM](#alternative-1---report-primary-pod-interface-name-per-vm)
    * [Alternative 2 - Use annotations instead of K8s API](#alternative-2---use-annotations-instead-of-k8s-api)
    * [Alternative 3 - Use the existing Downward API mechanism](#alternative-3---use-the-existing-downward-api-mechanism)
* [Adding a cluster-wide preferred pod NIC name](#adding-a-cluster-wide-preferred-pod-nic-name)
* [Specifying the primary pod NIC name in advance in K8s or on the node](#specifying-the-primary-pod-nic-name-in-advance-in-k8s-or-on-the-node)

<!-- TOC -->

## Inferring the primary interface name from Multus’ network-status annotation

### Alternative 1 - Report Primary Pod Interface Name Per VM

A new field will be added to the `VirtualMachineInstanceStatus` struct, which will be used to report the pod's primary
interface's name:

```Go
type VirtualMachineInstanceStatus struct {
...
// PrimaryPodInterfaceName represents the pod primary NIC's name
PrimaryPodInterfaceName string `json:primaryPodInterfaceName,omitempty`
...
}
```

The VMI controller will set this field's value from Multus' network-status annotation or fallback to `eth0`.

virt-handler and virt-launcher will be adjusted to read this value, and take it into account when performing the naming
scheme calculation.
In case the field will be empty, virt-handler and virt-launcher will fall back to the existing naming scheme calculation
based on the VMI.spec.

#### API Example

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: myvmi
spec:
  ...
status:
  ...
  primaryPodInterfaceName: custom-iface
  interfaces:
    - infoSource: 'domain, guest-agent, multus-status'
      interfaceName: eth0
      ipAddress: 10.131.0.29
      ipAddresses:
        - 10.131.0.29
      mac: '52:54:00:c0:35:78'
      name: passtnet
      queueCount: 1
  ...
```

#### Pros

- Manual configuration is not required
- Provides backward compatibility

#### Cons

- Adds API
- Exposes an implementation detail to VM owners which is probably not useful for them
- There is no correlation with the relevant interface (which might not even exist in case the VM is not connected to the
  pod network).
- This field is not relevant in case the VM is using only secondary networks

### Alternative 2 - Use annotations instead of K8s API

A new annotation will be added to the VirtualMachineInstance object.
It could represent either alternative 1 or 2 in annotation form.

#### Pros

- Provides backward compatibility

#### Cons

- Since information is going to be passed between KubeVirt's components it is not ideal to use semiformal API.
- Annotations usually flow from VirtualMachineInstance to Pod and not vice versa.

### Alternative 3 - Use the existing Downward API mechanism

Currently, the VMI controller sets the `kubevirt.io/network-info` annotation on the virt-launcher pod in two cases:

1. The VMI has an SR-IOV NIC.
2. The VMI uses a network binding plugin that is configured to use downward API.

The annotation's value is derived fom Multus' network-status annotation and is injected into the virt-launcher pod
using K8s' [Downward API](https://kubernetes.io/docs/concepts/workloads/pods/downward-api/) mechanism.

A new field will be added to the `NetworkInfo` struct:

```go
type NetworkInfo struct {
Interfaces []Interface `json:"interfaces,omitempty"`
PrimaryPodInterfaceName string `json:primaryPodInterfaceName,omitempty`
}
```

A new mechanism should be added between virt-handler and virt-launcher to share this information, because:

1. The primary pod interface name is also required by virt-handler
2. The virt-handler does not have access to pod objects

virt-handler will poll the virt-launcher's file-system for the `/etc/podinfo` file.
virt-handler will not be able to start the network setup until it successfully read and parsed the content of the file
above.

#### Cons

- All virt-launcher pods will have to use the downward API mechanism instead of just the subset using it today
- Complicates the network setup and overall VM startup sequence

## Adding a cluster-wide preferred pod NIC name

A new field will be added in KubeVirt's cluster-wide network configuration, which will specify the preferred pod
interface name that should be bound to the VM interface destined for pod network.

virt-handler and virt-launcher will be adjusted to prefer this value when trying to bind the primary network interface.
They will both fallback to `eth0` in case they cannot find the preferred interface name.

```go
type NetworkConfiguration struct {
...
PreferredPodPrimaryInterfaceName string `json:"preferredPodPrimaryInterfaceName,omitempty"`
...
}
```

### API Example

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    network:
      preferredPodPrimaryInterfaceName: special-nic
```

### Pros

- Does not require Multus to work

### Cons

- The cluster admin or an [Operator](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) needs to know in
  advance which interface name to use
- Error prone
- Has cluster-wide granularity
- Requires a change to the communication protocol between virt-handler and virt-launcher to pass the cluster-wide
  setting to virt-launcher

## Specifying the primary pod NIC name in advance in K8s or on the node

When creating a VM with secondary networks, KubeVirt specifies the pod interfaces’ names for all secondary interfaces.
It is currently not possible to specify the primary interface’s name when templating the virt-launcher pod, since there
is no K8s API that enables it.

The primary pod interface name is set by the container runtime (for example cri-o), before invoking the default CNI.

### No Go

- There is no K8s API to name the pod's primary NIC
- There is no Multus API or config to name the pod's primary NIC
