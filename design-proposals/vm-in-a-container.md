# KubeVirt Node driver

## Introduction

Running a VM inside a Pod, but starting it from stock libvirt, which has no
ideas about pods, we have to somehow derive cgroups and namespaces of where to
start the VM in.
This should be as transparent as possible to libvirt.

## Implementation

We can create a small binary which will be invoked by libvirt as custom
emulator. This binay makes sure the VM is started in the right cgroups and
namespaces of the target container.  This binary will have the `suid` bit set.

To find out the one shim process associated to a VM, we can run the shim
process with the uuid and the vm name in the commandline like this:

```bash
virt-launcher -kubevirt.vm.uuid 1234-5678-1234-1234 -kubevirt.vm.name testvm
```

Further we can for foward the same information to the emulator binary by
passing additinoal qemu environment variables to it with the same information:

```xml
<domain type='qemu' xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
<devices>
    <emulator>/usr/local/bin/myemulator</emulator>
</devices>
<qemu:commandline>
   <qemu:env value='kubevirt.io.vm.uuid 1234-5678-1234-1234'/>
   <qemu:env value='kubevirt.io.vm.name testvm'/>
</qemu:commandline>
</domain>
```

Note that we have to specify the qemu namespace to use these features.

Based on this information, the binary can check all runninig processes for the
matching shim process and derive namespaces and cgroups from it. Then it can
start the real qemu process inside them.  For namespaces it can use the
`setns()` system call.

### Security

Since the driver binary will have the `suid` bit set, to do all necessary
operations. This imposes one big security risk: The binary can start
applications in other namespaces and cgroups (even the one of the host).

A few things can be done to reduce this risk:

 1. Only allow the qemu user to execute the binary
 2. Introduce a shared secret between the VM target container and kubevirt
 3. This secret can be unique per pod creation and only valid for a specific
    amount of time
 4. The binary checks that  the secret is present in the kubernetes metadata
    file before it starts something in that container 

The secret can be delivered to the container through Kubernetes secrets. The
binary can receive the secret via libvirts <qemu:commandline/> tag as
environment variable:

```xml
<qemu:commandline>
   <qemu:env value='kubevirt.io.secret 3098tFJoswfwkjp4'/>
</qemu:commandline>
```

Note that this means that the secret is exposed in the Domain xml. If we think
that having a secret lifespan of e.g. one minute is not secure enough, the
secret could be RSA encrypted. 

### Flow

Pseudocode:

```
# Getting the pid os container runtime independent
pid = $(ps aux  | grep virt-launcher | grep "-kubevirt.vm.uuid 1234-5678-1234-1234 | tr -s ' ' | cut -d " " -f2)

if SecretFromPidMountNamspace(pid) == decrypt(providedSecret) && SecredFromPidMountNamespaceStillValid(pid) {
  executeQemuInPidNsAndCgroupContext(pid)
} else {
  fail()
}
```

### Running it inside a container

To run the binary from inside a container, it needs access to the host `/proc`
directory. The recommended implementaiton is, to have a config file for the
binary in the container on a well known location (e.g.
`/etc/kubevirt/qemu-kube.conf`). In there the path to `/proc` can be specified.
On the specified location, the host `/proc` can be mounted.

### Delivering the binary

At the end, the binary needs to be present in the mount namespace of where
libvirt is running. The binary should be delivered via an init container to the
host via `hostDir` mount (if libvirt is running on the host), or via an
`emptyDir` mount to a container (if libvirt is running in a container).

There are not many problems to be expected regarding to permissions since the
binary can

 * have the suid bit
 * have permission `-rwxr-xr-x`

when we implement the secret properly.

## Advantages

 * Cgroup and Namespace handling is done on one single place, the rest of
   KubeVirt does not have to know about them (good for migration, ...)
 * Container Runtimer independent
 * Can run directly on the node
 * Can run inside a container
 * Pretty secure because of the shared secrets
 * Can be delivered through init containers
 * Because of suid bit, libvirt/qemu don't require any extended privileges

## Disadvantages

 * If run inside a container, the container needs /proc mounted and needs to be
   privileged
 * Suid bit required, so it has to be done right
 * Still multiple compiled versions (musl libc, glibc)

## Other considered solutions

 * Talking to container runtime directly and have a container alongside
   virt-handler
 * Provide an api in virt-launcher which queries /proc/self to get namespaces
   and cgroups
 * Using libvirts own cgroup change mechanism

They all have the disadvantage of a lot of complex flows to find the right
information at the right time, while we would still required a minimal driver
where libvirt points to, to do the actual namespace changes.

# What is still missing?

Libvirt does preliminary check of the qemu process capabilities. There it does
not forward the additional commandline flags and environment variables.
Therefore we are still missing this one piece of information where to probe for
the qemu capabilities. This affects also all other solutions considered above.
