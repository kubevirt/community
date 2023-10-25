# Overview

Currently, KubeVirt supports both rootful and non-root VMs.
Since KubeVirt `v0.59.0`, non-root VMs are the default
implementation ([kubevirt/kubevirt#8563](https://github.com/kubevirt/kubevirt/pull/8563)).

The `Root` feature gate works both as a feature gate **and** as a global configuration.
When the `Root` feature gate is enabled:

1. All newly created VMs are rootful.
2. Non-root VMs are converted to rootful after migration.

When the `Root` feature gate is disabled:
1. All newly created VMs are non-root.
2. Rootful VMs are converted to non-root after migration.

There are five rootful CI lanes:

Periodic lanes:

- periodic-kubevirt-e2e-k8s-1.26-sig-storage-root
- periodic-kubevirt-e2e-k8s-1.26-sig-compute-root
- periodic-kubevirt-e2e-k8s-1.26-sig-operator-root

Pre-submit **optional** and non-required lanes:

- pull-kubevirt-e2e-k8s-1.26-sig-storage-root
- pull-kubevirt-e2e-k8s-1.26-sig-compute-root

## Motivation

Rootful VMs are less secure, pose a maintenance burden and are not tested when PRs are submitted
since [kubevirt/project-infra#2646](https://github.com/kubevirt/project-infra/pull/2646) (`v0.59.0`).

1. Remove multiple code paths that are spread across the code base.
2. Reduce the number of jobs running in the CI and the need to maintain them.

## Goals

1. Remove support for the creation of rootful VMs.
2. Remove support for the conversion of non-root VMs to rootful via migration.
3. Remove support for the conversion of rootful VMs to non-root via migration.

## Non Goals

N/A.

## Definition of Users

KubeVirt users using rootful VMs.

## User Stories

(list of user stories this design aims to solve)

## Repos

1. kubevirt/kubevirt
2. kubevirt/project-infra

# Design

1. Remove the `NonRootExperimental` feature gate.
2. Clean the e2e tests from using the `NonRoot` feature gate.
3. Deprecate rootful VMs in `v1.3.0`.

In `v1.4.0`:
1. Remove non-root to root conversion via migration.
2. Remove rootful VMs code paths from e2e tests.
3. VMI mutator: always create VMs as non-root.
4. Remove the "Root" feature gate.
5. Remove rootful CI lanes.
6. virt-launcher: remove `runWithNonRoot` argument (and the matching code in virt-controller).
7. Clean up the code from rootful related VMs.
8. Consider the deprecation / removal of the `runtimeUser` field in `VirtualMachineInstanceStatus`.

## API Examples

(tangible API examples used for discussion)

## Scalability

N/A.

## Update/Rollback Compatibility

Starting from KubeVirt version TBD, upgrade will be blocked in case rootful VMs exist.
An alternative could be to automatically migrate rootful VMs to non-root before the actual upgrade takes place.

## Functional Testing Approach

Currently, rootful and non-root tests are not separated, and are controlled by skips
in the code, or by ginkgo labels.

Reduce the number of skips dependent on the presence or absence of feature gates.

# Implementation Phases

