# Adding merge automation to your repository

In order to run tests and enable automated merging of pull requests for your repository on KubeVirt [prow] you need to have the following components configured:
- [tide] - responsible for maintaining the merge pool
- [lgtm] - sets `lgtm` label
- approve - sets label `approved`
- [trigger] - runs test jobs for a PR
- [branch-protection] tide takes from it required status for the [presubmit] jobs configured for the repository

In a nutshell: if tide is enabled then for any created PR that has `lgtm` and `approved` labels and all presubmit checks that are required are successful, tide will merge it. 

Example configuration PR to enable tide for KubeVirt [prow]: https://github.com/kubevirt/project-infra/pull/653/

Example workflow:
- tide is configured for PRs on a repository to require `lgtm` and `approved` labels and missing `needs-rebase` label
- Alice creates a pull request, which triggers presubmit jobs to run for the PR
- some presubmit jobs have `optional: false` set, which makes their status required to be successful
- All jobs succeed and add a success status to the commit object in GitHub
- Note: tide does not consider this PR to be in the merge pool as it misses the `lgtm` and `approved` labels
- Alice asks Bob and Charles for a review on the pull request
- Bob (a member) reviews and adds a `/lgtm` comment
- After that the default branch gets updated by another merge, where merging Alice's PR would cause a conflict, which adds a `needs-rebase` label
- Charles (a maintainer) reviews and adds an `/approve` comment
- Note: the `approved` label added to the PR enables automatic retesting on the PR
- Note: tide does not consider this PR to be in the merge pool as it has the `needs-rebase` label
- Alice rebases the PR on the default branch, which in turn removes the `lgtm` and `needs-rebase` labels, also this triggers presubmit jobs to run for the PR
- Bob (a member) reviews and adds a `/lgtm` comment
- Note: tide does not consider this PR to be in the merge pool as it matches the label configuration, but misses successful status on the required presubmit jobs (see above)
- All jobs succeed and add a success status to the commit object in GitHub
- Note: tide considers this PR to be in the merge pool as it has all required labels and none of the missing labels, additionally all required status are successful
- tide will then eventually merge the PR into the default branch

[branch-protection]: https://github.com/kubernetes/test-infra/blob/master/prow/cmd/branchprotector/README.md
[lgtm]: https://github.com/kubernetes/test-infra/tree/master/prow/plugins/lgtm
[presubmit]: https://github.com/kubernetes/test-infra/blob/master/prow/jobs.md#how-to-configure-new-jobs
[prow]: https://github.com/kubernetes/test-infra/tree/master/prow#
[tide]: https://github.com/kubernetes/test-infra/tree/master/prow/tide
[trigger]: https://github.com/kubernetes/test-infra/blob/master/prow/plugins/trigger/trigger.go#L107
