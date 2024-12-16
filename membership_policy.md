# Community membership

This document outlines the various responsibilities of contributor roles in KubeVirt.

| Role                                           | Responsibilities                                                                                             | Requirements                                                                                                         | Defined by                           |
|------------------------------------------------|--------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| [Contributor](#new-contributors)               | Adhere to the [KubeVirt code of conduct](https://github.com/kubevirt/community/blob/main/code-of-conduct.md) | Submit at least one pull request                                                                                     | Active GitHub account                |
| [Member](#member)                              | Active contributor in the community                                                                          | * Active contributor<br>* Sponsored by 2 members<br>* Multiple contributions to the project                          | KubeVirt GitHub org member           |
| [Reviewer](#reviewer)                          | Review contributions from others                                                                             | * Member<br>* History of review and authorship in the project<br> * Sponsored by an approver                         | [OWNERS_ALIASES] file reviewer entry |
| [Approver](#approver)                          | Approve accepting contributions                                                                              | * Reviewer<br>* Highly experienced<br> * Active reviewer & contributor to the project<br> * Sponsored by 2 approvers | [OWNERS_ALIASES] file approver entry |
| [SIG Chair](#special-interest-group-sig-chair) | Lead a SIG aligned to the goals of the SIG charter                                                           | * Can be sig-approver<br>* Highly experienced in SIG matters<br> * Active reviewer & contributor to the project      | [sigs.yaml] chair entry              |
| [SIG Subproject Lead](#sig-subproject-lead)    | Lead a subproject aligned to the goals of the SIG charter                                                    | * sig-reviewer<br>* Highly experienced in SIG subproject matters<br> * Active reviewer & contributor to the project  | [sigs.yaml] subproject leads entry   |
| [WG Chair](#working-group-wg-chair)            | Lead a WG aligned to the goals of the WG charter                                                             | * Highly experienced in WG matters<br> * Active reviewer & contributor to the project                                | [sigs.yaml] WG chairs entry          |
| [Project Maintainers](./GOVERNANCE.md)         | Steer the project from the highest level                                                                     | * Demonstrated leadership in the community<br>  See the [Governance doc](./GOVERNANCE.md) for more details | [MAINTAINERS](./MAINTAINERS.md) file |

## New contributors

[New contributors](https://github.com/kubevirt/kubevirt/blob/main/CONTRIBUTING.md) should be welcomed to the community by existing members, helped with PR workflow, and directed to relevant documentation and communication channels.
All contributors are expected to adhere to the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Established community members

Established community members are expected to demonstrate their adherence to the principles in this document, familiarity with project organization, roles, policies, procedures, conventions, etc., and technical and/or writing ability. Role-specific expectations, responsibilities, and requirements are enumerated below.

## Member

Members are continuously active contributors in the community. They can have issues and PRs assigned to them, participate in community related meetings (e.g. bug scrub), and pre-submit tests are automatically run for their PRs. Members are expected to remain active contributors to the community.
Defined by: Member of the KubeVirt GitHub organization

### Requirements

  * Enabled [two-factor authentication](https://help.github.com/articles/about-two-factor-authentication) on their GitHub account
  * Actively contributing to the project with multiple contributions to the project or community. Contribution may include, but is not limited to:
    * Authoring or reviewing PRs on GitHub
    * Filing or commenting on issues on GitHub
    * Contributing to community discussions (e.g. meetings, Slack, email discussion forums, Stack Overflow)
    * GitHub Insight metrics will be reviewed periodically
  * Subscribed to kubevirt-dev@googlegroups.com
  * Has read the [contributor guide](contributors/contributing.md)
  * Actively contributing to the project
    * Contributions according to GitHub and/or Kubernetes metrics in the past 90 days.
  * Sponsored by 2 org members. Note the following requirements for sponsors:
    * Sponsors must have close interactions with the prospective member
    * Sponsors must be reviewers or approvers in any of the OWNERS files found in KubeVirt's repositories, for example, [this](https://github.com/kubevirt/kubevirt/blob/main/OWNERS_ALIASES) or [this one](https://github.com/kubevirt/project-infra/blob/main/OWNERS).
    * Sponsors should be from multiple member companies to demonstrate integration across community.
  * Open a PR against the [org members section](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml#L20)
    * Ensure your sponsors are @mentioned on the issue
    * Complete every item on the [checklist](membership_checklist.md)
    * Make sure that the list of contributions included is representative of your work on the project.
  * Have your sponsoring org members reply confirmation of sponsorship: +1
  * Once your sponsors have responded, your request will be reviewed by KubeVirt project approvers. Any missing information will be requested.

> [!IMPORTANT]
> After the pull request against the org members section has been merged (see above), an invitation to the KubeVirt GitHub organization will be automatically sent to the new member. The new member needs to accept the invitation to receive member status.

### Responsibilities and privileges

  * Responsive to issues and PRs assigned to them
  * Responsive to mentions
  * Active owner their contributions (unless ownership is explicitly transferred)
  * Contribution is well tested
  * Tests consistently pass
  * Addresses bugs or issues discovered after contribution is accepted
  * Members can do /lgtm on open PRs.
  * They can be assigned to issues and PRs, and people can ask members for reviews with a /cc @username.
  * Tests can be run against their PRs automatically. No /ok-to-test needed.
  * Members can do /ok-to-test for PRs that have a needs-ok-to-test label, and use commands like /close to close PRs as well.

!! note
Members who frequently contribute are expected to proactively perform reviews and work towards becoming a primary reviewer.

## Reviewer

Reviewers are trusted to review contributions for quality and correctness on some part of the project. They are knowledgeable about both the project and engineering principles.
Defined by: reviewers entry in an OWNERS file of the KubeVirt project.
Note: Acceptance of contributions requires at least one approver in addition to the assigned reviewers.
Reviewer Status is scoped to a part of the project.

### Requirements

The following apply to the part of project for which one would be a reviewer in the [OWNERS_ALIASES] file.

  * Member for at least 3 months
  * Primary reviewer for at least 5 PRs to the project
  * Reviewed or merged at least 20 substantial PRs to the project
  * Knowledgeable about the project
  * Sponsored by an approver
    * With no objections from other approvers
    * Done through PR to update the OWNERS file
  * May either be self-nominate or nominated by an approver

### Responsibilities and privileges

  * The following apply to the part of project for which one would be a reviewer in the [OWNERS_ALIASES] file.
  * Tests are automatically run for PullRequests from members of the KubeVirt GitHub organization
  * Reviewer status may be a precondition to accepting L+ contributions
  * Responsible for project quality control via reviews
  * Focus on quality and correctness, including testing and factoring
  * May also review for more holistic issues, but not a requirement
  * Expected to be responsive to review requests as per community expectations
  * Assigned PRs to review related to project
  * Assigned test bugs related to project
  * Granted "read access" to KubeVirt repo
  * May get a badge on PR and issue comments

## Approver

Approvers are able to both review and approve contributions. While contribution review is focused on quality and correctness, approval is focused on holistic acceptance of a contribution including: backwards / forwards compatibility, adhering to API and flag conventions, subtle performance and correctness issues, interactions with other parts of the system, etc.
Defined by: Approvers entry in an OWNERS file in a repo owned by the KubeVirt project.
Approver Status is scoped to a part of the project.

### Requirements
The following apply to the part of project for which one would be an approver in the [OWNERS_ALIASES] file.

  * Reviewer of the project for at least 3 months
  * Primary reviewer for at least 10 substantial PRs to the project
  * Reviewed or merged at least 30 PRs to the project
  * Nominated by an approver
    * With no objections from other approvers
    * Done through PR to update the top-level OWNERS file

### Responsibilities and privileges

The following apply to the part of project for which one would be an approver in the OWNERS file (for repos using the bot).

  * Approver status may be a precondition to accepting contributions
  * Demonstrate sound technical judgement
  * Responsible for project quality control via reviews
  * Focus on holistic acceptance of contribution such as dependencies with other features, backwards / forwards compatibility, API and flag definitions, etc
  * Expected to be responsive to review requests as per community expectations
  * Mentor contributors and reviewers
  * May approve contributions for acceptance

## Special Interest Group (SIG) Chair

The KubeVirt project is organized primarily into SIGs, each with a common purpose of advancing the project with respect to a specific topic, such as Networking or Storage. A SIG Chair is required to help guide and coordinate the SIG. 

### Requirements

* Active reviewer for SIG matters for at least 3 months
* Track record of contributions to SIG matters
* Advanced knowledge in SIG subject

### Responsibilities and privileges

* Is an organizer, i.e.
  * Ensures that SIG meetings are moderated
  * Ensures that meeting notes are captured
* Is a facilitator, i.e.
  * Advances SIG topics
  * Approves & facilitates the creation of new subprojects
* Operates the SIG
  * Updates charter when required
  * Updates sigs.yaml SIG entry
* Communicates and coordinates:
  * with sponsored WGs
  * with other SIGs
  * to the broader community
* Claims approver rights in scope, i.e. by
  * Creating a sig-approver group in `OWNERS_ALIASES`,
  * Adding a reference to the above in an `OWNERS` file 
  * Adding an `owners` reference in sigs.yaml

## SIG Subproject Lead

Specific work efforts within SIGs can be divided into subprojects. A SIG Subproject Lead is required to help guide and coordinate the subproject.  

### Requirements

* Active reviewer for subproject matters for at least 3 months
* Track record of contributions to subproject matters
* Advanced knowledge in subproject scope

### Responsibilities and privileges

* Creates and maintains a subproject
* Leads within subproject scope, i.e.
  * Sets technical direction,
  * Makes design decisions
  * Mentors other contributors
  * Moderates technical discussions and decisions
* Claims approver rights in scope, i.e. by
  * Creating a subproject-approver group in `OWNERS_ALIASES`
  * Adding a reference to the above in an `OWNERS` file
  * Adding an `owners` reference in sigs.yaml

## Working Group (WG) Chair

Working groups are primarily used to facilitate topics of discussion that cross SIG lines. Working Group Chairs are required to help guide the working group and coordinate with SIG Chairs and the broader community.

### Requirements

* Active reviewer for WG matters for at least 3 months
* Track record of contributions to WG matters
* Advanced knowledge in WG scope

### Responsibilities and privileges

* Is an organizer, i.e.
  * Ensures that WG meetings are moderated
  * Ensures that meeting notes are captured
* Is a facilitator, i.e.
  * Advances exploration of the WG scope
  * Fosters collaboration across SIG boundaries related to WG scope
* Operates the WG
  * Updates charter when required
  * Updates sigs.yaml WG entry
* Communicates and coordinates
  * Gives updates to respective sponsoring SIG Chairs
  * The broader community

## Inactive members

[_Members are continuously active contributors in the community._](#member)

A core principle in maintaining a healthy community is encouraging active
participation. It is inevitable that people's focuses will change over time, and
they are not expected to be actively contributing forever.

However, being a member of the KubeVirt GitHub organization comes with
an elevated set of permissions. These capabilities should not be used by those
that are not familiar with the current state of the KubeVirt project.

Therefore, members with an extended period away from the project with no activity
will be removed from the KubeVirt GitHub Organization and will be required to
go through the org membership process again after re-familiarizing themselves
with the current state.

### How inactivity is measured

Inactive members are defined as members of the KubeVirt Organization
with **no** contributions within the KubeVirt organization within 12 months. 
This is measured by the [CNCF DevStats project].

**Note:** Devstats does not take into account non-code contributions. If a
non-code contributing member is accidentally removed this way, they may open an
issue to quickly be re-instated.

After an extended period away from the project with no activity
those members would need to re-familiarize themselves with the current state
before being able to contribute effectively.

[CNCF DevStats project]: https://kubevirt.devstats.cncf.io/d/9/developer-activity-counts-by-repository-group-table?orgId=1&var-period_name=Last%20year&var-metric=contributions&var-repogroup_name=All&var-country_name=All
[OWNERS_ALIASES]: https://github.com/kubevirt/kubevirt/tree/main/OWNERS_ALIASES
[sigs.yaml]: ./sigs.yaml
