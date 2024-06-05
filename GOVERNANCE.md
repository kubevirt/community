# Project Governance

- [Maintainers](#maintainers)
  - [Becoming a Maintainer](#becoming-a-maintainer)
  - [Meetings](#meetings)
  - [CNCF Resources](#cncf-resources)
  - [Code of Conduct Enforcement](#code-of-conduct)
  - [Voting](#voting)
- [SIGs (Special Interest Groups)](#sigs)
  - [SIG Subprojects](#subprojects)
- [WGs (Working Groups)](#working-groups)

## Maintainers

KubeVirt Maintainers govern the project. Maintainers collectively manage the 
project's resources and contributors, and speak for the project in public.  The
maintainers collectively decide any questions that cannot be resolved at the 
individual repository level, and provide strategic guidance for the project
overall.

The current maintainers can be found in [MAINTAINERS](./MAINTAINERS.md).  

This privilege is granted with some expectation of responsibility: maintainers
are people who care about the KubeVirt project and want to help it grow and
improve. A maintainer is not just someone who can make changes, but someone who
has demonstrated their ability to collaborate with the team, get the most
knowledgeable people to review code and docs, contribute high-quality code, and
follow through to fix issues.

A maintainer is a contributor to the project's success and a citizen helping
the project succeed.

### Selecting Maintainers

The current project maintainers will periodically review contributor activities
to see if additional project members may be promoted to maintainers. 

For nominations, the maintainers will look at the following criteria:

  * Commitment to the project: have they participated in discussions, 
    contributions, and reviews for 1 year or more?
  * Does the person show leadership in one of these areas?
    * Active approver or reviewer in core or subprojects
    * SIG leadership
    * Mentoring other project contributors
  * Does the candidate bring new perspectives or community connections to the 
    maintainers?
  * Do they understand how the project works (policies, processes, etc)?
  * Are they willing to take on the additional duties of a maintainer?

A candidate must be proposed by an existing maintainer by filing an PR in the
[Community Repo](https://github.com/kubevirt/community) against the MAINTAINERS.md file. 
A simple majority vote of +1s from existing Maintainers approves the application. 
Approved maintainers will be added to the [private maintainer mailing list](mailto:cncf-kubevirt-maintainers@lists.cncf.io).

### Meetings

Time zones permitting, Maintainers are expected to participate in the weekly public
community meeting. More details can be found [here](commnuity_meeting.md).

Maintainers will also have closed meetings in order to discuss security reports
or Code of Conduct violations.  Such meetings should be scheduled by any
Maintainer on receipt of a security issue or CoC report.  All current Maintainers
must be invited to such closed meetings, except for any Maintainer who is
accused of a CoC violation.

### CNCF Resources

Any Maintainer may suggest a request for CNCF resources, in the
[developer mailing list](https://groups.google.com/forum/#!forum/kubevirt-dev), 
the [Maintainer mailing list](mailto:cncf-kubevirt-maintainers@lists.cncf.io), on Github, 
or during a community meeting.  A simple majority of Maintainers approves the 
request.  The Maintainers may also choose to delegate working with the CNCF to 
non-Maintainer community members.

### Code of Conduct

[Code of Conduct](./code-of-conduct.md)
violations by community members will be discussed and resolved
on the [private Maintainer mailing list](mailto:cncf-kubevirt-maintainers@lists.cncf.io).  If the reported CoC violator
is a Maintainer, the Maintainers will instead designate two Maintainers to work
with CNCF staff in resolving the report.

### Removing Maintainers

Maintainers may voluntarily retire at any time.  Should a maintainer retire, 
it requires a majority vote of the current maintainers to reinstate them.

Maintainers may also be demoted at any time for one of the following reasons:

* Inactivity, including 6 months or more of non-participation or non-communication,
* Refusal to abide by this Governance,
* Violations of the Code of Conduct,
* Other actions that harm the reputation, stability, or harmony of the Kubevirt
  project.

Removing a maintainer requires a 2/3 majority vote of the other maintainers.

### Voting

While most business in KubeVirt is conducted by "lazy consensus", periodically
the Maintainers may need to vote on specific actions or changes.
A vote can be taken on [the developer mailing list](https://groups.google.com/forum/#!forum/kubevirt-dev) or
the private Maintainer mailing list for security or conduct matters.  
Votes may also be taken at the community meeting.  Any Maintainer may
request a vote be taken.

Most votes require a simple majority of all Maintainers to succeed. Maintainers
can be removed by a 2/3 majority vote of all Maintainers, and changes to this
Governance require a 2/3 vote of all Maintainers.

## SIGs

The KubeVirt project is organized primarily into Special Interest Groups, or SIGs.
Each SIG is comprised of members from multiple companies and organizations, with a
common purpose of advancing the project with respect to a specific topic, such as
Networking or Storage. Our goal is to enable a distributed decision structure
and code ownership, as well as providing focused forums for getting work done,
making decisions, and onboarding new contributors. Every identifiable subpart of
the project (e.g., github org, repository, subdirectory, API, test, issue, PR)
is intended to be owned by some SIG.

Our SIGs define mission and scope via their charters, ownership of code via
[OWNERS](https://www.kubernetes.dev/docs/guide/owners/), and members and roles in
[sigs.yaml].

For more details see also the [Kubernetes SIGs] model.

Examples:
* [sig-network - charter](./sig-network/charter.md)
* [sig-network - kubevirt/kubevirt OWNERS_ALIASES](https://github.com/kubevirt/kubevirt/blob/a7e0311d8704663351abd4bc9bbc8511753d2838/OWNERS_ALIASES#L60)
* [sig-ci - sigs.yaml](https://github.com/kubevirt/community/blob/4f63a79c0ed810aa332cd6716d4986001d28bcd7/sigs.yaml#L119)

### Subprojects

Specific work efforts within SIGs are divided into subprojects. Every part of the
KubeVirt code and documentation must be owned by some subproject. Some [SIGs](#sigs)
may have a single subproject, but many SIGs have multiple significant subprojects
with distinct (though sometimes overlapping) sets of contributors and owners, who
act as subproject’s technical leaders: responsible for vision and direction and
overall design, choose/approve design proposals approvers, field technical
escalations, etc.

For more details see also the [Kubernetes Subprojects] model.

## Working Groups

We need community rallying points to facilitate discussions/work regarding topics
that are short-lived or that span multiple SIGs.

Working groups are primarily used to facilitate topics of discussion that are in
scope for KubeVirt but that cross SIG lines.

 Our WGs define mission and scope via their charter, and members and roles in
 [sigs.yaml].

For more details see also the [Kubernetes WGs] model.

## Roles and Organization Management

All our SIGs and WGs follow the Roles and Organization Management outlined in [OARP]

[Kubernetes SIGs]: https://github.com/kubernetes/community/blob/master/governance.md#sigs
[Kubernetes Subprojects]: https://github.com/kubernetes/community/blob/master/governance.md#subprojects
[Kubernetes WGs]: https://github.com/kubernetes/community/blob/master/governance.md#working-groups
[OARP]: https://stumblingabout.com/tag/oarp/
[sigs.yaml]: ./sigs.yaml
