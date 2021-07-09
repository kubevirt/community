# Adding merge automation to your repository

In order to run tests and enable automated merging of pull requests for your repository on KubeVirt [prow] you need to have the following components and plugins configured:
- [tide] (component) - responsible for maintaining the merge pool
- [lgtm] (plugin) - sets `lgtm` label
- [approve] (plugin) - sets label `approved`
- [trigger] (plugin)  - runs test jobs for a PR
- [branch-protection] (component) - tide takes from it required status for the [presubmit] jobs configured for the repository

In a nutshell: if tide is enabled then for any created PR that has `lgtm` and `approved` labels and misses all of the `missingLabel` labels, and furthermore all presubmit checks that are required (aka `optional: false`) are successful, tide will merge it. 

Example configuration PR to enable tide for KubeVirt [prow]: https://github.com/kubevirt/project-infra/pull/653/

Example workflow:
- tide is configured for PRs on a repository to require `lgtm` and `approved` labels and missing `needs-rebase` label
- Alice creates a pull request, which triggers presubmit jobs to run for the PR
- Some presubmit jobs have `optional: false` set, which makes their status required to be successful
- All jobs succeed and add a success status to the commit object in GitHub
- Note: tide does not consider this PR to be in the merge pool as it misses the `lgtm` and `approved` labels
- Alice asks Bob and Charles for a review on the pull request
- Bob (a member) reviews and adds a `/lgtm` comment
- Note: a GitHub review automatically acts as if the `/lgtm` comment had been placed on the PR (see [config](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/kustom/base/configs/current/plugins/plugins.yaml#L463))
- After that the default branch gets updated by another merge, where merging Alice's PR would cause a conflict, which adds a `needs-rebase` label
- Charles (a maintainer) reviews and adds an `/approve` comment
- Note: tide does not consider this PR to be in the merge pool as it has the `needs-rebase` label
- Alice rebases the PR on the default branch, which in turn removes the `lgtm` and `needs-rebase` labels, also this triggers presubmit jobs to run for the PR
- Note: The approve label is not affected by PR modifications.
- Bob (a member) reviews and adds a `/lgtm` comment
- Note: the `lgtm` and `approved` labels added to the PR and the missing `needs-rebase` label enable automatic retesting on the PR (done by a [periodic commenter job](https://github.com/kubevirt/project-infra/blob/97ce8b6cc7bf8c66c58e02f47c1ce31e580c8181/github/ci/prow-deploy/files/jobs/kubevirt/project-infra/project-infra-periodics.yaml#L2))
- Note: tide does not consider this PR to be in the merge pool as it matches the label configuration, but misses successful status on the required presubmit jobs (see above)
- All jobs succeed and add a success status to the commit object in GitHub
- Note: tide considers this PR to be in the merge pool as it has all required labels and none of the missing labels, additionally all required status are successful
- tide will then [eventually merge](https://github.com/kubernetes/test-infra/blob/master/prow/cmd/tide/pr-authors.md) the PR into the default branch

[approve]: https://github.com/kubernetes/test-infra/blob/master/prow/plugins/approve/approve.go#L132
[branch-protection]: https://github.com/kubernetes/test-infra/blob/master/prow/cmd/branchprotector/README.md
[lgtm]: https://github.com/kubernetes/test-infra/tree/master/prow/plugins/lgtm
[presubmit]: https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#how-to-configure-new-jobs
[prow]: https://github.com/kubernetes/test-infra/tree/master/prow#
[tide]: https://github.com/kubernetes/test-infra/tree/master/prow/tide
[trigger]: https://github.com/kubernetes/test-infra/blob/master/prow/plugins/trigger/trigger.go#L107
