# Overview

This proposal deals with the enablement of the following technologies: Intel
TDX & AMD Secure Encrypted Virtualization with Secure Nested Paging (SEV-SNP)
in kubevirt. This enablement includes the support for creating confidential VMs
by using TDX & SEV-SNP by encrypting VM memory. For Intel TDX, This proposal
relies on the attestation by using trustee client and the trustee operator.
For AMD, SEV is currently supported in KubeVirt, and this VEP proposes to build
on top of this feature and integrates SEV-SNP support.

## Motivation

Organizations are now asking for workloads to include the strongest
confidential computing security guarantees.

For AMD, KubeVirt supports SEV and SEV-ES for memory encryption, but lacks
SEV-SNP enhanced memory integrity protection.

For Intel, Add support for TDX technology for in-use memory encryption of
confidential guests.

## Proposal

- Enable TDX VM Deployment: Provide users the ability to deploy VMs that are
  utilizing INTEL TDX technology.
- Enable TDX VM Blob Generation: Provide users the ability to generate the blob
  by using the INTEL TDX technology, in particular, by relying on the Quote
  Generation Service (QGS) component.
- Enable SEV-SNP VM Deployment: Provide users the ability to deploy VMs that
  are utilizing AMD SEV-SNP technology.
- Extend the existing SEV Feature: Currently the SEV sits behind the
  WorkloadEncryptionSEV adding SEV-SNP to feature is the next logical step.
- Extend Existing SEV Infrastructure: Use the existing SEV node labeling,
  scheduling, and domain generation infrastructure.

## Non Goals

- TDX & SEV-SNP Live Migration: Live migrations are currently not supported by
  either technologies.
- Custom Attestation Services: Only focus on the enabling the capabilities
  through the standard interfaces and not building a custom attestation
  infrastructure
- SEV migration: No automatic upgrade path from SEV/SEV-ES to SEV-SNP, this
  will be required to be done by explicit user configuration.

## Definition of Users
- Cluster Admins: Responsible for enabling and managing SEV-SNP & TDX
  capabilities in Kubernetes Clusters.
- Developers: Deploys confidential workloads that require SEV-SNP & TDX.

## User Stories

- As a Cluster Admin, I want to enable TDX on cluster nodes so that tenants can
  deploy Confidential Compute workloads with TDX enabled.
- As a Cluster Admin, I need to ensure that my clusters have the correct
  labeling that shows the correct TDX labels are showing up on my nodes.
- As a Developer, I want to deploy VMs on a platform with TDX protection to
  meet compliance requirements for memory integrity.
- As a Developer, I want to deploy VMs without needing to understand the low
  level INTEL technology.
- As a Developer, I want to generate the blob for VMs on a platform with TDX
  protection technology.
- As a Cluster Admin, I want to enable SEV-SNP on cluster nodes so that tenants
  can deploy Confidential Compute workloads with SEV-SNP enabled.
- As a Cluster Admin, I need to ensure that my clusters have the correct
  labeling that shows the correct SEV and SEV-SNP labels are showing up on my
  nodes.
- As a Developer, I want to deploy VMs on a platform with SEV-SNP protection to
  meet compliance requirements for memory integrity.
- As a Developer, I want to deploy VMs without needing to understand the low
  level AMD technology.

## Repos

- [KubeVirt](https://github.com/kubevirt/kubevirt)

## Design

### TDX Design
A new TDX Feature Gate should be added first. Extend the node labeller to
detect and label nodes capable of running TDX VMs, a corresponding node
selector can be added to VM pods so they are scheduled correctly. TDX
parameters need to be specified via the `launchSecurity` element of the Libvirt
domain xml:
```xml
   <domain>
     ...
     <launchSecurity type='tdx'>
       <policy>0x10000001</policy>
       <mrConfigId>xxx</mrConfigId>
       <mrOwner>xxx</mrOwner>
       <mrOwnerConfig>xxx</mrOwnerConfig>
       <quoteGenerationSocket path="/var/run/tdx-qgs/qgs.socket"/>
     </launchSecurity>
     ...
   </domain>
```
All parameters are optional. Note the Libvirt support is not ready in upstream,
refer to the
https://github.com/intel/libvirt-tdx/commit/962325568e67014c2ff2e41a673374fafd44e385
for the explanation of the parameters. From the implementation perspective,
extend the VMI spec to map to the parameters:

```xml
   spec:
     domain:
       launchSecurity:
         tdx:
           policy: 0x10000000
           mrConfigId: xxx
           mrOwner: xxx
           mrOwnerConfig: xxx
           qgs: true
```

For the Quote Generation Service(QGS), Libvirt currently only provides unix
socket support, and it will use `/var/run/tdx-qgs/qgs.socket` when the path is
not specified. As there is only one QGS service running on the host, and QGS is
started with the default unix socket, just let Libvirt configures QGS in the
default way. All parameters are optional to configure a TDX VM, please refer to
API Examples.

When the corresponding label applies, i.e., TDXLabel, the QGS service is
launched as a container in the virt-handler pod. We decide to use one instance
of QGS per node. The decision is based on leveraging QGS cache and the resource
consumption that could represent having one instance per VM.

### AMD Design

AMD SEV-SNP maps to a new type in the LibVirt XML of type “sev-snp” with its
own attributes and elements that are similar to the type “sev”.

```xml
   <domain>
     ...
     <launchSecurity type='sev-snp'>
       <cbitpos>47</cbitpos>
<policy>0x00030000</policy>
       <reducedPhysBits>1</reducedPhysBits>
	<guestVisibileWorkarounds>...</guestVisibleWorkaround>
	<idBlock>...</idBlock>
	<idAuth>...</idAuth>
	<hostData>...</hostData>
     </launchSecurity>
     ...
   </domain>
```

A new structure should be created for adding SEV-SNP on the `LaunchSecurity`
field.
```go
type SEVSNP struct {
    // 64-bit SEV-SNP Guest Policy
    // +optional
    Policy *string `json:"policy,omitempty"`
    // 16-byte base64 encoded guest hypervisor-defined workarounds.
    // +optional
    GuestVisibleWorkarounds *string `json:"guestVisibleWorkarounds,omitempty"`
    // 96-byte base64 encoded ID Block Structure.
    // +optional
    IDBlock *string `json:"idBlock,omitempty"`
    // 4096-byte base64 encoded ID Auth Structure.
    // +optional
    IDAuth *string `json:"idAuth,omitempty"`
    // 32-byte base64 encoded user-defined blob to provide to the guest.
    // +optional
    HostData *string `json:"hostData,omitempty"`
    // Whether the guest is allowed to use VCEK for attestation reports. Set to false to disable VCEK usage.
    // +optional
    AuthorKey *bool `json:"authorKey,omitempty"`
    // Whether idAuth contains VCEK field for attestation
    // +optional
    VCEK *bool `json:"vcek,omitempty"`
    // Optional attribute to indicate whether the hashes of the kernel, and command line should be included in the measurement done by the firmware.
    KernelHashes *bool `json:"kernelHashes,omitempty"`
}
```
Similar to TDX, all elements for type=’sev-snp’, users should be able to
provide additional settings to configure their Confidential VM. Some options
should have a default value such as the policy if the user does not specify the
VM should be using the QEMU default policy.  The node labller will detect
SEV-SNP capabilities from the LibVirt domain capabilities then apply the label
to the node, while the node selector renderer will be extended to include
SEV-SNP scheduling.

## API Examples
The bellow yaml snippets provide examples of how to request TDX feature in the
VMI spec.

### API Examples for Intel TDX

- TDX VM with defaults settings

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  ...
spec:
  domain:
    launchSecurity:
      tdx: {}
  ...
```

- TDX VM with user-defined parameters (no qgs)

```yaml
   spec:
     domain:
       launchSecurity:
         tdx:
           policy: 0x10000000
           mrConfigId: xxx
```

- TDX VM with qgs settings, note libvirt will use `/var/run/tdx-qgs/qgs.socket`
  when path is not specified

```yaml
   spec:
     domain:
       launchSecurity:
         tdx:
           policy: 0x10000000
           mrConfigId: xxx
           qgs: true
```

### API Examples for AMD

- Basic SEV-SNP with default options

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: sev-snp-vm
spec:
  domain:
    launchSecurity:
      snp: {}
    resources:
      requests:
        memory: 4Gi
    devices:
      disks:
      - name: containerdisk
        disk:
          bus: virtio
  ...
```

- SEV-SNP enablement with simple policy setting

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: sev-snp-vm
spec:
  domain:
    launchSecurity:
      snp:
        policy: "0x00030000"
    resources:
      requests:
        memory: 4Gi
    devices:
      disks:
      - name: containerdisk
        disk:
          bus: virtio
```

- The following is an example of a full implementation of all available options for SEV-SNP.

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: advanced-sev-snp-vm
spec:
  domain:
    launchSecurity:
      snp:
        policy: "0x30000"
        authorKey: true
        vcek: true
        guestVisibleWorkarounds: "YWJjZGVmZ2hpams="  # base64 encoded
        idBlock: "bG9yZW0gaXBzdW0gZG9sb3Igc2l0IGFtZXQ="    # base64 encoded 96-byte block
        idAuth: "Y29uc2VjdGV0dXIgYWRpcGlzY2luZyBlbGl0..."   # base64 encoded 4096-byte block
        hostData: "c2VkIGRvIGVpdXNtb2QgdGVtcG9yIGluY2k="    # base64 encoded 32-byte data
    resources:
      requests:
        memory: 4Gi
    devices:
      disks:
      - name: containerdisk
        disk:
          bus: virtio
  volumes:
  - name: containerdisk
    containerDisk:
      image: quay.io/kubevirt/fedora-cloud-container-disk-demo
```

## Scalability

## Security

## Update/Rollback Compatibility

- All new fields are fields should be optional
- This should not impact existing VMs
- AMD SEV Compatibility:
  - The AMD SEV-SNP feature sits behind the existing SEV feature gate without
    breaking changes
  - Node labels are added, no labels are removed.

## Functional Testing Approach

- Unit testing to detect TDX & SEV-SNP from the libvirt capabilities.
- Since TDX & SEV-SNP does not have support for nested virtualization this will
  require bare metal hardware to conduct e2e testing.


## Implementation Phases

### Intel TDX Phases:
This feature is split into two items. The first item corresponds to the
enablement of TDX in kubevirt thus allowing the creation and deployment of
confidential VMs using the TDX technology. The second item corresponds to the
generation of blobs that require the deployment of QGS and MPA registering
services in the node.

### AMD SEV-SNP Phases:
The initial phase of implementation will focus on integrating basic
functionality for all elements, allowing users to input configurations without
immediate validation. The subsequent phase will involve adding necessary checks
to prevent the creation of improperly configured VMs (e.g., preventing users
from setting KernelHashes without configuring kernel booting).
