# Overview
This proposal discusses a method to automate updating the machine type of VMs to the latest machine type version, if the VM has a machine type that is no longer compatible. Currently we can support manually changing the machine type of individual VMs, but we need an automated process that will be able to change the machine types of multiple (e.g. thousands of) VMs, limiting workload interruptions and preventing the user from having to manually update the machine type of every VM with a machine type version that is no longer supported. For example, CentOS Stream 9 will maintain compatibility with all `pc-q35-rhel8.x.x` machine types throughout its lifecycle. However, with the transition to CentOS Stream 10, compatibility with machine types prior to `pc-q35-rhel9.0.0` will not be maintained: 
```
$ cat /etc/centos-release 
CentOS Stream release 9
$ rpm -qa | grep qemu-kvm-core
qemu-kvm-core-8.0.0-5.el9.x86_64
$ /usr/libexec/qemu-kvm -machine ?
Supported machines are:
pc                   RHEL 7.6.0 PC (i440FX + PIIX, 1996) (alias of pc-i440fx-rhel7.6.0)
pc-i440fx-rhel7.6.0  RHEL 7.6.0 PC (i440FX + PIIX, 1996) (default) (deprecated)
q35                  RHEL-9.2.0 PC (Q35 + ICH9, 2009) (alias of pc-q35-rhel9.2.0)
pc-q35-rhel9.2.0     RHEL-9.2.0 PC (Q35 + ICH9, 2009)
pc-q35-rhel9.0.0     RHEL-9.0.0 PC (Q35 + ICH9, 2009)
pc-q35-rhel8.6.0     RHEL-8.6.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.5.0     RHEL-8.5.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.4.0     RHEL-8.4.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.3.0     RHEL-8.3.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.2.0     RHEL-8.2.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.1.0     RHEL-8.1.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel8.0.0     RHEL-8.0.0 PC (Q35 + ICH9, 2009) (deprecated)
pc-q35-rhel7.6.0     RHEL-7.6.0 PC (Q35 + ICH9, 2009) (deprecated)
none                 empty machine
```
This means each workload must be updated to use a `pc-q35-rhel9.x.x` machine type.

## Goals
* Create an opt-in workflow that automates mass converting machine types of VMs
	* if currently running, update VM machine type and add a label to alert the user the VM must be restarted for the change to take effect
	* alternatively, if currently running, take VM offline immediately and update machine type
	* if not currently running, update machine type immediately
* Allow user to specify certain VMs for updating machine type automation (e.g. by namespace)

## Non Goals
While these additions could be beneficial to cluster-admins who wish to use the mass machine type transition (MMTT), these are not covered in the scope of this design proposal: 
* Allowing user to filter VMs by running state: a cluster admin may want to only update VMs that are offline.
* Allowing user to filter VMs in ways other than by namespace; e.g. with label-selector

Also not covered in this design proposal is the method in which users/cluster admins will be informed that they have VMs with unsupported machine types.

## User Stories
As a cluster-admin, I want an automated way to update the machine type of all VMs to be compatible with CentOS Stream X without interrupting my workflow.
As a cluster-admin, I want to be able to include only certain VMs to be updated (e.g. by namespace).
As a cluster-admin, I want to be in control of the behavior of running VMs being updated; I can determine whether I want to manually restart running VMs being updated or let the automation restart all of them for me. 

## Repos
[kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design
Create a new virtctl command that will allow the user to automatically update the machine type of any VMs with a machine type that is no longer supported. This command invokes a k8s `Job` that iterates through all VMs within a specified namespace (or all namespaces if none is specified), determines if the machine type is no longer supported, and updates it to the latest supported version. Both the minimum supported machine type and the target machine type are constants that can be updated as future versions are released and more machine types are deprecated upon these releases. 

If the VM is running, the update will not take effect until it is restarted, so the label `restart-vm-required` will be applied to the VM until the VMI is removed (indicating the VM has been stopped), upon which the label will be cleared after verifying that its machine type has been successfully updated. The job will only terminate once all `restart-vm-required` labels have been cleared. By default, the user must handle stopping/restarting the labelled VMs manually. This is to allow the user to safely update the machine types of their VMs at a time they choose without worrying about workload interruptions. However, the user also has the option to allow the job to automatically restart every running VM that is being updated and apply the changes immediately.

The options to select a specific namespace or restart the VMs immediately are configurable with the `NAMESPACE` and `RESTART_NOW` environment variables respectively.

Given a VM with the following spec:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
...
spec:
  running: false
  template:
    spec:
      domain:
        machine:
          type: pc-q35-rhel8.2.0
```
When the MMTT is invoked, the VM spec will be updated to: 
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
...
spec:
  running: false
  template:
    spec:
      domain:
        machine:
          type: pc-q35-rhel9.2.0
```
where `pc-q35-rhel9.2.0` is the latest machine type version.

Along with machine types being in the format `pc-q35-rhelx.x.x`, a VM may have a machine type of `q35`, an alias for the current latest machine type version *at the time the VM is started*.

For example, this VM:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
...
spec:
  running: true
  template:
    spec:
      domain:
        machine:
          type: q35
```
may have been started when a RHEL-8 machine type was the latest machine type version and not been stopped or restarted since. Its VMI would thus look like this:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
...
spec:
  domain:
    machine:
      type: q35
status:
  machine:
    type: pc-q35-rhel8.2.0
```
and need to have its machine type updated.

To handle this, any running VMs with 'q35' machine type and an outdated machine type (reported in VMI's status.machine.type) will be labelled with `restart-vm-required`:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
	labels: "restart-vm-required"="true"
spec:
  running: true
  template:
    spec:
      domain:
        machine:
          type: q35
```
Once the VM has been stopped or restarted, the `restart-vm-required` label will be removed:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  running: false
  template:
    spec:
      domain:
        machine:
          type: q35
```
Now when a new VMI is created when starting this VM, it should have the latest machine type version:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachineInstance
...
spec:
  domain:
    machine:
      type: q35
status:
  machine:
    type: pc-q35-rhel9.2.0
```

## Update/Rollback Compatibility
As both the minimum supported machine type version and the latest machine type version change in the future, these values are global constants in the MMTT package. As the versions are updated in the future, these global constants can also be updated accordingly, allowing for easy maintainability.

## Functional Testing Approach
* Functional tests will follow the same basic procedure:
	* Create (a) VM(s) with the necessary machine type for the test case; start the VM if testing functionality of running VMs
	* Configure environment variables `NAMESPACE` and `RESTART_NOW` as necessary for the test case
	* Ensure VM spec has the correctly updated machine type version
	* Ensure the VM has/doesn't have the `restart-vm-required` label
	* Ensure all VMs have been updated accordingly
* Test cases:
	* Single VM (each of these will be their own individual test)
		* Running
			* VM machine type version less than the minimum supported version
				* `RESTART_NOW` is **false**
				* `RESTART_NOW` is **true**
			* VM machine type is `q35`
				* `RESTART_NOW` is **false**
				* `RESTART_NOW` is **true**
			* VM machine type version is greater than or equal to the minimum supported version
		* Not running
			* VM machine type version less than the minimum supported version
			* VM machine type version is greater than or equal to the minimum supported version
			* VM machine type is equal to `q35`
	* Multiple VMs (these cases will be split into 4 functional tests based on the environment variable configuration: `NAMESPACE`  specified/unspecified and `RESTART_NOW` **true**/**false**)
		* `NAMESPACE` is specified
			* `RESTART_NOW` is **true**
				* Running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
			* `RESTART_NOW` is **false**
				* Running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
		* `NAMESPACE` is not specified
			* `RESTART_NOW` is **true**
				* Running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
			* `RESTART_NOW` is **false**
				* Running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type version less than the minimum supported version
					* VM machine type version is greater than or equal to the minimum supported version
					* VM machine type is equal to `q35`

# Implementation Phases

 - [ ] Create MMTT package that Kubernetes job will use as an image
 - [ ] Create Kubernetes job yaml to be invoked with MMTT subcommand
 - [ ] Insert package and job into kubevirt/pkg/virtctl and implement the subcommand (name TBD)
