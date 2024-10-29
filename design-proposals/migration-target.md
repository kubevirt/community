# Overview

We are getting asks from multiple cluster admin that would like - in exceptionel cases - to explicitly specify the "destination" of the VM when doing Live migration.
While this may be less important in a cloud-native environment,
we get this ask from many users coming from other virtualization solutions, where this is a common practice.
The same result can already be achieved today with a few steps, this is only about simplifying it with a single direct API on the single `VirtualMachineInstanceMigration` without the need to alter a VM spec.

As a VM owner I can already constrain my [VM so that it is restricted to run on particular node](https://kubevirt.io/user-guide/compute/node_assignment/) instead of letting the k8s scheduler finding the best node by itself.
With the same mechanism a k8s user can already [constrain a generic Pod so that it is restricted to run on particular node](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/).
This proposal is aiming to expose the same capability to cluster admins, in a fully declarative way without the need to mess up the spec of the existing VM and revert at the end, when dealing with live migrations.

## Motivation
In the ideal cloud native design, the scheduler is supposed to be always able to correctly identify
the best node to run a pod (so the target pod for the VMI after the live-migration) on.
In the real world, we still see specific use cases where the flexibility do explicitly and directly define the target node for a live migration is a relevant nice-to-have:
- Experienced admins are used to control where their critical workloads are move to 
  > I as an admin, notice that a VM with guaranteed resources is having issues (I watched the cpu iowait metric). In order to resolve the performance issue and keep my user happy, I as admin want to move the VM, without interruption, to a node which is currently underutilized - and will make the user's vm perform better.
- Workload balancing solution doesn't always work as expected
  > I have configured my cluster with the descheduler and a load aware scheduler (trimaran), thus by default, my VMs will be regularly descheduled if utilization is not balanced, and trimaran will ensure that my VMs will be scheduled to underutilized nodes. Often this is working, however, in exceptional cases, i.e. if the load changes too quickly, or only 1 VM is suffering, and I want to avoid that all Vms on the cluster are moved, I need - for exception - a tool to move one VM, once to deal with this exceptional situation.
- Troubleshooting a node
- Validating a new node migrating there a specific VM
  > I, as a battle tested admin, ordered a new node, because the old one broken, got the HW team to install it in the dc, ask the net team to wire it up, installed it myself, and included it in the cluster - tainted.
  I drink a coffee, as this was a lot of work.
  Now I want to move one of my own VMs over to it to do some sanity testing.
  No other Vm on my 50 node cluster should be impacted. I just want my very own vm to move over and see if all behaves well.
  I've seen enough issues with misbehaving hardware, misbehaving live migrations, cpu firmware issues, broken storage etc. this is my sanity test

> [!IMPORTANT]
> Directly selecting named nodes as destinations is not assumed to be a default tool for balancing workloads or all the use-cases above. It's instead just a convenient tool for exceptional situations and one-offs to ensure that and admin can quickly react to emergencies, and spikes.
This proposal is part of a larger scheduling enhancement picture.
Cluster admins are also looking for, for instance [descheduler integration](https://github.com/kubevirt/community/pull/2580), [load aware scheduling](https://github.com/kubevirt/user-guide/pull/621).
This proposal is just an additional piece of the puzzle but, again, is the tool for exceptions and the corner cases, not the norm.

Such a capability is expected from traditional virtualization solutions but, with certain limitations, is still pretty common across the most popular cloud providers (at least when using dedicated and not shared nodes).
- For instance on Amazon EC2 the user can already live-migrate a `Dedicated Instance` from a `Dedicated Host` to another `Dedicated Host` explicitly choosing it from the EC2 console, see: https://repost.aws/knowledge-center/migrate-dedicated-different-host
- also on Google Cloud Platform Compute Engine the user can easily and directly live-migrate a VM from a `sole-tenancy` node to another one via CLI or REST API, see: https://cloud.google.com/compute/docs/nodes/manually-live-migrate#gcloud
- Project Harvester, an HCI solution built on top of Kubevirt is also offering the capability of live migrating a VM to a named node, see: https://docs.harvesterhci.io/v1.3/vm/live-migration/#starting-a-migration
  Project Harvester approach is although different from this proposal, its implications are analyzed in section [Project Harvester approach](#project-harvester-approach)

On the technical side something like this can already be indirectly achieved playing with node labels and affinity but nodeSelector and affinity are going to be defined as VM properties that are going to stay while here we are focusing just on setting the desired target of a one-off migration attempt without any future side effect on the VM.
The motivation is to better define a boundary between what is an absolute and long-lasting property of a VM (like affinity) with what is just an optional property of the single migration attempt.
This could also be relevant in terms of personas: we could have the VM owner/developer that is going to specify long-lasting affinity for a VM that is part of an application composed by different VMs and pods and a cluster admin/operator that needs to temporary override that for maintenance reasons.
On the other side the VM owner is not required/supposed to be aware of node names. 

## Goals
- A user allowed to trigger a live-migration of a VM and list the nodes in the cluster is able to rely on a simple and direct API to try to live migrate a VM to a specific node (or a node within a set of nodes identified by adding node affinity constraints).
- The live migration then can successfully complete or fail for various reasons exactly as it can succeed of fail today for other reasons.
- The target node that is explicitly required for the actual live migration attempt should not influence future live migrations or the placement in case the VM is restarted. For long-lasting placement, nodeSelectors or affinity/anti-affinity rules directly set on the VM spec are the only way to go.
- The constraints directly added on the one-off migration can only complement and limit constraints already defined on the VM object (pure AND logic).

## Non Goals
- this proposal is not defining a custom scheduler plugin nor suggesting to alter how the default k8s scheduler works with `nodeName`, `nodeSelector` and `affinity/anti-affinity` rules. See https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ for the relevant documentation

## Definition of Users
- VM owner: the user who owns a VM in his namespace on a Kubernetes cluster with KubeVirt
- Cluster-admin: the administrator of the cluster

## User Stories
- As a cluster admin I want - in exceptional but critical cases - to be able to try to live-migrate a VM to specific node (or node within a set of nodes) for various possible reasons such as:
  - I just added to the cluster a new powerful node and I want to migrate a selected VM there without trying more than once according to scheduler decisions
  - I'm not using any automatic workload rebalancing mechanism and I periodically want to manually rebalance my cluster according to my observations (see fon instance: https://github.com/kubernetes-sigs/descheduler/issues/225 )
  - Foreseeing a peak in application load (e.g. new product announcement), I'd like to balance in advance my cluster according to my expectation and not to current observations
  - During a planned maintenance window, I'm planning to drain more than one node in a sequence, so I want to be sure that the VM is going to land on a node that is not going to be drained in a near future (needing then a second migration) and being not interested in cordoning it also for other pods
  - I just added a new node and I want to validate it trying to live migrate a specific VM there
> [!NOTE]
> technically all of this can be already achieved manipulating the node affinity rules on the VM object, but as a cluster admin I want to keep a clear boundary between what is a long-lasting setting for a VM, defined by the VM owner, and what is single shot requirement for a one-off migration
- As a VM owner I don't want to see my VM object getting amended by another user just for maintenance reasons

## Repos
- https://github.com/kubevirt/kubevirt

# Design
## Proposed design
We are going to add a new optional `addedNodeSelectorTerm` stanza of type `*k8sv1.NodeSelectorTerm` on the `VirtualMachineInstanceMigration` object.
We are not going to alter by any mean the `spec` stanza of the VM or the VMI objects so future migrations or the node placement after a restart of the VM are not going to be affected by a `addedNodeSelectorTerm` set on a specific `VirtualMachineInstanceMigration` object.
If the target pod fails to be started, the `VirtualMachineInstanceMigration` object will be marked as failed as it can already happen today for other reasons.
The reason of the eventual failure will be reported back as it gets reported back today when a migration fails due to other reasons. 

## Why addedNodeSelectorTerm and not simply addedNodeSelector

According to the [k8s APIs](https://github.com/kubernetes/api/blob/71385f038c1097af36f3d2f68b415860b866c1f8/core/v1/types.go#L3355-L3363), a `nodeSelector` is a list of `NodeSelectorTerms` and
*it represents the OR of the selectors represented by the node selector terms*.
This means that a Pod can be scheduled onto a node if just one of the specified `NodeSelectorTerms` can be satisfied (terms are ORed).
This means that if a catch all `NodeSelectorTerm` is added in addition to existing `NodeSelectorTerms` already defined at VM level, it will completely defeat and bypass the constraints defined by the VM owner on the VM object while this proposal is only about being able to restrict the set of valid target nodes for a migration adding additional constraints (`pure AND logic`).
In k8s APIs, `NodeSelectorTerm` are [ORed](https://github.com/kubernetes/api/blob/71385f038c1097af36f3d2f68b415860b866c1f8/core/v1/types.go#L3360) while `NodeSelectorRequirements` within a single `NodeSelectorTerm` are [ANDed](https://github.com/kubernetes/api/blob/71385f038c1097af36f3d2f68b415860b866c1f8/core/v1/types.go#L3366).
This proposal is only about exposing **pure AND logic** to limit the set of candidate nodes for a live migration still respecting what is specified on the VM object so exposing the `NodeSelector` API is not an option.
Exposing a single `NodeSelectorTerm` on the `VMIM` object and adding all of the `NodeSelectorRequirements` defined there to all of the `NodeSelectorTerms` already defined on the VM object is instead a viable solution to achieve *pure AND logic*.

### Why to simply `NodeSelector map[string]string`
On Pod spec we also have `NodeSelector map[string]string` which is used in **AND** with `NodeAffinity` rules so from this point of view its a viable option.
On the other side `pod.spec.nodeSelector` is only matching labels and the predefined `kubernetes.io/hostname` [label is not guaranteed to be reliable](https://kubernetes.io/docs/reference/node/node-labels/#preset-labels).
`NodeSelectorTerm` offers more options, and in particular:
- the capability of matching `metadata.name` field that cannot be altered and its more reliable than `kubernetes.io/hostname` label.
- `NotIn` and `DoesNotExist` operators allowing tp define a node anti-affinity behavior as an alternative to node taints.

### How to propagate the additional constraint (the name of a target node for instance) to the target virt-launcher pod

If `addedNodeSelectorTerm` is defined on the VMIM object, all of the additional `NodeSelectorRequirements` defined there will be appended to all of the existing required `NodeSelectorTerms` defined in eventually set by the VM owner on the VM spec (`spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution`). 
From a logical perspective, this is a pure **AND**:

Nodes need to satisfy all the `NodeSelectorRequirements` defined in the `addedNodeSelectorTerm` stanza on `VirtualMachineInstanceMigration` object **AND** the preexisting `NodeAffinity` constraints defined by the VM owner on the VM spec (and taints and tolerations).
`addedNodeSelectorTerm` will adopt the well known `nodeSelectorTerm` grammar from k8s.
It's a pretty flexible and standard API. It can address:
- the simplest use case (migrating to a named nome identified by its node name):
  ```yaml
  addedNodeSelectorTerm:
    - matchFields:
      - key: metadata.name
        operator: In
        values:
          - <nodeName>
  ```
- more complex scenario like migrating to a scheduler chosen host within a group of hosts satisfying an additional requirement:
```yaml
addedNodeSelectorTerm:
  - matchExpressions:
    - key: disktype
      operator: In
      values:
        - ssd
```

From a coding perspective the additional logic required on the migration controller is pretty simple.
Pseudocode:
```go
	if migration.Spec.AddedNodeSelectorTerm != nil {
		if templatePod.Spec.Affinity.NodeAffinity == nil {
			templatePod.Spec.Affinity.NodeAffinity = &k8sv1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &k8sv1.NodeSelector{
					NodeSelectorTerms: []k8sv1.NodeSelectorTerm{*migration.Spec.AddedNodeSelectorTerm},
				},
			}
		} else if templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
			templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &k8sv1.NodeSelector{
				NodeSelectorTerms: []k8sv1.NodeSelectorTerm{*migration.Spec.AddedNodeSelectorTerm},
			}
		} else {
			for i, _ := range templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions = append(
					templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions,
					migration.Spec.AddedNodeSelectorTerm.MatchExpressions...)
				templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchFields = append(
					templatePod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchFields,
					migration.Spec.AddedNodeSelectorTerm.MatchFields...)
			}
		}
	}
```

## Alternative designs
During the review of this proposal alternative approached got debated.

### Amending node affinity rules on VM objects and propagating them as for LiveUpdate rollout strategy
One of the main reason behind this proposal is for improving the UX making it simpler and better defining boundaries between what is long-term placement requirement and what should simply be tried for this specific migration attempt.
According to:
https://kubevirt.io/user-guide/compute/node_assignment/#live-update
changes to a VM's node selector or affinities for a VM with LiveUpdate rollout strategy are now dynamically propagated to the VMI.

This means that, only for VMs with LiveUpdate rollout strategy, we can already force the target for a live migration with something like:
- set a (temporary?) nodeSelector/affinity on the VM
- wait for it to be propagated to the VMI due to LiveUpdate rollout strategy
- trigger a live migration with existing APIs (no need for any code change)
- wait for the migration to complete
- (eventually) remove the (temporary?) nodeSelector to let the VM be freely migrate to any node in the future

Such a flow can already be implemented today with a pipeline or directly from a client like `virtctl` without any backend change.
The drawback of that strategy is that we should tolerate having the spec of the VM amended twice with an unclear boundary about what was asked by the VM owner for long-lasting application specific reasons and what is required by a maintenance operator just for a specific migration attempt.

### Project Harvester approach
Harvester is exposing a stand-alone API server as the interface for its dashboard/UI.
The `migrate` method of `vmActionHandler` handler in this API server is accepting a `nodeName` parameter.
If `nodeName` is not empty, [the Kubevirt VMI object for the relevant VM is amended on the fly](https://github.com/harvester/harvester/blob/3bba1d6dcc17589fe9607aff38bea7614065b8be/pkg/api/vm/handler.go#L417-L439) setting/overriding a `nodeSelector` for a label matching the hostname just before creating an opaque (meaning not aware of the `nodeName` value) `VirtualMachineInstanceMigration` object.
`vmActionHandler` on the API server is also exposing a `findMigratableNodes` method exposing a list of viable nodes according node affinity rules on the given VM.
So, once the user selected a VM to be migrated, the UI is able to fetch a list of candidate nodes proposing them to the user that can select one of them. The `migrate` method on the API Server so can be called passing an explicit `nodeName` as a parameter.

Although this approach is working, we see some limits:
- it's implicitly prone to race conditions: with `LiveUpdate` rollout strategy for instance, another KubeVirt controller could reconcile the `VMI` with the `VM` before the `VirtualMachineInstanceMigration` got processed by the KubeVirt' migration controller resulting in the VM getting migrated to a different host
- having already a declarative API to request a live migration (`VirtualMachineInstanceMigration` CRD), it looks by far more natural and safe extending it with an additional declarative parameter so that the existing migration controller can properly consume it instead of building an imperative flow, backed by an API server, on top of it

### Exposing spec.nodeName and directly setting it to the target pod bypassing the k8s scheduler
An alternative naive approach would simply expose `nodeName` string on the `VirtualMachineInstanceMigration` API.
If the `nodeName` field is not empty, the migration controller will explicitly set `nodeName` on the virt-launcher pod that is going to be used as the target endpoint for the live migration.
If the `nodeName` field is not empty, the k8s scheduler will ignore the Pod that is going to be used as the target for the migration and the kubelet on the named node will directly try to place the Pod on that node.

Although simple, this approach is no-go for various reasons:
- the cluster admin can easily bypass/break useful or even potenatilly critical affinity/anti-affinity rules set by the VM owner for application specific needs (e.g. two VMs of an application level cluster spread over two different nodes for HA reasons)
- taints with `NoSchedule` and `PreferNoSchedule` effect are also going to be ignored with a potentially unexpected behaviour
- it will break/bypass also (Kubevirt application-aware-quota)[https://github.com/kubevirt/application-aware-quota]

### Exposing also addedTolerations to let the target pod tolerate something that was not originally tolerated by the VM
Injecting additional tolerations just as the result of a migration attempt could be an interesting option for emergency use case where bypassing a taint could be a recovery option.
Still this will enlarge the set of candidate nodes tolerating something that was not planned to be tolerated by the VM owner.
Then, even if directly added on the VM object, the additional toleration could be left there without compromising the ability to live migrate again in the future so it can be simply handled there if needed.

## Specifying a nodeSelector or affinity on an ad hoc migration policy
Migration policies provides a mechanism to let the cluster admin bind migration configurations to Virtual Machines so it could potentially look as a good candidate for this proposal.
But in practice a `MigrationPolicy` is not really a viable option to solve this:
in this proposal we are talking about configuring a `nodeSelector` for a one-off migration attempt; on the other side a `migrationPolicy` is actually designed to be matched to VMs using [NamespaceSelector and/or VirtualMachineInstanceSelector LabelSelectors](https://github.com/kubevirt/api/blob/51298a07198ee887ffcbf16a0b3ffb6e2fe07e9b/migrations/v1alpha1/types.go#L58-L63).
This means that the LabelSelector used to match the selected VM before the migration are still going to match it after the one-off migration attempt and since we'd like to target down to a named node, if the migration to that node was successful, the VM will become not migratable until the MigrationPolicy is removed/amended or the labels on the VM are altered to avoid matching that MigrationPolicy a second time.
So, if we decide to set a `nodeSelector` on `MigrationPolicy` CRD, the cluster admin would have to define a MigrationPolicy, trigger the live-migration, wait for the migration controller and then remove the MigrationPolicy to prevent future unwanted side effects.
So this would be a 3 steps imperative approach with potential concurrency risks without real benefits in terms of user experience.

## API Examples
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstanceMigration
metadata:
  name: migration-job
spec:
  vmiName: vmi-fedora
  addedNodeSelectorTerm:
    - matchFields:
      - key: metadata.name
        operator: In
        values:
          - <nodeName>
```

the API description should clearly emphasize that `addedNodeSelectorTerm` stanza is optional, and we recommend to not set it to let the scheduler find the best node (if trying to migrate to a specific named node is not strictly needed as a one-off migration).
Something like:
```go
// AddedNodeSelectorTerm is applied additionally to the NodeAffinity specified on the VM.
// The scheduler will automatically attempt a reasonable migration, addition constraints
// on the one-off migration are required only in special cases.
// In order to be valid migration targets, Nodes need to satisfy existing NodeAffinity as defined on the VM.
// AND the expressions on this added NodeSelectorTerm.
// AddedNodeSelectorTerm is empty by default (all Nodes match).
// AddedNodeSelectorTerm can only restrict the set of Nodes that are valid target for the migration.
// When multiple nodeSelectorTerms are specified in nodeAffinity types,
// then the Pod can be scheduled onto a node if one of the specified terms can be satisfied (terms are ORed).
// When multiple expressions are specified in a single nodeSelectorTerms,
// then the Pod can be scheduled onto a node only if all the expressions are satisfied (expressions are ANDed).
// To obtain the expected result (restrict the set of Nodes that are valid target for the migration),
// all the expressions specified here are going to be added to all the NodeSelectorTerms defined on the VM.
// +optional
AddedNodeSelectorTerm *k8sv1.NodeSelectorTerm `json:"addedNodeSelectorTerm,omitempty"`
```

## Scalability
Forcing additional node affinity constraints on `VirtualMachineInstanceMigration` could potentially lead to unbalanced node placement across the nodes.
But the same result can be already achieved today specifying a `nodeSelector` or `affinity` and `anti-affinity` rules on a VM. Nothing really new on this regard.
We assume that selecting nodes as destinations is not assumed to be a default tool for balancing workloads but just a tool for exceptional situations and one-offs.

## Update/Rollback Compatibility
`addedNodeSelectorTerm` on `VirtualMachineInstanceMigration` will be only an optional field so no impact in terms of update compatibility.

## Functional Testing Approach
- positive test 1: a VirtualMachineInstanceMigration with an explict addedNodeSelectorTerm pointing to a node able to accommodate the VM should succeed with the VM migrating to the named node
- negative test 1 a VirtualMachineInstanceMigration with an explict addedNodeSelector pointing to a node able to accommodate the VM but not matching a nodeSelector already present on the VM should fail
- negative test 2: a VirtualMachineInstanceMigration with an explict addedNodeSelector should fail if the required node doesn't exist
- negative test 3: a VirtualMachineInstanceMigration with an explict addedNodeSelector should fail if the VM is already running on the requested node
- negative test 4: a VirtualMachineInstanceMigration with an explict addedNodeSelector should fail if the selected target node is not able to accommodate the additional pod for virt-launcher

# Implementation Phases
A really close attempt was already tried in the past with https://github.com/kubevirt/kubevirt/pull/10712 but the PR got some pushbacks because it was not clear why a new API for one-off migration is needed.
We give here a better explanation why this one-off migration destination request is necessary.
Once this design proposal is agreed-on, a similar PR should be reopened, refined, and we should implement functional tests.
