# Overview

The [`kubevirt/common-instancetypes`](https://github.com/kubevirt/common-instancetypes)
project provides a set of useful [instance types and preferences](https://kubevirt.io/user-guide/virtual_machines/instancetypes/) for
use when creating
[`VirtualMachines`](http://kubevirt.io/api-reference/main/definitions.html#_v1_virtualmachine)
with KubeVirt. Until now operators and users could manually deploy these
resources from the [upstream project
tree](https://github.com/kubevirt/common-instancetypes#installation), from a
[specific release](https://github.com/kubevirt/common-instancetypes/releases/tag/v0.3.1)
or by using the [SSP operator](https://github.com/kubevirt/ssp-operator#ssp-operator).

This proposal suggests moving the deployment of the cluster-wide
`kubevirt/common-instancetypes` resources to `virt-operator` so that they are
consistently available across KubeVirt deployments.

## Motivation

The ultimate goal of this proposal is to improve KubeVirt's default user
experience when creating `VirtualMachines` across several different deployments.

The deployment of the cluster-wide `kubevirt/common-instancetypes` resources by
default will allow users to rely more freely on the presence of the resources
and make use of them during their `VirtualMachine` creation flows.

For example, thanks to recent improvements within `virtctl` the following
command would deploy a simple Fedora `VirtualMachine` work across several
different deployments:

```sh
$ virtctl create vm \
  --instancetype n1.medium \
  --preference fedora \
  --volume-containerdisk name:fedora,src:quay.io/containerdisks/fedora:latest \
  --name fedora | kubectl apply -f -
```

## Goals

* For new deployments, the deployment of the cluster-wide
  `kubevirt/common-instancetypes` resources by `virt-operator`

* For upgraded deployments, to have `virt-operator` claim ownership of any
  existing cluster-wide `kubevirt/common-instancetypes` resources deployed by
  the SSP operator.

## Non Goals

* For upgraded deployments, `virt-operator` will not try to claim ownership of
  any operator deployed cluster-wide `kubevirt/common-instancetypes` resources
  other than by SSP.

## User Stories

* As a user I want access to a consistent default set of cluster-wide instance
  types and preferences across all deployments I interact with

* As an operator I want my users to have access to a consistent set of
  cluster-wide instance types and preferences without the need for additional
  deployment of operators or resources

## Repos

* kubevirt/kubevirt
* kubevirt/common-instancetypes
* kubevirt/ssp-operator

# Design

`virt-operator` will be extended to deploy bundles of cluster-wide resources by
default from the `kubevirt/common-instancetypes` project.

These bundles will be injected into the `virt-operator` container image at build
time with the `kubevirt/common-instancetypes` project referenced as a submodule
checked out from a specific release tag from within `kubevirt/kubevirt`.

The submodule approach keeps much of the maintenance and testing cruft of the
`kubevirt/common-instancetypes` project out of the already congested
`kubevirt/kubevirt` project. It also allows us to retain the existing
`kustomize` project structure of `kubevirt/common-instancetypes` making the repo
useful for end operators and users looking to make modifications to or to only
generate a subset of the resources.

The `KubeVirt` `CR` will be extended to include a `deployInstancetypes`
toggle, enabled by default, to allow operators to disable this behaviour.

An additional toggle will also be introduced into the
`kubevirt/hyperconverged-cluster-operator` project to control this behaviour
when deploying `kubevirt/kubevirt`.

The `kubevirt/ssp-operator` project will also be modified to ignore
`kubevirt/common-instancetypes` resources deployed by the `virt-operator` in
version `v0.18.N` and later to stop deploying `kubevirt/common-instancetypes`
resources at all by default with version(s) >=`v0.19.0`.

## Scalability

The `kubevirt/common-instancetypes` project currently ships instance type and
preference cluster-wide resources. Their deployment by default should not
introduce any scalability issues.

## Update/Rollback Compatibility

For upgraded KubeVirt deployments `virt-operator` will attempt to take ownership
of existing `kubevirt/common-instancetypes` cluster-wide resources previously
deployed by the SSP operator. The SSP operator itself will need to be updated at
the same time to ensure it relinquishes ownership and no longer attempts to
reconcile the resources.

Rolling both `virt-operator` and SSP operator back to their previous versions
should return control of these resources to the SSP operator.

## Functional Testing Approach

* For new deployments, functional tests will be introduced asserting that the
  resources are deployed correctly.

* For upgraded deployments, functional tests will be introduced asserting that:
  * Operator defined resources are left alone by `virt-operator`
  * SSP operator deployed resources are claimed by `virt-operator`

# Implementation Phases

* Introduce `kubevirt/common-instancetypes` as a submodule of `kubevirt/kubevirt`
* Deploy `kubevirt/common-instancetypes` cluster resources in fresh deployments
* Claim ownership of existing SSP operator deployed
  `kubevirt/common-instancetypes` cluster resources on upgrade
* Add the ability to disable the deployment of `kubevirt/common-instancetypes`
  cluster resources from the `KV` `CR`
* Add the ability to disable the deployment of `kubevirt/common-instancetypes`
  cluster resources from the `HCO` `CR`
* Disable the deployment of `kubevirt/common-instancetypes`
  cluster resources from the `kubevirt/ssp-operator` project

# Alternatives

* Creation of a standalone `instancetype-operator` using
  [`kubebuilder-declarative-pattern`](https://github.com/kubernetes-sigs/kubebuilder-declarative-pattern)
  to deploy the `kubevirt/common-instancetypes` cluster-wide resources.
