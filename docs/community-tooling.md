# Community Repository Tooling

This guide covers the tools used to maintain this repository's SIG and alumni
documentation. For project governance and membership policies, see the
[repository overview](../README.md). Review the [AI Contribution Policy](../ai-contribution-policy.md)
when submitting AI-assisted contributions.

## Repository Layout

- `sig-*/` and `wg-*/` contain Special Interest Group and Working Group charters.
- `generators/` contains Go tools for generating community documentation.
- `validators/` contains Go tools for validating community configuration.
- `pkg/` contains shared Go libraries for these tools.
- `sigs.yaml` defines SIGs, chairs, meetings, and subprojects.
- `alumni.yaml` tracks alumni and emeritus members.

## Validate and Generate

Install Go using the version requirements in [go.mod](../go.mod). Run the
following commands from the repository root.

The validator also reads label and organization configuration from a sibling
[`kubevirt/project-infra`](https://github.com/kubevirt/project-infra) checkout at
`../project-infra/`. The default paths within that checkout are
`github/ci/prow-deploy/kustom/base/configs/current/labels/labels.yaml` and
`github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml`. Both
`make validate-sigs` and `make generate` require these files. To validate using
configuration files elsewhere, invoke the validator directly:

```bash
go run ./validators/cmd/sigs --labels-file-path /path/to/labels.yaml --orgs-file-path /path/to/orgs.yaml
```

When changing community structure, validate the SIG configuration:

```bash
make validate-sigs
```

This runs `go run ./validators/cmd/sigs` in its default dry-run mode. The validator
checks ownership links against GitHub, so it requires network access.

To update the configuration and regenerate SIG and alumni documentation:

```bash
make generate
```

Unlike validation alone, this runs the validator with `--dry-run=false`, which
can update `sigs.yaml`, before running both generators. Review the resulting
configuration and generated-file changes before submitting them.

The generators can also be run individually:

```bash
go run ./generators/cmd/sigs
go run ./generators/cmd/alumni
```

See the [SIG validator](../validators/cmd/sigs/README.md),
[SIG generator](../generators/cmd/sigs/README.md), and
[alumni generator](../generators/cmd/alumni/README.md) documentation for details.

## Tests and Formatting

For changes to the Go tooling, run:

```bash
make test
make lint
```

The lint target requires a locally installed `golangci-lint`; see
[hack/lint.sh](../hack/lint.sh).

The repository's [pre-commit configuration](../.pre-commit-config.yaml) defines
formatting and validation hooks, including Prettier for Markdown, Black for
Python, flake8, YAML checks, and yaspeller for Markdown spelling. With
[pre-commit](https://pre-commit.com/) installed, run the hooks on your changes:

```bash
pre-commit run --files docs/community-tooling.md
```

Replace the example path with the files you changed. The configuration is the
source of truth for hook versions and file selection.
