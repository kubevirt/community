Framework for E2E Tests
=

# Overview
This proposal is suggesting the introduction of a `Framework` to support common operations in e2e tests.

# Motivation

Kubevirt has an impressive amount of e2e test, covering a large amount scenarios and features.

Two paradigms can be used to make each test run on a clean environment:
- restore a clean situation before the execution of each test.
- make each test leave a clean situation.

Currently, in kubevirt the first option is most commonly used, but specific cleanups are also
performed after test execution. This mix can create leftovers and sometimes makes it difficult
to understand where a problem may have occurred.

It is also a symptom of a kind of non-independence of the tests from each other:
test A does not care about cleaning the cluster, so much as it knows that there will be
the setup of test B will do it.

Last but not least, each test file must specify the cleanup and setup function on its own, and in most cases
is something like this:
```go
BeforeEach(func() {
  tests.BeforeTestCleanup()
  
  virtClient, err = kubecli.GetKubevirtClient()
  util.PanicOnError(err)
  //other test specific setup
})
```

The `Framework`  can help resolve these issues by supporting tests in setting up and cleaning up the environment
in which they will operate. This will provide a separation of duties in the various phases of a test.
This will make it easier to script the tests since you will not have to worry about setup and cleanup
of the environment. It will also define a test architecture that is easier for new members to understand.

## Goals

- Support common operations.
- Increase e2e test stabilization.
- Minimizing the possibility of leftovers after a test run.

## Non Goals
- Refactoring all the e2e tests at once

# Proposal

## Definition of Personas

- E2E test authors.

## User Stories

As an e2e test author:

I want to create a new e2e test file ensuring that there will be no leftover that may influence subsequent testing.

## Repos
Kubevirt/kubevirt

## Design
The `Framework` should be similar to Kubernetes' [Framework](https://github.com/kubernetes/kubernetes/blob/master/test/e2e/framework/framework.go).
Phases of a test suite are:
1. **Environment setup**: prepares the cluster by installing namespaces, resources, etc..
2. **Test setup**: prepares the cluster with changes that are specific to a group of tests.
3. **Test execution**
4. **Test teardown**: must restore the environment to the previous state of the setup.
5. **Environment teardown**:  must restore the environment to the previous state of the environment setup.

| Phase                | Ginkgo function           | Repetition                                    |
|----------------------|:--------------------------|:----------------------------------------------|
| Environment setup    | SynchronizedBeforeSuite() | Once at the beginning of the whole test suite |
| Test setup           | BeforeEach()              | Before each test execution                    |
| Test execution       |                           |                                               |
| Test teardown        | AfterEach()               | After each test execution                     |
| Environment teardown | SynchronizedAfterSuite()  | Once at the end of the whole test suite       |

The `Framework` supports steps 2 and 4, freeing the test files from all the common steps of preparing and
cleaning the environment. Note that steps 2 and 4 will not disappear in the test files, but will contain
only the logic related to that specific set of tests.

## Phases

- PoC: Introduce the `Framework` into the e2e tests at kubevirt/kubevirt,
  migrating a small group of tests using it.
- Monitor these tests for a long enough period and check the stability.
- If the stability monitoring finishes with satisfying
  results, start migrating other tests to it.
****
