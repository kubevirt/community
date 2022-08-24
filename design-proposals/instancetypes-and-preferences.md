# Overview

A common pattern for IaaS is to have abstractions separating the resource sizing and performance of a workload from the user defined values related to launching their custom application. This pattern is evident across all the major cloud providers (also known as hyperscalers) as well as open source IaaS projects like OpenStack. AWS has [instance types](https://aws.amazon.com/ec2/instance-types/), GCP has [machine types](https://cloud.google.com/compute/docs/machine-types#custom_machine_types), Azure has [instance`VirtualMachine`sizes](https://docs.microsoft.com/en-us/azure/virtual-machines/sizes) and OpenStack has [flavors](https://docs.openstack.org/nova/latest/user/flavors.html).

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

KubeVirt's [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) API contains many advanced options for tuning a virtual machine performance that goes beyond what typical users need to be aware of. Users are unable to simply define the storage/network they want assigned to their [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) and then declare in broad terms what quality of resources and kind of performance they need for their [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine).

Instead, the user has to be keenly aware how to request specific compute resources alongside all of the performance tunings available on the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) API and how those tunings impact their guest’s operating system in order to get a desired result.

The [partially implemented `Flavor` API](https://github.com/kubevirt/kubevirt/commit/6dc548459fa17d7f9601cbc251088d2b70a2a96a) was an attempt to provide operators and users with a mechanism to define resource buckets that could be used during [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) creation. This implementation provided a cluster-wide [`VirtualMachineClusterFlavor`](http://kubevirt.io/api-reference/v0.52.0/definitions.html#_v1alpha1_virtualmachineclusterflavor) and a namespaced [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/v0.52.0/definitions.html#_v1alpha1_virtualmachineflavor) CRDs. Each containing an array of [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/v0.53.0/definitions.html#_v1alpha1_virtualmachineflavorprofile) that at present only encapsulates CPU resources by applying a full copy of the [`CPU`](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpu) type to the [`VirtualMachineInstance`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachineinstance) at runtime.

This approach has a few pitfalls such as:

* Using embedded profiles within the CRDs

* Tight coupling between the resource definition of a `VirtualMachineFlavorProfile` and workload preferences it might also contain that are required to run a specific [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine)

* This results in a reliance on the user selecting the correct [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/v0.52.0/definitions.html#_v1alpha1_virtualmachineflavorprofile) for their specific workload

* Not allowing a user to override some viable `VirtualMachineInstanceSpec` attributes at runtime

## User Stories

* As an operator, user or vendor I would like to provide instancetypes that define the resources available to a [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) along with any performance related attributes.

* As an operator, user or vendor I would like to separately provide preferences that relate to ensuring that a specific guest OS within the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) is able to run correctly.

* As an user I would like to be presented with and select from pre-defined options that encapsulate the resources, performance and any additional preferences required to run my workload.

## Goals

* Simplify the choices a user has to make in order to get a runnable [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) with the required amount of resources, desired performance and additional preferences to correctly run a given workload.

## Non-Goals

The following items have been discussed alongside instancetypes and preferences but are not within the current scope of this design proposal:

* Introspection of imported images to determine the correct guest OS related `VirtualMachinePreferences` to apply.

* Using image labels to determine the correct guest OS related `VirtualMachinePreferences` to apply.

## Repos

* [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design

As stated above the ultimate goal of this work is to provide the end user with a simple set of choices when defining a [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) within KubeVirt. To achieve this the existing [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/v0.52.0/definitions.html#_v1alpha1_virtualmachineflavor) CRDs will be renamed, refactored and extended to better encapsulate resource, performance or schedulable attributes of a [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine).

The rename of the [`VirtualMachineFlavor`](http://kubevirt.io/api-reference/v0.52.0/definitions.html#_v1alpha1_virtualmachineflavor) CRD to [`VirtualMachineInstancetype`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetype) and overall API to from `flavor` to `instancetype` will hopefully present a more familiar UX to end users accustom to the approaches of many public cloud providers.

The embedded [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/v0.51.0/definitions.html#_v1alpha1_virtualmachineflavorprofile) type within the CRDs will also be removed and replaced with a singular [`VirtualMachineInstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec) type per instancetype. The decision to remove [`VirtualMachineFlavorProfile`](http://kubevirt.io/api-reference/v0.51.0/definitions.html#_v1alpha1_virtualmachineflavorprofile) has been made as the concept isn't prevalent within the wider Kubernetes ecosystem and could be confusing to end users. Instead users looking to avoid duplication when defining instancetypes will be directed to use tools such as [`kustomize`](https://kustomize.io/) to generate their instancetypes. This tooling is already commonly used when defining resources within Kubernetes and should afford users plenty of flexibility when defining their instancetypes either statically or as part of a larger GitOps based workflow.

[`VirtualMachineinstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec) will also include elements of `CPU`, `Devices`, `HostDevices`, `GPUs`, `Memory` and `LaunchSecurity` defined fully below. Users will be unable to override any aspect of the instancetype (for example, `vCPU` count or amount of `Memory`) within the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) itself, any attempt to do so resulting in the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) being rejected.

A new set of [`VirtualMachinePreference`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreference) CRDs will then be introduced to define any remaining attributes related to ensuring the selected guestOS can run. As the name suggests the [`VirtualMachinePreference`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreference) CRDs will only define preferences, so unlike a instancetype if a preference conflicts with something user defined within the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) it will be ignored. For example, if a user selects a [`VirtualMachinePreference`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreference) that requests a `preferredDiskBus` of `virtio` but then sets a disk bus of `SATA` for one or more disk devices within the [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) the supplied `preferredDiskBus` preference will not be applied to these disks. Any remaining disks that do not have a disk bus defined will however use the `preferredDiskBus` preference of `virtio`.

Both the [`VirtualMachineInstancetype`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetype) and [`VirtualMachinePreference`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreference) CRDs will also be versioned during [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) creation using `ControllerRevisions` that will store a complete copy of each object referenced by a given [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine). This will ensure that any modifications to or even removals of the original CRDs will not impact the created [`VirtualMachine`](https://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine) at runtime.

## [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) mappings to [`VirtualMachinePreferenceSpec`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreferencespec) and [`VirtualMachineInstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec)
The following basic rules are applied when deciding which attributes of the [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) will be mapped to the new [`VirtualMachinePreferenceSpec`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreferencespec) and [`VirtualMachineInstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec) types:

* Resource, schedulable and performance related attributes of the [`DomainSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec) reside within [`VirtualMachineInstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec)

* Remaining guest OS related preferences reside within [`VirtualMachinePreferenceSpec`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreferencespec)

For example, the number of a vCPUs presented to the guest are defined within [`VirtualMachineInstancetypeSpec`](http://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachineinstancetypespec) but the preferred topology of these vCPUs resides within [`VirtualMachinePreferenceSpec`](https://kubevirt.io/api-reference/main/definitions.html#_v1alpha1_virtualmachinepreferencespec).

* [DomainSpec](http://kubevirt.io/api-reference/main/definitions.html#_v1_domainspec)
  * [CPU](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpu)
    * cores 
      * `VirtualMachineinstancetypeSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets` , `preferThreads` or `preferCores` (default).
    * sockets
      * `VirtualMachineinstancetypeSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets`, `preferThreads`  or `preferCores` (default).
    * threads
      * `VirtualMachineinstancetypeSpec.CPU.Guest` 
        * Number of vCPUs to expose to the guest
      * `VirtualMachinePreferenceSpec.CPU.PreferredTopology`
        * Guest visible CPU topology to expose, `preferSockets`, `preferThreads`  or `preferCores` (default).
    * dedicatedCpuPlacement
      * `VirtualMachineinstancetypeSpec.CPU.DedicatedCpuPlacement`
    * [features](http://kubevirt.io/api-reference/main/definitions.html#_v1_cpufeature)
      * `VirtualMachineinstancetypeSpec.CPU.Features`
    * isolateEmulatorThread
      * `VirtualMachineinstancetypeSpec.CPU.IsolateEmulatorThread`
    * model
      * `VirtualMachineinstancetypeSpec.CPU.Model`
    * [NUMA](http://kubevirt.io/api-reference/main/definitions.html#_v1_watchdog)
      * `VirtualMachineinstancetypeSpec.CPU.NUMA`
    * [Realtime](http://kubevirt.io/api-reference/main/definitions.html#_v1_realtime)
      * `VirtualMachineinstancetypeSpec.CPU.Realtime`
  * [Devices](http://kubevirt.io/api-reference/main/definitions.html#_v1_devices)
    * autoattachGraphicsDevice
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachGraphicsDevice`
    * autoattachMemBalloon
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachMemBalloon`
    * autoattachPodInterface
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachPodInterface`
    * autoattachSerialConsole
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachSerialConsole`
    * autoattachInputDevice
      * `VirtualMachinePreferenceSpec.Devices.PreferredAutoattachInputDevice`
    * blockMultiQueue
      * `VirtualMachineinstancetypeSpec.Devices.BlockMultiQueue`
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
        * `VirtualMachineinstancetypeSpec.Devices.PreferredDedicatedIOThread`
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
          * `VirtualMachineinstancetypeSpec.GPUs`
        * name
          * None
        * tag
          * None
        * virtualGPUOptions
          * `VirtualMachinePreferenceSpec.Devices.PreferredVirtualGPUOptions`
      * [HostDevices](http://kubevirt.io/api-reference/main/definitions.html#_v1_hostdevice)
        * deviceName
          * `VirtualMachineinstancetypeSpec.HostDevices`
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
    * `VirtualMachineinstancetypeSpec.ioThreadsPolicy`
  * [launchSecurity](http://kubevirt.io/api-reference/main/definitions.html#_v1_launchsecurity)
    * [sev](http://kubevirt.io/api-reference/main/definitions.html#_v1_sev)
      * `VirtualMachineinstancetypeSpec.LaunchSecurity.SEV`
  * [Machine](http://kubevirt.io/api-reference/main/definitions.html#_v1_machine)
    * type
      * `VirtualMachinePreferenceSpec.Machine.PreferredMachineType`
  * [Memory](http://kubevirt.io/api-reference/main/definitions.html#_v1_memory)
    * [guest](http://kubevirt.io/api-reference/main/definitions.html#_k8s_io_apimachinery_pkg_api_resource_quantity) 
      * `VirtualMachineinstancetypeSpec.Memory.Guest`
    * [hugepages](http://kubevirt.io/api-reference/main/definitions.html#_v1_hugepages)
      * `VirtualMachineinstancetypeSpec.Memory.HugePages`

# Acknowledgements

* David Vossel - Original downstream design proposal
* Fabian Deutsch - Counter downstream design proposal introducing domain preferences
* Andrej Krejcir - Initial implementation of flavors within KubeVirt
