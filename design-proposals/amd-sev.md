# Overview

[Secure Encrypted Virtualization (SEV)](https://developer.amd.com/sev/) is a feature of AMD's EPYC CPUs that allows the memory of a virtual machine to be encrypted on the fly. This document describes a design proposal for adopting the technology in KubeVirt.

This is the first step towards enabling confidential computing in KubeVirt. The proposal focuses on the basic AMD SEV. Later it can be extended to cover the extensions: `ES` (Encryptes State) and `SNP` (Secure Nested Paging) as well as other technologies like `Intel TDX`.

## Motivation

Without encryption any information stored in RAM can be compromised. This problem is especially relevant for the cloud environments where users run their workloads. Malicious software deployed on the patform or the administrator can get access to a running VM and that may lead to data leakage.

SEV attempts to address that and to reduce the risk of leaking sensitive information. It transparently encrypts the memory with a key that is unique for each VM. Additionally it also provides a measurement of the memory content that can be used by the owner of the VM to perform the attestation thus showing that the memory was encrypted correctly.

## Goals

- Provide a way for KubeVirt users to run a VM with SEV
- Provide an interface to allow KubeVirt users to perform the attestation

## Non Goals

- Initialization of the SEV platform, keys provisioning and management

    Potentially there can be other 'SEV users' on the platform apart from KubeVirt. Therefore initialization, keys provisioning and management is not included in the scope of this document at this step.

- Automation of the attestation process

    This directly affects user expeience but from the guest owner perspective KubeVirt is a part of the platform and therefore cannot be fully trusted.

## Definition of Users

The feature is intended for the users who operate VMs with sensitive data and who do not want to put additional 'trust' in the cloud platform.

## User Stories

- As a cluster administrator, in order to provide better confidentiality of the users' data, I want to allow running the VMs with memory encryption enabled.
- As a cluster user, in order to not have to share my secrets with the administrator, I want to be able to run my VMs with AMD SEV functionality enabled.
- As a cluster user, in order to prevent leaking of my data in case when the hypervisor software is compromised, I want to run my VMs with AMD SEV functionality enabled.

## Repos

- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

There are two major parts that need to be implemented. First the domain xml parameters need to be properly configured in order to match the launch requirements and run a VM with SEV. Then the second step implies providing the interfaces to allow a user to perform attestation of a running VM.

## Launching SEV guests

SEV parameters can be specified via the `launchSecurity` element of the domain xml:

```xml
<domain>
  ...
  <launchSecurity type='sev'>
    <policy>0x0001</policy>
    <cbitpos>47</cbitpos>
    <reducedPhysBits>1</reducedPhysBits>
    <dhCert>RBBBSDDD=FDDCCCDDDG</dhCert>
    <session>AAACCCDD=FFFCCCDSDS</session>
  </launchSecurity>
  ...
</domain>
```

To start a SEV VM there is a need to specify `cbitpos` and `reducedPhysBits`. Those parameters are hypervisor-dependent and can be obtained from the `<sev></sev>` element of the `virsh domcapabilities`:

```xml
<domainCapabilities>
...
  <features>
    ...
    <sev supported='yes'>
      <cbitpos>47</cbitpos>
      <reducedPhysBits>1</reducedPhysBits>
    </sev>
    ...
  </features>
</domainCapabilities>
```

The `policy` parameter allows setting the guest policy flags. Additionally the optional `dhCert` and `session` provide the guest owner's base64 encoded `DH` (Diffie-Hellman) key and the guest owner's base64 encoded `launch blob` respectively.

There are also several prerequisites to run SEV guests:

- All virtio devices need to be configured with the `iommu='on'` attribute in their `<driver>` configuration
- All memory regions used by the VM must be `locked` for Direct Memory Access (DMA) and to prevent swapping

## [Attestation](https://github.com/AMDESE/sev-tool#proposed-provisioning-steps)

### Certificate chain validation

In order to ensure that the platform is not compromised the guest owner first needs to validate the certificate chain provided by the platform owner. On a system where KubeVirt is running, the privileged `virt-handler` pod can export the required certificates in advance to a `kbevirt-private` directory. Whenever a new VM requiring SEV is launched, the directory can be mounted to the `virt-launcher` pod in `read-only` mode. A user can then invoke the `virtctl` tool to download the certificate chain for validation (for that an additional command will be needed, e.g. `virtctl fetchcertchain ...`).

### Validating the measurement

The attestation for the basic SEV is performed before a VM is booted. It requers a user to provide the launch blob and the certificate (via `session` and `dhCert` paramters in the libvirt domain xml). A VM is launched in paused state thus allowing to retrieve and validate the measurement. Once it is verified, user can inject the secret into the VM and resume the execution.

To support the attestation in KubeVirt `virtctl` can be extended to accept additional commands: `launchmeasure` and `injectsecret`. The commands will eventually end up with calling the corresponding libvirt API:

- [virDomainGetLaunchSecurityInfo](https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainGetLaunchSecurityInfo)

    The call returns the launch measurement for a SEV guest.

- [sev-inject-launch-secret](https://listman.redhat.com/archives/libvir-list/2021-May/msg00196.html)

    The API is still missing.

## Known limitations

- SEV is only going to work with OVMF (UEFI)
- SEV-encrypted VMs cannot contain directly-accessible host devices (that is, PCI passthrough)
- Live Migration is not supported with SEV at the moment
- The ammount of running SEV VMs on a node is limited
- libvirt currently does not provide an API to inject a secret in a running VM

Some of the limitations will probably be removed in the future.

## API Examples

**Requsting SEV feature in the VMI spec**

```yaml
spec:
  domain:
    launchSecurity:
      sev:
        policy:
          - EncryptedState
```

or alternatively

```yaml
spec:
  domain:
    launchSecurity:
      sev:
        policy:
          encryptedState: true
```

**Fetching the certificate chain**

```
virtctl fetchcertchain VMI
```

**Querying launch measure**

```
virtctl launchmeasure VMI
```

**Injecting launch secret**

```
virtctl injectsecret VMI --secret <base64>
```

## Scalability

 N/A

## Update/Rollback Compatibility

The feature should not impact update compatibility.

## Functional Testing Approach

Functional tests should be provided to cover the primary use-case scenarios:

- Launching of a SEV VM
- Attestation

**Note**: SEV-capable hardware is required to run the end-to-end tests.

# Implementation Phases

- Support running KubeVirt VMs with basic SEV
- Once supported by libvirt, implement the interfaces to allow attestation
- Extend the implementation to support new features/technologies (ES, SNP, TDX, etc.) as soon as libvirt and qemu adopt those
