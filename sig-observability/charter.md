# SIG Observability Charter

## Scope

SIG Observability is responsible for defining and implementing the observability
strategy for all KubeVirt components. This includes instrumenting, generating,
collecting, alerting, and exporting telemetry data such as traces, metrics, and
logs. 

### In scope

#### Code, Binaries and Services

- Defintion of observability resources (e.g. metrics, alerts, traces)
- Observability tooling and libraries such as exporters, agents, and collectors
- Observability unit and integration tests
- Observability documentation and best practices

### Out of scope

- Setting observability resources values (e.g. metric values, logging messages)
- Monitoring and alerting infrastructure
- Observability of external services

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OARP]

### Additional responsibilities of Chairs

- Uphold the KubeVirt Code of Conduct especially in terms of personal behavior
and responsibility
- Own sig-observability CI jobs
- Be welcoming to new contributors
- Resolve conflicts

[OARP]: https://stumblingabout.com/tag/oarp/
