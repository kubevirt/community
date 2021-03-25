# Community membership

This document outlines the various responsibilities of contributor roles in KubeVirt.

| Role  | Responsibilities | Requirements | Defined by
| ----- | ---------------- | ------------ | ----------
| Contributor | Adhere to the [KubeVirt code of conduct](https://github.com/kubevirt/community/blob/master/code-of-conduct.md) | Submit at least one pull request | Active GitHub account
| Member | Active contributor in the community | * Active contributor<br>* Sponsored by 2 members<br>* Multiple contributions to the project | KubeVirt GitHub org member
| Reviewer | Review contributions from others | * Member<br>* History of review and authorship in the project<br> * Sponsored by an approver | [OWNERS](https://github.com/kubevirt/kubevirt/blob/master/OWNERS_ALIASES) file reviewer entry |
| Approver| Approve accepting contributions | * Reviewer<br>* Highly experienced<br> * Active reviewer & contributor to the project<br>Sponsored by 2 approvers| [OWNERS](https://github.com/kubevirt/kubevirt/blob/master/OWNERS_ALIASES) file approver entry |


## New contributors

[New contributors](https://github.com/kubevirt/kubevirt/blob/master/CONTRIBUTING.md) should be welcomed to the community by existing members, helped with PR workflow, and directed to relevant documentation and communication channels.
All contributors are expected to adhere to the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Established community members

Established community members are expected to demonstrate their adherence to the principles in this document, familiarity with project organization, roles, policies, procedures, conventions, etc., and technical and/or writing ability. Role-specific expectations, responsibilities, and requirements are enumerated below.

## Member

Members are active contributors in the community. They can have issues and PRs assigned to them, participate in community related meetings (e.g. bug scrub), and pre-submit tests are automatically run for their PRs. Members are expected to remain active contributors to the community.
Defined by: Member of the KubeVirt GitHub organization

### Requirements

  * Enabled [two-factor authentication](https://help.github.com/articles/about-two-factor-authentication) on their GitHub account
  * Actively contributing to the project with multiple contributions to the project or community. Contribution may include, but is not limited to:
    * Authoring or reviewing PRs on GitHub
    * Filing or commenting on issues on GitHub
    * Contributing to community discussions (e.g. meetings, Slack, email discussion forums, Stack Overflow)
    * GitHub Insight metrics will be reviewed periodically
  * Subscribed to kubevirt-dev@googlegroups.com
  * Has read the [contributor guide](CONTRIBUTING.md)
  * Actively contributing to the project
    * Contributions according to GitHub and/or Kubernetes metrics in the past 90 days.
  * Sponsored by 2 org members. Note the following requirements for sponsors:
    * Sponsors must have close interactions with the prospective member
    * Sponsors must be reviewers or approvers in the KubeVirt OWNERS file.
    * Sponsors should be from multiple member companies to demonstrate integration across community.
  * Open a PR against the [org members section](https://github.com/kubevirt/project-infra/blob/master/github/ci/prow/files/orgs.yaml#L21)
    * Ensure your sponsors are @mentioned on the issue
    * Complete every item on the [checklist](membership_checklist.md)
    * Make sure that the list of contributions included is representative of your work on the project.
  * Have your sponsoring org members reply confirmation of sponsorship: +1
  * Once your sponsors have responded, your request will be reviewed by KubeVirt project approvers. Any missing information will be requested.

## Responsibilities and privileges

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

## Requirements

The following apply to the part of project for which one would be a reviewer in the [OWNERS](https://github.com/kubevirt/kubevirt/blob/master/OWNERS_ALIASES) file.

  * Member for at least 3 months
  * Primary reviewer for at least 5 PRs to the project
  * Reviewed or merged at least 20 substantial PRs to the project
  * Knowledgeable about the project
  * Sponsored by an approver
    * With no objections from other approvers
    * Done through PR to update the OWNERS file
  * May either be self-nominate or nominated by an approver

### Responsibilities and privileges

  * The following apply to the part of project for which one would be a reviewer in the [OWNERS](https://github.com/kubevirt/kubevirt/blob/master/OWNERS_ALIASES) file.
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
The following apply to the part of project for which one would be an approver in the [OWNERS](https://github.com/kubevirt/kubevirt/blob/master/OWNERS_ALIASES) file.

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
