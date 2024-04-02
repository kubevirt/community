# Overview

A common pattern for IaaS is to have abstractions separating the resource sizing and performance of a workload from the user defined values related to launching their custom application. This pattern is evident across all the major cloud providers (also known as hyperscalers) as well as open source IaaS projects like OpenStack. AWS has [instance types](https://aws.amazon.com/ec2/instance-types/), GCP has [machine types](https://cloud.google.com/compute/docs/machine-types#custom_machine_types), Azure has [instance VM sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes) and OpenStack has [flavors](https://docs.openstack.org/nova/latest/user/flavors.html).

Let’s take AWS for example to help visualize what this abstraction enables. Launching an EC2 instance only requires a few top level arguments, the disk image, instance type, keypair, security group, and subnet: 

```bash
$ aws ec2 run-instances --image-id ami-xxxxxxxx \
                        --count 1 \
                        --instance-type c4.xlarge \
                        --key-name MyKeyPair \
                        --security-group-ids sg-903004f8 \
                        --subnet-id subnet-6e7f829e
```

When creating the EC2 instance the user doesn't define the amount of resources, what processor to use, how to optimize the performance of the instance, or what hardware to schedule the instance on. Instead all of that information is wrapped up in that single `--instance-type c4.xlarge` CLI argument. `c4` denoting a specific performance profile version, in this case from the `Compute Optimized` family and `xlarge` denoting a specific amount of compute resources provided by the instance type, in this case 4 vCPUs,	7.5 GiB of RAM, 750 Mbps EBS bandwidth etc.

While hyperscalers can provide predefined types with performance profiles and compute resources already assigned IaaS and virtualization projects such as OpenStack and KubeVirt can only provide the raw abstractions for operators, admins and even vendors to then create instances of these abstractions specific to each deployment.

## Problem Statement

KubeVirt's [`VirtualMachine`](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachine) API contains many advanced options for tuning a virtual machine performance that goes beyond what typical users need to be aware of. Users are unable to simply define the storage/network they want assigned to their VM and then declare in broad terms what quality of resources and kind of performance they need for their VM.

Instead, the user has to be keenly aware how to request specific compute resources alongside all of the performance tunings available on the [`VirtualMachine`](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachine) API and how those tunings impact their guest’s operating system in order to get a desired result.

The [partially implemented currently v1alpha1 `Virtual Machine Flavors` API](https://github.com/kubevirt/kubevirt/commit/6dc548459fa17d7f9601cbc251088d2b70a2a96a) was an attempt to provide operators and users with a mechanism to define resource buckets that could be used during VM creation. At present this implementation provides a cluster-wide [`VirtualMachineClusterFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineclusterflavor) and a namespaced [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavor) CRDs. Each containing an array of [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavorprofile) that at present only encapsulates CPU resources by applying a full copy of the [`CPU`](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpu) type to the [`VirtualMachineInstance`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstance) at runtime.

This approach has a few pitfalls such as:

* Using embedded profiles within the CRDs

* Tight coupling between the resource definition of a `VirtualMachineFlavorProfile` and workload preferences it might also contain that are required to run a specific `VirtualMachine`.

* This results in a reliance on the user selecting the correct [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavorprofile) for their specific workload.

* Not allowing a user to override some viable `VirtualMachineInstanceSpec` attributes at runtime

## User Stories

* As an operator, user or vendor I would like to provide flavors that define the resources available to a VM along with any schedulable *or* performance related attributes.

* As an operator, user or vendor I would like to separately provide preferences that relate to ensuring that a specific guest OS within the VM is able to run correctly.

* As an user I would like to be presented with and select from pre-defined options that encapsulate the resources, performance and any additional preferences required to run my workload.

## Goals

* Simplify the choices a user has to make in order to get a runnable `VirtualMachine` with the required amount of resources, desired performance and additional preferences to correctly run a given workload.

## Non-Goals

The following items have been discussed alongside `Virtual Machine Flavor` and `Virtual Machine Preferences` but are not within the current scope of this design proposal:

* Introspection of imported images to determine the correct guest OS related `VirtualMachinePreferences` to apply.

* Using image labels to determine the correct guest OS related `VirtualMachinePreferences` to apply.

* Remove the need to define [`Disks`](http://kubevirt.io/api-reference/main/definitions.html#_v1_disk) within [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) when providing [`Volumes`](http://kubevirt.io/api-reference/main/definitions.html#_v1_volume) within a [`VirtualMachineInstanceSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstancespec).

* Remove the need to define [`Interfaces`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstancenetworkinterface) within [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) when providing [`Networks`](http://kubevirt.io/api-reference/main/definitions.html#_v1_network) within a [`VirtualMachineInstanceSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstancespec).

* Versioning of these CRDs through [Controller Revisions](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#controllerrevision-v1-apps) to ensure the generated [`VirtualMachineInstanceSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstancespec) of a `VMI` doesn't change over time if the CRDs are modified. This will be documented in a separate design proposal.

## Repos

* [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design

As stated above the ultimate goal of this work is to provide the end user with a simple set of choices when defining a `VirtualMachine` within KubeVirt. To achieve this the existing [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavor) CRDs will be heavily modified and extended to better encapsulate resource, performance or schedulable attributes of a VM.

A quick note on the original `Flavor` implementation before we continue. At present the [`FlavorMatcher`](http://kubevirt.io/api-reference/main/definitions.html#_v1_flavormatcher) type is used lookup a given flavor when admitting the initial `VirtualMachine` resource or when later starting the `VirtualMachine`. During the latter the flavor is applied to the associated `VirtualMachineInstance` object by the `VM controller` before the actual request to create the `VirtualMachineInstance` resource is submitted to the API.

For now this aspect of the design will stay but could be changed in future to allow the `Flavor` to be applied to the `VirtualMachineInstance` *after* submission to the API. This would allow users creating `VirtualMachine` and `VirtualMachineInstance` resources to have the same overall experience while ensuring things are expanded before any underlying `VM` launches.

For now however this proposal will start with the removal of the embedded `VirtualMachineFlavorProfile` type within the CRDs, this will be replaced with a singular `VirtualMachineFlavorSpec` type per flavor. The decision to remove `VirtualMachineFlavorProfile` has been made as the concept isn't prevalent within the wider Kubernetes ecosystem and could be confusing to end users. Instead users looking to avoid duplication when defining flavors will be directed to use tools such as [`kustomize`](https://kustomize.io/) to generate their flavors. This tooling is already commonly used when defining resources within Kubernetes and should afford users plenty of flexibility when defining their flavors either statically or as part of a larger GitOps based workflow.

`VirtualMachineFlavorSpec` will also include elements of `CPU`, `Devices`, `HostDevices`, `GPUs`, `Memory` and `LaunchSecurity` defined fully below. Users will be unable to override any aspect of the flavor (for example, `vCPU` count or amount of `Memory`) within the `VirtualMachine` itself, any attempt to do so resulting in the `VirtualMachine` being rejected.

A new set of `VirtualMachinePreference` CRDs will then be introduced to define any remaining attributes related to ensuring the selected guestOS can run. As the name suggests the `VirtualMachinePreference` CRDs will only define preferences, so unlike a flavor if a preference conflicts with something user defined within the `VirtualMachine` it will be ignored. For example, if a user selects a `VirtualMachinePreference` that requests a `preferredDiskBus` of `virtio` but then sets a disk bus of `SATA` for one or more disk devices within the `VirtualMachine` the supplied `preferredDiskBus` preference will not be applied to these disks. Any remaining disks that do not have a disk bus defined will however use the `preferredDiskBus` preference of `virtio`.

## Implementation Phases

* Remove [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavorprofile) from [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavor) and [`VirtualMachineClusterFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineclusterflavor).

* Introduce a new `Spec` attribute and new `VirtualMachineFlavorSpec` type  itself with `CPU`, `Memory`, `GPUs` and `HostDevices` attributes to [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavor) and [`VirtualMachineClusterFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineclusterflavor).

* Introduce new `VirtualMachinePreference` and `VirtualMachineClusterPreference` CRDs to the Flavors API and `VirtualMachinePreferenceSpec` type to define guestOS related preferences such as preferred device buses and models.

* Allow for `CPU` preferred topology policy to be defined within `VirtualMachinePreferenceSpec` through a `PreferredTopology` attribute.

* Extend [`VirtualMachineSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachinespec) to allow an optional reference to `VirtualMachinePreference` or `VirtualMachineClusterPreference`.

## [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) mappings to `VirtualMachinePreferenceSpec` and `VirtualMachineFlavorSpec`

The following basic rules are applied when deciding which attributes of the [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) will be mapped to the new `VirtualMachinePreferenceSpec` and `VirtualMachineFlavorSpec` types:

* Resource, schedulable and performance related attributes of the [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) reside within `VirtualMachineFlavorSpec`

* Guest OS related preferences reside within `VirtualMachinePreferenceSpec`

For example, the number of and topology of vCPUs presented to the guest are defined within `VirtualMachineFlavorSpec` but the preferred topology of these vCPUs resides within  `VirtualMachinePreferenceSpec`.

* [DomainSpec](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec)
  * [CPU](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpu)
    * cores 
      * `VirtualMachineFlavorSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets` , `preferThreads` or `preferCores` (default).
    * sockets
      * `VirtualMachineFlavorSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets`, `preferThreads`  or `preferCores` (default).
    * threads
      * `VirtualMachineFlavorSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets`, `preferThreads`  or `preferCores` (default).
    * dedicatedCpuPlacement
      * `VirtualMachineFlavorSpec.CPU.DedicatedCpuPlacement`
    * [features](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpufeature)
      * `VirtualMachineFlavorSpec.CPU.Features`
    * isolateEmulatorThread
      * `VirtualMachineFlavorSpec.CPU.IsolateEmulatorThread`
    * model
      * `VirtualMachineFlavorSpec.CPU.Model`
    * [NUMA](http://kubevirt.io/api-reference/main/definitions.html#_v1_watchdog)
      * `VirtualMachineFlavorSpec.CPU.NUMA`
    * [Realtime](http://kubevirt.io/api-reference/main/definitions.html#_v1_realtime)
      * `VirtualMachineFlavorSpec.CPU.Realtime`
  * [Devices](http://kubevirt.io/api-reference/main/definitions.html#_v1_devices)
    * autoattachGraphicsDevice
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachGraphicsDevice`
    * autoattachMemBalloon
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachMemBalloon`
    * autoattachPodInterface
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachPodInterface`
    * autoattachSerialConsole
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachSerialConsole`
    * blockMultiQueue
      * `VirtualMachineFlavorSpec.Devices.BlockMultiQueue`
    * clientPassthrough
      * None
    * disableHotplug
      * `VirtualMachinePreferenceSpec.Devices.PreferredDisableHotplug`
    * [Disks](http://kubevirt.io/api-reference/main/definitions.html#_v1_disk)
      * [blockSize](http://kubevirt.io/api-reference/main/definitions.html#_v1_blocksize) 
        * `VirtualMachinePreferenceSpec.Devices.PreferredDiskBlockSize`
      * bootOrder
        * None
      * cache
        * `VirtualMachinePreferenceSpec.Devices.PreferredDiskCache`
      * [Cdrom](http://kubevirt.io/api-reference/main/definitions.html#_v1_cdromtarget)
        * Bus
          * `VirtualMachinePreferenceSpec.Devices.PreferredCdRomBus`
        * readonly
          * None
        * Tray
          * None
      * dedicatedIOThread
        * `VirtualMachineFlavorSpec.Devices.PreferredDedicatedIOThread`
      * [Disk](http://kubevirt.io/api-reference/main/definitions.html#_v1_disktarget)
        * Bus
          * `VirtualMachinePreferenceSpec.Devices.PreferredDiskBus`
        * pciAddress
          * None
        * readonly
          * None
      * io
        * `VirtualMachinePreferenceSpec.Devices.PreferredIOMode`
      * [Lun](http://kubevirt.io/api-reference/main/definitions.html#_v1_luntarget)
        * bus
          * `VirtualMachinePreferenceSpec.Devices.PreferredLunBus`
        * readOnly
          * None
      * name
        * None
      * serial
        * None
      * shareable
        * None
      * tag
        * None
      * [Filesystems](http://kubevirt.io/api-reference/main/definitions.html#_v1_filesystem)
        * Name
          * None
        * virtiofs
          * None
      * [Gpus](http://kubevirt.io/api-reference/main/definitions.html#_v1_gpu)
        * deviceName
          * `VirtualMachineFlavorSpec.GPUs`
        * name
          * None
        * tag
          * None
        * virtualGPUOptions
          * `VirtualMachinePreferenceSpec.Devices.PreferredVirtualGPUOptions`
      * [HostDevices](http://kubevirt.io/api-reference/main/definitions.html#_v1_hostdevice)
        * deviceName
          * `VirtualMachineFlavorSpec.HostDevices`
        * name
          * None
        * tag
          * None
      * [Inputs](http://kubevirt.io/api-reference/main/definitions.html#_v1_input)
        * bus
          * `VirtualMachinePreferenceSpec.Devices.PreferredInputBus`
        * Name
          * None
        * Type
          * `VirtualMachinePreferenceSpec.Devices.PreferredInputType`
      * [Interfaces](http://kubevirt.io/api-reference/main/definitions.html#_v1_interface)
        * bootOrder
          * None
        * bridge
          * None
        * dhcpoptions
          * None
        * macaddress
          * None
        * [macvtap](http://kubevirt.io/api-reference/main/definitions.html#_v1_interfacemacvtap) 
          * None
        * [masquerade](http://kubevirt.io/api-reference/main/definitions.html#_v1_interfacemasquerade) 
          * None
        * model
          * `VirtualMachinePreferenceSpec.Devices.PreferredInterfaceModel`
        * name
          * None
        * pciAddress
          * None
        * ports
          * None
        * [slirp](http://kubevirt.io/api-reference/main/definitions.html#_v1_interfaceslirp) 
          * None
        * [sriov](http://kubevirt.io/api-reference/main/definitions.html#_v1_interfacesriov) 
          * None
        * Tag
          * None
      * networkInterfaceMultiqueue
        * `VirtualMachinePreferenceSpec.Devices.PreferredNetworkInterfaceMultiqueue`
      * [Rng](http://kubevirt.io/api-reference/main/definitions.html#_v1_rng)
        * `VirtualMachinePreferenceSpec.Devices.PreferredRNG`
      * [Sound](http://kubevirt.io/api-reference/main/definitions.html#_v1_sounddevice)
        * model
          * `VirtualMachinePreferenceSpec.Devices.PreferredSoundModel`
        * name
          * None
      * [TPM](http://kubevirt.io/api-reference/main/definitions.html#_v1_tpmdevice)
        * `VirtualMachinePreferenceSpec.Devices.PreferredTPM`
      * useVirtioTransitional
        * `VirtualMachinePreferenceSpec.Devices.PreferredUseVirtioTransitional`
      * [Watchdog](http://kubevirt.io/api-reference/main/definitions.html#_v1_watchdog)
        * [I6300esb](http://kubevirt.io/api-reference/main/definitions.html#_v1_i6300esbwatchdog)
          * action
            * None
        * name
          * None
  * [Features](http://kubevirt.io/api-reference/main/definitions.html#_v1_features)
    * [acpi](http://kubevirt.io/api-reference/main/definitions.html#_v1_featurestate)
      * `VirtualMachinePreferenceSpec.Features.PreferredAcpi`
    * [apic](http://kubevirt.io/api-reference/main/definitions.html#_v1_featureapic)
      * `VirtualMachinePreferenceSpec.Features.PreferredApic`
    * [hyperv](http://kubevirt.io/api-reference/main/definitions.html#_v1_featurehyperv) 
      * `VirtualMachinePreferenceSpec.Features.PreferredHyperv`
    * [kvm](http://kubevirt.io/api-reference/main/definitions.html#_v1_featurekvm)
      * `VirtualMachinePreferenceSpec.Features.PreferredKvm`
    * [pvspinlock](http://kubevirt.io/api-reference/main/definitions.html#_v1_featurestate) 
      * `VirtualMachinePreferenceSpec.Features.PreferredPvspinlock`
    * [smm](http://kubevirt.io/api-reference/main/definitions.html#_v1_featurestate)
      * `VirtualMachinePreferenceSpec.Features.PreferredSmm`
  * [Firmware](http://kubevirt.io/api-reference/main/definitions.html#_v1_firmware)
    * [bootloader](http://kubevirt.io/api-reference/main/definitions.html#_v1_bootloader)
      * [bios](http://kubevirt.io/api-reference/main/definitions.html#_v1_bios)
        * `VirtualMachinePreferenceSpec.Firmware.PreferredUseBios`
        * useSerial
          * `VirtualMachinePreferenceSpec.Firmware.PreferredUseBiosSerial`
      * [efi](http://kubevirt.io/api-reference/main/definitions.html#_v1_efi)
        * `VirtualMachinePreferenceSpec.Firmware.PreferredUseEfi`
        * secureBoot
          * `VirtualMachinePreferenceSpec.Firmware.PreferredUseEfiSecureBoot`
    * [kernelBoot](http://kubevirt.io/api-reference/main/definitions.html#_v1_kernelboot)
      * [container](http://kubevirt.io/api-reference/main/definitions.html#_v1_kernelbootcontainer)
        * image
          * None
        * imagePullPolicy
          * None
        * imagePullSecret
          * None
        * initrdPath
          * None
        * kernelPath
          * None
      * kernelArgs
        * None
    * serial
      * None
    * uuid
      * None
  * ioThreadsPolicy
    * `VirtualMachineFlavorSpec.ioThreadsPolicy`
  * [launchSecurity](http://kubevirt.io/api-reference/main/definitions.html#_v1_launchsecurity)
    * [sev](http://kubevirt.io/api-reference/main/definitions.html#_v1_sev)
      * `VirtualMachineFlavorSpec.LaunchSecurity.SEV`
  * [Machine](http://kubevirt.io/api-reference/main/definitions.html#_v1_machine)
    * type
      * `VirtualMachinePreferenceSpec.Machine.PreferredMachineType`
  * [Memory](http://kubevirt.io/api-reference/main/definitions.html#_v1_memory)
    * [guest](http://kubevirt.io/api-reference/main/definitions.html#_k8s_io_apimachinery_pkg_api_resource_quantity) 
      * `VirtualMachineFlavorSpec.Memory.Guest`
    * [hugepages](http://kubevirt.io/api-reference/main/definitions.html#_v1_hugepages)
      * `VirtualMachineFlavorSpec.Memory.HugePages`

## Types

### `VirtualMachineFlavorSpec`

```go
// VirtualMachineFlavorSpec
//
// +k8s:openapi-gen=true1
type VirtualMachineFlavorSpec struct {

	// Defines the CPU related attributes of the flavor, required.
	CPU CPUFlavor `json:"cpu"`

	// Defines the Memory related attributes of the flavor, required.
	Memory MemoryFlavor `json:"memory"`

	// Optionally defines any GPU devices associated with the flavor.
	//
	// +optional
	// +listType=atomic
	GPUs []v1.GPU `json:"gpus,omitempty"`

	// Optionally defines any HostDevices associated with the flavor.
	//
	// +optional
	// +listType=atomic
	HostDevices []v1.HostDevice `json:"hostDevices,omitempty"`

	// Optionally defines the IOThreadsPolicy to be used by the flavor.
	//
	// +optional
	IOThreadsPolicy *v1.IOThreadsPolicy `json:"ioThreadsPolicy,omitempty"`

	// Optionally defines the LaunchSecurity to be used by the flavor.
	//
	// +optional
	LaunchSecurity *v1.LaunchSecurity `json:"launchSecurity,omitempty"`
}
```

### `CPUFlavor`

```go
// CPUFlavor
//
// +k8s:openapi-gen=true
type CPUFlavor struct {

	// Number of vCPUs to expose to the guest.
	// The resulting CPU topology being derived from the optional PreferredCPUTopology attribute of CPUPreferences.
	Guest uint32 `json:"guest"`

	// Model specifies the CPU model inside the VMI.
	// List of available models https://github.com/libvirt/libvirt/tree/master/src/cpu_map.
	// It is possible to specify special cases like "host-passthrough" to get the same CPU as the node
	// and "host-model" to get CPU closest to the node one.
	// Defaults to host-model.
	// +optional
	Model string `json:"model,omitempty"`

	// DedicatedCPUPlacement requests the scheduler to place the VirtualMachineInstance on a node
	// with enough dedicated pCPUs and pin the vCPUs to it.
	// +optional
	DedicatedCPUPlacement bool `json:"dedicatedCPUPlacement,omitempty"`

	// NUMA allows specifying settings for the guest NUMA topology
	// +optional
	NUMA *v1.NUMA `json:"numa,omitempty"`

	// IsolateEmulatorThread requests one more dedicated pCPU to be allocated for the VMI to place
	// the emulator thread on it.
	// +optional
	IsolateEmulatorThread bool `json:"isolateEmulatorThread,omitempty"`

	// Realtime instructs the virt-launcher to tune the VMI for lower latency, optional for real time workloads
	// +optional
	Realtime *v1.Realtime `json:"realtime,omitempty"`
}
```

### `MemoryFlavor`

```go
// FlavorMemory
//
// +k8s:openapi-gen=true
type MemoryFlavor struct {

	// Guest allows to specifying the amount of memory which is visible inside the Guest OS.
	Guest *resource.Quantity `json:"guest,omitempty"`

	// Hugepages allow to use hugepages for the VirtualMachineInstance instead of regular memory.
	// +optional
	Hugepages *v1.Hugepages `json:"hugepages,omitempty"`
}
```

###  [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineflavor) and [`VirtualMachineClusterFlavor`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineclusterflavor)

```go
// VirtualMachineFlavor resource contains common VirtualMachine configuration
// that can be used by multiple VirtualMachine resources.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +genclient
type VirtualMachineFlavor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// VirtualMachineFlavorSpec for the flavor
	Spec VirtualMachineFlavorSpec `json:"spec"`
}
```

```go
// VirtualMachineClusterFlavor is a cluster scoped version of VirtualMachineFlavor resource.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +genclient
// +genclient:nonNamespaced
type VirtualMachineClusterFlavor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// VirtualMachineFlavorSpec for the flavor
	Spec VirtualMachineFlavorSpec `json:"spec"`
}
```

###  `VirtualMachinePreferenceSpec`

```go
// VirtualMachinePreferenceSpec
//
// +k8s:openapi-gen=true
type VirtualMachinePreferenceSpec struct {

	// Clock optionally defines preferences associated with the Clock attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	Clock *ClockPreferences `json:"clock,omitempty"`

	// CPU optionally defines preferences associated with the CPU attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	CPU *CPUPreferences `json:"cpu,omitempty"`

	// Devices optionally defines preferences associated with the Devices attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	Devices *DevicePreferences `json:"devices,omitempty"`

	// Features optionally defines preferences associated with the Features attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	Features *FeaturePreferences `json:"features,omitempty"`

	// Firmware optionally defines preferences associated with the Firmware attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	Firmware *FirmwarePreferences `json:"firmware,omitempty"`

	// Machine optionally defines preferences associated with the Machine attribute of a VirtualMachineInstance DomainSpec
	//
	//+optional
	Machine *MachinePreferences `json:"machine,omitempty"`
}
```

### `ClockPreferences`

```go
// ClockPreferences contains various optional defaults for Clock.
//
// +k8s:openapi-gen=true
type ClockPreferences struct {

	// ClockOffset allows specifying the UTC offset or the timezone of the guest clock.
	//
	// +optional
	PreferredClockOffset *v1.ClockOffset `json:"preferredClockOffset,omitempty"`

	// Timer specifies whih timers are attached to the vmi.
	//
	// +optional
	PreferredTimer *v1.Timer `json:"preferredTimer,omitempty"`
}
```

### `CPUPreferences`

```go
// PreferredCPUTopology defines a preferred CPU topology to be exposed to the guest
//
// +k8s:openapi-gen=true
type PreferredCPUTopology string

const (

	// Prefer vCPUs to be exposed as cores to the guest, this is the PreferredCPUTopology default.
	PreferCores PreferredCPUTopology = "preferCores"

	// Prefer vCPUs to be exposed as sockets to the guest
	PreferSockets PreferredCPUTopology = "preferSockets"

	// Prefer vCPUs to be exposed as threads to the guest
	PreferThreads PreferredCPUTopology = "preferThreads"
)

// CPUPreferences contains various optional CPU preferences.
//
// +k8s:openapi-gen=true
type CPUPreferences struct {

	// PreferredCPUTopology optionally defines the preferred guest visible CPU topology, defaults to PreferCores.
	//
	//+optional
	PreferredCPUTopology PreferredCPUTopology `json:"preferredCPUTopology,omitempty"`
}
```

### `DevicePreferences`

```go
// DevicePreferences contains various optional defaults for Devices.
//
// +k8s:openapi-gen=true
type DevicePreferences struct {

	// PreferredAutoattachGraphicsDevice optionally defines the preferred value of AutoattachGraphicsDevice
	//
	// +optional
	PreferredAutoattachGraphicsDevice *bool `json:"preferredAutoattachGraphicsDevice,omitempty"`

	// PreferredAutoattachMemBalloon optionally defines the preferred value of AutoattachMemBalloon
	//
	// +optional
	PreferredAutoattachMemBalloon *bool `json:"preferredAutoattachMemBalloon,omitempty"`

	// PreferredAutoattachPodInterface optionally defines the preferred value of AutoattachPodInterface
	//
	// +optional
	PreferredAutoattachPodInterface *bool `json:"preferredAutoattachPodInterface,omitempty"`

	// PreferredAutoattachSerialConsole optionally defines the preferred value of AutoattachSerialConsole
	//
	// +optional
	PreferredAutoattachSerialConsole *bool `json:"preferredAutoattachSerialConsole,omitempty"`

	// PreferredDisableHotplug optionally defines the preferred value of DisableHotplug
	//
	// +optional
	PreferredDisableHotplug *bool `json:"preferredDisableHotplug,omitempty"`

	// PreferredVirtualGPUOptions optionally defines the preferred value of VirtualGPUOptions
	//
	// +optional
	PreferredVirtualGPUOptions *v1.VGPUOptions `json:"preferredVirtualGPUOptions,omitempty"`

	// PreferredSoundModel optionally defines the preferred model for Sound devices.
	//
	// +optional
	PreferredSoundModel string `json:"preferredSoundModel,omitempty"`

	// PreferredUseVirtioTransitional optionally defines the preferred value of UseVirtioTransitional
	//
	// +optional
	PreferredUseVirtioTransitional *bool `json:"preferredUseVirtioTransitional,omitempty"`

	// PreferredInputBus optionally defines the preferred bus for Input devices.
	//
	// +optional
	PreferredInputBus string `json:"preferredInputBus,omitempty"`

	// PreferredInputType optionally defines the preferred type for Input devices.
	//
	// +optional
	PreferredInputType string `json:"preferredInputType,omitempty"`

	// PreferredDiskBus optionally defines the preferred bus for Disk Disk devices.
	//
	// +optional
	PreferredDiskBus string `json:"preferredDiskBus,omitempty"`

	// PreferredLunBus optionally defines the preferred bus for Lun Disk devices.
	//
	// +optional
	PreferredLunBus string `json:"preferredLunBus,omitempty"`

	// PreferredCdromBus optionally defines the preferred bus for Cdrom Disk devices.
	//
	// +optional
	PreferredCdromBus string `json:"preferredCdromBus,omitempty"`

	// PreferredDedicatedIoThread optionally enables dedicated IO threads for Disk devices.
	//
	// +optional
	PreferredDiskDedicatedIoThread *bool `json:"preferredDiskDedicatedIoThread,omitempty"`

	// PreferredCache optionally defines the DriverCache to be used by Disk devices.
	//
	// +optional
	PreferredDiskCache v1.DriverCache `json:"preferredDiskCache,omitempty"`

	// PreferredIo optionally defines the QEMU disk IO mode to be used by Disk devices.
	//
	// +optional
	PreferredDiskIO v1.DriverIO `json:"preferredDiskIO,omitempty"`

	// PreferredBlockSize optionally defines the block size of Disk devices.
	//
	// +optional
	PreferredDiskBlockSize *v1.BlockSize `json:"preferredDiskBlockSize,omitempty"`

	// PreferredInterfaceModel optionally defines the preferred model to be used by Interface devices.
	//
	// +optional
	PreferredInterfaceModel string `json:"preferredInterfaceModel,omitempty"`

	// PreferredRng optionally defines the preferred rng device to be used.
	//
	// +optional
	PreferredRng *v1.Rng `json:"preferredRng,omitempty"`

	// PreferredBlockMultiQueue optionally enables the vhost multiqueue feature for virtio disks.
	//
	// +optional
	PreferredBlockMultiQueue *bool `json:"preferredBlockMultiQueue,omitempty"`

	// PreferredNetworkInterfaceMultiQueue optionally enables the vhost multiqueue feature for virtio interfaces.
	//
	// +optional
	PreferredNetworkInterfaceMultiQueue *bool `json:"preferredNetworkInterfaceMultiQueue,omitempty"`

	// PreferredTPM optionally defines the preferred TPM device to be used.
	//
	// +optional
	PreferredTPM *v1.TPMDevice `json:"preferredTPM,omitempty"`
}
```

### `FeaturePreferences`

```go
// FeaturePreferences contains various optional defaults for Features.
//
// +k8s:openapi-gen=true
type FeaturePreferences struct {

	// PreferredAcpi optionally enables the ACPI feature
	//
	// +optional
	PreferredAcpi *v1.FeatureState `json:"preferredAcpi,omitempty"`

	// PreferredApic optionally enables and configures the APIC feature
	//
	// +optional
	PreferredApic *v1.FeatureAPIC `json:"preferredApic,omitempty"`

	// PreferredHyperv optionally enables and configures HyperV features
	//
	// +optional
	PreferredHyperv *v1.FeatureHyperv `json:"preferredHyperv,omitempty"`

	// PreferredKvm optionally enables and configures KVM features
	//
	// +optional
	PreferredKvm *v1.FeatureKVM `json:"preferredKvm,omitempty"`

	// PreferredPvspinlock optionally enables the Pvspinlock feature
	//
	// +optional
	PreferredPvspinlock *v1.FeatureState `json:"preferredPvspinlock,omitempty"`

	// PreferredSmm optionally enables the SMM feature
	//
	// +optional
	PreferredSmm *v1.FeatureState `json:"preferredSmm,omitempty"`
}
```

### `FirmwarePreferences`

```go
// FirmwarePreferences contains various optional defaults for Firmware.
//
// +k8s:openapi-gen=true
type FirmwarePreferences struct {

	// PreferredUseBios optionally enables BIOS
	//
	// +optional
	PreferredUseBios *bool `json:"preferredUseBios,omitempty"`

	// PreferredUseBiosSerial optionally transmitts BIOS output over the serial.
	//
	// Requires PreferredUseBios to be enabled.
	//
	// +optional
	PreferredUseBiosSerial *bool `json:"preferredUseBiosSerial,omitempty"`

	// PreferredUseEfi optionally enables EFI
	//
	// +optional
	PreferredUseEfi *bool `json:"preferredUseEfi,omitempty"`

	// PreferredUseSecureBoot optionally enables SecureBoot and the OVMF roms will be swapped for SecureBoot-enabled ones.
	//
	// Requires PreferredUseEfi and PreferredSmm to be enabled.
	//
	// +optional
	PreferredUseSecureBoot *bool `json:"preferredUseSecureBoot,omitempty"`
}
```

### `MachinePreferences`

```go
// MachinePreferences contains various optional defaults for Machine.
//
// +k8s:openapi-gen=true
type MachinePreferences struct {

	// PreferredMachineType optionally defines the preferred machine type to use.
	//
	// +optional
	PreferredMachineType string `json:"preferredMachineType,omitempty"`
}
```

### `VirtualMachinePreference` and `VirtualMachineClusterPreference`

```go
// VirtualMachinePreference
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +genclient
type VirtualMachinePreference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *VirtualMachinePreferenceSpec `json:"spec"`
}
```

```go
// VirtualMachineClusterPreference
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:openapi-gen=true
// +genclient
// +genclient:nonNamespaced
type VirtualMachineClusterPreference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec *VirtualMachinePreferenceSpec `json:"spec"`
}
```

### `VirtualMachinePreferenceMatcher`

```go
// PreferenceMatcher references a set of preference that is used to fill fields in the VMI template.
type PreferenceMatcher struct {
	// Name is the name of the VirtualMachinePreference or VirtualMachineClusterPreference
	Name string `json:"name"`

	// Kind specifies which preference resource is referenced.
	// Allowed values are: "VirtualMachinePreference" and "VirtualMachineClusterPreference".
	// If not specified, "VirtualMachineClusterPreference" is used by default.
	//
	// +optional
	Kind string `json:"kind,omitempty"`
}
```

### `VirtualMachine`

```go
// VirtualMachineSpec describes how the proper VirtualMachine
// should look like
type VirtualMachineSpec struct {

[..]

	// FlavorMatcher references a flavor that is used to fill fields in Template
	Flavor *FlavorMatcher `json:"flavor,omitempty" optional:"true"`

	// PreferenceMatcher references a set of preference that is used to fill fields in Template
	Preference *PreferenceMatcher `json:"preference,omitempty" optional:"true"`

[..]

}
```

## API Examples

The following example shows a `clarge` flavor used alongside `Windows` preferences:

### `VirtualMachineFlavor`

```yaml
---
apiVersion: flavor.kubevirt.io/v1alpha1
kind: VirtualMachineFlavor
metadata:
  name: vmf-clarge
spec:
  cpu:
    guest: 4
  memory:
    guest: 2Gi
```

### `VirtualMachinePreference`

```yaml
---
apiVersion: flavor.kubevirt.io/v1alpha1
kind: VirtualMachinePreference
metadata:
  name: vmpwindows
spec:
  clock:
    preferredClockOffset:
      utc: {}
    preferredTimer:
      hpet:
        present: false
      hyperv: {}
      pit:
        tickPolicy: delay
      rtc:
        tickPolicy: catchup
  cpu:
    preferredCPUTopology: preferSockets
  devices:
    preferredDiskBus: sata
    preferredInterfaceModel: e1000
    preferredTPM: {}
  features:
    preferredAcpi: {}
    preferredApic: {}
    preferredHyperv:
      relaxed: {}
      spinlocks:
        spinlocks: 8191
      vapic: {}
    preferredSmm: {}
  firmware:
    preferredUseEfi: true
    preferredUseSecureBoot: true

```

### `VirtualMachine`

```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm: vm-windows-clarge-windows
  name: vm-windows-clarge-windows
spec:
  flavor:
    kind: VirtualMachineFlavor
    name: vmf-clarge
  preference:
    kind: VirtualMachinePreference
    name: vmpwindows
  running: false
  template:
    metadata:
      labels:
        kubevirt.io/vm: vm-windows-clarge-windows
    spec:
      domain:
        devices:
          disks:
          - disk: {}
            name: containerdisk
        resources: {}
      terminationGracePeriodSeconds: 0
      volumes:
      - containerDisk:
          image: registry:5000/kubevirt/windows-disk:devel
        name: containerdisk
```

### `VirtualMachineInstance`

```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  annotations:
    kubevirt.io/flavor-name: vmf-clarge
    kubevirt.io/latest-observed-api-version: v1
    kubevirt.io/preference-name: vmpwindows
    kubevirt.io/storage-observed-api-version: v1alpha3
  creationTimestamp: "2022-04-19T10:51:53Z"
  finalizers:
  - kubevirt.io/virtualMachineControllerFinalize
  - foregroundDeleteVirtualMachine
  generation: 9
  labels:
    kubevirt.io/nodeName: node01
    kubevirt.io/vm: vm-windows-clarge-windows
  name: vm-windows-clarge-windows
  namespace: default
  ownerReferences:
  - apiVersion: kubevirt.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: VirtualMachine
    name: vm-windows-clarge-windows
    uid: 8974d1e6-5f41-4486-996a-84cd6ebb3b37
  resourceVersion: "8052"
  uid: 369e9a17-8eca-47cc-91c2-c8f12e0f6f9f
spec:
  domain:
    clock:
      timer:
        hpet:
          present: false
        hyperv:
          present: true
        pit:
          present: true
          tickPolicy: delay
        rtc:
          present: true
          tickPolicy: catchup
      utc: {}
    cpu:
      cores: 1
      model: host-model
      sockets: 4
      threads: 1
    devices:
      disks:
      - disk:
          bus: sata
        name: containerdisk
      interfaces:
      - bridge: {}
        name: default
      tpm: {}
    features:
      acpi:
        enabled: true
      apic:
        enabled: true
      hyperv:
        relaxed:
          enabled: true
        spinlocks:
          enabled: true
          spinlocks: 8191
        vapic:
          enabled: true
      smm:
        enabled: true
    firmware:
      bootloader:
        efi:
          secureBoot: true
      uuid: bc694b87-1373-5514-9694-0f495fbae3b2
    machine:
      type: q35
    memory:
      guest: 2Gi
    resources:
      requests:
        memory: 2Gi
  networks:
  - name: default
    pod: {}
  terminationGracePeriodSeconds: 0
  volumes:
  - containerDisk:
      image: registry:5000/kubevirt/windows-disk:devel
      imagePullPolicy: IfNotPresent
    name: containerdisk
status:
  activePods:
    557c7fef-04b2-47c1-880b-396da944a7d3: node01
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2022-04-19T10:51:57Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: null
    message: cannot migrate VMI which does not use masquerade to connect to the pod
      network
    reason: InterfaceNotLiveMigratable
    status: "False"
    type: LiveMigratable
  guestOSInfo: {}
  interfaces:
  - infoSource: domain
    ipAddress: 10.244.196.149
    ipAddresses:
    - 10.244.196.149
    - fd10:244::c494
    mac: 66:f7:21:4e:d9:30
    name: default
  launcherContainerImageVersion: registry:5000/kubevirt/virt-launcher@sha256:40b2036eae39776560a73263198ff42ffd6a8f09c9aa208f8bbdc91ec35b42cf
  migrationMethod: BlockMigration
  migrationTransport: Unix
  nodeName: node01
  phase: Running
  phaseTransitionTimestamps:
  - phase: Pending
    phaseTransitionTimestamp: "2022-04-19T10:51:53Z"
  - phase: Scheduling
    phaseTransitionTimestamp: "2022-04-19T10:51:53Z"
  - phase: Scheduled
    phaseTransitionTimestamp: "2022-04-19T10:51:57Z"
  - phase: Running
    phaseTransitionTimestamp: "2022-04-19T10:51:59Z"
  qosClass: Burstable
  runtimeUser: 0
  virtualMachineRevisionName: revision-start-vm-8974d1e6-5f41-4486-996a-84cd6ebb3b37-2
  volumeStatus:
  - name: cloudinitdisk
    size: 1048576
    target: sdb
```


# Acknowledgements

* David Vossel - Original downstream design proposal
* Fabian Deutsch - Counter downstream design proposal introducing domain preferences
* Andrej Krejcir - Initial implementation of Flavors within KubeVirt
