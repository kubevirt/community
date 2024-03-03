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
1. virt-handler [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/cmd/virt-handler/virt-handler.go#L185) the node with `kubevirt.io/schedulable=false` at the start and the termination (by SIGTERM only) of the virt-handler process to indicate (best-effort) that the node is not ready to handle VMs. This is meant to improve user experience with VM scheduling.

2. virt-handler/heartbeat:
  * [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L96) the node with `kubevirt.io/schedulable=false` when the stop channel closes (i.e. virt-handler exits).
  * [labels and annotations](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L139) the node, updating NodeSchedulable, CPUManager, KSMEnabledLabel labels and VirtHandlerHeartbeat KSMHandlerManagedAnnotation annotations once per minute.
    * On the other side - Virt-controller's node-controller watches which monitors the heartbeat annotation.

3. Virt-handler/node labeller [labels](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/node-labeller/node_labeller.go#L254) various information which are mostly static in nature: cpu model, hyperv features, cpu timer, sev/-es, RT...
  * This is triggered both by interval (3 min) and on changes of cluster config. However, as these labels usually don't change—this intervalled reconciliation is usually no-op.

## Solution option #1: Use Validating Admission Policy
Add new [validationAdmissionPolicy and ValidatingAdmissionPolicyBinding](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/) objects to prevent virt-handler from adding/deleting/updating non kubevirt-owned labels/annotations.
kubevirt-owned labels/annotations are defined as follows:
- labels/annotations that contain the string `kubevirt.io`.
- `cpu-manager` label, which is an exception to the above rule.

The validationAdmissionPolicy and ValidatingAdmissionPolicyBinding objects will be configured with the following parameters:
- [matchConstraints](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#getting-started-with-validating-admission-policy): apply only on Nodes; Operations: UPDATE
- predefined [variables](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#variable-composition):
  - `isPartitionedServiceAccount`: true if virt-handler is the user requesting the update.
  - `oldLabels`: the nodes' labels before the update was issued
  - `oldNonKubevirtLabels`: a sublist of the old labels that are not kubevirt owned
  - `newLabels`: the nodes' labels after the update was issued
  - `newNonKubevirtLabels`: a sublist of the new labels that are not kubevirt owned
  - (same set of parameters for annotations)
- [matchConditions](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#matching-requests-matchconditions) rules (only if all are "true" then the validation will be evaluated):
  - `isPartitionedServiceAccount` == true
- [validation](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/#creating-a-validatingadmissionpolicy) rules (written in pseudocode to keep things readable):

  | validation (all rules need to be true in order to pass policy)                                                                    | error message in case of validation failure                                    | Remarks                                                                                                                                                                |
  |-----------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
  | object.spec == oldObject.spec                                                                                                     | virt-handler user cannot modify spec of the nodes                              |                                                                                                                                                                        |
  | object.status == oldObject.status                                                                                                 | virt-handler user cannot modify status of the nodes                            |                                                                                                                                                                        |
  | object.(not labels/annotations/resourceVersion/managedFields) == oldObject.(not labels/annotations/resourceVersion/managedFields) | virt-handler user can only change allowed sub-metadata fields.                 |                                                                                                                                                                        |
  | size(`newNonKubevirtLabels`) == size(`oldNonKubevirtLabels`)                                                                      | virt-handler user cannot add/delete non kubevirt-owned labels                  |                                                                                                                                                                        |
  | for label in `newNonKubevirtLabels`: {(label in `oldNonKubevirtLabels`) AND `newLabels`[label] == `oldLabels`[label]}             | virt-handler user cannot update non kubevirt-owned labels                      |                                                                                                                                                                        |

### Backwards compatibility:
validationAdmissionPolicy [going GA](https://kubernetes.io/blog/2024/03/12/kubernetes-1-30-upcoming-changes/) on k8s 1.30, so it should be OK for deployment on Kubevirt main and next minor release branches.
However, since kubevirt needs to be supported up to 2 k8s versions back, then the validationAdmissionPolicy deployment will be conditioned by whether v1/validationAdmissionPolicy resource is available on the cluster (see best-effort approach on [appendix A](#appendix-a-using-new-k8s-resources-on-kubevirt)).

As the discussion on this behavior goes beyond the scope of this proposal - it is thoroughly discussed on [appendix A](#appendix-a-using-new-k8s-resources-on-kubevirt).

### Next step: Prevent patches from other nodes' virt-handlers
It is possible to expand the rules of the validationAdmissionPolicy to also prohibit virt-handler running on a node from patching other nodes.
This could be defined by setting a new variable `requestHasNodeName` which will be true if request has `userInfo.extra.authentication.kubernetes.io/node-name` field, then adding this validation:
```
!`requestHasNodeName` OR (object.Name == request.nodeName)
```
However, this field is [beta](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/#additional-metadata-in-pod-bound-tokens) on k8s 1.30  and is currently protected by the [ServiceAccountTokenPodNodeInfo](https://github.com/kubernetes/kubernetes/blob/c4bce63d9886e5f1fc00f8c3b5a13ea0d2bdf772/pkg/features/kube_features.go#L753) k8s feature gate.
This field can be added once it is GA.

Pros:
* Avoid adding CRDs or changing the current virt-handler code to fix the CVE issue.
* A k8s built-in feature which is straightforward to maintain and configure.
* validationAdmissionPolicy is [going GA](https://kubernetes.io/blog/2024/03/12/kubernetes-1-30-upcoming-changes/) on k8s 1.30, so it could be used for the next Kubevirt minor release.

Cons:
* Currently, can't add the next step that prevents patches from other nodes' virt-handlers.

## Solution option #2: ShadowNode CRD
The security issue is being addressed by introducing a new light weight CRD: "ShadowNode" for each Kubernetes node.
Virt-handler will keep read access to nodes but will lose write access. Virt-handler will instead write to the ShadowNodes object, and this will subsequently be selectively copied to the node by a new controller. This controller will be the only entity who writes to the node object.

Pros:
- The virt-controller is running only on subset of nodes, therefore, the exposure is not as wide as the virt-handler. Moreover, as virt-controller is advised to run on control-plane (see [node-placement](https://kubevirt.io/user-guide/operations/installation/#restricting-kubevirt-components-node-placement)) - the risk can be further reduced.
- A malicious user can't manipulate virt-handler's RBAC permissions to exploit non-kubevirt params/labels.
- Adding a new CRD is a good opportunity to create a real spec for all that info. For instance, the reason we use heartbeat in metadata is that we don't have a place on the k8s node object to properly express information. Now we can.
- Scalability:
  - #shadowNodes will be equivalent #nodes, negligible space overhead
  - If heartbeat annotation is handled only on the shadowNode and not directly patched to the node as been done now, then the nodes will get much fewer API-calls. Ergo, Node-watching elements will get "bothered" less frequently.

Cons:
- Scalability:
  - Albeit a simple and lightweight CRD (will only contain metadata), adding shadowNodes adds a 1:1 shadowNodes to node resources, therefore increasing:
    - the #API-calls kubevirt does - At first #writes could double due to writing via the CRD as a proxy. However, this should be transiently reduced, as the labels kubevirt patches don't tend to change often (except the heartbeat that as mentioned above can be eliminated). Considering the heartbeat annotation removal from the node patch, then this con can be flagged as negligible.
    - the storage kubevirt takes.
- Adding virt-controller as a "proxy" agent, we now depend on its availability. This could cause an increase in heartbeat timeout in corner cases like upgrade.
- To a degree, it is still possible to manipulate pods by causing VMs to migrate from other nodes. For example, a malicious user can add the control-plane label to its own node, then adding other labels to lure control plane pods to their node.

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

### Solution option #3: Creating a REST/GRPC communication between virt-handler and virt-controller
Pass the labels/annotations through a REST/GRPC channel to the virt-controller. REST option is preferred because the server-side is already implemented.

Virt-handler (Server): Instead of directly patching labels/annotation to the node, Virt-handler will maintain a "cache" (NodeInfoMgr) where all the components will "set" their data.
Virt handler can use its existing [web-service](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/cmd/virt-handler/virt-handler.go#L556) to expose one more endpoint - GetNodeInfo. Once reached - this endpoint will return the node data.

Virt-Controller (Client): Add a new web-service client infra that will:
1. run as a go routine and poll the client for updates every one minute. This frequency is chosen because it is the highest of the frequencies (=heartbeat).
2. ask every Virt-handler server separately for the NodeInfo (need to make sure it's possible to address them in a unique manner).
3. filter non-kubevirt owned annotations/labels.
4. patch the node if necessary.

In case that virt-handler gets a SIGTERM, it will no longer be able to set the node as unschedulable. This will be replaced with readiness and liveness probes and a new controller that will watch the virt-handler Daemonset readiness and set it to unschedulable if necessary.

Notes:
- The heartbeat annotation and virt-controller's [Node-controller](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/pkg/virt-controller/watch/node.go) can be removed in favor using the web-service timeout mechanism (if received timeout - set node as unreachable).

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
- Virt-handler will no longer be able to immediately set the node as unschedulable (in case of virt-handler SIGTERM). However, replacing it with the probes and controllers should provide a sufficient substitute.

## Solution option #4: Moving to third party entity
The idea is to eliminate the reason(s) to patch to the node from virt-handler in the first place.

1. Use tools like [Node Feature Discovery](https://kubernetes-sigs.github.io/node-feature-discovery/v0.15/get-started/introduction.html) (personally preferred) OR utilize virt-handler's [device-manager](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/) to publish node features.
   - How these tools solve the CVE issue:
     - Node Feature Discovery's master Daemonset (that patches the node) only runs on control-plane nodes, so there is less security concern.
     - device-manager is a plugin for kubelet so the patching is actually done by kubelet.
2. `"kubevirt.io/schedulable"` label:
   - Remove writing this label in case virt-handler crashes. It is racy, and has no real value.
   - Move setting it entirely to virt-controller (right now it only sets it to value "false")
3. Heartbeat: Instead of adding the heartbeat TS annotation in order to later monitor its "freshness" from virt-controller's [Node-controller](https://github.com/kubevirt/kubevirt/blob/b102c56f0fcd52feff3ff7a6296737b8e8b99131/pkg/virt-controller/watch/node.go),  simply monitor the readiness of the virt-handler Deamonset/pods.

### utilize device-manager option
Instead of publishing the node's cpu-feature,supported_features,hostCPUModels,etc... as labels on the node, add them as devices on the node:
Example:
From this:
```yaml
apiVersion: v1
kind: Node
metadata:
labels:
    cpu-feature.node.kubevirt.io/abm: "true"
    cpu-feature.node.kubevirt.io/adx: "true"
#...
```
To this:
```yaml
apiVersion: v1
kind: Node
#...
status:
    allocatable:
      devices.kubevirt.io/kvm: 1k
      cpu-feature.node.kubevirt.io/abm: 1k
      cpu-feature.node.kubevirt.io/adx: 1k
      #...
    capacity:
      devices.kubevirt.io/kvm: 1k
      cpu-feature.node.kubevirt.io/abm: 1k
      cpu-feature.node.kubevirt.io/adx: 1k
      #...
...
```
 When a VMI wants a VMI with specific feature, it would set the virt-loauncher pod with the desired resource:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-pod
spec:
  containers:
    - name: demo-container-1
      resources:
        limits:
          cpu-feature.node.kubevirt.io/adx: 1
```

### utilize Node Feature Discovery (NFD) option
NFD can do the same thing the kubevirt's node-labeller is doing.
When deployed, NFD:
- allows for non-standard user-specific feature labels created with the [NodeFeature CRD](https://kubernetes-sigs.github.io/node-feature-discovery/v0.15/usage/custom-resources.html#nodefeature).
  - virt-handler could use this in order to relay its labels, instead of patching them themselves.
- patches built-in labels with `feature.node.kubernetes.io/` prefix. See list [here](https://kubernetes-sigs.github.io/node-feature-discovery/v0.15/usage/features.html#built-in-labels).
  - These labels are similar but do not exactly match the ones Kubevirt is deploying right now (See [comparison](https://www.diffchecker.com/qA2EvefO/) of local dev node deployed using KubevirtCI)
  - This NFD feature will not be used for this solution.

Example:
Instead of directly patching these labels to the node: `cpu-feature.node.kubevirt.io/abm: "true"`, `cpu-feature.node.kubevirt.io/adx: "true"`, it would deploy the following CR:
```yaml
apiVersion: nfd.k8s-sigs.io/v1alpha1
kind: NodeFeature
metadata:
  labels:
    nfd.node.kubernetes.io/node-name: node-1
  name: node-1-labeller-sync
spec:
  labels:
    cpu-feature.node.kubevirt.io/abm: "true"
    cpu-feature.node.kubevirt.io/adx: "true"
#...
```
See full CR [here](https://paste.centos.org/view/ef218827)

Pros:
- No need to invent the wheel.
- Re-examine current methods to make sure if we still need them is a good practice. In this case we may find out that patching is no longer needed, and in the long run need to maintain less code.
- These steps can be broken to multiple seemingly unrelated PRs, that are easier to review.

Cons:
- Relying on a third part tools.
- Using third-party tools cannot filter only the kubevirt.io related labels.
- annotations are not addressed (for example `kubevirt.io/ksm-handler-managed` annotation)
- with NFD one cannot filter only kubevirt owned labels, making it a bit less safe.

## Solution option #5: [Ruled-out] Using status section
Same proxy solution as the primary solution where virt-hadler updates the labels/annotations somewhere, and it gets copied to the node by virt-controller, but instead of using a new CRD as the medium - writing the labels/Annotations to the status section of an existing CRD like kubevirt CR.

Pros:
- Avoid introducing a new CRD to the project.

Cons:
- [Deal-breaker] The status section is considered a user facing API section. Dynamically adding/updating 150+ technical labels there may only confuse the user and thus is not advisable.

## Solution option #6: [Ruled-out] Using the virt-handler pod as a medium to pass the labels/annotations
Same proxy solution as primary solution where virt-handler updates the labels/annotations somewhere, and it gets copied to the node by virt-controller, but instead of using a new CRD as the medium - writing the labels/annotations to the virt-handler pods's labels/annotations.
Labels/Annotations that need to pass to the node should get a special prefix (like "nodesync/"), so that regular labels don't get passed to the node by mistake.

Pros:
- Avoid introducing a new CRD to the project.
- A simple and clean approach.

Cons:
- [Deal-breaker] It requires giving the virt-handler RBAC permission to update pods, which is generally considered risky.
- It's only relevant for passing labels/annotations. CRD approach allows for future expansions.

## Solution option #7: [Ruled-out] Fine-tune each virt-handler RBAC to only patch its own node.
Introducing a new controller called token-distributor-controller. It will create a token for each virt-handler pod and deliver it to the node.

Pros:
- Avoid introducing a new CRD to the project.
- A malicious user can't manipulate pods from other nodes to migrate to their node.

Cons:
- A malicious user can still manipulate its own node with non-kubevirt controlled params/labels, then possibly lure control plane components to its node in order to steal their credentials. For example, it can add the control-plane label to its own node, then add other labels to attract components that seek specific functionalities, and take over them.

This solution is not a bad one, but it's more important to address not being able to affect non-kubevirt labels/annotations before doing this (see Follow-up Steps section).

# chosen solution

Considering the comments of all the reviewers, the solution chosen is validationAdmissionPolicy.

# Update/Rollback Compatibility

Note: TBD per chosen solution.

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
  * new virt-handler: shadowNode is written to, but it is not synced to the node yet. However, virt-handler keeps patching the node directly, so there is no upgrade issue.
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

# Appendix A: using new K8s resources on Kubevirt

When using new k8s resources on kubevirt in cases that solve a bug or CVE like the proposal above,
there is an inherent conflict with the fact that kubevirt needs to support two k8s versions backwards.
To handle this conflict, there are few possible options that can be taken:
- condition the new kubevirt fix (that uses new k8s resource) with a kubevirt Feature gate.
  - Pro:
    - allows the cluster-admin to opt in to the fix if they want to use it on k8s versions where these resources are not GA yet.
    - allows the cluster-admin to opt out of the solution. However, it is unclear why would this option be needed (especially in CVE cases).
  - Con:
    - There is no new "feature" added. Using the Feature-gate tool is not appropriate here, for several reasons:
      - It forces the cluster-admin to "enable" it to get the fix. This is not scalable, and usually not something the cluster-admin would want to specifically enable for each bug/CVE solved.
      - SW-Architecture wise - it is not the role of the cluster-admin to enable deployment of new fixes—that is the job of kubevirt's virt-operator.
      - New k8s resources also come with their own Feature-gate. Adding a kubevirt FG on top of the k8s just to comply with the 2 version backwards compatibility constraint is a misuse of the feature-gate tool.
- Best effort approach: condition the new kubevirt fix by whether the new resource is available on the cluster. This approach can also become more strict in later steps as mentioned on Cons.
  - Pro:
    - The cluster admin does not have to enable anything—the fix is enabled by default on versions where the new resource is supported.
  - Cons:
    - The cluster-admin will not "know" that a CVE fix was not deployed on their now upgraded kubevirt - although the version it upgraded to states it fixed it.
      - This could be mitigated by introducing a new kubevirt CR configuration "enforceSecurityFixes" which will have two options: "strict"/"permissive".
        - strict mode - kubevirt will prevent upgrade of kubevirt if a CVE fix requires a resource that is not currently enabled on the cluster.
        - permissive mode - kubevirt will not prevent upgrades—but instead will add an event/status to the kubevirt CR stating that in order for security fix to fully take effect, the user needs to upgrade k8s version.
