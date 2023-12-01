# Overview
This proposal discusses a method to automate updating the machine type of VMs.
Currently we can support manually changing the machine type of individual VMs,
but we need an automated process that will be able to change the machine types
of multiple (e.g. thousands of) VMs, limiting workload interruptions and
preventing the user from having to manually update the machine type of every VM
with a machine type version that is no longer supported. For example, CentOS
Stream 9 will maintain compatibility with all `pc-q35-rhel8.x.x` machine types
throughout its lifecycle. However, with the transition to CentOS Stream 10,
compatibility with machine types prior to `pc-q35-rhel9.0.0` will not be
maintained: 
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
	* if currently running, update VM machine type and update its Status to alert
  the user the VM must be restarted for the change to take effect
	* alternatively, if currently running, take VM offline immediately and update
  machine type
	* if not currently running, update machine type immediately
* Allow user to specify certain VMs for updating machine type automation
(e.g. by namespace)

## Non Goals
While these additions are beneficial to cluster-admins who wish to use the mass
machine type transition (MMTT), these will not be included in the initial
implementation:
* Allowing user to filter VMs by running state: a cluster admin may want to only
update VMs that are offline.
* Subcommands that allow the user to monitor and manage the status of the MMTT
job and affected VMs

These will eventually be implemented in follow ups once the initial design is implemented.

Additionally, this design proposal does not cover the method in which users
(cluster-admins) will be informed that they have VMs with unsupported machine
types.

## User Stories
* As a cluster-admin, I want an automated way to update the machine type of all
VMs with outdated machine types without interrupting my workflow.
* As a cluster-admin, I want to be able to include only certain VMs to be updated
(e.g. by namespace, label-selector).
* As a cluster-admin, I want to be in control of the behavior of running VMs being
updated; I can determine whether I want to manually restart running VMs being
updated or let the automation restart all of them for me. 

## Repos
[kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)

# Design
Create a new virtctl command `virtctl update machine-types` that will allow
the user to automatically update the machine type of any VMs with a machine type
that matches the machine type regex provided by the user. This command invokes
a Kubernetes `Job` that iterates through all VMs within a specified namespace
(or all namespaces if none is specified), and if the machine type matches the
regex, it removes the machine type field. Removing the machine type means that
the next time that VM is started/restarted, its machine type will automatically
be populated with the default machine type stored in the Kubevirt CR, which will
always be a supported machine type.

If the VM is running, the update will not take effect until it is restarted. 
By default, the user must handle stopping/restarting the affected VMs manually.
This is to allow the user to safely update the machine types of their VMs at a
time they choose without worrying about workload interruptions. However, the
user also has the option to request the job to automatically restart every
running VM that is being updated and apply the changes immediately. In order
for the job to know when all VMs have been updated and all affected running VMs
have been restarted, we will add a field to the VM Status `MachineTypeRestartRequired`,
which the job will set for all running VMs (unless the user requests an
automatic restart). When the job detects the VMI removal, after verifying that
its machine type has been successfully updated in the VM Spec, it will update
the VM Status `MachineTypeRestartRequired` field. Once all the VMs' Status
indicate that they have been restarted, the job will finally terminate. 

If the Job fails or is killed, it will be restarted with `restartPolicy=OnFailure`
in the Job spec. The `MachineTypeRestartRequired` field may then remain set in
the status of a VM that has already been shut down, so the Job will first check
VMs with the field set and confirm if the machine types have been updated. If
they have, then the Status will be updated accordingly; otherwise, the Status
will remain and the VM will continue to be tracked until it is restarted and its
machine type is updated.

The user may use the `--which-matches-glob` flag to specify the regex of the
machine types the user wishes to update. The user may also specify by namespace
or label-selector, or restart the VMs immediately are using the `--namespace`,
`--label-selector`, and `--restart-now` respectively. In follow-ups to the
initial implementation, other methods of specifying which VMs to convert the
machine types of will be added, such as specifying a single VM by name, or
configuring a limit on the number of VMs that can be restarted at once when
restarting the running VMs immediately.

Using the CentOS Stream example from the overview, given a VM with the following spec:
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
When the user calls the virtctl command:
`virtctl update machine-types --which-matches-glob *rhel-8.*`

The MMTT Job is deployed, and the VM spec will be updated to: 
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
          type:
```
Where the machine type in the VM spec is now empty. Upon starting this VM,
the machine type will automatically be populated with the default machine type
in the Kubevirt CR.

A VM may have a machine type of `q35`, an alias for the current latest machine
type *at the time the VM is started*.

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

To handle this, any running VMs with 'q35' machine type in its spec and a machine
type in the VMI's status.machine.type matching the regex will have its Status
field `MachineTypeRestartRequired` updated:
```yaml
---
apiVersion: kubevirt.io/v1
kind: VirtualMachine
spec:
  running: true
  template:
    spec:
      domain:
        machine:
          type: q35
status:
  machineTypeRestartRequired: true
```
Once the VM has been stopped or restarted, the `MachineTypeRestartRequired` field
will be updated once again:
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
status:
  machineTypeRestartRequired: false
```
Now when a new VMI is created when starting this VM, it should have the default
machine type from Kubevirt CR.
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
change in the future when new versions are released, Kubevirt CR is already able
to determine the latest machine type, so that will automatically be updated.

## Functional Testing Approach
Functional tests will follow the same basic procedure:
* Create VMs with the necessary machine type for the test case; start the VM if
testing functionality of running VMs
* Create and execute the `virtctl update machine-types` command with the
designated flags for that test case
* Ensure that the VMs with the specified machine type have been updated
* Ensure that flag-specific conditions have been fulfilled
* Each virtctl command flag will have its own individual test
* There will also be a complex example that combines all the flags into one test.
* All tests will use `--which-matches-glob *rhel-8.*` as the machine type to update.

* `--namespace` flag test:
	* Across 2 different namespaces, each namespace will have one stopped VM and
  one running VM with a machine type that needs to be updated.
	* Verify that only the VMs in the specified namespace were updated
	* Verify that only the running VM in the specified namespace has its
  `Status.MachineTypeRestartRequired` set true.
* `--label-selector` flag test:
	* Two VMs with specified label, two without the label. Each set of VMs will
  have one running and one stopped VM.
	* Verify that only the VMs with the specified label were updated.
	* Verify that only the running VM with the specified label has its
  `Status.MachineTypeRestartRequired` set true.
* `--restart-required` flag test:
	* One running VM and one stopped VM with machine types that need to be updated.
	* Verify that the  VMs are updated.
	* Verify that the running VM does not have its
  `Status.MachineTypeRestartRequired` set true.
	* Verify that a new VMI object has been created with the updated machine type
  in its Status.
* The complex example will be a combination of these three test cases.


# Implementation Phases
## Initial Phase
This will be a bare-bones design that will be functional for the user, but with
limited features. The user must specify the machine type they wish to update
and have the ability to use the command with the flags to specify by namespace,
label-selector, or to restart any running VMs immediately. Additionally, to
monitor and manage/delete the job, the user can utilize the `kubectl` commands.
In future phases, we will add subcommands that allow the user to monitor and manage the specific MMTT job directly.
 - [ ] Create mass machine type transition package that k8s job will use as an
 image, and build the image.
 - [ ] Add `virtctl` subcommand that will create the k8s job and run it, with
 flags for specifying by namespace, label-selector, and immediately restarting
 running VMs that have been affected.
 - [ ] Add unit and functional tests.
## Future Phases
As follow-ups, we will implement the following features:
 - Dynamically retrieving the supported and deprecated QEMU machine types as a
 subcommand to provide the user with a list of machine types to choose from. This
 will be implemented in conjuction with the ability for the user to specify a
 target machine type to update to; this machine type will be selected from the
 list of supported machine types provided to the user.
 - Subcommands that the user can use to monitor and manage the jobs. These could
 include `virtctl update machine-types delete` which will safely terminate and
 clean up the job and delete it, and `virtctl update machine-types list` which will allow the user to see the status of the VMs being updated by the job.
 - Other flags that the user can specify VMs to be affected by. For example, if
 the user wants to only convert a single specific VM, we might want to implement
 a flag to allow the user to specify the VM by name.
 - When using the `restart-now` flag, allow the user to configure how many VMs
 they want to restart at once; depending on the number of running VMs, it may be
 very intensive to trigger the restart of possibly thousands of VMs at once.
