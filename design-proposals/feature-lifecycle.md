# KubeVirt Feature Lifecycle

## Summary

KubeVirt requires a clear policy on how features are introduced,
evaluated and finally graduated or removed.

This proposal defines the steps and policies to follow in order
to manage a feature and its lifecycle in KubeVirt.

The proposal is focusing on introducing features in
a stable API (CRD) version, e.g. `kubevirt.io/v1`.

## Overview

### Motivation
KubeVirt has grown into a matured virtualization management solution
with a large set of features.

New features are being proposed and added regularly to its portfolio.

With time, the challenge of supporting and maintaining such a large
set of features raised the need to re-examine their relevance.
It also raised the need to examine with more care features graduation.

The KubeVirt community has tried to control the flow of features
informally through feature-gates, similar to Kubernetes.
However, as time passed, several challenges presented themselves:
- Evaluated features rarely got graduated to GA or removed, causing
  feature consumption to be risky for users and a maintenance burden
  for the project contributors.
  > **Note**: As of this writing (pre v1.2), there are 37 FGs,
  > out of which 5 GA-ed and 2 marked for deprecation.
  > For more information, explore the
  > [source](https://github.com/kubevirt/kubevirt/blob/5ff12ae931cefd81514ec96f97a189ab2c179ad7/pkg/virt-config/feature-gates.go).
- We do not have agreed-upon procedures on how and when features
  graduate or discontinue.
  This causes each feature to take different approaches, possibly
  surprising users.

We conclude that the current use of FGs is insufficient
due to the lack of well-defined processes and policy on how a feature
should progress in its lifecycle.

Eventually we would like to see features being evaluated carefully
before they are introduced, while they are experimented with and
proven to be actually in use (and useful) before graduating.

> **Note**: Once a feature graduates, it is included in a
> General Availability (GA) release with its functionality available
> to all users. GA features need to comply with [semver](https://semver.org/)
> which add constraints on their ability to change (including deprecation).

### Goals
- Define the process a feature needs to pass in order to be
  Generally Available.
- Define the process a feature needs to pass in order to be removed.
- Provide policies and rules on how to manage a feature during its
  lifetime.

### Non Goals
- Implement enforcement tooling to keep features in sync with
  the lifecycle rules.

### Definition Of Users
- Development contributors.
- Cluster operators.

### User Stories
- As a KubeVirt contributor, I would like to introduce a new useful
  feature and follow it to graduation (GA).
- As a KubeVirt contributor, I would like to remove a feature
  that has not yet reached its formal graduation.
- As a KubeVirt contributor, I would like to remove a feature
  that has already graduated.
  > **Note**: Removal of a GA feature is considered an exception.
  > Strong arguments and a wide agreement is required for such an action.
- As a KubeVirt cluster operator, I would like to experiment with
  a newly-proposed ("Alpha") feature in a controlled environment, to see
  if it makes sense to me.
- As a KubeVirt cluster operator, I would like to evaluate an
  undergraduate ("Beta") feature in a real-life environment with actual users.
- As a KubeVirt cluster operator, I would like to keep using a feature
  that got graduated after I used it during the "Beta" evaluation period.
- As a KubeVirt cluster operator, I would like to know that
  an experimental ("Alpha") feature is planned to be removed.
- As a KubeVirt cluster operator, I would like to know that
  an undergraduate ("Beta") feature is planned to be removed.
- As a KubeVirt cluster operator, I would like to stop using a feature
  that got removed.

### Repos
This is a cross repo project policy under the
[kubevirt](https://github.com/kubevirt) organization.

## Proposal Design
The proposal on how to define a feature lifecycle is influenced by
processes and policies from the Kubernetes project.
These sources are scattered around, each focusing on different
aspects of a feature:
- [Feature Gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/)
- [Graduation](https://kubernetes.io/blog/2020/08/21/moving-forward-from-beta/)
- [Changing the API](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md)
- [Deprecation](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)

The proposal takes the top-down approach, starting with the high level
flow that a common feature will traverse through.
Continuing with actions that need to be taken and timeline suggestions.

Both feature graduation and discontinuation flows are covered.
Including the implications on users.

Depending on individual topics, follow-up proposals may extend the basic
points raised in this proposal.

### Terminology
Feature Gates and feature configuration are often used interchangeable or not differentiated.
- Feature Gate: A flag that controls the presence or availability of a feature in the cluster.
- Feature configuration: Cluster or workload level configuration that allows an admin or user
  (depending on the feature) to control aspects of a feature operation.
  A common usage is to determine if features are opt-in or opt-out by default.

### Feature Stages
A feature is expected to pass in the following order through the following stages:
1. Enhancement proposal.
2. Implementation.
3. Release as Alpha (experimental).
4. Release as Beta (pre-release for evaluation).
5. Release as General Availability (graduation).
6. Removal.

Starting from the Alpha release, it can be removed with restrictions that
depend on the release stage (Alpha, Beta, GA).

[Removal](#removal) of features is widely discussed later
in this proposal.

#### Enhancement proposal
As the first step for introducing a new feature, a formal proposal is
expected to be shared for public review via mailinglist and
a [design proposal](https://github.com/kubevirt/community/tree/main/design-proposals).

This is the first opportunity to evaluate a new feature.
The proposal needs to include motivation, goals, implementation details
and phases. Review the [proposal template](https://github.com/kubevirt/community/blob/main/design-proposals/proposal-template.md)
for more information.

#### Implementation
The development work on the feature is expected to include coding,
testing, integration and documentation.

#### Releases
- **Alpha**:
  An initial release of the feature for experimental purposes.
  Recommended for non-production usages, evaluation or testing.

  The API is considered unstable and may change significantly.
  There are no backward compatability considerations and it can
  be removed at any time.

  The period in which a feature can remain in Alpha is limited,
  assuring features are not piling up without control.
  See [release stage transition table](#release-stage-transition-table)
  for more information.

  The feature presence is controlled using a Feature-Gate (FG) during
  runtime. It must be specified for the feature to be active.

- **Beta**:
  The first release that can be evaluated with care in production.
  Acting as a pre-release, its main objective is to collect feedback
  from users to assure its usefulness and readiness for graduation.
  If there is no confidence of usage or usefulness, it may remain in
  this stage for some time.

  However, the period in which a feature can remain in Beta is limited,
  assuring features are not piling up without control.
  See [release stage transition table](#release-stage-transition-table)
  for more information.

  The API is considered stable with care not to break backward compatibility
  with previous beta releases.
  This implies that fields may only be added during this stage,
  not removed or renamed.

  The feature presence is controlled using a Feature-Gate (FG) during
  runtime. It must be specified for the feature to be active.

- **GA**:
  The feature graduated to general-availability (GA) and is now part of
  the core features.

  The API is considered stable with care not to break backward compatibility
  with the previous releases.

  The feature functionality is no longer controlled by a FG.

> **Warning**: A Feature Gate flag is solely intended to control
> feature lifecycle. It should not be confused and used as a cluster
> configurable enablement of the functionality.
> In cases where the cluster admin should control a functionality,
> regardless to the feature stage, dedicated configuration field/s
> should be included.

#### Removal
If a feature is targeted for deprecation and retirement,
it needs to pass a deprecation process, depending on its current
release stage (Alpha, Beta, GA).

For more details, see [here](#deprecation-and-removal).

#### Release Stage Transition Table
The following table summarized the different release stages with their
transition requirements and restrictions.

| Stage          | Period range    | F.Gate | Removal Availability       |
|----------------|-----------------|--------|----------------------------|
| Alpha          | 1 to 2 releases | YES    | Between **minor** releases |
| Beta           | 1 to 3 releases | YES    | Between **minor** releases |
| GA             | -               | NO     | Between **major** releases |

Through Alpha and Beta feature releases, a FG must be set in order
for the feature to function.
By default, no FG is specified, therefore the feature is disabled.

If a feature is not able to transition to the next stage in the defined period,
it should be removed automatically.

> **Note**: Exceptions to the period range may apply
> if 2/3 of active maintainers come to agreement to prolong
> a specific feature.

### Deprecation and Removal
One reason for features to go through the Alpha and Beta stages,
is the opportunity to examine their usefulness and adoption.
Same goes with major releases that intentionally allow breaking
backward compatibility (as specified by [semver](https://semver.org/)).

Therefore, it is only natural that some features will not graduate
between the stages, or will be found irrelevant after some time and be
removed when transitioning between major releases.

#### Major Releases
KubeVirt follows semver versioning, in which major versions may
break API compatibility. Therefore, discontinuation of features
is somehow simpler when incrementing the major version.

However, this is not without a cost.
When a new major release is introduced, the previous one is still maintained
and supported, something that does not exist with minor releases.

#### The Deprecation Flow (for Minor releases)
Only Alpha and Beta features can be removed during a minor release.

These are the steps needed to deprecate & remove a feature:
- Proposal: Prepare a proposal to remove a feature with proper
  reasoning, phases, exact timelines and functional alternatives (if any).
  The proposal should be reviewed and approved.
- Notification: Notify the project community about the feature
  discontinuation based on the approved proposal.
  All details of the plan should be provided to allow users and possibly
  down-stream projects to adjust.
  Use all community media options to distribute this information
  (e.g. mailing list, slack channel, community meetings).
- Deprecation warnings: Add deprecation warnings at runtime to warn users
  that the feature is planned to be removed.
  Warnings should be raised when:
  - Feature API fields are accessed.
  - Feature FG is activated.
  - Behavior related to the feature is detected (optional).
- Removal: Feature removal involves removing the core functionality
  of a feature and its exposed API.
  - The core implementation can be removed in two steps:
    - The FG is removed by assuring it is never reported as set
      (i.e. even if it is left by the operator configured, internally
       it is ignored).
      At this stage, the core implementation will follow the FG conditions
      and therefore from the outside the feature is inactive.
    - In case there are no side effects, the core implementation code can
      be removed.
  - The API types are not to be removed, as it may have implications
    with the underlying storage which has already persisted them.
    Kubernetes has not removed fields, it just kept them around with the
    warning that they have been deprecated and no longer available.

    While keeping fields around for a period of a release or two makes
    sense, beyond a limited period it adds a burden on dragging leftover
    fields around to eternity.

> **Note**: The only reference seen on why fields should not be removed
> was mentioned [here](https://github.com/kubernetes/kubernetes/issues/52185).
> But it is unclear if this is relevant for Alpha stage features.
> Starting with a strict policy, similar to Kubernetes is recommended,
> i.e. once fields are introduced, they should not be removed no matter
> the feature release stage.
> Per need, the topic can be revisited in follow-up adjustments.

### Exceptions
While the project strives to maintain a stable contract with its users,
there may be scenarios where the policy described here will not be a fit.

Therefore, it should be acceptable to have exceptions from time to time
given a very good reasoning and an agreement from 2/3 of the project
maintainers (also known as "approvers").

## Implementation Phases
- Add a section in the
  [design proposal template](https://github.com/kubevirt/community/blob/main/design-proposals/proposal-template.md)
  that describes the planned timelines for the feature stages.
- Add a reference to the feature-lifecycle documentation to assure contributors
  know the process and policy.
- Prepare a user-facing document that describes the usability implications of
  this feature lifecycle.

## Miscellaneous

- New features are to be introduced to major and minor release versions only.
  For clarification, this implies that new features are **not** to be backport.
- CI:
  - Alpha stage features should not be gating on CI.
  - Beta and GA features should be gating on CI.
- API fields may be marked with the following information:
  - Description
  - Release stage (alpha/beta/ga) and release version.

  It is left to a follow-up implementation proposal to define the exact format and
  required/optional information.
