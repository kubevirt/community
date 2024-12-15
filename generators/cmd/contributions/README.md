# contributions

A cli tool that fetches GitHub contributions of a user either per organization or per repository.

This tool was built with the focus on helping guide decisions for community administrators about user inactivity. Concrete applications are:
* generating evidence that a specific user has been active in a repository or organization
* finding inactive community members for a GitHub organization
* finding inactive reviewers in a GitHub repository

There are three flags that are complementary to each other when using the tool:
* `--username`: the username to check for contributions. This is the most basic use case - checking the contributions of a single user
* `--orgs-file-path`: file path to the orgs.yaml file, which defines the organization structure (see [peribolos] from Kubernetes). This checks the contributions of all organization members defined in target organization.
* `--owners-file-path`: file path to an [OWNERS] file that we want to check reviewers and approvers from, possibly accompanied by `--owners-aliases-file-path` to make given aliases resolvable

The flags `--org` and `--repo` then target what contributions to check, where omitting `--repo` will generate a report over contributions for the org vs. contributions in a given repository. 

# `--username`

The username check aims to quickly check contributions of a specific github user.

## per organization

Example:
```bash
$ go run ./generators/cmd/contributions \
    --github-token /path/to/oauth \
    --months 12 \
    --username dhiller
{"level":"info","msg":"creating report for user \"dhiller\"","time":"2024-12-05T15:01:07+01:00"}
activity log:
    user:          dhiller
    org:           kubevirt
    since:         2023-12-05 15:01:07

    issues
        created:   59
        commented: 74
    pull requests:
        reviewed:  512
        created:   182
        commented: 768
    commits:       180
{"level":"debug","msg":"user activity log: \"/tmp/user-activity-dhiller-kubevirt-3984018412.yaml\"","time":"2024-12-05T15:01:13+01:00"}
$ # showing user contribution details
$ head -20 /tmp/user-activity-dhiller-kubevirt-3984018412.yaml
issuesCreated:
    issueCount: 59
    nodes:
        - issue:
            number: 13436
            title: 'arm lane: clustered failure'
            URL: https://github.com/kubevirt/kubevirt/issues/13436
            repository:
                name: kubevirt
            author:
                login: dhiller
            createdAt: 2024-12-04T10:11:01Z
        - issue:
            number: 13435
            title: Issue of postcopy migration with the check-tests-for-flakes lane
            URL: https://github.com/kubevirt/kubevirt/issues/13435
            repository:
                name: kubevirt
            author:
                login: dhiller
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
{"level":"info","msg":"creating report for user \"dhiller\"","time":"2024-12-05T15:05:54+01:00"}
activity log:
    user:          dhiller
    repository:    kubevirt/project-infra
    since:         2023-12-05 15:05:54

    issues
        created:   17
        commented: 15
    pull requests:
        reviewed:  247
        created:   124
        commented: 333
    commits:       119
{"level":"debug","msg":"user activity log: \"/tmp/user-activity-dhiller-kubevirt_project-infra-2764291029.yaml\"","time":"2024-12-05T15:06:05+01:00"}
$ # showing user contribution details
$ head -20 /tmp/user-activity-dhiller-kubevirt_project-infra-2764291029.yaml
issuesCreated:
    issueCount: 17
    nodes:
        - issue:
            number: 3786
            title: 'flakefinder: live filtering in report - exclude lane'
            URL: https://github.com/kubevirt/project-infra/issues/3786
            repository:
                name: project-infra
            author:
                login: dhiller
            createdAt: 2024-11-27T10:39:57Z
        - issue:
            number: 3768
            title: 'prowjob: remove outdated job config for kubevirt versions 0.3x.xx'
            URL: https://github.com/kubevirt/project-infra/issues/3768
            repository:
                name: project-infra
            author:
                login: dhiller
```
# `--orgs-file-path`

The orgs-file-path check is targeted to produce machine consumable output for later consumption by other processes. Therefore the flag `--report-output-file-path` is used to write the report output file and consume the `.report.inactiveUsers` yaml element.

```bash
$ go run ./generators/cmd/contributions \
    --github-token /path/to/oauth \
    --orgs-file-path ../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml \
    --report-output-file-path /tmp/contributions-report.yaml
{"level":"debug","msg":"active user: 0xFelix","time":"2024-12-05T15:36:59+01:00"}
{"level":"debug","msg":"skipping user aburdenthehand (reason: invisibleContributions)","time":"2024-12-05T15:36:59+01:00"}
{"level":"debug","msg":"active user: acardace","time":"2024-12-05T15:37:06+01:00"}
{"level":"debug","msg":"active user: Acedus","time":"2024-12-05T15:37:12+01:00"}
{"level":"debug","msg":"active user: aerosouund","time":"2024-12-05T15:37:17+01:00"}
{"level":"debug","msg":"active user: aglitke","time":"2024-12-05T15:37:21+01:00"}
{"level":"debug","msg":"active user: akalenyu","time":"2024-12-05T15:37:27+01:00"}
...
{"level":"info","msg":"inactive user: jobbler","time":"2024-12-05T15:41:02+01:00"}
{"level":"debug","msg":"user activity log: \"/tmp/user-activity-jobbler-kubevirt-2735337715.yaml\"","time":"2024-12-05T15:41:02+01:00"}
...
{"level":"debug","msg":"active user: nunnatsa","time":"2024-12-05T15:42:21+01:00"}
{"level":"debug","msg":"skipping user openshift-ci-robot (reason: bots)","time":"2024-12-05T15:42:21+01:00"}
{"level":"debug","msg":"skipping user openshift-merge-robot (reason: bots)","time":"2024-12-05T15:42:21+01:00"}
...
{"level":"debug","msg":"active user: zhlhahaha","time":"2024-12-05T15:44:32+01:00"}
inactive users:
- gouyang
- jobbler
- VirrageS
$ # show full contribution report (directed by --report-output-file-path , see command above)
$ cat /tmp/contributions-report.yaml
reportOptions:
    org: kubevirt
    repo: ""
    username: ""
    githubTokenPath: /home/dhiller/.tokens/github/kubevirt-bot/oauth
    months: 6
    orgsConfigFilePath: ../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml
    ownersFilePath: ""
    reportAll: false
    reportOutputFilePath: /tmp/contributions-report.yaml
    ownersAliasesFilePath: ""
reportConfig: # see default-config.yaml
    skipInactive:
        kubevirt:
            - name: bots
              github:
                - kubevirt-bot
                - kubevirt-commenter-bot
                - ...
            - name: orgAdmins
              ...
        ...
result:
    activeUsers:
        - 0xFelix
        - acardace
        - ...
    inactiveUsers:
        - gouyang
        - jobbler
        - VirrageS
    skippedUsers:
        bots:
            - kubevirt-commenter-bot
            - openshift-ci-robot
            - openshift-merge-robot
        invisibleContributions:
            - aburdenthehand
            - jberkus
log:
    - |
      activity log:
          user:          gouyang
          org:           kubevirt
          since:         2024-06-05 15:40:21

          issues
              created:   0
              commented: 0
          pull requests:
              reviewed:  0
              created:   0
              commented: 0
          commits:       0
    - activity log written to "/tmp/user-activity-gouyang-kubevirt-*.yaml"
    ...
```

# `--owners-file-path`

The owners-file-path is targeted to check a specific [OWNERS] file and produce a machine-consumable output by setting the flag `--report-output-file-path`  and consuming the `.report.inactiveUsers` yaml element.

## automatic OWNERS alias resolution

For [OWNERS] files using aliases and having an adjacent `OWNERS_ALIASES` file (most likely in the root directory of the repository) those aliases will automatically get resolved into the list items of the alias list. If the [OWNERS] files are referring to an `OWNERS_ALIASES` file in a different location, the flag `--owners-aliases-file-path` needs to get set with the path to that file. 

```bash
$ go run ./generators/cmd/contributions \
    --github-token /path/to/oauth \
    --repo project-infra \
    --owners-file-path ../project-infra/OWNERS \
    --report-output-file-path /tmp/contributions-report.yaml
{"level":"debug","msg":"active user: aglitke","time":"2024-12-05T16:02:52+01:00"}
...
{"level":"debug","msg":"active user: xpivarc","time":"2024-12-05T16:03:56+01:00"}
inactive users:
[]
$ # show the report
$ cat /
reportOptions:
    org: kubevirt
    repo: project-infra
    username: ""
    githubTokenPath: /home/dhiller/.tokens/github/kubevirt-bot/oauth
    months: 6
    orgsConfigFilePath: ../project-infra/github/ci/prow-deploy/kustom/base/configs/current/orgs/orgs.yaml
    ownersFilePath: ../project-infra/OWNERS
    reportAll: false
    reportOutputFilePath: /tmp/contributions-report.yaml
    ownersAliasesFilePath: ""
reportConfig:
    skipInactive:
        kubevirt:
            - name: bots
              github:
                - kubevirt-bot
                - kubevirt-commenter-bot
                - ...
...
result:
    activeUsers:
        - brianmcarey
        - davidvossel
        - dhiller
        - enp0s3
        - phoracek
        - tiraboschi
        - vladikr
        - xpivarc
    inactiveUsers:
        - aglitke
        - rmohr
    skippedUsers: {}
log:
    - |
      activity log:
          user:          aglitke
          repository:    kubevirt/project-infra
          since:         2024-06-05 16:12:38

          issues
              created:   0
              commented: 0
          pull requests:
              reviewed:  0
              created:   0
              commented: 0
          commits:       0
    - activity log written to "/tmp/user-activity-aglitke-kubevirt_project-infra-*.yaml"
    ...
```

# automated query retry

Sometimes there might appear error messages indicating that a query failed, likely (at the time of writing) with a 502 or 504 http error. In general every query will get retried a number of times, after which the tool will give up and display the causing error.

[peribolos]: https://docs.prow.k8s.io/docs/components/cli-tools/peribolos/
[OWNERS]: https://www.kubernetes.dev/docs/guide/owners/