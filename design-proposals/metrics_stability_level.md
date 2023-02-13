# Metric Stability

[KEP-1209] introduced the concept of metric stability to Kubernetes. When metrics graduate to the `STABLE` `StabilityLevel`, we provide guarantees to consumers of those metrics so they can confidently build alerting and monitoring platforms.

## Stability Classes

There are currently two stability classes for metrics: (1) Alpha, (2) Stable. These classes are intended to make explicit the API contract between the control-plane and the consumer of control-plane metrics.

### Alpha

__Alpha__ metrics have __*no*__ stability guarantees; as such they can be modified or deleted at any time. All Kubernetes metrics begin as alpha metrics.

An example of an alpha metric follows:

```go
var alphaMetricDefinition = kubemetrics.CounterOpts{
    Name: "some_alpha_metric",
    Help: "some description",
    StabilityLevel: kubemetrics.ALPHA, // this is also a custom metadata field
	DeprecatedVersion: "1.15", // this can optionally be included on alpha metrics, although there is no change to contractual stability guarantees
}
```

### Stable

__Stable__ metrics can be guaranteed to *not change*, except that the metric may become marked deprecated for a future Kubernetes version.

An example of a stable metric follows:

```go
var deprecatedMetricDefinition = kubemetrics.CounterOpts{
    Name: "some_deprecated_metric",
    Help: "some description",
    StabilityLevel: kubemetrics.STABLE, // this is also a custom metadata field
    DeprecatedVersion: "1.15", // this is a custom metadata field
}
```

By *not change*, we mean three things:

1. the metric itself will not be deleted ([or renamed](#metric-renaming))
2. the type of metric will not be modified
3. no labels can be added **or** removed from this metric

From an ingestion point of view, it is backwards-compatible to add or remove possible __values__ for labels which already do exist (but __not__ labels themselves). Therefore, adding or removing __values__ from an existing label is permissible. Stable metrics can also be marked as __deprecated__ for a future Kubernetes version, since this is a metadata field and does not actually change the metric itself.

**Removing or adding labels from stable metrics is not permissible.** In order to add/remove a label to an existing stable metric, one would have to introduce a new metric and deprecate the stable one; otherwise this would violate compatibility agreements.

## API Review

Graduating a metric to a stable state is a contractual API agreement, as such, it would be desirable to require an api-review (to sig-instrumentation) for graduating or deprecating a metric (in line with current Kubernetes [api-review processes](https://github.com/kubernetes/community/blob/master/sig-architecture/api-review-process.md)).

We use a verification script to flag stable metric changes for review by SIG Instrumentation approvers.

## Metric Renaming

Metric renaming is be tantamount to deleting a metric and introducing a new one. Accordingly, metric renaming will also be disallowed for stable metrics.

## Deprecation Lifecycle

Metrics can be annotated with a Kubernetes version, from which point that metric will be considered deprecated. This allows us to indicate that a metric is slated for future removal and provides the consumer a reasonable window in which they can make changes to their monitoring infrastructure which depends on this metric.

While deprecation policies only actually change stability guarantees for __stable__ metrics (and not __alpha__ ones), deprecation information may however be optionally provided on alpha metrics to help component owners inform users of future intent, to help with transition plans (this change was made at the request of @dashpole, who helpfully pointed out that it would be nice to be able signal future intent even for alpha metrics).

When a stable metric undergoes the deprecation process, we are signaling that the metric will eventually be deleted. The lifecyle looks roughly like this (each stage represents a Kubernetes release):

__Stable metric__ -> __Deprecated metric__ -> __Hidden metric__ -> __Deletion__

__Deprecated__ metrics have the same stability guarantees of their counterparts. If a stable metric is deprecated, then a deprecated stable metric is guaranteed to *not change*. When deprecating a stable metric, a future Kubernetes release is specified as the point from which the metric will be considered deprecated.

```go
var someCounter = kubemetrics.CounterOpts{
    Name: "some_counter",
    Help: "this counts things",
    StabilityLevel: kubemetrics.STABLE,
    DeprecatedVersion: "1.15", // this metric is deprecated when the Kubernetes version == 1.15
}
````

__Deprecated__ metrics will have their description text prefixed with a deprecation notice string '(Deprecated from x.y)' and a warning log will be emitted during metric registration (in the spirit of the official [Kubernetes deprecation policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/#deprecating-a-flag-or-cli)).

Before deprecation:

```text
# HELP some_counter this counts things
# TYPE some_counter counter
some_counter 0
```

During deprecation:
```text
# HELP some_counter (Deprecated from 1.15) this counts things
# TYPE some_counter counter
some_counter 0
```
Like their stable metric counterparts, deprecated metrics will be automatically registered to the metrics endpoint.

On a subsequent release (when the metric's deprecatedVersion is equal to current_kubernetes_version - 1)), a deprecated metric will become a __hidden metric__. _Unlike_ their deprecated counterparts, hidden metrics will __*no longer be automatically registered*__ to the metrics endpoint (hence hidden). However, they can be explicitly enabled through a command line flag on the binary (i.e. '--show-hidden-metrics-for-version=<previous minor release>'). This is to provide cluster admins an escape hatch to properly migrate off of a deprecated metric, if they were not able to react to the earlier deprecation warnings. Hidden metrics should be deleted after one release.


[KEP-1209]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-instrumentation/1209-metrics-stability
