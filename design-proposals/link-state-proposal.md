Adding Link State Management for vNICs
=

# Overview

The purpose of this design proposal is to introduce support for link state management (up/down) in KubeVirt. This feature
will enable VM owners to dynamically control the link state. It aligns KubeVirt with traditional virtualization platforms
and enhances its utility in environments requiring precise network state control.

By introducing link state management, KubeVirt will:
- Provide VM-level controls to configure the link state of individual interfaces.
- Report link state info via the VirtualMachineInterface status.

This proposal outlines the necessary changes to KubeVirt, including API changes.

## Motivation

Currently, KubeVirt does not support modifying or reporting the link state of virtual network interfaces.
This limitation can lead to challenges in implementing advanced networking scenarios, such as:
- Simulating link failures for testing and debugging.
- Managing multi-network setups, where some interfaces need to be intentionally disabled.

## Goals

1. Allow users to configure the link state (up/down) of individual interfaces through the KubeVirt API.
2. Enable link state configuration at both VM creation and runtime.
3. Report the current link state of each VM interface.

## Non Goals

1. Affect network plumbing on the host - the setup process will remain unchanged change.

## Definition of Users

- VM owner

## User Stories

As a VM owner I would like to be able to:

- Start the VM with one or more interfaces with their link set to `down`.
- Toggle the link state of one or more interfaces of a running VM.
- Hot plug an interface with a link state set to `down`.

- I want my interfaces link states to persist following a migration.

## Repos

kubevirt/kubevirt.

# Design

## API Addition
Currently, the `Interface` struct has the `State` field (currently used for hot-unplug):
```go
type Interface struct {
    ...
    State InterfaceState `json:"state,omitempty"`
    ...
}
```

Type `InterfaceState` is an enum, currently having a single value:  
```go
type InterfaceState string

const (
	InterfaceStateAbsent InterfaceState = "absent"
)
```

The following values will be added to specify the required link state:
```go
const (
    InterfaceStateLinkUp   InterfaceState = "linkUp"
    InterfaceStateLinkDown InterfaceState = "linkDown"
)
```

An empty value will be considered as `linkUp`.

The network validator will be adjusted to allow the two new values for:
1. All interfaces connected to primary and secondary networks.
2. All bindings - core and plugins.

virt-launcher's `Converter` component will be adjusted to take the interface State field into account when creating a new domain XML.
In case the field's value is `InterfaceStateLinkDown`, the interface will be created as follows:
```xml
<interface>
  ...
  <link state='down'/>
  ...
</interface>
```

> [!NOTE]
> For additional details please see libvirt's [documentation](https://libvirt.org/formatdomain.html#modifying-virtual-link-state).

virt-launcher's `LibvirtDomainManager.SyncVMI` will be extended to support updating the interfaces link state while the VM is running.

virt-launcher's NIC hotplug logic will be adjusted to support hot-plugging a NIC while its link state is down.

## Network Binding Plugins Support

When using the functionality purposed by this proposal with network binding plugins using a sidecar container -
its up to the plugin's vendor to support this functionality - as they control the interface section in the domain XML.

## Link State Reporting

A new field will be added to the VMI interface status that will be used for reporting the interface's current link state (up/down):
```go
type VirtualMachineInstanceNetworkInterface struct {
	...
	LinkState string `json:"linkState,omitempty"`
	...
}
```

virt-handler will be responsible for reporting this field's value, based on domain information received from virt-launcher.

## API Examples
### Controlling The Link State
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: my-vm
spec:
  template:
    spec:
      domain:
        devices:
          interfaces:
            - name: default
              state: linkDown
              masquerade: { }
      networks:
        - name: default
          pod: { }
```
### Link State Reporting
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: my-vm
spec:
  domain:
    devices:
      interfaces:
      - name: default
        state: linkDown
        masquerade: { }
  networks:
  - name: default
    pod: { }
status:
  interfaces:
    - name: default
      linkState: down
```

## Scalability

This proposal does not affect scalability.

## Update/Rollback Compatibility

As this functionality will require the cooperation of the virt-launcher component, it could only be supported by
up-to-date virt-launcher pods.

## Functional Testing Approach

The following e2e scenarios will be added:
- A VM started with interface with a link state set to `down`, then changed to `up`.
- Setting a link state of a running VM to `down`, then `up`.
- Hot-plugging an interface connected to a secondary network with link state set to `down`, then changed to `up`.
- Migrating a VM that has an interface with a link state set to `down`, then changed to `up`.

# Implementation Phases

1. Addition of the new `InterfaceState` values.
2. Adjusting the network validator to accept the new `InterfaceState` values.
3. Adjusting virt-launcher:
   - On creation
   - When VM is running
   - On hotplug
4. Addition of e2e tests.
5. Addition of the `LinkState` field.
6. Adjusting virt-handler's reporting logic.
7. Adjusting e2e tests.