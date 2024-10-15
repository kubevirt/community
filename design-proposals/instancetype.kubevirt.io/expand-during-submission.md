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
and eventual restore of a given `VirtualMachine` to happen.

## Goals

* Provide simple configurables at the cluster and `VirtualMachine` level to expand
  instance types and preferences

## Non Goals

* The default behavior will not change as part of this work

## User Stories

* As a cluster owner I want to configure the expansion of all new
  `VirtualMachines` referencing instance types and preferences
* As a VirtualMachine owner I want to optionally enable or disable the expansion
  my VirtualMachine referencing an instance type or preference

## Repos

* kubevirt/kubevirt

## Design

A new `KubeVirt` configurable will be introduced to have `VirtualMachines`
referencing instance types or preferences expanded.

Additional `VirtualMachine` configurables will also be introduced to allow a
specific newly submitted VirtualMachine referencing instance types or
preferences to be expanded.

Both sets of configurables will default to `reference`.

The `VirtualMachine` configurable will override the cluster-wide `KubeVirt`
configurable in situations where they are not equal.

When requested the `VirtualMachine` controller will expand any referenced
instance type or preference when the `VirtualMachine` is reconciled.

Note that this will only occur when no revision has already been collected for
the resource and as such wouldn't impact existing `VirtualMachines` unless a
user explicitly requests to expand by removing the existing reference to the
revision.

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
Users will need to explicitly request to not expand their `VirtualMachines`
referencing instance types or preferences to avoid this.

## Alternatives

### Eventual expansion of *all* VMs

A possible alternative to this design proposal is to have the VirtualMachine
controller eventually expand all VirtualMachines referencing instance types or
preferences regardless of a revision already being stored.

This behaviour could be controlled by the same cluster-wide and VirtualMachine
specific configurables as suggested by this proposal.

This could however result in drastic changes to existing VirtualMachine objects
without user interaction.

We would also still need to retain webhook validation logic to ensure that a
given `VirtualMachine` using an instance type or preference is valid during
submission.

This would also put additional load on the controller if the cluster-wide
configurable was enabled in an environment with a large number of existing
`VirtualMachines` using instance types and preferences.

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

We could alter the default `policy` to `expand` and in doing so deprecate the
revision `reference` behaviour for eventual removal from the project ahead of
`instancetype.kubevirt.io` finally making it to `v1`.

### Keep existing behaviour

Ultimately we can also decide not to implement the core proposal or any of the
above alternatives and continue to support the original revision based flows.
Supporting users and third-party integrators with better documentation and
tooling for VirtualMachines referencing instance types or preferences.

## API Examples

The default `policy` will be `reference`.

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
      policy: expand
    preference:
      policy: expand
```

Additionally a VM owner can explicitly request `expand` by setting the policy
directly on the `VirtualMachine`:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    name: foo
    policy: expand
  preference:
    name: bar
    policy: expand
```

Likewise users can explicitly request the `reference` policy to counter a
different default being defined in the `KubeVirt` `CR`:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    name: foo
    policy: reference
```

Users can also expand an existing `VirtualMachine` by removing any referenced
revision from a matcher to invoke the expand behaviour, for example starting
with the following:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    name: foo
    revisionName: bar
```

A user can explicitly ask for the associated instance type to be expanded into
the `VirtualMachine` by removing `revisionName` and setting `policy` to
`expand`:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  instancetype:
    name: foo
    policy: expand
```

This will in turn expand the instance type into the `VirtualMachine` and remove
the matcher:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: cirros
spec:
  [..]
```

```go
// KubeVirtConfiguration holds all kubevirt configurations
type KubeVirtConfiguration struct {
[..]
    // Instancetype configuration
    Instancetype *InstancetypeConfiguration `json:"instancetype,omitempty"`

    // Preference configuration
    Preference *PreferenceConfiguration `json:"preference,omitempty"`
}

type ExpansionPolicy string

const (
    Expand    ExpansionPolicy = "expand"
    Reference ExpansionPolicy = "reference"
)

type InstancetypeConfiguration struct {
    Policy *ExpansionPolicy `json:"policy,omitempty"`
}

type PreferenceConfiguration struct {
    Policy *ExpansionPolicy `json:"policy,omitempty"`
}

[..]

// InstancetypeMatcher references a instancetype that is used to fill fields in the VMI template.
type InstancetypeMatcher struct {
    [..]
    Policy *ExpansionPolicy `json:"policy,omitempty"`
}

[..]

// PreferenceMatcher references a set of preference that is used to fill fields in the VMI template.
type PreferenceMatcher struct {
    [..]
    Policy *ExpansionPolicy `json:"policy,omitempty"`
}
```

## Scalability

The resulting mutation of the `VirtualMachine` with this proposal will
cause an additional reconciliation loop to trigger for the VM.

Work should be carried out to ensure that the substantial mutation of the
`VirtualMachine` during submission doesn't negatively impact the control plane.

## Update/Rollback Compatibility

Existing `VirtualMachines` referencing instance types or preferences will not be
changed as a result of this behaviour.

There will be no ability to automatically rollback new `VirtualMachines` once
they have their instance type or preference expanded by this new functionality.

Users will also be unable to resize their `VirtualMachines` by making a singular
choice of instance type in the future.

## Functional Testing Approach

TBD

## Implementation Phases

TBD