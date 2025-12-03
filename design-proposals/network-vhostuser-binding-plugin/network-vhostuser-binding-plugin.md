# Overview
`vhostuser` interfaces are supported by qemu but not implemented in Kubevirt. Network Binding Plugin is a good framework to add support for `vhostuser` interfaces to Kubervirt. 

## Motivation
`vhostuser` interfaces are required to attach VMs to a userspace dataplane such as OVS-DPDK or VPP and achieve a fast datapath from the VM to the physical NIC.
This is a mandatory feature for networking VMs such as vRouter, IPSEC gateways, firewall or SD-WAN VNFs, that usually bind the network interfaces using DPDK. Expected performance with DPDK can only be met if the whole datapath is userspace and not go through kernel interfaces like with usual bridge interfaces.

## Goals
Be able to add `vhostuser` secondary interfaces to the VM definition in Kubevirt.

## Non Goals
The `vhostuser` secondary interfaces configuration in the dataplane is under the responsibility of Multus and the CNI such as `userspace CNI`.

## Definition of Users
- **VM User** is the persona that configures `VirtualMachine` or `VirtualMachineInstance`
- **Cluster Admin** is the persona that configures `KuberVirt` resources
- **Network Binding Plugin Developer** is the persona that implements the `network-vhostuser-binding` plugin
- **CNI Developer** is the persona that implements the CNI that configures the dataplane with vhostuser sockets
- **Dataplane Developer** is the persona that implements the userspace dataplane

## User Stories
- As a VM User, I want to create a VM with one or serveral `vhostuser` interfaces attached to a userspace dataplane.
- As a VM User, I want the `vhostuser` interface to be configured with a specific MAC address.
- As a VM User, I want to enable multi-queue on the `vhostuser` interface
- As a VM User, I want to be able to configure the `vhostuser` interface as transitional
- As a Cluster Admin, I want to be able to enable `network-vhostuser-binding`
- As a Network Binding Plugin Developer, I want the shared socket path to be accessible to `virt-launcher` pod 
- As a Dataplane Developer, I want to access all `vhostuser` sockets of VM pods
- As a CNI Developer, I want to know where vhostuser sockets are located
 
## Repos
As far as possible, `network-vhostuser-binding` could be hosted in Kubevirt repo, and most specificaly in [cmd/sidecars](https://github.com/kubevirt/kubevirt/tree/main/cmd/sidecars).  
However, if Kubevirt repo is not meant to host other plugins than reference ones, `network-vhostuser-binding` could be hosted out-of-tree.


## Design
This proposal leverages the KubeVirt Network Binding Plugin sidecar framework to implement a new `network-vhostuser-binding-plugin`.

`network-vhostuser-binding-plugin` role is to implement the modification to the domain XML according to the VMI definition passed through its gRPC service by the `virt-launcher` pod on `OnDefineDomain` event from `virt-handler`.

`vhostuser` interfaces are defined in the VMI under `spec/domain/devices/interfaces` using the binding name `vhostuser`:

```yaml
spec:
  domain:
    devices:
      networkInterfaceMultiqueue: true
      interfaces:
      - name: default
        masquerade: {}
      - name: net1
        binding:
          name: vhostuser
        macAddress: ca:fe:ca:fe:42:42
```

`network-vhostuser-binding` translates the VMI definition into libvirt domain XML modifications on `OnDefineDomain`:
1. Creates a new interface with `type='vhostuser'`
2. Set the MAC address if specified in the VMI spec
3. Define model type according to `useVirtioTransitional` VMI spec
4. If `networkInterfaceMultiqueue` is set to `true`, add the number of queues calculated after the number of cores of the VMI
5. Add `memAccess='shared'` to all NUMA cells elements
6. Define the device name according to Kubevirt naming schema
7. Define the `vhostuser` socket path, immutable accross Live Migration

As `OnDefineDomain` hook can be called multiple times by KubeVirt, `network-vhostuser-binding` modification must be idempotent.

Below is an example of modified domain XML:

```xml
<cpu mode="host-model">
        <topology sockets="2" cores="8" threads="1"></topology>
        <numa>
            <cell id="0" cpus="0-7" memory="2097152" unit="KiB" memAccess="shared"/>
            <cell id="1" cpus="8-15" memory="2097152" unit="KiB" memAccess="shared"/>
        </numa>
</cpu>
<interface type='vhostuser'>
    <source type='unix' path='/var/run/kubevirt/vhostuser/net1/poda08a0fcbdea' mode='server'/>
    <target dev='poda08a0fcbdea'/>
    <model type='virtio-non-transitional'/>
    <mac address='ca:fe:ca:fe:42:42'/>
    <driver name='vhost' queues='8' rx_queue_size='1024' tx_queue_size='1024'/>
    <alias name='ua-net1'/>
</interface>
```

### Implementation details

The socket path have to be available to both `virt-launcher` pod (and `compute` container) and dataplane pod.  
In order to not use hostPath volumes that requires pod to be privileged, we propose to implement a **vhostuser Device Plugin** that will be able to inject mounts to the sockets directory into unprivileged pods, and annotations.

#### Device Plugin for **vhostuser sockets** resources

Device plugins can instructs kubelet to add mounts into the containers when managed resources are requested.

This design proposal relies on a device plugin that would manage two kinds of resources on the userspace dataplane that we can think of a virtual switch:
- **dataplane**: `1`  
  This resource give access to all sub directories of `/var/run/vhostuser`, and to sockets inside.  
  It is requested by the dataplane itself.  
  Kubelet injects `/var/run/vhostuser` mount in the container.
- **vhostuser sockets**: `n`  
  This resource can be thought as a virtual switch port, and can have a limit related to dataplane own limitation (performance, CPU, etc.).  
  It can help schedule workloads on node where dataplane has available resources.  
  It is requested through VM or VMI definition in resources request spec. In turn the `compute` container of the `virt-launcher` pod will request the same resources.  
  This makes the device plugin allocates a sub directory `/var/run/vhostuser/<socketXX>`, and mount it into the `virt-launcher` pod.

The device plugin has to comply with [`device-info-spec`](https://github.com/k8snetworkplumbingwg/device-info-spec/blob/main/SPEC.md#device-information-specification). This allows information sharing between device plugin and the CNI. Thanks to Multus being compliant with this spec, the CNI can retrieve device information (socket path and and type) to be used to configure the dataplane accordingly. Multus will  annotate the `virt-launcher` pod with this information, KubeVirt extracts only a part into `kubevirt.io/network-info`.
   
The device plugin has to care about directory permissions and SELinux, for the sockets to be accessible from requesting pods.

#### Network Binding Plugin and Kubevirt requirements

Network Binding Plugin then can leverage `downwardAPI` feature available from Kubevirt v1.3.0, in order to retrieve the `kubevirt.io/network-info` annotation values, and extract the socket path to configure the interface in the domain XML.

But it can't use it directly as it would break Live Migration of VMs:   
The socket directories `/var/run/vhostuser/<socketXX>` are not predictable, and new ones get allocated when the destination pod is being created.  
Unfortunately the domain XML is the one from the source pod (migration domain), and references sockets paths allocated to source pod.

Hence, Network Binding Plugin needs to use immutable paths to sockets. This can be achieved using the interface name (or its hash version) in symbolic links to the real socket path: `/var/run/kubevirt/vhostuser/net1` -> `/var/run/vhostuser/<socketXX>`.

This requires an enhancement in KubeVirt in order for `virt-launcher` pod to have a shared `emptyDir` volume, mounted in both `compute` and `vhostuser-network-binding-plugin` containers. This `emptyDir` path can be either:
- a paramater in Network Binding Plugin KubeVirt CRD spec, but would require a spec evolution.
- a well-known fixed path for this `emptyDir`, that may be used by any binding plugin.

It appears that a well-known fixed path is sufficient and can be simpler to implement.


#### Implementation diagram

![kubevirt-vhostuser-shared-sockets](kubevirt-vhostuser-binding-plugin-device-plugin.drawio.png)

## API Examples

### KubeVirt CRD

No modification tothe Network Binding Plugin spec of the KubeVirt CR is necessary as we will use a well-known fixed path for the shared `emptyDir` volume.

### No modification to VM

Example of a `VirtualMachine` definition using `network-vhostuser-binding` plugin and device plugin resources requests:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vhostuser-vm
  namespace: tests
spec:
  running: true
  template:
    metadata:
      labels:
        kubevirt.io/domain: vhostuser-vm
    spec:
      architecture: amd64
      domain:
        cpu:
          cores: 4
        devices:
          disks:
          - disk:
              bus: virtio
            name: containerdisk
          interfaces:
          - masquerade: {}
            name: default
          - binding:
              name: vhostuser
            macAddress: ca:fe:ca:fe:42:42
            name: net1
          networkInterfaceMultiqueue: true
        machine:
          type: q35
        memory:
          hugepages:
            pageSize: 1Gi
        resources:
          limits:
            vhostuser/sockets: 1
          requests:
            memory: 2Gi
            vhostuser/sockets: 1
      networks:
      - name: default
        pod: {}
      - multus:
          networkName: vhostuser-network
        name: net1
      nodeSelector:
        node-class: dpdk
      volumes:
      - containerDisk:
          image: os-container-disk-40g
        name: containerdisk
```

## Scalability
(overview of how the design scales)

## Update/Rollback Compatibility
Kubevirt Network Binding plugin relies on `hooks/v1alpha3` API for a clean termination of the `network-vhostuser-binding` container in the virt-launcher pod.

## Functional Testing Approach
Create a VM with several `vhostuser` interfaces then:
- check the generated domain XML contains all interfaces with appropriate configuration
- check the vhostuser sockets are created in the expected directory of virt-launcher pod
- check the vhostuser sockets are available to the dataplane pod
- check the VM is running
- check VM network connectivity
- live migrate the VM
- check the VM is migrated and is running
- check VM network connectivity

# Implementation Phases
- [ ] Implement network binding plugin sharedDir spec in KubeVirt 
- [x] First implementation of the `network-vhostuser-binding`
- [x] Implement vhostuser device plugin, based on [generic-device-plugin](https://github.com/squat/generic-device-plugingeneric-device-plugin)
- [ ] Upstream `network-vhostuser-binding`
