# Overview
This KubeVirt design proposal discusses possible implementations for adding -
and removing - network interfaces from running Virtual Machines, without
requiring a restart.

## Motivation
Hot-plug / hot-unplug (add / remove) nics to running VMs is an industry
standard available in multiple platforms, allowing the dynamic attachment of L2
networks. This is useful when the workload (VM) cannot tolerate a restart when
attaching / removing networks, of for scenarios where, for instance, the
workload is created prior to the network.

This would help bridge the gap between KubeVirt's feature set and what the user
base expects from it.

## Goals
Adding network interfaces to running VMs.
Removing networking interfaces from running VMs.
A VM can have multiple interfaces connected to the same (secondary) network(s).

**Note:** the above are goals for the KubeVirt project, whose scope are VMs.
Achieving the same goals for pods (on the multus project) is a **requirement**
for KubeVirt.

## Non Goals
- Providing multiple connections to the primary cluster network (managed by
Kubernetes).
- Dynamic attachment / removal of SR-IOV (or macvtap) networks; since these rely
on a device plugin - and thus require the update of the pod's resources - i.e.
request the VF resource - something which, is not possible on a running pod.
This particular goal requires more study, and should be discussed in a separate
design document.
- Propagating networking changes / reconfiguration of existing interfaces;
the API is a bit misleading, since a user might want to think that updating an
existing interface is possible (since these are actually `NetworkSelectionElement`s
from the [multus api](https://github.com/k8snetworkplumbingwg/multus-cni/blob/dc9315f12549d70a9fa40a95a11bd8ea88b95577/pkg/types/types.go#L118)).

## Definition of Users
This feature is intended for cluster users who want to connect their existing
Virtual Machines to additional existing networks.
These users may - or may not - have permissions to create the aforementioned
additional networks; if so, they would need to rely on the cluster
administrator to provision it for them.

## User Stories
* as a user, I want to add a new network interface to a running VM.
* as a user, I want to remove an existing network interface from a running VM.

## Repos
- [CNAO](https://github.com/kubevirt/cluster-network-addons-operator)
- [KubeVirt](https://github.com/kubevirt/kubevirt)
- [Multus](https://github.com/k8snetworkplumbingwg/multus-cni)

# Design

## Multus
Multus - a CNI plugin - only handles the ADD / REMOVE verb, and is triggered
by kubelet only when the pod's sandbox is created - or removed. Given its
simplicity, it assumes no networks exist whenever it is executed, and procceeds
to call the ADD / DEL for **all** networks listed in its
`k8s.v1.cni.cncf.io/networks` annotation.

As such, the first part of the design implementation must focus on multus; it
must be refactored to enable it to be triggered not only when the pod's sandbox
is created, but also on-demand - i.e. whenever the pod's
`k8s.v1.cni.cncf.io/networks` are updated.

To do that, a controller residing on a long lived process must be introduced.
An important detail to take into account is this controller **must** be
executed on the host's PID, network, and mount namespaces, in order to access
whatever resources the delegate CNI requires which are available in the host
file system (e.g. log files, sockets, etc). The alternative would be for the
multus configuration to somehow know which resources to bind mount into the pod,
which, in my opinion is not reasonable without major changes in multus - we'd
need some discovery mechanism between multus & the CNI plugins.

There are multiple alternatives to run the controller process, each of which
with advantages / disadvantages:
1. the controller simply listens for pod network annotation updates, and spawns
   a new process (the "thin" CNI plugin) that runs on the host's PID, net, and
   mount namespaces.

   This alternative **requires** multus to share the PID namespace of the host,
   but only shares the host's mount namespace on the CNI process running the CNI
   binary - the controller part runs in its own mount namespace.
2. multus is re-architected as a thick plugin. This could achieved - for
   instance - by bind mounting the host's filesystem as a volume in the multus
   pod.

   While this solution does **not** require the pod to share the host's PID
   namespace, it does grant the controller side access to the host's
   filesystem.

   This solution requires each delegate CNI's configuration to be updated,
   pointing at the mounted path, rather than the path on the host - something
   that is dependent of each individual plugin's implementation.

   This solution also implies a larger attack surface on multus, since it is a
   privileged container with access to the host's filesystem.

### Simple controller
This controller would just start a new process on the host's mount namespace.

To achieve it, the following code snippet could be explored:

```golang
hostMountNamespace := "/proc/1/ns/mnt"
fd, err := os.Open(hostMountNamespace)
if err != nil {
    return fmt.Errorf("failed to open mount namespace: %v", err)
}
defer fd.Close()

if err = unix.Unshare(unix.CLONE_NEWNS); err != nil {
    return fmt.Errorf("failed to detach from parent mount namespace: %v", err)
}
if err := unix.Setns(int(fd.Fd()), unix.CLONE_NEWNS); err != nil {
    return fmt.Errorf("failed to join the mount namespace: %v", err)
}

const cniBinDir = "/opt/cni/bin/"

netCont := ...
runtimeConfig := ...

cniDriver := libcni.NewCNIConfig([]string{cniBinDir}, nil)
cniDriver.AddNetwork(context.Background(), netConf, runtimeConfig)
...
```

### Thickening multus
A "thin CNI plugin" runs as a one-shot process, typically as a binary on disk
executed on a Kubernetes host machine.

A "thick CNI Plugin", on the other hand, is a CNI component composed of two (or
more) parts, usually composed of "shim", and a long lived process (daemon)
resident in memory. The "shim" is a lightweight "thin CNI plugin" component that
simply passes CNI parameters (such as JSON configuration, and environment
variables) to the daemon component, which then processes the CNI request.

To transform multus into a thick plugin, it is needed to instantiate a long
lived process - which will be the multus pod entrypoint - listening to a unix
domain socket - this socket must be available both in the multus pod and the
hosts's mount namespaces; as such, a bind mount to host this socket must be
provided for the multus pod.

The CNI shim will then be invoked by kubelet, and send the CNI ADD/DELETE
commands to the server side of multus via the unix domain socket previously
mentioned. The multus daemon would also contact the shim whenever a new network
is to be added - or removed.

As previously mentioned in the [multus section](#multus), this approach
requires the full host's filesystem to be bind mounted into the multus pod,
and each of the delate CNI plugins to use these bind mounted paths instead of
whatever their defaults are. This could be inconvenient - at best - or, break
under plugins having an hard-coded path.

The upside of this approach, is that it seems to not require to share the
host's PID namespace.

Refer to the sequence diagram below for more information:

![Create pod flow for thick plugin](create-pod-flow.png)

### Further multus requirements
Multus must also be updated to act only on specific networks (the ones being
added or removed) rather than the full annotation list.

Another conflicting behavior of multus is its interface naming scheme; it
iterates all networks in the `k8s.v1.cni.cncf.io/networks` annotation, and for
each creates an interface in the pod named `netX`, where `X=<net index> + 1`.
The simplest way to work-around this limitation is to explicitly define the
interface names being plugged into the pod, as documented in the
[multus documentation](https://github.com/k8snetworkplumbingwg/multus-cni/blob/master/docs/how-to-use.md#launch-pod-with-text-annotation-with-interface-name).

At a later stage, multus will be patched to take into account the existing
interfaces when hot-plugging new ones. The existing interfaces can be derived
from the `k8s.v1.cni.cncf.io/network-status` annotation on the pod.

Finally, and to also account for the multus users who run it under severe
memory and CPU contraints, the "thin" plugin architecture must also be
preserved, and the user should be able to choose the type of plugin via
configuration. This means the multus maintainers will actively maintain the
two different alternatives.

To implement the aforementioned requirement, we plan on:
- build two **different** binaries: one for the `thin` plugin architecture,
  another for the `thick` plugin architecture.
- provide two different multus daemonset specs - one whose entrypoint is the
  multus daemon, which be use to provide the `thick` plugin alternative,
  another whose entrypoint is the bash script currently used - which'll be
  used on the `thin` plugin.
- both binaries will be shipped in the multus images.

#### Annotation driven - controller reacting to pod updates
This implementation relies on a control loop added to the multus server part.
It will listen to pod updates, and react when the `k8s.v1.cni.cncf.io/networks`
have changed. It will then issue ADD operations for each new network, and DELETE
operations for networks that were removed, to the responsible routine, within
the same process.

Refer to the sequence diagram below for more information:

![Kubernetes native solution](controller-based-approach.png)

This alternative seems to be better aligned to Kubernetes patterns, requires
less code, is aligned with
[Kaloom's multus fork](https://github.com/kaloom/kubernetes-kactus-cni-plugin/#how-the-podagent-communicate-the-additiondeletion-of-a-network-attachment-into-a-running-pod), and finally, also follows
[KubeVirt's razor](https://github.com/kubevirt/kubevirt/blob/main/docs/architecture.md#the-razor),
since this solution also allows adding and removing network interfaces to/from
pods.

In order to help troubleshooting when things go wrong, we should introduce two
annotations to the `k8s.v1.cni.cncf.io` namespace: one for tracking the
revision number of the `networks` annotation, another to track the revision
number of the `network-status` annotation.

These will be helpful to figure out if the resources are out of sync - e.g. the
desired state and the current state have not converged yet.

## KubeVirt
In order to mimic the disk hotplug feature, the proposed API also follows the
notion of a sub-resource, which will be used to trigger changes in the VM, the
VMI, or both - by sending an HTTP PUT to the correct URL endpoint.
This also enables a simple - and coherent - integration with `virtctl`, which
would help preserve the user's expectations of how the system works.

Furthermore, by design, KubeVirt only allows the update of the VMI spec by the
Kubevirt service accounts, thus via the VMI subresource path.

The proposed API changes for VM objects can be seen in
[the VM API examples section](#vms), while the proposed API changes for the VMI
object can be seen in [the VMI API examples section](#vmis).

## VMI flows

### virtctl
Two new commands will be introduced to the `virtctl` root command:
`addinterface`, and `removeinterface`, which result in an HTTP PUT being sent
to the corresponding VM/VMI subresource, on `virt-api`.

Refer to the image below for more information.

![Update VMI subresource](patching-vmi-annotations.png)

**NOTE:** this step is common for both VM / VMI objects. When used for a VM,
the user should provide the `--persist` flag.

### virt-api
The `virt-api` subresource handlers will then proceed to patch the VMI spec
`spec.domain.devices.interfaces`, and `spec.networks`.

### virt-controller
A VMI update will be trigered in virt-controller, during which we must patch
the `k8s.v1.cni.cncf.io/networks` annotation on the pod holding the VM, which
in turn causes multus to hotplug an interface into the pod.

The request to plug this newly created pod interface into the VM will then be
forwarded to the correct `virt-handler`.

### virt-handler
Finally, `KubeVirt`s agent in the node will create - and configure - any
required networking infrastructure, and finally tap into the correct
`virt-launcher`s namespaces to execute the commands required to hot plug / hot
unplug the network interfaces.

**NOTE:**  The feature is protected by the `HotplugInterfaces` feature gate.

## VM flows
The flows to patch up the VMI object are a subset of the steps required to
hot-plug an interface into a VM. This means that some extra initial steps are
required to update the corresponding VMI networks and interfaces specs, but
afterwards, the flows are common.

As with VMIs, it starts with issueing a `virtctl` command.

### virtctl
To hot-plug a new NIC into a running VMI, the user would execute the following
command:
```bash
$ virtctl addinterface <vmi-name> \
    --name <kubevirt-spec-iface-name> \
    --net <nad-name> \
    --iface-name <desired-pod-interface-name> \
    --persist
```

For hot-unplugging, use the `removeinterface` command instead.

### virt-api
The `virt-api` subresource handlers will then proceed to patch the VM status
with a `VirtualMachineInterfaceRequest`.

### virt-controller
The `virt-controller` will then see this status update and will proceed to
update the VM template with the added / removed interfaces, and all call the
VMI `AddInterface` endpoint described in the [VMI flows section](#vmi-flows).

## API Examples

### VMs
```golang
// VirtualMachineStatus represents the status returned by the
// controller to describe how the VirtualMachine is doing
type VirtualMachineStatus struct {
...
    InterfaceRequests []VirtualMachineInterfaceRequest `json:"interfaceRequests,omitempty" optional:"true"`
}

// VirtualMachineInterfaceRequest encapsulates all dynamic network operations on
// a VMI.
type VirtualMachineInterfaceRequest struct {
    // AddInterfaceOptions when set indicates an interface should be added.
    AddInterfaceOptions *AddInterfaceOptions `json:"addInterfaceOptions,omitempty" optional:"true"`

    // RemoveInterfaceOptions when set indicates an interface  should be removed.
    RemoveInterfaceOptions *RemoveInterfaceOptions `json:"removeInterfaceOptions,omitempty" optional:"true"`
}

// AddInterfaceOptions are used when hot plugging network interfaces
type AddInterfaceOptions struct {
    // Name is the name of the interface in the KubeVirt VMI spec
    Name string `json:”name”`

    // NetworkName is the name of the multus network
    NetworkName string `json:”networkName”`
}

// RemoveInterfaceOptions are used when hot unplugging network interfaces
type RemoveInterfaceOptions struct {
    // Name is the name of the interface in the KubeVirt VMI spec. Must match
    // the name of the network defined in the VMI spec.
    Name string `json:"name"`
}
```

**Note:** KubeVirt will **always** explicitly define the pod interface name for
multus-cni. It will be computed from the VMI spec interface name, to allow
multiple connections to the same multus provided network.

### VMIS
```golang
type VirtualMachineInstanceNetworkInterface struct {
...
    // If the interface is hotplugged, this will contain its status
    HotplugStatus *InterfaceHotplugStatus `json:"hotplugStatus,omitempty"`
}

type InterfaceHotplugStatus struct {
    // Phase are specific phases for the hotplug volume.
    Phase InterfaceHotplugPhase `json:"phase,omitempty"`

    // DetailedMessage has more information in case of failed operations.
    DetailedMessage string `json:"DetailedMessage,omitempty"`
}

type InterfaceHotplugPhase string

const InterfaceHotplugPhasePending    InterfaceHotplugPhase = "Pending"    # network configuration did not start yet
const InterfaceHotplugPhaseInfraReady InterfaceHotplugPhase = "InfraReady" # network configuration phase1 completed (networking infra created and configured)
const InterfaceHotplugPhaseReady      InterfaceHotplugPhase = "Ready"      # network configuration phase1 and phase2 completed (all is OK from the networking perspective)
const InterfaceHotplugPhaseFailed     InterfaceHotplugPhase = "Failed"     # plugging the new interface failed. More details on `DetailedMessage`.
```


### Hotplug for pods
Assuming the following two `network-attachment-definition`s:
```yaml
---
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf-1
spec:
  config: '{
            "cniVersion": "0.3.0",
            "type": "macvlan",
            "master": "eth1",
            "mode": "bridge",
            "ipam": {
                "type": "host-local",
                "ranges": [
                    [ {
                         "subnet": "10.10.0.0/16",
                         "rangeStart": "10.10.1.20",
                         "rangeEnd": "10.10.3.50",
                         "gateway": "10.10.0.254"
                    } ]
                ]
            }
        }'
---
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf-2
spec:
  config: '{
            "cniVersion": "0.3.0",
            "type": "macvlan",
            "master": "eth1",
            "mode": "bridge",
            "ipam": {
                "type": "host-local",
                "ranges": [
                    [ {
                         "subnet": "12.10.0.0/16",
                         "rangeStart": "12.10.1.20",
                         "rangeEnd": "12.10.3.50"
                    } ]
                ]
            }
        }'
```

And a pod with the following spec:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-case-03
  annotations:
    k8s.v1.cni.cncf.io/networks: macvlan-conf-1
spec:
  containers:
  - name: pod-case-03
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
```
Update the pod to:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-case-03
  annotations:
    k8s.v1.cni.cncf.io/networks: macvlan-conf-1,macvlan-conf-2
spec:
  containers:
  - name: pod-case-03
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
```

The aforementioned update will trigger multus to start the CNI ADD flow for the
network named `macvlan-conf-2`.

## Functional Testing Approach
Functional testing will use the network sig KubeVirt lanes -
`k8s-<x.y>-sig-network`. These lanes must be used since this feature is network
related, and, they have multus CNI installed in the cluster via CNAO.
In KubeVirt, a new test suite will be added, where the following tests will be
performed:

* plug a new NIC into a running VM
* unplug a NIC from a running VM (can be performed in the previous test
  teardown)
* migrate a VM having an hot-plugged interface

All these tests have as pre-requirements that the `HotplugInterfaces` feature
gate is enabled, **and** a secondary network provisioned.

### Multus functional tests
In multus, new functional tests must be added that cover the following
scenarios:
* plug a new network interface into a running pod
* plug multiple network interfaces - using the same attachment - into a running
  pod
* unplug a network interface from a running pod

# Implementation Phases
1. **M** Refactor multus as a thick cni plugin.
2. **M** Refactor multus to allow adding / removing only new networks
3. **C** Consume this updated multus functionality via CNAO
4. **K** Add the hot-plug / hot-unplug functionality to KubeVirt
5. **M** Take into account the existing interfaces when multus implicitly generates
   new interface names
6. **M** update the default route on the pod whenever the NIC holding the
   default route is hot-unplugged. It should now point to the cluster's default
   network - managed by Kubernetes.

**Notes:**
* the action items listed above have either `M`, `K`, or `C` to
indicate in which project should it be implemented.
* the MVP version would be composed of steps 1 through 4, inclusive.
