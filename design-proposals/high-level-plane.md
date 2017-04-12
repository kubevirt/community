# High-level plane

Author: Fabian Deutsch \<fdeutsch@redhat.com\>

## Introduction

Today KubeVirt is exposing a pretty low-level interface to virtualization capabilities.

But this low-level API might be too low-level for i.e. UIs to integrate. On the higher level its typically required to set policies, the lower layer then provides the mechanics to implement this policy. The suggested high-level plane is the middle layer, translating policies into the mechanics.

The introduction of a high level plane also tries to prevent that too much logic is pushed into the lower level mechanics.

There will be users for a high-level plane, but we assume that the low-level plane is still relevant, that is the reason why a separate API is suggested, instead of changing the existing one.

### Use-case

The primary focus is to provide a convenient high-level API suitable for integration into UIs.

A couple of aspects will be the focus of this initial part:

* VM abstraction (support stopped VMs, OS flavors, disk abstraction, network abstraction, )
* VM templating
* …

## API

TBD
The API is probably just a more high-level VM Spec for now, it would exist in parallel to the low-level representation.
Example:

kind: VirtualMachine
spec:
  state: down
  os:
    flavor: Windows 10
  disks:
    - name: windows-disk-pvc-from-kube
      type: disk
    - name: windpws-iso-pvc-from-kube
      type: cdrom
  nics:
    - network: foonet-from-kube
      model: virtio
  display:
    heads: 2

FIXME Is there a bidirectional conncetion between high- and low-level VM Spec?

## Implementation

The implemantation will reside in a separate controller, which can be an addon to KubeVirt.

TBD
* Separate controller
* Watches for high-level specs
* Links to low level specs

