# Overview

This design provides an overview of adding a dedicated kubevirt ephemeral csi driver. This csi driver will be part of the kubevirt installation and will be required for certain features such as volume hotplug.

## Motivation

The ephemeral csi driver will allow one to specify [csi ephemeral volumes](https://kubernetes.io/docs/concepts/storage/ephemeral-volumes/#csi-ephemeral-volumes) inside a pod spec. Csi ephemeral volumes act very similar to an [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) in that they are created when a pod starts, and will be removed when the pod stops. Like emptyDir, the csi ephemeral volumes provide a stable path for the volume inside a pod (like the virt-launcher), that is also visible on the node, the pod is running on. This allows the node (or a privileged pod on that node that has the host file system mounted, like the virt-handler pod) to manipulate the contents of the path and have them show up inside the pod. This is important for features like volume hotplug that bind mount volumes into the virt-launcher pod using that volume path. 

Another property of csi ephemeral volumes is that they can be specified inline in a pod spec and do not require an explicit persistent volume to function. When a pod using the csi ephemeral volume terminates, the csi driver is free to clean up the contents of the ephemeral volume, or perform any other operation it wants on the ephemeral volume.

As said above a csi ephemeral volume is very close in function to an emptyDir, so why not simply use an emptyDir? The main motivation in suggesting an ephemeral csi driver is control over what happens when a pod exits. The following scenario can happen when using an emptyDir in combination with volume hotplug:

1. We have a running VM, and have hotplugged one or more volumes into it. We have used an emptyDir as a stable path inside the virt-launcher which we used to attach the disks in those volumes to the VM (this is what the current volume hotplug implementation does).
2. Something causes the virt-launcher to be force deleted (user action, resource pressure, whatever)
3. When we normally terminate a VM, the virt-handler is notified we are going to terminate the VM, and is given a chance to unmount any volumes before the virt-launcher pod exits.
4. When a pod that uses an emptyDir terminates, the kubelet will remove all the contents of the volume backing the emptyDir.
5. Because we force terminated the virt-launcher pod, we are now in a race between the kubelet, and the virt-handler (normally the grace period will give virt-handler enough time). If the virt-handler is called to unmount and hotplugged volumes before the kubelet clears the emptyDir, we are fine. However if the kubelet is run before the virt-handler, it will blindly remove the contents of the emptyDir volume, which, in the hotplug volume case will be the disk.img files that have not been unmounted. Since it is a bind mount this will result in the disk.img file on the actual volumes being removed, and we have lost data.

If we use a csi ephemeral volume controlled by a kubevirt specific csi driver, we can control the behavior in that scenario and unmount any mounted volumes inside the csi ephemeral volume before clearing it.

## Goals

* Design and implement a kubevirt specific csi driver that only supports **ephemeral volumes**.
* Use csi ephemeral volumes to fully control what happens when virt-launcher pods terminate with regards to volume hotplug.

## Non Goals

Implement a full fledged CSI driver that also supports persistent volumes. We have other projects like the [hostpath csi driver](https://github.com/kubevirt/hostpath-provisioner) for that.


## User Stories

* As a KubeVirt user I want to be guaranteed that if I have hotplugged volumes in my VM, and the virt-launcher pod terminates unexpectedly, I will not lose any data on my volumes.

## Repos
kubevirt/kubevirt

# Design
The idea is to implement just the ephemeral parts of the csi spec. Examples of drivers that do this are [csi-driver-image-populate](https://github.com/kubernetes-csi/csi-driver-image-populator) and [secrets-store-csi-driver](https://github.com/kubernetes-sigs/secrets-store-csi-driver).

To support just the ephemeral portions, we need to implement the appropriate parts of the csi driver, and start the [node-driver-registrar](https://github.com/kubernetes-csi/node-driver-registrar) so we can register the driver with kubernetes. The csi driver will have to run on each node, so will require a daemonset. We could make an additional daemonset, or add additional containers to the virt-handler daemonset. The daemonset will require privileged permissions due to the nature of gRPC socket used to communicate between the different CSI components.

In addition we will have to create the CSI driver object as shown below (I used example naming, we can pick something different if we want)

```yaml
apiVersion: storage.k8s.io/v1
kind: CSIDriver
metadata:
  name: kubevirt.csi.k8s.io
  labels:
    app.kubernetes.io/instance: kubevirt.csi.k8s.io
    app.kubernetes.io/part-of: kubevirt.io
    app.kubernetes.io/name: kubevirt.csi.k8s.io
    app.kubernetes.io/component: csi-driver
spec:
  # Supports only ephemeral inline volumes.
  volumeLifecycleModes:
  - Ephemeral
```

One implementation detail to work out, is what to use where to create the volumes. The natural initial implemention would be a directory on the node somewhere, maybe the same directory as where other kubevirt specific configuration is stored?

## Scalability

Since the volumes are ephemeral and will be created on the node the virt-launcher pod is started on, the scaling should be the same as an emptyDir or other ephemeral volumes.

## Update/Rollback Compatibility

Since this is a change in how hotplugged volumes would be mounted inside the virt-launcher pod, there is a potential problem in the virt-handler. The virt-handler uses a specific path format to locate the emptyDir volume used in the virt-launcher. This path is then used to bind mount the volumes with the disk.img files. We would have to update virt-handler to temporarily support both emptyDir and csi ephemeral volumes to allow running VMs to still function properly after upgrading.

## Functional Testing Approach

There is an extensive test suite for hotplugged volumes which should still pass when using the csi ephemeral volumes. In addition we should add some csi ephemeral volume specific tests from the k8s csi suite.

# Implementation Phases

* Implement the ephemeral csi driver.
* Add/Modify daemonset to use it.
* Modify volume hotplug implement to use csi ephemeral volumes