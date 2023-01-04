## Introduction

A more indepth proposal can be found at the following google doc https://docs.google.com/document/d/1KYc3lsHHhri0gbCZTOGe5lz2dNJogRRwrXwN9Do9kT0/edit?usp=sharing

### Background & Recent Discussions
As it stands, Kubevirt has provided support for both X86 and ARM-based clusters but the assumption is that the architecture of the cluster is homogeneous. 

This means that all components are compiled for the architecture where they will be running. The codebase uses the runtime package to determine the compiled architecture for applicable logic branches in the code. 

For users running a mixture of VM workloads, the only currently available option is to introduce multiple clusters. The complexities associated with federating multiple clusters in unison can serve as a deterrent to the usefulness of KubeVirt in these types of use cases.

### Purpose 
This document aims to explain the justification for any changes required to KubeVirt in order to provide the ability for users to run in a multi-architecture environment.  In addition, this document aims to provide an open forum to discuss whether the proposed changes are the most effective way of providing the desired functionality.

## Motivation

The motivation for adding the proposed functionality stemmed from a personal use case that required running X86 and ARM-based workloads. After investigation, the conclusion reached was that no such functionality was currently available in KubeVirt. This is largely due to the lack of users working with ARM-based workloads. 

Regardless of this fact, it was identified that the framework already has accomplished most of the heavy lifting by providing support for each of the architecture exclusively. Given this, the amount of effort required to expand the existing logic was identified to be minimal and a POC was created. 

## Goals

- Support the deployment of Kubevirt into multi-architecture clusters
- Ability to specify architecture when deploying VM/VMIs 
- Feature changes should not break any existing functionality in KubeVirt
- Multi-architecture support should be placed behind a feature-gate and be an optional config item based on user-need

## User Stories

As a user, I want to be able to run an ARM compute-plane with an x86 control-plane and vice versa.

## Repos
 - https://github.com/kubevirt/kubevirt

# Design

## Proposed Code Changes

*In addition to the following code changes, all KubeVirt components were compiled as multi-architecture images*

**Kubevirt API Spec**

- Addition of an Architecture field to the VirtualMachineInstanceSpec will allow the user to indicate the architecture of the workload being deployed. If not specified, the default from the cluster config will be applied.
- Addition of ArchitectureConfiguration field to the KubevirtConfiguration. The ArchitectureConfiguration allows for setting defaults for the existing arch-specific config items. These items are `OVMFPath`, `EmulatedMachines`, and `MachineType`. Additionally, user can specify the `defaultArchitecture` to be applied. Currently if you do not set the defaultArchitecture, it will default to runtime.GOARCH. The same goes for all the arch specific defaults (OVMF,Emulated Machines, etc), they will default to the hardcoded X86,ARM defaults that already existed  as consts
  

**Virt-api**

- VMI create admitter will now validate the architecture field provided in the VMI spec to determine if the multi-architecture feature gate has been enabled
- VMI create admitter will retrieve the supported machine types using the added architecture field of the VMI spec
- Validation of ARM64 specific settings will now rely on the architecture field of the VMI spec
- VMI mutator’s setting of ARM64 specific settings will now rely on the architecture field of the VMI spec
- VMI mutator will set the default architecture based on the cluster config item
- VMI mutator will retrieve machine type based on architecture

**Virt-controller**

- Topology Hinter is no longer initialized with an architecture. Any methods that rely on architecture information will now be supplied from the VMI spec
- Retrieval of the ovmfPath now relies on the architecture field of VMI spec instead of runtime.GOARCH

**Virt-config**

- Addition of MultiArchitecture feature gate 
- Addition of defaults for each of the added KubevirtConfiguration fields
 
**Virt-operator**

- The addition of node-labeler initContainer for virt-handler daemonsets no longer performs architecture checks. The logic for determining whether the initContainer’s node-labeller.sh script runs has now been moved to the script itself.

**Virt-launcher**

- Given that the virt-launcher pod runs on a per VM basis, relying on runtime.GOARCH is an acceptable

**Virt-handler**

- Given that the virt-handler runs as a daemonset, relying on runtime.GOARCH is acceptable


## API Examples

**Example VM Spec**

```yaml
apiVersion: kubevirt.io/v1alpha3
kind: VirtualMachine
metadata:
  name: mytestvm
spec:
  running: true
  template:
    spec:
      architecture: arm64
      domain:
        devices:
          disks:
          - name: disk1
            disk:
              bus: virtio
          interfaces:
          - name: network1
            bridge: {}
      networks:
              - name: network1
                multus:
                  networkName: mytestvm/mytestvm
      volumes:
           - name: disk1
             emptyDisk:
               capacity: 50G
```

**Example Kubevirt Configuration**

```yaml	
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  certificateRotateStrategy: {}
  configuration:
    developerConfiguration:
      featureGates: ["MultiArchitecture"]
    architectureConfiguration:
      arm64:
        emulatedMachines: ["virt"]
        ovmfPath: "/usr/share/AAVMF"
      amd64:
        emulatedMachines: ["pc-*"]
        machineType: "q35"
      defaultArchitecture: "amd64"
  customizeComponents: {}
  imagePullPolicy: IfNotPresent
```

## Scalability

The proposed changes have no anticipated impact on scalability capabilities of the KubeVirt framework


## Update/Rollback Compatibility

The only issue as far as rollback compatibility goes would be around the KubeVirt components within k8s no longer being compiled for multi-architecture environments. If a user were to downgrade their deployment of KubeVirt while still running a multi-architecture cluster, there may be deployment issues for the core components but this would be an expected outcome. 

## Functional Testing Approach

Currently, the only area of this proposal that still requires clarity is what direction the KubeVirt maintainers would like to proceed with regarding functional testing. The existing cluster that is stood up as part of the functional tests leverages a helper script from the k8s repo that checks system architecture prior to bringing up a cluster. How to leverage this script but bring up a multi-architecture cluster still needs discussion.

# Implementation Phases

Proposed Implementation Rollout involves 2 stages:

1. Merge code changes. This will allow us to still test that pods are created in the e2e test for the expected architecture, despite not being able to start up.
2. Update the https://github.com/kubevirt/user-guide 
3. Followup PR involving the pushing of multi-arch images as part of the release process