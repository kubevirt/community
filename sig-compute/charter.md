# SIG Compute Charter

## Scope

The scope of sig-compute is enormous, and includes node configurations,
live-migration, virtualization, and much more.  

In fact, currently sig-compute's scope captures the vast majority of the areas
in the project, being the "default" sig for everything not directly related to
the scope of other SIGs (such as storage and network).

### In scope

Following topics are spread across a number of subprojects to help maintainers
better focus their time and energy on specific areas of the codebase:

#### sig-compute-node

- SELinux
- Cgroups
- TPM
- TSC
- Swap
- KSM
- External kernel boot
- seccomp
- virt-launcher maintenance
- node-labeller
- libvirt / QEMU configuration
- CPU features
- Dedicated CPUs
- Hypervisor features (e.g. HyperV)
- devices management: generic host devices interface, GPU and vGPU discovery configuration

#### sig-compute-cluster

- Migration policies
- Migration configuration
- Migration performance
- virt-{operator,controller,api} maintenance

#### sig-compute-virtctl

- virtctl maintenance

#### sig-compute-instancetype

- Instance type and preference graduation and maintenance

### Out of scope

- Specific network/storage related aspects of live-migration

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OWNERS_ALIASES](https://github.com/kubevirt/kubevirt/blob/main/OWNERS_ALIASES)
file.

SIG chairs:

- [@jean-edouard](https://github.com/jean-edouard)
- [@iholder101](https://github.com/iholder101)

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts
- Ensure node stability is not compromised
