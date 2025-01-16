# Overview
The request to live migrate a VM is represented by an instance of VirtualMachineInstanceMigration kind.
VirtualMachineInstanceMigration is a namespaced CRD and its instances are expected to be in the namespace of the VM they are referring to.
As for today, by default, a namespace admin (usually a namespace "owner" in a less formal definition) is able to create VMs and also VirtualMachineInstanceMigrations (VMIMs) objects to enqueue a live migration request for a VM within his namespace.
On the other side, live migration is also a building block for infrastructure critical operations like node drains or upgrades, so potentially every namespace admin allowed to create or delete VMIMs objects can delay or even prevent a cluster critical operation such as an upgrade.
This proposal is about amending KubeVirt default RBAC roles according to the principle of least privilege so that namespace admins will be allowed to create or delete VirtualMachineInstanceMigrations objects only if explicitly allowed by a cluster admin.
The present proposal is not an API change since the API will remain exactly the same, the change is only about who is allowed by default to call the API.

## Motivation
Live migration is an essential building block for cluster critical operations like node drains or upgrades.
A less privileged user should not be allowed by default to be able to potentially interfere with infrastructure critical operations.

## Goals
- Namespace admins will not be allowed anymore by default to enqueue or delete migration request for the VMs in their namespaces
- If needed, the right to enqueue or delete migration requests could be still granted by cluster admins to individual users or groups

## Non Goals
- This is only affecting explicit live migration requests, other operations like hot-plugging a device to a VM can rely on live migration but, in this case, the migration request will be handled by a KubeVirt controller and so is not going to be limited by this change.
- Building a priority mechanism to prioritize infrastructure required migrations is out of scope for this proposal: while a priority mechanism will still be beneficial, it would not solve by itself the issue of unprivileged users being able to dequeue infra required migrations affecting their VMs delaying or blocking infra tasks.

## Definition of Users
See also [the default k8s user facing roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles). 
- cluster-admin: the role allows super-user access to perform any action on any resource. Allows super-user access to perform any action on any resource. When used in a ClusterRoleBinding, it gives full control over every resource in the cluster and in all namespaces. When used in a RoleBinding, it gives full control over every resource in the role binding's namespace, including the namespace itself.
- admin: the role allows admin access, intended to be granted within a namespace using a RoleBinding.
  If used in a RoleBinding, allows read/write access to most resources in a namespace, including the ability to create roles and role bindings within the namespace.

We are assuming that binding the default admin role to namespace "owners" (inappropriate but common concept, read admins) is a common practice. 

## User Stories
- as a cluster-admin, I will be allowed to enqueue or delete migration requests exactly as today
- as a namespace admin (read as a namespace owner), I will not be allowed by default to enqueue or delete migration requests for the VMs in my namespace
- as a namespace admin (read as a namespace owner), I will be still allowed by default to hotplug a device to my VMs even if this is implicitly triggering a live migration
- as a cluster-admin, I will have an easy way to grant a role to allow existing users or groups to manage live migrations in their own namespace or even at cluster scope
- as a cluster-admin, I will have an escape hatch to restore the previous behavior at cluster scope even for future users

## Repos
https://github.com/kubevirt/kubevirt

# Design
KubeVirt is currently defining three default cluster roles: `kubevirt.io:admin`, `kubevirt.io:edit`, `kubevirt.io:view` that are respectively labeled with `rbac.authorization.k8s.io/aggregate-to-admin=true`, `rbac.authorization.k8s.io/aggregate-to-edit=true` and `rbac.authorization.k8s.io/aggregate-to-view=true` to be aggregated to `admin`, `edit` and `view` default k8s cluster roles.
The admin cluster role is usually bound with a namespaced RoleBinding to namespace owners.
In the past users with the admin and the edit role at namespace were allowed to `get`, `delete`, `create`, `update`, `patch`, `list`, `watch` and `deletecollection` on `VirtualMachineInstanceMigrations` objects and `update` the `/migrate` subresource on VM objects.
The proposal is to not grant anymore that by default to namespace admins and create instead a new dedicated role named `kubevirt.io:migrate` defined as:
```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    app.kubernetes.io/component: kubevirt
    app.kubernetes.io/managed-by: virt-operator
  name: kubevirt.io:migrate
rules:
- apiGroups:
  - subresources.kubevirt.io
  resources:
  - virtualmachines/migrate
  verbs:
  - update
- apiGroups:
  - kubevirt.io
  resources:
  - virtualmachineinstancemigrations
  verbs:
  - get
  - delete
  - create
  - update
  - patch
  - list
  - watch
  - deletecollection
```

Cluster-admins can still grant it to individual users at namespace scope with something like:
```
kubectl create -n usernamespace rolebinding kvmigrate --clusterrole=kubevirt.io:migrate --user=user1 --user=user2 --group=group1`
```
or at cluster scope:
```
kubectl create clusterrolebinding kvmigrate --clusterrole=kubevirt.io:migrate --user=user1 --user=user2 --group=group1
```
or even restore the previous default behavior with something like:
```
kubectl label --overwrite clusterrole kubevirt.io:migrate rbac.authorization.k8s.io/aggregate-to-admin=true rbac.authorization.k8s.io/aggregate-to-edit=true`
```

The cluster-admin role is defined as: 
```
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
```
and k8s RBAC rule is purely addictive, so no other rules are required to let cluster-admin rules handle `VMIMs`.

## Alternative designs
### Make VirtualMachineInstanceMigrations cluster scoped
VirtualMachineInstanceMigrations are namescoped exposing only `vmiName` field under spec assuming that it can only refer to a VM in the same namespace.
Amending VirtualMachineInstanceMigrations CRD to make it cluster scoped adding a `vmiNamespace` field is a bold backward incompatible change and so it should avoided if other simpler options are available.

### Introduce a priority mechanism for VirtualMachineInstanceMigrations
As a next step, we can implement a priority queue (similar to the [controller-runtime mechanism](https://github.com/kubernetes-sigs/controller-runtime/pull/3014)) for live migrations. With such a mechanism we would be able to easily accommodate the "semi-user-triggered" migrations, such as various hotplugs. These will come with a lower priority compared to upgrades and node drains.
But the priority mechanism is just a nice to have second order optimization: by itself it's not enough to solve the issue of a less privileged user blocking a node drain or an upgrade by continuously deleting the VirtualMachineInstanceMigrations objects as soon as created by KubeVirt controllers while performing cluster wide activities.  

## API Examples
No API changes

## Scalability
No concerns, it's only amending the default KubeVirt cluster roles restricting them.  

## Update/Rollback Compatibility
For maintainability and harmonization purposes, we aim to have clusters deployed starting with different releases to behave by default in the same way.
After an upgrade, KubeVirt operator will amend `kubevirt.io:admin`, `kubevirt.io:edit` cluster roles removing the rules that allow namespace admin to manage VMIMs so an upgraded cluster will behave exactly a fresh deployed one.
 A cluster admin could still restore the previous behavior simply relabeling the new `kubevirt.io:migrate` clusterrole. The KubeVirt operator will ignore custom labels there honoring the cluster-admin configuration.
A really conscious cluster-admin that does not allow any disruption due to the upgrade process could still create temporary ClusterRole for the migration before the upgrade labeling it with `rbac.authorization.k8s.io/aggregate-to-admin=true`.
Something like:
```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    rbac.authorization.k8s.io/aggregate-to-admin=true
  name: kubevirt.io:upgrademigrate
rules:
- apiGroups:
  - subresources.kubevirt.io
  resources:
  - virtualmachines/migrate
  verbs:
  - update
- apiGroups:
  - kubevirt.io
  resources:
  - virtualmachineinstancemigrations
  verbs:
  - get
  - delete
  - create
  - update
  - patch
  - list
  - watch
  - deletecollection
```
This cluster role will be aggregated to the `admin` role before the Kubevirt upgrade and the upgrade process will not touch it enforcing the previous behavior.
After the upgrade, the cluster admin will have all the time to grant the new `kubevirt.io:migrate` clusterrole to selected users before dropping the temporary clusterrole.

## Functional Testing Approach
### Unit Testing Approach
Unit tests will enforce that `kubevirt.io:admin`, `kubevirt.io:edit` and `kubevirt.io:migrate` are configured as defined in this proposal regarding `VirtualMachineInstanceMigrations` objects and `/migrate` subresource.

### Functional Testing Approach
`[rfe_id:500][crit:high][vendor:cnv-qe@redhat.com][level:component][sig-compute]User Access` tests will be amended to match the behavior described here. 

# Implementation Phases
This change can be handled with a single PR. [A POC is already available](https://github.com/kubevirt/kubevirt/pull/13497).
