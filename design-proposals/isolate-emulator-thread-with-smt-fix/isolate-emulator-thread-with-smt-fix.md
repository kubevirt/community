# Preface
High performance VMs, such as VMs using [DPDK](https://en.wikipedia.org/wiki/Data_Plane_Development_Kit) or [realtime](https://en.wikipedia.org/wiki/Real-time_computing) applications are sensitive to interference from neighboring workloads.
In order to avoid noisy CPU neighbors, such VMs need to use the [dedicatedCPUPlacement and IsolateEmulatorThread](https://kubevirt.io/user-guide/virtual_machines/dedicated_cpu_resources/#requesting-dedicated-cpu-for-qemu-emulator) features, along with other custom cluster configuration.
One of these cluster changes is to set the Kubelet’s CPUManager Policy to static - [full-pcpus-only](https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/). This forces kubelet to only accept Pods that have CPUs with full physical cores.

However, on nodes with [hyperthreading](https://en.wikipedia.org/wiki/Hyper-threading) (SMT) enabled, this causes an error, described on [BZ#2228103](https://bugzilla.redhat.com/show_bug.cgi?id=2228103).
This design proposal is meant to fix this issue.

Considering that hyperthreading (SMT) is enabled by default in many [Intel](https://www.intel.com/content/www/us/en/architecture-and-technology/hyper-threading/hyper-threading-technology.html) processors - it means that DPDK and Realtime users may be impacted.

## Motivation
Running High performance VMs with an even number of CPUs with the conditions below.

### The problem
In order to replicate this issue, the following conditions must be met:
* The node has hyperthreading (SMT) enabled.
* Kubelet’s CPUManager is set to static - full-pcpus-only.
* Kubevirt CR's defaultRuntimeClass is set to a high performance runtimeClass.
* VM is set with `dedicatedCPUPlacement` and `IsolateEmulatorThread` enabled.
* VM requests an even number of CPUs.

When enabling `IsolateEmulatorThread` on a VM with even CPU count, Kubevirt is requesting one extra CPU for the virt-launcher pod. This is an internal change that is not exposed to the user.
After the pod is scheduled, the now odd number of CPUs will collide with the CPU-Manager policy of using only full physical cores, and cause the kubelet to [reject](https://github.com/kubernetes/kubernetes/pull/101432) this pod.
For example, let’s consider this VM spec:
```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  running: true
  template:
    spec:
      domain:
        cpu:
          dedicatedCpuPlacement: true
          isolateEmulatorThread: true
          sockets: 1
          cores: 4
          threads: 2
```
kubelet will emit following event:
```
SMT Alignment Error: requested 9 cpus not multiple cpus per core = 2
```
The following picture helps show how this is a problem
<img src="smt-slignment-error-diagram.png">

## Goals
Fix [BZ#2228103](https://bugzilla.redhat.com/show_bug.cgi?id=2228103) on the Kubevirt scope.

## Non  Goals
Fix the actual scheduling of the virt-launcher to a node that will reject it.

## Definition of Users
This bug addresses all KubeVirt users running VMs with isolateEmulatorThread on nodes with hyperthreading enabled - specifically DPDK and Realtime users.

## User Stories
A user should be able to run VM with dedicatedCPUPlancement and isolateEmulatorThread enabled on a node with hyperthreading enabled, regardless of the amount of guest CPUs requested.

## Repos
- [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Bug fix design proposal
The bug fix proposal introduces an alpha annotation `kubevirt.io/CPUManagerPolicyBetaOptions:full-pcpus-only`:
```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  running: true
  template:
    metadata:
      annotations:
        kubevirt.io/CPUManagerPolicyBetaOptions:full-pcpus-only
    spec:
      domain:
        cpu:
          dedicatedCpuPlacement: true
          isolateEmulatorThread: true
          sockets: 1
          cores: 4
          threads: 2
```

When the VM is configured with `isolateEmulatorThread: true` and the above annotation, the virt-controller will set the virt-launcher pod with an even amount of CPUs, meaning:

| User requested CPUs               |Even (2,4,6,8,...)| odd (1,3,4,5,7,9,...) |
|-----------------------------------|------------------|-----------------------|
| Extra emulatorThread CPUs for pod |        +2        | +1 (as usual)         |

Notes:
1. In both scenarios the guest will get the same amount of CPUs it requested.
2. The housekeeping cgroup will also update to now contain the 2 CPUs.

Pros:
* Cleaner user experience. No prior knowledge of CPUs added behind the scene is needed. 
* Matches the current crio convention of adding annotations in order to get high-performance pods, like `cpu-load-balancing.crio.io: disable`
* Reduces the “wasted” CPUs to minimum possible per scenario, i.e. only for VMs that specifically add this annotation, and only when they ask for even CPUs. 
* As this is an alpha annotation, it can be removed/ignored when and if solutions inside/outside Kubevirt such as CPU Manager handles the situation differently.

Cons:
* We’re adding yet another semi alpha knob to an already complicated set of knobs.

# Fix Alternatives
## Alternative #1 Adding `full-pcpus-only` option as a global configuration
As opposed to the suggested solution where the user needs to add this annotation to every VM, it is possible to make this behavior global. The assumption is that clusters run homogeneous Nodes, so this will eliminate the need to add this annotation every time.

Pros:
* VM configuration is less complex.
* Similar to the way default runtimeclass is configured on Kubevirt.

Cons:
* The suggested annotation is waisting a CPU when used on nodes with hyperthreading disabled. A cluster-wide configuration may be wasteful for heterogeneous clusters, where node-level granularity is needed.

## Alternative #2 Configure both pod and Guest
It is possible to use both the current Guest API and the old yet existing pod API, like so:
```
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  running: true
  template:
    spec:
      domain:
        resources:
          requests:
            cpu: '10'
          limits:
            cpu: '10'
        cpu:
          cores: 4
          sockets: 1
          threads: 2
          dedicatedCpuPlacement: true
          isolateEmulatorThread: true
```
In order to do that the current validating webhooks regarding dedicatedCpuPlacement and IsolatedEmulatorThread will have to be loosened.

Pros:
* This option does not introduce any alpha knobs.

Cons:
* This option exposes the user to unnecessary implementation details and possible waste of physical CPUs. They have to do things manually, opening the door for misuse. Loosening the validating webhook doesn't make sense in the general case.
* KubeVirt will need to adjust internal assumptions that are based on the fact that VMs {req, limits}.cpu = vCPUs. Specifically, how the housekeeping cgroup is being formed, but possibly other aspects.
* VMs that were defined like this will have to change when the long term solution will arrive (backwards-compatibility issue)

## Scalability
The solution proposed is “wasting” one unused CPU for the purpose of maintaining SMT alignment. I believe it is a minor waste considering the alternatives, such as disabling SMT on the node. As I consider the solution adequately scalable.

## Update/Rollback Compatibility
Since the annotation is not an API change, it is backportable.
When a better solution is possible, Kubevirt can simply replace the annotation support with a better approach, thus not making the change relatively seamless for the client. 

## Functional Testing Approach
Will add unit tests to `renderresources_test.go` to check extra housekeeping logic, and in `templates.go` to check the virt-launcher pod spec output.

# Implementation Phases
## Phase 1: Align CPU to SMT requirements when feature gate is enabled and annotation is set on the VMI.
* Virt-config:
  * Add support for the new feature gate: `FullPCPUsOnly`.
* Virt-API:
  * Add support for the new alpha annotation: `kubevirt.io/CPUManagerPolicyBetaOptions:full-pcpus-only`.
* Virt-controller:
  * In case annotation is set, feature gate is enabled and IsolateEmulatorThread is true, add logic to Align the total number of CPUs to be even.
## Phase 2: Adjust housekeeping cgroup to use the extra CPU added by the alignment, if such a one is available.
* Virt-launcher:
  * Try to fit a full core to the housekeeping cgroup. If not enough CPUs - allocate a single thread like before (best-effort approach).
* Virt-handler:
  * Update code to allow accepting a housekeeping CPUSet of >1 CPUs.
