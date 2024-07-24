# Overview
This proposal introduces a new feature in KubeVirt to automatically
generate serial numbers for each disk attached to a virtual machine.

## Motivation
- Windows guest OS in the VM requires that disk serial numbers are defined and persistent across reboots.
- Generating the serial reduces manual configuration.

## Goals
- Add an option to the VM's `.spec` that will enable automatic serial number generation for disks that
  don't have a serial number specified.

- Generate unique serial numbers for each disk. The number has to be unique compared to all other disks attached
  to all VMs with this disk.
 
- Don't store the serial number in the VM's `.spec`. It would confuse gitops operators.

- The serial number should remain the same across storage migration.

## Non Goals

- No customization options for serial number format.

## Definition of Users

- VM owners

## User Stories

- As a VM owner, I want to use the automated disk serial number assignment to reduce manual configuration.
- As a VM owner, I want each disk to have a unique serial number.
- As a VM owner, I want newly created Windows VMs to have serial numbers set on disks by default.

## Repos
- [kubevirt](https://github.com/kubevirt/kubevirt)

# Design

A new boolean option `generateDiskSerials` will be added to the VMI `.spec`. 
When it is set to `true`, kubevirt will generate a serial number for each disk, that does not have one specified.
These serial numbers will then be set in the libvirt domain xml.

This option can also be added to `VirtualMachinePreference`, so it can be set on new VMs and
the user does not need to be aware of it.

## Persistence of generated serial numbers

The serial number has to be the same across VM restarts and storage migration,
so it has to be stored somewhere, or generated deterministically.

This section lists several options how to store serial numbers.
Then the rest of the document describes details of option: **Annotation on the PVC object**.

### Annotation on the PVC object
The serial number will be stored as an annotation on the PVC object. It would be created when the VMI is created.
Then it will be read by `virt-handler` when creating libvirt domain XML.

#### Advantages:
- Serial is associated with PVC.
- Serial is clearly visible to anyone who can read PVC.

#### Disadvantages:
- Does not store serials for other volume types except PVC.


### Storing it in the `status` of the VM
Serial numbers will be stored in the `.status` of the VM.
GitOps does not use the `/status` subresource to update the objects it creates, so it should not overwrite the serials.
More information about `.status` is [here](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status).

#### Advantages:
- We can store serial numbers for all volume types.
- Serial is clearly visible to anyone who can read VM.

#### Disadvantages:
- Harder to make sure that a disk has the same serial if it is mounted to multiple VMs.
- It may be an unexpected place to store persistent information that cannot be regenerated if `.status` is wiped.


### Deterministic generation
We can use name and namespace, or UID fields of the PVC to deterministically generate the serial number. 
For non-PVC volumes, we can use other metadata to generate a serial number.

#### Advantages:
- We don't need to store anything.

#### Disadvantages:
- If the storage is migrated to a PVC with different metadata, it would change the serial number.


### Storing the serial in a ConfigMap  
We can create a ConfigMap that would store all disk serials for a VM.

#### Advantages:
- We can store serial numbers for all volume types.

#### Disadvantages:
- Keeping the ConfigMap in sync with the VM is more complicated.
- Harder to make sure that a disk has the same serial if it is mounted to multiple VMs.


## API Examples

### Existing API to set serial number manually

The serial number can be specified manually in VM spec here:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example-vm
spec:
  template:
    spec:
      domain:
        devices:
          disks:
            - name: disk
              serial: 1234567890 # <-- serial number
              disk:
                bus: virtio
      volumes:
        - name: disk
          persistentVolumeClaim:
            claimName: example-pvc
```

### New API

To enable the serial generation, new filed `generateDiskSerials` will be added to the VMI `.spec`:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: example-vm
spec:
  template:
    spec:
      domain:
        devices:
          generateDiskSerials: true # <-- new field
          disks:
            - name: disk
              disk:
                bus: virtio
      volumes:
        - name: disk
          persistentVolumeClaim:
            claimName: example-pvc
```

The generated serial will be stored as an annotation in the PVC:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: example-pvc
  annotations:
    kubevirt.io/disk-serial: 123456789 # <-- new annotation
spec:
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 10Gi
```

## Scalability
Serials are generated during VMI creation, if the annotation does not exist.

## Update/Rollback Compatibility
The feature is backward compatible. Existing VMs remain unaffected unless the new field is set.
Rollback ignores the new field.

## Testing Approach

### Unit tests
We can add new unit tests to check if the generated libvirt XML contains the `<serial>` tag.

### Functional Tests
Unit tests will probably be enough. They will check that correct libvirt XML is created.

If we want to add a functional test, we can copy [this test](https://github.com/kubevirt/kubevirt/blob/a37a2b317c335bc1a662d57bd9137d715f52e202/tests/storage/storage.go#L375),
and modify it to enable `generateDiskSerials` field instead of specifying the serial manually.

## Design questions

- Do we want the serials to be generated by default on all VMs,
  or do we want a configuration option to enable it?
  - We would not need any API for generating them. 
  - The downside is that existing disks that previously did not have a serial configured,
    will have it. This could confuse the guest OS that expects the disk serial to not change.

- Where is a good place for the `generateDiskSerials` field?
- Is it a good idea to store the serial in an annotation, or a different approach is better?

## Implementation considerations
The length of the serial number is limited. The [libvirt documentation](https://libvirt.org/formatdomain.html#hard-drives-floppy-disks-cdroms) says:

> Note that depending on hypervisor and device type the serial number may be truncated silently. IDE/SATA devices are commonly limited to 20 characters. SCSI devices depending on hypervisor version are limited to 20, 36 or 247 characters.
>
> Hypervisors may also start rejecting overly long serials instead of truncating them in the future so it's advised to avoid the implicit truncation by testing the desired serial length range with the desired device and hypervisor combination.