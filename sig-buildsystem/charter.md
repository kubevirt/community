# SIG Buildsystem Charter

## Scope

SIG buildsystem takes care of all things building and automating for the KubeVirt project - they ensure new and existing contributors have a smooth contribution experience, i.o.w. contributors shouldn't have the feeling that they are wasting their time.
This includes bazel, test automation and updates, i.e. WORKSPACE file and kubevirtci/cluster-up.
This also includes the support of failing builds or test lanes whose failures are caused by infrastructural or configuration problems.

### In scope

#### Code

- [kubevirt/kubevirt]
  - all test automation scripts inside folder `automation/`
  - KubeVirt builder and other build related scripts inside `hack/`
  - `cluster-up/`, [KubeVirtCI] updates
- [kubevirt/project-infra]
  - sig-lane support
    - helping fix and supporting issues with the sig-lanes as long as these are not code related
    - maintaining automation for kubevirt jobs progression and test lane bumping
    - updating jobs for newer k8s versions inside `github/ci/prow-deploy/files/jobs/kubevirt/kubevirt`
  - supporting bumping of other test lanes to newer k8s providers
  - supporting configuration changes wrt prow plugins, i.e. tide configuration, for [kubevirt/kubevirt]

#### Binaries

- [quay.io/kubevirt]
  - creating new repositories for images
  - supporting image update issues
- [kubevirt/kubevirt]
  - supporting issues with automated updates for WORKSPACE, uploads of dependencies
  - supporting the caches for images and binaries used with quay, docker and prow

#### Services

- [Prow] issues
  - helping to resolve or forward infrastructural issues or configuration problems
  - connecting or forwarding to [KubeVirt CI operations group]
- [kubevirt/kubevirt] ci health
  - Flake detection and remediation - regularly consulting the available [ci reports] wrt flaky tests, ensuring the flake process gets executed
  - keeping the merge queue healthy - regularly consulting the number of merges and unblocking merge queue traffic jams
  - continuously improving overall CI throughput, i.e. reducing lane cycle times

### Out of scope

Fixing or debugging failing tests in general,  fixing code issues, i.e. flaky tests, faulty imports - that is the responsibility of the specialist SIGs.

Also fixing general prow issues, that is the [KubeVirt CI operations group] responsibility.

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts

[OARP]: https://stumblingabout.com/tag/oarp/
[kubevirt/kubevirt]: https://github.com/kubevirt/kubevirt
[kubevirt/project-infra]: https://github.com/kubevirt/project-infra
[Prow]: https://prow.ci.kubevirt.io
[KubeVirtCI]: https://github.com/kubevirt/kubevirtci
[KubeVirt CI operations group]: https://github.com/kubevirt/project-infra/blob/main/docs/ci-operations-group.md
[quay.io/kubevirt]: https://quay.io/kubevirt
[ci reports]: https://github.com/kubevirt/project-infra/blob/main/docs/reports.md
