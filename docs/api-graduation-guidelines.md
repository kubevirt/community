This document defines the expectations of how API graduation is handled across the github.com/kubevirt organization. Individual projects within the KubeVirt org may decide to append additional criteria and timeline constraints to the items outlined here, but should not deviate from the broad expectations as defined in this document.

# Graduation Phase Expecations

## Alpha
* Feature incomplete, meaning the API might be half implemented as functionality lands in phases.
* Basic end to end tests are merged that validate the application flow behaves as expected.
* No backwards compatibility guarantees meaning API can mutate and be even be removed without warning.
* Available for early adopters to test and provide initial feedback.

## Beta
* User guide documentation exists.
* Has early adopter commitment, meaning at least one end user or vendor has provided feedback that the feature meets their needs and will be adopted. 
* Contains well tested end to end tests that include upgrade tests that guarantee forward compatibility.
* Stable API, meaning no backwards incompatible API changes will occur.
* Support for the API will not be dropped without a lengthy deprecation phase.

## General Availability
* Core functionality is feature complete, meaning the API isn't lacking any critical functionality.
* Proven in production use cases.

# Graduation Timelines Guidelines

## Alpha -> Beta graduation timeline guidelines
* Alpha APIs should exist in an official release for adopter feedback to occur before beta graduation.
* An API that hasn’t graduated to beta within a year should either be deprecated or graduated.

## Beta -> GA graduation timeline guidelines
* A beta API should exist in a multiple releases to allow adoption of the API and proven stability in production across updates before GA graduation.
* A beta API known to be used in production should be graduated to GA within a year.

