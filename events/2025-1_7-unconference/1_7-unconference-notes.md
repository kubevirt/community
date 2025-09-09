# KubeVirt v1.7 Unconference notes (08 July, 2025)

These notes come from the hedgedoc notes that we used for the unconference and have been lightly edited for formatting.

## VEP process retrospective

- Review load was huge. How to address that better?
    - Too many VEPs.
        - We could limit #VEPs
        - Always count with buffer (1-2 people being free) ? 
    - We need to have "Assigned reviewers" to budget for bw
    - Reviewers capacity was affected and we did have CI issues as well. We are working on the CI issues for smoother next release.
- Alpha features should have no effect on production code in theory. But:
    - In practice, FG enablement check is hard to do in virt-launcher.
    - Migration controllers were heavily refactored to support an alpha VEP, but these changes are not feature gated...
- We needed many exceptions this time. How can we improve this moving forward?
- Acks from all SIGs
    - We need to ack from every SIG before merge and deadline
    - [lyarwood] Enforce with labels?
        - [lyarwood] I've created https://github.com/kubevirt/enhancements/issues/74 to track this
    - [vladikr] Make sure that each SIG tracks their own VEPs
- More subtopics are welcome...

## Adding VirtualMachineTemplates to KubeVirt

- external repo but deploy by virt-operator?

## Advancing VirtualMachinePool to beta

- [Sreeja1725]Current status of VirtualMachinePool is in alpha, For kubevirt v1.7, the main objective is to advance the feature from alpha to beta by implementing the valuable lifecycle aspects of this API
    - https://github.com/kubevirt/enhancements/issues/69


## Collect Guest CPU-Load Metrics from libvirt

**Context:** https://github.com/kubevirt/enhancements/pull/67

- https://github.com/kubevirt/kubevirt/pull/14879
- https://groups.google.com/g/kubevirt-dev/c/UhcV9B3z_eY/m/azWVxUbTAQAJ - exception ML
- Too late for v1.6.0, lets at least land in main for v1.7.0 for now.

## Reducing KubeVirt's dependancies on external projects

- [omisan] Reducing KubeVirt's dependancies on external projects
    - [Lubo] I generally agree

## Miscellaneous

- [aburden] Adding a 'docs-needed' label
    - Should we ask PR authors to provide user-guide docs PRs before their PR can merge?
    - General agreement that this is a good idea and should proceed. 

- CentOS 10 rebase when?
    - No ideas in the room

- Linting
    - lots of free linting paths found by omisan, sig-compute to review
