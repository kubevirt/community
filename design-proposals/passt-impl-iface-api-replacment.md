## Overview
Following the emerging [Kubevirt Network Binding Plugin API](https://docs.google.com/document/d/13fLaqZuM8El_CT5XzxPv-1epKsB1-Spv_jtXveEkkdQ/edit#heading=h.z27rudfpwy44)
and introduction of Passt network binding plugin [2](https://github.com/kubevirt/kubevirt/pull/10425) [3](https://github.com/kubevirt/kubevirt/pull/10621), 
it's now possible to remove Passt network binding implementation from Kubevirt core.

This proposal outlines the replacement plan for the existing Passt implementation and API.

# Motivation
Removing Passt network binding implementation from Kubevirt core reduces the maintenance burden.
Having Passt binding code outside Kubevirt's core may ease contribution process as there is no need to be familiar 
with [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt) core code.

## Goals
- Reduce maintenance of Passt network binding in Kubevirt core.
  
## Non Goals

## User Stories

## Repos
[kubevirt/kubevirt](https://github.com/kubevirt/kubevirt):
- [Passt binding plugin hook sidecar](https://github.com/kubevirt/kubevirt/tree/main/cmd/sidecars/network-passt-binding)
- [Passt binding plugin CNI](https://github.com/kubevirt/kubevirt/tree/main/cmd/cniplugins/passt-binding)

# Design
Replace Kubevirt core Passt network binding implementation and interface API with Passt binding plugin.
Users should be aware of the coming changes and be prepared.

The proposed plan for replacing Kubevirt core Passt binding implementation and API is as follows:

1. Announce the coming changes for Passt binding described in this section at kubevirt-dev mailing list.
   This will give users some time to prepare for changes in Kubevirt next release v1.2.0. 

2. Kubevirt next release v1.2.0 should have the following changes:
   1. Raise warning when Passt feature gate is enabled saying its deprecated.
   2. Raise warning when Passt interface API is being used saying its deprecated and Passt binding plugin should be used instead.
   3. Mark relevant filed as deprecated.

   VM that use Passt interface API will work regularly.
 
3. Kubevirt next release v1.3.0 should have the following changes:
   1. [Passt interface API](http://kubevirt.io/api-reference/main/definitions.html#_v1_interfacepasst) 
      usage will be blocked by the admission webhook.
      Creating new VMs that use Passt interface API will be blocked.
      Existing VMs that use Passt interface API won't manage to re-start.
    
      Users who like to use Passt binding will have to follow the network binding plugin API:
      1. Install Passt binding CNI (make sure the binary available on all nodes).
      2. Create and register Passt binding network-attachment-definition.
      3. Register Passt binding sidecar image.
         
   2. Remove Passt binding implementation from Kubevirt core.
   
   3. Deprecate Passt feature gate.
      Set Passt feature gate to always false as a discontinued feature, as it will be replaced by a network binding plugin.

> **Note:**
> Passt binary remain part of virt-launcher compute container, as it executed by libvirt it should be aligned with its version.
> Decoupling Passt binary from the compute container may require additional changes to enable libvirt process
> access passt binary.

## API Examples
### VM with Passt interface using the retiring Passt interface API:
1. Enable Passt feature:
```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - Passt
  ...
```

2. Specify Passt interface in the VM spec template:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: myvm
spec:
  template:
    spec:
      domain:
        devices:
          interfaces:
          - name: blue
            passt: {}
      networks:
      - name: blue
        pod: {}
  ...
```

### VM with Passt interface using network binding API:

1. Passt network binding plugin network-attachment-definition should be created, see the following example:
```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: kubevirt-passt-binding
spec:
  config: '{
            "cniVersion": "0.3.1",
            "type": "kubevirt-passt-binding"
  }'
```

2. Register Passt network binding network-attachment-definition and sidecar image, by specifying them in Kubevirt configuration, and enable network binding plugin feature:
```yaml
apiVersion: kubevirt.io/v1
kind: KubeVirt
metadata:
  name: kubevirt
  namespace: kubevirt
spec:
  configuration:
    developerConfiguration:
      featureGates:
      - NetworkBindingPlugins
    network:
      Binding:
        passt:
          sidecarImage: quay.io/kubevirt/network-passt-binding:v1.0.0
          networkAttachmentDefinition: kubevirt-passt-binding
  ...
```

3. Specify the registered Passt network binding plugin in VM spec template:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: myvm
spec:
  template:
    spec:
      domain:
        devices:
          interfaces:
          - name: blue
            binding:
              name: passt
      networks:
      - name: blue
        pod: {}
  ...
```


## Scalability

## Update/Rollback Compatibility
In Kubevirt v1.3.0: 
- Existing running VMs that use Passt interface API will keep running with no disruption.
- Stopped VMs or new ones, will not manage to start and require using the network binding plugin API.
> **Note:** Passt doesnt support migration.

## Functional Testing Approach
Kubevirt's Passt network e2e tests shall spin up VMs with Passt interface using network binding plugin API.

## Implementation Phases
1. Kubevirt v1.2.0: 
   1. Raise warnings when Passt feature is enabled and Passt interface API is used.
   2. Declare relevant filed deprecated.

2. Kubevirt v1.3.0:
   1. Set admission webhook to block VM creation when Passt interface is being used.
   2. Remove Passt binding implementation from Kubevirt core
   3. Deprecate Passt feature gate, i.e.: always enabled, no-op.

# Open questions
## 1. Can we set the webhook to raise warnings at Kubevirt v1.1.1?
   Kubevirt core Passt binding implementation and Passt network binding plugin already live along each other.
   We need to have Kubevirt raise warnings when Passt feature gate is enabled and when Passt interface API is being used.

   By moving these changes to v1.1.1, changes targeted to v1.3.0 can be done at v1.2.0,
   and eventually shorten the time to complete Passt implementation replacement.