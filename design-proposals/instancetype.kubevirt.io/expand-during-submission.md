# Overview

The initial design of instance types and preferences captured a point in time
revision of the original resource for future use by the `VirtualMachine` to
ensure we always generate the same `VirtualMachineInstance` at runtime.

This additionally allows users to easily switch between classes, sizes and
generations of these resources at some point in the future. However this
flexibility comes at the cost of complexity within KubeVirt itself when handling
more advanced `VirtualMachine` lifecycle flows.

This design proposal aims to set out new alternative and configurable behavior
where these revisions are no longer captured and are instead expanded into the
`VirtualMachine`.

## Motivation

The complexity of managing these point in time revisions of instance types and
preferences has grown over time as more features and versions of the CRDs have
landed.

Some of this complexity is exposed to users and third party integrations
such as back-up tooling or user-interfaces built on top of KubeVirt. This can
take the form of differing entry points for certain functionality such as hot
plug to requiring knowledge of the stored revisions to allow for a valid back up
and eventual restore of a given VirtualMachine to happen.

## Goals

* Provide a simple cluster configurable to control how
  instance types and preferences are referenced from a VirtualMachine

## Non Goals

* The default behavior will not change as part of this work

## User Stories

* As a cluster owner I want to control how instance types and preferences are
  referenced from VirtualMachines within my environment

## Repos

* kubevirt/kubevirt

## Design

A new KubeVirt configurable will be introduced to control how instance types and
preferences are referenced from VirtualMachines.

This configurable will provide an `InstancetypeReferencePolicy` that
encapsulates this behaviour. The following policies will be initially provided:

* `reference` (default) - This is the original reference behaviour of instance
  types and preferences where a ControllerRevision is captured and referenced
  from the VirtualMachine.
* `expand` - This is a new behaviour where any instance type and preferences are
  expanded into the VirtualMachine if a ControllerRevision hasn't already been
  captured and referenced.
* `expandAll` - The same behaviour as expand but regardless of a
  ControllerRevision being captured already.

## Concerns

### Exposing users to API complexity within VirtualMachines

One of the original design goals with instance types and preferences was to
simplify creation by reducing a users exposure to the core
`VirtualMachineInstanceSpec` API and thus their overall decision matrix.

While this proposal doesn't change the ability to simplify creation of
`VirtualMachines` it can result in a fully flattened `VirtualMachine` exposing
all of this complexity at that level once again.

### Breaking declarative management of VirtualMachines using instance types

This expansion behavior will break any declarative management of these
`VirtualMachines` as they will substantially change after initial submission.
VM owners will need to explicitly request to not expand their `VirtualMachines`
referencing instance types or preferences to avoid this.

## Alternatives

### Immutable instance types and preferences

The need to retain point in time revisions of instance types and preferences is
due to the simple fact that in the current implementation these resources are
mutable and can change over time. Thus to ensure we always get the same
`VirtualMachineInstance` at runtime revisions need to be taken and referenced
from the `VirtualMachine`.

We could possibly remove this requirement by making these object immutable and
thus dropping the need capture and reference `ControllerRevisions` from the
`VirtualMachines` at all.

This however still retains the need for additional logic in more complex
`VirtualMachine` lifecycle operations where we need to expand these now
immutable resources in the `VirtualMachine`.

### Expand by default and deprecate revision references

We could alter the default `policy` to `expandAll` and in doing so deprecate the
revision `reference` behaviour for eventual removal from the project ahead of
`instancetype.kubevirt.io` finally making it to `v1`.

### Keep existing behaviour

Ultimately we can also decide not to implement the core proposal or any of the
above alternatives and continue to support the original revision based flows.
Supporting users and third-party integrators with better documentation and
tooling for VirtualMachines referencing instance types or preferences.

## API Examples

The default `policy` will be `reference` and as such there should be no change
in behavior when this configurable is not provided.

A cluster admin can default the policy of all new `VirtualMachines` by setting
`expand` within the `KubeVirt` `CR`:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kv
spec:
  configuration:
    instancetype:
      referencePolicy: expand
```

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example
spec:
  instancetype:
    name: foo
  preference:
    name: bar
```

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example
spec:
  template:
    spec:
      domain:
        cpu:
          sockets: 1
          cores:   1
          threads: 1
```

A cluster admin can also use the `expandAll` policy to have all VirtualMachines
expanded regardless of `revisionNames` already being captured.

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kv
spec:
  configuration:
    instancetype:
      referencePolicy: expandAll
```

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example
spec:
  instancetype:
    name: foo
    revisionName: revision-foo
  preference:
    name: bar
    revisionName: revision-bar
```

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example
spec:
  template:
    spec:
      domain:
        cpu:
          sockets: 1
          cores:   1
          threads: 1
```

```go
// KubeVirtConfiguration holds all kubevirt configurations
type KubeVirtConfiguration struct {
[..]
    // Instancetype configuration
    Instancetype *InstancetypeConfiguration `json:"instancetype,omitempty"`
}

type InstancetypeConfiguration struct {
 // ReferencePolicy defines how an instance type or preference should be referenced by the VM after submission, supported values are:
 // reference (default) - Where a copy of the original object is stashed in a ControllerRevision and referenced by the VM.
 // expand - Where the instance type or preference are expanded into the VM during submission with references removed.
 // +nullable
 // +kubebuilder:validation:Enum=reference;expand;expandAll
 ReferencePolicy *InstancetypeReferencePolicy `json:"referencePolicy,omitempty"`
}

type InstancetypeReferencePolicy string

const (
 // Copy any instance type or preference and reference from the VirtualMachine
 Reference InstancetypeReferencePolicy = "reference"
 // Expand any instance type or preference into VirtualMachines without a revisionName already captured
 Expand InstancetypeReferencePolicy = "expand"
 // Expand any instance type or preferences into all VirtualMachines
 ExpandAll InstancetypeReferencePolicy = "expandAll"
)

```

## Scalability

The resulting mutation of the VirtualMachine with this proposal will
cause an additional reconciliation loop to trigger for the VM.

Work should be carried out to ensure that the substantial mutation of the
VirtualMachine during submission doesn't negatively impact the control plane.

## Update/Rollback Compatibility

There will be no ability to automatically rollback new VirtualMachines once
they have their instance type or preference expanded by this new functionality.

Users will also be unable to resize their VirtualMachines by making a singular
choice of instance type in the future without making further modifications to
their VirtualMachine.

## Functional Testing Approach

TBD

## Implementation Phases

TBD
