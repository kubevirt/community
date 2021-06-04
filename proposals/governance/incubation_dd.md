# KubeVirt Incubation Stage Review Proposal

## What is KubeVirt: Refresh

KubeVirt technology addresses the needs of development teams that have adopted or want to adopt Kubernetes but possess existing Virtual Machine-based workloads that cannot be easily containerized. More specifically, the technology provides a unified development platform where developers can build, modify, and deploy applications residing in both Application Containers as well as Virtual Machines in a common, shared environment.

Benefits are broad and significant. Teams with a reliance on existing virtual machine-based workloads are empowered to rapidly containerize applications. With virtualized workloads placed directly in development workflows, teams can decompose them over time while still leveraging remaining virtualized components as is comfortably desired.

## Statement on alignment with the CNCF mission

## Community

### Governance

* [Membership policy](https://github.com/kubevirt/community/blob/master/membership_policy.md)

### SIGs

* [SIG blah](https://github.com/kubevirt/community/blob/master/sigs.yaml)

### Maintainers

*
*
*

### GitHub source code repository

* About 50 contributors measured by multiple L or greater sized contributions
* Avg of 5 commits per week
* 171 official members in the organization
* 2562 stars
* 615 forks
* 96 watched tags

### Communications
The project utilizes Google Groups as a mailing list where where users have a chance to interact with core developers and discuss general topics. There are currently (2021/06/01) 436 subscribers.  The mailing list typically receives 0-5 new threads per day.

The project holds a video conference every week where users have a chance to interact with core developers, discuss general topics and participate in bug triage. Recently the performance and scale working group started conducting a weekly meeting based on the topic. Both meetings are recorded and posted to the project YouTube channel. Meeting notes are also emailed to the general mailing list to keep the community informed.

The project uses two channels on the CNCF/Kubernetes Slack server. #virtualization is used to handle general use topics. #kubevirt-dev is used for developer oriented topics.  Slack channels typically receive 0-5 new threads per day.

Important announcements are relayed via mailing list, website blog and Twitter.

KubeVirt advertises communications channels via https://kubevirt.io/community as well as project [README](https://github.com/kubevirt/kubevirt/blob/master/README.md).

### Integrations

* [oVirt](https://www.ovirt.org)
* [Gardener](https://gardener.cloud/blog/2020-10/00/)
* [Kubermatic Virtualization](https://www.kubermatic.com/products/kubevirt/)
* [SUSE/Rancher Harvester](https://github.com/rancher/harvester/blob/766abd06561b059c1af623aacc4e505db471ceee/deploy/charts/harvester/README.md)
* [Google Anthos](https://youtu.be/RE0A3kHT3LA?t=126)

### End Users

Project end users are maintained in the [Adopters](https://github.com/kubevirt/kubevirt/blob/main/ADOPTERS.md) files

Significant users include:
* NVIDIA
* SUSE
* Kubermatic
* H3C
* CoreWeave
* Civo
* CloudFlare
* Ateme
* Cloudbase Solutions

### CNCF Sponsored Security Audit

## Product information

### Feature design proposals
Design proposals to allow community members to gain feedback on their designs from the repo approvers before the community member commits to executing on the design. By going through the design process, developers gain a have a high level of confidence that their designs are viable and will be accepted.

* [design-proposals](https://github.com/kubevirt/community/tree/master/design-proposals)

### Release cadence

Kubevirt has a establihed and documented release process and cadence

* [Release process](https://github.com/kubevirt/kubevirt/blob/main/docs/release.md)
* [Release cadence](https://github.com/kubevirt/kubevirt/blob/main/docs/release.md#cadence-and-timeline)

### Delivered features

* [DONE] GA v1 API for core KubeVirt APIs
 * API v1 features need to rely on GA’ed Kubernetes entities, fully fledged (incl e.g. explain, validation)
 * An OpenAPI definition as the only source of truth for KubeVirt’s API
 * https://github.com/kubevirt/kubevirt/pull/3349
* [DONE] Zero downtime live updates
* [DONE] Stabilize or replace bridge network binding
* [DONE] Disk hotplug
* [DONE] IPv6 support
* [DONE] Device passthrough support
* [DONE] CPU pinning support
* [DONE] Memory metrics gathering support
* [DONE] Affinity / Anti-Affinity rules
* [DONE] Live-Migration support
* [DONE] Offline disk snapshots
* [DONE] SRIOV support
* [DONE] Dynamic SSH Key Injection
* [DONE] Multus support for multiple network interfaces attached to Virtual Machines
* [DONE] Dedicated prow deployment for CI functional tests and automation

### Future Roadmap

* [WIP] Non-root VMI Pods
* [WIP] Establish predictable community release and support patterns
* [WIP] Define a deprecation policy
* [WIP] Review and Revise User Guide
* [WIP] Templating mechanism for VMs
* [WIP] Monitoring and metrics standardization
* [WIP] CPU NUMA topology support
* [WIP] Macvtap support
* [WIP] SSH proxy ingress support

## Incubation Stage Requirements

The KubeVirt project maintainers propose that KubeVirt move to Incubation based on:

* Use in production by 3 significant end users
* A healthy number of committers and a growing committer base in addition, to a healthy online community.
* Demonstrating a substantial ongoing flow of commits and merged contributions that focused on delivering a defined project roadmap and integrations.
* A clear versioning scheme with dev and stable releases.
