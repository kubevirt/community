# SIG Lifecycle

## SUMMARY

This document covers everything you need to know about the creation and retirement (“lifecycle”) of a special interest within the KubeVirt project. General project governance information can be found in the [steering committee repo].

## Creation

### Pre-requisites for a SIG

- [ ] Read [sig-governance.md]
- [ ] Ensure all SIG Chairs, Technical Leads, and other leadership roles are [community members]
- [ ] Send an email to <kubevirt-dev@googlegroups.com> to scope the SIG and get provisional approval
- [ ] Look at the checklist below for processes and tips that you will need to do while this is going on. It's best to collect this information upfront so you have a smoother process to launch
- [ ] Follow the [sig-charter-guide] to propose and obtain approval for a charter
- [ ] Announce new SIG on <kubevirt-dev@googlegroups.com>

### GitHub

- [ ] Submit a PR that will
  - [ ] Adds rows to [sigs.yaml]
    - You’ll need:
    - SIG Name
    - Directory URL
    - Mission Statement
    - Chair Information
    - Meeting Information
    - Contact Methods
    - Any SIG Stakeholders
    - Any Subproject Stakeholders
  - [ ] Creates `kubevirt/community/sig-_YOURSIG_`
  - [ ] Adds SIG-related docs like charter.md, schedules, roadmaps, etc/

#### TODO

- Add automation to create SIG directory and skeleton documents after [sigs.yaml] is updated
- Add process for adding labels to project-infra

### Communicate

Each one of these has a linked canonical source guideline from set up to moderation and your role and responsibilities for each. We are all responsible for enforcing our [code of conduct].

## Retirement

(merging or disbandment)
Sometimes it might be necessary to sunset a SIG. SIGs may also merge with an existing SIG if deemed appropriate, and would save project overhead in the long run. Working Groups in particular are more ephemeral than SIGs, so this process should be followed when the Working Group has accomplished it's mission.

### Prerequisites for SIG Retirement

- [ ] SIG’s retirement decision follows the decision making and communication processes as outlined in their charter

[steering committee repo]: https://github.com/kubernetes/steering
[sig-governance.md]: /committee-steering/governance/sig-governance.md
[sig-charter-guide]: sig-charter-guide.md
[sigs.yaml]: /sigs.yaml
[code of conduct]: /code-of-conduct.md
[community members]: /membership_policy.md
