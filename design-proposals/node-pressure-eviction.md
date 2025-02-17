# Overview
Kubernetes has two types of pod eviction.

[API-Initiated Eviction](https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/) and [Node Pressure Eviction](https://kubernetes.io/docs/concepts/scheduling-eviction/node-pressure-eviction/). The API-Initiated Eviction is coordinated at the cluster level. This is the kind of eviction used during node draining events and it’s the type of eviction that observes Pod Disruption Budgets.

Node Pressure Eviction is completely independent from the API-Initiated Eviction. This eviction occurs directly on the node driven by the kubelet. Pod Disruption Budgets are not observed and neither is the pod’s termination grace period. This means when a node is under pressure, the kubelet has the ability to kill workloads independently of any cluster wide rules in order to free up resources.

There are ways to influence and even disable Node Pressure Eviction using the kubelet config, but be aware that these are node level configurations and not directly influenced by the workload itself. **Meaning, unlike API-Initiated Eviction, workload owners cannot directly influence how Node Pressure Eviction behaves for their workload.**

## Goals
Currently, the VMI api’s EvictionStrategy tunable only works with API-Initiated Eviction. The goal is to extend the behavior of the EvictionStrategy to Node Pressure Eviction as well. This means if a VMI has an EvictionStrategy of LiveMigrate, we expect the system to attempt to live migrate the VMI regardless of which eviction method initiated the eviction.

## Non Goals
* There is no expectation that NodePressure and API-Initiated eviction will provide the same guarantees.

## Definition of Users
- Cluster Admins who wish to influence at the cluster level how VMs behave during Node Pressure Eviction

## User Stories
* As a cluster admin, I would prefer that VMs across the cluster default to attempting best effort live migrated during Node Pressure Eviction rather than being restarted.

# Design
This section outlines the theory behind how this design works and the code changes required to implement this theory into practice.

## Design Theory

The theory behind this design can be summed up in the following points.

* Currently a user expresses their desire to perform an API initiated shutdown of a VMI by deleting that VMI. This results in vmi.DeletionTimestamp != nil.
* Currently when the virt-launcher pod intercepts a signal to shutdown (SIG_TERM), it passes that knowledge to virt-handler and virt-handler coordinates how to shutdown the VMI.
* This means if virt-handler is told by virt-launcher that a shutdown signal has been received, and the vmi.DeletionTimestamp == nil, then this shutdown is not an expression of the user’s intent but instead an expression of the systems attempt to evict the workload.

Essentially, if the system observes a virt-launcher receive a shutdown signal (like SIG_TERM), but the vmi.DeletionTimestamp == nil, then the system can process that as an eviction.

## Theory in Practice

The EvictionStrategy tunable is configurable at the cluster scope and on a per VM/VMI scope. This tunable aims to give cluster admins and VM owners control over how VM’s are processed during eviction.

The EvictionStrategy can be extended to work for Node Pressure Eviction by intercepting the VMI’s shutdown signal in virt-handler and transform that into an eviction requisition on the VMI. An eviction is signalled on a VMI by setting the vmi.Status.EvacuationNodeName to equal the name of the node that the VMI should be moved from.

In practice, this can be implemented through three changes into virt-handler

First, create some helper functions to determine if the VMI should evacuate. This logic detects if a vmi is being torn down externally of the user’s intent, and makes a decision on whether that should invoke an evacuation or not based on the VMI’s EvictionStrategy.

```code
diff --git a/pkg/virt-handler/vm.go b/pkg/virt-handler/vm.go
index b0ab86ace6..c87f5e0b0e 100644
--- a/pkg/virt-handler/vm.go
+++ b/pkg/virt-handler/vm.go
@@ -459,6 +459,61 @@ func (c *VirtualMachineController) startDomainNotifyPipe(domainPipeStopChan chan
 	return nil
 }
 
+func domainIsAlive(domain *api.Domain) bool {
+	if domain == nil {
+		return false
+	}
+	return domain.Status.Status != api.Shutoff &&
+		domain.Status.Status != api.Crashed &&
+		domain.Status.Status != ""
+}
+
+func (c *VirtualMachineController) shouldEvacuateVMI(vmi *v1.VirtualMachineInstance, domain *api.Domain) bool {
+
+   // TODO put Node Pressure Eviction feature gate check here
+
+	// If VMI or Domain is no longer active, don't evacuate
+	if vmi == nil ||
+		domain == nil ||
+		!vmi.IsRunning() ||
+		!domainIsAlive(domain) {
+		return false
+	}
+
+	// If VMI is being torn down due to deletion, don't evacuate
+	if vmi.DeletionTimestamp != nil {
+		return false
+	}
+
+	// If virt-launcher has not signaled graceful shutdown, don't evacuate
+	gracefulShutdown := c.hasGracefulShutdownTrigger(domain)
+	if !gracefulShutdown {
+		return false
+	}
+
+	markForEviction := false
+
+	// At this point we know we have an active VMI that is not being
+	// deleted, but virt-launcher is signalling to us that the VMI's
+	// pod is being torn down (node level eviction).
+	//
+	// Choose to evacuate (livemigrat) based on the EvictionStrategy
+	// and capabilities of the VMI
+	evictionStrategy := migrations.VMIEvictionStrategy(c.clusterConfig, vmi)
+	switch *evictionStrategy {
+	case v1.EvictionStrategyLiveMigrate:
+		if vmi.IsMigratable() {
+			markForEviction = true
+		}
+	case v1.EvictionStrategyLiveMigrateIfPossible:
+		if vmi.IsMigratable() {
+			markForEviction = true
+		}
+	case v1.EvictionStrategyExternal:
+		markForEviction = true
+	}
+
+	return markForEviction
+}
```

Second, within the sync logic of the reconcile loop, ignore the request to shutdown if the VMI should be evacuated instead.

```code
diff --git a/pkg/virt-handler/vm.go b/pkg/virt-handler/vm.go
@@ -1923,8 +1983,16 @@ func (c *VirtualMachineController) defaultExecute(key string,
 	gracefulShutdown := c.hasGracefulShutdownTrigger(domain)
 	if gracefulShutdown && vmi.IsRunning() {
 		if domainAlive {
-			log.Log.Object(vmi).V(3).Info("Shutting down due to graceful shutdown signal.")
-			shouldShutdown = true
+			if c.shouldEvacuateVMI(vmi, domain) {
+				// Node level eviction detected.
+				// ignore graceful shutdown and let virt-controller's evacuation controller
+				// make the decision on how to proceed with the vmi shutdown.
+				log.Log.Object(vmi).V(3).Info("Received node level eviction signal.")
+			} else {
+				log.Log.Object(vmi).V(3).Info("Shutting down due to graceful shutdown signal.")
+				shouldShutdown = true
+			}
+
```

And lastly, set the vmi.EvacuationNodeName on the VMI’s status during the update status part of the reconcile loop if the vmi should be evacuated.

```code
diff --git a/pkg/virt-handler/vm.go b/pkg/virt-handler/vm.go
@@ -1440,6 +1495,14 @@ func (c *VirtualMachineController) updateVMIStatus(origVMI *v1.VirtualMachineIns
 
 	controller.SetVMIPhaseTransitionTimestamp(origVMI, vmi)
 
+	// process eviction
+	if vmi.Status.EvacuationNodeName != vmi.Status.NodeName &&
+		c.shouldEvacuateVMI(vmi, domain) {
+
+		vmi.Status.EvacuationNodeName = vmi.Status.NodeName
+		log.Log.Object(vmi).V(3).Info("Marked node level eviction signal.")
+	}
+
 	// Only issue vmi update if status has changed
 	if !equality.Semantic.DeepEqual(oldStatus, vmi.Status) {
 		key := controller.VirtualMachineInstanceKey(vmi)
```

From here, the evacuation controller (pkg/virt-controller/watch/drain/evacuation/evacuation.go) will handle the VMI evacuation

## API

Having KubeVirt respond to Node Pressure Eviction should be an opt-in feature gate (NodePressureEvictionLiveMigration) at the cluster level that cluster admins toggle to enable the behavior. Once the `NodePressureEvictionLiveMigration` feature gate is enabled, KubeVirt will make decisions on how to respond to Node Pressure Eviction based on the per VM or cluster wide EvictionStrategy.

## Update/Rollback Compatibility

Attempting to enable the NodePressureEvictionLiveMigration feature gate on previous versions of KubeVirt will fail when this feature gate does not exist.

## Functional Testing Approach

This feature can be tested by modifying the kubelet’s config on a set of nodes. Below are some example values to use. These example values set the soft eviction memory threshold high in order to make it easier to invoke the node pressure eviction. The evictionMaxPodGracePeriod corresponds to how long the pod can block shutting down until it is force killed. This value translates to how long the system has to evacuate the VMI via live migration before the pod dies.

```yaml
evictionSoft:
  memory.available: "3500Mi"
evictionSoftGracePeriod:
  memory.available: "5s"
evictionMaxPodGracePeriod: 300
evictionHard:
  memory.available: "10Mi"
```

To test invoking node pressure eviction, fill up a node with VMI’s that are configured with EvictionStrategy: LiveMigrate until the eviction memory threshold is hit. Then verify the migration is triggered and succeeds before the source pod is killed.

# Open Questions

## Recording Cause of Migration on VMIM

Today when a VMIM is created, we don’t have context into why the system created the VMIM. It would be nice to have information tied to the VMIM that indicates why the migration was invoked… such as “API Eviction”, “Node Pressure Eviction, “virt-launcher update”, etc…

## Migration Time Optimizations

Even with a substantial grace period, there is some risk that the live migration will not complete in time before the VMI pod is killed when Node Pressure Eviction takes place. This is because we don’t have a mechanism to block the pod from being killed after the evictionMaxPodGracePeriod is exceeded or a hard eviction threshold is hit.

It’s possible we should consider pausing the VMI during the live migration in an attempt to converge on the Live Migration quicker.

