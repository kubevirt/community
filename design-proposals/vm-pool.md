# Overview

This design provides an approach for creating a VM grouping and replication abstraction for KubeVirt called a VirtualMachinePool.

## Motivation

The ability to manage a group (or pool) of similar VMs using a higher level abstraction is a staple among commonly utilized Iaas operational patterns. By bringing stateful VM group management to KubeVirt, we open the door for operation teams to utilize their existing patterns for managing KubeVirt VMs. This feature further aligns KubeVirt as an Iaas offering comparable to the public clouds which makes KubeVirt a more attractive option for Iaas management on baremetal hardware.

## Goals

* VM pool abstraction capable of managing replication of stateful VMs at scale
* Automated rollout of spec changes and other updates across a pool of stateful VMs
* Automated and manual scale-out and scale-in of VMs associated with a pool
* Autohealing (delete and replace) of VMs in a pool not passing health checks
* Ability to specify unique secrets and configMap data per VM within a VMPool
* Ability to detach VMs from VMPool for debug and analysis

## Non-Goals

* Not designing a VM fleet abstraction capable of managing multiple VM groupings containing VMs which are dissimilar to one another. A VMPool consists only of VMs which are similar in shape to one another that are derived from a single config.

## Terms

* **VirtualMachine [VM]** - Refers specifically to the KubeVirt VirtualMachine API object.
* **VirtualMachinePool [VMPool]** - KubeVirt API for scaling out/in replicas of KubeVirt VirtualMachines.
* **Scale-out and Scale-in** - Terms to describe the action of modifying the replica count of VMs within a VMPool.
* **Detach VM** - The process of manually separating a VirtualMachine from a VirtualMachinePool.
* **Update Strategy** - The policy used to define how a VMPool handles rolling out VM spec updates to VMs within the pool.
* **Scale-In Strategy** - The policy used to define how a VMPool handles removing VMs from a pool during scale-in.

## User Stories

* As a cluster user, I want to automate batch rollout of changes (CPU/Memory/Disk/PubSSHKeys/etc…) across a pool of VM replicas.
* As a cluster user managing a pool of VM replicas I want to automate scale out of VM instances based on utilization
* As a cluster user managing a pool of VM replicas I want to automate scale-in of VM instances to optimize cluster resource consumption
* As a user transitioning workloads to KubeVirt I want to use similar management patterns provided by existing Iaas platforms (AWS, Azure, GCP)
* As a cluster admin managing nested Kubernetes clusters on top of KubeVirt VMs, I want the ability to elastically scale the underlying KubeVirt VM infrastructure.
* As a SRE managing the availability VM replicas in a pool, I want to automate VM recovery by auto detecting and deleting misbehaving VMs and having the platform spin up fresh new instances as a replacement.
* As a pool user/manager I want to remove a VM from the pool without modifying it for debugging. The missing VM can be replaced by the pool.

# Design

The VMPool design introduces a new API represented as a CRD called the VirtualMachinePool (VMPool) object. This object contains tunings related to managing a set of replicated stateful VMs as well as a template that defines the configuration applied creating the VM replicas. Conceptually, The VMPool's templating mechanism is very similar to how Kubernetes Deployments operate.

## VirtualMachinePool (VMPool) API

The VMPool API represents all the tunings necessary for managing a pool of stateful VMs. The VMPools spec contains the following tunings and values

* **Template** - (Required) A VirtualMachine spec used as a template when creating each VM in the pool.
* **Replicas** - (Required) An integer representing the desired number of VM replicas
* **MaxUnavailable**  - (Optional) (Defaults to 25%) Integer or string pointer, that when set represents either a percentage or number of VMs in a pool that can be unavailable (ready condition false) at a time during automated update.
* **NameGeneration** - (Optional) Specifies how objects within a pool have their names generated
	* **AppendPostfixToSecretReferences** - (default false) Boolean that indicates if VM’s unique postfix should be appended to references to Secrets in the VMI’s Volumes list. This is useful when needing to pre-generate unique secrets for VMs within a pool.
	* **AppendPostfixToConfigMapReferences** - (default false) Boolean that indicates if VM’s unique postfix should be appended to ConfigMap references in the VMI’s Volumes list. This is useful when needing to pre-generate unique secrets for VMs within a pool.

* **UpdateStrategy** - (Optional) Specifies how the VMPool controller manages updating VMs within a VMPool
	* **Unmanaged** - No automation during updates. The VM is never touched after creation. Users manually update individual VMs in a pool.
	* **Opportunistic** - Opportunistic update of VMs which are in a halted state.
	* **Proactive** - (Default) Proactive update by forcing VMs to restart during update.
		* **SelectionPolicy** - (Optional) (Defaults to "random" base policy when no SelectionPolicy is configured) The priority in which VM instances are selected for proactive scale-in
			* **OrderedPolicies** - (Optional) Ordered list of selection policies. Initial policies include [LabelSelector]. Future policies may include a [NodeSelector] or other selection mechanisms.
			* **BasePolicy** - (Optional) Catch all polices [Oldest|Newest|Random]
* **ScaleInStrategy** - (Optional) Specifies how the VMPool controller manages scaling in VMs within a VMPool
	* **Unmanaged** - No automation during scale-in. The VM is never touched after creation. Users manually delete individual VMs in a pool. Persistent state preservation is up to the user removing the VMs
	* **Opportunistic** - Opportunistic scale-in of VMs which are in a halted state.
		* **StatePreservation** - (Optional) specifies if and how to preserve state of VMs selected for scale-in.
			* **Disabled** - (Default) all state for VMs selected for scale-in will be deleted
			* **Offline** - PVCs for VMs selected for scale-in will be preserved and reused on scale-out (decreases provisioning time during scale out)
			* **Online** - [NOTE we can't implement this until we have the ability to suspend VM memory state to a PVC] PVCs and memory for VMs selected for scale-in will be preserved and reused on scale-out (decreases provisioning and boot time during scale out)
Each VM’s PVCs are preserved for future scale out
	* **Proactive** - (Default) Proactive scale-in by forcing VMs to shutdown during scale-in.
		* **SelectionPolicy** - (Optional) (Defaults to "random" base policy when no SelectionPolicy is configured) The priority in which VM instances are selected for proactive scale-in
			* **OrderedPolicies** - (Optional) Ordered list of selection policies. Initial policies include [LabelSelector]. Future policies may include a [NodeSelector] or other selection mechanisms.
			* **BasePolicy** - (Optional) Catch all polices [Oldest|Newest|Random]
		* **StatePreservation** - (Optional) specifies if and how to preserve state of VMs selected for scale-in.
			* **Disabled** - (Default) all state for VMs selected for scale-in will be deleted
			* **Offline** - PVCs for VMs selected for scale-in will be preserved and reused on scale-out (decreases provisioning time during scale out)
			* **Online** - [NOTE we can't implement this until we have the ability to suspend VM memory state to a PVC] PVCs and memory for VMs selected for scale-in will be preserved and reused on scale-out (decreases provisioning and boot time during scale out)
Each VM’s PVCs are preserved for future scale out
* **Autohealing** - (Optional)  (Defaults to disabled with nil pointer) Pointer to struct which specifies when a VMPool should should completely replace a failing VM with a reprovisioned instance. 
	* **StartupFailureThreshold** - (Optional) (Defaults to 3) An integer representing how many consecutive failures to reach a running state (which includes failing to pass liveness probes at startup when liveness probes are enabled) should result in reprovisioning.

## VMPool API Examples

**Automatic rolling updates and scale-in strategy with state preservation to optimization of boot times during scale-out**

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachinePool
metadata:
  name: my-vm-pool
spec:
  replicas: 100
  maxUnavailable: 10
  scaleInStrategy:
    proactive:
      statePreservation: Offline
      selectionPolicy:
        basePolicy: "Oldest"
  updateStrategy:
    proactive:
      selectionPolicy:
        basePolicy: "Oldest"
  template:
    spec:
      dataVolumeTemplates:
      - metadata:
          name: alpine-dv
        spec:
          pvc:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 2Gi
          source:
            http:
              url: http://cdi-http-import-server.kubevirt/images/alpine.iso
      running: false
      template:
        spec:
          domain:
            devices:
              disks:
              - disk:
                  bus: virtio
                name: datavolumedisk
          terminationGracePeriodSeconds: 0
          volumes:
          - dataVolume:
              name: alpine-dv
            name: datavolumedisk
```

**Manual rolling updates and Manual scale-in strategy**

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachinePool
metadata:
  name: my-vm-pool
spec:
  replica: 100
  scaleInStrategy:
    unmanaged: {}
  updateStrategy:
    unmanaged: {}
  template:
    spec:
      dataVolumeTemplates:
      - metadata:
          name: alpine-dv
        spec:
          pvc:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 2Gi
          source:
            http:
              url: http://cdi-http-import-server.kubevirt/images/alpine.iso
      running: false
      template:
        spec:
          domain:
            devices:
              disks:
              - disk:
                  bus: virtio
                name: datavolumedisk
          terminationGracePeriodSeconds: 0
          volumes:
          - dataVolume:
              name: alpine-dv
            name: datavolumedisk
```

**Automatic rolling updates and scale-in strategy with VM ordered selection policy on scale-in**

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachinePool
metadata:
  name: my-vm-pool
spec:
  replicas: 100
  maxUnavailable: 10
  scaleInStrategy:
    proactive:
      selectionPolicy:
        orderedPolicies:
          - labelSelector
            - non-important-vms
        basePolicy: "Oldest"
      statePreservation: Offline
  updateStrategy:
    proactive:
      selectionPolicy:
        basePolicy: "Oldest"
  template:
    spec:
      dataVolumeTemplates:
      - metadata:
          name: alpine-dv
        spec:
          pvc:
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 2Gi
          source:
            http:
              url: http://cdi-http-import-server.kubevirt/images/alpine.iso
      running: false
      template:
        spec:
          domain:
            devices:
              disks:
              - disk:
                  bus: virtio
                name: datavolumedisk
          terminationGracePeriodSeconds: 0
          volumes:
          - dataVolume:
              name: alpine-dv
            name: datavolumedisk
```

# Special Topics

The topics in this section tackles in detail how specific use cases and functionality are handled with the VMPool abstraction.

## Manually Detaching VM from VMPool

A VM in a VMPool can be detached from a VMPool by removing the owner reference. This removes that VM from being actively managed by the VMPool.

Since VMs within a VMPool each have a unique sequential postfix applied to each VM name, a detached VM’s sequence number will be skipped during scale-in and scale-out operations until the detached VM is either returned to the Pool (by manualing adding the VMPool owner reference back) or the VM is deleted which frees the resource name.

## VM Naming

By default, VM names are generated from the VMPool’s name by appending the VMPool’s name with a sequential unique postfix. This is similar to how pods are generated from a StatefulSet’s name.

Here is an example of how the VMPools sequential postfix will work in practice.

Starting with a VMPool that has 3 VMs.

* `my-vm-1`
* `my-vm-2`
* `my-vm-3`

During scale-in (`replicas: 2`), my-vm-2 is removed, which results in the set looking like this.

* `my-vm-1`
* `my-vm-3`

On the next scale-out event, the VMPool is going to search sequentially to fill any gaps before appending new sequence numbers. If the VMPool's replica count is set to 4 (`replicas: 4`), the new set will fill in the missing my-vm-2 and create a new my-vm-4.

* `my-vm-1`
* `my-vm-2` (recreated from previous state if `scaleInPreserveState=Offline`)
* `my-vm-3`
* `my-vm-4` (newly provisioned)

## State Preservation during Scale-in

During scale-in when the `scaleInStrategy` is set to `Proactive` with `StatePreservation=Offline`, the VM’s being removed from the pool will have their PVC state preserved. In order to ensure on the next scale-out event that VMs using the exact same state are started again, the previous VM names will be reused during scale-out. This is similar in concept to how a StatefulSet uses predictable sequential names.

When `scaleInStrategy` is set to `Proactive` with `Preservation=Disabled`, all PVC state will be completely removed from VMs during scale-in and reprovisioned during scale-out.

## Handling of Annotations and Labels

VMs inherit the Labels/Annotations from the VMPool.Spec.Template.Metadata section.

VMIs inherit the Labels/Annotations from the VMPool.Spec.Template.Spec.Template.Metadata section of a VMPool.

## Handling Persistent Storage

Usage of a DataVolumeTemplate within a VMPool.Spec.Template will result in unique persistent storage getting created for each VM within a VMPool. The DataVolumeTemplate name will have the VM’s sequential postfix appended to when the VM is created from the VMPool.Spec.Template which makes each VM a completely unique stateful workload.

## Handling Unique VM CloudInit and ConfigMap Volumes at Scale

By default, any secrets or configMaps references in a VMPool.Spec.Template.Spec.Template’s Volume section will be used directly as is, without any modification to the naming. This means if you specify a secret in a CloudInitNoCloud volume, that every VM instance spawned from the VMPool with this volume will get the exact same secret used for their cloud-init user data.

This default behavior can be modified by setting the **AppendPostfixToSecretReferences** and **AppendPostfixToConfigMapReferences** booleans to true on the VMPool spec. When these booleans are enabled, references to secret and configMap names will have the VM’s sequential postfix appended to the secret and configmap name. This allows someone to pre-generate unique per VM secret and configMap data for a VMPool ahead of time in a way that will be predictably assigned to VMs within the VMPool.

## Autohealing

When managing VMs at large scale, it is useful to automate the recovery of VMs which continually fail to reach a running phase or pass the liveness probes. This automatically fixes scenarios where a VM’s state has somehow been corrupted and needs to be completely refreshed.

Autohealing has two layers.

* VMI Recovery - This is configured at vmi layer through the use of LivenessProbes on the VMI template within the VMPool. More info about Liveness probes can be found [here](http://kubevirt.io/user-guide/virtual_machines/liveness_and_readiness_probes/#liveness-and-readiness-probes).
* VM Recovery - This is configured at the VMPool management layer through the use of the VMPool.Spec.Autohealing tunable.

VMI recovery involves simply tearing down a VM's active VMI and restarting it. The VM's volumes are preserved and the new VMI is launched using the existing volumes. This is essentially an automated VM restart.

VM recovery involves a VMPool completely deleting a VM and the VM's state, then reprovisioning the VM. This is a complete state refresh.

By enabling `Autohealing: {}` on the VMPool’s spec, VMs which continually fail to successfully launch and pass an initial liveness check will automatically be deleted (including persistent storage) and re-provisioned. This allows for auto recovery in application scenarios that can withstand such an action.

Autohealing must have an exponential backoff mechanism to prevent it from causing unnecessary strain on the cluster due to state reprovisioning.

## Throttling Parallel VM Creation/Update

The VMPool controller should establish some default upper limits when it comes to how many VMs can be batch created and updated at a time. This is similar to the VMIReplicaSet controller's internal `BurstReplicas` global variable. By default VMIReplicaSets will only create at most 250 VMIs at a time, but this is configurable as a global setting. The VMPool controller should adopt the same pattern of a max upper limit of 250 VMs per a VMPool with a global configurable setting.

This setting is primarily meant as a way of preventing the VMPool controller from creating an unintentionally cluster DoS.

## VM Selection during Scale-In and Updates

The VMPool spec includes a `selectionPolicy` field for proactive scale-in and proactive updates. This field allows creators of a VMPool to define how VMs will be selected to proactively act upon.

Within the `selectionPolicy` there's a tuning called the `basePolicy` that is meant to act as a "catch all" policy, meaning it is always possible to find a VM which matches the base policy. Examples of base policies are values such as "oldest", "newest", and "random". The "random" policy will include optimizations that attempt to select VMs in the pool based on the least amount of disruption.

In addition to the base policy, there's an ordered list called `orderedPolicies` which allows the VMPool creator to define custom criteria for selecting VMs as well as a priority for the criteria. Initially for the first implementation of VMPools, the ordered policies will be limited to a `labelSelector`. Multiple label selectors can be defined in the ordered policies list and the priority each label selector has is based on its order in the list. In the future, new types of ordered policies may exist, such as node selector for example. The types of ordered policies can expanded as new use cases arise.

## VirtualMachinePool vs VirtualMachineInstanceReplicaSet

VMPools are designed to manage stateful VM workloads while VMIReplicaSets are designed to manage stateless VMI workloads.

A VMIReplicaSet makes sense for a user who wants to manage stateless VMI objects using similar patterns to how Kubernetes ReplicaSets manage pods.

A VMPool makes sense for a user who wants to manage stateful VM objects using similar patterns found in Iaas public clouds, like AWS, GCP, Azure. In this way, a VMPool is more similar to an AWS AutoscalingGroup or a GCP ManagedInstanceGroup than it is to any core Kubernetes API.

## VirtualMachinePool Metrics

While it is expected that the type and number of metrics collected that are specific to the VMPool object will change as production use cases continue to evolve, we do have an idea of some metrics that are useful.

* A rate metric measuring the number of VM starts within a pool
* A rate metric measuring the number of VM stops within a pool
* Adding a vmpool label to the phase transition time histograms allowing us to isolate transition times by pool

