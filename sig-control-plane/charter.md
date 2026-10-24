# SIG Control Plane Charter

## Scope

SIG Control Plane is responsible for the development and enhancement of
KubeVirt control plane.

### In scope

#### Code, Binaries and Services

Maintenance of core control plane components such as virt-api, virt-controller
and virt-operator. This includes the deployment, scalability and
overall performance of these components.

This work in the `kubevirt/kubevirt` repository is recorded under the `kubevirt`
subproject in [sigs.yaml](../sigs.yaml), which lists its leads and the OWNERS files
defining code ownership.

### Out of scope

The node-side virt-handler component is owned by SIG Compute.

The contents of individual APIs are overseen by SIG Architecture and owned by
specific SIGs.

SIG specific implementations within control plane components are also owned by
specific SIGs. For example, network specific controller logic within
virt-controller is owned by sig-network.

## Roles and Organization Management

This SIG follows the Roles and Organization Management outlined in [OARP]

SIG chairs:

- [@lyarwoood](https://github.com/lyarwood)
- [@xpivarc](https://github.com/xpivarc)

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts

[OARP]: https://stumblingabout.com/tag/oarp/
