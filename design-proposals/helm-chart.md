# Overview

Provide and support Helm charts as an alternative installation method for KubeVirt and CDI.

## Motivation

Helm is one of the most popular choices when it comes to managing Kubernetes applications and their release lifecycle.
Many software platforms prefer Helm charts for their configuration, automation and distribution capabilities.

## Goals

- Provide Helm charts for all upcoming versions of KubeVirt and CDI

## Non Goals

- Removing manifests installation method is out of scope.
  Users will continue to have the option to deploy KubeVirt and CDI using the manifests provided at build time.

## Definition of Users

This feature is intended for administrators of Kubernetes clusters who are managing KubeVirt's (and CDI's) release lifecycles.

## User Stories

- As a cluster administrator, I'd want to install, upgrade and uninstall KubeVirt using Helm chart
- As a cluster administrator, I'd want to install, upgrade and uninstall CDI using Helm chart

## Repos

* [kubevirt/kubevirt](https://github.com/kubevirt/kubevirt)
* [kubevirt/containerized-data-importer](https://github.com/kubevirt/containerized-data-importer)

# Design

In order to establish a solution design we need to answer the following questions:
* How are the Helm charts going to look like?
* How will the Helm charts be managed?

## Chart Design

Both KubeVirt and CDI installations depend on CRD, CR and Operator Deployment resources.

The current installation procedure involves creating the CRD and Operator Deployment first closely followed by creating the CR.

Helm, however, has limitations when it comes to [handling CRDs](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/).
Upgrading and deleting is not currently supported (unless templated) and CRDs must be present before creating a CR.
The CR is required in order for the operator to deploy the rest of KubeVirt's dependencies and applications.

This leaves us with the following Helm chart layouts:

### Approach 1: Single chart

```text
├── Chart.yaml
├── crds
│   └── kubevirt-crd.yaml
├── templates
│    ├── kubevirt-cr.yaml
│    ├── kubevirt-hooks.yaml
│    └── kubevirt-operator.yaml
└── values.yaml
```

This method allows us to have all resources in the same chart.

The `CustomResourceDefinition` is extracted out of `kubevirt-operator.yaml` under the special `/crds` directory which Helm will pick up and install first.
Note that `kubevirt-crd.yaml` cannot be templated.

The installation process for this approach is simple (`$ helm install kubevirt <repo-url>`), however,
there are some obstacles when it comes to upgrade and uninstall procedures.

Upgrading a CRD is not currently possible in Helm. Once installed, further upgrades will only detect that the CRD already exists and will ignore it.

Uninstalling depends on the CR being removed before the Operator. The CRD removal is also not taken care of by Helm automatically.

In order to resolve these issues, we can leverage [Helm's hooks mechanism](https://helm.sh/docs/topics/charts_hooks/).
It allows us to intervene at certain points of the release lifecycle and perform custom operations.
These operations will be defined in `kubevirt-hooks.yaml` and will consist of the following:
- `pre-upgrade` hook manually applying the CRD manifest in order to perform modifications to it
- `pre-delete` hook removing the CR first in order for the operator to gracefully uninstall KubeVirt's components before Helm removes it
- `post-delete` hook removing the leftover CRD

This approach was used in [SUSE Edge's initial Helm chart](https://github.com/suse-edge/charts/tree/main/charts/kubevirt/0.1.0).
[PR](https://github.com/suse-edge/charts/pull/13) for additional context.

### Approach 2: One CRD chart and one CR + Operator chart

Perhaps the most common solution to tackle the CRD problem is to extract the CRD in a separate Helm chart.

```text
├── Chart.yaml
├── templates
│   └── kubevirt-crd.yaml
```

```text
├── Chart.yaml
├── templates
│    ├── kubevirt-cr.yaml
│    ├── kubevirt-hooks.yaml
│    └── kubevirt-operator.yaml
└── values.yaml
```

Here we will install the CRD chart first and the CR / Operator chart after that.
Upcoming releases of KubeVirt can ship newer versions of both charts in order to avoid discrepancies between chart versions.

The benefits are that we don't need to worry about manually upgrading or deleting the CRD. We'd still need a `pre-delete` hook to remove the CR first though.

The negatives are that we'd maintain two charts instead of a single one.

### Approach 3: One CR chart and one CRD + Operator chart

The showstopper is that the CRD must exist before deploying the CR.
We could also achieve this with the following split:

```text
├── Chart.yaml
├── templates
│    ├── kubevirt-crd.yaml
│    └── kubevirt-operator.yaml
└── values.yaml
```

```text
├── Chart.yaml
├── templates
│   └── kubevirt-cr.yaml
└── values.yaml
```

This approach is also similar to the current installation / uninstallation flows:

```shell
$ helm install kubevirt-operator <repo-url>
$ helm install kubevirt-cr <repo-url>
...
$ helm uninstall kubevirt-cr
$ sleep 45 # wait for the operator to uninstall the different components
$ helm uninstall kubevirt-operator
```

Note that the CRD is under `/templates` for this example.
This means that we can execute upgrades / deletes without custom hooks.
This isn't possible if the CR is also a template in the same chart.

### Approach N: Exotic options

As seen above, Helm provides a lot of customization possibilities.
I'd only mention some more and move on.
Let me know if you're interested in either of the following:
- One CRD + Operator chart, post-install hook creating the CR
- One CRD, one CR and one Operator chart

### Overview

Many Helm chart layouts can work. Choosing one depends on how much we'd want to rely on chart hooks.

[Approach 1](#approach-1-single-chart) was chosen for the initial Helm chart linked above in order to:
- stay as close as possible to the original manifest setup
- demonstrate chart hooks in action

As mentioned earlier, I believe [Approach 2](#approach-2-one-crd-chart-and-one-cr--operator-chart) is the most popular so far.

[Approach 3](#approach-3-one-cr-chart-and-one-crd--operator-chart) is most similar to the current setup.

The CDI project has the same `CRD-Operator-CR` setup so all of the above applies to it as well.

Note that the file structures above are purely for illustrative purposes.
Different Kubernetes resources (e.g. `rbac`) could be extracted as separate templates (files) in order to keep the charts cleaner and easier to navigate.

## Chart Management

The second problem which needs to be addressed is how Helm charts are going to be created and modified.

Currently, the manifests in KubeVirt are generated in two steps:
1. `resource-generator` generates semi-populated configurations under `/manifests/generated`
2. `manifest-templator` produces the final version of the manifests which are then shipped as part of each GitHub release

### Approach 1: Generate Helm chart(s)

Helm charts are in general not that different from plain Kubernetes manifests.
It all depends on how much customization we'd allow the user but for the SUSE Edge initial chart we kept as little modifications as possible.

We should be able to plug in the functionality to generate a `Chart.yaml`, `values.yaml` and possibly hooks and organize those in the necessary file structure.
The charts can then be generated, packaged and pushed to the configured Helm repository at build time.

### Approach 2: Manually curate a Helm chart

We can depend on the Kubernetes manifests provided as part of GitHub releases and create a Helm chart based on those.
This is the approach I've seen most people attempt when trying to come up with a custom KubeVirt installation method.

It does not involve any changes to the manifest generation flow, however, some degree of automation would still be necessary in order to patch, package and publish the charts.
The end result probably needs to be kept outside the main KubeVirt repository.

### Overview

In theory, both approaches should work.

[Approach 1](#approach-1-generate-helm-charts) will probably require more effort when it comes to the implementation,
but it should be stable and both manifests and charts would be released simultaneously.

[Approach 2](#approach-2-manually-curate-a-helm-chart) is the more popular choice,
however, we need to keep in mind that the complete manifests are not version controlled.
This would also mean that updated Helm charts are released separately from the manifests.

Unfortunately, I'm still very new to KubeVirt so feedback in this section would be more than welcome.

I'm not familiar with CDI's current manifest creation flow yet, but I *believe* it is similar.

## Helm repository

Chart repository is a location where packaged charts can be stored and shared.

We need to configure one and publish KubeVirt and CDI charts there.

There are multiple options to achieve this. The ones worth considering are:

- GitHub pages

This is the most popular way of setting up repositories for GitHub based projects.
GitHub can be configured to serve static content using simple configuration.
[Example guide](https://helm.sh/docs/topics/chart_repository/#github-pages-example).

- OCI registry

The more interesting option is to use OCI registry as Helm repository.
[This feature](https://helm.sh/docs/topics/registries/#using-an-oci-based-registry) has graduated to GA since Helm v3.8.0.
We should be able to use Quay OCI considering all other artefacts are already stored there.
[Example guide](https://cloud.redhat.com/blog/quay-oci-artifact-support-for-helm-charts)

## API Examples

Working with Helm charts is fairly straightforward. Users should be able to manage the applications by:

```shell
$ helm repo add kubevirt <repo-url>
$ helm install kubevirt kubevirt/kubevirt --version <version> \
      # additional configuration
$ helm install cdi kubevirt/cdi --version <version> \
      # additional configuration
$ helm uninstall kubevirt
$ helm uninstall cdi
```

## Scalability

N/A

## Update/Rollback Compatibility

The feature will not impact older versions of KubeVirt or CDI.
Helm charts will, however, be a viable way of updating & rolling back future releases.

## Functional Testing Approach

- Deploy KubeVirt from Helm chart
- Ensure all components are created and work as expected
- Update KubeVirt from N-1 to N release using Helm charts
- Ensure all components are upgraded and working as expected
- Uninstall KubeVirt chart
- Ensure all components have been removed

The same tests should be performed for CDI.

# Implementation Phases

- Create a KubeVirt Helm chart
- Configure a Helm repository
- Create a CDI Helm chart and add it to the same Helm repository
