# Properly use contexts in functional tests

## Overview
This proposal is about properly use golang contexts in KubeVirt functional tests.

## Introduction
Since the start of the KubeVirt project, contexts became a very important concept
in golang, when writing an asynchronous logic. In a way, the KubeVirt code left
behind regarding the proper usage of golang contexts.

Using `context.Background()` or even worse, `context.TODO()` is a code smell, because
doing that won't allow the application to cancel async-operation, like I/O, for example,
in case of termination.

While closing this gap in the production code is a huge effort, the situation in the
functional test code is a much easier to handle. This is why this proposal deals with
the functional tests.

This proposal is based on the ginkgo documentation, that describes how to deal with 
interrupts by letting ginkgo inject context object to the tests. See here for more
details:
[https://onsi.github.io/ginkgo/#spec-timeouts-and-interruptible-nodes](https://onsi.github.io/ginkgo/#spec-timeouts-and-interruptible-nodes)

## Goals
- Define coding standards for the functional tests, regarding the proper way of using contexts
- Modify the current functional tests code to use the new standards
- Using the new standards as a gate for code review of functional tests

## Design 
### Coding Standards for using contexts in Functional Tests
ginkgo needs to be able to cancel asynchronous operations when it should
terminate a test or a go-routine. For that, ginkgo provides its own contexts.

The test code should use the contexts provided by ginkgo, rather than new contexts
like `context.Background()` or `context.TODO()`.

The following guidelines are base on the [ginkgo documentation](https://onsi.github.io/ginkgo/#spec-timeouts-and-interruptible-nodes).

Avoid using new contexts (e.g. `context.Background()` or `context.TODO()`).
Instead, prefer using the context injected by ginkgo.

First, add the `ctx` optional parameter of the anonymous functions in `It`, `BeforeEach`,
`AfterEach` or `DescribeTable`. Then use this parameter for all places where a context is 
needed in this container.

ginkgo will inject a context to the test, by populating the `ctx` parameter with a context 
controlled by ginkgo.

For example:
```go
It("should test something", func(ctx context.Context){
	...
	kv, err = virtClient.KubeVirt(kvName).Update(ctx, kv, metav1.UpdateOptions{})
	Expect(err).ToNot(HaveOccurred())
	...
})
```

When using asynchronous assertion, i.e. `Eventually` or `Consistently`, pass the container's context
using the `WithContext` method, and add the context parameter to the checked function; e.g.
```go
It("should test something", func(ctx context.Context){
	...
	
	Eventually(func(ctx context.Context) error {
		kv, err = virtClient.KubeVirt(kvName).Update(ctx, kv, metav1.UpdateOptions{})
		...
		
		return err
	}).WithTimeout(120 * time.Seconds).
		WithPolling(10 * time.Seconds).
		WithContext(ctx).
		Should(Succeed())
	
	...
})
```

Notice that in case of also passing the `Gomega` parameter, the context parameter will be the second one, i.e.
```go
It("should test something", func (ctx context.Context){
	...
	
	Eventually(func(g Gomega, ctx context.Context) {
		kv, err = virtClient.KubeVirt(kvName).Update(ctx, kv, metav1.UpdateOptions{})
		g.Expect(err).ToNot(HaveOccurred())
		...
	}).WithTimeout(120 * time.Seconds).
		WithPolling(10 * time.Seconds).
		WithContext(ctx).
		Should(Succeed())
	
	...
})
```

Also, notice that gingko cancels contexts when exiting a container, therefore avoid using the same context across 
multiple containers. This example will fail with canceled context error:
```go
var _ = Describe("test something", func(){
	var ctx context.Context
	
	BeforeEach(func(gctx context.Context){
		ctx = gctx // or something like ctx = context.WithTimeout(gctx, ...)
	})
	
	It(func(){
		kv, err = virtClient.KubeVirt(kvName).Update(ctx, kv, metav1.UpdateOptions{}) // <= will fail here
		...
	})
})
```

#### Bad Example
```go
It("should test something", func(){
	...
	kv, err = virtClient.KubeVirt(kvName).Update(context.Backgound(), kv, metav1.UpdateOptions{})
	Expect(err).ToNot(HaveOccurred())
	...
})
```
The above example, uses a new context when performing an I/O operation - using the client to perform
a call to the kubernets cluster. This context can't be canceled and in case of interrupt, this is not 
handled properly. Using the ginkgo context, will cause a proper cancellation of the I/O operation in
this case.

## Additional Reading
* [ginkgo documentation](https://onsi.github.io/ginkgo/#spec-timeouts-and-interruptible-nodes)
* [kubernetes coding standards](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/writing-good-e2e-tests.md#interrupting-tests)

## Related PRs
* HCO implementation: [hyperconverged-cluster-operator/pull/2952](https://github.com/kubevirt/hyperconverged-cluster-operator/pull/2952)
* First KubeVirt PR to start implementing the new coding standards: [kubevirt/pull/12140](https://github.com/kubevirt/kubevirt/pull/12140)
  (Halt. Waiting for this proposal)
* KubeVirt coding standards: [kubevirt/pull/12148](https://github.com/kubevirt/kubevirt/pull/12148)
  (Reverted. Waiting for this proposal)
