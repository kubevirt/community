# SIG Responsibilities

All KubeVirt SIGs must define a charter defining the scope and governance of the SIG.

- The scope must define what areas the SIG is responsible for directing and maintaining.
- The governance must outline the responsibilities within the SIG as well as the roles
  owning those responsibilities.

## Steps to create a SIG charter

1. Copy [the template][Short Template] into a new file under community/sig-*YOURSIG*/charter.md ([sig-scale example])
2. Fill out the template for your SIG
3. Update [sigs.yaml] with the individuals holding the roles as defined in the template.
4. Add subprojects owned by your SIG in the [sigs.yaml]
5. Create a pull request with a draft of your charter.md and sigs.yaml changes. Communicate it within your SIG
   and get feedback as needed.
6. Send the SIG Charter out for review to kubevirt-dev@googlegroups.com. Include the subject "SIG Charter Proposal: YOURSIG"
   and a link to the PR in the body.
8. Typically expect feedback within a week of sending your draft. Expect longer time if it falls over an
   event such as KubeCon/CloudNativeCon or holidays. Make any necessary changes.
9. Once accepted, the maintainers will ratify the PR by merging it.

## How to use the templates

SIGs should use [the template][Short Template] or [prefilled template][Short Template with examples] as a starting point.


## Goals

The primary goal of the charters is to define the scope of the SIG within KubeVirt and how the SIG leaders exercise ownership of these areas by taking care of their responsibilities. A majority of the effort should be spent on these concerns.

[Short Template]: sig-charter-template.md
[Short Template with examples]: sig-charter-template-prefilled.md
[sigs.yaml]: ../sigs.yaml
[sig-scale example]: ../../sig-scale/charter.md

