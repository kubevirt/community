# Title

Privilege dropping, one capability at a time

# Abstract

KubeVirt's architecture is composed of two main components: virt-handler, a trusted DaemonSet, running in each node, which operates as the virtualization agent, and virt-launcher, an untrusted Kubernetes pod encapsulating a single libvirt + qemu process.

To reduce the attack surface of the overall solution, the untrusted virt-launcher component should run with as little linux capabilities as possible.

The goal of this talk is to explain the journey to get there, and the steps taken to drop CAPNETADMIN, and CAPNETRAW from the untrusted component.

This talk will encompass changes in KubeVirt and Libvirt, and requires some general prior information about networking (dhcp / L2 networking).

# Presenters

- Miguel Duarte Barroso, Software Developer, Red Hat, mdbarroso@redhat.com

[x] The presenters agree to abide by the
    [Linux Foundation's Code of Conduct for Events](https://events.linuxfoundation.org/about/code-of-conduct/)

# Session details

- Track: Contributors
- Session type: Presentation
- Duration: 40m
- Level: Intermediate

# Additional notes
None
