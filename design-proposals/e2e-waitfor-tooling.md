Event driven E2E Testing
=

**Authors**: [Edward Haas](https://github.com/eddev)

# Summary

This proposal is suggesting the usage of events over polling the cluster
resources & objects in e2e tests.

# Motivation

Kubevirt has an impressive amount of e2e test, covering a large amount
scenarios and features.

The scenarios that the tests are exercising, frequently use asynchronous
API operations against the Kubevirt cluster. In most cases, there is a
need to *wait* for the (overall) operation to complete (be examining
the current state) and only then proceed with other steps.

Waiting for a state of an object (e.g. VMI) is currently implemented
in the e2e tests through polling repeatedly its state.
This polling strategy has one clear pro, it is simple.
On the other hand, there are several cons:
- Loads the API server (depending on the number of retries, frequency of
  retries, concurrent running tests,  mount of such "waits").
  This load is also dependent on the resources the API Server has in the
  test cluster.
- Intermediate states may be missed. The sampling of the state may just
  miss some conditions, just because it happened very fast.
  Mitigating this usually involves the need to increase the frequency of
  such sampling, which again raises the load issue.

Therefore, one can claim that using polling to wait for a condition to
be met is by definition flaky and inaccurate.

Fortunately, Kubernetes core strategy is to use events to distribute
facts about the objects/resources state.
Production application that integrate with Kubernetes are expected to
use events (e.g. watchers, informers) and not polling.

The e2e tests can be classified as applications that integrate with the
Kubernetes cluster, and therefore, could (and should) use the same
patterns expected from production applications.

## Goals

- Use event driven approach to wait for a condition. Encourage it over
  the polling approach.
- Increase e2e test stabilization.

## Non Goals
- Replace all existing polling implementation at once.

# Proposal

## Definition of Personas

- E2E test authors and reviewers.

## User Stories

As a e2e test author:

I want to execute an asynchronous operation on the cluster (e.g. create
and run a VMI) and wait for it to complete, given a condition I can
define and express (e.g. VMI phase `running`).
I am able to add a timeout and abort the "waiting" explicitly.

## Existing solution

The current e2e tests at kubevirt/kubevirt use several methods to
implement the "waiting" part, all are based on polling:
- Ginkgo [Eventually](https://onsi.github.io/ginkgo/#patterns-for-asynchronous-testing).
  Repeatedly executing the body function until the resulting assertion
  succeeds, the timeout is reached or a body assertion fails.

  Example:
  ```
  Eventually(func() error {
      return libnet.PingFromVMConsole(vmi, ipAddress)
  }, 15*time.Second, time.Second).Should(Succeed())
  ```
- Kubernetes apimachinery wait tooling.
  Repeatedly executing the body function until the resulting assertion
  succeeds, the timeout is reached or an error returned from the body.

  Example:
  ```
  err := wait.Poll(time.Second*5, time.Minute*3, func() (bool, error) {
      _, err := clusterApi.CoreV1().Secrets(namespace).Get(secret.Name, metav1.GetOptions{})
      if err != nil {
          if errors.IsNotFound(err) {
              return true, nil
          }
          return false, nil
      }
      return false, fmt.Errorf("secret %s already exists", secret.Name)
  })
  ```
- Kubevirt e2e wrapper for VMI state.
  These are repeatedly reading the VMI state until the required
  state/phase is reached. It does so using Ginkgo `Eventually`.
  The required state condition is abstracted away.

  Example:
  ```
  vmi = tests.CreateVmiOnNode(vmi, nodeName)
  vmi = tests.WaitUntilVMIReady(vmi, console.LoginToFedora)
  ```

## Event driven waiting

### Summary

Kubernetes API is providing a "Watch" operation on all available
resources. Watchers are a powerful toolset that enable [efficient
detection of changes](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)
with assurance that no event is missed.

Fortunately, the Kubernetes community has curated
[tooling](https://pkg.go.dev/k8s.io/client-go/tools/watch) to take
advantage of watchers by providing most of the mechanics required
to setup and manage them.
Its users are left with just adding basic logic regarding the
conditions and possibly filtering (e.g. label selectors).

The main function in use is `Until`, supporting a timeout, aborting
and conditions.

Example (waiting for namespace deletion):
```
_, err := k8swatchtools.Until(ctx, namespace.ResourceVersion, w, func(event k8swatch.Event) (bool, error) {
    if event.Type != k8swatch.Deleted {
        return false, nil
    }
    _, ok := event.Object.(*corev1.Namespace)
    return ok, nil
})
if err == nil {
    log.Printf("Namespace %q deleted", namespace.Name)
}
```

## Phases

- PoC: Introduce `Until` into the e2e tests at kubevirt/kubevirt,
  replacing a single usage of one of the polling implementations.
- Start migrating a group of tests to this new method and monitor
  it for a period of several weeks (for stability).
- Once (and if) the stability monitoring finishes with satisfying
  results, start migrating other tests to it. This should include
  a plan to distribute the work.

