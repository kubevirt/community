# instancetype.kubevirt.io - pkg/instancetype refactor

## Overview

The implementation of instance types (and formally flavors) started with KubeVirt v0.46.0 and has slowly grown in scope and features. The structure of the code however hasn’t changed much since the initial implementation of flavors and as a result is mostly concentrated within a handful of files within `pkg/instancetype`. The initial design decision of providing a single public `interface` and implementation also has not scaled well as new features and functionality have been added to the codebase.

## Motivation

Ultimately the instance type code base is becoming more difficult to both maintain and extend. Refactoring the code to make it more manageable for future contributors and maintainers ahead of the API graduating to `v1` is extremely important as a result.

## Goals

* Break up `pkg/instancetype/instancetype.go` and `pkg/instancetype/instancetype_test.go`  
* Provide finer grain implementations of common instance type and preference functionality such as find, apply, store and infer  
* Replace the public `InstancetypesMethods` interface with private smaller interfaces defined closer to their consumers
* All existing functionality should be unchanged by this work, all functional tests should therefore be untouched

## Non Goals

* Any behavioral changes to the API during the release v1.5 cycle should happen outside of this work

## User Stories

* As a developer working on instance types and preferences I want to see the code logically separated for easier maintenance and future improvements
* As a developer working on instance types and preferences I want to see the API graduate to `v1`

## Repos

* kubevirt/kubevirt

## Design

### Break out preference code into a `pkg/instancetype/preference` sub package

Much of the business logic around preferences is heavily intertwined with instance types. This should be extracted into a separate sub package for easier maintenance.

### Move lookup code under `pkg/{instancetype,instancetype/preference}/find`

Resource lookup code for instance types and preferences should also be extracted into new `find` packages.

Any common code placed under `pkg/instancetype/find`.

### Move apply code under `pkg/{instancetype,instancetype/preference}/apply`

The logic around applying instance types and preferences is significantly different and should be clearly separated.

Further separation should also allow for specific preferences to be applied in specific use cases.

An example being during device hot-plug where a subset of preferences need to be applied, for example network interface or disk preferences.

### Move infer code under `pkg/{instancetype,instancetype/preference}/infer`

Any common code placed under `pkg/instancetype/infer`.

### Move store code under `pkg/{instancetype,instancetype/preference}/store`

Any common code placed under `pkg/instancetype/store`.

### Extract webhook validation code under `pkg/{instancetype,instancetype/preference}/validate`

Moving this code under `pkg/{instancetype,instancetype/preference}` will ease future maintenance.

### Replace use of `InstancetypeMethods` with specific interfaces per consumer requirements

At present the `InstancetypeMethods` interface is requested by various webhooks, subresource APIs and controllers:

* `VMsAdmitter`  
* `VMsMutator`  
* `SubresourceAPIApp`  
* `VMController`  
* `VMExportController`

There are also several other locations where public functions from `pkg/instancetype` are called directly.

The use of a single public interface exported by `pkg/instancetype` was a mistake.

This should be replaced with smaller simple private interface definitions alongside consumers of the implementations provided by `pkg/{instancetype,instancetype/preference}`.

Once replaced, `InstancetypeMethods` can be removed from `pkg/instancetype` entirely.

### API Examples

Taking a very simplistic example from  `VMController`, we can replace the use of `InstancetypeMethods` with a private interface defined to meet the requirements of the specific consumer.

For the sake of this example the only requirement for the handler is to find an instance type, thus we can define the following interface within the controller:

```go
type instancetypeHandler interface {
    Find(vm *v1.VirtualMachine) (*v1beta1.VirtualMachineInstancetypeSpec, error)
}
```

As set out above the implementation of this logic will reside under `pkg/instancetype/find` and will satisfy this interface.

```go
type Finder struct {
    store                   cache.Store
    clusterStore            cache.Store
    controllerRevisionStore cache.Store
    client                  kubevirtcli.KubevirtClient
}

func (f *Finder) Find(vm *v1.VirtualMachine) (*v1beta1.VirtualMachineInstancetypeSpec, error) {
 [..]
}
```

We can then pull in the required implementation into a handler in order to satisfy the interface within the controller:

```go
import "kubevirt.io/kubevirt/pkg/instancetype/find"

type instancetypeHandler interface {
    Find(vm *v1.VirtualMachine) (*v1beta1.VirtualMachineInstancetypeSpec, error)
}

type handler struct {
    find.Finder
}

var _ instancetypeHandler := &handler{}

finder := find.NewSpecFinder(instancetypeStore, clusterInstancetypeStore, controllerRevisionStore, virtClient),

instancetypeHandler := &handler{
    Finder: finder,
}

c := &VMController{
    [..]
    instancetypeHandler: instancetypeHandler
    [..]
}
```

Before using this handler throughout the controller:

```go
func (c *VMController) Sync(){
    [..]
    spec, err := c.instancetypeHandler.Find(vm)
    [..]
}
```

## Update/Rollback Compatibility

N/A - There should be no functional changes as a result of this refactor.

## Functional Testing Approach

N/A - There should be no functional changes as a result of this refactor.

## Implementation Phases

* Introduce `pkg/instancetype/preference`
* Make internal changes to `pkg/instancetype` and `pkg/instancetype/preference` without breaking `InstancetypeMethods` consumers
* Once complete rework consumers to use locally defined interfaces with individual PRs for easier review
