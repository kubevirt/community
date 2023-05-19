# Overview
This proposal outlines the establishment of an API review process within kubevirt project. The primary objective of this
process is to ensure the quality, consistency, and usability of the project's APIs, enabling developers to effectively
interact with our software. By implementing an API review process, the aim is to enhance the project's overall
development experience, encourage collaboration, and foster community engagement.

Kubernetes core APIs have been very successful in maintaining strong backward compatibility along with good usability
over the years. This proposal uses the Kubernetes api review process as a guide for implementing a similar process for
kubevirt.


## Motivation
Since kubevirt is reaching v1 soon and has already had some APIs with v1 version, implementing an API review process
ensures code quality, usability, stability, and fosters community collaboration, ultimately leading to a successful and
sustainable software ecosystem.  

More importantly, it allows for a common framework through which all stake-holders can discuss and decide API facing 
changes in kubevirt.

## Goals

- Establish a process for api reviews such that:
  - contributors have a clear idea about how to implement an API facing change
  - reviews have a guideline about the necessary checks required for a successful API facing change
  - community has a guideline about how to handle API breakages upon upgrade
  - propose necessary tools to make process easier and reliable, and avoid human errors

## Non Goals
- Upgrades breaking due to behavioural changes, e.g. change in implementation of a controller

## Definition of Users
- API Reviewers
- Contributors

## User Stories
- As a user of kubevirt project, how can I feel confident that kubevirt maintains an intuitive, stable and crisp API that
  can be used as a foundational block to build products and projects.
- As a contributor, how do I know the right way to approach API facing change?
- As a contributor, how do I know get feedback about an API facing change?
- As a reviewer, what are the list of checks needed for approving an API facing change?
- As a reviewer, what tools could be used to make api reviews easier and reliable?

## Repos
- https://github.com/kubevirt/kubevirt

# Design

In order to achieve the stability, quality and consistency of APIs like the core Kubernetes APIs this document proposes
the following changes:

- Process for api reviewers: Process changes to add more resources and visibility for api reviews
- Process for contributors: Documenting and sharing a guide for adding api facing changes to kubevirt
- Tools and tests: Tools and automation that can be helpful to reduce human burden and errors to carry-out api changes

## Process for api reviewers
Recent changes to reviewer guidelines recommend forming small groups in specific areas of expertise. sig-api-reviews is 
one such group. This group will be responsible for:
- Reviewing all the PRs with `kind/api-change` labels
- Maintaining a high quality, stable and crisp api surface that is backward compatible

### How to achieve this?

Kubernetes has a very well-defined process for api-reviews. Taking an inspiration from that kubevirt should have the 
following:

- A specific community call for sig-api-reviewer
- In this call all the PRs with api-facing changes will be reviewed. Check list of items to check in the call
  - does the PR introduce breaking changes?
  - can the API changes be better?
  - Any other communication that is needed for contributor to move forward.
- While goal of the project is to prohibit introducing API breaking changes, automation tools proposed in the Tools section
  will help in attaining this goal. However some APIs have reached v1 (GA) without any such tool. Hence a well-defined 
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
  - Any api-facing change with design doc can be reviewed by the sig-api-reviewers for initial feedback.
  - Any api-facing change PR can be brought up for the discussion in sig-api-reviewers call
  - Link to a conventions document for good practices and guidance

## Tools for reviewers

The easiest part about checking for good api-facing change is one that maintains backward compatibility. API objects are
serialized to and from JSON when clients interact with API server. The testing for serialization could be automated to 
make sure that APIs continue to be serializable upon upgrades

### Tool to test serialization upon upgrade
Here is a [demo](https://github.com/alaypatel07/kubevirt-api-fuzzer) tool that identifies api breakages while reading 
older objects using newer clients. 

##### Description and Usage
1. This tool creates JSON and YAML files for all the API exposed by kubevirt in group-version "kubevirt.io/v1",
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

During KubeVirt upgrade, the apiserver is updated last, i.e. for a moment in time until the upgrade rolls out, kubevirt
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
(How/if this design will get broken up into multiple phases)
