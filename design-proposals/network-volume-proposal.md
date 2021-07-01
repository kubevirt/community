# Overview
The purpose of this proposal is to introduce a new volume source known as NetworkVolumeSource. NetworkVolumeSource will be able to access network endpoints, such as NBD. This will allow for users to access images over a network.

## Motivation
The motivation for this proposal is to allow for the use of cdroms from a network source. This can eventually be modified to allow for the dynamic changing of cdroms.

## Goals
Provide ISO images that can be used by multiple VMs simultaneously and that can be inserted and ejected at runtime.

## Non Goals
* There is no need for the ability to hotplug or hot-unplug storage devices.
* There is no need for the ability to dynamically change cdroms.

## Definition of Users
* VM owner
* KubeVirt administrator

## User Stories
* As a VM owner I want to boot multiple VMs from an ISO image. I want to import the ISO image content into my cluster only once.
* As a KubeVirt administrator I want to easily package ISO images and make them available in the cluster so that my end users can access them easily.

## Repos
* https://github.com/kubevirt/kubevirt

# Design
This design introduces a new volume source to KubeVirt known as NetworkVolumeSource.  NetworkVolumeSource takes advantage of libvirt's ability to connect network endpoints to utilize them as a source for a cdrom, providing readonly access to the endpoint. However, access should still be configurable to allow the volume to be utilized for other disk types. To accomplish this, NetworkVolumeSource requires the following fields:
  * URI:
    * The URI necessary to connect to the endpoint. An optional port can be specified if the endpoint's port is not the default or if the default port for a particular service is unkown. 
  * Readonly:
    * An optional field which defaults to true. If not specified, the source will be attached as readonly.
  * Format:
    * Accepts "raw" or "qcow2."
  * SecretRef:
    * An optional field which can contain authorization information.
  * CertConfigMap:
    * An optional field which can reference a map containing any public keys necessary to connect to the endpoint. In the case of NBD endpoints, this field can contain public TLS information.
    
These fields should provide the necessary information to construct a libvirt disk xml which can connect to the network endpoint. The initial implementation is primarily concerned with utilizing an NBD endpoint as a cdrom disk. In this example, the endpoint is a service known as "diskservice" which exports several images, does not have authorization requirements, and has TLS disabled. The service, “diskservice”, will have multiple running instances. In the event that one crashes, the VM will still have the ability to connect. The following is a fragment of the resulting libvirt xml:

```XML
<disk type='network' device='cdrom'>
  <driver name='qemu' type='raw' cache='none' error_policy='stop'/>
  <source protocol='nbd' name='Fedora-Workstation-Live-x86_64-34-1.2.iso' tls='no' index='1'>
    <host name='diskservice' port='10809'/>
  </source>
  <target dev='sda' bus='sata'/>
  <readonly/>
  <alias name='ua-cdromnetworkdisk'/>
  <address type='drive' controller='0' bus='0' target='0' unit='0'/>
</disk>
```

This implementation also includes all the required code to translate the KubeVirt API NetworkVolumeSource into the appropriate libvirt XML. For an example of the API that leads to this domain xml, see the API section of this document.

## API Examples
* Example from schema.go
  * Creates NetworkVolumeSource with fields for URI and Format.
```go
type NetworkVolumeSource struct {
	// Will force the ReadOnly setting in VolumeMounts.
	// Default false.
	// +optional
	ReadOnly bool `json:"readOnly,omitempty"`
	//URI represents the URI of the network volume
	Uri string `json:"uri"`
	//Format represents the format of the network volume
	Format string `json:"format"`
}
```
* Example from service.yaml
  * Follows the format created in schema.go with the URI and Format specified within the yaml.
```yaml
- networkVolume:
    uri: nbd://diskservice/Fedora-Workstation-Live-x86_64-34-1.2.iso
    format: raw
  name: networkdisk
```

## Scalability
N/A

## Update/Rollback Compatibility
N/A

## Functional Testing Approach
Single VM RAW
 * Start disk-service with a raw disk
 * Create a VM that uses a CDROM with a network volume that points to the disk service.
 * Boot VM
 * Log into VM and attempt to mount cdrom
 * Verify mount and that data is available
 
Single VM QCOW2
 * Start disk-service with a qcow2 disk
 * Create a VM that uses a CDROM with a network volume that points to the disk service.
 * Boot VM
 * Log into VM and attempt to mount cdrom
 * Verify mount and that data is available
 
Multiple VMs RAW
 * Start disk-service with a raw disk
 * Create multiple VMs that use a CDROM with a network volume that points to the disk service.
   * Point at the same image.
 * Boot each VM
 * Log into each VM and attempt to mount cdrom
 * Verify mount and that data is available for each VM
 
Multiple VMs QCOW2
 * Start disk-service with a qcow2 disk
 * Create multiple VMs that use a CDROM with a network volume that points to the disk service.
   * Point at the same image.
 * Boot each VM
 * Log into each VM and attempt to mount cdrom
 * Verify mount and that data is available for each VM
 
Multiple VMs different images
 * Start disk-service with multiple disks being served (raw/qcow2)
 * Create multiple VMs that use a CDROM with a network volume that points to the disk service.
   * Point at different images.
 * Boot each VM
 * Log into each VM and attempt to mount cdrom

# Implementation Phases
Phase One:
 * Implement API and conversion logic.
 
Phase Two:
 * Add authentication and certificate management logic.
 * Add tests to make sure authentication and encryption is functional.
