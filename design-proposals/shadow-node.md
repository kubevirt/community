# Overview

A reported [Advisory](https://github.com/kubevirt/kubevirt/security/advisories/GHSA-cp96-jpmq-xrr2) is outlining how of our virt-handler component can be abused to escalate local privileges when a node, running virt-handler, is compromised. A flow and more details can be found on reported [issue](https://github.com/kubevirt/kubevirt/issues/9109).

This proposal is outlining mitigation in a form of internal change.

## Motivation

Kubevirt should be secure by default. While mitigation is available, it is not part of Kubevirt.

## Goals

Mitigate the advisory by default.

## Non Goals

## Definition of Users

Not user-facing change

## User Stories

Not user-facing change

## Repos

Kubevirt/Kubevirt

# Design
(This should be brief and concise. We want just enough to get the point across)

In order to easy reviews and understand what needs to be change the first section will describe different usage of Node object in the virt-handler.

## Node usage
1. [markNodeAsUnschedulable](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/cmd/virt-handler/virt-handler.go#L185) is used at the start and the termination (by SIGTERM only) of virt-handler to indicate (best-effort) that the node is not ready to handle VMs.

2. Parts of Heartbeat
    1. [labelNodeUnschedulable](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L96) is similarly as markNodeAsUnschedulable used only when virt-handler is stopping.

    2. As part of [do](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L119) we do 2 calls:
     [Get](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L139) - irrelevant as this is read access.
    [Patch](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/heartbeat/heartbeat.go#L139) - updating NodeSchedulable, CPUManager, KSMEnabledLabel labels and VirtHandlerHeartbeat KSMHandlerManagedAnnotation annotations once per minute.

3. As part of node labeller
    1. [run](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/node-labeller/node_labeller.go#L195) - irrelevant as this is read access

    2. [patchNode](https://github.com/kubevirt/kubevirt/blob/5b61a18f4c3b6c549d50fd201664c21720e3dbe2/pkg/virt-handler/node-labeller/node_labeller.go#L254) - node labeller patching node with labels for various informations which are mostly static in nature:  cpu model, hyperv features, cpu timer, sev/-es, RT...
This is triggered both by interval (3 min) and on changes of cluster config. 

## Shadow node

The proposed solution is to introduce new CRD that will be shadowing a node "ShadowNode". The virt-handler will be writting to ShadowNode and virt-controller will have a responsibility to sync allowed information to Node.

## API Examples
```yaml
apiVersion: <apiversion>
kind: ShadowNode
metadata:
  name: <Equivalent to existing name>
  // Labels and Annotations only
spec: {} //Empty, allows futher extension
status: {} //Empty, allows futher extension
```

## Scalability

There are few aspects to consider. 

1.  #shadowNodes will be equivalent #nodes, negligible space overhead

2. #writes could double. Here it is importent to look at the usage and sort each case to 2 categories (low and high frequent write). The first usage is clearly low frequent as long as virt-handler operates as designed.
The second usage consist of two cases which might seem different but in fact these are same most of the time because NodeSchedulable, CPUManager, KSMEnabledLabel labels and KSMHandlerManagedAnnotation annotation is mostly static. What is left is VirtHandlerHeartbeat that is not necessary to sync on a Node (requires re-working node controller).
The third case is propagating labels that are also mostly static.
With all the above doubling #writes is unlikely. 


## Update/Rollback Compatibility

Virt-handler is going to continue writting to Node object, therefore update should be without compatibility issues.

## Functional Testing Approach

Existing functional tests should be enough to cover this change.

# Implementation Phases
(How/if this design will get broken up into multiple phases)
