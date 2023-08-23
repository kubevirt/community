# SIG Charter Guide

All KubeVirt SIGs must define a charter defining the scope and governance of the SIG.

- The scope must define what areas the SIG is responsible for directing and maintaining.
- The governance must outline the responsibilities within the SIG as well as the roles
  owning those responsibilities.

## Steps to create a SIG charter

1. Copy [the template][sig-charter-template] into a new file under community/sig-_YOURSIG_/charter.md
2. Read [sig-governance-requirements] so you have context for the template
3. Fill out the template for your SIG
4. Update [sigs.yaml] with the individuals holding the roles as defined in the template.
5. Add subprojects owned by your SIG in the [sigs.yaml]
6. Create a pull request with a draft of your charter.md and sigs.yaml changes. Communicate it within your SIG
   and get feedback as needed.
7. Send the SIG Charter out for review to kubevirt-dev@googlegroups.com. Include the subject "SIG Charter Proposal: YOURSIG"
   and a link to the PR in the body.
8. Typically expect feedback within a week of sending your draft. Expect longer time if it falls over an
   event such as KubeCon/CloudNativeCon or holidays. Make any necessary changes.
9. Once accepted, the steering committee will ratify the PR by merging it.

## Steps to update an existing SIG charter

- For significant changes, or any changes that could impact other SIGs, such as the scope, create a
  PR and send it to the steering committee for review with the subject: "SIG Charter Update: YOURSIG"
- For minor updates to that only impact issues or areas within the scope of the SIG the SIG Chairs should
  facilitate the change.

## SIG Charter approval process

When introducing a SIG charter or modification of a charter the following process should be used.
As part of this we will define roles for the [OARP] process (Owners, Approvers, Reviewers, Participants)

- Identify a small set of Owners from the SIG to drive the changes.
  Most typically this will be the SIG chairs.
- Work with the rest of the SIG in question (Reviewers) to craft the changes.
  Make sure to keep the SIG in the loop as discussions progress with the Steering Committee (next step).
  Including the SIG mailing list in communications with the steering committee would work for this.
- Work with the steering committee (Approvers) to gain approval.
  This can simply be submitting a PR and sending mail to [kubevirt-dev@googlegroups.com].
  If more substantial changes are desired it is advisable to socialize those before drafting a PR.
  - The steering committee will be looking to ensure the scope of the SIG as represented in the charter is reasonable (and within the scope of KubeVirt) and that processes are fair.
- For large changes alert the rest of the KubeVirt community (Participants) as the scope of the changes becomes clear.
  Sending mail to [kubevirt-dev@googlegroups.com] and (optionally) announcing it in the #kubevirt-dev slack channel.

If there are questions about this process please reach out to the steering committee at [kubevirt-dev@googlegroups.com].

## How to use the templates

SIGs should use [the template][sig-charter-template] as a starting point. This document links to the recommended [SIG Governance][sig-governance] but SIGs may optionally record deviations from these defaults in their charter.

## Goals

The primary goal of the charters is to define the scope of the SIG within KubeVirt and how the SIG leaders exercise ownership of these areas by taking care of their responsibilities. A majority of the effort should be spent on these concerns.

[sig-governance-requirements]: sig-governance-requirements.md
[sig-governance]: sig-governance.md
[sig-charter-template]: sig-charter-template.md
[sigs.yaml]: https://github.com/KubeVirt/community/blob/master/sigs.yaml
[kubevirt-dev@googlegroups.com]: mailto:kubevirt-dev@googlegroups.com
[OARP]: https://stumblingabout.com/tag/oarp/
