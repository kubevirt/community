# Overview

`VirtualMachines` referencing instance types or preferences have
`ControllerRevisions` created containing a point in time copy of the referenced
resource. See the user guide for more on this:

https://kubevirt.io/user-guide/virtual_machines/instancetypes/#versioning

While this allows KubeVirt to always produce the same `VirtualMachineInstance`
at run time regardless of changes to the original resource it does currently
mean that KubeVirt has to support older versions of the
`instancetype.kubevirt.io` CRDs for the lifetime of these `VirtualMachines`.

This document proposes the introduction of two new controllers to
`virt-controller` to migrate the stashed `instancetype.kubevirt.io` objects
within these `ControllerRevisions` to their latest version, hopefully allowing
support for older deprecated versions to be dropped from the codebase.

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

The following design is heavily inspired by the existing
`kube-storage-version-migrator` project, with both using two controllers,
one to first trigger and then another to facilitate the migration.

This is unlike the recently proposed [mass machine type
transition](https://github.com/kubevirt/community/pull/225) that uses a
standalone `Job` to orchestrate the update of a `VirtualMachines` machine type.
While this is a valid approach for a targeted user invoked workflow the
migration outlined below is more involved and shares logic with KubeVirt itself
when converting these stashed objects.

As such the current suggestion is to mirror the design of
`kube-storage-version-migrator` within `virt-controller` with two additional
controllers. An initial controller to discover stashed objects that need to be
migrated and then trigger their migration with another controller handling the
actual migration.

## `ControllerRevisionUpgradeTrigger` controller

As suggested above this controller will seek out stashed
`instancetype.kubevirt.io` objects to migrate. When found it then creates a
request to migrate the object that is handled by a second
`ControllerRevisionUpgrade` CRD and controller.

```go
// ControllerRevisionUpgradeTrigger
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
type ControllerRevisionUpgradeTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControllerRevisionUpgradeTriggerSpec   `json:"spec,omitempty"`
	Status ControllerRevisionUpgradeTriggerStatus `json:"status,omitempty"`
}

type ControllerRevisionUpgradeTriggerSpec struct {
	// Optional namespace to target specific VirtualMachines
	Namespace string `json:"namespace,omitempty"`

	// Optional label selector to use to target specific VirtualMachines
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}
```

* Watch for `ControllerRevisionUpgradeTrigger` requests
  
* Once triggered check for `VirtualMachines` or `VirtualMachineSnapshots`
  referencing instance type or preference `ControllerRevisions` [1]

* Check if an existing `ControllerRevisionUpgrade` exists for the
  object, skip the following if it does

* Check if the associated `ControllerRevisions` has the latest version listed
  by their `instancetype.kubevirt.io/object-{version}` label

* If not create `ControllerRevisionUpgrade` requests for any
  requiring a migration

* Track all created `ControllerRevisionUpgrade` requests and
  report status

[1] This initial step will be removed in the future once all instance type
`ControllerRevisions` are labelled with at least
`instancetype.kubevirt.io/object-version` making this a simple lookup using said
label.

## `ControllerRevisionUpgrade` controller

Once stashed `instancetype.kubevirt.io` objects have been identified by the
`ControllerRevisionUpgradeTrigger` controller and a
`ControllerRevisionUpgrade` object submitted this controller then
attempts to upgrade the object to the latest version.

```go
// ControllerRevisionUpgrade encapsulates a specific upgrade of a stashed ControllerRevision instance type object to the latest available version
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +genclient
type ControllerRevisionUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControllerRevisionUpgradeSpec `json:"spec,omitempty"`
	Status ControllerRevisionUpgradeStatus `json:"status,omitempty"`
}

type ControllerRevisionUpgradeSpec struct {
	// Name of the ControllerRevision to migrate
	TargetName string `json:"targetName"`
}
```

* Watch for `ControllerRevisionUpgrade` requests

* Extract and convert a stashed `instancetype.kubevirt.io` object within a given
  `ControllerRevision` to the latest version

* Recreate [2] the `ControllerRevision` and store the upgraded object

* Patch the `Owner` object with a reference to the new `ControllerRevision`

* Remove the original `ControllerRevision`

[2] The object within a `ControllerRevisions` is immutable so we have to create
a fresh `ControllerRevision` for the upgraded object.

## Scalability


## Update/Rollback Compatibility

Once the initial `ControllerRevision` is deleted there will be no rollback
possible to the original object.

## Functional Testing Approach

* Tests should populate an environment with `ControllerRevisoins` containing all
  of the currently supported versions (`v1alpha{1,2}`, `v1beta1` etc..) and
  migrate these to the latest version.

## Implementation Phases

* Introduce `ControllerRevisionUpgrade` CRD and controller

* Introduce `ControllerRevisionUpgradeTrigger` CRD and controller

## Alternatives

### Upgrading on VirtualMachine resync

This would rely on all `VirtualMachines` being resync'd periodically or on
upgrade to first check and then upgrade any associated
`instancetype.kubevirt.io` ControllerRevisions.

*TODO* Finish the following section

For

* This would avoid the need to add additional CRDs and controllers to the codebase

* `VirtualMachines` could be labelled to highlight that no upgrade was required *or* the upgrade was complete?

Against

* Difficult to tell when all `VirtualMachines` had been upgraded without additional labels

* `VirtualMachineStatus` would need to be extended to capture failures

* This would add overhead to each resync of a VirtualMachine using
  `instancetype.kubevirt.io` ControllerRevisions

* This would also possibly overload the VirtualMachine controller on upgrade if
  a large number of `VirtualMachines` are using `instancetype.kubevirt.io`
  ControllerRevisions that need to be upgraded