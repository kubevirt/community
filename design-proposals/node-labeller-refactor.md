# Overview

Node-labeller is a subcomponent of virt-handler responsible for two key tasks:

1. **Inferring host capabilities**: It gathers information such as  `host capabilities`, which is shared with the VM Controller during its setup, creating an unnecessary coupling between these components.
2. **Labeling nodes**: Based on the inferred capabilities, node-labeller labels nodes at startup, periodically, or when triggered by a `kubevirt` change callback.

## Motivation

The node-labeller has deviated from its primary purpose of simply labeling nodes. It now also handles inferring host capabilities for the VM Controller, a task unrelated to its core function. Most of the capabilities it gathers aren't relevant to node labeling, leading to an overcomplicated design and unnecessary coupling with the VM Controller.

This has resulted in a bloated implementation, especially with the use of the `loadAll` function, which loads data for both node-labeller and VM Controller, making the code harder to test and maintain.

Additionally, node-labeller uses an internal `api` package within the `capabilities` directory that mirrors the functionality of the official libvirtxml package. This duplication exacerbates the aforementioned issues in the codebase and adds redundant layers of complexity, particularly in testing.

## Goals

* **Decouple the inference functionality**: Separate capability inference from node labeling to streamline the node-labeller’s functionality. Improve testing coverage for both labeling and capability parsing, focusing on eliminating file reads in unit tests and introducing tests for affected components that previously lacked test coverage.
* **Decouple node-labeller and VM Controller**: Deprecate the `node-labeller/api` package and replace it with libvirtxml, ensuring that the VM Controller obtains its required capabilities without relying on node-labeller.

## Non-Goals

* This proposal does **not** aim to extend the functionality of node-labeller or VM Controller, such as adding new labels or introducing more parsing options.

## Definition of Users

This change is not user-facing but affects all users indirectly. The node-labeller’s original functionality remains intact, so no user intervention is required.

## Repos

[KubeVirt](https://github.com/kubevirt/kubevirt)

# Design

The redesigned node-labeller will focus solely on labeling nodes, with all capability inference logic moved to a separate package. This new package will still be part of virt-handler but will provide capabilities as a dependency for both node-labeller and VM Controller. 

The `node-capabilities.sh` script, which currently infers capabilities using the libvirt API and stores them in `/var/lib/kubevirt-node-labeller`, will be adjusted accordingly as node-labeller will no longer handle capability inference.

Node-labeller will now only use specific host-related data (e.g., CPU counters), and its structures and methods will be streamlined to reflect this simpler scope. It will no longer serve as a dependency provider for the VM Controller.

The redundant `node-labeller/api/capabilities` package will be replaced with the existing libvirtxml library. All components currently relying on this package, along with their tests, will be updated to use libvirtxml, completing the decoupling process.

To maintain flexibility, node-labeller's construction will be refactored to use the builder pattern, simplifying its structure while allowing for future extensibility.

At its end state, node-labeller should be a lean, single-file component with core logic and relevant tests.

## Update/Rollback Compatibility

Since this change does not alter the core functionality of any components, updates and rollbacks will be fully compatible.

## Functional Testing Approach

The existing functional tests in `tests/infrastructure/node-labeller.go` cover these changes, ensuring proper functionality.

# Implementation Phases

The node-labeller currently uses the `loadAll` function to infer several host capabilities during its construction. This will be refactored across the following phases:

## Phase I - Remove `loadHostCapabilities`

The `loadHostCapabilities` function populates node-labeller with host capabilities and shares them with the VM Controller. It reads the `capabilities.xml` file from `/var/lib/kubevirt-node-labeller` and unmarshals it into the `Capabilities` struct from `node-labeller/capabilities/api`.

In this phase, `loadHostCapabilities` will be removed, and the `node-labeller/capabilities/api` package will be deprecated in favor of libvirtxml. Additional tests will be added to cover previously untested functionality, such as in `pkg/virt-handler/options.go`.

## Phase II - Remove `loadDomCapabilities`

Similar to Phase I, the `loadDomCapabilities` function retrieves domain capabilities by unmarshaling `virsh_domcapabilities.xml` into the `HostDomCapabilities` struct in `node-labeller/model.go`.

This phase will remove `loadDomCapabilities` and eliminate `node-labeller/model.go`, transitioning to libvirtxml. It will also remove `loadHostSupportedFeatures`, which relies on the same structs.

## Phase III - Introduce the `node-capabilities` Package

A new lightweight `node-capabilities` package (name TBD) will be created to handle capability inference for both node-labeller and VM Controller. Any remaining capability inference logic in node-labeller will be migrated to this package. Node-labeller’s construction will be simplified to use only basic types.
