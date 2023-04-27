# Network Interface Resource Request

Author: [Edward Haas](edwardh@redhat.com)

Contributors: [Andrea Bolognani](abologna@redhat.com)

Status: Ready

## Summary
Support the ability to define how many network devices a VM may potentially contain.
Allowing users to hotplug network interfaces at least to that desired capacity.

## Motivation
The introduction of network devices hotplug introduced a limitation
to the amount of possible devices one can attach.

Users of libvirt who can estimate how many potential network devices are needed,
can define the additional PCI/e controllers in advance.
KubeVirt does not expose such domain details through its API.
But even if these details had been exposed by the VMI object,
a common user would find it challenging to define such low domain details.

### Goals
- Allow users to define a guaranteed amount of network devices that can be potentially hotplug.
- Do not expose domain architecture details (e.g. PCI/PCIe controllers).

### Non-Goals
- Define an upper limit to the number of network devices that can be defined. 

## Proposal
Users are to express the potential amount of network devices through
the VMI spec.
Internally, kubevirt will use libvirt logic to generate the needed
devices to support the amount specified (e.g. PCIe controllers).

Using hotplug, users may add network devices at least to the amount specified
by the new user *domain-resources-requests-kubevirt.io/interface* input.

> **Note**: While this solution focuses on network interfaces, the approach can be
> applied to any device type which has dependencies on other devices.

### Definition of Personas
- VM/VMI administrator: Any user who can create VM and VMI objects.

### User Stories
- As a VM/VMI administrator, I would like to create a VM that is able
  to increase its network interfaces count in the future.
  I would explicitly define the amount of potential network interfaces
  by providing a value through the VM/VMI Spec.

### API Extensions
The VMI Spec will be extended to include a new domain resource name: `kubevirt.io/interface`.
It will be available as a domain-resources-requests entry.

#### Validation
- Should be immutable.
- Should accept an integer value between 0 and 32.
- Should be accepted only as `requests` (not as `limits`).
- In case the value is less than the amount of interfaces defined already,
  it has no effect.

> **Note**: These are API level validations. The backend (e.g. hypervisor)
> may have other validations/limitation.

#### VMI Spec example
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: vmi-iface-request-example
spec:
  domain:
    devices:
      disks:
        - disk:
            bus: virtio
          name: containerdisk
      interfaces:
        - masquerade: {}
          name: default
        - bridge: {}
          name: secondary
    resources:
      requests:
        kubevirt.io/interface: 5
    rng: {}
  networks:
    - name: default
      pod: {}
    - multus:
        networkName: some-nad
      name: secondary 
  volumes:
    - containerDisk:
        image: registry:5000/kubevirt/fedora-with-test-tooling-container-disk:devel
      name: containerdisk
```

## Design Details
The solution is taking advantage of the existing libvirt capability to auto-generate
the needed configuration (e.g. controllers) when specifying network interfaces.

The main change is in the logic of creating a domain in the virt-launcher.

This is the main flow suggested:
- Base on the resources-requests-kubevirt.io/interface and the current defined interfaces, generate
  [interface-placeholders](#network-interface-placeholder) and add them to the domain spec.
- Following the creation of the domain configuration, remove these placeholders
  from he configuration and start the domain without them.
- At this stage, the domain has been started without the additional interface devices
  but with the additional configuration (e.g. controllers) they originally generated.

#### Network Interface Placeholder
The interface placeholder is a temporary dummy definition of an interface in the domain spec
of libvirt. It is used to trigger libvirt to auto-generate dependent devices (e.g. PCIe controllers).

After the domain configuration is constructed, these placeholders are removed.

This is an example of such a placeholder:
```xml
<interface type='ethernet'>
  <target dev='placeholder-1' managed='no'/>
  <model type='virtio-non-transitional'/>
</interface>
```

## Appendix A: Alternatives
There have been several other implementation alternatives to answer the goals.
This section lists the alternatives to allow an open discussion and record other proposals.

- Allow users to define explicitly domain controllers.

  This option includes the full exposure of the controllers devices details or a partial
  abstraction of it.

  Such an approach would require deep knowledge of machine type-specific details
  by the user.

- While using the same suggested API towards the user, construct the required controllers
  directly (and not through network interface placeholders).

  Similar to the previous point, intimate knowledge of the machine type-specific details
  is required, this time by KubeVirt (and not the user).
  For example, `q35` requires `pcie-root-port` controllers to be added but `pc` doesn't,
  and would in fact break if those were present.

- Always define the maximum number of resources by default (or as a boolean input).

  Each additional PCI controller introduces some overhead in terms of memory usage and
  guest OS boot time, so minimizing their amount is beneficial.

> **Note**: That said, the exact overhead hasn't been measured, and it could turn out
  not to be a deal-breaker compared to the baseline resource usage.

  Considering hotplug will not be used on most deployments, it would not be ideal
  to pay the penalty by all users.
  That said, if this option is found reasonable in the future, it can be smoothly
  integrated into the proposed solution by defining a default (e.g. 32).
  In turn it will make the suggested input redundant (i.e. a no-op).

- Use a PCI bridge.

  This solution would require implementing PCI address allocation logic in KubeVirt,
  as libvirt will normally refuse to use the conventional PCI slots provided by
  a `pci-bridge` controller for devices on a PCI Express-based machine type such as `q35`.
  The logic would have to be used for all devices, not just network interfaces,
  in order to avoid clashes with the addresses allocated by libvirt.

## Appendix B: Extensions
The following extension options are presented here as optional to extend this proposal.

- Define resource limits for interfaces: Limit the amount of interfaces that can be defined.
- Define other resource types (e.g. disks) in the same manner.
