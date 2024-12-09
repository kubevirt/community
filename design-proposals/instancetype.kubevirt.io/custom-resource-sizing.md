# Overview

Instance types and preferences provide a way to define common VirtualMachine workload resource sizing and preferences.

The common-instancetypes project within KubeVirt currently provides a set of standardized instance types and preferences that can be deployed by `virt-operator` alongside all of the other core KubeVirt components.

# Motivation

While the resources provided by the common-instancetypes project are a great starting point they are generalized and could never possibly cover all workload resourcing and preference use cases.

As such this design proposal seeks to make it as easy as possible for both cluster admins and users to create their own customized versions of these commonly deployed resources specific to their workloads. Additionally the instance type API itself will be made more flexible by making vCPU and memory resource sizing optional within an instance type, allowing users to provide their own sizing within their VirtualMachine while still retaining the remaining benefits of using an instance type.

## Goals

* Allow users to provide their own vCPU and memory resource sizing while using commonly deployed instance types  
* Make it as easy as possible for cluster admins and users to create and consume their own versions of commonly deployed instance types and preferences

## Non Goals

## User Stories

* As a VM owner I want to provide my own vCPU and memory resource sizing while using commonly deployed instance types  
* As a VM owner I want to create customized versions of commonly deployed instance types and preferences specific to my workload use case  
* As a cluster admin I want users to be able to provide their own vCPU and memory resource sizing when using commonly deployed instance types  
* As a cluster admin I want to create customized versions of commonly deployed instance types and preferences specific to my workload use case

## Repos

* `kubevirt/kubevirt`  
* `kubevirt/common-instancetypes`

# Design

## Support for user-supplied vCPU and memory instance type resource sizing

The original design of the instance type API and CRDs required vCPU and memory resource sizing to always be provided when creating an instance type.

This obviously constrained users to the resource sizing already baked into commonly deployed instance types within their environment.

It would be more flexible if the resource sizing within an instance type was optional and could instead be provided by the users VirtualMachine.

### Make `spec.cpu.guest` and `spec.memory.guest` optional

The first step here is to make both the quantity of vCPUs and memory provided by an instance type optional. This behavioral change will require a new version of the API, `instancetype.kubevirt.io/v1beta2`.

### Introduce `spec.template.spec.domain.cpu.guest`

This aims to align the VirtualMachine and instance type APIs in terms of expressing the number of vCPUs to provide to the eventual guest.

Until now the VM could only express a quantity of vCPU resource through resource requests or through the guest visible CPU topology with the number of vCPUs being the total of sockets\*cores\*threads.

By introducing a standalone `guest` attribute that expresses a number of vCPUs we can also allow this value to be used by any associated preference when determining the guest visible CPU topology.

Without the use of a preference, the vCPUs provided by this field would be made visible to the guest OS as single core sockets by default.

This field will be the only supported way of providing a number of vCPUs for the VM when also using an instance type without a set amount of vCPUs.

The field will also be mutually exclusive with the original `spec.template.spec.domain.cpu.{cores,sockets,threads}` fields.

### Provide a `custom` size for each common-instancetypes instance type class

A new `custom` size would be introduced to all but the universal instance type class without vCPUs or memory defined, allowing users to provide their own values through the VirtualMachine spec.

A `custom` size obviously doesn't make sense for the universal instance type class which only provides vCPUs and memory.

## Allow for the easy creation of customization common-instancetype resources

### Introduce `virtctl customize`

This new `virtctl` sub-command would allow users to make subtle changes to existing resources within the cluster, outputting a fresh version they can either save locally or apply directly.

Initial support for instance types would be provided and allow the amount of vCPU and RAM supplied by an instance type to be changed. Support for more switches could then follow depending on feedback from users.

By default the command will remove the all `instancetype.kubevirt.io` labels and annotations from the resulting resource to avoid confusion.

The command will also support switching between the cluster and namespaced kinds of instance types and preferences. By default the command will look for cluster wide resources and output namespaced resources.

For example, a user could provide the name of a cluster wide deployed `common-instancetype` such as `u1.medium` and replace it with a namespaced version specific to their workload use case within a namespace they control.

## API Examples

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
[..]
spec:
  instancetype:
    name: cx1.custom
  template:
    spec:
      domain:
        cpu:
          guest: 4
        memory:
          guest: 4Gi
---
apiVersion: instancetype.kubevirt.io/v1beta2
kind: VirtualMachineClusterInstancetype
metadata:
  annotations:
    instancetype.kubevirt.io/description: |-
      The CX Series provides exclusive compute resources for compute
      intensive applications.

      *CX* is the abbreviation of "Compute Exclusive".

      The exclusive resources are given to the compute threads of the
      VM. In order to ensure this, some additional cores (depending
      on the number of disks and NICs) will be requested to offload
      the IO threading from cores dedicated to the workload.
      In addition, in this series, the NUMA topology of the used
      cores is provided to the VM.
    instancetype.kubevirt.io/displayName: Compute Exclusive
  labels:
    instancetype.kubevirt.io/class: compute.exclusive
    instancetype.kubevirt.io/cpu: custom
    instancetype.kubevirt.io/dedicatedCPUPlacement: "true"
    instancetype.kubevirt.io/hugepages: "true"
    instancetype.kubevirt.io/icon-pf: pficon-registry
    instancetype.kubevirt.io/isolateEmulatorThread: "true"
    instancetype.kubevirt.io/memory: custom
    instancetype.kubevirt.io/numa: "true"
    instancetype.kubevirt.io/vendor: kubevirt.io
    instancetype.kubevirt.io/version: "1"
    instancetype.kubevirt.io/size: custom
    instancetype.kubevirt.io/common-instancetypes-version: v1.2.0
  name: cx1.custom
spec:
  cpu:
    dedicatedCPUPlacement: true
    isolateEmulatorThread: true
    numa:
      guestMappingPassthrough: {}
  ioThreadsPolicy: auto
  memory:
    hugepages:
      pageSize: 2Mi
```

```shell
$ virtctl customize u1.medium --name u1.mysize --vcpus 4 --memory 4Gi | kubectl apply -f -
$ virtctl create vm --instancetype virtualmachineinstancetype/u1.mysize [..] | kubectl apply -f
```

## Scalability

This should hopefully reduce the common-instancetypes footprint by removing some more extreme sizes from the generated resources.

## Update/Rollback Compatibility

All existing `VirtualMachines` using instance types should continue to work.

## Functional Testing Approach

As with every API-related change, extensive functional and unit tests will be written to assert the above behavior, but there isn't anything overly special to cover here.

# Implementation Phases

* Introduce `instancetype.kubevirt.io/v1beta2`
* Make `spec.cpu.guest` and `spec.memory.guest` optional
* Introduce `spec.template.spec.domain.cpu.guest`
* Introduce `virtctl customize`
* Introduce `custom` sizes of `common-instancetype` resources