# Automating GitHub org management

**TLDR; any changes that are applied manually using the GitHub UI will be reverted within one hour! Please file a PR against the [org configuration file]!**

**Contributors applying for KubeVirt org membership: please have a look at the requirements [here](../membership_policy.md#requirements)**

We are using [peribolos] to make transparent and automate management of our GitHub organization, covering members, org and repo settings, teams and access rights of teams to repositories.

Our [org configuration file] contains the settings for our organization. These settings are synchronized by a periodic job.

If you need to change settings within kubevirt org regarding

- org settings, i.e.
    - org admins
    - org members
- repos, i.e.
    - new repos
    - repo settings
- teams, i.e.
    - team maintainers and members
    - team access to repositories

Please file a PR against the [org configuration file].

After the PR is merged changes will be applied within one hour.

Examples:
* [Adding a github user to the org as member](https://github.com/kubevirt/project-infra/pull/1765)
* [Adding a repository and a team, giving the team access to the repository](https://github.com/kubevirt/project-infra/pull/990/files)

[org configuration file]: https://github.com/kubevirt/project-infra/blob/fe7457d449e0d03d5a0dd62359f415b44c3fa323/github/ci/prow/files/orgs.yaml#L1
[peribolos]: https://github.com/kubernetes/test-infra/tree/master/prow/cmd/peribolos
