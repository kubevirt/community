# SIG Governance Requirements

## Goals

This document outlines the recommendations and requirements for defining SIG and subproject governance.

This doc uses [rfc2119](https://www.ietf.org/rfc/rfc2119.txt) to indicate keyword requirement levels.
Sub elements of a list inherit the requirements of the parent by default unless overridden.

## Checklist

Following is the checklist of items that should be considered as part of defining governance for
any subarea of the KubeVirt project.

### Roles

- _MUST_ enumerate any roles within the SIG and the responsibilities of each
- _MUST_ define process for changing the membership of roles
  - When and how new members are chosen / added to each role
  - When and how existing members are retired from each role
- _SHOULD_ define restrictions / requirements for membership of roles
- _MAY_ define target staffing numbers of roles

### Organizational management

- _MUST_ define when and how collaboration between members of the SIG is organized

  - _SHOULD_ define how periodic video conference meetings are arranged and run
  - _SHOULD_ define how conference / summit sessions are arranged
  - _MAY_ define periodic office hours on slack or video conference

- _MAY_ define process for new community members to contribute to the area

  - e.g. read a contributing guide, show up at SIG meeting, message the google group

- _MUST_ define how subprojects are managed
  - When and how new subprojects are created
  - Subprojects _MUST_ define roles (and membership) within subprojects

### Project management

The following checklist applies to both SIGs and subprojects of SIGs as appropriate:

- _MUST_ define how milestones / releases are set

  - How target dates for milestones / releases are proposed and accepted
  - What priorities are targeted for milestones
  - The process for publishing a release

- _SHOULD_ define how priorities / commitments are managed
  - How priorities are determined
  - How priorities are staffed

### Technical processes

All technical assets _MUST_ be owned by exactly 1 SIG subproject. The following checklist applies to subprojects:

- _MUST_ define how technical decisions are communicated and made within the SIG or project

  - Process for proposal, where and how it is published and discussed, when and how a decision is made
  - Who are the decision makers on proposals (e.g. anyone in the world can block, just reviewers on the PR,
    just approvers in OWNERs, etc)
  - How disagreements are resolved within the area (e.g. discussion, fallback on voting, escalation, etc)
  - How and when disagreements may be escalated
  - _SHOULD_ define expectations and recommendations for proposal process (e.g. escalate if not progress towards
    resolution in 2 weeks)
  - _SHOULD_ define a level of commitment for decisions that have gone through the formal process
    (e.g. when is a decision revisited or reversed)

- _MUST_ define how technical assets of project remain healthy and can be released
  - Publicly published signals used to determine if code is in a healthy and releasable state
  - Commitment and process to _only_ release when signals say code is releasable
  - Commitment and process to ensure assets are in a releasable state for milestones / releases
    coordinated across multiple areas / subprojects (e.g. the KubeVirt OSS release)
  - _SHOULD_ define target metrics for health signal (e.g. broken tests fixed within N days)
  - _SHOULD_ define process for meeting target metrics (e.g. all tests run as presubmit, build cop, etc)
