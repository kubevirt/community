# SIG CI Charter

## Scope

The SIG CI - aka [KubeVirt CI Operations Group] - maintains the KubeVirt CI infrastructure.

### In scope

The group is **primarily** responsible for keeping the [KubeVirt CI infrastructure](https://github.com/kubevirt/project-infra/blob/main/docs/infrastructure-components.md#infrastructure-components) operational, such that CI jobs are executed in a timely manner and PRs of [any of the onboarded projects](https://github.com/kubevirt/project-infra/tree/main/github/ci/prow-deploy/files/jobs) are not blocked.

Additional **secondary** responsibilities are:

* keeping an eye on prow job failures and notify members of the sig teams if required
* supporting and educating sig members in CI matters related to prow job configuration
* regularly updating the prow deployment (effectively meaning looking at the [automated bump jobs](https://github.com/kubevirt/project-infra/pulls/kubevirt-bot))
* maintaining cluster nodes (in coordination with the KNI infrastructure team)
* maintaining cluster configuration (i.e. prow concertation, onboarding and updating other cluster configs inside the [secrets repository](https://github.com/kubevirt/secrets/), also adding secrets that folks can use in their jobs or actions)

### Out of scope

The SIG is NOT responsible for

* fixing flaky tests, as long as those tests are not caused by the CI infrastructure itself
* fixing the overload of the CI infrastructure if it is caused by misuse of the infrastructure
* improving the runtime of specific lanes as long as the creators of the lane are capable of handling this themselves
* anything else that people outside this group are capable of handling on their own

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts

[KubeVirt CI Operations Group]: https://github.com/kubevirt/project-infra/blob/main/docs/ci-operations-group.md
[OARP]: https://stumblingabout.com/tag/oarp/
