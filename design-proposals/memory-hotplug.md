# Overview

This proposal is about adding Memory hotplug support to KubeVirt. Memory hotplug will allow users to dynamically add
Memory resources to running VMs.

## Motivation

Memory hotplug is now a common virtualization feature, but KubeVirt lacks support for it. Live vertical scaling of VM
resources in general can be very useful in a variety of scenarios.

## Goals

* Hotplug and hot-unplug of Memory resources from VMs should be achievable by editing the VM object itself (no
  sub-resources)
* Underlying technology used for hotplug should be independent of the VM API, meaning we should be able to swap out how
  hotplug is technically performed in the future without impacting API compatibility
* Implementation should be achievable today, with Kubernetes APIs that are at least in beta. Unfortunately, at the time
  of writing, the Kubernetes vertical pod scaling API is only alpha

## Non Goals

* Revising existing hotplug mechanisms like Volume hotplug
* Dynamically applying affinity rules to VMIs

## Definition of Users

A user with (at a minimum) create, update, and patch access for VM objects within a namespace.

## User Stories

As a VM user, I would like to dynamically increase or decrease the number of Memory used by a VM without requiring the
VM to restart.

## Repos

kubevirt/kubevirt

## Design

In VM objects, the existing Memory `guest` field will now be dynamic, provided that it has been declared as such.
Declaring the field as dynamic will be done by adding a `memory` entry to a section under the VM spec,
called `liveUpdateFeatures` (definition during CPU hotplug implementation).   
Please note that prior to defining this field the hotplug action would not take place i.e. changing the number of guest
will be staged until further reboot. Under the `memory` entry, it will be possible to define a maximum number of guest.
That number is needed by LibVirt and will default to 4 times the initial number of sockets if not set. That default
value will be configurable in the KubeVirt CR. On VM startup, that value will translate to a new `maxGuest` entry
under `spec.domain.memory` in the VMI object. However, all VMI Memory fields will stay immutable, and Memory guest will
only be adjustable on the VM object. Once the hoplug action begins, subsequent changes to the number of guest would be
rejected until the hotplug process completes. Memory Resource update on the pod level must occur before the actual
hotplug action on the Libvirt level, hence, during the hotplug action a notifitcation will be created upon successful
update of the Memory resources on the pod level. It will indicate that now it safe to proceed to the hotplug action on
the Libvirt/QEMU level.

## Drawbacks / Linmitations

* Workload update method should be defined explicitly as `VMLiveUpdateFeatures` under the Kubevirt configuration.
* This feature will be incompatible with memory requests/limits, at least initially, since those values need to
  dynamically scale according to the current number of enabled memory
* Set the memory increment (new_guest - old_guest) for online memory expansion An integer multiple of 128 MB (operating
  system requirements)
* Memory hotplug depends on numa, VMs will have to be `NUMAable` to enable Memory hotplug

#### API Examples

##### VM Spec API

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  running: false
  liveUpdateFeatures:
  +   memory
+     maximumGuest: 4Gi
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-cirros
    spec:
      domain:
        memory:
          guest: 4Gi
```

##### VMI Spec API

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  labels:
    special: vm-cirros
  name: vm-cirros
spec:
  domain:
    memory:
+     maxGuest: 10Gi
  guest: 8Gi

```

## Status API

The VMI status must indicate whether Memory hotplug is taking place. As the hotplug process begins the current state
of `spec.domain.memory` needs to be propagated back to the VM/VMI `status.currentMemory` to allow the user to have an
indication of the hotplug processing.

```
	// Memory allow specified the detailed inside the vmi.
		CurrentMemory *Memory `json:"currentMemory,omitempty"`
```

Conditions on the VM and VMI should indicate the status of the pending Memory hotplug actions. When errors occur, the
conditions should propagate the errors to the VM.Status.Conditions field so users can receive feedback into the status
of their declared state.

```
// Indicates that the VMI is in progress of Hot memory Plug/UnPlug
VirtualMachineInstanceMemoryChange VirtualMachineInstanceConditionType = "HotMemoryChange"
```

Upon succesful hotplug action the condition will be removed. Otherwise, a human readable reason and message indicating
whether the hotplug change is pending or encountering errors.

## Extensibility

We will be able to use the `VMLiveUpdateFeatures` section to manage things like cpu/memory hotplug .

## Update/Rollback Compatibility

The feautre doesn't intorduce any breaking changes so existing functionality of versions prior to this feature should
not be impacted upon update/rollback. However if existing workloads would like to opt-in for this feature after an
update, the VMI should be re-created since the addition of new fields to the VMI API.

## Functional Testing Approach

Functional tests will simply:

* Create and start a VM with `memeory` set as one of th `liveUpdateFeatures`
* Modify the number of memory guest for the VM
* Ensure the guest sees the changed number of memory
* Ensure the memory resources on the virt-launcher pod were adjusted accordingly
* Ensure interoperability with non-hotplug related live-migration occasions (workload-update, evacuation,
  user-triggered)

## Implementation Phases

### Phase 1: Creating the KubeVirt API and linking it to the LibVirt memory hotplug API

### Phase 2: Triggering a live migration when the number of sockets changes, to adjust the virt-launcher memory resources

The
VMI [workload update controller](https://github.com/kubevirt/kubevirt/blob/ce8aab7874d2e1586787e3e2a17306b7edca1b8a/pkg/virt-controller/watch/workload-updater/workload-updater.go#L476)
is extended to live migrate VMIs to satisfy updates to a VMI resource request/limits field.

Sequence of events:

* User updates memory guest on a vm.spec.template.spec.domain.memory
* VM controller determines the request/limits required and writes the necessary changes to the active vmi.Spec
* VM controller writes the required request/limits in the vmi.Status
* VMI workload update controller detects the mismatch between the resource request/limits on the pod vs vmi.Spec and
  live migrates the VMI to satisfy vmi.spec
* Migration controller constructs a new target pod with the desired memory request/limits
* Migration controller identfies a match between the required request/limits on the vmi.Status and those of the VMI
  target pod
* Migration controller marks the VMI as safe to hotplug.
* VMI live migrates.
* virt-handler identifies the mark and triggers the hotplug/hot-unplug LibVirt action after the live migration completes
* VM status is updated to reflect that the pending hotplug/hot-unplug actions have completed

### Phase 3 (future): Using the [Kubernetes vertical scaling API](https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources) once it reaches a more mature state

The inplace vertical pod feature allows us to dynamically change the VMI pods resource request/limits. This feature will
allow us to avoid live migration, however the issue for us is timing. This feature is alpha in Kubernetes 1.27. We need
it to be at least in beta to integrate it without risking stability and API reliability.

Sequence of events:

* User updates memory guest on a vm.spec.template.spec.domain.memory
* VM controller writes these change to the active vmi.Spec
* VM controller writes the required request/limits in the vmi.Status
* VMI controller updates the active pod spec to reflect the request/limit changes.
* Pod changes are applied.
* VMI controller identfies a match between the required request/limits on the vmi.Status and those of the VMI pod
* VMI controller marks the VMI as safe to hotplug.
* virt-handler identifies the mark and triggers the hotplug action.
* VM status is updated to reflect that the pending hotplug actions have completed.

#### References

* Libevirt domain XML memory hotplug

  https://libvirt.org/formatdomain.html#memory-allocation

  https://libvirt.org/formatdomain.html#memory-devices

* Kubernetes inplace pod updates

  https://github.com/kubernetes/kubernetes/pull/102884

  https://github.com/kubernetes/enhancements/tree/master/keps/sig-node/1287-in-place-update-pod-resources

* cpu-hotplug design proposals

  https://github.com/kubevirt/community/blob/main/design-proposals/cpu-hotplug.md