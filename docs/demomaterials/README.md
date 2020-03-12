_Table of contents_

<!-- TOC depthFrom:2 insertAnchor:false orderedList:true updateOnSave:true withLinks:true -->

- [Purpose of this repo](#purpose-of-this-repo)
- [Demo](#demo)
  - [Demo Script](#demo-script)
    - [Basic](#basic)
    - [Advanced](#advanced)
  - [Online resources](#online-resources)
    - [Technology hands-on](#technology-hands-on)
    - [Videos](#videos)
  - [Offline resources](#offline-resources)
    - [Technology hands-on](#technology-hands-on-1)
- [Additional resources on the technology](#additional-resources-on-the-technology)
  - [Networking](#networking)
  - [VM's](#vms)
    - [Windows workloads](#windows-workloads)
  - [Console access or UI](#console-access-or-ui)
  - [Development](#development)

<!-- /TOC -->

## Purpose of this repo

This repository and its contents, aim to provide instructions and materials to be able to host a demo on KubeVirt technology.

It contains links to different materials already available on the website or videos in the KubeVirt channel that could help set up a demonstration environment.

Many links point to the documentation at `user-guide`, going exactly to the topic commented at that point.

**Note**: Videos from YouTube can be downloaded 'locally' for offline usage via `youtube-dl` or similar tools so, that's also an option for booth display.

## Demo

In this section you'll find links to resources to prepare your demonstration about KubeVirt technology.

### Demo Script

Check materials from [KubeVirt-Tutorial](https://github.com/kubevirt/kubevirt-tutorial) for preparing the contents. KubeVirt-tutorial contains a list of scenarios aimed at exploring interactively the technology, from deploying KubeVirt, then doing Virtual Machine operations, using DataVolumes, KubeVirt UI and Multus.

#### Basic

The basic demo should contain:

- KubeVirt installation
- [VM Lifecycle](https://kubevirt.io/user-guide/#/usage/life-cycle?id=life-cycle)
  - create VM
  - start VM
  - pause VM
  - resume VM
  - stop VM
  - [create container disk](https://kubevirt.io/user-guide/#/creation/disks-and-volumes?id=containerdisk)

So the demonstration should start with:

- Quick introduction to KubeVirt
- Lifecycle basis (create, start, pause, resume, stop)
- Consider demonstrating KubeVirt UI as we target new users

#### Advanced

The advanced demo should contain:

- Advanced operations
  - [Live Migration](https://kubevirt.io/user-guide/#/installation/live-migration?id=live-migration)
  - [Node Drain/Eviction](https://kubevirt.io/user-guide/#/installation/node-eviction?id=how-to-evict-all-vms-on-a-node)
- Install from ISO ([Article covering Windows installation using an ISO](https://kubevirt.io/2020/KubeVirt-installing_Microsoft_Windows_from_an_iso.html))
  - Run a Windows VM
    - Access via RDP over a dedicated network [Expose virtual machine ports](https://kubevirt.io/user-guide/#/usage/network-service-integration?id=expose-virtualmachineinstance-as-a-loadbalancer-service)

So the demonstration would start where 'basic' was left and perform:

- live migrate a vm
- perform a node drain
- install VM from ISO
- Access VM console in graphical mode

### Online resources

#### Technology hands-on

[Katacoda scenarios](http://katacoda.com/kubevirt) covers already some of the features and requires just a browser with internet connectivity from participants:

- [Installing KubeVirt](https://katacoda.com/kubevirt/scenarios/kubevirt-101)
- [Using Containerized Data Importer](https://katacoda.com/kubevirt/scenarios/kubevirt-cdi)
- [Upgrading KubeVirt](https://katacoda.com/kubevirt/scenarios/kubevirt-upgrades)

#### Videos

The [KubeVirt YouTube Channel](https://www.youtube.com/channel/UC2FH36TbZizw25pVT1P3C3g) features videos on the technology and community meetings.

- [Kube Virt 2 minutes introduction](https://www.youtube.com/watch?v=uusM5SyK-vc&feature=youtu.be)

Above 'Katacoda' labs have their video counterpart:

- [Using KubeVirt](https://www.youtube.com/watch?v=eQZPCeOs9-c)
- [Experimenting with CDI](https://www.youtube.com/watch?v=ZHqcHbCxzYM)
- [KubeVirt Upgrades](https://www.youtube.com/watch?v=OAPzOvqp0is)

Additional videos showcase how to demonstrate KubeVirt UI:

- Managing KubeVirt Using OpenShift Web Console
  - [Using compiled binary](https://www.youtube.com/watch?v=XQw4GkGHs44)
  - [Using container](https://www.youtube.com/watch?v=xoL0UFI657I)
- [KubeVirt Basic Operations](https://www.youtube.com/watch?v=KC03G60shIc)
- [KubeVirt + Neutron](https://asciinema.org/a/7nB3vgIJcz05TxRNiaD2vLLdE)
- [KubeVirt basic presentation](https://asciinema.org/a/182627)

### Offline resources

#### Technology hands-on

For offline demonstrations you can use:

- [MiniKube](https://kubevirt.io/quickstart_minikube)
- [Kind](https://kubevirt.io/quickstart_kind)

Both of them makes it easy to have attendees to deploy KubeVirt using virtual machine or containers.

## Additional resources on the technology

### Networking

- [Network Deep Dive](https://kubevirt.io/2018/KubeVirt-Network-Deep-Dive.html)
- [Network updates on top of Deep Dive](https://kubevirt.io/2018/KubeVirt-Network-Rehash.html)
- [Attach to multiple networks](https://kubevirt.io/2018/attaching-to-multiple-networks.html)

### VM's

- [Containerized Data importer](https://kubevirt.io/2018/containerized-data-importer.html) for importing VM's.
- [How to import VM into KubeVirt](https://kubevirt.io/2019/How-To-Import-VM-into-Kubevirt.html)

#### Windows workloads

- [Install Windows from an ISO](https://kubevirt.io/2020/KubeVirt-installing_Microsoft_Windows_from_an_iso.html)

### Console access or UI

- [KubeVirt UI options](https://kubevirt.io/2019/KubeVirt_UI_options.html)
- [OKD UI console installation for KubeVirt](https://kubevirt.io/2020/OKD-web-console-install.html)

### Development

- [Use VSCode for KubeVirt development](https://kubevirt.io/2018/Use-VS-Code-for-Kube-Virt-Development.html)
