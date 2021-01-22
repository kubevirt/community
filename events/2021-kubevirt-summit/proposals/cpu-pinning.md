# Title

## Workarounds with Kubevirt 
###  CPU Pinning with custom policies

# Abstract

CPU Pinning : Kubevirt supports CPU pinning via the Kubernetes CPU Manager. However there are a few gaps with achieving CPU pinning only via CPU Manager:
It supports only static policy and doesn’t allow for custom pinning. 
It supports only Guaranteed QoS class.
The insistence by CPU Manager to keep a shared pool means that it is impossible to overcommit in a way that allows all CPUs to be bound to guest CPUs. 
It provides a best-effort allocation of CPUs belonging to a socket and physical core. In such cases it is susceptible to corner cases and might lead to fragmentation.
That is, Kubernetes keeps us from deploying VMs as densely as we can without Kubernetes. An important requirement for us is to do away with the shared pool and let kubelet and containers that do not require dedicated placement to use any CPU, just as system processes do. Moreover, system services such as the container runtime and the kubelet itself can continue to run on these exclusive CPUs. The exclusivity offered by the CPU Manager only extends to other pods. In this session we’d like to discuss the workarounds we use for supporting a custom CPU pinning using a dedicated CPU device plugin and integrating it with Kubevirt and discuss use cases. 

# Presenters

- Sowmya Seetharaman (sseetharaman@nvidia.com, github: sseetharaman6)
- Dhanya Bhat (dbhat@nvidia.com, github: dbbhat)

[X] The presenters agree to abide by the
    [Linux Foundation's Code of Conduct for Events](https://events.linuxfoundation.org/about/code-of-conduct/)

# Session details

- Track: Users
- Session type: BoF
- Duration: 20m
- Level: Any

