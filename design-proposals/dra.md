# Overview

This proposal is about adding supporting DRA (dynamic resource allocation) in KubeVirt.
DRA allows vendors fine-grained control of devices. The device-plugin model will continue
to exist in Kubernetes, but DRA will offer vendors more control over the device topology.

## Motivation

DRA adoption is important for KubeVirt so that vendors can expect the same
control of their devices using Virtual Machines and Containers.

## Goals

- Align on the API changes needed to consume DRA enabled devices in KubeVirt
- Align on how KubeVirt will consume external and in-tree DRA drivers
- Align on what drivers KubeVirt will support in tree
TODO: @alayp Q: Should the decision around in-tree drivers be in scope for this proposal?

## Non Goals

- Replace device-plugin support in KubeVirt

## Definition of Users

- A user is a person that wants to attach a device to a VM
- An admin is a person who manages the infrastructure and decide what kind of devices will be exposed via DRA
- A developer is a person familiar with CNCF ecosystem that can develop automation using the APIs discussed here

## Assumptions

- KubeVirt will support DRA API from kubernetes version 1.31.
- The example demonstrate a pgpu driver, but the same mechanism could be used for vgpu
TODO: @alayp Q: Can device plugins and DRA drivers co-exists in the same cluster?

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

1. API changes and the plumbing required in KubeVirt to read env var and convert it into right domxml
2. Driver Implementation to set the env var

This design focuses on part 1 of the problem.

## API Changes

```
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
	Name string `json:"name"`
	// DeviceName is the name of the device provisioned by device-plugins
	DeviceName string `json:"deviceName,omitempty"`
	// Claim is the name of the claim that is going to provision the DRA device
	Claim             *k8sv1.ResourceClaim `json:"claim,omitempty"`
	VirtualGPUOptions *VGPUOptions         `json:"virtualGPUOptions,omitempty"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
}

type HostDevice struct {
	Name string `json:"name"`
	// DeviceName is the name of the device provisioned by device-plugins
	DeviceName string `json:"deviceName,omitempty"`
	// Claim is the name of the claim that is going to provision the DRA device
	Claim *k8sv1.ResourceClaim `json:"claim,omitempty"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
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
// +k8s:openapi-gen=true
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
	// DeviceResourceClaimStatus reflects the DRA related information for the degive
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
available as a list will allow users to use the device available from this list in GPU section or HostDevices section of
the DomainSpec API. 

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
```

### VM API with a GPU and a Host device

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
