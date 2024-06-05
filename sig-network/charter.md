# SIG Network

## Scope

SIG Network owns KubeVirt networking topics.
The main objective of the SIG is to connect the Virtual Machine to the primary and secondary pod networks,
allowing migration and other advantages of virtualization in addition to maintaining services,
port-forwarding, service mesh, and other Kubernetes capabilities.

### In scope
Some notable examples of areas owned by SIG Network:
- Masquerade, Bridge and Passt binding plugins for the primary network.
- Bridge, SRIOV and MacVTap bindings for secondary networks.
- VM networking lifecycle on: creation, migration, hotplug/unplug, restart, stop etc.

#### Code, Binaries and Services
- [Main repo](https://github.com/kubevirt/kubevirt).
- Networking related APIs on: VM, VMI, Kubevirt CR, etc.
- Core network binding infrastructure.
- Network binding plugin infrastructure.
- Examples of binding plugin CNIs and sidecars.
- Virtual Machine/Virtual Machine Instance networking status management.
- Domain networking configuration.
- Pod network definition.
- Unit and e2e tests.
- Developer and user documentation around networking.
- Help for onboarding new SIG Network members.
- Grooming network related bugs.
- IPAM for secondary networks.

### Out of scope
- Node networking, it may be managed by [kubernetes-nmstate](https://github.com/nmstate/kubernetes-nmstate).
- Host to Pod connectivity configuration, it is managed by external CNIs.

## Roles and Organization Management

This sig adheres to the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs
- Uphold the KubeVirt Code of Conduct especially in terms of personal behavior and responsibility.
- Own sig-network CI jobs.
- There are some other projects written and managed by the SIG members:
  - [Cluster Network Addons Operator](https://github.com/kubevirt/cluster-network-addons-operator).
  - [Kube Secondary DNS](https://github.com/kubevirt/kubesecondarydns).
  - [Bridge Marker](https://github.com/kubevirt/bridge-marker).

[OARP]: https://stumblingabout.com/tag/oarp/
