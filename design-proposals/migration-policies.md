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

## Non Goals
* Allowing users to transparently choose migration policies for their workloads
* Preventing misconfigurations regarding overlapping selectors

## Definition of Users
This feature's most obvious user is the cluster administrator.

VM owners/creators, by contrast, should not know or care about their migration configuration in most cases.
Moreover, the design should take into account that one of the goals is to allow the cluster administrator to dynamically
define and change existing policies with as little participation as possible from the VM owner.

Another goal is to let the admin the possibility to expose different policies to the users in a way he can then
use. For example, an admin can set two policies for fast/slow migration that will apply to any VM with a
`migration-type: slow` or `migration-type: fast` and this way a user can apply different policies by defining
the relevant label.

## User Stories
As an admin I would like to have different migration policies for different groups of VMs.
* By "migration policies" I mean controlling one or more of the following migration aspects:
    * Sensitivity to overall migration time
    * Migration progress timeout
    * Determining if and when post-copy / auto-converge are safe.
    * Control number of migration threads, dirty rate (in the future)
* By “different groups of VMs” I could mean:
    * All VMs that have certain label defined
    * All VMs that have certain label with certain value defined
    * All VMs that belong to a namespace with certain label key/value
* As an admin I know that some workloads are in higher priority than others and therefore need different 
  migration policies.
* As an admin I know that some workloads are not migration-friendly and demand careful configurations in order
  for migrations to converge.
* As an admin I want to have the ability to temporarily tune migrations for a specific event such as draining
  a node or performing an upgrade.
* As an admin I want that the migration policy would be as transparent as possible for the VM owner / creator.
* As a user I want to be able to tell admins the workload characteristics of my VMs so that admins can create
  policies and/or namespaces with the desired characteristics.

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
* By VMI's labels
* By namespace's labels

All of these methods can be combined, for example a policy can require both VMI labels and namespace labels in
order to match.

## Policies' Precedence

It is possible that a multiple policies apply to the same VMI. In such cases, the precedence is in the
same order as the bullets above (VMI labels first, then namespace labels). It is not allowed to define
two policies with the exact same selectors.

If multiple policies apply to the same VMI:
* The most detailed policy will be applied, that is, the policy with the highest number of matching labels
* If multiple policies match to a VMI with the same number of matching labels, the policies will be sorted by the 
  lexicographic order of the matching labels keys. The first one in this order will be applied.
  
### Example

For example, let's imagine a VMI with the following labels:
* size: small
* os: fedora
* gpu: nvidia

And let's say the namespace to which the VMI belongs contains the following labels:
* priority: high
* bandwidth: medium
* hpc-workload: true

The following policies are listed by their precedence (high to low):
1) VMI labels: `{size: small, gpu: nvidia}`, Namespace labels: `{priority:high, bandwidth: medium}`
    * Matching labels: 4, First key in lexicographic order: `bandwidth`.
2) VMI labels: `{size: small, gpu: nvidia}`, Namespace labels: `{priority:high, hpc-workload:true}`
    * Matching labels: 4, First key in lexicographic order: `gpu`.
3) VMI labels: `{size: small, gpu: nvidia}`, Namespace labels: `{priority:high}`
    * Matching labels: 3, First key in lexicographic order: `gpu`.
4) VMI labels: `{size: small}`, Namespace labels: `{priority:high, hpc-workload:true}`
    * Matching labels: 3, First key in lexicographic order: `hpc-workload`.
5) VMI labels: `{gpu: nvidia}`, Namespace labels: `{priority:high}`
    * Matching labels: 2, First key in lexicographic order: `gpu`.
6) VMI labels: `{gpu: nvidia}`, Namespace labels: `{}`
    * Matching labels: 1, First key in lexicographic order: `gpu`.
7) VMI labels: `{gpu: intel}`, Namespace labels: `{priority:high}`
    * VMI label does not match - policy cannot be applied.

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

**This policy applies for the VMs in namespaces that have all the required labels:**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    namespaceSelector:
      matchLabels:
        hpc-workloads: true       # Matches a key and a value 
        xyz-workloads-type: ""    # Matches the key only, any value applies
```

**This policy applies for the VMs that have all the required labels:**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    virtualMachineInstanceSelector:
      matchLabels:
        workload-type: db       # Matches a key and a value 
        operating-system: ""    # Matches the key only, any value applies
```

**It is also possible to combine the previous two:**
```yaml
kind: MigrationPolicy
  spec:
  selectors:
    namespaceSelector:
      matchLabels:
        hpc-workloads: true
        xyz-workloads-type: ""
    virtualMachineInstanceSelector:
      matchLabels:
        workload-type: db
        operating-system: ""
```

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
    namespaceSelector:
      matchLabels:
        hpc-workloads: true
        xyz-workloads-type: ""
    virtualMachineInstanceSelector:
      matchLabels:
        workload-type: db
        operating-system: ""
```

## Functional Testing Approach
- Reliably distinguish after a migration if it was pre- or post-copy
- create scenarios where throttled VMIs fail the migration reliably

# Implementation Phases
[This PR](https://github.com/kubevirt/kubevirt/pull/6399) is a very basic implementation of migration policies (POC)
to serve as a reference for the final implementation and design discussion.
