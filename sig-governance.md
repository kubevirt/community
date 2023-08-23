# SIG Roles and Organizational Governance

This charter adheres to the conventions described in the
[sig-charter-guide]. It will be updated as needed to meet the current
needs of the KubeVirt project.

To standardize Special Interest Group efforts, create maximum transparency, and
route contributors to the appropriate SIG; SIGs should follow these guidelines:

- Have an approved Charter [sig-charter-guide].
- Meet regularly, at least for 30 minutes every 3 weeks, except November and
  December.
- Keep up-to-date meeting notes, linked from the SIG's page in the community
  repo.
- Participate in release planning meetings, retrospectives, and burndown
  meetings, as needed.
- Ensure related work happens in a project-owned GitHub org and repository,
  with code and tests explicitly owned and supported by the SIG, including
  issue triage, PR reviews, test-failure response, bug fixes, etc.
- Use the forums provided (mailing lists, slack etc) as the primary means of
  working, communicating, and collaborating, as opposed to private emails and
  meetings.
- Ensure contributing instructions are defined in the SIGs
  folder located in the KubeVirt/community repo if the groups contributor
  steps and experience are different or more in-depth than the documentation
  listed in the general [contributor guide].

The process for setting up a SIG is listed in the [sig-lifecycle] document.

## Roles

### Notes on Roles

Within this section, "Lead" refers to someone who is a member of the union of a
Chair, Tech Lead, or Subproject Owner role. Leads may, and frequently do, hold
more than one role. There is no singular lead to any KubeVirt community
group. Leads have specific decision-making power over some part of a group and
thus, additional accountability. Each role is detailed below.

Initial roles are defined at the founding of the SIG or Subproject as part of
the acceptance of that SIG or Subproject.

### Leads

#### Activity Expectations

- Leads _SHOULD_ remain active and responsive in their Roles.
- Leads taking an extended leave of 1 or more months _SHOULD_ coordinate with
  other leads to ensure the role is adequately staffed during their leave.
- Leads going on leave for 1-3 months _MAY_ work with other Leads to identify a
  temporary replacement.
- Leads _SHOULD_ remove any other leads that have not communicated a leave of
  absence and either cannot be reached for more than one month or are not
  fulfilling their documented responsibilities for more than one month.
  - Removal may be done through a [super-majority] vote of the active Leads.
  - If there is not enough _active_ Leads, then a [super-majority] vote
    between the remaining active Chairs, Tech Leads, and Subproject Owners may
    decide the removal of the Lead.

#### Requirements

- Leads _MUST_ be at least a ["member" on our contributor ladder] to be
  eligible to hold a leadership role within a SIG.
- SIGs _MAY_ prefer various levels of domain knowledge depending on the role.
  This should be documented.
- Interest or skills in people management.

#### Escalations

- Lead membership disagreements _MAY_ be escalated to the SIG Chairs. SIG Chair
  membership disagreements may be escalated to the Steering Committee.

#### On-boarding and Off-boarding Leads

- Leads _MAY_ decide to step down at anytime and propose a replacement. Use
  lazy consensus amongst the other Leads with fallback on majority vote to
  accept the proposal. The candidate _SHOULD_ be supported by a majority of
  SIG contributors or Subproject contributors (as applicable).
- Leads _MAY_ select additional leads through a [super-majority] vote amongst
  leads. This _SHOULD_ be supported by a majority of SIG contributors or
  Subproject contributors (as applicable).

#### Chair

- Number: 2+
- Membership tracked in [sigs.yaml]
- _SHOULD_ define how priorities and commitments are managed. _MAY_ delegate to
  other leads as needed.
- _SHOULD_ drive charter changes (including creation) to get community buy-in
  but _MAY_ delegate content creation to SIG contributors.
- _MUST_ in conjunction with the Tech Leads identify, track, and maintain the
  metadata of the SIGs enhancements for the current release and serve as
  point of contact for the release team, but _MAY_ delegate to other
  contributors to fulfill these responsibilities.
- _MAY_ delegate the creation of a SIG roadmap to other Leads.
- _MUST_ organize a main group meeting and make sure [sigs.yaml] is up to date,
  including subprojects and their meeting information, but _SHOULD_ delegate
  the need for subproject meetings to subproject owners.
- _SHOULD_ facilitate meetings but _MAY_ delegate to other Leads or future
  chairs/chairs in training.
- _MUST_ ensure there is a maintained CONTRIBUTING.md document in the
  appropriate SIG folder if the contributor experience or on-boarding knowledge
  is different than in the general [contributor guide]. _MAY_ delegate to
  contributors to create or update.
- _MUST_ organize KubeCon/CloudNativeCon Intros and Deep Dives with CNCF Event
  staff and approve presented content but _MAY_ delegate to other contributors.
- _MUST_ ensure meetings are recorded and made available.
- _MUST_ coordinate sponsored working group updates to the SIG and the wider
  community.
- _MUST_ coordinate communication and be a connector with other community
  groups like SIGs and the Steering Committee but _MAY_ delegate the actual
  communication and creation of content to other contributors where appropriate.

#### Tech Lead

- Number: 2+
- Membership tracked in [sigs.yaml]
- _MUST_ Approve & facilitate the creation of new subprojects
- _MUST_ Approve & facilitate decommissioning of existing subprojects
- _MUST_ Resolve cross-Subproject and cross-SIG technical issues and decisions
  or delegate to another Lead as needed
- _MUST_ in conjunction with the Chairs identify, track, and maintain the
  metadata of the SIGs enhancement proposals for the current release and serve
  as point of contact for the release team, but _MAY_ delegate to other
  contributors to fulfill these responsibilities
- _MUST_ Review & Approve SIG Enhancement Proposals, but _MAY_ delegate to
  other contributors to fulfill these responsibilities for individual proposals

#### Subproject Owner

- Number: 2+
- Scoped to a subproject defined in [sigs.yaml]
- Seed leads and contributors established at subproject founding
- _SHOULD_ be an escalation point for technical discussions and decisions in
  the subproject
- _SHOULD_ set milestone priorities or delegate this responsibility
- Membership tracked in [sigs.yaml] via links to OWNERS files

#### All Leads

- _SHOULD_ maintain health of their SIG or subproject
- _SHOULD_ show sustained contributions to at least one subproject or the the
  SIG at large.
- _SHOULD_ hold some documented role or responsibility in the SIG and / or at
  least one subproject (e.g. reviewer, approver, etc)
- _MAY_ build new functionality for subprojects
- _MAY_ participate in decision making for the subprojects they hold roles in

## Subprojects

### Subproject Creation

Subprojects may be created with a simple majority vote of SIG Technical Leads.

- [sigs.yaml] _MUST_ be updated to include subproject information and OWNERS
  files with subproject owners.
- Where subprojects processes differ from the SIG governance, they must
  document how. e.g. if subprojects release separately - they must document
  how release and planning is performed

### Subproject Requirements

Subprojects broadly fall into two categories, those that are directly part
of [KubeVirt] core and those that are tool, driver, or other component that
do not adhere to the KubeVirt release cycle.

#### KubeVirt Core Subprojects

- _MUST_ use the [design-proposal] process for introducing new features and decision
  making.
- _MUST_ Adhere to release test health requirements.

#### Non KubeVirt Core Subprojects\*\*

- _SHOULD_ define how releases are performed.
- _SHOULD_ setup and monitor test health.

Issues impacting multiple subprojects in the SIG should be resolved by the
SIG's Tech Leads or a federation of Subproject Owners.

## SIG Retirement

In the event that the SIG is unable to regularly establish consistent quorum or
otherwise fulfill its Organizational Management responsibilities

- after 3 or more months it _SHOULD_ be retired
- after 6 or more months it _MUST_ be retired

[KubeVirt]: https://github.com/kubevirt/
[sigs.yaml]: /sigs.yaml
[sig-charter-guide]: /sig-charter-guide.md
[sig-lifecycle]: /sig-lifecycle.md
[design-proposal]: design-proposals/README.md
["member" on our contributor ladder]: /membership_policy.md
[contributor guide]: https://kubevirt.io/user-guide/contributing/
[super-majority]: https://en.wikipedia.org/wiki/Supermajority#Two-thirds_vote
