# contributions

A tool that fetches GitHub contributions by a user per organization or per repository.

## per organization

Example:
```bash
$ go run ./generators/cmd/contributions \
    --github-token /path/to/oauth \
    --months 12 \
    --username dhiller
activity log:
        user         dhiller
        organization kubevirt
        since        2023-12-02 17:50:50

        hasContributions:                    true
        totalIssueContributions:             57
        totalPullRequestContributions:       181
        totalPullRequestReviewContributions: 455
        totalCommitContributions:            199

user activity log: "/tmp/user-activity-dhiller-kubevirt-2101388502.yaml"
$ cat /tmp/user-activity-dhiller-kubevirt-2101388502.yaml
hasAnyContributions: true
totalCommitContributions: 199
totalIssueContributions: 57
totalPullRequestContributions: 181
totalPullRequestReviewContributions: 455
issueContributions:
    totalCount: 57
    nodes:
        - issue:
            URL: https://github.com/kubevirt/community/issues/359
          occurredAt: 2024-11-28T09:18:48Z
pullRequestContributions:
    totalCount: 181
    nodes:
        - pullRequest:
            URL: https://github.com/kubevirt/community/pull/361
          occurredAt: 2024-12-02T09:59:10Z
pullRequestReviewContributions:
    totalCount: 455
    nodes:
        - pullRequestReview:
            repository:
                nameWithOwner: kubevirt/kubevirt
            pullRequest:
                URL: https://github.com/kubevirt/kubevirt/pull/13274
            createdAt: 2024-11-29T09:01:07Z
            state: APPROVED
commitContributionsByRepository:
    - contributions:
        totalCount: 119
        nodes:
            - repository:
                nameWithOwner: kubevirt/project-infra
              user:
                name: Daniel Hiller
              occurredAt: 2024-11-26T08:00:00Z
...
```

## per repository

Example:
```bash
$ go run ./generators/cmd/contributions \
    --github-token /path/to/oauth \
    --months 12 \
    --repo project-infra \
    --username dhiller
activity log:
        user                            dhiller
        repository                      kubevirt/project-infra
        since                           2023-12-02 17:54:45

        issues
                created:                17
                commented:              15
        pull requests:
                reviewed:               244
                created:                124
                commented:              330
        commits:                        117

user activity log: "/tmp/user-activity-dhiller-kubevirt_project-infra-3828787132.yaml"
$ cat /tmp/user-activity-dhiller-kubevirt_project-infra-3828787132.yaml
issuesCreated:                                                          issueCount: 17
    nodes:                                                                  - issue:
            number: 3786
            title: 'flakefinder: live filtering in report - exclude
lane'
            URL: https://github.com/kubevirt/project-infra/issues/3786
            repository:                                                             name: project-infra
            author:
                login: dhiller                                                  createdAt: 2024-11-27T10:39:57Z
        - issue:
            number: 3768
            title: 'prowjob: remove outdated job config for kubevirt
 versions 0.3x.xx'
            URL: https://github.com/kubevirt/project-infra/issues/37
68
            repository:
                name: project-infra
            author:
                login: dhiller
            createdAt: 2024-11-19T10:53:26Z
        - issue:
            number: 3712
            title: Remove label needs-approver-review after either c
losed/merged or review by approver has happened
            URL: https://github.com/kubevirt/project-infra/issues/37
12
            repository:
                name: project-infra
            author:
                login: dhiller
            createdAt: 2024-10-25T11:08:15Z
        - issue:
            number: 3711
            title: Check prow update mechanism
            URL: https://github.com/kubevirt/project-infra/issues/37
11
            repository:
                name: project-infra
            author:
                login: dhiller
            createdAt: 2024-10-25T10:31:19Z
...
```
