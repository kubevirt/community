# Overview
This design provides a way to migrate VMIs (Virtual Machine Instances) over a dedicated network.  
VMI migrations currently happen over the main Kubernetes network all pods are connected to.

The KubeVirt pod in charge of node-to-node communication for VMI migration is virt-handler.  
There is one virt-handler pod running per node, and they currently only have access to the main network.  
This proposal aims at changing that.

## Motivation
In order for VMI migrations to succeed, there must be a fair amount of bandwidth available.  
If for example the memory of a migrating VMI changes faster than the current speed of the main network, the migration will never succeed.  
This is especially problematic during mass eviction events like platform upgrades.  
This problem is currently mitigated by throttling the number of concurrent migrations and using long timeouts.

## Goals
- Allow a cluster admin to setup and dedicate a secondary physical cluster network to migrations.
- Keep the main network as the default medium for migrations.

## Non Goals
- QoS. That is an obvious next step after enabling the use of secondary networks. However, this is out of scope for the initial implementation.
- Advanced routing. The secondary network is expected to assign IPs on a subnet that does not intersect with the subnet of the main network. Advanced/customized routing is out of scope for this proposal.
- Network configuration. A functional secondary physical network should be provided.
- Network definition. A valid NetworkAttachmentDefinition should be written by the cluster admin and provided to the KubeVirt CR (Custom Resource).
- Cluster components. A multi-network CNI like multus should be properly installed, as well as necessary plugins, and any required daemon should be running on each node.

## User story
As an administrator of the cluster I want all my VMIs to be migrated over a dedicated network to provide a stable connection with a larger bandwidth comparing to the default k8s pod network

## Definition of Users
This feature is intended for cluster admins who wish to accelerate and/or control migration network usage.

## Repos
- [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt):
  - add a CR option and handle it at the virt-handler level
- [kubevirt/kubevirtci](https://github.com/kubevirt/kubevirtci):
  - install [whereabouts](https://github.com/k8snetworkplumbingwg/whereabouts) on network test lanes
  - add a secondary network to network lanes

# Design
If a dedicated migration network is defined in the KubeVirt CR, migration target virt-handler should communicate their IP of that network instead of the main one.   
Dedicated migration networks will be exposed to virt-handlers as "migration0". Their presence and assigned IP can be discovered using the [Downward API](https://kubernetes.io/docs/tasks/inject-data-application/downward-api-volume-expose-pod-information/#the-downward-api).  
The source virt-handler will then use that IP to communicate with the target, which will get routed through the secondary network.

## API
In the KubeVirt CR, the ability to define network roles will be added under spec.configuration.network, like so:
```
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - LiveMigration
    network:
      roles:
      - networkAttachmentDefinition: my-secondary-network
        role: migration
```

## Cluster setup example 1: (recommended) IPs managed by whereabouts
- Every node has an eth1 connected to the secondary network that will be used for migrations
- Multus is installed on the cluster:
```
git clone https://github.com/k8snetworkplumbingwg/multus-cni.git
cat ./images/multus-daemonset.yml | kubectl apply -f -
```
- Whereabouts is installed on the cluster: see https://github.com/k8snetworkplumbingwg/whereabouts#installing-whereabouts
- A NetworkAttachmentDefinition is defined like so:
```
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: my-secondary-network
  namespace: kubevirt
spec:
  config: '{
      "cniVersion": "0.3.1",
      "name": "bridge2",
      "type": "macvlan",
      "master": "eth1",
      "mode": "bridge",
      "ipam": {
        "type": "whereabouts",
        "range": "10.1.1.0/24"
      }
    }'
```

## Cluster setup example 2: DHCP provided by the network
- Every node has an eth1 connected to the secondary network that will be used for migrations
- A DHCP server is running on that network
- Multus is installed on the cluster:
- The CNI DHCP daemon is started on every node:
```
rm -f /run/cni/dhcp.sock
/opt/cni/bin/dhcp daemon
```
- A NetworkAttachmentDefinition is defined like so:
```
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: my-secondary-network
  namespace: kubevirt
spec:
  config: '{
      "cniVersion": "0.3.1",
      "name": "bridge2",
      "type": "macvlan",
      "master": "eth1",
      "mode": "bridge",
      "ipam": {
           "type": "dhcp"
      }
    }'
```

## Functional Testing Approach
Functional testing will use the network lanes with CNAO, whereabouts, multiple nodes and multiple networks, like so:
```
export KUBEVIRT_NUM_NODES=2 KUBEVIRT_NUM_SECONDARY_NICS=1 KUBEVIRT_WITH_CNAO=true
make cluster-up
```
The functional test suite will first:
- Create a NetworkAttachmentDefinition bridged to eth1, like the above, except using the `whereabouts` CNI plugin to avoid requiring a DHCP server.
- Alter the KubeVirt CR to set that as the dedicated migration network
- Wait for virt-handlers to respawn
Then the functional tests will verify that migrations go through that network.

# Future improvements
- Multiple migration networks could be supported, with more granular settings for which VMIs migrate over which network
