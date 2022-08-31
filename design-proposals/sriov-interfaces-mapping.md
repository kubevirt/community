# Overview
When creating a VM with SR-IOV interfaces, the VFs from the same resource pool are not differentiated by KubeVirt and therefore potentially wrongly plugged to the domain. 
</br>
This design proposes to fix the current mapping algorithm, using information from multus annotations.

## Motivation
[KubeVirt’s SR-IOV support](https://kubevirt.io/user-guide/virtual_machines/interfaces_and_networks/#sriov) enables wiring up SR-IOV Virtual Functions (VF) to VMs
by passing its PCI device ID from the node, through the virt-launcher pod, and then into the VM using [Libvirt](https://libvirt.org/) API.
<br/>
Each VM SR-IOV device (i.e: VF) is associated with a VM secondary network.

When a VM is defined with multiple SR-IOV devices from the same [resource pool](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin#configurations)
or connected to the same `NetworkAttachmentDefinition`, Kubevirt cannot differentiate which VF is associated with which requested device.
The root cause is that the PCI information passed by the sriov-network-device-plugin (through kubelet) is not enough to directly map the VFs with their secondary networks.

The outcome is a non-consistent association between the underlying VFs and the VM networks,
leaving the VM in a state where the VFs are plugged into the domain wrongly.
<br/>
For example:
<br/> 
When a SR-IOV interface is defined with a custom MAC address (e.g: `02:00:00:00:00`), a custom guest PCI address of (e.g: `0000:01:01.1`) and pointing to a `NetworkAttachmentDefinition` with VLAN configuration (e.g: 100).
<br/>
It is not guaranteed that the correct VF (with VLAN 100) will be plugged into the domain on the specified PCI address (`0000:01:01.1`).

To clarify, in case there is only one SR-IOV interface it will be mapped correctly as there is a 1:1 correlation, but the more interfaces there are it is more likely that the mapping will be wrong.

This issue also manifests when the VM interface boot order is set, the VM may end up booting from the wrong interface.

Moreover, from the user's side, this issue will also have the unfortunate side of effect of reflecting the wrong network-attachment-definition used in the VM status. <br/>

Kubevirt should plug the correct VF with the correct properties into the domain,
and the user inside the guest should be able to predict and assume a consistent SR-IOV NIC setup.

## Goal
Accurately associate VM's SR-IOV interfaces with the underlying SR-IOV Virtual Functions (VF) devices (PCI) allocated by the [sriov-network-device-plugin](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin#sr-iov-network-device-plugin) and [sriov-cni](https://github.com/k8snetworkplumbingwg/sriov-cni).

## Non-Goals
Fix/change third-party components (e.g: Multus, SR-IOV CNI, etc..)

## Existing Implementation Flow
In order to discover and allocate VFs to Pods the [sriov-network-device-plugin](https://github.com/k8snetworkplumbingwg/sriov-network-device-plugin#sr-iov-network-device-plugin) (from now on sriov-dp) and [sriov-cni](https://github.com/k8snetworkplumbingwg/sriov-cni) are being used.
<br/>
The sriov-dp provides an API to create resource pools of VFs according to various properties [[1]](https://github.com/openshid/sriov-network-device-plugin#configurations), and each resource pool is labeled with a resource name.

In the VMI spec, secondary networks point to the desired SR-IOV `NetworkAttachmentDefinition` object.
<br/>
The `NetworkAttachmentDefinition` is set with the resource-name annotation:
<br/>`k8s.v1.cni.cncf.io/resourceName=<resource name>`<br/>
which later on [passed down by Multus](https://github.com/k8snetworkplumbingwg/multus-cni/blob/a28f5cb56c79a582f5ea2b35a61b38f34b937930/examples/README.md#passing-down-device-information) to the sriov-cni.
```yaml
kind: VirtualMachine
  ...
spec:
  domain:
    devices:
      interfaces:
      - name: sriovnet1
        sriov: {}
    ...
  networks:
  - name: sriovnet1
    multus:
      networkName: default/sriov-network-vlan100
  ...
```
```yaml
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-network-vlan100
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: kubevirt.io/sriov_net
spec:
  config: |-
    {
        "type":"sriov",
  ...
```

As part of Pod creation flow, kubelet executes the required device plugins according to the Pod spec, in our case the sriov-dp.
<br/>
The sriov-dp sets the `PCIDEVICE` environment variable inside the pod which indicates the allocated device PCI address and its resource pool.
<br/>
For example:
<br/>
Given the resource name `kubevirt.io/sriov_net` the following environment variable is set:
<br/>
`PCIDEVICE_KUBEVIRT_IO_SRIOV_NET=<PCI address>`.

As part of the VM create flow, virt-controller renders the given VMI spec and creates the corresponding virt-launcher Pod [[3]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-controller/services/template.go#L443) (renderLaunchManifest).
<br/>
For each secondary network:
- Realize the device `resourceName` by fetching the specified `NetworkAttachmentDefinition`.
- Add the following environment variable to the compute container (spec.Env):<br/>`KUBEVIRT_RESOURCE_NAME_<network name>=<resource name>`

Once virt-launcher Pod is ready, virt-handler process the VMI object and eventually trigger the virt-launcher to synchronize with the new spec (i.e: virt-handler sends gRPC call to virt-launcher [[4]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/manager.go#L822)).
<br/>
virt-launcher renders the VMI spec, converts it to Libvirt domain XML and eventually passes it to Libvirt, which in turn creates the domain.

As part of the VMI spec rendering, virt-launcher realizes the VFs PCI addresses that been allocated by loading `KUBEVIRT_RESOURCE_NAME` (formerly created by the virt-controller) and `PCIDEVICE` (created by the sriov-dp) environment variables.

Each SR-IOV device is represented by the Hostdev device type in the Libvirt domain XML [[5]](https://wiki.libvirt.org/page/Networking#PCI_Passthrough_of_host_network_devices).

### Example
Given the following VMI spec:
```yaml
kind: VirtualMachineInstance
metadata:
  name: sriovvmi1
  ...
spec:
  domain:
    devices:
      interfaces:
      - name: sriovnet1
        sriov: {}
    ...
  networks:
  - name: sriovnet1
    multus:
      networkName: default/sriov-network
  ...
```

The following environment variables will be present in the Pod:
```bash
# kubectl exec -it virt-launcher-sriovvmi1-jpk58 -- env
...
PCIDEVICE_KUBEVIRT_IO_SRIOV_NET=0000:04:0a.3
...
KUBEVIRT_RESOURCE_NAME_sriovnet1=kubevirt.io/sriov_net
```

`sriovnet1` network is represented by Hostdev device type in the domain XML like so:
```bash
# kubectl exec -it virt-launcher-sriovvmi1-jpk58 -- virsh dumpxml 1
...
<devices>
...
    <hostdev mode='subsystem' type='pci' managed='no'>
        <alias name='ua-sriov-sriovnet1'/>    
        <driver name='vfio'/>
        <address type='pci' domain='0x0000' bus='0x06' slot='0x00' function='0x0'/>
        <source>
            <address domain='0x0000' bus='0x04' slot='0x0a' function='0x3'/>
        </source>
    </hostdev>
...

```

## The Problem
The VFs PCI address information that being passed down to virt-launcher through `PCIDEVICE` (by sriov-device-plugin) and `KUBEVIRT_RESOURCE_NAME` (by Kubevirt) environment variables, is not enough to map the VFs with their secondary network.
In the scenario where a VM has more than one network that points to a `NetworkAttachmentDefinition` with the same resource name, VFs might get assigned to the wrong network.

For example:
In the scenario described in https://github.com/kubevirt/kubevirt/issues/6351,
there are two `SriovNetwork` objects - each configures a different VLAN [[1]](https://github.com/kubevirt/kubevirt/issues/6351#issuecomment-918467186).
But the VM's networks point to the wrong VFs:
VM network `sriov-vsrx-if30` points to `default/sriov-vsrx-if30` NetworkAttachmentDefinition, which is configured with VLAN 30, but the underlying VF is configured with a different VLAN (i.e. the wrong PCI address was chosen).

The issue manifests on Kubevirt as follows:
<br/>
The SR-IOV network PCI address is determined on virt-launcher code that converts the VMI spec to Libvirt domain.<br/>
Libvirt domain Hostdev is represented in Kubevirt code by the HostDevice object.<br/>
For each SR-IOV network a corresponding HostDevice object is created, the PCI address is assigned as follows:
- Create SR-IOV host devices [[2]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/manager.go#L798)  [[3]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/hostdev.go#L35)  [[4]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/hostdev.go#L38).
- Map between network-name to resource-name [[5]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/pcipool.go#L40-L44)  [[6]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/pcipool.go#L49-L59).
- Map between resource-name and PCI addresses [[7]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/pcipool.go#L45)  [[8]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/sriov/pcipool.go#L61-L67)  [[9]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/addresspool.go#L35-L61).
- Allocate VF PCI address for each VM network host device [[10]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/device/hostdevice/hostdev.go#L64).

### Example 1: Two different secondary networks pointing to different NetworkAttachmentDefinitions with the same resourceName
1. VMI spec with two SR-IOV devices, each connected to different network:
```yaml
kind: VirtualMachineInstance
metadata:
  name: sriovvmi2
...
spec:
  domain:
    devices:
      interfaces:
      - name: sriovnet-vlan100
        sriov: {}
      - name: sriovnet-vlan200
        sriov: {}
...
networks:
- multus:
    networkName: default/sriov-network-vlan100
    name: sriovnet-vlan100
- multus:
    networkName: default/sriov-network-vlan200
    name: sriovnet-vlan200
  ...
```

2. Two `NetworkAttachmentDefinition` objects pointing to the same resource:
```yaml
---
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-network-vlan100
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: kubevirt.io/sriov_net
spec:
  config: |-
    {
      "cniVersion":"0.3.1",
      "name":"sriov-network-vlan100",
      "type":"sriov",
      "vlan":100,
      "spoofchk":"on",
      "trust":"off",
      "vlanQoS":0,
      "link_state":"enable",
      "ipam":{}
    }
---
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-network-vlan200
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: kubevirt.io/sriov_net
spec:
  config: |-
    {
      "cniVersion":"0.3.1",
      "name":"sriov-network-vlan200",
      "type":"sriov",
      "vlan":200,
      "spoofchk":"on",
      "trust":"off",
      "vlanQoS":0,
      "link_state":"enable",
      "ipam":{}
  }
```

3. The following environment variables will be present in the pod:
```bash
# kubectl exec virt-launcher-sriovvmi2-wcszx -- env
...
PCIDEVICE_KUBEVIRT_IO_SRIOV_NET=0000:04:02.4,0000:04:02.5
...
KUBEVIRT_RESOURCE_NAME_sriovnet-vlan100=kubevirt.io/sriov_net
KUBEVIRT_RESOURCE_NAME_sriovnet-vlan200=kubevirt.io/sriov_net
```

4. `sriovnet-vlan100` and `sriovnet-vlan200` networks will be represented by a Hostdev in the domain XML:
```bash
# kubectl exec -it virt-launcher-sriovvmi2-wcszx -- virsh dumpxml 1
...
  <hostdev mode='subsystem' type='pci' managed='no'>
    <alias name='ua-sriov-sriovnet-vlan100'/>
    <driver name='vfio'/>
    <source>
          <address domain='0x0000' bus='0x04' slot='0x02' function='0x4'/>
    </source>
    <address type='pci' domain='0x0000' bus='0x06' slot='0x00' function='0x0'/>
  </hostdev>
  <hostdev mode='subsystem' type='pci' managed='no'>
    <alias name='ua-sriov-sriovnet-vlan200'/>
    <driver name='vfio'/>
    <source>
          <address domain='0x0000' bus='0x04' slot='0x02' function='0x5'/>
    </source>
    <address type='pci' domain='0x0000' bus='0x07' slot='0x00' function='0x0'/>
  </hostdev>
...
```

#### Summary
| SR-IOV VFs device pool name | `kubevirt.io/sriov_net`        |
|:----------------------------|:-------------------------------|
| VM networks names           | `sriovnet100`, `sriovnet200`   |
| Allocated VFs PCI addresses | `0000:04:03.6`, `0000:04:03.7` |

When `HostDevice` object is created, the PCI address is picked as follows:
<br/>
`networkToResource["sriovnet100"]`  ----------------------------> `kubevirt.io/sriov_net`
<br/>
`addressesByResource["kubevirt.io/sriov_net"]` ---------> `0000:04:03.6`

There is no guarantee that `0000:04:03.6` is the correct PCI address for the network `sriovnet100`  and that the VF is configured with VLAN 100.
This example shows that the current environment variable `PCIDEVICE` (given by the sriov-dp) is simply not enough to directly map between the VM’s networks and  the VFs PCI address.

### Example 2: Two VM networks  connected to the same `NetworkAttachmentDefinition`
VMI spec with two SR-IOV devices, each connected to different networks, that are attached to the `NetworkAttachmentDefinition` (different MAC addresses):

```yaml
kind: VirtualMachineInstance
metadata:
  name: sriovvmi2
...
spec:
  domain:
    devices:
      interfaces:
      - name: sriovnet-vlan100-primary-mac
        macAddress: aa:bb:cc:dd:ee:01
        sriov: {}
      - name: sriovnet-vlan100-secondary-mac
        macAddress: aa:bb:cc:dd:ee:02
        sriov: {}
      ...
  networks:
  - multus:
      networkName: default/sriov-network-vlan100
    name: sriovnet-vlan100-primary-mac
  - multus:
      networkName: default/sriov-network-vlan100
    name: sriovnet-vlan100-secondary-mac
  ...
```

Single `NetworkAttachmentDefinition`:
```yaml
---
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-network-vlan100
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: kubevirt.io/sriov_net
spec:
  config: |-
    {
        "cniVersion":"0.3.1",
        "name":"sriov-network-vlan100",
        "type":"sriov",
        "vlan":100,
        "spoofchk":"on",
        "trust":"off",
        "vlanQoS":0,
        "link_state":"enable",
        "ipam":{}
    }
```

The following environment variables will be present in the pod:
```bash
# kubectl exec virt-launcher-sriovvmi2-wcszx -- env
...
PCIDEVICE_KUBEVIRT_IO_SRIOV_NET=0000:65:01.2,0000:65:01.3
...
KUBEVIRT_RESOURCE_NAME_sriovnet-vlan100=kubevirt.io/sriov_net
```

#### Summary
In this example we can see that both NICs points the same `NetworkAttachmentDefinition`, but since they have different MAC addresses, there is no way to know which PCI address holds which MAC address.

> **_Note_**: In order to pass-trough SR-IOV VF to a VM, the VF should be configured to use the vfio-pci driver [[1]](https://kubevirt.io/user-guide/virtual_machines/interfaces_and_networks/#sriov).
> The vfio-pci driver is an userspace driver, when a VF is bound to it, the VF is no longer recognized by the kernel and therefore will not be presented by iplink (‘ip a’ command).

# Proposal <a id='proposal'></a>
## General Flow
The proposal is based on information multus-cni provides as an annotation on the virt-launcher pod. The objective is to pass down the relevant information to the virt-launcher application, so it in turn can consume it and create the correct domain configuration.

In order to pass the VM network to VF PCI address mapping from the pod to the virt-launcher application, the following step are to be taken:
- Process the `k8s.v1.cni.cncf.io/network-status` annotation, and compose a structure that contains one-to-one mapping between a (VMI) network and the device PCI address.
- Create an annotation with the mapping information (e.g: `kubevirt.io/network-to-pci-address`) on the virt-launcher pod. <br/> This annotation is using the downward API to make it available to the application.
- Once the virt-launcher application requires the network-to-pci information, it can read the content. <br/>virt-launcher will use this data when creating the domain for the first time and when hotplug-ing the devices.

> **_Note_**:
This approach diverges from the standard data communication methodology of passing down information to the virt-launcher
via virt-handler with gRPC client. </br>
> However, as this is also true for all current host-devices implementation (i.e. vGPU, Generic host-device etc..).
> </br>
This kind of change is in some ways preferable but should be made as part of an overall refactoring design change of how host-devices information is passed down to the virt-launcher, and not as part of this remapping fix scope.

The following sections go into more details.

## Utilizing `k8s.v1.cni.cncf.io/network-status` annotation
This annotation includes the information required to do one-to-one mapping between a VM interface and VF PCI address:

### Example
```yaml
k8s.v1.cni.cncf.io/network-status: |-
    [
        {
            "name": "kindnet",
            "interface": "eth0",
            "ips": [
                "10.244.2.131"
            ],
            "mac": "82:cf:7c:98:43:7e",
            "default": true,
            "dns": {}
        },{
            "name": "default/sriov-network-vlan100",
            "interface": "net1",
            "dns": {},
            "device-info": {
                "type": "pci",
                "version": "1.0.0",
                "pci": {
                    "pci-address": "0000:04:02.5"
                }
            }
        },{
            "name": "default/sriov-network-vlan200",
            "interface": "net2",
            "dns": {},
            "device-info": {
                "type": "pci",
                "version": "1.0.0",
                "pci": {
                    "pci-address": "0000:04:02.2"
                }
            }
        }
    ]
```

> **_Note_**: Multus provides PCI device information since version [v3.7](https://github.com/k8snetworkplumbingwg/multus-cni/releases/tag/v3.7).<br/>
> In order to maintain backward compatibility, when the annotation is missing, virt-launcher will fall to the current mapping method based on environment variables.<br/>
> Once a Pod reaches Running state, the CNI operation is finished successfully and the annotation will present on the Pod.

### Annotations Exposure Method
A subset of the information in the `network-status` annotation needs to be passed down to the virt-launcher for composing the right domain configuration.

There have been several methods explored, recorded in [Appendix 1: Alternative Annotations Exposure Methods](#alternative-solutions).

In this section the chosen method is explained.

#### Expose `network-status` annotation to virt-launcher Pod
The annotation content can be exposed to virt-launcher application using Kubernetes Downward API [[1]](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/):

virt-controller creates an annotation that will hold the mapping between the (VMI) SR-IOV networks and PCI addresses, e.g: `kubevirt.io/network-to-pci-address`.

1. Given the `network-status` annotation, virt-controller process the content and compose 1:1 mapping between a (VMI) network from it and the device PCI address information.
   Create `network-to-pci-address` annotation with the mapping information, for example:
    ```yaml
    kubevirt.io/network-to-pci-address: |-
      {
        “sriovnet-vlan100-primary-mac”: “0000:04:02.5”,
        “sriovnet-vlan200-primary-mac”: “0000:04:02.2”
      }
    ```

2. Mount `network-to-pci-address` annotation using Kubernetes downward API as part of virt-launcher Pod creation:
    ```yaml
    kind: Pod
    metadata:
      generateName: virt-launcher-
      annotations:
        kubevirt.io/network-to-pci-address: '{“sriovnet-vlan100-primary-mac”: “0000:04:02.5”, “sriovnet-vlan200-primary-mac”: “0000:04:02.2”}'
    ...
    spec:
    containers:
    - name: “compute”
      ...
      volumeMounts:
        - name: network-to-pci-address-annotation
          mountPath: /etc/podinfo
          ...
          volumes:
    - name: network-to-pci-address-annotation
      downwardAPI:
      items:
        - path: "network-to-pci-address"
          fieldRef:
          fieldPath: metadata.annotations['network-to-pci-address']
          ...
    ```

    > **_Note_**: The annotation details will be populated at `/etc/podinfo/network-to-pci-address` file inside the pod.<br/>
    > Any change to the annotation will be reflected in the file.

3. As part of virt-launcher VMI spec rendering, given the downward API file exists (e.g: `/etc/podinfo/network-to-pci-address`),
   access the `network-to-pci-address` data, lookup for the PCI address based on the network name and create each SR-IOV HostDevice accordingly.

4. If the file doesn't exist, fall back to the legacy mapping method (based on `PCIDEVICE` and `KUBEVIRT_RESOURCE_NAME` environment variables).

##### Backward compatibility
- On post Kubevirt upgrade new virt-controller will treat new and old VMIs and their pod in the same manner - adding the `network-to-pci-address` annotation on the pods.

- No disruption to already existing VMIs workloads, though the SR-IOV mapping may not be correct.

##### Pros
1. This solution involves only virt-controller and virt-launcher.
2. The mapping info is passed down using the standard Kubernetes mechanism (Downward API) with no need to add additional logic.
3. Flexibility, no API change to VM or VMI objects.

##### Cons
1. Passing down the mapping info to virt-launcher via Downward API is bypassing the standard communication channel.
2. Depends on Downward API feature and volume mount on the Pod.

### New Mapping Algorithm
Currently, virt-launcher uses the information from the `PCIDEVICE` and `KUBEVIRT_RESOURCE_NAME` environment variables,
the proposal is to use the information from Multus `k8s.v1.cni.cncf.io/network-status` annotation in order to map between the SR-IOV interfaces and the underlying VF PCI address.

#### Mapping logic
1. Map the `interface` field in `network-status` annotation to VMI spec.networks:
- On pod creation, a network request annotation (`k8s.v1.cni.cncf.io/networks`) is created on the virt-launcher pod [[1]](https://github.com/kubevirt/kubevirt/blob/c4b6ae63c5a7f642ab86b0755dabca3b814ecb39/pkg/virt-controller/services/template.go#L1692).
  - The annotation creates a network request per each non-default multus network.<br/>
    For each CNI entry, each name is given an indexed name `net#` saved on is set in the `interface` field, acting as an identifier.
  - The VMI spec.networks `interface` filed equal to `interface` field the `network-status` annotation given by multus-cni.
- Duplicating this logic gives us a 1:1 map between the `interface` field on the `network-status` annotations (net1, net2, etc..) and the network name.

##### Example
The following VMI has 1 bridge and 2 SR-IOV secondary networks:

```yaml
kind: VirtualMachineInstance
...
spec:
  domain:
    devices:
      Interfaces:
      - name: bridge-primary-mac
        macAddress: aa:bb:cc:dd:ee:00
        sriov: {}
      - name: sriovnet-vlan100-secondary-mac
        macAddress: aa:bb:cc:dd:ee:01
        sriov: {}
      - name: sriovnet-vlan100-third-mac
        macAddress: aa:bb:cc:dd:ee:02
        sriov: {}
      ...
  Networks:
  - multus:
        networkName: default/bridge-network
    name: bridge-primary-mac
  - multus:
        networkName: default/sriov-network-vlan100
    name: sriovnet-vlan100-secondary-mac
  - multus:
        networkName: default/sriov-network-vlan100
    name: sriovnet-vlan100-third-mac
  ...
```

The `interface` field is mapped to `network-name` as follows:
- `bridge-primary-mac` ------------------------> `net1`
- `sriovnet-vlan100-secondary-mac` -----> `net2`
- `sriovnet-vlan100-third-mac` -----------> `net3`

With this mapping we can easily retrieve the PCI address, as the one with the corresponding `interface` name:
```yaml
k8s.v1.cni.cncf.io/network-status: |-
    [{
          "name": "kindnet",
          "interface": "eth0",
          "ips": [
              "10.244.1.9"
          ],
          "mac": "3a:7e:42:fa:37:c6",
          "default": true,
          "dns": {}
      },{
          "name": "default/bridge-network",
          "interface": "net1",
          "mac": "aa:bb:cc:dd:ee:00",
          "dns": {}
      },{
          "name": "default/sriov-network-vlan100",
          "interface": "net2",
          "mac": "aa:bb:cc:dd:ee:01",
          "dns": {},
          "device-info": {
              "type": "pci",
              "version": "1.0.0",
              "pci": {
                  "pci-address": "0000:65:00.2"
              }
      },{
          "name": "default/sriov-network-vlan100",
          "interface": "net3",
          "mac": "aa:bb:cc:dd:ee:02",
          "dns": {},
          "device-info": {
              "type": "pci",
              "version": "1.0.0",
              "pci": {
                  "pci-address": "0000:65:00.3"
              }
          }
      }]
```

Since we’re only interested in the SR-IOV networks, the network-name to PCI address mapping will be:
- `sriovnet-vlan100-secondary-mac ` ---> `0000:65:00.2`
- `sriovnet-vlan100-third-mac` ----------> `0000:65:00.3`

> **_Note_**:
> We are relying on the fact that the `interface` naming convention in [[2]](https://github.com/kubevirt/kubevirt/blob/c4b6ae63c5a7f642ab86b0755dabca3b814ecb39/pkg/virt-controller/services/template.go#L1692) will remain the same.
> To ensure that, we plan move this logic to a new package that will be used for all scenarios related to these annotations.
> Moreover, this will be covered in e2e test in case the naming logic is changed without using the new package.
> One could suggest relying on the MAC address as an identifier instead of relying on the indexation logic, but that is simply not good enough since the MAC address field is not mandatory field.

##### Backward compatibility:
The mapping mentioned above could fail for several reasons, among them:
- `network-status` annotation is not present on the VMI.
- `network-status` annotation does not hold the PCI address information for all the SR-IOV interfaces.
- Annotation is not valid, such a case can happen if the `network-status` annotation is breaking the API or Multus deployed on the cluster is older than [v3.7](https://github.com/k8snetworkplumbingwg/multus-cni/releases/tag/v3.7).<br/>In any case, in the event of a mapping failure, Kubevirt will fall back to the old mapping method, in order to not break old VMIs on old deployments.

# Appendix 1: Alternative Annotations Exposure Methods <a id="alternative-solutions"></a>
## 1. virt controller exposes network to pci mapping through the VMI annotations <a id="alternative-option-1"></a>
1. virt-controller processes the VMI (virt-launcher) pod and adds it to the VMI annotations.

2. As part of virt-launcher VMI spec rendering:
- Given the `network-status` annotation exists on the VMI object, extract the mapping between SR-IOV networks and the VFs PCI addresses and create each HostDevice accordingly.
- In case `network-status` annotation doesn't exist, fall back to the legacy mapping method (based on environment variables).

### Backward compatibility
|                                                       | old virt-controller                                | new virt-controller                                                                                                                                                                                                              |
|:------------------------------------------------------|:---------------------------------------------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Already existing VMs<br/>(old virt-launcher)          | 1. No regression <br/> VFs mapping might be wrong. | 2. No regression <br/><br/>VFs mapping might be wrong.<br/>virt-controller adds the pod `network-status` annotations to VMI annotations,but virt-launcher wont perform the assignment according to it.                           |
| VM creation/restart/migration<br/>(new virt-launcher) | -                                                  | 4. No regression<br/>Once the VM is restarted, VMI or Pod is re-created, the new SR-IOV interfaces assignment logic will apply.<br/>In case the VM is migrated, the VM on target will have SR-IOV interfaces assigned correctly. |

#### Conclusion
- (2) virt-controller will add the `network-status` to the VMI, the guest will have no disruption and its VFs mapping will not change.
- (4) When virt-controller pods are is upgraded, new VMs will be created with the correct SR-IOV interfaces assignment.

### Pros
1. No need to use Downward API and mount its volume.
3. No API changes.

### Cons
1. virt-controller adds an annotation to the VMI.
2. During migration, there is a need to choose the right pod data to be processed (source or target).

## 2. virt-controller exposes network-to-pci mapping through two VMI annotations <a id="alternative-option-2"></a>
Following [option 1](#alternative-option-1) con (2), this solution mitigates it as follows:
In order to ensure that the correct mapping information is passed down to virt-launcher, it is necessary to record each of the VMI virt-launcher pods `network-status` annotation.

1. virt-controller processes all VMI (virt-launcher) pods and for each, creates a VMI annotation that has the network-to-pci mapping.<br/>During migration, a VMI has two pods (source and target).<br/>   The pod identifier is embedded into the annotation key (e.g: pod name, UID, etc..).
   #### Example
    ```yaml
    kind: VirtualMachineInstance
    ...
    metadata:
      annotations:
        kubevirt.io/virt-launcher-abc123-network-status: "...",
        kubevirt.io/virt-launcher-xyz456-network-status: "..."
    ...
    ```

2. virt-launcher picks the relevant annotation and perform the mapping accordingly.

#### Pros
1. No need to synchronize the `network-status` annotation from the pods (source and target).

#### Cons
2. virt-controller need to maintain two annotations during migration and one on the regular flow.

## 3. **virt-handler** exposes network-to-pci mapping through VMI annotations
As part of the virt-handler VM create/update flow, the VMI spec is processed and virt-launcher is triggered to synchronize with the new VMI spec [[1]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-handler/vm.go#L2639) (i.e: sends [SyncVMI](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/virt-launcher/virtwrap/manager.go#L822) gRPC call to virt-launcher).

1. Extend virt-handler to virt-launcher command options to include the mapping information.

2. As part of virt-handler VMI update flow, fetch the corresponding virt-launcher pod and read its `network-status` annotation.<br/>If the annotation exists, extract the network-to-pci mapping and send it to virt-launcher.

3. As part of virt-launcher VMI spec rendering:<br/> if the network-to-pci mapping information is valid, create each SR-IOV HostDevice accordingly. <br/> In case the information is invalid fall back to the legacy mapping method (based on environment variables).

### Backward compatibility
|                                                       | old virt-handler                                                                                                                                                    | new virt-handler                                                                                                                                                                                                                                                          |
|:------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------|:--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Already existing VMs<br/>(old virt-launcher)          | 1. No regression<br/>VFs mapping might be wrong.                                                                                                                    | 2. Regression<br/>virt-handler fetches virt-launcher pod and annotates the VMI with the `network-status` annotation.<br/>Next, virt-handler sends the network-to-pci mapping information to virt-launcher trough gRPC. virt-launcher fall backs to legacy mapping method. |
| VM creation/restart/migration<br/>(new virt-launcher) | 3. Regression<br/>Virt-handler sends the network-to-pci mapping information to virt-launcher through gRPC, virt-launcher should fall back to legacy mapping method. | 4. No regression                                                                                                                                                                                                                                                          |

#### Conclusion
- (2) and (3):<br/>
  There might be a disruption to VMs when they were created:
    - After the virt-controller is updated (new virt-launcher), on a node with an old instance of virt-handler.
    - Before the virt-controller is updated (old virt-launcher), on a node with a new instance of virt-handler.
      <br/><br/>
      Since the gRPC command server (virt-launcher) and client (virt-handler) runs with different API versions, virt-launcher may fail to synchronize with the new VMI state.
      <br/><br/>

- (4) When both virt-handler and virt-controller are updated, new VMs will be created with the correct SR-IOV interfaces assignment.

### Pros
1. No need to involve virt-controller.
2. No need to create annotations on the pod or VMI.

### Cons
1. Requires API changes to virt-handler to virt-launcher commands API [[7]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/handler-launcher-com/cmd/v1/cmd.pb.go) [[8]](https://github.com/kubevirt/kubevirt/blob/1a8f08103e48c5f2bb2f5826d118507ce7ec1f0c/pkg/handler-launcher-com/cmd/v1/cmd.pb.go#L334).
2. virt-handler needs to fetch the pod object in order to get the mapping information.

## 4. Expose network-to-pci mapping through the VMI status
This option uses the same basic ideas from [option 2](#alternative-option-2) but instead of using VMI annotations it
uses VMI Status additional fields to persist the mapping.

1. Extend the VMI status API to reflect each SR-IOV interface underlying PCI address.

2. The virt-controller updates the status fields with the mapping information.

3. virt-launcher VMI extract the network-to-pci mapping from the VMI status and create each HostDevice accordingly.<br/>
   In case there is not enough information in the VMI status (i.e: missing host PCI address) fall back to legacy mapping method.

#### Pros
1. No need for annotations.

#### Cons
1. The status field will have to support mapping from two pods, supporting migration.
2. It is confusing to use the VMI status as an input to the virt-launcher.
3. The host PCI address is now exposed to the user and may be confusing.

## 5. Expose network-to-pci mapping through general propose annotation
This option uses the same basic ideas from the [proposed solution](#proposal) but instead of using dedicated annotation just for the network-to-pci mapping,
there will be a general propose annotation that enable passing any kind of information down to virt-launcher.

In order to pass the information from the pod to the virt-launcher application, the following step are to be taken:
- Create a generic annotation for passing information from the pod manifest to the application that runs in it. This annotation is using the downward api to make it available to the application.
- Given the `network-status` annotation, process the content and compose structure that contains a 1:1 mapping between a (VMI) network from it and the device PCI address. This data is added to the generic annotation from earlier.
- Once the virt-launcher application requires this network-to-pci information, it can read the content.
- In the context of SR-IOV, virt-launcher will use this data when creating the domain for the first time and when hotplug-ing the devices.

1. virt-controller creates `virt-metadata` annotation that will hold generic information about the pod that we want to expose to the virt-launcher pod user.

2. Given the `network-status` annotation, process the content and compose 1:1 mapping between a (VMI) network from it and the device PCI address information and add it to `virt-metadata` annotation.
   For example:
```yaml
kubevirt.io/virt-metadata: |-
{
  network-to-pci-address: 
  {
    “sriovnet-vlan100-primary-mac”: “0000:04:02.5”,
    “sriovnet-vlan200-primary-mac”: “0000:04:02.2”
  }
}
```

2. Mount `virt-metadata` annotation using Downward API as part of virt-launcher Pod creation:
```yaml
kind: VirtualMachineInstance
metadata:
  name: sriovvmi
...
spec:
containers:
- name: “compute”
  ...
  volumeMounts:
    - name: virt-metadata-annotation
      mountPath: /etc/podinfo
      ...
      volumes:
- name: virt-metadata-annotation
  downwardAPI:
  items:
    - path: "virt-metadata"
      fieldRef:
      fieldPath: metadata.annotations['virt-metadata']
      ...
```

4. As part of virt-launcher VMI spec rendering, given the downward API file exists (e.g: `/etc/podinfo/virt-metadata` file),
   access the network to PCI address data, lookup for the PCI address based on the network name and create each SR-IOV HostDevice accordingly.

5. If the file doesn't exist, fall back to the legacy mapping method (based on environment variables).

##### Pros
1. Enable passing down any kind of information to virt-launcher application with no need to synchronize the information during migration.

#### Cons
1. Introduce new communication channel between virt-handler to virt-launcher.
