# KubeVirt Incubation Stage Review Proposal

## What is KubeVirt: Refresh
KubeVirt technology addresses the needs of development teams that have adopted or want to adopt Kubernetes but possess existing Virtual Machine-based workloads that cannot be easily containerized. More specifically, the technology provides a unified development platform where developers can build, modify, and deploy applications residing in both Application Containers as well as Virtual Machines in a common, shared environment.

Benefits are broad and significant. Teams with a reliance on existing virtual machine-based workloads are empowered to rapidly containerize applications. With virtualized workloads placed directly in development workflows, teams can decompose them over time while still leveraging remaining virtualized components as is comfortably desired.

## Statement on alignment with the CNCF mission:

## Project progress since sandbox

## Metrics

### Mailing list
KubeVirt currently (2021/06/01) has 436 subscribers.  The mailing list typically get 0-5 new threads per day.

### GitHub
* 171 official members in the organization
* About 50 contributors measured by multiple L or greater sized contributions
* 96 watched tags
* 2562 stars
* 615 forks

### Slack
Slack channels typically get 0-5 new threads per day.

### Downloads

### Integrations
* Kubernetes / minikube
* Red Hat OKD / OpenShift 4.x
* oVirt / Red Hat Virtualization

### Community

#### Communications

Kubevirt uses several technologies to maintain the community.

Source code is stored in git repositories under a GitHub organization provided by the CNCF (https://www.github.com/kubevirt).  GitHub issues are used to track reqs for engineering and bug tracking.  GitHub pull requests are used to peer review contributions.

The project utilizes Google Groups as a mailing list where use, support, proposal topics are discussed.

The project holds a video conference every week via a Zoom account provided by the CNCF.  General topics, support and bug triage are conducted in this meeting.  The project also conducts a weekly meeting based on the topic of performance and scale.  Both meetings are recorded and posted to the project YouTube channel.  Meeting notes are emailed to the general mailing list.

The project uses two channels on the CNCF/Kubernetes Slack server.  #virtualization is used to handle general use and support topics.  #kubevirt-dev is used for developer oriented communication.

KubeVirt utilizes email, website blog and Twitter for important announcements such as version releases.

KubeVirt advertises communications channels via https://kubevirt.io/community as well as project README (https://github.com/kubevirt/kubevirt/blob/master/README.md).

#### End Users

### CNCF Sponsored Security Audit

## Features & Roadmap

### Roadmap at Sandbox
* [DONE] GA v1 API for core KubeVirt APIs
 * API v1 features need to rely on GA’ed Kubernetes entities, fully fledged (incl e.g. explain, validation)
 * An OpenAPI definition as the only source of truth for KubeVirt’s API
 * https://github.com/kubevirt/kubevirt/pull/3349
* [DONE] Stabilize or replace bridge network binding

### Future Roadmap
* [WIP] Non-root VMI Pods
* [WIP] Persistent containerDisk volumes
* [WIP] Establish predictable community release and support patterns
* [WIP] Define a deprecation policy
* [WIP] Review and Revise User Guide
* [WIP] Virt-launcher live updates
* [WIP] Templating mechanism for VMs
