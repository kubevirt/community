# Overview
This proposal outlines the establishment of an API review process within KubeVirt project. The primary objective of this
process is to ensure the quality, consistency, and usability of the project's APIs, enabling developers to effectively
interact with our software. By implementing an API review process, the aim is to enhance the project's overall
development experience, encourage collaboration, and foster community engagement.

Kubernetes core APIs have been very successful in maintaining strong backward compatibility along with good usability
over the years. This proposal uses the Kubernetes API review process as a guide for implementing a similar process for
KubeVirt.


## Motivation
Since KubeVirt has reached v1 and has some APIs with v1 version, implementing an API review process
ensures code quality, usability, stability, and fosters community collaboration, ultimately leading to a successful and
sustainable software ecosystem.  

More importantly, it allows for a common framework through which all stake-holders can discuss and decide API facing 
changes in KubeVirt.

Currently, the project has a very hand-wavy mechanism for contributing, reviewing and discussion API facing changes. This
document attempts to specific focus areas where improvement is needed in the Goals section below and proposal for 
implementing the improvements in the design section.


## Goals

- Establish a policy for API evolution
- Establish a process for API reviews such that:
  - contributors have a clear idea about how to implement an API facing change
  - reviews have a guideline about the necessary checks required for a successful API facing change
  - community has a guideline about how to handle API breakages upon upgrade
    - API fields in an object could belong to different features. Depending on the maturity of the feature, the field 
      could evolve differently. 
  - propose necessary tools to make process easier and reliable, and avoid human errors

## Non Goals
- Upgrades breaking due to behavioural changes, e.g. change in implementation of a controller

## Definition of Users
- API Reviewers
- Contributors

## User Stories
- As a stakeholder of KubeVirt project, I would like the project to have a crip policy of how APIs will evolve across
  different versions.
- As a KubeVirt cluster operator, consuming the API by creating objects (manually or programmatically), I want KubeVirt to
   maintain an intuitive, stable and simple API, that can be used as a foundational block to build products and projects
- As a contributor, I need guidance on the right way to approach API facing change. Ideally this guidance should include
    all the steps: design docs, contributing the change and post contribution steps
- As a reviewer, I need to have a comprehensive list of checks needed for approving an API facing change
- As a reviewer, ideally, I need to be able to leverage automation wherever possible to make reviewing easier

## Repos
- https://github.com/kubevirt/kubevirt

# Design

In order to achieve the stability, quality and consistency of APIs like the core Kubernetes APIs this document proposes
the following changes:

- For all stakeholders: Describe a policy of API evolution
- For API reviewers: There should be a one or more engineers that review api breaking changes on a regular basis.
- For contributors: There should be a guide explaining how to merge and api change
- Tools and tests: Tools and automation that can be helpful to reduce human burden and errors to carry-out api changes

More details on each of the aforementioned items is highlighted below in separate sections.

## Policy of API Evolution
- Alpha features should always introduce an optional field to APIs that have GA'ed or in Beta
- Alpha features and hence by extensions fields that are introduced behind a feature flag, is not guaranteed to work
  in all future versions. Hence, they should not be deployed in production. This ensures that new feature development 
  work is not slowed down by requirement to maintain api-compatibility. Even though compatibility for experimental or 
  alpha APIs is not strictly required, but breaking compatibility should not be done lightly, as it disrupts all users
  of the feature.
- Beta or GA features that introduces new fields to beta or stable APIs should not break compatibility

## Contributors responsibilities
- Contributors must explain why the API change is needed, which functionality it adds and who requested it. They should
  include links to github issues, email threads or any other public document where users request this functionality.
- Contributors should contact the users who requested the functionality via the channel they have used, to give them the
  opportunity to review the API. We do not want to define an API that the user finds inconvenient.

## Process for api reviewers
Recent changes to reviewer guidelines recommend forming small groups in specific areas of expertise. sig-api is 
one such group. This group will be responsible for:
- Reviewing all the PRs with `kind/api-change` labels
- Maintaining a high quality, stable and crisp api surface that is backward compatible

The focus of this group should be to lay out the guidelines of an intuitive, maintainable and usable set of APIs for
KubeVirt project. 

### How to achieve this?

Kubernetes has a very well-defined process for api-reviews. Taking an inspiration from that, KubeVirt should have the 
following:

- A specific community call for sig-api
- In this call all the PRs with api-facing changes will be reviewed. Check list of items to go through in the call
  - does the PR introduce breaking changes?
  - can the API changes be better?
  - Any other communication that is needed for contributor to move forward.
- While goal of the project is to prohibit introducing API breaking changes, automation tools proposed in the Tools section
  will help in attaining this goal. However, some APIs have reached v1 (GA) without any such tool. Hence, a well-defined 
  process is needed to address breaking changes discovered in previous versions:
  - If a break in API is reported, the next release:
    - will introduce the fix
    - if the fix is burdensome, the next release will have deprecation warnings
    - Depending on the support burden, after letting the fix is in place and appropriate deprecation warnings raised for
    a minimum of 3 releases (around a year), the community can start the process of removing the deprecated fields using
    the normal API removal process.

## Process for contributors

In order for the process to work efficiently, contributors should receive the right support and guidance when 
contributing api-facing changes. 

### How to achieve this?

- Have a document to describe the right process for contributors with the following details:
  - Any API-facing change with design doc can be reviewed by the sig-api for initial feedback.
  - Any API-facing change PR can be brought up for the discussion in sig-api call
  - Link to a conventions document for good practices and guidance

## Tools for reviewers

The easiest part about checking for good api-facing change is one that maintains backward compatibility. API objects are
serialized to and from JSON when clients interact with API server. The testing for serialization could be automated to 
make sure that APIs continue to be serializable upon upgrades

### Tool to test serialization upon upgrade
Here is a [demo](https://github.com/alaypatel07/KubeVirt-api-fuzzer) tool that identifies api breakages while reading 
older objects using newer clients. 

##### Description and Usage
1. This tool creates JSON and YAML files for all the API exposed by KubeVirt in group-version "kubevirt.io/v1",
   versioned by the release. The current version is in `HEAD` directory, previous versions are in `release-0.yy` release
   directory. The following APIs are included, more APIs can be added in the future:
    ```
    VirtualMachineInstance
    VirtualMachineInstanceList
    VirtualMachineInstanceReplicaSet
    VirtualMachineInstanceReplicaSetList
    VirtualMachineInstancePreset
    VirtualMachineInstancePresetList
    VirtualMachineInstanceMigration
    VirtualMachineInstanceMigrationList
    VirtualMachine
    VirtualMachineList
    KubeVirt
    KubeVirtList
    ```
2. Upon any change to API, the json and YAML files will be updated in HEAD directory.
3. When KubeVirt cuts a new release of the project, the files in HEAD directory will be copied to the release version and
   future development branch will add a unit test for past two releases:
    ```
    $ VERSION=release-0.60
    $ cp -fr testdata/{HEAD,${VERSION}} 
    ```
   
##### How will it help?

During KubeVirt upgrade, the apiserver is updated last, i.e. for a moment in time until the upgrade rolls out, KubeVirt
components like virt-handler, virt-controller will have newer client, but apiserver will be serving older objects.

Using this tests, it can be asserted that the current newer clients can roundtrip (serialized and de-serialized) past
two releases, which will make the upgrade safer.


## API Examples
Tool usage example: https://github.com/alaypatel07/kubevirt-api-fuzzer#usage

## Scalability
TODO

## Update/Rollback Compatibility
(does this impact update compatibility and how)

## Functional Testing Approach
(an overview on the approaches used to functional test this design)

# Implementation Phases

The process for achieving api-stability will involve the following: 

### Phase 1

- creation of a sig-api that will meet on a regular cadence for the following:
  - triage all the open PRs with `kind/api-change` label
  - interested reviewers will pick up PRs and design documents for review
  - discuss common observed patterns in the api reviews
  - the PRs are not gated on api-review approval
- This is the initial phase where api reviews will be tried in the community. This phase will last atleast 1 release, 
  currently planned for release 1.2

### Phase 2

- develop automation to help PR reviewers find breaking changes
- PRs will be gated on api-review approval
- identify additional work like e2e upgrade test to enhance API stability

### Phase 3
- implementation and addition of upgrade tests