This repo contains design proposals for features impacting repos across the
KubeVirt github organization.

# Purpose of Design Proposals

The purpose of a design proposal is to allow community members to gain feedback
on their designs from the repo approvers before the community member commits to
executing on the design. By going through the design process, developers gain a
have a high level of confidence that their designs are viable and will be
accepted.


NOTE: This is process is not mandatory. Anyone can execute on their own design
without going through this process and submit code to the respective repos.
However, depending on the complexity of the design and how experienced the
developer is within the community, they could greatly benefit from going through
this design process first. The risk of not getting a design proposal approved
is that a developer may not arrive at a viable design that the community will
accept.

# How to create Design Proposals

To create a design proposal, it is recommended to use the `proposal-template.md`
file an outline. The structure of this template is meant to provide a starting
point for people. Feel free to edit and modify your outline to best fit your
needs when creating a proposal.

Once your proposal is done, submit it as a PR to the design-proposals folder.

If you want to bring further attention to your design, ping individuals
who are listed as `approvers` for the impacted repos. You may also want to
raise the design during the weekly community members call and on the mailing
list (kubevirt-dev@googlegroups.com) as well.

# Getting a design approved

For a design to be considered viable, an approver from each repo impacted by
the design needs to provide a `/lgtm` comment on the proposal's PR. Once
all `/lgtm` are collected, the design can be approved and merged. A merged
design proposal means the proposal is viable to be executed on.

# Design proposal drift

After a design proposal is merged, it's likely that the actual implementation
will begin to drift slightly from the original design. This is expected and
there is no expectation that the original design proposal needs to be updated
to reflect these differences.

The code and our user guide are the ultimate sources of truth. Design proposals
are merely the starting point for the implementation.

