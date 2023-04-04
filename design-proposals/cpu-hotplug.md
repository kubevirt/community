# Overview
This proposal is about adding CPU hotplug support to KubeVirt.
CPU hotplug will allow users to dynamically add/remove CPU resources to/from running VMs.
While this design is focused on CPU hotplug, it aims to provide a generic API design for resource hotplug that will be reusable for things like memory.

## Motivation
CPU hotplug is now a common virtualization feature, but KubeVirt lacks support for it.
Live vertical scaling of VM resources in general can be very useful in a variety of scenarios.

## Goals
- Hotplug and hot-unplug of CPU resources from VMs should be achievable by editing the VM object itself (no sub-resources)
- Underlying technology used for hotplug should be independent of the VM API, meaning we should be able to swap out how hotplug is technically performed in the future without impacting API compatibility
- Implementation should be achievable today, with Kubernetes APIs that are at least in beta. Unfortunately, at the time of writing, the Kubernetes vertical pod scaling API is only alpha

## Non Goals
- Revising existing hotplug mechanisms like Volume hotplug
- Dynamically applying affinity rules to VMIs

## Definition of Users
A user with (at a minimum) create, update, and patch access for VM objects within a namespace.

## User Stories
As a VM user, I would like to dynamically increase or decrease the number of CPUs used by a VM without requiring the VM to restart.

## Repos
kubevirt/kubevirt

# Design
In VM objects, the existing CPU `sockets` field will now be dynamic, provided that it has been declared as such.
Declaring the field as dynamic will be done by adding a `cpu` entry to a new section under the VM spec, called `liveUpdateFeatures`.
Under the `cpu` entry, it will be possible to define a maximum number of sockets. That number is needed by LibVirt and will default to 4 times the initial number of sockets if not set. That default value will be configurable in the KubeVirt CR.
On VM startup, that value will translate to a new `maxSockets` entry under `spec.domain.cpu` in the VMI object. However, all VMI CPU fields will stay immutable, and CPU sockets will only be adjustable on the VM object.

Increasing the number of CPU sockets will not only add CPUs to the guest but also increase the CPU resources available to the virt-launcher pod, via a migration (more on that in the Implementation section).
The way LibVirt handles CPU hotplug is by exposing the maximum number of CPUs to the guest and turning off the unused CPUs. In this document, when we refer to "adding" CPUs to the guest, we're really just turning them on. This is why we need to define a maximum number of sockets in advance.

## Drawbacks / Limitations
- Since CPU hot-(un)plug involves a live migration, VMs will have to be `LiveMigratable` to enable CPU hotplug
- Each disabled vCPU, so `(maxSockets - sockets) * cores * thread`, consumes 8MiB of overhead memory
- This feature will be incompatible with CPU requests/limits, at least initially, since those values need to dynamically scale according to the current number of enabled CPUs

## API Examples

### VM Spec API

```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  running: false
  liveUpdateFeatures:
    cpu
      maximumSockets: 4
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-cirros
    spec:
      domain:
        cpu:
          sockets: 2
          cores: 4
          threads: 2
```
In this example, the user will be able to double the guest CPU resources by increasing the value of `sockets`.

### VMI Spec API

The VM definition above would translate to the following VMI:


```
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  labels:
    special: vm-cirros
  name: vm-cirros
spec:
  domain:
    cpu:
      maxSockets: 4
      sockets: 2
      cores: 4
      threads: 2
```

### Status API

The VMI status must indicate the current number of CPU sockets. The value needs to be propagated back to the VM status to allows the user to have an indication of when their changes applied to the running VMI.

The user should be able to inspect the VM.Status value below to determine the state of their changes. This is similar in concept to someone inspecting a Deployments status to determine the the number of pods that are available.

```
// The number of sockets the vm has currently active.
vmi.Status.Resources.Sockets.Ready
vm.Status.Resources.Sockets.Ready
```

Conditions on the VM and VMI should indicate the status of the pending CPU hotplug actions. When errors occur, the conditions should propagate the errors to the VM.Status.Conditions field so users can receive feedback into the status of their declared state.

```
// VirtualMachineCPUReady indicates that the desired amount of cpu is active on the vmi
VirtualMachineCPUReady VirtualMachineConditionType = "CPUReady"
```

When the desired state is equal to the ready state, then these conditions will have `Status=True`, otherwise `Status=False` with a human readable reason and message indicating whether the hotplug change is pending or encountering errors.

## Extensibility
We will be able to use the new `liveUpdateFeatures` section to manage things like memory hotplug in the future.

## Update/Rollback Compatibility
This is an entirely new feature, so updates from older versions will not be a problem.

## Functional Testing Approach
Functional tests will simply:
- Create and start a VM with `cpu` set as one of the `liveUpdateFeatures`
- Modify the number of CPU sockets for the VM
- Ensure the guest sees the changed number of CPUs
- Ensure the CPU resources on the virt-launcher pod were adjusted accordingly

# Implementation Phases

The first two phases will happen together in the same pull request (PR), since the first phase is mostly useless without the second one.
In the unlikely case that hot-unplugging CPUs is more complicated than anticipated, hotplug-only could be submitted as a first PR, then hot-unplug later on as a second PR.
The third phase will happen later as a separate PR. It should be able to transparently replace the migration mechanism without API changes.

## Phase 1: Creating the KubeVirt API and linking it to the LibVirt CPU hotplug API
## Phase 2: Triggering a live migration when the number of sockets changes, to adjust the virt-launcher CPU resources

The VMI [workload update controller](https://github.com/kubevirt/kubevirt/blob/ce8aab7874d2e1586787e3e2a17306b7edca1b8a/pkg/virt-controller/watch/workload-updater/workload-updater.go#L476) is extended to live migrate VMIs to satisfy updates to a VMI resource request/limits field.

Sequence of events:
- User updates CPU sockets on a vm.spec.template.spec.domain.cpu
- VM controller determines the request/limits required and writes the necessary changes to the active vmi.Spec
- VMI workload update controller detects the mismatch between the resource request/limits on the pod vs vmi.Spec and live migrates the VMI to satisfy vmi.spec
- Migration controller constructs a new target pod with the desired CPU request/limits
- VMI live migrates
- virt-handler triggers the hotplug/hot-unplug LibVirt action after the live migration completes
- VM status is updated to reflect that the pending hotplug/hot-unplug actions have completed

## Phase 3 (future): Using the [Kubernetes vertical scaling API](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources ) once it reaches a more mature state

The inplace vertical pod feature allows us to dynamically change the VMI pods resource request/limits. This feature will allow us to avoid live migration, however the issue for us is timing. This feature is alpha in Kubernetes 1.27. We need it to be at least in beta to integrate it without risking stability and API reliability.

Sequence of events
- User updates CPU sockets on a vm.spec.template.spec.domain.cpu
- VM controller writes these change to the active vmi.Spec
- VMI controller updates the active pod spec to reflect the request/limit changes
- Pod changes are applied
- virt-handler triggers the hotplug/hot-unplug action after the inplace update completes.
- VM status is updated to reflect that the pending hotplug/hot-unplug actions have completed

# References

- LibVirt Domain XML CPU hotplug
https://libvirt.org/formatdomain.html#cpu-allocation
- Kubernetes inplace pod updates
https://github.com/kubernetes/kubernetes/pull/102884
https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources