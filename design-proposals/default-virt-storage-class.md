# Overview
Storage providers in the field may supply many ways of utilizing a back end storage solution;  
They do so by providing different Kubernetes StorageClass parameters.  
Some examples are `csi.storage.k8s.io/fstype` which indicates which filesystem type is created on a raw block before being passed on to a pod, or even lower level kernel [mapping options](https://access.redhat.com/articles/6978371).  
With the growth of kubevirt adoptions, we may sometimes see a certain combination of storage class parameters is preferrable for VM workloads.

This design presents a possible way of steering kubevirt users
towards using such identified storage class.

## Motivation
- Certain Microsoft Windows versions use a peculiar I/O pattern. This particular pattern may cause [CRC errors](https://github.com/kubevirt/kubevirt/pull/9741) on certain types of storage back ends (ceph) when data is written and read across different processes or threads on a Windows VM.  
- Storage providers like Portworx provide more than 10 storage class variations out of the box. It's possible that a specific variation is preferrable for VM workloads.

## Goals
* Provide a method for preferring a certain storage class in the cluster specifically for VM workloads

## Non Goals
* Make the preferrable storage class variation a part of CDI knowledge via [StorageProfiles](https://github.com/kubevirt/containerized-data-importer/blob/main/doc/storageprofile.md).  
This will be a cluster admin/storage provider decision.

## Definition of Users
This feature is intended for cluster admins/storage providers who wish to advertise a specific set of parameters for provisioning VM disks.

## User Stories
* As a KubeVirt user, I want my storage to be tuned properly to run virtual machines and I want to consume this tuning automatically or with as little manual intervention as possible.

## Repos
* **containerized-data-importer**: Various controllers related to PVC population
* **kubevirt**: storage class preference adjusted as well, if it exists

# Design
Introduce a new storage class annotation `storageclass.cdi.kubevirt.io/is-default-class`  
which marks a storage class as the default for virtualization workloads (`contentType=kubevirt`),  
and thus is preferred over the k8s cluster [default storage class](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/) `storageclass.kubernetes.io/is-default-class` for those cases.


## API Examples
(tangible API examples used for discussion)

## Scalability
It is assumed scalability was one of the choosing factors when identifying a storage class to be superior for VM workloads

## Update/Rollback Compatibility
* Existing populated PVCs will not be affected, however,  
clone sources from previous versions may have been provisioned using a different storage class variation of a provisioner.
* Storage classes can still be explicitly requested on manifests to override virt/k8s default storage class

## Functional Testing Approach
* DV created without explicit SC  
&emsp;virt storage class takes precedence over a default sc  
&emsp;virt storage class is used even if there is no default sc
* DV created with explicit SC  
&emsp;overrides virt storage class
* Multiple virt storage classes  
&emsp;Follows deterministic [k8s behavior](https://github.com/kubernetes/kubernetes/pull/110559);  
&emsp;&emsp;Primary sort by creation timestamp, newest first  
&emsp;&emsp;Secondary sort by class name, ascending order
* Clone from "standard" storage class variation to "virt" variation
