# Overview

With the introduction
of [KubeVirt Feature Lifecycle](https://github.com/kubevirt/community/blob/main/design-proposals/feature-lifecycle.md)
policy, features reaching General Availability (GA) need to drop their use of feature gates. This applies also to
configurable features that we may want to disable.

## Motivation

Users may want certain features to be configurable, for example to make the best use out of given
resources or for compliance reasons features may expose sensitive information from the host to the virtual machines
instances (VMI) or add additional containers to the launcher pod, which are not required by the user. 

The downward metrics feature is a good example of why some clusters may want to have it enabled or disabled.
The downward metrics feature exposes some metrics about the host node where the VMI is running to the guest. This
information may be considered sensitive information.
If there is no mechanism to disable the feature, any VMI could request the metrics and inspect information that, in some
cases, the admin would like to hide, creating a potential security issue, "need-to-know principle".

The behavior of other features might be changed by editing configurables, e.g. the maximum of CPU sockets allowed for
each VMI can be configured.

Before the introduction
of [KubeVirt Feature Lifecycle](https://github.com/kubevirt/community/blob/main/design-proposals/feature-lifecycle.md)
policy, many feature gates remained after feature's graduation to GA with the sole purpose of acting as a switch for the
feature. Generally speaking, this is a bad practice, because feature gates should be reserved for controlling a feature
until it reaches maturity. i.e., GA. Therefore, in the case that a developer wants to provide the ability to tune/change
a feature, configurables exposed in the KubeVirt CR should be provided. This should be
accomplished while achieving [eventually consistency](https://en.wikipedia.org/wiki/Eventual_consistency). This forces
us to avoid the feature configuration control checking on webhooks and moving the feature configuration control closer to the
responsible code. Moreover, it has to be decided how the system should behave if a VMI is
requiring a feature in a configuration different from what was expressed in the KubeVirt CR, or what should happen if the
configuration of a feature in use is changed. (see matrix below).

## Goals

- Get a clear understanding about the features configurations.
- Establish how the feature configurables should work.
- Describe how the system should react in these scenarios in the case that the VMI exposes an API field to configure the
  features:
    - A feature in KubeVirt is set to state A and a VMI requests the feature to be in state B.
    - A feature in KubeVirt is set to state A, there are running VMIs using the feature in state A, and the feature is 
      changed in KubeVirt to state B.
    - A feature in KubeVirt is set to state A, and pending VMIs want to use it.
    - A feature in KubeVirt is set to state A, and running VMIs using the feature in state B wants to live migrate.
- Graduate features by dropping their gates and (optionally) adding spec options for them.

## Non Goals

- Describe how features protected with features gates should work.
- Change how feature gates are managed. Feature gating and configuration are two completely distinct issues.

## Definition of Users

Development contributors.

Cluster administrators.

## User Stories

* As a cluster administrator, I want to be able to change the cluster-wide configuration of a feature by editing configurables.

* As VMI owner, I want to use a given feature.

* As a VMI owner / cluster admin, I want to understand what's the current configuration of the various features.

## Repos

Kubevirt/Kubevirt

# Design

Ideally, a graduated feature would just work out the box, with no further complexity to the cluster admin.
Features that must be configured must add new fields to the KubeVirt CR under `spec`:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
[...]
spec:
  certificateRotateStrategy: {}
  feature-A:  {}
  feature-C:
    configA: integer
    configB: string
[...]
```

The VMI object may or may not include a configuration field inside the relevant spec.

> **NOTE:** The inclusion of these new KubeVirt API fields should be carefully considered and justified. The feature
> configurables should be avoided as much as possible.


Current feature gates will require an evaluation to determine if they need to be dropped or graduated to a configurable.
This is current list of GA'd features present in KubeVirt/KubeVirt which are still using feature gates and are shown as
[configurables in HCO](https://github.com/kubevirt/hyperconverged-cluster-operator/blob/main/controllers/operands/kubevirt.go#L166-L174):

- DownwardMetrics
- Root (not sure about this one)
- DisableMDEVConfiguration
- PersistentReservation
- AutoResourceLimitsGate
- AlignCPUs

This is the current list of GA'd features present in KubeVirt/KubeVirt which are still using feature gates and are [always
enabled by HCO](https://github.com/kubevirt/hyperconverged-cluster-operator/blob/main/controllers/operands/kubevirt.go#L125-L142):

- CPUManager
- Snapshot
- HotplugVolumes
- GPU
- HostDevices
- NUMA
- VMExport
- DisableCustomSELinuxPolicy
- KubevirtSeccompProfile
- HotplugNICs
- VMPersistentState
- NetworkBindingPlugins
- VMLiveUpdateFeatures

Please note that only feature gates included in KubeVirt/KubeVirt are listed here.

Section [Interactions with the VMIs requests](#interactions-with-the-vmis-requests) details how the system should
react to the different scenarios different to scenarios where the VMI feature configuration is different from what it is
configured in the KubeVirt CR. Also, Section [Update/Rollback Compatibility](#updaterollback-compatibility) explains how
feature gates should be graduated to configurables.

## Interactions with the VMIs requests

In case that, the VMI exposes a configuration field to request the feature as well as the KubeVirt CRD, the system may
encounter some inconsistent states that should be handled in the following way:

- If the feature is set to state A in the KubeVirt CR and the VMI is requesting the feature in state B, the VMIs must
  stay in `Pending` state. The VMI status should be updated, showing a status message, highlighting the reason(s) for the
  `Pending` state. Moreover, an event could be triggered. For instance, in the following KubeVirt CR, `feature-B` is not
  enabled:

```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
[...]
spec:
  certificateRotateStrategy: {}
  feature-A:  {}
```
but a given VMI is requesting it:

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
metadata:
  name: vmi-feature-b
spec:
  domain:
    feature-B: {}
[...]
```
Therefore, the VMI PHASE should stay in `Pending` until `feature-B` is enabled in KubeVirt CR:

```bash
$ kubectl get vmis
NAME           AGE    PHASE     IP    NODENAME   READY
vmi-feature-b   2s    Pending                    False
```
Moreover, the VMI status should reflect the specific feature configuration that is preventing VMI to start:
```bash
$ kubectl get vmis vmi-feature-b
[...]
status:
  conditions:
  - lastProbeTime: "2024-08-28T10:16:57Z"
    lastTransitionTime: "2024-08-28T10:16:57Z"
    message: virtual machine is requesting the disabled feature: feature-B 
    reason: FeatureNotEnabled
    status: "False"
    type: Synchronized
```

and a warning event is triggered:

```event
LAST SEEN   TYPE      REASON                    OBJECT                                 MESSAGE
[...]
2s          Warning   FeatureNotEnabled         virtualmachineinstance/vmi-feature-b   feature-B feature not enabled
```

- Feature configuration checks that could prevent a VMI from starting should only be performed during the VMI
  reconciliation process, and not at runtime if the changes cannot be applied without restarting the VMI. While this
  approach ensures that the system does not actively block, stop, or kill running VMIs due to configuration changes in
  the KubeVirt CR, it is important to note that VMIs may still experience issues or termination if critical features
  become unavailable or incompatible. 
- The system should not block live migration unless the requested feature
  is not supported in the destination host. However, as stated before, if the changes can be applied without 
  restarting VMI, it can be done at runtime.
- Updates to KubeVirt CR to update a feature configuration should not be rejected.

## Scalability

The feature configurables should not affect in a meaningful way the cluster resource usage.

## Update/Rollback Compatibility

The feature configurables should not affect forward or backward compatibility once the feature GA. A given feature,
after 3 releases in Beta, all feature gates must be dropped. Those features that need a configurable should define it ahead
of time.

## Functional Testing Approach

The unit and functional testing frameworks should cover the relevant scenarios for each feature.

# Implementation Phases

The feature configuration checks should be placed in the VMI reconciliation loop. In this way, the feature configuration
evaluation is close to the VMI scheduling process, as well as allowing KubeVirt to reconcile itself if it is out of sync
temporally.

Regarding already existing features transitioning from feature gates as a way to enable/disable a feature to configurable
fields, this change is acceptable, but it should be marked as a breaking change and documented. Moreover, all feature
gates should be evaluated to determine if they need to be dropped and transitioned to configurables.

## About implementing the checking logic in the VM controller

KubeVirt should not allow starting a VM if it is requesting a feature that it is not available in the cluster.
The VM controller must report the reasons in the `status` field of the VM.

Optionally, another check in the VM controller could be added to let the user know if a VM has requested a feature
configuration which is different from what it is specified in the KubeVirt CR. This check would be performed when the
user creates the VM, and it should update the `status` field of the VM. 
