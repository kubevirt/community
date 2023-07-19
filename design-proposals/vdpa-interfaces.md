# Overview
A vDPA device means a type of device whose datapath complies with the virtio specification, but whose control path is vendor specific. Currently, two vDPA drivers are implemented:
- vhost interface (vhost-vdpa) for userspace or guest virtio driver, like a VM running in QEMU
- virtio interface (virtio-vdpa) for bare-metal or containerized applications running in the host

This design provides an approach for creating a VM with vhost-vdpa interface.

## Goals
Have a mechanism to allow creation of virtual machines with vdpa interfaces.

## Non Goals
Because the implementation of vDPA can be SR-IOV based or not, this mechanism should not be coupled with the [Intel SR-IOV device plugin](https://github.com/intel/sriov-network-device-plugin) and [SR-IOV CNI plugin](https://github.com/k8snetworkplumbingwg/sriov-cni).

## User Stories
As a Kubevirt user I would like to create a vm with vdpa interfaces.

## Repos
[KubeVirt](https://github.com/kubevirt/kubevirt)

# Design
Design can be inspired by the existing [sriov type interface](https://kubevirt.io/user-guide/virtual_machines/interfaces_and_networks/#sriov), but you can see the difference by comparing the libvirt domain XML:
```xml
<hostdev type="pci" managed="no" mode="subsystem">
        <source>
                <address type="" domain="0x0000" bus="0xcc" slot="0x00" function="0x4"></address>
        </source>
        <address type="pci" domain="0x0000" bus="0x07" slot="0x00" function="0x0"></address>
        <alias name="ua-sriov-offload-ovn"></alias>
</hostdev>
```
```xml
<interface type="vdpa">
        <source dev="/dev/vhost-vdpa-1"></source>
        <model type="virtio"></model>
        <alias name="ua-offload-ovn"></alias>
        <driver name="vhost" queues="4"></driver>
</interface>
```
- The former is hostdev while the latter is interface
- The source of the former is a bdf while the source of the latter is a device path.

## API
Add a vDPA type to InterfaceBindingMethod:
```go
// Represents the method which will be used to connect the interface to the guest.
// Only one of its members may be specified.
type InterfaceBindingMethod struct {
	Bridge     *InterfaceBridge     `json:"bridge,omitempty"`
	Slirp      *InterfaceSlirp      `json:"slirp,omitempty"`
	Masquerade *InterfaceMasquerade `json:"masquerade,omitempty"`
	SRIOV      *InterfaceSRIOV      `json:"sriov,omitempty"`
	VDPA       *InterfaceVDPA       `json:"vdpa,omitempty"`
	Macvtap    *InterfaceMacvtap    `json:"macvtap,omitempty"`
	Passt      *InterfacePasst      `json:"passt,omitempty"`
}
```

## Example
VMI spec with two vDPA devices, connected to the same network:
```yaml
          networkInterfaceMultiqueue: true
          interfaces:
          - masquerade: {}
            name: default
          - name: offload
            vdpa: {}
          - name: offload2
            vdpa: {}
      networks:
      - name: default
        pod: {}
      - multus:
          networkName: default/offload-ovn
        name: offload
      - multus:
          networkName: default/offload-ovn
        name: offload2
```

The default/offload-ovn NetworkAttachmentDefinition spec:
```yaml
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: offload-ovn
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: intel.com/vdpa_jaguar_vhost
spec:
  config: >-
    {
        "cniVersion": "0.3.0",
        "type": "kube-ovn",
        "server_socket": "/run/openvswitch/kube-ovn-daemon.sock",
        "provider": "offload-ovn.default.ovn"
    }
```

The Configuration file for the Intel SR-IOV device plugin:
```yaml
apiVersion: v1
data:
  config.json: |
    {
        "resourceList": [
            {
                "resourceName": "vdpa_jaguar_vhost",
                "selectors": [{
                    "vendors": ["****"],
                    "devices": ["1000"],
                    "drivers": ["jaguar"],
                    "vdpaType": "vhost"
                }]
            }
        ]
    }
```

The following environment variables will be present in the Pod:
```bash
# kubectl exec -it virt-launcher-myvmi-zmdbh -- env
PCIDEVICE_INTEL_COM_VDPA_JAGUAR_VHOST_INFO={"0000:cc:00.2":{"generic":{"deviceID":"0000:cc:00.2"},"vdpa":{"mount":"/dev/vhost-vdpa-1"}},"0000:cc:00.3":{"generic":{"deviceID":"0000:cc:00.3"},"vdpa":{"mount":"/dev/vhost-vdpa-2"}}}
PCIDEVICE_INTEL_COM_VDPA_JAGUAR_VHOST=0000:cc:00.2,0000:cc:00.3
KUBEVIRT_RESOURCE_NAME_offload=intel.com/vdpa_jaguar_vhost
KUBEVIRT_RESOURCE_NAME_offload2=intel.com/vdpa_jaguar_vhost
```

The domain XML like so:
```xml
    <interface type='vdpa'>
      <mac address='52:54:00:ff:36:37'/>
      <source dev='/dev/vhost-vdpa-1'/>
      <model type='virtio'/>
      <driver name='vhost' queues='4'/>
      <alias name='ua-offload'/>
      <address type='pci' domain='0x0000' bus='0x02' slot='0x00' function='0x0'/>
    </interface>
    <interface type='vdpa'>
      <mac address='52:54:00:38:f0:ac'/>
      <source dev='/dev/vhost-vdpa-2'/>
      <model type='virtio'/>
      <driver name='vhost' queues='4'/>
      <alias name='ua-offload2'/>
      <address type='pci' domain='0x0000' bus='0x03' slot='0x00' function='0x0'/>
    </interface>
```