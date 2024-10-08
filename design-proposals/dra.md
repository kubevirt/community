# Overview

This proposal is about adding supporting DRA (dynamic resource allocation) in KubeVirt.
DRA allows vendors fine-grained control of devices. The device-plugin model will continue
to exist in Kubernetes, but DRA will offer vendors more control over the device topology.

## Motivation

DRA adoption is important for KubeVirt so that vendors can expect the same
control of their devices using Virtual Machines and Containers.

## Goals

- Align on the API changes needed to consume DRA enabled devices in KubeVirt
- Align on how KubeVirt will consume devices by external DRA drivers

## Non Goals

- Replace device-plugin support in KubeVirt
- Align on what drivers KubeVirt will support in tree

## Definition of Users

- A user is a person that wants to attach a device to a VM
- An admin is a person who manages the infrastructure and decide what kind of devices will be exposed via DRA
- A developer is a person familiar with CNCF ecosystem that can develop automation using the APIs discussed here

## Assumptions

- KubeVirt will support DRA API from kubernetes version 1.31.
- The example demonstrate a pGPU driver, but the same mechanism could be used for vGPU

## User Stories

- As a user, I would like to use my GPU dra driver with KubeVirt
- As a user, I would like to use KubeVirt's default driver
- As a developer, I would like APIs to be extensible so I can develop drivers/webhooks/automation for custom use-cases

## Use Cases

### Supported Usecases

1. Devices where the DRA driver that set agreed upon attributes for resources in ResourceSlice.
   1. PCIe Bus address for pGPU
   2. Device uuid for vGPU
1. Devices have to either be of the type GPU or HostDevices. 

### Unsupported Usecases

1. Devices where the DRA driver does not set the attributes needed to configure libvirt dom XML for the devices will not
   be supported.
   1. In the future a standardization could be envisioned for a PCI device where the attributes are set
      automatically through the DRA framework. When this is achieved, any DRA device should be available in KubeVirt VMs

## Repos

kubevirt/kubevirt

# Design

For allowing users to consume DRA devices, there are two main changes needed:

1. API changes and the plumbing required in KubeVirt to generate the domxml with the devices.
2. Driver Implementation to set the env var

This design focuses on part 1 of the problem.

## API Changes

```go
type VirtualMachineInstanceSpec struct {
    ..
    ..
    // ResourceClaims defines which ResourceClaims must be allocated
	// and reserved before the VMI and hence virt-launcher pod is allowed to start. The resources
	// will be made available to the domain which consume them
	// by name.
	//
	// This is an alpha field and requires enabling the
	// DynamicResourceAllocation feature gate in kubernetes
	//  https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/
	//
	// This field is immutable.
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	ResourceClaims []k8sv1.PodResourceClaim `json:"resourceClaims,omitempty"`
}

type GPU struct {
	// Name of the GPU device as exposed by a device plugin
	Name string                    `json:"name"`
	DeviceSource DeviceSource      `json:",inline"`
	VirtualGPUOptions *VGPUOptions `json:"virtualGPUOptions,omitempty"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
}

type HostDevice struct {
	Name string               `json:"name"`
	DeviceSource DeviceSource `json:",inline"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
}

type DeviceSource struct {
	// DeviceName is the name of the device provisioned by device-plugins
	DeviceName *string `json:"deviceName,omitempty"`
	// Claim is the name of the claim that is going to provision the DRA device
	Claim *k8sv1.ResourceClaim `json:"claim,omitempty"`
}

type VirtualMachineInstanceStatus struct {
    ..
    ..
	// DeviceStatus reflects the state of devices requested in spec.domain.devices. This is an optional field available
	// only when DRA feature gate is enabled
	// +optional
	DeviceStatus *DeviceStatus `json:"deviceStatus,omitempty"`
}

// DeviceStatus has the information of all devices allocated spec.domain.devices
type DeviceStatus struct {
	// GPUStatuses reflects the state of GPUs requested in spec.domain.devices.gpus
	// +listType=atomic
	// +optional
	GPUStatuses []DeviceStatusInfo `json:"gpuStatuses,omitempty"`
	// HostDeviceStatuses reflects the state of GPUs requested in spec.domain.devices.hostDevices
	// DRA
	// +listType=atomic
	// +optional
	HostDeviceStatuses []DeviceStatusInfo `json:"hostDeviceStatuses,omitempty"`
}

type DeviceStatusInfo struct {
	// Name of the device as specified in spec.domain.devices.gpus.name or spec.domain.devices.hostDevices.name
	Name string `json:"name"`
	// DeviceResourceClaimStatus reflects the DRA related information for the device
	// +optional
	DeviceResourceClaimStatus *DeviceResourceClaimStatus `json:"deviceResourceClaimStatus,omitempty"`
}

// DeviceResourceClaimStatus has to be before SyncVMI call from virt-handler to virt-launcher
type DeviceResourceClaimStatus struct {
	// ResourceClaimName is the name of the resource claims object used to provision this resource
	// +optional
	ResourceClaimName *string `json:"resourceClaimName,omitempty"`
	// DeviceName is the name of actual device on the host provisioned by the driver as reflected in resourceclaim.status
	// +optional
	DeviceName *string `json:"deviceName,omitempty"`
	// DeviceAttributes are the attributes published by the driver running on the node in
	// resourceslice.spec.devices.basic.attributes. The attributes are distinguished by deviceName
	// and resourceclaim.spec.devices.requests.deviceClassName.
	// +optional
	DeviceAttributes map[string]DeviceAttribute `json:"deviceAttributes,omitempty"`
}
```

The first section vmi.spec.resourceClaims will have a list of devices needed to be allocated for the VM. Having this
available as a list will allow users to use the device from this list in GPU section or HostDevices section of the 
DomainSpec API. 

In v1alpha3 version of [DRA API](https://pkg.go.dev/k8s.io/api@v0.31.0/resource/v1alpha3#DeviceClaim), multiple drivers
could potentially provision devices that are part of a single claim. For this reason, a separate list of claims required
for the VMI (section 1) is needed instead of mentioning the resource claim in devices section 
([see Alternate Designs](#Alternative 1))

The second sections allows for the resource claim to be used in the spec.domain.devices section. The two uses cases
currently handled by the design are:

1. allowing the devices to be used as a gpu device (spec.domain.devices.gpu)
2. allowing the devices to be used as a host device (spec.domain.device.hostDevices)

The status section of the VMI will contain information of the allocated devices for the VMI when the information is
available in DRA APIs. The same information will be accessible in virt-handler, virt-launcher and sidecar containers. 
This allows for device information to flow from DRA APIs into KubeVirt stack. 

The virt-launcher will have the logic of converting a GPU device into its corresponding domxml. For use-cases that are 
not handled in-tree, a sidecar container could be envisioned which will convert the information available in status to 
the corresponding domxml.

### Examples

### VM API with PassThrough GPU

```
---
# this is a cluster scoped resource
apiVersion: resource.k8s.io/v1alpha3
kind: DeviceClass
metadata:
  name: gpu.example.com
spec:
  selectors:
  - cel:
      expression: device.driver == 'gpu.example.com'
---
apiVersion: resource.k8s.io/v1alpha3
kind: ResourceClaimTemplate
metadata:
  name: pgpu-claim-template
spec:
  spec:
    devices:
      requests:
        - name: pgpu-request-name
          deviceClassName: gpu.example.com
---
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  resourceClaims:
  - name: pgpu-claim-name
    source:
      resourceClaimTemplateName: pgpu-claim-template
  domain:
    devices:
      gpus:
      - name: pgpu
        claim: 
          name: pgpu-claim-name
          request: pgpu-request-name
status:
  deviceStatus:
    gpuStatuses:
    - deviceResourceClaimStatus:
        deviceAttributes:
          driverVersion:
            version: 1.0.0
          index:
            int: 0
          model:
            string: LATEST-GPU-MODEL
          uuid:
            string: gpu-8e942949-f10b-d871-09b0-ee0657e28f90
          pciAddress: 
            string: 0000:01:00.0
        deviceName: gpu-0
        resourceClaimName: virt-launcher-vmi-fedora-9bjwb-gpu-resource-claim-m4k28
      name: pgpu     
–--
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-cirros
spec:
  containers:
  - name: virt-launcher
    image: virt-launcher
    resources:
      claims:
      - name: pgpu-claim-name
        request: pgpu-request-name
  resourceClaims:
  - name: pgpu-claim-name
    source:
      resourceClaimTemplateName: pgpu-claim-template
status:
  resourceClaimStatuses:
  - name: gpu-resource-claim
    resourceClaimName: virt-launcher-vmi-fedora-9bjwb-gpu-resource-claim-m4k28
```

### VM API with a Host device

```
---
# this is a cluster scoped resource
apiVersion: resource.k8s.io/v1alpha3
kind: DeviceClass
metadata:
  name: pci-nvme.kubevirt.io
spec:
  selectors:
  - cel:
      expression: device.driver == 'pci-nvme.example.com'
---
apiVersion: resource.k8s.io/v1alpha3
kind: ResourceClaimTemplate
metadata:
  name: pci-nvme-claim-template
spec:
  spec:
    devices:
      requests:
        - name: pci-nvme-request-name
          deviceClassName: pci-nvme.kubevirt.io
          allocationMode: ExactCount
          count: 1
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-cirros
  name: vm-cirros
spec:
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-cirros
    spec:
      resourceClaims:
      - name: nvme1-claim-name
        source:
          resourceClaimTemplateName: test-pci-claim-template
      domain:
        devices:
          hostDevices:
          - name: nvme1-host-device
            claim: 
              name: nvme1-claim-name
              request: pci-nvme-request-name
–--
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-cirros
spec:
  containers:
  - name: virt-launcher
    image: virt-launcher
    resources:
      claims:
      - name: nvme1-host-device
        claim: 
          name: nvme1-claim-name
          request: pci-nvme-request-name
  resourceClaims:
  - name: nvme1-claim-name
    source:
      resourceClaimTemplateName: test-pci-claim-template
```

### DRA API for reading device related information

The examples below shows the APIs used to generate the vmi.status.deviceStatuses section: 
1. the pod status has reference to the resourceClaimName, pod.status.resourceClaimStatuses[*].resourceClaimName where
   the name of the claim is same as vmi.spec.resourceClaims[*].Name
1. pod spec has node name, pod.spec.nodeName
1. the resourceclaim status has device name and driver use for allocating the device, 
   resourceclaim.status.allocation.devices[*].deviceName and resourceclaim.status.allocation.devices[*].driver, where
   resourceclaim.status.allocation.devices[*].request is same as vmi.spec.domain.devices[*].gpus[*].claim.request
1. Using node name and driver name, the resource slice for that node could be found. Using device name, the attributes 
   of the device could be found

```
---
apiVersion: v1
kind: Pod
metadata:
  name: virt-launcher-vmi-fedora-9bjwb
  namespace: gpu-test1
spec:
  containers:
  - name: compute
    resources:
      claims:
      - name: gpu-resource-claim
  resourceClaims:
  - name: gpu-resource-claim
    resourceClaimTemplateName: single-gpu
status:
  resourceClaimStatuses:
  - name: gpu-resource-claim
    resourceClaimName: virt-launcher-vmi-fedora-9bjwb-gpu-resource-claim-m4k28
---
apiVersion: resource.k8s.io/v1alpha3
kind: ResourceClaim
metadata:
  annotations:
    resource.kubernetes.io/pod-claim-name: gpu-resource-claim
  generateName: virt-launcher-vmi-fedora-9bjwb-gpu-resource-claim-
  name: virt-launcher-vmi-fedora-9bjwb-gpu-resource-claim-m4k28
  namespace: gpu-test1
  ownerReferences:
  - apiVersion: v1
    blockOwnerDeletion: true
    controller: true
    kind: Pod
    name: virt-launcher-vmi-fedora-9bjwb
spec:
  devices:
    requests:
    - allocationMode: ExactCount
      count: 1
      deviceClassName: gpu.example.com
      name: gpu
status:
  allocation:
    devices:
      results:
      - device: pgpu-0
        driver: gpu.example.com
        pool: kind-1.31-dra-control-plane
        request: gpu
    nodeSelector:
      nodeSelectorTerms:
      - matchFields:
        - key: metadata.name
          operator: In
          values:
          - kind-1.31-dra-control-plane
  reservedFor:
  - name: virt-launcher-vmi-fedora-9bjwb
    resource: pods
    uid: 8ffb7e04-6c4b-4fc7-bbaa-c60d9a1e0eaa
---
apiVersion: resource.k8s.io/v1alpha3
kind: ResourceSlice
metadata:
  generateName: kind-1.31-dra-control-plane-gpu.example.com-
  name: kind-1.31-dra-control-plane-gpu.example.com-drr27
  ownerReferences:
  - apiVersion: v1
    controller: true
    kind: Node
    name: kind-1.31-dra-control-plane
spec:
  devices:
  - basic:
      attributes:
        driverVersion:
          version: 1.0.0
        index:
          int: 0
        model:
          string: LATEST-GPU-MODEL
        uuid:
          string: gpu-8e942949-f10b-d871-09b0-ee0657e28f90
        pciAddress:
          string: 0000:01:00.0 
    name: pgpu-0
  driver: gpu.example.com
  nodeName: kind-1.31-dra-control-plane
  pool:
    generation: 0
    name: kind-1.31-dra-control-plane
    resourceSliceCount: 1
---
```

### Web hook changes
1. Allow DRA devices to be requested only if the corresponding DRA feature gate is enabled in kubevirt configuration
Note: All the following sections will assume that DRA feture gate is enabled

### Virt controller changes

1. If devices are requested using DRA, virt controller needs to render the virt-launcher manifest such that 
   pod.spec.resourceClaims and pod.spec.containers.resources.claim sections are filled out.
1. virt-controller needs a mechanism to watch for virt-launcher pods, resourceclaims and resourceslices to populate the 
   vmi.status.deviceStatus using the steps mentioned in above section that has all the attributes (for example the 
   pciAddress for the gpu device):
   1. The pod status has information about the allocated/reserved resourceClaim.
   1. The resourceClaim has information about the individual requests in the claim and their allocated device names.
   1. The resourceslice corresponding to the node running the VMI has information about the allocated device.

### Virt launcher changes

1. For devices generated using DRA, virt-launcher needs to use the vmi.status.deviceStatus to generate the domxml
   instead of environment variables as in the case of device-plugins
1. The standard env variables `PCI_RESOURCE_<deviceName>` and `MDEV_PCI_RESOURCE_<deviceName>` may continue to be set
   as fallback mechanisms but the focus here is to ensure we can consume the device PCIe bus address atrribute from the
   allocated devices in virt-launcher to generate the domxml.
1. Both GPU and HostDevice devices requested in the domain spec will have corresponding entries in the VMI status
   at `status.deviceStatus.gpuStatuses[*]`/`status.deviceStatus.hostDeviceStatuses[*]`. From here, the relevant
   device attributes can be inferred by virt-launcher (`pcieAddress` attr) to generate the domxml with the appropriate
   gpu/hostdev spec.

# Alternate Designs

## Alternative 1

```
type GPU struct {
	// Name of the GPU device as exposed by a device plugin
	Name              string       `json:"name"`
	DeviceName        string       `json:"deviceName"`
	VirtualGPUOptions *VGPUOptions `json:"virtualGPUOptions,omitempty"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag   string `json:"tag,omitempty"`
	Claim string `json:"claim,omitempty"`
}

type HostDevice struct {
	Name string `json:"name"`
	// DeviceName is the resource name of the host device exposed by a device plugin
	DeviceName string `json:"deviceName"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
	// If specified, the ResourceName of the host device will be provisioned using DRA driver . which will not require the deviceName field
	//+optional
	ResourceClaim *ResourceClaim `json:"resourceClaim,omitempty"`
}

// ResourceClaim represents a resource claim to be used by the virtual machine
type ResourceClaim struct {
	// Name is the name of the resource claim
	Name string `json:"name"`
	// Source represents the source of the resource claim
	Source ResourceClaimSource `json:"source"`
}

// ResourceClaimSource represents the source of a resource claim
type ResourceClaimSource struct {
	// ResourceClaimName is the name of the resource claim
	ResourceClaimName string `json:"resourceClaimName"`
	// ResourceClaimTemplateName is the name of the resource claim template
	//
	// Exactly one of ResourceClaimName and ResourceClaimTemplateName must
	// be set.
	ResourceClaimTemplateName string `json:"resourceClaimTemplateName"`
}

```

This design misses the use-case where more than one DRA device is specified in the claim template, as each
device will have its own template in the API.

This design also assumes that the deviceName will be provided in the ClaimParameters, which requires the DRA drivers
to have a ClaimParameters.spec.deviceName in their spec.

# References

- Structured parameters
  https://github.com/kubernetes/kubernetes/pull/123516
- Structured parameters KEP
  https://github.com/kubernetes/enhancements/issues/4381
- DRA
  https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/