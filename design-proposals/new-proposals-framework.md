# Overview

This proposal outlines a new process for managing KubeVirt Enhancement Proposals (VEPs), emphasizing centralized prioritization and enhanced SIG involvement and collaboration. The proposed changes aim to focus the community's efforts on prioritized pull requests, increase the review bandwidth, and ensure clear visibility of feature progress and associated issues.

## Motivation

KubeVirt’s current approach to proposals is unstructured and has outgrown the community repo, making it difficult to adequately manage both the repo and the design proposals. The current process does not provide a way to prioritize development and review efforts during the development cycle, nor is there a way to derive a roadmap for a specific release. The process also lacks any formal SIG involvement. Being in the community repo also makes it difficult to grow the reviewer and approver list as there are conflicting purposes.


With this proposal, we aim to:

- Optimize contributors' and reviewers' bandwidth and experience.
- Strengthen SIG ownership and accountability.
- Improve the alignment of enhancements with the project's goals and architecture.
- Separate the enhancement process from community matters.

# Goals

1. Introduce a new repository, `kubevirt/enhancements`, to house all VEPs.
2. Prioritize VEPs at the start of each development cycle to streamline review and focus efforts.
3. Establish SIG-specific review processes with a dedicated reviewer for each VEP.
4. Create issues to track the progress, maturity, and associated bugs for each accepted VEP.
5. Establish ownership of the VEP and any additional SIG collaboration.

## Non-Goals

- Removing the central approval mechanism at this stage.
- Full decentralization of VEP review and approval processes.

## Repos

- **Impacted Repository**: `kubevirt/enhancements` and `kubevirt/community`

# Design

The process includes the following key components:
1. **VEP Creation**: VEP authors will initiate proposals via PRs to the `kubevirt/enhancements` repository. [Design proposal template](https://github.com/kubevirt/community/blob/main/design-proposals/proposal-template.md)
2. **SIG Review and Collaboration**: Each VEP will have a target SIG, and the SIG will assign a dedicated reviewer to oversee the proposal, collaborate with other SIGs as needed, and provide feedback or veto when necessary.
3. **Centralized Prioritization**: At the start of each release cycle, all accepted VEPs will be designated as the project’s priority, focusing community efforts on the associated pull requests. Acceptance will be based on community support and a commitment to implementation.
4. **Visibility and Tracking**: The Author of an accepted VEPs will open an issue to track their progress, maturity stages (alpha, beta, GA), list the associated bugs, and user feedback
5. **Single source of truth**: Each VEP will be the authoritative reference for the associated feature. This aligns with the Kubernetes KEP process. It will ensure that each enhancement
Includes all the relevant information, including the design and the state.
The VEP owner is responsible to update it as its development progresses, until it is fully mature (or deprecated).

## Scalability

This design allows scalability by:
- Distributing review responsibility across SIGs while maintaining central oversight.
- Providing a framework for gradual decentralization as SIGs mature.
- Centralized tracking to set priorities and ensure alignment across the project.
- Uniform process 

# Approvers
The following individuals are proposed as the initial approvers for the kubevirt/enhancements repository. These approvers will be responsible for ensuring VEPs meet the required standards, align with the project’s goals and best practices.
 - Luboslav Pivarc @xpivarc
 - Vladik Romanovsky @vladikr

# Implementation Phases

1. **Alpha Rollout (v1.5 Cycle)**:
   - Create the `kubevirt/enhancements` repository.
   - Introduce a template for VEP submissions.
   - Migrate one or two active designs to test the process.
   - Refine the process based on feedback from initial VEPs.

2. **Full Rollout (v1.6 Cycle)**:
   - Transition all enhancements to the new process.
   - Empower SIGs to take increased ownership while maintaining central prioritization.

3. **Future Considerations**:
   - Gradual reduction in centralized coordination as SIGs become self-sufficient.

