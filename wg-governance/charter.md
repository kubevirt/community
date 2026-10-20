# WG Governance

## Scope

The Governance Working Group coordinates the implementation of automated governance tooling and processes to ensure consistency between the project's governance definitions (sigs.yaml), GitHub organization structure (orgs.yaml), and code review permissions (OWNERS_ALIASES).

## Problem Statement

KubeVirt currently maintains governance data in multiple locations:
- GitHub organization membership and teams in `orgs.yaml` (project-infra)
- SIG/WG structure, leadership, and roles in `sigs.yaml` (community)
- Code review permissions in `OWNERS_ALIASES` files (multiple repos)

This creates maintenance burden and drift:
- When members leave, they may be removed from orgs.yaml but remain in sigs.yaml or OWNERS_ALIASES
- Additional roles (reviewer/approver/chair) may not be properly cleaned up
- No automated synchronization exists between these sources
- OWNERS and OWNERS_ALIASES files are manually maintained

## Goals

### Short-term
- Implement automated governance validation and sync tooling (`sig-sync`)
- Establish sigs.yaml (or per-group group.yaml) as the single source of truth
- Automate syncing from sigs.yaml to OWNERS_ALIASES and GitHub teams
- Implement drift detection and reporting

### Long-term
- Create systematic SIG/WG activity tracking
- Establish clear lifecycle management for leadership roles
- Reduce manual governance maintenance overhead
- Enable SIGs/WGs to self-manage their group.yaml with automated approval workflows

## In Scope

- Design and implementation of `sig-sync` tooling (see project-infra#5317)
- Migration strategy from current manual processes to automated sync
- Validation that governance data remains consistent across systems
- Documentation of governance automation for contributors and SIG chairs
- Extraction of group data into per-SIG/WG files (see community#451)
- SIG/WG activity tracking and health metrics (see community#449)

## Out of Scope

- Changes to the fundamental governance model (maintainers, voting, etc.) — governed by GOVERNANCE.md
- SIG/WG creation/dissolution decisions — those remain with maintainers
- Individual SIG/WG technical direction

## Deliverables

1. **sig-sync tooling**: CLI tool for validating and syncing governance data
2. **Documentation**: Migration guide, SIG chair guide, architecture docs
3. **Automation infrastructure**: Prow periodic jobs for drift detection and automated sync
4. **Per-group data extraction**: Transition from monolithic sigs.yaml to per-group group.yaml files

## Timeline

The WG will disband when:
- sig-sync tooling is implemented, tested, and in production
- All SIGs/WGs have migrated to the new structure
- Automated syncing is proven stable over at least 2 quarterly cycles
- Documentation is complete and SIG chairs are trained

**Estimated timeline**: 6-12 months

## Governance

### Stakeholder SIGs

- SIG CI
- SIG Buildsystem
- All SIGs (consumers of the governance automation)

### Roles and responsibilities

#### Chairs

* @aburdenthehand
* @dhiller

#### Additional reviewers

* @iholder101
* @xpivarc

## References

- [project-infra#5317: sig-sync automation proposal](https://github.com/kubevirt/project-infra/issues/5317)
- [community#451: Extract sig data into per-group files](https://github.com/kubevirt/community/pull/451)
- [community#449: Systematic SIG/WG activity tracking](https://github.com/kubevirt/community/pull/449)
- [Kubernetes SIG governance model](https://github.com/kubernetes/community/blob/master/governance.md)
