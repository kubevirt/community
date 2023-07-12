# Overview

`VirtualMachines` referencing instance types or preferences have
`ControllerRevisions` created containing a point in time copy of the referenced
resource. See the user guide for more on this:

https://kubevirt.io/user-guide/virtual_machines/instancetypes/#versioning

While this allows KubeVirt to always produce the same `VirtualMachineInstance`
at run time regardless of changes to the original resource it does currently
mean that KubeVirt has to support older versions of the
`instancetype.kubevirt.io` CRDs for the lifetime of these `VirtualMachines`.

This proposal outlines a solution to this by upgrading older
`ControllerRevisions` during a `VirtualMachine`  resync by the
`vm-controller`.

## Motivation

In order to drop support for older deprecated versions of the
`instancetype.kubevirt.io` API and CRDs we need to ensure the following
conditions are met:

* All stored objects in `etcd` are migrated to the latest version

* All stashed objects in `ControllerRevisions` are migrated to the latest
  version

The former is taken care by KubeVirt's existing recommendation to use the
`kube-storage-version-migrator` tool with KubeVirt >= `v1.0.0` to handle the
migration of stored objects from `kubevirt.io/v1alpha3` to `kubevirt.io/v1`.

However without any way of migrating stashed `instancetype.kubevirt.io` objects
within `ControllerRevisions` we can never drop support for older deprecated
versions.

## Goals

* All `instancetype.kubevirt.io` stashed objects within `ControllerRevisions`
  should be migrated to the latest supported version allowing for support of
  older deprecated versions to eventually be removed.

## Non Goals

* Currently a `ControllerRevision` is captured for each `VirtualMachine` and
  combination or instance type or preference. The deduplication of these
  `ControllerRevisions` is saved for follow up work but might be an additional
  step when migrating to a new `ControllerRevision` using labels to identify
  existing `ControllerRevisions` containing the same version and importantly
  generation of an object.

## User Stories

* As a developer I want deprecated versions of `instancetype.kubevirt.io` to
  be eventually removed from KubeVirt, reducing the on-going maintenance effort
  required for future versions of KubeVirt.

* As a user I want my existing `VirtualMachines` and running
  `VirtualMachineInstances` to be unaffected by the migration and the eventual
  removal of older versions of `instancetype.kubevirt.io`.

* As a user I want my existing `VirtualMachineSnapshots` to be unaffected by the
  migration and the eventual removal of older versions of
  `instancetype.kubevirt.io`.

## Repos

- https://github.com/kubevirt/kubevirt

## Design

On resync of a `VirtualMachine` the `vm-controller` will look for any instance
type or preference `ControllerRevisions` associated with the `VirtualMachine`.
This lookup is done using the `{Instancetype,Preference}Matchers` held within
the `VirtualMachineSpec` but might change shortly to the `Status` of the
`VirtualMachine` via
[kubevirt/kubevirt/pull/10229](https://github.com/kubevirt/kubevirt/pull/10229).

If a `ControllerRevision` is found it is checked for the
`instancetype.kubevirt.io/object-version` label, introduced by
[kubevirt/kubevirt/pull/9932](https://github.com/kubevirt/kubevirt/pull/9932)
with KubeVirt v1.0.0.

If the label is found the value will be checked against a new constant within
the API definition tracking the latest version of the API.

When the label is not found an upgrade will be attempted. At present this might
not be required if the object contained within the `ControllerRevision` is
already `v1beta1` however this will also recreate the `ControllerRevision`
ensuring it is now labelled and not upgraded in future resyncs of the
`VirtualMachine`.

For older objects existing compatibility code will be used to generate the
latest version of the object before this is stashed in a newly created
`ControllerRevision` with the `VirtualMachine` patched to reference this.

## Scalability

TBD

## Update/Rollback Compatibility

Once the initial `ControllerRevision` is deleted there will be no rollback
possible to the original object.

## Functional Testing Approach

* Tests should populate an environment with `ControllerRevisoins` containing all
  of the currently supported versions (`v1alpha{1,2}`, `v1beta1` etc..) and
  migrate these to the latest version.

## Implementation Phases


## Alternatives

### Introduction of standalone CRDs and a controller to handle upgrades of these ControllerRevisions

See [kubevirt/community/pull/226](https://github.com/kubevirt/community/pull/226)
for more details on this approach that was previously discussed.

### Stashing the version of the associated objects within the Status of a VirtualMachine

As with
[kubevirt/kubevirt/pull/10229](https://github.com/kubevirt/kubevirt/pull/10229)
we could extend the `VirtualMachine` `Status` to also include details of the
version of objects stashed within the referenced instance type and preference
`ControllerRevisions` to avoid looking up each time in order to check labels.


