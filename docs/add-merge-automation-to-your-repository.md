# Adding merge automation to your repository

In order to run tests and enable automated merging of pull requests for your repository on KubeVirt [prow] you need to have the following components and plugins configured:
- [tide] (component) - responsible for maintaining the merge pool
- [lgtm] (plugin) - sets `lgtm` label
- [approve] (plugin) - sets label `approved`
- [trigger] (plugin)  - runs test jobs for a PR
- [branch-protection] (component) - tide takes from it required status for the [presubmit] jobs configured for the repository

In a nutshell: if tide is enabled then for any created PR that has `lgtm` and `approved` labels and misses all of the `missingLabel` labels, and furthermore all presubmit checks that are required (aka `optional: false`) are successful, tide will merge it. 

Example configuration PR to enable tide for KubeVirt [prow]: https://github.com/kubevirt/project-infra/pull/653/

## Example workflow

Initial situation: `tide` is configured for PRs on a repository to require `lgtm` and `approved` labels and missing `needs-rebase` label

- Alice creates a pull request, which triggers presubmit jobs to run for the PR
- Some presubmit jobs have `optional: false` set, which makes their status required to be successful
- All jobs succeed and add a success status to the commit object in GitHub
> [!NOTE]
> `tide` does **not** consider this PR to be in the merge pool as it misses the `lgtm` and `approved` labels
- Alice asks Bob and Charles for a review on the pull request
- Bob (a member) reviews and adds a `/lgtm` comment
> [!NOTE]
> A GitHub review automatically _may_ act as if the `/lgtm` comment had been placed on the PR (see [config](https://github.com/kubevirt/project-infra/blob/fbb84ab7a4206079c94c1ee226a5af12915d9f0b/github/ci/prow-deploy/kustom/base/configs/current/plugins/plugins.yaml#L747))
- After that the default branch gets updated by another merge, where merging Alice's PR would cause a conflict, which adds a `needs-rebase` label
- Charles - an approver listed in the [OWNERS] files for _all the files_ touched in the PR - reviews and adds an `/approve` comment
> [!NOTE]
> tide does **not** consider this PR to be in the merge pool as it has the `needs-rebase` label
- Alice rebases the PR on the default branch, which in turn removes the `lgtm` and `needs-rebase` labels, also this triggers presubmit jobs to run for the PR
> [!NOTE]
> The `approved` label is not affected by PR modifications.
- Bob - a member of KubeVirt GitHub org - reviews and adds a `/lgtm` comment
> [!NOTE]
> the `lgtm` and `approved` labels added to the PR and the missing `needs-rebase` label enable automatic retesting on the PR (done by a [periodic commenter job](https://github.com/kubevirt/project-infra/blob/97ce8b6cc7bf8c66c58e02f47c1ce31e580c8181/github/ci/prow-deploy/files/jobs/kubevirt/project-infra/project-infra-periodics.yaml#L2))

> [!NOTE]
> `tide` does **not** consider this PR to be in the merge pool as it matches the label configuration, but misses successful status on the required presubmit jobs (see above)
- All jobs succeed and add a success status to the commit object in GitHub
> [!NOTE]
> tide considers this PR to be in the merge pool as it has all required labels and none of the missing labels, additionally all required status are successful
- tide will then [eventually merge](https://github.com/kubernetes-sigs/prow/blob/main/site/content/en/docs/components/core/tide/pr-authors.md) the PR into the default branch

[prow]: https://docs.prow.k8s.io/docs/
[approve]: https://github.com/kubernetes-sigs/prow/blob/0e909e33e02e45dcb2f2fc5b605f8057e44f1c5a/pkg/plugins/approve/approve.go#L132
[branch-protection]: https://docs.prow.k8s.io/docs/components/optional/branchprotector/
[lgtm]: https://docs.prow.k8s.io/docs/components/plugins/approve/approvers/#lgtm-label
[presubmit]: https://docs.prow.k8s.io/docs/jobs/
[tide]: https://docs.prow.k8s.io/docs/components/core/tide/
[trigger]: https://github.com/kubernetes-sigs/prow/blob/0e909e33e02e45dcb2f2fc5b605f8057e44f1c5a/pkg/plugins/trigger/trigger.go#L107
[OWNERS]: https://www.kubernetes.dev/docs/guide/owners/#owners-spec
