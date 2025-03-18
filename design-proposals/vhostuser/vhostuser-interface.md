
# Overview
- User Space networking is a feature that involves packet path from the VM to the data plane running in userspace like OVS-DPDK etc, by-passing the kernel.
- With VhostUser interface, the kubevirt VM created can utilize the dpdk virtio functionality and avoid using kernel path for the traffic when DPDK is involved, allowing usage of fast datapath and improve performance.

## Motivation
- Any kubevirt VM uses linux kernel interfaces to connect to its host and to reach out externally or to communicate within the cluster. The current kubevirt VMs can only support the kernel data path for which the traffic needs to be through the kernel of the node/compute hosting the VM.

- As a result, when we have a host with DPDK dataplane running in Userspace, the traffic from kubevirt VM will go through the kernel and reach the dataplane running in userspace. This longer path slows down the traffic between the application and the dataplane when both are running in userspace.

- In order to improve the traffic performance and make use of the dpdk, we need to use a different path that can by-pass the kernel and reach the dataplane from the user space directly. This so called fast path will run within the user space and will allow fast packet processing.

## Goals
- The kubevirt VM when created with DPDK features, should be able to use fast datpath if the host supports the DPDK dataplane allowing fast packet processing to improve performance of the traffic.

## Non Goals
- The Userspace CNI with the dataplane running in userspace like dpdk/vpp, should make sure proper handling of mounting and unmounting the volumes. These changes are CNI specific and might vary from CNI to CNI.
- The kubevirtci creates kernel mode dataplane hosts for testing the VMs. It needs to be updated to support DPDK.

## Definition of Users
- Any user who has a (k8s) cluster with a dpdk dataplane can use the interface/feature.
- The feature will be supported only if a multus userspace CNI is being used, for instance userspace cni from intel.

## User Stories
- As a user/admin, I want to use user space networking feature if I have a VM with dpdk virtio interfaces and the host has a DPDK dataplane and userspace cni with multus support, by-passing the traffic from kernel of the host.

## Repos

- kubevirt/kubevirt
   - Introduction of vhostuser type interface.
- kubevirt/kubevirtci
   - Enable usage of dpdk on the setups created by kubevirtci.

# Design
- The design involves introducing vHostUser interface type and required parameters to support it. An EmptyDIR volume (shared-dir) and DownwardAPI (podinfo) volume are mounted on the virtlauncher pod to have the support for new interface.
- These mounts will allow the virt-launcher pod to create a VM with an additional interface which can be reached using the vhostSocket mentioned in the DownwardAPI.

- Creating the VMs will follow the same process with a few changes highlighted as below:
    1. **Once the VM spec virt-controller will add two new volumes to the virt-launcher pod.**\
        a. **EmptyDir volume named (shared-dir) "/var/lib/cni/usrcni/" from virt-launcher pod is used to share the vhostuser socket file with the virt-launcher pod and dpdk dataplane. This socket file acts as an UNIX socket between the VM interface and the dataplane running in the UserSpace.**\
        b. **DownwardAPI volume named (podinfo) is "/etc/podnetinfo/annotations" used to share vhostuser socket file name with the virt-launcher via pod annotations. The downwardAPI is used here to have the pod know the socket details like name and path, which is created by CNI/kubemanager while bringingup the virt-launcher pod and this info is only availbale after the pod is created**
    2. **The CNI should mount the shared-dir  to "/var/lib/cni/usrcni". The CNI can link the EmptyDir volume's host path /var/lib/kubelet/pods/<podID>/volumes/kubernetes.io~empty-dir/shared-dir to a custom path CNI prefers to have. The CNI should create the sockets in the host in the custom path. For each interface, CNI should create a socket.**
    3. **The CNI should update the virt-launcher pod annotations with vhostsocket-file name and details using the downwarAPI. The /etc/podnetinfo/annotations from the virt-launcher pod will hold the information of the socket(s) details.**
    4. **The virt-launcher reads the DownwardAPI volume and retrieves the vhostsocket-file name specified in pod annotations and uses it while launching the VM using libvirtd.**
    4. **The virt-launcher is modified to skip establishing the networking between VM and the virt-launcher pod using bridge(Refer in Kubevirt Networking section in https://kubevirt.io/2018/KubeVirt-Network-Deep-Dive.html).**
    5. **Instead of using the bridge through the launcher pod, the vHostUser interface of the VM will be directly connected to the DPDK datplane using the vhost socket**

- Removing the VMs will follow the below chnages:
    1. **The CNI should delete the socket files of each interface. If 'n' interfaces are present, 'n' number of socket files will be deleted allowing k8s to proceed with a clean pod deletion.**
    
   No additional changes are required while deleting the VM, it will be deleted as a regular VM, the CNI should handle the deletion of interfaces. In the case of vHostUser, CNI should delete the socket files before k8s initiates pod deletion.

- With the above process a kubevirt VM with a vHostUser interface can be acheived. All the steps highlighted will be the changes made as part of the design.

## API Examples

Since, KubeVirt will always explicitly define the pod interface name for multus-cni. It will be computed from the VMI spec interface name, to allow multiple connections to the same multus provided network.

The vHostUser Interface will be defined in the VM spec as  shown below:

```yaml
          interfaces:
          - name: vhost-user-vn-blue
            vhostuser: {}
          useVirtioTransitional: true
      networks:
      - name: vhost-user-vn-blue
        multus:
          networkName: vn-blue
```
No additional definition of vloumes or annotations are required. The DownwardAPI will provide all the necessary annotations required for vhostUser Interface.

## Scalability
- There should not be any scalability issues, as the feature is to create an interface when useful.
- Regarding VM live migration, we need to check if the socket files created will be able to support the VM migration.

## Update/Rollback Compatibility
- Should have no impact.

## Functional Testing Approach
Functional test can:

- Create VM with vhostUser interface
- Create VM with multiple interface types. (Bridge + vhostuser type)

# Implementation Phases
- The design will be implemented in two phases:
    - Add DPDK support on kubevirt/kubevirtci
        - Enable OVS with DPDK
        - Enable an option in gocli for DPDK support
        - Add UserSpace CNI in the setup instead of or along with calico.
    - Add vhostUser Interface type in kubevirt/kubevirt
        - Add vHostUser Interface type
        - Create Pod template to support the Interface
        - Create appropriate virsh xml elements
        - Add E2E test to send traffic between the 2 VMs created.

## Annex


A sample DownwardAPi mght look this way:

```yaml
k8s.v1.cni.cncf.io/network-status="[{\n    \"name\": \"k8s-kubemanager-kubernetes-CNI/default-podnetwork\",\n    \"interface\": \"eth0\",\n    \"ips\": [\n        \"10.244.2.4\"\n    ],\n    \"mac\": \"02:d7:b3:27:b3:b2\",\n    \"default\": true,\n    \"dns\": {}\n},{\n    \"name\": \"kubevirttest/vn-blue\",\n    \"interface\": \"net1\",\n    \"ips\": [\n        \"19.1.1.2\"\n    ],\n    \"mac\": \"02:c6:da:d3:ab:07\",\n    \"dns\": {},\n    \"device-info\": {\n        \"type\": \"vhost-user\",\n        \"vhost-user\": {\n            \"mode\": \"server\",\n            \"path\": \"2ea1931c-2935-net1\"\n        }\n    }\n}]"
k8s.v1.cni.cncf.io/networks="[{\"interface\":\"net1\",\"name\":\"vn-blue\",\"namespace\":\"kubevirttest\",\"cni-args\":{\"interface-type\":\"vhost-user-net\"}}]"
kubectl.kubernetes.io/default-container="compute"
kubernetes.io/config.seen="2023-05-16T16:50:04.224937236Z"
kubernetes.io/config.source="api"
kubevirt.io/domain="vm-single-virtio"
kubevirt.io/migrationTransportUnix="true"
post.hook.backup.velero.io/command="[\"/usr/bin/virt-freezer\", \"--unfreeze\", \"--name\", \"vm-single-virtio\", \"--namespace\", \"kubevirttest\"]"
post.hook.backup.velero.io/container="compute"
pre.hook.backup.velero.io/command="[\"/usr/bin/virt-freezer\", \"--freeze\", \"--name\", \"vm-single-virtio\", \"--namespace\", \"kubevirttest\"]"
pre.hook.backup.velero.io/container="compute"
```
The CNI should update the /etc/podnetinfo/annotaions with something similar, to update the annotations of the virt-launcher pod.

The NAD can be generic which can just be used for defining the networks. The below is a NAD definition from userspace CNI based on ovs-dpdk by intel.

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: userspace-ovs-net-1
spec:
  config: '{
        "cniVersion": "0.3.1",
        "type": "userspace",
        "name": "userspace-ovs-net-1",
        "kubeconfig": "/etc/kubernetes/cni/net.d/multus.d/multus.kubeconfig",
        "logFile": "/var/log/userspace-ovs-net-1-cni.log",
        "logLevel": "debug",
        "host": {
                "engine": "ovs-dpdk",
                "iftype": "vhostuser",
                "netType": "bridge",
                "vhost": {
                        "mode": "client"
                },
                "bridge": {
                        "bridgeName": "br-dpdk0"
                }
        },
        "container": {
                "engine": "ovs-dpdk",
                "iftype": "vhostuser",
                "netType": "interface",
                "vhost": {
                        "mode": "server"
                }
        },
        "ipam": {
                "type": "host-local",
                "subnet": "10.56.217.0/24",
                "rangeStart": "10.56.217.131",
                "rangeEnd": "10.56.217.190",
                "routes": [
                        {
                                "dst": "0.0.0.0/0"
                        }
                ],
                "gateway": "10.56.217.1"
        }
    }'

```
