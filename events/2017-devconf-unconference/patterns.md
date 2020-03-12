# Design patterns

Owner: fabiand

## Declarative

- You declare what you want: A new VM with a new disk.
- You do _not_ tell how to get there: Create a VM then add a disk.

The benefit of this approach is that the knowledge which is required to fulfill your wish, will be moved into a software component. I.e. you will create a component that knows what needs to be done to create a new VM with a new disk.
This allows the component to do the work for you.

The declarations are stored inside Kubernetes.

The entities/objects/resources can be modified using CRUD operations.

## Reactive

This is the pattern which is waiting for a change in your declaration and reacting to it.
You do not need to reach out to a component to perform instruct it.

## 3rd Party Resources

[//]: # "Reference Links"
[k8sprinciples]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/principles.md
[k8sarch]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/architecture.md
[k8swatch]: https://github.com/kubernetes/community/blob/master/contributors/design-proposals/apiserver-watch.md
