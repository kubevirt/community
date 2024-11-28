# Overview

The [VM Rollout Strategy](https://kubevirt.io/user-guide/operations/vm_rollout_strategies/) feature introduced in [KubeVirt 1.2](https://github.com/kubevirt/kubevirt/releases/tag/v1.2.0) now allows for specific changes made to a running VM to propagate to the VMI without a restart. This design proposal covers extending this support to cover changes made to the referenced instance type or preference of a running VM.

## Motivation

The initial implementation of the VM Rollout Strategy feature does not support instance types or preferences with their use blocked by the VM validation webhook. Given the ability to now plug additional vCPUs (as sockets) and memory into a running VMI it would be beneficial to VM owners to extend this support to also plug additional resources introduced when switching to a new instance type and/or preference.

## Goals

- Allow for suitable changes introduced by a new instance type and/or preference to be propagated to the running VMI

## Non Goals

- Support for attributes of an instance type or preference not already supported by the `liveUpdate` VM Rollout Strategy feature. For example GPUs, Host Devices, NUMA, disk buses etc
- Moving the `revisionName` of an `InstancetypeMatcher` or `PreferenceMatcher` to the `Status` field of the VM. The work to extract this field into the `Status` of a VM will continue through a previously documented [issue](https://github.com/kubevirt/kubevirt/issues/10145) and [PR](https://github.com/kubevirt/kubevirt/pull/10229).

## User Stories

- As a VM Owner with the `liveUpdate` VM Rollout Strategy feature enabled and configured in my environment I want changes made when switching to a new instance type or preference to be propagated to the running VMI where possible

## Repos

- kubevirt/kubevirt

# Design

The ultimate goal of this design is to provide the same user experience when updating a VM in a `liveUpdate` VM Rollout Strategy enabled environment both with and without the use of instance types or preferences. As such the existing behaviour and logic of the `liveUpdate` `vmRolloutStrategy` feature is utilized with minimal changes made to accommodate instance types and preferences.

No new feature gates or core KubeVirt configurables will be introduced as part of this work.

## virt-api

At present [requests to update a VM referencing an instance type when VM Rollout Strategy is set to `LiveUpdate` are rejected](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/pkg/virt-api/webhooks/validating-webhook/admitters/vms-admitter.go#L435-L441), this limitation will be removed.

## virt-controller (VM)

When `LiveUpdate` is configured as the VM Rollout Strategy and only `LiveUpdate` supported attributes of an instance type such as socket count or guest visible memory change then the same hot plug logic the traditional `LiveUpdate` workflow will be used to hot plug these additional resources.

This will be achieved by expanding any referenced instance type and preference of a VM before the original `LiveUpdate` logic runs within the VM controller. This should ultimately provide the same behaviour as an owner updating a VM without a referenced instance type or preference.

[Defaulting of `maxGuest` and `maxSockets` will also be moved out of the VM controller](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/pkg/virt-controller/watch/vm.go#L1847-L1877) and into the VMI mutation webhook. This defaulting would previously happen within the VM controller while building the initial VMI before a referenced instance type and/or preference were also applied. Moving this defaulting to the mutation webhook means we can now also provide `maxGuest` and `maxSockets` through an instance type.

## instancetype.kubevirt.io/v1beta1

New optional [`maxSocket`](https://kubevirt.io/user-guide/operations/cpu_hotplug/#optional-set-maximum-sockets-or-hotplug-ratio) and [`maxGuest`](https://kubevirt.io/user-guide/operations/memory_hotplug/) attributes will be added to the instance type CRDs modelling the associated attributes of a VM.

```go
type CPUInstancetype struct {
[..]
	// MaxSockets specifies the maximum amount of sockets that can be hotplugged
	// +optional
	MaxSockets uint32 `json:"maxSockets,omitempty"`
}

```
```go
type MemoryInstancetype struct {
[..]
	// MaxGuest allows to specify the maximum amount of memory which is visible inside the Guest OS.
	// The delta between MaxGuest and Guest is the amount of memory that can be hot(un)plugged.
	// +optional
	MaxGuest *resource.Quantity `json:"maxGuest,omitempty"`
}
```

## virtctl

New `vm` subcommands will be added to `virtctl` to update the instance type or preference of a VM.

```
$ # $kind is optional and defaults to virtualmachineclusterinstancetype
$ virtctl vm update-instance-type $vm $kind/$new-instance-type

$ # $kind is optional and  defaults to virtualmachineclusterpreference
$ virtctl vm update-preference $vm $kind/$new-preference
```

This subcommands will provide an optional `--wait/-w` switch that will wait until either a `RestartRequired` condition appears on the VM *or* for the `liveUpdate` process to complete and for all resources to be plugged into the running VMI.

## API Examples

```yaml
./cluster-up/kubectl.sh apply -k https://github.com/kubevirt/common-instancetypes

./cluster-up/kubectl.sh patch kv/kubevirt -n kubevirt --type merge -p '{"spec":{"workloadUpdateStrategy":{"workloadUpdateMethods":["LiveMigrate"]},"configuration":{"developerConfiguration":{"featureGates": ["VMLiveUpdateFeatures"]},"vmRolloutStrategy": "LiveUpdate", "liveUpdateConfiguration": {"maxGuest": "8Gi"}}}}'

./cluster-up/kubectl.sh apply -f - <<EOF
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: alpine
spec:
  instancetype:
   name: u1.nano
  runStrategy: Always
  template:
    metadata:
      creationTimestamp: null
    spec:
      domain:
        devices:
          interfaces:
          - masquerade: {}
            name: default
        machine:
        resources: {}
      networks:
      - name: default
        pod: {}
      volumes:
      - containerDisk:
          image: registry:5000/kubevirt/alpine-container-disk-demo:devel
        name: alpine
      - cloudInitNoCloud:
          userData: |
            #!/bin/sh
            echo 'printed from cloud-init userdata'
        name: cloudinitdisk
EOF

./cluster-up/kubectl.sh wait vms/alpine --for=condition=Ready

./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.spec.domain.memory.guest}'="512Mi"
./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestAtBoot}'="512Mi"
./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestRequested}'="512Mi"
./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestCurrent}'="512Mi"

./cluster-up/kubectl.sh patch vms/alpine --type merge -p '{"spec":{"instancetype":{"name":"u1.micro","revisionName":""}}}'

./cluster-up/kubectl.sh wait vmis/alpine --for=condition=HotMemoryChange

while ! ./cluster-up/kubectl.sh get virtualmachineinstancemigrations -l kubevirt.io/vmi-name=alpine; do
  sleep 1
done 

./cluster-up/kubectl.sh wait virtualmachineinstancemigrations -l kubevirt.io/vmi-name=alpine --for=jsonpath='{.status.phase}'="Succeeded"

./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestAtBoot}'="512Mi"
./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestRequested}'="1Gi"
./cluster-up/kubectl.sh wait vmis/alpine --for=jsonpath='{.status.memory.guestCurrent}'="1Gi"
```

## Scalability

N/A

## Update/Rollback Compatibility

The original ControllerRevisions are retained after the VM has been updated allowing for a user to revert to the previous state of the VM if needed. These ControllerRevisions can be manually removed by the owner of the VM if this is not required. If they are not they will be removed as normal with the deletion of the VirtualMachine.

## Functional Testing Approach

New functional tests will be written to ensure behaviour is consistent between an instance type and non-instance type `liveUpdate` VM Rollout Strategy enabled update of a VM.

# Implementation Phases

TODO
