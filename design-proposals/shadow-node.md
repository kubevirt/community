# Overview

A reported [Advisory](https://github.com/kubevirt/kubevirt/security/advisories/GHSA-cp96-jpmq-xrr2) is outlining how of our virt-handler component can be abused to escalate local privileges when a node, running virt-handler, is compromised. A flow and more details can be found on reported [issue](https://github.com/kubevirt/kubevirt/issues/9109).

This proposal is outlining mitigation in a form of internal change.

## Motivation

Kubevirt should be secure by default. While mitigation is available, it is not part of Kubevirt.

## Goals

- Prevent a malicious user that has taken over a Kubernetes node where virt-handler daemonset is deployed from exploiting the daemonset's RBAC permissions in order to elevate privileges beyond the node until potentially having full privileged access to the whole cluster.
- Prevent virt-handler from modifying non-kubevirt-owned fields on the node. This includes not being able to change the spec or any label/annotation which are not strictly kubevirt owned.

## Non Goals

Redesign/optimize existing virt-hanlder features to not update/patch the node in the first place.

## Definition of Users

Not user-facing change

## User Stories

Not user-facing change

## Repos

Kubevirt/Kubevirt

# Design Options

## Background
In order to easy review and understand the different options, this section will describe different usage of Node object in the virt-handler.

### Node writers on virt-handler
Kubevirt sets ~170 labels on each node. This is set by the following entities
1. virt-handler [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/cmd/virt-handler/virt-handler.go#L185) the node unschedulable for VMs at the start and the termination (by SIGTERM only) of the virt-handler process to indicate (best-effort) that the node is not ready to handle VMs.

2. virt-handler/heartbeat:
  * [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L96) the node unschedulable for VMs when the stop channel closes (i.e. virt-handler exits).
  * [labels and annotates](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L139) the node, updating NodeSchedulable, CPUManager, KSMEnabledLabel labels and VirtHandlerHeartbeat KSMHandlerManagedAnnotation annotations once per minute.
    * On the other side - Virt-controller's node-ctonroller watches which monitors the heartbeat annotation.

3. Virt-handler/node labeller [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/node-labeller/node_labeller.go#L254) various information which are mostly static in nature: cpu model, hyperv features, cpu timer, sev/-es, RT...
  * This is triggered both by interval (3 min) and on changes of cluster config.

## Solution option #1: ShadowNode CRD
The security issue is being addressed by introducing a new light weight CRD: "ShadowNode" for each Kubernetes node.
Virt-handler will keep read access to nodes but will lose write access. Virt-handler will instead write to the ShadowNodes object, and this will subsequently be selectively copied to the node by a new controller. This controller will be the only entity who writes to the node object.

Pros:
- Since the new controller runs virt-controller context which only runs on the control-plane nodes, the risk of taking over the cluster is reduced.
- A malicious user can't manipulate virt-handler's RBAC permissions to exploit non-kubevirt params/labels.
- Adding a new CRD is a good opportunity to create a real spec for all that info. For instance, the reason we use heartbeat in metadata is that we don't have a place on the k8s node object to properly express information. Now we can.
- Scalability:
  - #shadowNodes will be equivalent #nodes, negligible space overhead

Cons:
- Scalability:
  - Albeit a simple and lightweight CRD (will only contain metadata), adding shadowNodes adds a 1:1 shadowNodes to nodes resources, therefor increasing:
    - the #API-calls kubevirt does - #writes could double sue to writing via the CRD as a proxy. Having said that, if the heartbeat (which is the highest frequency write) is not copied to the node than it would be less (requires re-working node controller to watch for the heartbeat annotation the shadowNode instead).
    - the storage kubevirt takes.
- Adding virt-controller as a "proxy" agent, we now depend on its availability. This could cause an increase in heartbeat timeout in corner cases like upgrade.
- To a degree, it is still possible to manipulate pods, by causing VMs to migrate from other nodes.

### API Examples
```yaml
apiVersion: <apiversion>
kind: ShadowNode
metadata:
  name: <Equivalent to existing name>
  # Labels and Annotations only
spec: {} //Empty, allows further extension
status: {} //Empty, allows further extension
```

### Solution option #2: Creating a REST/GRPC communication between virt-handler and virt-controller
Pass the labels/annotations through a REST/GRPC channel to the virt-controller. REST option is preferred because the server-side is already implemented.

Virt-handler (Server): Instead of directly patching labels/annotation to the node, Virt-handler will maintain a "cache" (NodeInfoMgr) where all the components will "set" their data.
Virt handler can use its existing [web-service](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/cmd/virt-handler/virt-handler.go#L556) to expose one more endpoint - GetNodeInfo. Once reached - this endpoint will return the node data.

Virt-Controller (Client): Add a new web-service client infra that will:
1. run as a go routine and poll the client for updates every one minute.This frequency is chosen because it is the highest of the frequencies (=heartbeat).
2. ask every Virt-handler server separately for the NodeInfo (need to make sure it's possible to address them in a unique manner).
3. filter non-kubevirt owned annotations/labels.
4. patch the node if necessary.

The heartbeat annotation and virt-controller's [Node-controller](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/pkg/virt-controller/watch/node.go) can be removed in favor for the web-service timeout mechanism (if received timeout - set node as unreachable).

Pros:
- Avoid introducing a new CRD to the project.
- Using existing web-server infra on the server side.
- security:
  - the web-service is secure using the cert-manager internal API.
  - Virt-handler does not need to access kube-API at all, and cannot modify unrelated nodes or shadow nodes. Virt-controller does all the risky work
- Scalability:
  - Reduces #API requests (i.e. node patches) to a minimum.

Cons:
- More complex. It's easier and more appealing to use the k8s API, i.e. patching labels/annotations, than to implement and maintain GRPC/REST case logic internally.
- Virt-handler will no longer be able to immediately set the node as unschedulable (in case of virt-handler SIGTERM). However, this use-case is pretty rare, and IMO should be reconsidered.

## Solution option #3: Moving to third party entity
The idea is to eliminate the reason(s) to patch to the node from virt-handler in the first place.

1. Use tools like [Node Feature Discovery](https://kubernetes-sigs.github.io/node-feature-discovery/v0.15/get-started/introduction.html) or utilize virt-handler's [device-manager](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) to publish node features (pending checking whether they don't have the CVE issue as well).
2. Remove writing the unschedulable label in case virt-handler crashes. It is racy, and has no real value.
3. Heartbeat - Remove adding the annotation in favor of watching the readiness of the virt-handler Deamonset/pods.

Pros:
- No need to invent the wheel.
- Re-examine current methods to make sure if we still need them is a good practice. In this case we may find out that patching is no longer needed, and in the long run need to maintain less code.
- This steps can be broken to multiple seemingly unrelated PRs, that are easier to review.

Cons:
- Relying on a third part tools.

## Solution option #4: [Ruled-out] Using status section
Same proxy solution as primary solution where virt-hadler updates the labels/annotations somewhere, and it gets copied to the node by virt-controller, but instead of using a new CRD as the medium - writing the labels/Annotations to the status section of an existing CRD like kubevirt CR.

Pros:
- Avoid introducing a new CRD to the project.

Cons:
- [Deal-breaker] The status section is considered a user facing API section. Dynamically adding/updating 150+ technical labels there may only confuse the user and thus is not advisable.

## Solution option #5: [Ruled-out] Using the virt-handler pod as a medium to pass the labels/annotations
Same proxy solution as primary solution where virt-handler updates the labels/annotations somewhere, and it gets copied to the node by virt-controller, but instead of using a new CRD as the medium - writing the labels/annotations to the virt-handler pods's labels/annotations.
Labels/Annotations that need to pass to the node should get a special prefix (like "nodesync/"), so that regular labels don't get passed to the node by mistake.

Pros:
- Avoid introducing a new CRD to the project.
- A simple and clean approach.

Cons:
- [Deal-breaker] It requires giving the virt-handler RBAC permission to update pods, which is generally considered risky.
- It's only relevant for passing labels/annotations. CRD approach allows for future expansions.

## Solution option #6: [Ruled-out] Fine-tune each virt-handler RBAC to only patch its own node.
Introducing a new controller called token-distributor-controller. It will create a token for each virt-handler pod and deliver it to the node.

Pros:
- Avoid introducing a new CRD to the project.
- A malicious user can't manipulate pods from other nodes to migrate to their node.

Cons:
- A malicious user can still manipulate its own node with non-kubevirt controlled params/labels, then possibly lure control plane components to its node in order to steal their credentials. For example, it can add the control-plane label to its own node, then add other labels to attract components that seek specific functionalities, and take over them.

This solution is not a bad one, but it's more important to address not being able to affect non-kubevirt labels/annotations before doing this (see Follow-up Steps section).

# chosen solution

TBD until further discussion.

# Update/Rollback Compatibility

The following reasoning is relevant for all the proposed solutions.
virt-operator upgrade [order](https://github.com/kubevirt/kubevirt/blob/2e3a2a32d88a2e08c136c051c30754d1b4af423b/pkg/virt-operator/resource/apply/reconcile.go#L526-L641) is:
```
CRDs ==> RBACs ==> Daemonsets (=virt-handler) ==> Controller Deployments (=virt-controller) ==> API Deployments (=virt-API)
```

## Option 1: Ignore/Mask patches to the node in case of "RBAC forbidden" error received
In order to handle backwards compatibility issues and minimize heartbeat timeouts, Virt-handler is going to continue writing to Node object.
At some point during the upgrade, it would start being blocked due to the removal of the patch ability. In order to mitigate this and minimize heartbeat timeout during upgrade, the node patch will start ignoring RBAC "forbidden" error.
Having said the above, this is a specific mention to each upgrade scenario - before and after the upgrade itself:

### During upgrade rollout transient case scenarios:
* New CRD; old RBAC; old virt-handler; old virt-controller; old virt-API:
  * Old RBAC: no upgrade issue
  * Old virt-handler: no upgrade issue
  * Old virt-controllers: no upgrade issue
  * Old Virt-API: no upgrade issue
* New CRD, new RBAC; old virt-handler; old virt-controller; old virt-API:
  * Old virt-handler: during upgrade, the old RBACs are kept as backup until upgrade is done. Hence, patching node should not fail so there is no upgrade issue.
  * Old virt-controllers: no upgrade issue, virt controller RBAC is not changed due to this solution.
  * Old Virt-API: no upgrade issue.
* New CRD; new RBAC; new virt-handler; old virt-controller; old virt-API:
  * new virt-handler: shadowNode is written to but it is not synced to the node yet. However, virt-handler keeps patching the node directly, so there is no upgrade issue.
  * [same as before] Old controllers: no upgrade issue, virt controller RBAC is not changed.
  * [same as before] Old Virt-API: no upgrade issue.
* New CRD; new RBAC; new virt-handler; new virt-controller; old virt-API:
  * New virt-handler: no upgrade issue. virt-handler patches both node and shadow node.
  * New controllers: will start getting shadowNode requests once they are created. no upgrade issue.
  * [same as before] Old Virt-API: no upgrade issue.

### After upgrade rollout:
Virt-handler will keep trying (and failing) to patch the node since the backup RBACs are now removed. Virt-handler will ignore these cases by ignoring RBAC "forbidden" errors. This behavior should be kept until we no longer support upgrades from the current release.

## Option 2: Use a new flag in the kubevirt CR status to signal upgrade is done

Until the upgrade is done, patches to the node will still occur. Once the upgrade is finished (after the temporary validating webhooks are cleared) - set a new status field on virt-operator to signal that you can use the new way.
Doing so will ensure that all new components are installed and ready.

## Functional Testing Approach

Existing functional tests should be enough to cover this change.

# Implementation Phases

This change will be implemented in one phase.

# Follow-up Steps

These steps could be presented in future design proposals:
- Restrict the virt-handler Daemonset to only modify KubeVirt owned CRs (such as virtualmachineinstances) related to its own node.
- Restrict the virt-handler Daemonset from being able to update/patch other nodes.
