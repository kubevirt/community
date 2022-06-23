
# Overview
Add ability to find best match for CPU Model dynamically after a vmi is scheduled to a node.

The story is about calculating non-obsolete CPU Model 
that is supported by enough nodes ( * ) that respect
VMI's nodeAffinity and nodeSelectors.

( * ) Number that is high enough for the VMI's migratability
      but not too high to avoid using obsolete CPU Model that 
      is more likely to be widely supported.

    
This will allow users to choose CPU Model configuration in
`KubevirtCR` or in`vmi.spec.domain.CPU.model` that will result
dynamic calculation of CPU Model that is relatively
new and migratable in the cluster.

That can be useful when the user's cluster is heterogeneous.

## Motivation
In KubeVirt, the default CPU model is `host-model` 
which is highly coupled to the initial node of a VMI.

For `host-model` VMIs the CPU model is being determined by the first node on which the VMI 
lands on. This means that the initial node a VMI runs on significantly limit the 
possible destinations for future migrations.

If a VMI with `host-model` CPU model is scheduled to a node with a unique `host-model-required-feature`
in heterogeneous cluster it won't be able to migrate and this will probably **SURPRISE** the user.

An alternative approach would be to set the CPU Model to be a hardcoded model that is widely supported
in the cluster.  
This would solve the unique `host-model-required-feature` problem because hard coded models have
a defined set of CPU features.  
The problem with that approach is that obsolete CPU Models are more likely to be widely supported
and **nodes that don't support such CPU Model will be ignored** also the cluster might change
as nodes are being added and removed from the cluster.

It isn't a trivial task for a user to calculate optimal CPU Model that will be relatively new and
won't prevent VMIs from being live migrated.

Even if a user has this information he would have to set the nodeAffinity himself and as mentioned
this might cause him to ignore some nodes that don't support this optimal CPU Model. 

Therefore, some users might benefit from such configuration options.

## Goals
- Dynamically choose a CPU Model that is relatively new and will allow users with heterogeneous 
cluster to live migrate VMIs.

- Avoid Ignoring some of the nodes in a cluster as it can cause issues with resource 
allocation and availability. it could lead to over-utilization of other nodes and potential performance issues. 


## Non Goals
Not changing the default form in which we schedule virtual machines.

## Definition of Users
Every user that is interested in migration of VMIs in heterogeneous cluster.

## User Stories
- As a user with a long living heterogeneous cluster, I'm **SURPRISED** to encounter `host-model` migration 
are always pending and i don't know what can i do to migrate these VMIs.

- As a user that use KubeVirt in heterogeneous cluster i set the CPU model to be
`Opteron_G1` because it is supported in all the cluster nodes and live migration is 
important to my use cases, but I realized that `Opteron_G1` CPU model does not include 
some of the features and capabilities of newer CPU models, such as support for virtualization 
extensions like Intel VT-x or AMD-V, which are required for running certain types 
of my workloads.

- As a user that use KubeVirt in heterogeneous cluster i set the CPU model to be
`Skylake-Client-IBRS` because it relatively new and performance is important to 
my workloads, but I realized that my VMIs always land on the same set 
of nodes  because some of the nodes don't support `Skylake-Client-IBRS`.

Configuration option that calculate best match for CPU Model dynamically after a vmi is scheduled 
to a node can help these users.

## Repos
https://github.com/kubevirt/kubevirt

# Design
- We will hold a `map[string]string` from CPU Model to launch year ,this heuristic will allow
us choosing relatively new CPU Model when using CPU Model matcher.

- When using the CPU Model matcher a VMI's CPU Model will be calculated **AFTER** virt-launcher is
scheduled to a node.

**Calculation** : First we will get all the potential nodes that respect the VMI's node affinity, 
node selectors and has the same vendor as the initial node.  
Then for each CPU Model that has newer launch year then `MinimalLaunchYear` that is supported 
on the initial node we will calculate how many potential nodes support it.  
We will define a `ThresholdSupport` and we will choose the CPU Model that is
supported by at least `ThresholdSupport` of the cluster's node with 
the newest launch year.


This can be done through the `virt-controller` before the handoff to `virt-hanlder` . 


## API Examples
VMI:
```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
...
spec:
  domain:
    cpu:
      model: best-match-model
...
status:
...
  prefferedModel: <best_match_model>
...
```
KubevirtCR:
```
    apiVersion: kubevirt.io/v1
    kind: KubeVirt
    metadata:
      name: kubevirt
      namespace: kubevirt
    spec:
      ...
      configuration:
        cpuModel: best-match-model
    ...
```


## Scalability

The work can be done by virt-controller because virt-controller 
already has shared informer that watch & list nodes.  
The calculation will be done only for the initial node of a vmi
after virt-launcher is scheduled.

We will use the shared informer's store ,therefore there aren't any API calls involved.

## Update/Rollback Compatibility
(does this impact update compatibility and how)

Shouldn't impact previous versions.

## Functional Testing Approach
Can be tested with functional tests and unit tests

Also this will allow us to use `best-match-model` configuration in the functional tests
and make our functional tests more heterogeneous friendly.

# Implementation Phases
- Add k8s.io/component-helpers to vendor and sync
to be able to use Match function to verify
that nodes match vmis nodeSelectors and
NodeAffinity.
- Add ability to find best match for cpuModel dynamically 
with `DynamicCpuModelMatcher` and use that model after a VMI
lands on the initial node
- Add `prefferedModel` field to `vmi.status`
, `prefferedModel` will represent the best match for cpuModel
on the initial node where virt-launcher landed.
- Add nodeSelector to the bestMatch model on live migration