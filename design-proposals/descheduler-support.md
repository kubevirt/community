# Overview
A [Descheduler](https://github.com/kubernetes-sigs/descheduler) is a Kubernetes application that causes the control plane to re-arrange the workloads in a better way.
It operates every pre-defined period and goes back to sleep after it had performed its job.

The descheduler uses the Kubernetes [eviction API](https://kubernetes.io/docs/concepts/scheduling-eviction/api-eviction/) in order to evict pods, and receives feedback from `kube-api` whether the eviction request was granted or not.

Currently, there is an issue for a descheduler to operate on `virt-launcher` pods because of the way KubeVirt handles eviction requests (see PR [kubevirt/kubevirt#11286](https://github.com/kubevirt/kubevirt/pull/11286)):

From the descheduler's point of view, `virt-launcher` pods fail to be evicted, but they actually do migrate to another node in the background.
The descheduler notes the failure to evict the `virt-launcher` pod and keeps trying to evict other pods, typically resulting in it attempting to evict substantially all pods from the node.
In other words, the way KubeVirt handles eviction requests causes the descheduler to make wrong decisions and take wrong actions that could destabilize the cluster.

This document proposes addition of functionality to KubeVirt intended to support the descheduler's operation on `virt-launcher` pods. 

## Motivation
Cluster administrators want to have a descheduler running on their cluster.
The descheduler should be able to operate on `virt-launcher` pods.

## Goals
1. Enable KubeVirt to work on existing K8s clusters that already have a descheduler deployed.
2. Enable cluster admins to introduce a descheduler to an existing cluster with KubeVirt deployed.

## Non Goals
- The Kubernetes eviction API will not be changed.
- The descheduler is expected to respect the agreement described bellow.

## Definition of Users
- Cluster administrator

## User Stories
- As a cluster admin running a K8s cluster with a descheduler, I want to deploy and use KubeVirt.
- As a cluster admin running a K8s cluster with KubeVirt, I want to deploy a descheduler that will correctly operate on my VMs.
- As a cluster admin I want to enable cluster-wide KubeVirt's descheduler support.

## Repos
1. kubevirt/kubevirt
2. kubevirt/hyperconverged-cluster-operator

# Design
## Descheduler Requirements
1. In order for the descheduler to be able to differentiate `virt-launcher` pods from other pods, all `virt-launcher` pods should be annotated with `descheduler.alpha.kubernetes.io/request-evict-only`.
2. In cases where the VM will be migrated, eviction requests on `virt-launcher` pods should return 429 HTTP code with a special reason message.
3. In order for the deschduler to understand that a `virt-launcher` pod is pending migration, it should be annotated with `descheduler.alpha.kubernetes.io/eviction-in-progress` so it will not try to evict it again.
4. In case the migration failed, the `descheduler.alpha.kubernetes.io/eviction-in-progress` annotation will be removed from the source `virt-launcer` so the descheduler will understand that it could try evicting it again.

## Feature Gate
A `DeschedulerSupport` feature gate will be introduced in order to enable this feature.

## Changes to Pod Eviction Webhook
Currently, the descheduler observes the one of the following responses:

| Eviction Strategy     | Is VMI migratable | Does Webhook approve eviction | Does PDB allow eviction | Response                         |
|-----------------------|-------------------|-------------------------------|-------------------------|----------------------------------|
| None                  | N/A               | True                          | True                    | 200 - Eviction granted           |
| LiveMigrate           | True              | True                          | False                   | 429 - Eviction blocked by PDB    |
| LiveMigrate           | False             | False                         | False                   | 429 - Eviction denied by webhook |
| LiveMigrateIfPossible | True              | True                          | False                   | 429 - Eviction blocked by PDB    |
| LiveMigrateIfPossible | False             | True                          | True                    | 200 - Eviction granted           |
| External              | N/A               | True                          | False                   | 429 - Eviction blocked by PDB    |

In the following cases instead of letting the eviction fail on the PDB, the pod eviction webhook should deny the request with reason stating that the VM will be migrated:

| Eviction Strategy     | Is VMI migratable | Response                  |
|-----------------------|-------------------|---------------------------|
| LiveMigrate           | True              | 429 - VM will be migrated |
| LiveMigrateIfPossible | True              | 429 - VM will be migrated |
| External              | N/A               | 429 - VM will be migrated |

## Changes to VMI Controller
`virt-controller` has a VMI controller.
It will:
1. Add the `descheduler.alpha.kubernetes.io/request-evict-only` to newly created pods, or to existing pods that do not have them.
2. Add the `descheduler.alpha.kubernetes.io/eviction-in-progress` annotation to the `virt-launcher` pod, if the VMI is marked for eviction.
3. Remove the `descheduler.alpha.kubernetes.io/eviction-in-progress` annotation from the `virt-launcher` pod, in case it failed to migrate.

## Scalability
This feature should not affect scalability.

## Update/Rollback Compatibility
The VMI-controller reconciles all `virt-launcher` pods, so it will annotate existing `virt-launcher` pods as well.

## Functional Testing Approach
The pod eviction webhook logic will be tested by unit tests.
The addition / removal of annotations will be tested by unit tests.

The flow will be tested e2e without a real descheduler by manually evicting a `virt-launcher` pod of a running VMI.
This should be done with all eviction strategies.

# Implementation Phases
1. Add a feature gate in KubeVirt.
2. Add the logic to pod eviction webhook.
3. Add the `descheduler.alpha.kubernetes.io/request-evict-only` annotation to newly created `virt-launcher` pods.
4. Add the `descheduler.alpha.kubernetes.io/request-evict-only` annotation to existing `virt-launcher` pods that do not have it.
5. Add the `descheduler.alpha.kubernetes.io/eviction-in-progress` annotation to `virt-launcher` pods that their controlling VMI is marked for eviction.
6. Remove the `descheduler.alpha.kubernetes.io/eviction-in-progress` annotation to `virt-launcher` pods which failed migration.
7. Add an E2E test simulating an eviction.
8. Add an enabler in HCO.
