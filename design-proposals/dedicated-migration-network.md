# Overview
This design provides a way to migrate VMIs (Virtual Machine Instances) over a dedicated network.
VMI migrations currently happen over the main Kubernetes network all pods are connected to.

## Motivation
In order for VMI migrations to succeed, there must be a fair amount of bandwidth available.
If for example the memory of a migrating VMI changes faster than the current speed of the main network, the migration will never succeed.
This is especially problematic during mass eviction events like platform upgrades.
This problem is currently mitigated by throttling the number of concurrent migrations and using long timeouts.

## Goals
- Allow a cluster admin to setup and dedicate a secondary physical cluster network to migrations.
- Keep the main network as the default medium for migrations.

## Non Goals
- Network configuration. A functional secondary physical network should be provided.
- Network definition. A valid NetworkAttachmentDefinition should be written by the cluster admin and provided to the KubeVirt CR (Custom Resource).
- Cluster components. A multi-network CNI like multus should be properly installed, as well as necessary plugins, and any required daemon (such as CNI DHCP) should be running on each node.

## User story
As an administrator of the cluster I want all my VMIs to be migrated over a dedicated network to provide a stable connection with a larger bandwidth comparing to the default k8s pod network

## Definition of Users
This feature is intended for cluster admins who wish to accelerate and/or control migration network usage.

## Repos
- kubevirt/kubevirt:
  - add a CR option and handle it at the virt-handler level
- kubevirt/kubevirtci:
  - install whereabouts on CNAO lanes

# Design
If a dedicated migration network is defined in the KubeVirt CR, migration target virt-handler should communicate their IP of that network instead of the main one.
The source virt-handler will then use that IP to communicate with the target, which will get routed through the secondary network.

## Cluster setup example
- Every node has an eth1 connected to the secondary network that will be used for migrations
- A DHCP server is running on that network
- Multus is installed on the cluster:
```
git clone https://github.com/k8snetworkplumbingwg/multus-cni.git
cat ./images/multus-daemonset.yml | kubectl apply -f -
```
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
      "cniVersion": "0.3.0",
      "name": "bridge2",
      "type": "macvlan",
      "master": "eth1",
      "mode": "bridge",
      "ipam": {
           "type": "dhcp"
      }
    }'
```

## API Examples
In the KubeVirt CR, a field named dedicatedMigrationNetwork will be added under spec.configuration.migrations, like so:
```
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - LiveMigration
    migrations:
      dedicatedMigrationNetwork: my-secondary-network
```

## Functional Testing Approach
Functional testing will use the CNAO lanes with multiple nodes and multiple networks, like so:
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
