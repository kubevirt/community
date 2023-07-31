# Monitoring code refactor

**Authors:** João Vilaça <jvilaca@redhat.com>

## What

This design document is proposing a code refactor for the monitoring logic in
all KubeVirt components. The goal is to have a consistent monitoring package
between all projects, a code struct that is easy to maintain and evolve, while
moving closer to the Kubernetes metric implementation style.

## Why

It’s important to refactor the monitoring code to separate monitoring logic from
business logic. It will make the codebase more modular, readable, and
maintainable. It will reduce complexity, reduce the risk of introducing errors,
and allow us to test each component independently.

Right now, every change regarding monitoring needs to happen at many different
places. For example, when adding the linter for metrics names, it was necessary
to go to all KubeVirt components repositories and implement the linter
validation in each. This is demanding, very time-consuming, requires going
through multiple duplicated code reviews, uses many CI resources, and makes it
mentally harder to decide to change anything.

### Pitfalls of the current solution

We are facing multiple challenges with the current solution:

- It is hard to keep a mental model of how monitoring is implemented in all the
different repositories and is difficult to maintain all of them in sync with new
features

- Hard to add new features to metrics and alerts, for example, because of how
metrics are being created it is hard to add a new metric stability field to all
of them since they are independently created in many different places

- Complex to collect all existing metrics and alerts for automatic documentation

- High levels of intertwined monitoring and business logic code

### Goals

- Decouple monitoring logic from business logic
- Encapsulate the monitoring best-practices and the common patterns into a
library and have it as a dependency for all KubeVirt components
- Keep monitoring code and utilities easy to maintain and evolve
- Document how to use the monitoring library for adding new metrics and rules
- Have a structure and tools to accurately and easily generate monitoring
documentation, lint metrics and alerts, define allow-list, deny-list and
opt-in metrics and other future features without having to change the code in
multiple places

### Non-Goals

- Update any metrics or alerts
- Apply these ideas to events in this phase
- Create a new component to run in the cluster alongside other components to
collect metrics and create alerts

## How

### General relationship diagram:

Each component will use a generic package for operator observability,
abstracting all common operations between the components (such as registering
them in `controller-runtime` metrics `Registry`), deduplicating code, and
providing a way simplify the definition of metrics, recording rules, and alerts.

The `docs/` package will list the metrics, alerts, and recording rules and
generate the documentation files using the operator observability docs
utilities.

### Utilities

The operator observability package will also have utilities to:
- Generate the documentation for metrics, alerts, and recording rules
- Lint metrics and recording rules names
- Validate alerts configuration
- Generate resource definitions PrometheusRule, ServiceMonitor, etc
- Define allow-list, deny-list, and opt-in metrics

## Action Plan

### Decouple monitoring logic from business logic

1) Move the monitoring logic from the business logic code to a dedicated package
in each component repository

### Use a generic operator-observability package

1) Create the operator-observability package with the generic monitoring code
   and utilities

2) Document how to use the operator-observability package for adding new
   metrics, alerts, and recording rules

3) Update the monitoring code in each component to use the 
   operator-observability package
