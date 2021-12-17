# Overview

[Secure Encrypted Virtualization (SEV)](https://developer.amd.com/sev/) is a feature of AMD's EPYC CPUs that allows the memory of a virtual machine to be encrypted on the fly. This document describes a design proposal for adopting the technology in KubeVirt.

This is the first step towards enabling confidential computing in KubeVirt. The proposal focuses on the basic AMD SEV. Later it can be extended to cover the extensions: `ES` (Encrypted State) and `SNP` (Secure Nested Paging) as well as other technologies like `Intel TDX`.

## Motivation

Without encryption any information stored in RAM can be compromised. This problem is especially relevant for the cloud environments where users run their workloads. Malicious software deployed on the patform or the administrator can get access to a running VM and that may lead to data leakage.

SEV attempts to address that and to reduce the risk of leaking sensitive information. It transparently encrypts the memory with a key that is unique for each VM. Additionally it also provides a measurement of the memory content that can be used by VM owners to perform the attestation thus showing that the memory was encrypted correctly.

## Goals

- Provide a way for KubeVirt users to run VMs with SEV
- Provide an interface to allow KubeVirt users to perform the attestation

## Non Goals

- Initialization of the SEV platform, keys provisioning and management

    Potentially there can be other 'SEV users' on the platform apart from KubeVirt. Therefore initialization, keys provisioning and management is not included in the scope of this document at this step.

- Automation of the attestation process

    This directly affects user expeience but from the guest owner perspective KubeVirt is a part of the platform and therefore cannot be fully trusted. To facilitate the attestation process a guest owner can use external tools like [sev-tool](https://github.com/AMDESE/sev-tool) or [sevctl](https://github.com/enarx/sevctl).

## Definition of Users

The feature is intended for the users who run VMs with sensitive data and who do not want to put additional 'trust' in the cloud platform.

## User Stories

- As a cluster administrator, in order to provide better confidentiality of the users' data, I want to allow running the VMs with memory encryption enabled.
- As a cluster user, in order to not have to share my secrets with the administrator, I want to be able to run my VMs with AMD SEV functionality enabled.
- As a cluster user, in order to prevent leaking of my data in case when the hypervisor software is compromised, I want to run my VMs with AMD SEV functionality enabled.

## Repos

- [KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

There are two major parts that need to be implemented. First the domain xml parameters need to be properly configured in order to match the launch requirements and run a VM with SEV. Then the second step implies providing the interfaces to allow a user to perform attestation of a running VM.

## Launching SEV guests

SEV parameters need to be specified via the `launchSecurity` element of the libvirt domain xml:

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

Here the `policy` parameter allows setting the guest policy flags. The optional `dhCert` and `session` provide the guest owner's base64 encoded `DH` (Diffie-Hellman) key and the guest owner's base64 encoded `launch blob` respectively (those are needed for the attestation).

Additionally to start a SEV VM there is a need to provide `cbitpos` and `reducedPhysBits`. The parameters are hypervisor-dependent and can be obtained from the `<sev></sev>` element of the `virsh domcapabilities`:

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

**Note:** in recent libvirt `cbitpos` and `reducedPhysBits` can be ommited.

Apart from setting the parameters, there are several more prerequisites to run SEV guests:

- all virtio devices need to be configured with the `iommu='on'` attribute in their `<driver>` configuration
- all memory regions used by the VM must be `locked` for Direct Memory Access (DMA) and to prevent swapping

From the implementation perspective the required parameters can be mapped to the VMI spec:

```yaml
spec:
  domain:
    launchSecurity:
      sev: {}
```

Additional logic and validation shall be added to ensure that the domain xml meets the prerequisites. Apart from that `/dev/sev` device needs to be exposed to the VM pod with proper access bits (currently QEMU requires `read-write` access).

SEV-capable nodes should be properly labeled. Then a corresponding node selector can be added to VM pods so they are scheduled correctly.

## Attestation

[Attestation](https://github.com/AMDESE/sev-tool#proposed-provisioning-steps) is an interactive process between `platform owner` and `guest owner`. It is carried out to establish a certain level of trust. To leverage the benefits of memory encryption a guest owner provides an encrypted VM disk. To decrypt the data and to run a VM with such a disk there is a need to `inject a secret` into the VM when the system is about to boot. The secret is known only by the guest owner. The goal of the attestation process is to validate that the platform is genuine and it really runs a non-compromised SEV hardware. Also it is needed to establish a private communication channel with the platform so the guest owner can securely provide the secret. Attestation consists of the following steps:

- VM gets scheduled on a SEV-capable node
- guest owner validates the platform on that specific node
- guest owner establishes a secure channel with the platform by providing the session parameters
- VM is started in a paused state
- guest owner requests and verifies the measurement
- guest owner signals the VM to unpause by injecting the secret

More details about each individual step are provided bellow.

### Step #1 Validation of the platform

The guest owner should get the assets listed bellow from the platform owner:

- the `PDH` (Platform Diffie-Hellman) key exported from the SEV hardware
- the complete key chain (up to the `root of trust`) used to sign the `PDH`
- the OVMF binary

KubeVirt can expose a VMI subresource endpoint `/sev/fetchcertchain` which will return base64 encoded certificates. Also a new command can be added to virtctl tool to facilitate the process: `virtctl sev fetchcertchain ...`.

On a node where KubeVirt is running, the privileged `virt-handler` pod can export the required assets in advance to a `kubevirt-private` directory. Whenever a new VM requiring SEV is launched, the directory can be mounted to the `virt-launcher` pod in `read-only` mode. A user can then invoke the `virtctl` tool to fetch the certificate chain for validation.

In order to ensure that the platform is not compromised the guest owner needs to validate the `PDH` key using the provided chain. For that `sev-tool --validate_cert_chain ...` can be used (AMD root certificate should be downloaded from the web for verification). Additionally the UEFI image should be considered 'trustable' by the guest owner.

### Step #2 Establishing secure channel with the platform

To establish a secure communication channel with the platform a guest owner needs to provide the launch blob with the parameters of the session and the certificate derived from `PDH`. This can be suported in KubeVirt by `/sev/setupsession` VMI endpoint with the parameters `session` and `dhCert` (which will be mapped to the corresponding elements in the domain xml). New `virtctl sev setupsession ...` command can be introduced to trigger propagation of the parameters.

The implementation in `virt-launcher` in turn can 'postpone' the creation of the domain until the parameters are specified.

The construction of the launch blob is currently not in the scope of KubeVirt as it requires the guest owner to expose the secrets. It should be performed 'offline' by the guest owner (e.g. by using `sev-tool --generate_launch_blob ...`).

### Step #3 Retrieving the measurement of a running VM

The measurement of the memory is performed before a VM is booted. Initially a VM is launched in a paused state thus allowing to retrieve and validate the measurement. KubeVirt can provide the `/sev/querylaunchmeasure` VMI endpoint complemented by `virtctl sev querylaunchmeasure ...` command to expose it to the guest owner.

The implementation will need to call the corresponding libvirt API to retrieve the measurement for a particular VM:

- [virDomainGetLaunchSecurityInfo](https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainGetLaunchSecurityInfo)

    The call returns the launch measurement for a SEV guest.

Currently the measurement is calculated basing on the negotiated session parameters and SHA256 digest of the OVMF binary. The guest owner can verify it against the binary provided by the platform owner.

### Step #4 Launching the VM by injecting the secret

After the measurement is successfully verified, the guest owner can inject the encrypted secret into the VM and resume its execution. A new `/sev/injectsecret` VMI endpoint and `virtctl sev injectsecret ...` command can be added to support secret injection in KubeVirt VMs.

The implementation will need to call the corresponding libvirt API once it becomes available:

- [sev-inject-launch-secret](https://listman.redhat.com/archives/libvir-list/2021-May/msg00196.html)

    The libvirt API is still missing.

The blob with the encrypted secret can be built by the guest owner using `sev-tool --package_secret ...`.

## Known limitations

- SEV is only going to work with OVMF (UEFI)
- SEV-encrypted VMs cannot contain directly-accessible host devices (that is, PCI passthrough)
- Live Migration is not supported with SEV at the moment
- The ammount of running SEV VMs on a node is limited
- libvirt currently does not provide an API to inject a secret in a running VM

Some of the limitations will probably be removed in the future.

## API Examples

**Requesting SEV in the VMI spec**

The bellow yaml snippets provide examples of how to request SEV feature in the VMI spec.

*SEV VM with default policy (no attestation):*

```yaml
spec:
  domain:
    launchSecurity:
      sev: {}
```

*SEV VM with user-defined policy (no attestation):*

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

**VMI subresource endpoints**

The following VMI subresource endpoints can be introduced for the attestation process:

- *Fetching the certificate chain*: `/sev/fetchcertchain`
    ```
    virtctl sev fetchcertchain VMI
    ```
- *Fetching OVMF binary info*: TBD (???)

- *Providing session parameters (launch blob and DH cert)*: `/sev/setupsession`
    ```
    virtctl sev setupsession VMI --session <base64> --dhcert <base64>
    ```
- *Querying launch measure*: `/sev/querylaunchmeasure`
    ```
    virtctl sev querylaunchmeasure VMI
    ```
- *Injecting launch secret*: `/sev/injectsecret`
    ```
    virtctl sev injectsecret VMI --secret <base64>
    ```

## Scalability

The attestation process needs to be automated and offloaded to an external (trusted) service in order to make the solution scalable. This is currently not in the scope since the proposal is focused on the initial integration and basic support of the SEV technology. It can be extended at a leter point though.

## Update/Rollback Compatibility

The feature should not impact update compatibility.

## Functional Testing Approach

Functional tests should be provided to cover the primary use-case scenarios:

- Launching of a SEV VM
- Attestation

**Note**: SEV-capable hardware is required to run the end-to-end tests.

# Implementation Phases

- Support running KubeVirt VMs with basic SEV (without attestation)
- Expose the new VMI subresource endpoints to allow attestation
- Implement virtctl commands for interactive attestation
- Extend the implementation to support new features/technologies (ES, SNP, TDX, etc.) as soon as libvirt and qemu adopt those
