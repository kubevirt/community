# Overview

VirtualMachine Delete Protection prevents the accidental deletion of VMs from a cluster. As long as deletion protection
is activated for certain VM objects, they cannot be deleted unless the protection is intentionally removed.

## Motivation

KubeVirt clusters are often managed by automations, CLI commands, 3rd party tools or UI actions, for example to deploy,
reconfigure or delete VMs. These automations may result in accidental deletion of VMs that should have not been deleted.
Affected VMs could be business critical to application/s or services the cluster is running, this could cause them a
degradation or out of service, e.g., a SQL server is not available to handle requests.
Moreover, the affected VMs may contain crucial data if underlining PVC is deleted as a result of a cascaded delete. 
Therefore, the VM deletion protection aims to avoid accidental degradation of applications or loss of data.

## Goals

- Create a delete protection for VirtualMachine objects.

## Non Goals

- Create a delete protection for VirtualMachineInstance objects.
- Create a delete protection for virt-launcher Pod objects.

## Definition of Users

VM owners with permissions to delete VMs.
Cluster administrators with permissions to delete VMs.

## User Stories

As a VM owner, I want to protect my VMs from accidental deletion.

## Repos

kubevirt/ssp-operator

# Design

The design is divided into two sections. The first section shows how the VM protection can be enabled by a user, the
second section depicts possible implementations of the feature in the backend.

### Enable the VM Object Protection

To protect VMs against deletion, they need to be labeled with  `kubevirt.io/vm-delete-protection` set to `true` or
`True`. If the label is present with one of the aforementioned values, requests to delete this VM will be rejected
by KubeVirt. By default, this label is not present on a VM. It is up to the user to enable the deletion protection.
The label looks like this in the VM manifests:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  labels:
    kubevirt.io/vm-delete-protection: "true"
spec:
[...]
```

### Disable the VM Object Protection
There are two possible ways to disable the VM protection:

1. Removing the `kubevirt.io/vm-delete-protection` label with the following patch operation:
```bash
$ kubectl patch vm <vm-name> --type=json -p='[{"op": "remove", "path": "/metadata/labels/kubevirt.io~1vm-delete-protection"}]'
```
2. Setting the value of the label `kubevirt.io/vm-delete-protection` to another different from `true` or `True`. 
   For instance, `kubevirtio.io/vm-delete-protection: "false"`. This can be achieved by using the following patch operation:
```bash
$ kubectl patch vm <vm-name> --type=json -p='[{"op": "replace", "path": "/metadata/labels/kubevirt.io~1vm-delete-protection", "value": "false"}]'
```

### Backend

#### Chosen Solution: Deploying a new ValidatingAdmissionPolicy (VAP) in SSP

The VM protection is implemented by adding
a  [ValidationAdmissionPolicy and ValidationAdmissionPolicyBinding](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
objects. On every delete action against a VM object, the ValidatingAdmissionPolicy will check for presence of the
label. If it is present and its value equals “true” or “True” the request will be rejected. The
ValidationAdmissionPolicy and ValidationAdmissionPolicyBinding objects will be configured with the following parameters:

- matchConstrains: Apply to \`virtualmachine\` objects; Operations: DELETE.
- Variables:
    - \`label\`: Returns the VM label field;
        - string('kubevirt.io/vm-delete-protection')
- Validation rules:
    - Checks if the label is present at all, if present, checks if it is equal to “true” or “True”;
        - (\!(variables.label in oldObject.metadata.labels) || \!oldObject.metadata.labels\[variables.label\].matches('^(true|True)$'))


Pros:

- Kubernetes built-in feature which is straightforward to maintain, implement and configure.
- ValidationAdmisssionPolicy is [GA](https://kubernetes.io/blog/2024/03/12/kubernetes-1-30-upcoming-changes/) since Kuberentes version 1.30.

Cons:

- It will not work in deployments using Kubernetes versions \< 1.30

#### Rejected: Deploying a new ValidatingAdmissionPolicy (VAP) in virt-operator

This is the same solution proposal
as detailed in
Section [Deploying a new ValidatingAdmissionPolicy (VAP) in SSP](#solution-deploying-a-new-validatingadmissionpolicy-vap-in-ssp).
The difference is that in this alternative solution kubevirt/kubevirt is the responsible to deploy and keep in sync
deletion protection VAP.

Check [https://github.com/kubevirt/community/blob/main/design-proposals/shadow-node.md\#backwards-compatibility](https://github.com/kubevirt/community/blob/main/design-proposals/shadow-node.md#backwards-compatibility)
for backwards compatibility details.

A proof of concept implementation can be found
at: [https://github.com/kubevirt/kubevirt/commit/7167e68723201b3899f8f36b02e5bf11701ee722](https://github.com/kubevirt/kubevirt/commit/7167e68723201b3899f8f36b02e5bf11701ee722)

#### Alternative Solution: Virtual Machine Admission Webhook

A new validation rule is added to the VM admitters. The admitter will check for presence of
the \`[kubevirt.io/vm-delete-protection](http://kubevirt.io/vm-delete-protection)\` label, filter delete
operations, and will reject any attempt to delete VM objects if the label is present and its value is set to “true”
or “True”.

Pros:

- It works with any Kubernetes version.

Cons:

- It will increase the maintenance effort.

A proof of concept implementation can be found
at:  [https://github.com/kubevirt/kubevirt/commit/2fafcf649019c26c8802838f2bbe673a9e93f04c](https://github.com/kubevirt/kubevirt/commit/2fafcf649019c26c8802838f2bbe673a9e93f04c)

#### Discarded Alternative Solution: Finalizers

According
to [Kubernetes finalizers documentation](https://kubernetes.io/docs/concepts/overview/working-with-objects/finalizers/)
the deletion operations will be accepted, and objects will not be deleted until a particular condition is meet. However,
the goal is to reject any deletion operation preformed against a protected VMs, showing an appropriate warning message.
By using finalizers, this is not possible. Therefore, finalizers are not suitable for this particular scenario.

#### Comparison between VAP Solution and Admission Webhook

We have carried out some benchmark experiments to determine if there is any impact in terms of performance when the user
deletes VMs. In the test, we created a bunch of VMs and delete them at the same time using the command:

```bash
$ kubectl delete –all 
```

We have defined three sets of experiments changing the number of VMs created, and the running state of the VM when it is
deleted, i.e., stopped or running. The values shown here are the mean of 31 runs:

| Solution          | Deletion time stopped 400 VMs(s) | Deletion time stopped 40 VMs(s) | Deletion time running 40 VMs(s) |
|:------------------|:---------------------------------|:--------------------------------|:--------------------------------|
| Baseline          | 78.97                            | 6.03                            | 198.52                          |
| VAP               | 78.96                            | 6.23                            | 199.61                          |
| Admission Webhook | 79.06                            | 6.16                            | 199.8                           |

Please note that these tests are not aimed to compare exact deletion times between different experiment sets, but to
compare them to a baseline and check if any  implies a degradation and if there is a faster option than the other.
Differences between options are negligible in any scenario analyzed. Moreover, **no performance degradation has been
observed.**

Given the pros and cons, ** VAP deployed by SSP is preferred in this case.**

## Scalability

This change should not limit the current scalability of the system.

## Update/Rollback Compatibility

The inclusion of this feature should not create any update/rollback compatibility issue.

## Functional Testing Approach

Unit and functional test should be enough to cover this change.  
The quick functional test without starting the VM should be part of the conformance tests.

# Implementation Phases

The implementation will be done in one phase.