# SIG BUILDSYSTEM Charter

## Scope

SIG buildsystem takes care of all things building and automating for the KubeVirt project.
This includes bazel, test automation and updates, i.e. WORKSPACE file and kubevirtci/cluster-up.
This also includes the support of failing builds or test lanes whose failures are caused by infrastructural or configuration problems.

### In scope

#### Code, Binaries and Services

- all test automation scripts inside folder `automation/`
- KubeVirt builder and other build related scripts inside `hack/`
- `cluster-up/`, [KubeVirtCI] updates

### Out of scope

Fixing failing tests in general, that's the other SIGs responsibility.
Fixing flaky tests, fixing faulty imports, fixing bugs inside KubeVirt project itself.

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts

[OARP]: https://stumblingabout.com/tag/oarp/
[KubeVirtCI]: https://github.com/kubevirt/kubevirtci
