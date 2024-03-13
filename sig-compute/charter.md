# SIG Compute Charter

## Scope

SIG-compute's scope is enormous, and includes node configurations, live-migration,
virtualization, and much more.  

In fact, currently sig-compute's scope captures the vast majority of the ares in the
project, being the "default" sig for everything not directly related to the scope of other
SIGs (such as storage and network).

In the future we aim to break sig-compute into more sub-SIGs, which will be more granular
and help define a concrete scope.

### In scope

Following topics are in-scope.

We can use these topics as a base ground for thinking how to break sig-compute in the future.

#### Node
- SELinux
- Cgroups
- TPM
- TSC
- Swap
- KSM
- External kernel boot
- seccomp
- node-labeller
- and more

#### Virtualization
- virt-launcher
- libvirt / QEMU configuration
- CPU features
- Dedicated CPUs
- Hypervisor features (e.g. HyperV)
- devices management: generic host devices interface, GPU and vGPU discovery configuration
- and more

#### Live-migration
- Migration policies
- Migration configuration
- Migration performance
- and more

### Out of scope

- Specific network/storage related aspects of live-migration

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OWNER_ALIASES](https://github.com/kubevirt/kubevirt/blob/main/OWNERS_ALIASES)
file.

SIG chairs:
- jean-edouard
- iholder101

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts
- Ensure node stability is not compromised
