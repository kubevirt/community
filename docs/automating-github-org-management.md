# Automating GitHub org management

**TLDR; any changes that are applied manually using the GitHub UI will be reverted within one hour! Please file a PR against the [org configuration file]!**

**Contributors applying for KubeVirt org membership: please have a look at the requirements [here](../membership_policy.md#requirements)**

We are using [peribolos] to make transparent and automate management of our GitHub organization, covering members, org and repo settings, teams and access rights of teams to repositories.

Our [org configuration file] contains the settings for our organization. These settings are synchronized by a periodic job.

If you need to change settings within kubevirt org regarding

- org settings, i.e.
    - org admins
    - org members
- repo settings
- teams, i.e.
    - team maintainers and members
    - team access to repositories

please file a PR against the [org configuration file].

## New repositories

Creating new repositories is part of the [VEP process](https://github.com/kubevirt/enhancements#process), since there the VEP authors and the reviewers decide where to host the implementation. New repos are described in section [`Repos`](https://github.com/kubevirt/enhancements/blob/main/veps/NNNN-vep-template/vep.md#repos) of the VEP template.

After the VEP is accepted the VEP authors can create a PR to describe the repositories and the teams maintaining them. See below for an example.

## Applying changes

Please file a PR against the [org configuration file].

After the PR is merged changes will be applied within one hour.

Examples:
* [Adding a github user to the org as member](https://github.com/kubevirt/project-infra/pull/1765)
* [Adding a repository and a team, giving the team access to the repository](https://github.com/kubevirt/project-infra/pull/990/files)

[org configuration file]: https://github.com/kubevirt/project-infra/blob/fe7457d449e0d03d5a0dd62359f415b44c3fa323/github/ci/prow/files/orgs.yaml#L1
[peribolos]: https://github.com/kubernetes/test-infra/tree/master/prow/cmd/peribolos
