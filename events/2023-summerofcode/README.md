# Google Summer of Code 2023

"Google Summer of Code (GSoC) is a global, online program that brings new contributors into open source software organizations." - [Google Summer of Code Contributor Guide](https://google.github.io/gsocguides/student/)

The KubeVirt community is applying to be a Google Summer of Code organization, to provide mentorship opportunity to applicants interested in learning about open source software development in the cloud native ecosystem. 

See the [Google Summer of Code website](https://summerofcode.withgoogle.com/) for more information about the program.

## Key Dates

Feb 22: List of accepted organizations announced <br />
Feb 22 - Mar 19: Potential contributors discuss project application ideas with organizations <br />
Apr 4: Contributor application deadline <br />
May 4 - 28: Community Bonding Period <br />
May 29 - Aug 28: The Summer of Code

See the [Google Summer of Code timeline](https://developers.google.com/open-source/gsoc/timeline) for more detailed timeline information.

## Project Ideas

KubeVirt is proposing the following project ideas as starting points for GSoC contributors to develop their own project applications.

### Create KubeVirt seccomp Profiles
**Github issue**: https://github.com/kubevirt/community/issues/205

**Description**: [Seccomp](https://man7.org/linux/man-pages/man2/seccomp.2.html) is a security facility from the Linux Kernel that prevents processes to execute unauthorized syscalls.  By limiting the number of permitted syscalls, seccomp is being utilized in conjunction with [Kubernetes](https://kubernetes.io/docs/tutorials/security/seccomp/) to reduce the attack surface of the containers.
Container engines offer their own default profile. However, we cannot assume that one size fits all. Therefore, the default profile may either permit syscalls that are in fact not required by the workload or prohibit legitimate syscalls.

**Expected outcomes**: Seccomp custom profiles are already supported by KubeVirt, although they are still based on the cri-o seccomp profile. The goal of this internship is to integrate and automate the syscall auditing during testing, then utilize the test results to create a seccomp profile with the syscalls actually used at runtime. Finally, the seccomp profile will be applied to the test suite to ensure that it does not block any needed syscalls.

As an optional addition, the intern could look into if various seccomp profiles can be applied depending on the feature specified by the Virtual Machine description if the project's time constraints permit it.

**Links**: <br />
[Seccomp Linux man page](https://man7.org/linux/man-pages/man2/seccomp.2.html) <br />
[Kubernetes with seccomp](https://kubernetes.io/docs/tutorials/security/seccomp/) <br />
[Seccomp profile for KubeVirt](https://github.com/kubevirt/kubevirt/pull/8917)

**Project size**: 12 weeks

**Required skills**: Kubernetes knowledge

**Desirable skills**: Virtualization and GoLang programming skills

**Mentors**: Alice Frosi <afrosi@redhat.com>, Co-mentor: Luboslav Pivarcl <lpivarc@redhat.com>


### POC Virtual Machine Runtime Interface
**Github issue**: https://github.com/kubevirt/community/issues/206

**Description**: Kubevirt is a Kubernetes extension to run virtual machines on Kubernetes clusters leveraging Libvirt + Qemu&KVM stack. It does this by exposing a custom resource called VirtualMachine which is then translated into a Pod (called virt-launcher). This Pod is treated as any other application pod, and includes a monitoring process, virt-launcher, that manages the Libvirt+Qemu processes.
Libvirt needs to run in the same context as QEMU, therefore is launched in each virt-launcher pod together with the monitorning process “virt-launcher”. 

Unfortunately, one of the drawbacks of this additional component is the increment of the required memory per pod or additional CPUs to run on. This differs from traditional virtual machine platforms as they usually can use one Libvirt instance per node and therefore isolate the CPU and memory consumption. Kubevirt is tightly integrated with Libvirt which makes it hard to integrate with another VMM (this is similar to how Kubernetes can leverage containerd/crio or others). The stretch goal of this project is to design an interface that would allow Kubevirt to interact with any VMM without need for Libvirt.

**Expected outcomes**: The main goal of this project is to create a proof of concept to refactor the current virt-launcher code into a node level component (running one instance per node vs per Pod). This would reduce the memory overhead of each VM and demonstrate that abstraction of communication with Pods would be possible.

**Links**: <br />
https://github.com/kubevirt/kubevirt/blob/main/docs/README.md <br />
https://github.com/kubevirt/kubevirt/blob/main/docs/architecture.md

**Project size**: 12 weeks

**Required skills**: Kubernetes knowledge and Golang

**Desirable skills**: GRPC, Unix 

**Mentors**: Luboslav Pivarcl <lpivarc@redhat.com>, Co-mentor:  Alice Frosi <afrosi@redhat.com>

### Custom project proposals

You can submit your own project idea by emailing the [kubevirt-dev Google Group](https://groups.google.com/forum/#!forum/kubevirt-dev) and CC'ing Andrew Burden <aburden@redhat.com>.

If a mentor from the KubeVirt community supports the proposed project idea, we can add it to the KubeVirt project ideas list.


