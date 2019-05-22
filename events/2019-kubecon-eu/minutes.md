# Meetup Minutes
May 22 2019

## Intro
- Quite a few poeple, ~30
- The meetup is quite far away from conference venue

## Updates
- Basics landet in 0.17 built into the virt-operator
- Used by a few, not all yet

## Use-cases
- 3x Kube on Kube using KubeVirt (Triple-k?)
- 1x "One-off-VMs"

## Question: Does live migration interrupt the network connection?
- No - For additional networks as long as the underlying network provider supposrts to pick arbitrary MAC (i.e. multus)
- Yes - for pod network vNIC because it is kube like, which does not care about live migration
- Remark: endpoints can be used to point to interfacs of additional networks

## Network providers
- multus
- contrail
- flannel
- cillium
- CNI + vlan

## Security thoughts
- a user should not get more privileges if he is accessing the vm pod than he would have from the vm itself
- Idea: A proposal to have handler to be performing privileged operations, instead of DP or launcher

## Question: What about a network metadataserver?
- No objections, but no driving force so far

## Device passthrough
- SR-IOV NICs, working
- vGPU might be coming up (nVIDIA)
- interest in general - from multiple parties

## Fencing - why do we need it at all?
- To automatically resolve unknown node states on bare-metal
- depends on "cloud provider" 
- we can not distinguish between node not communicating and non-operatoinal
- to resolve the unknown state iof  amachine
- ties into machine api
- machiner api to power off the machine ( depends on cloud provider)
- in bare metal case we cre powering it offthe continues with node removal

## Bare metal management
- All running KubeVirt on bare-metal
- Nobody is using bare-metal management system

## Cluster API
- "everybody" does it - but in many different ways
- Room for convergence?

## Cloud Provider
- Request to have official releases and more testing
- sync offline - on meetings and ml

## Ceph
- used by many
- more parties started to look at it

## Hot-plug
- is hotp-plug important or is it performance?
  For the kube on kube case, yes
  overcloud wants to use storage of undercloud
  over cloud node asking under cloud storage for a new volume to attach to node to expose to overcloud node pod
- Still an issue

## Question: Other VM runtimes?
- vm runtimes?
- qemu - nemu, firecracker?
- Sharing some technical improvements
- No conceptual convergence

## CNCF?
- Need a sponsor, unsure
- Need enough supporting parties, probably
- Trademark transfer, possible (likely)
