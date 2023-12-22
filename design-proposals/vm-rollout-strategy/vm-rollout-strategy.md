# Overview

Today changes to a VM object (i.e. changing VM.spec.template) get applied, but only take effect upon the next restart - let's call these pending changes “staged”. This is a proposal to introduce a custer-level rollout strategy field to control how changes to the VM object will propagate to the VirtualMachineInstance object without interupting the workload.


## Motivation

In most cases, the user would prefer the changes to the VM object to eventually propagate to the VMI object, as long as the rollout does not interrupt the workload. This behavior would be similar to how the Deployments roll out ReplicaSets.


## Goals

-   Provide a way to eventually propagate changes from a running VM object to its corresponding VMI without interrupting the workload.
    
-   Maintain backwards compatibility for cluster owners that prefer keeping staged changes until next VM reboot.
    

## Non Goals

-   Live uppdate for **all** fields of the Virtual Machine API. Only select fields can be propagated to the guest. 
-   Make LiveUpdate the default. We defer this decision.
-   Support upgrades of VMs that are using the `liveUpdateFeatures`.

## Definition of Users

A VM owner - with edit permissions for VM objects within a namespace.


## User Stories

As a VM owner I want changes that I applied to the VM object to propagate to the VMI if possible.

As a VM owner, I want to be notified about changes that require a restart of the VM.

## Repos

kubevirt/kubevirt


# Design

> **Note:** Below we are only discussing changes to fields that support live updates.
  The term `all` only refers to fields that can be live-updated. 

## KubeVirt CR

A new field called `vmRolloutStrategy` will be introduced in the KubeVirt CR. This field will control the behavior of the VM objects. Two possible values will be available:

1. `LiveUpdate` means all changes to the VM object will be eventually applied to the VMI.
    
2. `Stage` would mean the current familiar behavior remains - all changes to the VM object will be staged and applied after a restart.

In both cases, a condition called `RestartRequired` will be added to the VM to reflect the need for a restart and the reason why.

The default value of the `vmRolloutStrategy` field if not specified will be `Stage`.

The `vmRolloutStrategy` field will be usable only if the `VMLiveUpdateFeatures` feature gate is enabled.


## VM API

The VM level `liveUpdateFeatures` field will be removed.
VMs using the `liveUpdateFeatures` will be rejected by the API.

If `RestartRequired` condition is present, the VM has to be restarted.

## Changes to CPU/Memory hotplug workflow

### CPU Hot Plug
Since the LiveUpdateFeatures API will be deprecated the `spec.template.spec.domain.cpu.maxSockets`field of the VM object will become mutable and stays optional.
Changes to `spec.template.spec.domain.cpu.maxSockets` would require a restart of the VM. This would also set the `RestartRequired` condition.

If it is not set by the user then the maxSockets value will continue to be internally computed according the cluster-wide `maxHotPlugRatio`, as it is done today.

CPU hot plug is triggered by changing `spec.template.spec.domain.cpu.sockets`. This change will be propagated to the VMI object as long as its value is lower than the value of `MaxSockets` on the last VM startup.
Increasing the value of the `spec.template.spec.domain.cpu.sockets` beyond the `MaxSockets` will require a restart to take effect.

The `RestartRequired` condition is guaranteed to be present on the VM object right after a change to the VM object was made *and* a restart is required.

### Memory Hot Plug

The logic of memory hot plug will be similar to CPU hot plug, with the exception of the fields names.
`spec.template.spec.domain.memory.maxGuest` will become mutable and will remain optional.
A default value for the `MaxGuest` memory value will continue to be internally calculated based on the cluster setting of `maxHotplugRatio` 

As part of this change, the `MaxGuest` field will be renamed to `MaxMemory`.

All other hot plug behavior wil remain the same.

## Drawbacks / Limitations

- No cluster- and VM-level fine-grained control over rollout settings of specific features.
- No VM-specific controll over the rollout strategy.
- Changing the cluster-level rollout strategy while VMs are running will lead to an udefined behavior 

## API 

 
### Before the changes in this proposal:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
    name: vm-cirros
spec:
  liveUpdateFeatures:
    cpu:
      maxSockets: 8
    memory:
      maxGuest: 256Mi
  running: false
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
        memory:
          guest: 128Mi
```

### After the changes in this proposal:

```yaml
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
      domain:
        cpu:
          sockets: 2
          cores: 4
          threads: 2
          maxSockets: 8  # optional
        memory:
          guest: 128Mi
          maxGuest: 256Mi  # optional
```
### CPU Hot plug example
 Before CPU hot plug:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  template:
    spec:
      domain:
        cpu:
          sockets: 2
          cores: 1
          threads: 1
```

After CPU hot plug:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  template:
    spec:
      domain:
        cpu:
          sockets: 3  # This value changed from '2' to '3', this is the only change
          cores: 1
          threads: 1
```

### Status API

A condition called `RestartRequired` will be added to the VM requesting a restart if the changes cannot be immediately applied.
The condition will contain the reason for not being able to reach the desired state requested in the VM spec.

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  template:
    spec:
      domain:
        cpu:
          sockets: 42  # Changed to an unreasonable value, which does require a restart
          cores: 1
          threads: 1
status:
  conditions:
    - type: RestartRequired
      value: true
      reason: 'The number of sockets is too high and requires a VM restart in order to become effective".
```

## Update/Rollback Compatibility

No upgrade consideration has to be made, as this is currently a non-stable API.
