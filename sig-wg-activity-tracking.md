# Proposal: Systematic SIG and WG Activity Tracking for KubeVirt

## Problem Statement

As KubeVirt matures as a CNCF Incubating project (accepted September 2019, moved to Incubating in April 2022), maintaining healthy SIGs and WGs requires knowing who is actively participating and who has moved on. KubeVirt already has significant infrastructure for this — a centralized `sigs.yaml` registry with `emeritus_leads` support, defined inactivity thresholds in `membership_policy.md`, a contribution checker tool, and a quarterly Prow periodic that automatically removes inactive org members and adds them to the alumni list. These are strong foundations.

However, this automation currently covers **org membership** only. SIG/WG leadership, reviewer/approver roles in OWNERS_ALIASES, and overall SIG health are not yet subject to the same automated lifecycle. Some SIGs have no active chairs listed. There is no regular process to verify that SIG leads are still engaged, or to retire SIGs that have gone dormant.

Without extending the existing automation to cover these areas, we risk:

- **Stale leadership records** — chairs and leads who have changed roles or left the project remain listed, creating confusion about who is actually steering a SIG or WG.
- **Blocked decision-making** — when listed leads are unreachable, approvals, design reviews, and charter updates stall.
- **Contributor discouragement** — active contributors see no clear path to step up when leadership seats are effectively vacant.
- **Governance credibility** — as a CNCF Incubating project with graduation in view, our governance documents should reflect reality.

The Kubernetes community has solved these problems with a comprehensive lifecycle for SIG/WG membership, automated inactivity detection at all levels, emeritus transitions, and annual health reporting. We should extend our existing automation to cover similar ground, tailored to KubeVirt's scale.

## What Kubernetes Does

The Kubernetes community maintains a comprehensive system for tracking SIG and WG health. The key components worth studying are:

### Centralized Registry (`sigs.yaml`)

KubeVirt already has this. Kubernetes uses an identical format — the `sigs.yaml` file is the single source of truth for every SIG, WG, and committee, including leadership roles, meeting schedules, contact information, and subproject ownership.

### Defined Inactivity Thresholds

Kubernetes defines inactivity with concrete numbers:

- **Organization members**: removed after 12 months with no contributions (measured via CNCF DevStats).
- **OWNERS file reviewers/approvers**: flagged inactive with fewer than 10 DevStats contributions _and_ fewer than 10 PR comments within 12 months. Approvers are moved to `emeritus_approvers`; reviewers are removed.
- **SIG leads**: expected to coordinate absences longer than 1 month. Removal after >1 month of no communication or >1 month not fulfilling duties, via super-majority vote of remaining leads.

### Emeritus Status

Rather than simply deleting inactive people, Kubernetes uses an emeritus process that preserves institutional knowledge:

- `emeritus_approvers` in OWNERS files — retains history, removes merge authority.
- `emeritus_leads` in `sigs.yaml` — distinguished from active leads by the absence of a `company` field.

This matters because it signals "this person contributed significantly" while making clear they no longer hold decision-making authority.

### SIG Health Signals and Retirement

Kubernetes has explicit retirement triggers:

- After **3+ months** of missing consistent meeting quorum: the SIG _should_ retire.
- After **6+ months** of missing consistent meeting quorum: the SIG _must_ retire.

### Annual Reports

Every January, the Steering Committee generates report templates for each SIG. Chairs complete them by mid-February covering current initiatives, project health, and membership metrics (Slack members, mailing list subscribers, meeting attendees, unique reviewers/approvers). The Steering Committee publishes a project-wide summary by end of March.

### Generator Tooling

A Go-based generator validates `sigs.yaml` structure, produces per-SIG README files, compiles `OWNERS_ALIASES`, and generates annual report templates. This automation catches inconsistencies early and reduces manual bookkeeping.

### GitHub Team Sync (Peribolos)

Kubernetes uses Peribolos to declaratively manage GitHub organization membership and team composition from YAML configuration. Team definitions flow from `sigs.yaml` leadership into GitHub teams automatically.

## Where KubeVirt Stands Today

KubeVirt already has many of the building blocks in place — more than one might initially expect:

| Component | Status | Details |
|-----------|--------|---------|
| `sigs.yaml` | Exists (493 lines, 10 SIGs, 4 WGs) | Includes [`emeritus_leads`](https://github.com/kubevirt/community/blob/main/sigs.yaml#L284) support (e.g. sig-buildsystem) |
| `OWNERS_ALIASES` | Exists | SIG-scoped reviewer/approver groups |
| `alumni.yaml` / `ALUMNI.md` | Exists (150+ alumni) | Automatically updated by Prow periodic job |
| Contribution checker | Exists (`generators/cmd/contributions/`) | Queries GitHub API, configurable inactivity window (default 6 months for reviewers/approvers/chairs, 12 months for org members) |
| Automated inactivity removal | **Operational** | Prow periodic [`periodic-community-remove-inactive-users-from-orgs-yaml`](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/community/community-periodics.yaml#L2) runs quarterly (March, June, September, December), removes inactive users from `orgs.yaml` and adds them to `alumni.yaml`, opening PRs automatically |
| Inactivity policy | **Defined** in [`membership_policy.md`](https://github.com/kubevirt/community/blob/main/membership_policy.md#L220) | Org members: 12 months (CNCF DevStats); Reviewers/Approvers/Chairs/Leads: 6 months |
| Presubmit validation | **Operational** | [`pull-community-make-generate`](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/community/community-presubmits.yaml#L15) runs `make generate` on every PR and fails if generated output drifts from committed state |
| Annual reports | Do not exist | No regular health check process |
| SIG retirement criteria | Not defined | Some SIGs appear dormant (no chairs, no meetings) |

### What is already automated

The quarterly Prow periodic job is worth highlighting because it closes the loop from detection to action:

1. Runs the contribution checker with `--months 12` against the full org membership.
2. For each inactive user, removes them from `orgs.yaml` (in `project-infra`).
3. Adds them to `alumni.yaml` (in `community`) and regenerates `ALUMNI.md`.
4. Opens PRs in both repos, tagging the affected users and SIG buildsystem for review.

This means org-level membership hygiene is already handled. The presubmit `make generate` job ensures that `sigs.yaml` changes stay consistent with generated files (sig-list, OWNERS_ALIASES, etc.).

### Where the gaps remain

The automation covers **org membership** well, but does not yet address:

- **SIG/WG leadership lifecycle** — chairs and leads who become inactive are not automatically flagged or transitioned. The `emeritus_leads` field exists in `sigs.yaml` but the transition is manual and ad-hoc.
- **Reviewer/approver cleanup** — the contribution checker can identify inactive reviewers and approvers, but there is no periodic job that acts on this for OWNERS_ALIASES the way the org membership job does for `orgs.yaml`.
- **SIG health monitoring** — no regular process to check whether SIGs are holding meetings, maintaining quorum, or have sufficient leadership.
- **SIG retirement criteria** — no defined process for winding down a dormant SIG.

## Proposal

Building on the existing foundation, the following proposals address the remaining gaps — SIG/WG leadership lifecycle, reviewer/approver cleanup, health monitoring, and retirement criteria.

### 1. Automate SIG/WG Leadership Inactivity Detection

The quarterly Prow periodic already handles org membership. Extend it (or create a companion job) to also check `sigs.yaml` chairs and leads against the contribution checker:

- Run the contribution checker with `--owners-file-path` against `sigs.yaml` leadership entries.
- For each chair or lead inactive for 6+ months, open an issue in `kubevirt/community` tagging the person and the relevant SIG.
- Give the person 30 days to respond before a maintainer opens a PR to move them to `emeritus_leads`.

This closes the gap between "we can detect inactive leads" and "we act on it." The `emeritus_leads` field already exists in `sigs.yaml` (e.g., Brian Carey in sig-buildsystem) — the missing piece is the automated trigger.

### 2. Automate Reviewer/Approver Cleanup in OWNERS_ALIASES

Similarly, extend the periodic job to scan OWNERS_ALIASES for inactive reviewers and approvers:

- **Reviewers** inactive for 6+ months: open a PR to remove them.
- **Approvers** inactive for 6+ months: open a PR to move them to an `emeritus_approvers` section (following the Kubernetes OWNERS convention).

```yaml
# OWNERS_ALIASES — proposed emeritus section
sig-example-approvers:
  - active-approver
sig-example-emeritus-approvers:
  - former-approver    # 2026-06-23
```

This ensures that OWNERS_ALIASES reflects actual review capacity, not historical participation.

### 3. Establish SIG/WG Health Checks

Introduce a lightweight annual (or semi-annual) health check. Unlike Kubernetes' detailed annual report, KubeVirt can start with a simpler format.

**Per-SIG/WG checklist (filed as a GitHub issue template):**

- [ ] At least 2 active chairs listed in `sigs.yaml`
- [ ] Regular meetings held in the past 6 months (with notes/recordings)
- [ ] Subproject OWNERS files have at least 2 approvers each
- [ ] No single-company concentration in leadership (aspiration, not hard requirement)
- [ ] Charter is current and reflects actual scope

SIGs that cannot satisfy the first three items for two consecutive check-ins should enter a probationary period, with retirement as the eventual outcome if health does not improve.

This could be generated automatically by a Prow periodic (similar to how Kubernetes' Steering Committee generates annual report templates in January) or filed manually by maintainers.

### 4. Define SIG/WG Retirement Criteria

Add to `GOVERNANCE.md`:

- **3 months** without a quorum at meetings or without active chairs: maintainers should discuss the SIG's future at the community meeting.
- **6 months** without recovery: the SIG should be formally retired. Its subprojects are either absorbed by another SIG or archived.

Retirement is not failure — it is recognition that the community's energy has shifted. Retired SIGs can be restarted if interest returns.

### 5. Enhance the Generator Tooling

The existing `make generate` pipeline (enforced by the [`pull-community-make-generate`](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/community/community-presubmits.yaml#L15) presubmit) already validates that generated output stays in sync. Extend the generators to:

- **Validate `sigs.yaml`** for required fields (at least 2 chairs, meeting info, charter link) — failing the presubmit if a SIG is missing critical data.
- **Generate per-SIG README files** from `sigs.yaml` (Kubernetes does this; it keeps docs in sync with the central registry).
- **Cross-reference OWNERS_ALIASES with `sigs.yaml`** to detect inconsistencies between leadership entries and alias group membership.

### 6. Formalize the Transition Process

When the contribution checker or a health check flags someone as inactive:

1. **Notify**: Open a PR or issue mentioning the person, giving them 30 days to respond.
2. **Transition**: If no response, move them to emeritus status via PR (chairs/leads to `emeritus_leads` in `sigs.yaml`, approvers to `emeritus_approvers` in OWNERS_ALIASES).
3. **Announce**: Note the transition at the community meeting — frame it as gratitude, not removal.
4. **Update alumni.yaml**: If the person has fully departed the project, the existing quarterly periodic will handle this automatically.

This process respects contributors' time while keeping governance files accurate. The key improvement over the current state is that steps 1-3 happen proactively via automation rather than waiting for someone to notice.

## Immediate Next Steps

1. **Audit current state** — Run the contribution checker against all chairs, leads, and approvers to identify who is active and who has moved on.
2. **Extend the quarterly periodic** — Add OWNERS_ALIASES and `sigs.yaml` leadership scanning to the existing Prow job (or create a companion job).
3. **File health check issues for SIGs without active chairs** — sig-architecture, sig-documentation, sig-testing need immediate attention.
4. **Draft retirement criteria for `GOVERNANCE.md`** — Propose the 3-month / 6-month thresholds via PR.
5. **Enhance `make generate` validation** — Add checks for minimum chair count and required fields in `sigs.yaml`.

## Related Work

- Kubernetes community governance: [kubernetes/community](https://github.com/kubernetes/community)
- Kubernetes SIG lifecycle: [sig-wg-lifecycle.md](https://github.com/kubernetes/community/blob/master/sig-wg-lifecycle.md)
- Kubernetes annual reports: [annual-reports.md](https://github.com/kubernetes/community/blob/master/committee-steering/governance/annual-reports.md)
- KubeVirt contribution checker: [generators/cmd/contributions/](https://github.com/kubevirt/community/tree/main/generators/cmd/contributions)
- KubeVirt membership policy: [membership_policy.md](https://github.com/kubevirt/community/blob/main/membership_policy.md)
- KubeVirt inactivity periodic: [community-periodics.yaml](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/community/community-periodics.yaml)
- KubeVirt presubmit validation: [community-presubmits.yaml](https://github.com/kubevirt/project-infra/blob/main/github/ci/prow-deploy/files/jobs/kubevirt/community/community-presubmits.yaml)
- KubeVirt CNCF project page: [cncf.io/projects/kubevirt](https://www.cncf.io/projects/kubevirt/)
