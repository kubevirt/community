# Initial networking for VMs using pod NICs

Author: Fabian Deutsch \<fabiand@redhat.com\>

## Introduction

VMs need network connectivity. This proposal is about a first start in this direction.

The working assumption is that libvirtd is running with a host-view, thus in the host's
network namespace.
On the other hand we want to connect VMs to the NIC(s) of the pod in which they are running.

To keep it simple and work with what we have today, we just handle a single NIC (thus multiple should
not make a difference).

This proposal does not require changes to Kubernetes.

### Use-case

The primary usecase for this proposal are VMs running in pods with libvirtd running on the host
or in a DaemonSet.
It is an easy way to get the VM simply connected to the NIC available inside the pod where
it is intended to run.

## API

This proposal will enhance the `VM` Resoufce to also accept interfaces like:

```
"interfaces": [
  {
    "source": {
      "dev": "eth0",
      "mode": "bridge"
      },
    "type": "direct"
    }
  }
}
```

The `dev` value above is refering to the `dev` name _inside_ the pod.
Thus the `virt-controller` needs to translate this pod specific name into the host sided endpoint.

The `virt-controller` would transfer this into:

``    <interface type='direct'>
      <source dev='veth0@mypod' mode='vepa'/>
    </interface>
```

The `dev` name is now referring to the veth endpoint `veth0@mypod` in th ehost's network namespace.

## Implementation

Obviously the 1:1 transformation of the VM Spec to the domxml can be reused, except that the `virt-controller` needs to translate the pod centric name into a host centric name.
