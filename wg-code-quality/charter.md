# WG Code-Quality

## Scope

The Code-Quality Working Group improves the quality of the KubeVirt codebase and creates an inter-SIG process for identifying, agreeing on, and addressing code quality issues.

The short-term mission is the improvement of code quality. The long-term mission is met when that inter-SIG process is implemented and executed as required.

### In scope

* Create and maintain code-quality guidelines, best practices, and standards.
* Distribute knowledge about code quality, including presentations, documentation, lectures, and linters.
* Enforce selected standards through automation, review, and shared codebase infrastructure such as reusable libraries.
* Prioritize quality work and focus the group on selected topics.
* Monitor and identify code-quality issues.
* Define the inter-SIG process: the tooling and documentation required, the code-structure metrics to establish and how they are implemented, how reviewer and maintainer awareness is achieved, and how and when those measures are executed.

The working group facilitates this effort until enough tooling, documentation, code structure, and reviewer and maintainer awareness is in place.

### Out of scope

* Owning feature code. SIGs that claim code ownership keep approval of that code through their OWNERS files.
* Deciding product or feature design, except where a change is required to meet an agreed quality standard.
* Operating CI infrastructure. [SIG CI](../sig-ci/charter.md) keeps the infrastructure running. This group coordinates with SIG CI when a quality measure needs automation.
* Fixing functional bugs that are not code-quality issues.

### Stakeholder SIGs

Every SIG that owns code and produces it is a stakeholder. The current primary counterparts are:

* [SIG Compute](../sig-compute/charter.md)
* [SIG Network](../sig-network/charter.md)
* [SIG Storage](../sig-storage/charter.md)
* [SIG CI](../sig-ci/charter.md), for automation of code-quality measures

## Roles and Organization Management

This working group follows the Roles and Organization Management outlined in the [membership policy]. Chair requirements, responsibilities, and inactivity handling are defined there. Chairs and the liaison are recorded in [sigs.yaml].

### Chairs

* [Edward Haas](https://github.com/EdDev) (@EdDev), Red Hat
* [Itamar Holder](https://github.com/iholder101) (@iholder101), Red Hat

### Liaison

* [Daniel Hiller](https://github.com/dhiller) (@dhiller)

### Additional responsibilities of Chairs

* Welcome new contributors
* Resolve conflicts
* Keep this charter and the [sigs.yaml] entry current
* Coordinate with the chairs of affected SIGs before a quality standard changes how those SIGs review or structure code
* Give updates to stakeholder SIG chairs and to the broader community

## Decision making

Most business is conducted by lazy consensus, as described in [GOVERNANCE]. Proposals are raised in the working group meeting, on the [kubevirt-dev mailing list](https://groups.google.com/forum/#!forum/kubevirt-dev), or in a GitHub issue or pull request. Objections are raised in those same channels. Charter changes follow the review window in [wg-TEMPLATE/README.md](../wg-TEMPLATE/README.md): feedback is expected within a week, longer when the window falls over an event or a holiday.

A proposal that changes review practice, automation, or code structure inside a SIG's owned code is agreed with that SIG. This group recommends and coordinates. Approvers named in the owning SIG's OWNERS files accept the change.

When lazy consensus fails, the chairs record the disagreement and seek a resolution with the chairs of the affected SIGs. Questions that stay unresolved are escalated to the maintainers, who decide them as described in [GOVERNANCE].

Changes to this charter, and changes to chairs or the liaison, are made by pull request against this repository. The pull request is reviewed with the working group and the stakeholder SIGs, then sent to kubevirt-dev@googlegroups.com with the subject "WG Charter Proposal: code-quality". Maintainers ratify it by merging.

## Deliverables

The inter-SIG process is written down in one or more documents that define:

* the tooling and documentation required
* the code-structure metrics to establish, and how they are implemented
* how reviewer and maintainer awareness is achieved
* how and when the measures are executed

Those documents are delivered to the stakeholder SIGs and to the kubevirt-dev mailing list.

## Timelines and disbanding

The working group disbands when the chairs agree, by the decision process above, that the basic goal is met: documentation, tooling, and metrics for the targets in this charter are in place and in use. Disbanding is a charter change and follows the same pull-request process.

## Meeting Mechanics

* Where: Zoom meeting ID 93168599900
  * Link to join is [here](https://zoom.us/j/93168599900)
* When: Thursday at 09:00 UTC, every two weeks. See the [KubeVirt calendar](https://calendar.google.com/calendar/embed?src=kubevirt@cncf.io).
* [Meeting minutes](https://docs.google.com/document/d/1okUpVvBkzpocnJrIDnSRhDyTJJoM8v6qWqhgGRWu3bQ/edit) are sent to the kubevirt-dev mailing list:
  * mail-to: kubevirt-dev@googlegroups.com
  * Subject: [WG-code-quality] Meeting Notes <DATE>
* Slack: [#kubevirt-dev](https://kubernetes.slack.com/archives/C0163DT0R8X)

Meetings are open. Anyone may join and contribute.

## Agenda

* Discuss code-quality topics important to community members.
* Raise awareness of issues and pull requests that affect code quality across SIGs.
* Track progress on guidelines, automation, metrics, and the inter-SIG process.

[membership policy]: ../membership_policy.md
[sigs.yaml]: ../sigs.yaml
[GOVERNANCE]: ../GOVERNANCE.md
