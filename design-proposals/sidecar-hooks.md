# Sidecar hooks

## Overview

A hooking mechanism that allows customizations to the VMI before it starts without changing KubeVirt
core components. The common user of those hooks are Sidecars that will run in the same Pod that the
VMI will be deployed.

## Motivation

A way to accommodate use cases that we might not want to bring and support in KubeVirt itself or to
give room to test and develop features while they are still not mature enough.

## Goals

- Provide and maintain a set of gRPC APIs for users to do modifications to the VMI
- Provide and maintain sidecars that are important for core functionalities

## Non Goals

- Provide and maintain a wide variety of sidecars that are not part of KubeVirt's core
  functionalities.

## User Stories

- As a VM owner, I want to apply custom configurations to QEMU command line
  - For debugability, by adding support or tweak [to run with debug tools][]
  - To test unsupported features or apply changes that [could workaround bugs][]
- As a KubeVirt developer, I want to provide alternative solutions that might not be ready to core
  KubeVirt components
  - For Network configurations like [Slirp][] and [Passt][]

## Design

KubeVirt should provide a versioned gRPC Hook module, with all the APIs that can be used including
parameters and responses.

The gRPC server runs in the Sidecar. The Sidecars will use the module to establish the communication
with virt-launcher and exchange capabilities to define which Hook version is being used and what are
the hooks that are implemented.

Virt-launcher will be entitled with managing the hooks. It'll verify and connect to all the Sidecars
that are running by checking the pre-defined folder where the unix sockets for gRPC communication
will be created.

Virt-launcher should apply the requested changes to the VMI.

## Hooks API promotion

Should follow the same rules as stated in [API Graduation Guidelines][]

[to run with debug tools]: https://kubevirt.io/user-guide/debug_virt_stack/launch-qemu-strace/
[could workaround bugs]: https://github.com/kubevirt/kubevirt/issues/8420
[Slirp]: https://github.com/kubevirt/kubevirt/pull/10272
[Passt]: https://github.com/kubevirt/kubevirt/pull/10425
[API Graduation Guidelines]: ../docs/api-graduation-guidelines.md
