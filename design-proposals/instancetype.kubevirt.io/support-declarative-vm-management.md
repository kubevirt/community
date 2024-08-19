# Overview

At present a VirtualMachine using instance types and/or preferences will have runtime derived data such as the name of a ControllerRevision capturing the state of each resource mutated into the core spec during submission.

This breaks declarative management of VirtualMachines as an owner has no way of pre-populating these runtime derived values and will always see changes made to the spec of their VirtualMachine after submission.

## Motivation

This design proposal aims to enable declarative management of VirtualMachines using instance types and/or preferences by using status to track runtime derived data while retaining all existing behavior and lifecycle support of a VirtualMachine using an instance type and/or preference.

## Goals

* No runtime derived instance type or preference data should be mutated into the core VirtualMachine spec during submission  
* Users should still be able to snapshot and restore a VirtualMachine using instance types and/or preferences  
* Users should still be able to switch between instance types potentially invoking vCPU and memory hot plug

## Non Goals

* Deduplication of instance type and preference ControllerRevisions  
* A new instancetype API group version to v1beta2, all of the API changes actually land in the core v1 API group.

## User Stories

* As a VirtualMachine owner I want to declaratively manage my VirtualMachines using instance types and/or preferences  
* As a VirtualMachine owner I want existing VirtualMachine lifecycle features to continue to work such as snapshot and restore  
* As a VirtualMachine owner I want an easy to use mechanism for switching between instance types and preferences

## Repos

* kubevirt/kubevirt

# Design

## `spec.{instancetype,preference}.revisionName` deprecation

The original RevisionName field of the {Instancetype,Preference}Matchers will be deprecated and no longer populated by the VirtualMachine controller.

Any existing or new values will be mirrored into the new `status.{instancetype,preference}` fields discussed below.

As the `{Instancetype,Preference}Matchers` are part of the core `kubevirt.io/v1` API this field should only be dropped if we ever make it to `v2` of this API.

## `spec.{instancetype,preference}.controllerRevisionRef`

A new ControllerRevisionRef field will be introduced. This corrects a wrinkle with the original API and now follows [best practices for reference field naming in k8s](https://github.com/kubernetes/community/blob/f88bbbabe89e01bd8f436c339128154e4032d0d5/contributors/devel/sig-architecture/api-conventions.md\#references-to-related-objects).

This field will continue to allow users and controllers a way for explicitly defining a ControllerRevision to use for a given instance type or preference.

This is required for use cases such as snapshot and restore where the Status of the VirtualMachine is not captured. Thus we need to store these ControllerRevision details in the spec of the VirtualMachine during the snapshot for use in the eventual restore.

The VirtualMachine controller will not populate this field when the VirtualMachine is first seen, thus ending the needless mutation of the object after submission.

## `spec.{instancetype,preference}.InferFromVolume`

The use of inferFromVolume will no longer result in the Matcher being mutated within the spec.

All runtime derived details will now be populated within the status of the VirtualMachine.

As with controllerRevisionRef above these details will however be written into the spec during a snapshot to ensure we accurately restore the VirtualMachine later.

## `status.{instancetypeRef,preferenceRef}`

New types will be introduced under the status of a VirtualMachine to capture the details of each resource.

These details include the `Name,` `Kind` and `ControllerRef` for each resource:

```go
type ControllerRevisionRef struct {
  // Name of the ControllerRevision
  Name string `json:"name"`
}

type InstancetypeRef struct {
  // Name is the name of the instance type
  Name string `json:"name"`

  // Kind specifies the kind of instancetype resource referenced
  Kind string `json:"kind"`

  // ControllerRef specifies the ControllerRevision storing a copy of the object captured
  // when it is first seen by the VirtualMachine controller
  ControllerRevisionRef ControllerRevisionRef `json:"controllerRevisionRef"`
}

type PreferenceRef struct {
  // Name is the name of the preference
  Name string `json:"name"`

  // Kind specifies the kind of preference resource referenced
  Kind string `json:"kind"`

  // ControllerRef specifies the ControllerRevision storing a copy of the object captured
  // when it is first seen by the VirtualMachine controller
  ControllerRevisionRef ControllerRevisionRef `json:"controllerRevisionRef"`
}

type VirtualMachineStatus struct {
[..]
  // InstancetypeRef captures the state of any referenced instance type from the VirtualMachine
  //+nullable
  //+optional
  InstancetypeRef *InstancetypeRef `json:"instancetypeRef,omitempty"`

  // PreferenceRef captures the state of any referenced preference from the VirtualMachine
  //+nullable
  //+optional
  PreferenceRef *PreferenceRef `json:"preferenceRef,omitempty"
}
```

Storing the `Name` and `Kind` is useful in a few ways. Namely when inferring the details of the resource *and* also checking if a user has requested to switch to a new resource by patching the core spec of the VirtualMachine.

## VirtualMachine subresource API to clear `ControllerRevisionRef`

A user should not be able to modify the status of a VirtualMachine.

As such a new `refresh-{instancetype,preference}` subresource API will be provided to allow users to clear both `spec.{instancetype,preference}.controllerRevisionRef` and `status.{instancetypeRef,preferenceRef}.controllerRevisionRef` in order to allow users to move between generations of an instance type or preference object.

A new `virtctl refresh instancetype ${vm}` command should also be created to help users call this subresource API.

## LiveUpdate support

Users will be able to invoke the existing LiveUpdate workflow by patching the `spec.{instancetype,preference}`.

The VirtualMachine controller will attempt to align the status with the spec and in doing so will invoke the required logic to attempt a LiveUpdate that may result in the hot plug of additional supported resources etc.

The same will also be possible when moving between generations of an instance type if the new generation introduces additional supported resources such as vCPUs exposed as sockets or memory.

## API Examples

For a simple VirtualMachine using a VirtualMachineClusterInstancetype named `foo` the following details should end up in status:

```yaml
spec:
  instancetype:
    name: foo
[..]
status:
  instancetypeRef:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

When moving between instance types users will now only need to modify the name (and kind if originally provided) with the VirtualMachine controller then attempting to reconcile the spec with the status of the VirtualMachine:

```yaml
spec:
  instancetype:
    name: foo
[..]
status:
  instancetypeRef:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

A user patches `spec.instancetype.name` with `fooNew`, `foo` is still referenced by `status.instancetypeRef.name`:

```yaml
spec:
  instancetype:
    name: fooNew
[..]
status:
  instancetypeRef:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

Eventually the VirtualMachine controller aligns `status.instancetypeRef.name` with `spec.instancetype.name`:

```yaml
spec:
  instancetype:
    name: fooNew
[..]
status:
  instancetypeRef:
    name: fooNew
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

When `kind` and/or `revisionName` are already populated in existing VirtualMachines these values are mirrored into status by the controller:

```yaml
spec:
  instancetype:
    name: foo
    kind: virtualmachineclusterinstancetype
    revisionName: bar
[..]
status:
  instancetypeRef:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

This will also happen with `controllerRevisionRef` in the spec:

```yaml
spec:
  instancetype:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
[..]
status:
  instancetypeRef:
    name: foo
    kind: virtualmachineclusterinstancetype
    controllerRevisionRef:
      name: bar
```

When using `inferFromVolume` with a new VirtualMachine the Matcher within the spec is no longer mutated with all runtime derived details now stashed in status:

```yaml
spec:
  instancetype:
    inferFromVolume: foo
status:
  instancetypeRef:
    name: derivedName
    kind: derivedKind
    controllerRevisionRef:
      name: derivedRevisionName

```

If a snapshot is taken of this VirtualMachine however the values from status will be written back into the spec alongside `inferFromVolume`:

```yaml
spec:
  instancetype:
    name: derivedName
    kind: derivedKind
    inferFromVolume: foo
    controllerRevisionRef:
      name: derivedRevisionName
```

On restore these values will again be populated in the status of the VirtualMachine by the controller and cleared from the spec:

```yaml
spec:
  instancetype:
    name: derivedName
    kind: derivedKind
    inferFromVolume: foo
    controllerRevisionRef:
      name: derivedRevisionName
status:
  instancetypeRef:
    name: derivedName
    kind: derivedKind
    controllerRevisionRef:
      name: derivedRevisionName
```

Users looking to use `inferFromVolume` once again with a restored VirtualMachine will need to manually clear the `name`, `kind` and `controllerRevisionRef` fields before using the `refresh-{instancetype,preference}` subresource APIs to force the controller to infer these values once again in `status`.

## Scalability

The existing patching of the spec will simply be replaced with similar operations against the status of the VirtualMachine and thus the proposed changes should not add any additional performance penalty at scale. The only exception to this will be on initial upgrade as discussed in the next section.

## Update/Rollback Compatibility

On upgrade the VirtualMachine controller will need to patch the status of any VirtualMachine referencing an instance type and/or preference. This additional operation may slow the initial reconciliation of VirtualMachines in an environment if many such objects exist.

## Functional Testing Approach

Existing functional tests should be extended to assert the contents of status.

New functional tests should ensure that user provided data is copied from the spec into status. These new tests should also assert that the matchers are no longer mutated after submission.

Snapshot and restore functional tests should be extended to assert the new behavior.

InferFromVolume functional tests should be extended to assert the new behavior.

New functional tests should be written covering the behavior of the `refresh-{instancetype,preference}` subresource API.

# Implementation Phases

* Introduce the new API structs and fields
* Implement new ControllerRevision metadata capture logic within the VirtualMachine controller  
* Deprecate `RevisionName`  
* Implement new LiveUpdate logic based around the contents of status  
* Implement the new `refresh-{instancetype,preference}` subresource API and `virtctl` command
