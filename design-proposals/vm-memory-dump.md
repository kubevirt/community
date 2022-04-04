# Overview
Proposal for taking VM memory dump for analysis purposes.

## Motivation
One of the methods of troubleshooting guest virtual machines is getting a memory dump and analyzing it. As this is a standard capability in most VM management systems, it is also required for kubevirt’s VMs.

## Goals
Have a mechanism to get a VM memory dump to be inspected with [Volatility3](https://github.com/volatilityfoundation/volatility3).

## Non Goals
Memory dump is not a restorable output and currently will not be a part of the VM snapshot.

## User Stories
As a Kubevirt user I would like to get a memory dump of a running VM so that I can later inspect its memory.

## Repos
[KubeVirt](https://github.com/kubevirt/kubevirt)

# Design
To trigger a memory dump a virtctl memoryDump command will be added.
The command will either use an existing PVC or create a new PVC on demand. This PVC will be bound to the VM and will appear in the VM volumes spec with a fitting status.
During the memory dump process this PVC will be mounted to the virt-launcher pod. After the mounting the guest memory will be dumped to that pvc and eventually it will be unmounted from the virt launcher.
The PVC will remain bound - in the VM spec. It is the users' responsibility to export/use the memory dump results before reusing the PVC for a new memory dump. As long as the PVC is bound as a memory dump "container" each memory dump command will overwrite the previous content.
It will be possible to track the last memory dump in the VM status with a new memory dump status that will be added there.
It will be possible to unbound the PVC and then do a memory dump to another PVC that will be bound instead.

## API

### Get memory dump
#### trigger memory dump
Run a virtctl command to get a memory dump. This new API will look as follows:

`$ virtctl memory-dump my-vm --volume-name=memory-pvc`

In this case `memory-pvc` should already exist and be of a size big enough to contain the memory dump. A check will be made to make sure.
It will be possible to ask for a PVC to be created with `--create` flag. In such case the required size will be calculated.

##### The process
The trigger of memory dump will call a VM subresource endpoint which will look like this:
`/apis/subresources.kubevirt.io/v1alpha3/namespaces/<vm-namespace>/virtualmachines/<vm-name>/memorydump`

The VM subresource endpoint will recieve a rest request containing the pvc name the user want to dump the memory to.
It will patch the vm status with a memoryDumpStatus stating the volume-name and Phase `Binding`.
The memoryDumpStatus will look something like this:

```golang
// MemoryDumpStatus represent the memory dump request status and info
type MemoryDumpStatus struct {
	// ClaimName is the name of the pvc that will contain the memory dump
	ClaimeName string `json:"claimName"`
	// Phase represents the memory dump phase
	Phase MemoryDumpPhase `json:"phase,omitempty"`
	// TimeStamp represents the time the memory dump was completed
	TimeStamp *metav1.Time `json:"timestamp,omitempty"`
}

type MemoryDumpPhase string

const (
	// The memorydump is during pvc binding
	MemoryDumpBinding MemoryDumpPhase = "Binding"
	// The memorydump is in progress
	MemoryDumpInProgress MemoryDumpPhase = "InProgress"
	// The memorydump is completed
	MemoryDumpCompleted MemoryDumpPhase = "Completed"
	// The memorydump is being unbound
	MemoryDumpUnBinding MemoryDumpPhase = "Unbinding"
)
```

In the virt controller once getting the update of the memoryDump in state of binding it will bind the pvc to the vm by updating the vm volumes with the memoryDump pvc.
After that the status of the memoryDump will be updated to `InProgress` which will be cause the volume to be mounted to the virt-launcher, same as done in the hotplug process, other then the end of the process that instead of attaching the volume to the VMI it will trigger the guest memory dump `virDomainCoreDump` command with `VIR_DUMP_MEMORY_ONLY` flag to be executed by virt-launcher.
(Check out [virDomainCoreDump](https://libvirt.org/html/libvirt-libvirt-domain.html#virDomainCoreDump) for reference)
After the dump is complete the timestamp and status of the memory dump will be updated and it will unmount the pvc from virt-launcher.
The PVC will remain bound to the VM - will be kept in the VM volumes and have the updated status and timestamp in the VM status as mentioned.
It will look something like that: (In the VMI there will be no information)

```yaml
---
apiVersion: v1
items:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachine
  name: my-vm
  ....
  Snip
  ....
    template:
      spec:
        domain:
          devices:
            disks:
            - disk:
                bus: virtio
              name: datavolumedisk
            - disk:
                bus: virtio
              name: cloudinitdisk
        volumes:
        - dataVolume:
            name: my-dv
          name: datavolumedisk
        - cloudInitNoCloud:
            userData: |-
              #cloud-config
              password: fedora
              chpasswd: { expire: False }
          name: cloudinitdisk
        - memoryDump:
            persistentVolumeClaim:
                claimName: memory-dump-pvc
          name: memory-dump
  status:
    ....
    Snip
    ....
    memoryDumpStatus:
      claimName: memory-dump
      phase: Completed
      timestamp: "2022-03-29T11:00:04Z"
```


#### remove memory dump from VM
Run a virtctl command to remove a memory dump from a VM. This new API will look as follows:

`$ virtctl remove-memory-dump my-vm`

##### The process
The trigger of remove memory dump will call a VM subresource that will add a removeMemorydumpRequest to the VM status.
The request will un the PVC from the vm. The PVC will be deleted from the VM spec and the memoryDumpStatus will be removed.

### Handle the memory dump
After the memory dump is completed the user will be able to export this PVC and also if there is a snapshot supported storage class the PVC can be a part of the VM snapshot and then it can be unbound and deleted.
The output of the memory dump can be used for memory analysis with different tools for example [Volatility3](https://github.com/volatilityfoundation/volatility3) and maybe also [sleuthkit](https://www.sleuthkit.org/autopsy/)


## Update/Rollback Compatibility
New API should not affect updates / rollbacks.

# Implementation Phases
The implementation wil be split to several phases:
* Add memory dump command in virt launcher server
* Add virtctl command to execute memory-dump with existing pvc
* Add remove memory dump virtctl command
* Extend the virtctl command to support on-demand creation of PVC
