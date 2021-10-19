# Overview
KubeVirt allows [Live Migrations](https://kubevirt.io/user-guide/operations/live_migration/) of Virtual Machine workloads.
Today migration settings can be configurable only on the cluster-wide scope by editing [KubevirtCR's spec](https://kubevirt.io/api-reference/master/definitions.html#_v1_kubevirtspec)
or more specifically [MigrationConfiguration](https://kubevirt.io/api-reference/master/definitions.html#_v1_migrationconfiguration)
CRD. 

Several aspects (although not all) of migration behaviour that can be customized are:
- Bandwidth
- Auto-convergence
- Post/Pre-copy
- Max number of parallel migrations
- Timeout

This design proposal asks to generalize the concept of defining migration configurations so it would be
possible to apply different configurations to specified groups of VMs. These are meant to be called "**Migration
Policies**".

## Motivation
Today there is no way of applying different configurations for different groups of VMs. Such capability can be useful
for a lot of different use cases on which there is a need to differentiate between different workloads. Differentiation
of different configurations could be needed because different workloads are considered to be in different priorities,
security segregation, workloads with different requirements, help to converge workloads which aren't migration-friendly,
and many other reasons (see user story section below).

## Goals
The goal is to have a mechanism through which it will be possible to apply all relevant cluster-wide configurations
to a custom specified group of VM. Also, at any given time every VM should deterministically know which migration policy
it needs to obey to.

Another goal is to allow the cluster administrator to define these policies without the participation of the VM's owner
(as written in "Definition of Users" below).

## Definition of Users
This feature's most obvious user is the cluster administrator.

VM owners/creators, by contrast, should not know or care about their migration configuration at most cases.
Moreover, the design should take into account that one of the goals is to allow the cluster administrator to dynamically
define and change existing policies with as little participation as possible from the VM owner.

## User Stories
As an admin I would like to have different migration policies for different groups of VMs.
* By "migration policies" I mean controlling one or more of the following migration aspects:
    * Bandwidth
    * Scheduling
    * Storage
* By “different groups of VMs” I could mean:
    * All VMs in the same namespace
    * All VMs that have certain label defined
    * All VMs that have certain label with certain value defined
* As an admin I know that some workloads are in higher priority than others and therefore need different 
  migration policies.
* As an admin I know that some workloads are not migration-friendly and demand careful configurations in order
  for migrations to converge.
* As an admin I want to have the ability to temporarily tune migrations for a specific event such as draining
  a node or performing an upgrade.
* As an admin I want that the migration policy would be as transparent as possible for the VM owner / creator.

## Repos
Kubevirt/kubevirt

# Design
The design will be similar to Kubernetes' [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
which is a similar concept. Therefore, a new MigrationPolicy CRD will be introduced. This CRD's spec will contain all
relevant fields from KubevirtCR's [MigrationConfiguration](https://kubevirt.io/api-reference/master/definitions.html#_v1_migrationconfiguration)
(i.e. which are not inherently cluster-wide configurations). In the future, additional
configurations (or status related fields) that inherently relate to non-global context could be added,
thus the configurations on NetworkPolicy and KubevirtCR's MigrationConfigurations may diverge to some degree.

In addition, the new CRD's spec will include the following ways of specifying the groups of VMIs on which
to apply the policy:
* _(By VMI names / regular expressions)?_
* By VMI's labels
* By namespace's name
* By namespace's labels

All of these methods can be combined, for example a policy can require both VMI labels and namespace labels in
order to match.

It is possible that a multiple policies apply to the same VMI. In such cases, the precedence is in the
same order as the bullets above. It is not allowed to define two policies with the exact same selectors.

## API Examples
### Migration Configurations
Currently, MigrationPolicy's spec will only include the currently relevant configurations from KubevirtCR's
MicrationConfiguration:
```yaml
kind: MigrationPolicy
  spec:
    allowAutoConverge: true
    bandwidthPerMigration: 217Ki
    completionTimeoutPerGiB: 23
    allowPostCopy: false
    disableTLS: false
```

In the future more configurations can be added, either being cloned from KubevirtCR's MigrationConfiguration or
by adding a configuration that is not suited for cluster-wide context.

All above fields are optional. When omitted, the behaviour that will apply is the currently defined behaviour in
KubevirtCR's MigrationConfiguration. This way, KubevirtCR will serve as a configurable set of defaults for both
VMs that are not bounded to any MigrationPolicy and VMs that are bounded to a MigrationPolicy that does not
define all the configurations.

### Matching Policies to VMs

Next in the spec are the selectors that define the group of VM on which to apply the policy. The options to do so
are the following.

**This policy applies for the VMs in a given namespace name(s):**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    - namespaceSelector:
        matchName: my-namespace
```
In the future it's possible to also support matching by regular expressions.

**This policy applies for the VMs in namespaces that have all the required labels:**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    - namespaceSelector:
        matchLabels:
          hpc-workloads: true       # Matches a key and a value 
          xyz-workloads-type: ""    # Matches the key only, any value applies
```

**This policy applies for the VMs that have all the required labels:**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    - virtualMachineInstanceSelector:
        matchLabels:
          workload-type: db       # Matches a key and a value 
          operating-system: ""    # Matches the key only, any value applies
```

**It is also possible to combine the previous two:**

```yaml
kind: MigrationPolicy
  spec:
  selectors:
    - namespaceSelector:
        matchLabels:
          hpc-workloads: true
          xyz-workloads-type: ""
    - virtualMachineInstanceSelector:
        matchLabels:
          workload-type: db
          operating-system: ""
```

_NOTE_: It's possible to add `matchName` to `virtualMachineInstanceSelector` as well to match to VMs by name (or regular expressions).

### Full Manifest:

```yaml
kind: MigrationPolicy
metadata:
  name: my-awesome-policy
spec:
  # Migration Configuration
  allowAutoConverge: true
  bandwidthPerMigration: 217Ki
  completionTimeoutPerGiB: 23
  allowPostCopy: false
  disableTLS: false
  
  # Matching to VMs
  selectors:
    - namespaceSelector:
        matchLabels:
          hpc-workloads: true
          xyz-workloads-type: ""
    - virtualMachineInstanceSelector:
        matchLabels:
          workload-type: db
          operating-system: ""
```

## Functional Testing Approach
Open for discussion, might be tricky.

# Implementation Phases
[This PR](https://github.com/kubevirt/kubevirt/pull/6399) is a very basic implementation of migration policies (POC)
to serve as a reference for the final implementation and design discussion.
