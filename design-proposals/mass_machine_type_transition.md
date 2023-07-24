# Overview
This proposal discusses a method to automate updating the machine type of VMs
to the latest machine type version, if the VM has a machine type that is no
longer compatible. Currently we can support manually changing the machine type of
individual VMs, but we need an automated process that will be able to change the
machine types of multiple (e.g. thousands of) VMs, limiting workload interruptions
and preventing the user from having to manually update the machine type of every
VM with a machine type version that is no longer supported. For example, CentOS
Stream 9 will maintain compatibility with all `pc-q35-rhel8.x.x` machine types
throughout its lifecycle. However, with the transition to CentOS Stream 10,
compatibility with machine types prior to `pc-q35-rhel9.0.0` will not be maintained: 
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
While these additions are beneficial to cluster-admins who wish to use the
mass machine type transition (MMTT), these will not be included in the initial implementation:
* Allowing user to filter VMs by running state: a cluster admin may want to only update VMs that are offline.
* Subcommands that allow the user to monitor and manage the status of the MMTT job and affected VMs

These will eventually be implemented in follow ups once the initial design is implemented.

Additionally, this design proposal does not cover the method in which users/cluster admins
will be informed that they have VMs with unsupported machine types.

## User Stories
As a cluster-admin, I want an automated way to update the machine type of all
VMs with unsupported machine types without interrupting my workflow.
As a cluster-admin, I want to be able to include only certain VMs to be updated (e.g. by namespace, label-selector).
As a cluster-admin, I want to be in control of the behavior of running VMs being
updated; I can determine whether I want to manually restart running VMs being
updated or let the automation restart all of them for me. 

## Repos
[kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design
Create a new virtctl command `virtctl update machine-types` that will allow
the user to automatically update the machine type of any VMs with a machine type
that is no longer supported. This command invokes a k8s `Job` that iterates
through all VMs within a specified namespace (or all namespaces if none is
specified), determines if the machine type is no longer supported, and updates
it to the latest supported version. If the Job fails or is killed, it will be
restarted with `restartPolicy=OnFailure` in the Job spec. The `restart-vm-required`
label may then remain on a VM that has already been shut down, so the Job will
first check VMs with the label and confirm if the machine types have been updated.
If they have, then the label will be removed, but if not, the label will remain
and the VM will continue to be tracked until it is restarted and its machine
type is updated. Both the minimum supported machine type and the target machine
type will be determined (retrieved) and configured internally when the command is
invoked, and these values will automatically be updated as future versions are
released and more machine types are deprecated upon these releases.

If the VM is running, the update will not take effect until it is restarted,
so the label `restart-vm-required` will be applied to the VM until the VMI is
removed (indicating the VM has been stopped), upon which the label will be cleared
after verifying that its machine type has been successfully updated. The job will
only terminate once all `restart-vm-required` labels have been cleared. By default,
the user must handle stopping/restarting the labelled VMs manually. This is to allow
the user to safely update the machine types of their VMs at a time they choose
without worrying about workload interruptions. However, the user also has the
option to allow the job to automatically restart every running VM that is being
updated and apply the changes immediately.

The options to specify by namespace, label-selector, or restart the VMs
immediately are configurable with the `--namespace`, `--label-selector`,
and `--restart-now` respectively. In follow-ups to the initial implementation,
other methods of specifying which VMs to convert the machine types of will be
added, such as specifying a single VM by name, or configuring a limit on the
number of VMs that can be restarted at once when restarting the running VMs
immediately.

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

Along with machine types being in the format `pc-q35-rhelx.x.x`, a VM may have
a machine type of `q35`, an alias for the current latest machine type *at the time the VM is started*.

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
may have been started when a RHEL-8 machine type was the latest machine type and
not been stopped or restarted since. Its VMI would thus look like this:
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

To handle this, any running VMs with 'q35' machine type and an outdated machine
type (reported in VMI's status.machine.type) will be labelled with `restart-vm-required`:
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
Now when a new VMI is created when starting this VM, it should have the latest machine type:
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
As both the minimum supported machine type version and the latest machine type
change in the future when new versions are released, Kubevirt is already able
to determine the latest machine type, so that will automatically be updated.
The command will also be fetching the minimum supported machine type using
the list of supported and deprecated QEMU machine types, thus minimizing the
need to manually maintain these values.

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
			* VM machine type version less than the minimum supported machine type
				* `RESTART_NOW` is **false**
				* `RESTART_NOW` is **true**
			* VM machine type is `q35`
				* `RESTART_NOW` is **false**
				* `RESTART_NOW` is **true**
			* VM machine type version is greater than or equal to the minimum supported machine type
		* Not running
			* VM machine type version less than the minimum supported machine type
			* VM machine type version is greater than or equal to the minimum supported machine type
			* VM machine type is equal to `q35`
	* Multiple VMs (these cases will be split into 4 functional tests based on the environment variable configuration: `NAMESPACE`  specified/unspecified and `RESTART_NOW` **true**/**false**)
		* `NAMESPACE` is specified
			* `RESTART_NOW` is **true**
				* Running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
			* `RESTART_NOW` is **false**
				* Running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
		* `NAMESPACE` is not specified
			* `RESTART_NOW` is **true**
				* Running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
			* `RESTART_NOW` is **false**
				* Running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`
				* Not running
					* VM machine type is less than the minimum supported machine type
					* VM machine type is greater than or equal to the minimum supported machine type
					* VM machine type is equal to `q35`

# Implementation Phases
## Initial Phase
This will be a bare-bones design that will be functional for the user, but with
limited features that focus on the specific use case of upgrading the machine
types of VMs with outdated or deprecated machine types to the latest supported
machine type. The user will have the ability to use the command with the flags
to specify by namespace, label-selector,  or to restart any running VMs immediately.
In this stage, the minimum supported machine type and the latest machine type will
be determined and stored internally; the user will not be able to configure what
machine types they would like to use at this time. Additionally, to monitor and
manage/delete the job, the user can utilize the `kubectl` commands. In future phases,
we will add subcommands that allow the user to monitor and manage the specific machine
type transition job directly.
 - [ ] Create mass machine type transition package that k8s job will use as an image, and build the image.
 - [ ] Add `virtctl` subcommand that will create the k8s job and run it, with flags for specifying by namespace, label-selector, and immediately restarting running VMs that have been affected.
 - [ ] Add unit and functional tests.
## Future Phases
As follow-ups, we will implement the following features:
 - Dynamically retrieving the supported and deprecated QEMU machine types and allowing the user to select a specific machine type to migrate to. Some more discussion and planning will be required to flesh out exactly how to implement this; namely how the user will know what machine types are supported and any restrictions we want to place on the user when selecting a machine type to convert to.
 - Subcommands that the user can use to monitor and manage the jobs. These could include `virtctl update machine-types delete` which will safely terminate and clean up the job and delete it, and `virtctl update machine-types list` which will allow the user to see the status of the VMs being updated by the job.
 - Other flags that the user can specify VMs to be affected by. For example, if the user wants to only convert a single specific VM, we might want to implement a flag to allow the user to specify the VM by name.
 - When using the `restart-now` flag, allow the user to configure how many VMs they want to restart at once; depending on the number of running VMs, it may be very intensive to trigger the restart of possibly thousands of VMs at once.
