# Overview
This proposal is about adding supporting DRA (dynamic resource allocation) in KubeVirt.
DRA allows vendors fine-grained control of devices.  The device-plugin model will continue
to exist in Kubernetes, but DRA will offer vendors more control over the device topology.

## Motivation
DRA adoption is important for KubeVirt so that vendors can expect the same
control of their devices using Virtual Machines and Containers.

## Goals
- Align on how KubeVirt will consume external and in-tree DRA drivers
- Align on what drivers KubeVirt will support in tree
- Align on the API changes needed to consume DRA enabled devices in KubeVirt

## Non Goals
- Replace device-plugin support in KubeVirt

## Definition of Users
- A user is a person that wants to attach a device to a VM
- An admin is a person who manages the infrastructure and decide what kind of devices will be exposed via DRA
- An developer is a person familiar with CNCF ecosystem that can develop automation using the APIs discussed here  

## User Stories
- As a user, I would like to use my GPU dra driver with KubeVirt
- As a user, I would like to use KubeVirt's default driver
- As a developer, I would like APIs to be extensible so I can develop drivers/webhooks/automation for custom use-cases

## Repos
kubevirt/kubevirt

# Design

At the high level there are two sets of changes needed for integrating DRA with KubeVirt:
1. How DRA resources can be consumed by users of KubeVirt?
2. What changes are needed for existing device-plugin drivers to be DRA compatible?

This design focuses on part 1 of the problem. 

For allowing users to consume DRA devices, there are two main changes needed:
1. API changes
2. Driver Implementation

## API Changes

```
type VirtualMachineInstanceSpec struct {
    // ResourceClaims defines which ResourceClaims must be allocated
    // and reserved before the virt-launcher pod and hence the VMI is allowed to start. The resources
    // will be made available to the domain which consume them
    // by name.
    //
    // This is an alpha field and requires enabling the
    // DynamicResourceAllocation feature gate in kubernetes
    //  https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/
    //
    // This field is immutable.
    //
    // +featureGate=DynamicResourceAllocation
    ResourceClaims []k8sv1.PodResourceClaim `json:"resourceClaims,omitempty"`
}

type GPU struct {
	// Name of the GPU device as exposed by a device plugin
	Name              string       `json:"name"`
	DeviceName        string       `json:"deviceName"`
	ClaimName         string       `json:"claimName"`
	VirtualGPUOptions *VGPUOptions `json:"virtualGPUOptions,omitempty"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag   string `json:"tag,omitempty"`
	// Claim is the name of resource, defined in spec.resourceClaims,
	// that is used to provision this GPU device.
	//
	// This is an alpha field and requires enabling the
	// DynamicResourceAllocation feature gate.
	//
	// This field is immutable. It can only be set for containers.
	//
	// +featureGate=DynamicResourceAllocation
	// +optional
	Claim string `json:"claim,omitempty"`
}

type HostDevice struct {
	Name string `json:"name"`
	// DeviceName is the resource name of the host device exposed by a device plugin
	DeviceName string `json:"deviceName"`
	// If specified, the virtual network interface address and its tag will be provided to the guest via config drive
	// +optional
	Tag string `json:"tag,omitempty"`
	// Claim is the name of resource, defined in spec.resourceClaims,
	// that is used to provision this host device.
	//
	// This is an alpha field and requires enabling the
	// DynamicResourceAllocation feature gate.
	//
	// This field is immutable. It can only be set for containers.
	//
	// +featureGate=DynamicResourceAllocation
	// +optional
	Claim string `json:"claim,omitempty"`
}
```

The first section vmi.spec.resourceClaims will have a list of devices needed to be allocated for the VM. Having this 
available as a list is important for the following use-cases:
1. A user can use the device available from this list in GPU section or HostDevices section of the API
2. A developer can use the list of for other extensions like checking for availability of resources
3. An admin can write policies on what kind of device (and drivers) are available in the cluster and hence allowed

The API changes here are directly modeled from the Pod Spec. While the use-cases for DRA in a pod and in a VM might
differ, it is important to have similarity between the pod spec and vmi spec. For example, the newer versions of DRA
allows for multiple devices in a single resource claim. The proposed API could be easily re-modeled for the expected 
changes to DRA

The second sections allows for the resource claim to be used in the spec.domain.devices section. The two uses cases 
currently handled by the design are:
1. allowing the devices to be used as a gpu device (spec.domain.devices.gpu)
2. allowing the devices to be used as a host device (spec.domain.device.hostDevices)

### Examples

### VM API with PassThrough GPU

```
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
name: gpu.resource.nvidia.com
driverName: gpu.resource.nvidia.com
---
apiVersion: rtx4090.gpu.resource.nvidia.com/v1
kind: ClaimParameters
name: rtx4090-claim-parameters
spec:
  driver: vfio
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
  name: rtx4090-claim-template
spec:
  spec:
    resourceClassName: gpu.resource.nvidia.com
    parametersRef:
      apiGroup: rtx4090.gpu.resource.nvidia.com/v1
      kind: ClaimParameters
      name: rtx4090-claim-parameters
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
      - name: rtx4090
        source:
          resourceClaimTemplateName: rtx4090-claim-template
      domain:
        devices:
          gpu:
            name: pgpu
            claim: rtx4090
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
      - name: rtx4090
  resourceClaims:
  - name: rtx4090
    source:
      resourceClaimTemplateName: rtx4090-claim-template
```

### VM API with vGPU

```
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
name: gpu.resource.nvidia.com
driverName: gpu.resource.nvidia.com
---
apiVersion: a100.gpu.resource.nvidia.com/v1
kind: ClaimParameters
name: a100-40C-claim-parameters
spec:
  driver: nvidia
  profile: A100DX-40C # Maximum 2 40C vGPUs per GPU
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
  name: a100-40C-claim-template
spec:
  spec:
    resourceClassName: gpu.resource.nvidia.com
    parametersRef:
      apiGroup: a100.gpu.resource.nvidia.com/v1
      kind: ClaimParameters
      name: a100-40C-claim-parameters
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
      - name: a100-40C
        source:
          resourceClaimTemplateName: a100-40C-claim-template
      domain:
        devices:
          gpu:
            name: vgpu
            claim: a100-40C
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
      - name: a100-40C
  resourceClaims:
  - name: a100-40C
    source:
      resourceClaimTemplateName: a100-40C-claim-template
```

### VM API with a GPU and a Host device
```
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
metadata:
  name: pci.kubevirt.io
driverName: pci.resource.kubevirt.io
parametersRef:
  apiGroup: pci.resource.kubevirt.io
  kind: PciClassParameters
  name: pci-params
---
apiVersion: pci.resource.kubevirt.io/v1alpha1
kind: PciClaimParameters
metadata:
  name: nvme-params
  namespace: pci-nvme-test1
spec:
  count: 1
  deviceName: "devices.kubevirt.io/nvme"
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
 namespace: pci-nvme-test1
 name: test-pci-claim-template
spec:
 spec:
   resourceClassName: pci.kubevirt.io
   parametersRef:
     apiGroup: pci.resource.kubevirt.io
     kind: PciClaimParameters
     name: nvme-params
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClass
name: gpu.resource.nvidia.com
driverName: gpu.resource.nvidia.com
---
apiVersion: a100.gpu.resource.nvidia.com/v1
kind: ClaimParameters
name: a100-40C-claim-parameters
spec:
  driver: nvidia
  profile: A100DX-40C # Maximum 2 40C vGPUs per GPU
---
apiVersion: resource.k8s.io/v1alpha2
kind: ResourceClaimTemplate
metadata:
  name: a100-40C-claim-template
spec:
  spec:
    resourceClassName: gpu.resource.nvidia.com
    parametersRef:
      apiGroup: a100.gpu.resource.nvidia.com/v1
      kind: ClaimParameters
      name: a100-40C-claim-parameters
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
      - name: a100-40C
        source:
          resourceClaimTemplateName: a100-40C-claim-template
      - name: nvme1
        source:
          resourceClaimTemplateName: test-pci-claim-template
      domain:
        devices:
          gpu:
            name: vgpu
            claim: a100-40C
          hostDevices:
          - name: example-host-device
            claim: nvme1
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
      - name: a100-40C
      - name: pci
  resourceClaims:
  - name: a100-40C
    source:
      resourceClaimTemplateName: a100-40C-claim-template
  - name: pci
    source:
      resourceClaimTemplateName: test-pci-claim-template
```

# References

- Structured parameters
https://github.com/kubernetes/kubernetes/pull/123516
- Structured parameters KEP
https://github.com/kubernetes/enhancements/issues/4381
- DRA
https://kubernetes.io/docs/concepts/scheduling-eviction/dynamic-resource-allocation/
