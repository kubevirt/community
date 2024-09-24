# Working Group Responsibilities

All KubeVirt Working Groups (WGs) must define a charter.

- The scope must define:
  - what problem the WG is solving
  - what deliverable(s) will be produced, and to whom
  - when the problem is expected to be resolved
- The governance must outline:
  - the responsibilities within the WG
  - the roles owning those responsibilities
  - which SIGs are stakeholders

## Steps to create a WG charter

1. Copy [the template][WG Template] into a new file under community/wg-*YOURWG*/charter.md
2. Fill out the template for your WG
3. Update [sigs.yaml] with the individuals holding the roles as defined in the template.
4. Create a pull request with a draft of your `charter.md` and `sigs.yaml` changes. Communicate it within your WG and the stakeholder SIGs and get feedback as needed.
5. Send the WG Charter out for review to kubevirt-dev@googlegroups.com. Include the subject "WG Charter Proposal: YOURWG"
   and a link to the PR in the body.
6. Typically expect feedback within a week of sending your draft. Expect longer time if it falls over an
   event such as KubeCon/CloudNativeCon or holidays. Make any necessary changes.
7. Once accepted, the maintainers will ratify the PR by merging it.

## How to use the templates

WGs should use [the template][WG Template] as a starting point.


## Goals

The primary goal of the charters is to define the scope of the WG within KubeVirt and how the WG can determine whether it has achieved their solution. A majority of the effort should be spent on these concerns.

[WG Template]: wg-charter-template.md
[sigs.yaml]: ../sigs.yaml

