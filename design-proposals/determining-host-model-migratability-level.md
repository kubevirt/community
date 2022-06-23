
# Overview
Add a label to determine the host-model migratability level of each node 
and use this label to configure the cluster to prefer\require node with
high host-model migratability level when scheduling vm for the first time.
## Motivation
In KubeVirt, the default CPU model is `host-model` .

For host-model VMs the CPU model is being determined by the first node on which the VM 
lands on. This means that the initial node a VM runs on significantly limit the 
possible destinations for future migrations.

If a vm with host-model CPU model is scheduled to a node with a unique `host-model-required-feature`
in heterogeneous cluster it won't be able to migrate and this will probably **SURPRISE** the user.

It isn't a trivial task for a user to calculate the set of nodes that 
a vm with `host-model` can migrate to (for each node) especially when the cluster change 
and nodes are being added\removed to\from the cluster.

Even if a user has this information he would have to set nodeAffinity himself to prefer/require
nodes only for the first time he schedules a vm.

Also, some users\developers might want to use our testing framework in heterogeneous cluster,
and we usually never check if a host-model migration is even possible before we migrate vms in tests.

Therefore, some users might benefit from such label and configuration options.

## Goals
- Add a label to each schedulable node in the cluster with a value of
the relative amount of nodes that a vm with `host-model` CPU model can
migrate to when the node is the initial node of the vm.


- Allow users to configure KubeVirtCR to add node affinity for host-model 
migratability level to virt-launcher when scheduling a vmi for the first time.

## Non Goals
Not changing the default form in which we schedule virtual machines.

## Definition of Users
Every user that is interested of migration of vms in heterogeneous cluster.

## User Stories
- As a user with a long living heterogeneous cluster, I'm **SURPRISED** to encounter host-model migration failures and
i don't know how to prevent the failures.


- As a user that use KubeVirt's tests in heterogeneous cluster i don't know why the tests are flacky:

    In KubeVirt-ci the cluster is homogeneous, so we aren't exposed to test failures that
occur because of a host-model migration failures.

  For instance Openshift-Virtualization project run our functional tests in heterogeneous
clusters this issue cause flakiness, and it is hard to follow.

  
Such configuration option that prefer/require host-model migratible nodes can solve this.

## Repos
https://github.com/kubevirt/kubevirt

# Design
- List and calculate the value of the relative amount of nodes that a vm with host-model
  CPU model can be scheduled to for each node  when adding\removing a node from the cluster.

  This can be done through the `node-controller` which is 
part of `virt-controller`. 


- Add an optional fields to `kubevirt.spec.configuration.migrations` to determine if we want our vm's to prefer\require 
nodes with at least `<some-numerical-number>` value of the new `kubevirt/host-model-migratability-level`
label.

## API Examples
- add the following label to a node that allows host-model migration to 50% of the
  cluster's schedulable nodes.
```
kubevirt/host-model-migratability-level: 50
```

- add fields to KubeVirtCR:
`kubevirt.spec.configuration.migrations.requireHostModelMigratabilityLevel:<some-number> `
`kubevirt.spec.configuration.migrations.preferedHostModelMigratabilityLevel.level:<some-number> `
`kubevirt.spec.configuration.migrations.preferedHostModelMigratabilityLevel.weight:<some-number> ` 
(the default weight will be the lowest)


## Scalability

The work is being done by only one Pod of virt-controller to avoid 
adding node's informer for each virt-handler and load work on them
every node removal\addition.

Complexity:



Each iteration the controller will:

**(1) If this is the first iteration or this information isn't up-to-date:**

The controller will calculate the amount of nodes that a vm with `host-model`
can migrate to.

We would have to iterate through every permutation of a pair of nodes and through all their features
and then update the value of the label to be:

`number_of_potential_destination_nodes / total_number_of_nodes`

for each node

the Complexity of this operation is O(n²m).

n- number of nodes

m- number of features

**(2) If a new node is added to the cluster:**

The controller would have to iterate every node in the cluster and 
check if a host-model migration is valid to the new node and if it is
we would have to update the label of the source node to be:

`x / y` ->  (`x / y` *  `y / y+1` )  + `1 / y + 1`

otherwise:

`x / y` ->  (`x / y` *  `y / y+1` )

the Complexity of this operation is O(nm).

**(3) If a node is being removed from the cluster:**

The controller would have to iterate every node in the cluster and
check if a host-model migration was valid to the node that is being delted and if it is
we would have to update the label of the source node to be:

`x / y` ->  (`x / y` *  `y / y-1` )  - `1 / y - 1`

otherwise:

`x / y` ->  (`x / y` *  `y / y-1` )

the Complexity of this operation is O(nm).

## Update/Rollback Compatibility
(does this impact update compatibility and how)

Shouldn't impact previous versions.

## Functional Testing Approach
This will allow us to require\require `host-model` migratible nodes in the functional tests
and make our functional tests more heterogeneous friendly.

# Implementation Phases
First PR:
- Add the label to each node through the `node-controller`.

- Add new fields to KubeVirtCR and implement the desired behavior.

Second PR:

- Configure KubeVirtCR before running functional tests with migrations.

KubeVirtCR configuration should add node affinity to virt-launcher
when scheduling a vmi for the first time.

something like:
```
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
          - key: kubevirt/host-model-migratability-level
            operator: Gt
            values:
            - <some-numerical-number>
```
