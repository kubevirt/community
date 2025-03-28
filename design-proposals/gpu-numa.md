# Overview

This proposal proposes a solution for implementing GPU NUMA topology mapping
for virtual machines based on KubeVirt.

## Motivation

Support associating NUMA nodes with GPU devices in KubeVirt VMs to optimize
computing performance in virtual machines.

## Goals

- Update KubeVirt's Libvirt schema to support VM GPU NUMA node setting.
- Add a feature gate to control the enabling and disabling of this
functionality.

## Non Goals

- Do nothing if VM CPU NUMA is not set.

## Definition of Users

- VM owners

## User Stories

Currently, KubeVirt does not support setting the GPU NUMA node. In multi-NUMA
virtual machine scenarios, the GPU devices in the created virtual machine are
directly attached to the root complex(pcie.0) without mapping the GPU NUMA
topology from the physical machine. This will result in degraded GPU
communication efficiency within the virtual machine, especially on machines
with 8 NUMA nodes.

## Repos

- [kubevirt](https://github.com/kubevirt/kubevirt)

# Design

We propose using emulating a `pcie-expander-bus (pxb-pcie)` in the VM to
configure NUMA node and expose a `pcie-root-port` for PCIe device (GPU devices)
to plug into, according to the [QEMU device placement strategy](
https://github.com/qemu/qemu/blob/master/docs/pcie.txt#L37-L74):

PCIe endpoint devices are not themselves associated with NUMA nodes, rather the
bus they are connected to has affinity. The root complex(pcie.0) is not
associated with any NUMA node, but extra `pcie-expander-bus (pxb-pcie)` can
be added and associated with a NUMA node.

It is not possible to plug PCIe endpoint devices directly into the
`pcie-expander-bus (pxb-pcie)`, so it is necessary to add `pcie-root-port` into
each `pcie-expander-bus (pxb-pcie)` - we will need one port per device to be
added.

The PCI/PCIe topology in the VM will be like this:

```
   pcie.0 bus
   ----------------------------------------------------------------------------------------------
        |                |                    |                  |                        |
   -----------   ------------------   -------------------   --------------         --------------
   | PCI Dev |   | PCIe Root Port |   | PCIe-PCI Bridge |   |  pxb-pcie  |         |  pxb-pcie  |
   -----------   ------------------   -------------------   | (set NUMA) |         | (set NUMA) |
                                                            --------------         --------------
                                                                 |                        |
                                                         --------------------    --------------------
                                                         |  PCIe Root Port  |    |  PCIe Root Port  |
                                                         --------------------    --------------------
                                                                 |                        |
                                                           ------------             ------------
                                                           | GPU Card |             | GPU Card |
                                                           ------------             ------------
```

Compared to directly providing a topology file, this approach more
accurately reflects the physical GPU NUMA relationship and is more friendly to
NUMA topology awareness because user do not need to manually obtain and depend
on a topology file.

To implement the VM PCI/PCIe topology proposed above, the libvirt XML
configuration should be set like this:

```xml
<controller type='pci' index='0' model='pcie-root'>
  <alias name='pcie.0'/>
</controller>
<controller type='pci' index='1' model='pcie-expander-bus'>
  <target busNr='180'>
    <node>0</node>
  </target>
  <alias name='pci.1'/>
  <address type='pci' domain='0x0000' bus='0x00' slot='0x02' function='0x0'/>
</controller>
<controller type='pci' index='2' model='pcie-root-port'>
  <target chassis='2' port='0x0'/>
  <alias name='pci.2'/>
  <address type='pci' domain='0x0000' bus='0x01' slot='0x00' function='0x0'/>
</controller>
<hostdev mode='subsystem' type='pci' managed='no'>
  <model type='virtio'/>
  <source>
    <address domain='0x0000' bus='0x27' slot='0x00' function='0x0'/>
  </source>
  <alias name='ua-gpu-gpu-1'/>
  <address type='pci' domain='0x0000' bus='0x02' slot='0x00' function='0x0'/>
</hostdev>
```

According to the [libvirt official documentation](
https://libvirt.org/pci-addresses.html#pcie-expander-bus), QEMU uses the `bus`
property of a device's PCI address only to match it with the PCI controller
that has the same `index` property, and not to set the actual PCI address,
which is decided by the guest OS.

So, by looking at the XML snippet above, we can see that the gpu device plugs
into the pcie-root-port controller, which plugs into the pcie-expander-bus
controller, which plugs into pcie-root: the guest OS sees the same topology,
but assigns different PCI addresses to some of its component.

KubeVirt is implemented based on Libvirt (QEMU), but

- the [XML schema used by KubeVirt](
https://github.com/kubevirt/kubevirt/blob/v1.4.0/pkg/virt-launcher/virtwrap/api/schema.go#L636-L643)
to render Libvirt virtual machines does not support NUMA node configuration.
- Only after the virtual machine pod (KubeVirt virt-launcher) is successfully
scheduled can we determine which GPU devices have been allocated and associate
them with the corresponding NUMA nodes.

Therefore, it is necessary to modify the KubeVirt source code to associate GPU
devices within the virtual machine with the corresponding NUMA nodes.

```go
// pkg/virt-launcher/virtwrap/api/schema.go
type Controller struct {
	Type    string            `xml:"type,attr"`
	Index   string            `xml:"index,attr"`
	Model   string            `xml:"model,attr,omitempty"`
	Driver  *ControllerDriver `xml:"driver,omitempty"`
	Alias   *Alias            `xml:"alias,omitempty"`
	Address *Address          `xml:"address,omitempty"`
	Target  *ControllerTarget `xml:"target,omitempty"`  // Add struct `ControllerTarget`
}

type ControllerTarget struct {
	Node *uint32 `xml:"node,omitempty"`
}
```

In the virtual machine pod (KubeVirt virt-launcher), another main question is
how to determine the associated NUMA node within the virtual machine based on
the PCI numbers of the allocated physical GPU devices. Here are the steps:

- `/sys/bus/pci/devices/${GPU_PCI}/numa_node` shows the NUMA node in the
physical machine.

```bash
cat /sys/bus/pci/devices/0000\:60\:00.0/numa_node
0
```

- On each K8S node, the init container of `virt-handler` (a Kubernetes
DaemonSet) executes the script
`cmd/virt-launcher/node-labeller/node-labeller.sh`. This script collects the
full NUMA topology information of the physical machine node based on
`virsh capabilities`. The collected data is stored in
`/var/lib/kubevirt-node-labeller/capabilities.xml` and mounted into the main
`virt-handler` container via hostPath. Each time `virt-handler` creates a
virtual machine, it passes `capabilities.xml` as a gRPC request parameter to
`virt-launcher`.`/var/lib/kubevirt-node-labeller/capabilities.xml` records the
association between the NUMA node and CPU in the physical machine.

```xml
...
    <topology>
      <cells num='8'>
        <cell id='0'>
          <cpus num='32'>
            <cpu id='0' socket_id='0' die_id='0' core_id='0' siblings='0,128'/>
            <cpu id='1' socket_id='0' die_id='0' core_id='1' siblings='1,129'/>
...
```

- The virt-launcher pod will determine the physical CPUs allocated to it based
on its own cgroup settings. Then, using the full CPU topology information from
the `capabilities.xml` file above, it will render the vCPU topology inside the
virtual machine. The mapping between physical CPUs and vCPUs is stored in the
following memory data structure: [virt-launcher CPUTune](
https://github.com/kubevirt/kubevirt/blob/v1.4.0/pkg/virt-launcher/virtwrap/api/schema.go#L218)

```xml
<domain type="kvm" xmlns:qemu="http://libvirt.org/schemas/domain/qemu/1.0">
    <cputune>
        <vcpupin vcpu="0" cpuset="1"></vcpupin>
        <vcpupin vcpu="1" cpuset="65"></vcpupin>
...
```

- Similarly, the association between vCPUs and NUMA node in the VM is stored in
the following memory data structure:[virt-launcher NUMA](
https://github.com/kubevirt/kubevirt/blob/v1.4.0/pkg/virt-launcher/virtwrap/api/schema.go#L281)

```xml
<domain type="kvm" xmlns:qemu="http://libvirt.org/schemas/domain/qemu/1.0">
    <cpu mode="host-model">
        <topology sockets="1" cores="16" threads="1"></topology>
        <numa>
            <cell id="0" cpus="0,1,2,3,4,5,6,7" memory="32212254720" unit="b"></cell>
            <cell id="1" cpus="40,41,42,43,44,45,46,47" memory="32212254720" unit="b"></cell>
        </numa>
    </cpu>
...
```

## API Examples

Based on the proposed changes above, when we create a virtual machine with the
following configuration, the NUMA mapping functionality can be achieved.

First, open the new feature gate in the KubeVirt CR:
```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - GPUDeviceNUMA # the feature gate to control the enabling and disabling of this functionality
```

Second, create a VM with CPU/memory NUMA settings, GPU NUMA will be automatically
set inside the VM:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: testvm1
spec:
  runStrategy: Once
  template:
    spec:
      domain:
        cpu:
          dedicatedCpuPlacement: true   # CPU NUMA settings
          numa:                         # (KubeVirt already supported)
            guestMappingPassthrough: {} #
        memory:              # memory NUMA settings
          hugepages:         # (KubeVirt already supported)
            pageSize: "2Mi"  #
```

## Scalability

No impact.

## Security

No impact.

## Update/Rollback Compatibility

The feature gate `GPUDeviceNUMA` can control the enabling and disabling of
this functionality.

## Functional Testing Approach

- Unit tests will cover the new function.
- Create a VM with GPU NUMA:
  - A correct CPU/memory/GPU NUMA topology mapping inside VM.
  - An [NCCL-tests](https://github.com/NVIDIA/nccl-tests) evaluation to verify the performance changes of the VM.

# Implementation Phases

- [ ] Implement associating NUMA nodes with GPU devices in KubeVirt VMs.
- [ ] Functional Testing.
- [ ] Upstream `gpu-numa`.