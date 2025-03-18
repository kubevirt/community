# KubeVirt v1.6 Unconference notes (Feb 25-26)

These notes come from the hedgedoc notes that we used for the unconference and have been lightly edited for formatting.

## Split live migration VEP

**Context:** https://github.com/kubevirt/enhancements/pull/6

- Is this sig-compute or sig-storage?
- What do we call the "source" and "target"?
-  Actor? Modality? Endpoint? Role?
- This VEP covers the mechanics of the migration itself. It assumes that a higher level orchestrator will set up the resources in advance.

### Follow-up to split live migration VEP

- How much of the final cross-cluster design make its way into what we do today. ie how many code paths will we need to maintain
- Should be the same path. The vmi will go over a different channel
 

## SIG-release Retrospective of release 1.5 (also overlapped with VEP discussion)

- Add a guide for VEP process
-  Add notes for expectations of VEP approvers
-  One of the problems is understanding the new process
        - A step by step doc should be provided
        - A person who will guide authors would be helpful
-  What is the process bringing
-  Checkpoints, more specifically it was mentioned that it would be valuable to have "gut check" if the feature is on time
-  More experience/KEPs is needed to iron this out
-  Who should approve PRs of a VEP (should it be approver of VEP)? A: that is not the intention and SIGs should be responsible for the code.
-  State milestones - here we need discussion on what the milesotnes will mean and what deadlines we set. Think about exception process.
-  Rules - It was specificly called out to be more explicit about what is allowed and what not
-  Communication between sigs and with release - this was identified as something which needs to be improved
-  Exceptions - Document the process
-  What happens if deadline is missed ? Who to contact?
-  Target release, what it means, how do we make this clear? 
-  Labels? Can they help us?

## Virtualization Enhancement Processes (VEPs)

**Context:** https://github.com/kubevirt/enhancements

Notes: 

- Targeting SIGs in the VEP
- SIG representatives then involved to sponsor the VEP
- When VEP merged, issue is created (in enhancements repo)
- Issue given milestone; tracked on ~project board

Timeline:

- KEPs: merged at least 1 month before code freeze to be eligible
- For us, this would be ~beta tag

Concerns:

- Lack of people power (comparative to k8s)
  - Intention is to move the majority of the review burden to the SIGs
  - VEP review/triage should be a part of regular SIG meetings
- k8s decentralised; our approach currently centralised with only 2 approvers
  - SIGs should be involved in owning the code
  - Most (all?) SIGs will still need to review the VEPs to ensure they agree with the approach, and that it does not mess with them
- Different SIGs likely have different capacity
- VEP deadline for a release needs to be clearly communicated in the lead-up

Other repos: How does this apply to them?

- Let's create a list of repos that are under the umbrella of k/enhancements

Related:

- Better documentation on purpose of alpha and beta, and definition of feature freeze

**Resulting issue/pr:** https://github.com/kubevirt/enhancements/pull/13

## What is the state of Bazel in kubevirt/kubevirt?

- Reasons for Bazel:
-  Reproducible builds
-  Decreased build time
-  Bazel cache is used in CI
-  RPM dependency mirroring 

- Issues with Bazel:
-  KV/KV is using an old version of Bazel
-  Lack of knowledge around Bazel
-  Tech debt: 
        - deprecated rules_docker module blocking upgrades of bazel version
            - https://github.com/kubevirt/kubevirt/pull/13111
-  s390x is not supported by upstream (crossbuilding on x86 for s390x or building s390x bazel manually)
-  Stay close to Kubernetes, they dropped Bazel as well (investigate what K8S is using for CI caching)

- Potential issues when removing Bazel:
-  How do we handle backports to branches that still use Bazel?
-  Need to handle cross-compile or use a different method
-  rpm caching could potentially be handled standalone by bazeldnf?
-  BUILD.bazel files are mostly generated, targets and containerfiles are not
-  Build caching could be replaced by backing up / restoring the GO cache

- Security updates could be possible with Prow / Bazel

**Resulting issue/pr:** https://github.com/kubevirt/kubevirt/issues/14038

## DRA Support in KubeVirt

**Context:** https://github.com/kubevirt/community/pull/331

-  [scheduling] preferably the last session of the day
-  Discussion until now: https://docs.google.com/document/d/1bQdLoxwSC1ILvyIb4ljSm5RZmsVpvqfWKIbN0KOBKsw/edit?tab=t.0#heading=h.fqe36hzerfrn
-  What sig does this belong to?

## Libvirt's XML library

- Why can't we replace virt-launcher's custom implemetation with libvirt
- libvirt has a go library that converts structs to xml and back.
- We have a custom implementation, that implements a subset of what they have - which is a maintenance burden for us.
- Our custom structs are used in one of the communication channels between virt-launcher and virt-handler.
- Every time we add stuff our custom implementation - this communication channel is automatically affected and more info is transmitted.

- objections to using libvirt's library:
  - Claiming it will take more memory.
  - It will open us to feature requests that we don't want to do - exposing more libvirt API on the VM objects.


## SIG-compute 1.6 feature prioritization

- Change the default rolloutStrategy to LiveUpdate
  - Add periodic lane to test Staging
- Live migration bug(s) [+1 @ibezukh]
  - Post-copy hotplug
-  Removing (deprecate/GA) some feature gates
  - List TBD
-  SWAP?
  - David's Node Pressure Eviction VEP (https://github.com/kubevirt/community/pull/390)
-  DRA?
  - See dedicated session
-  Persistent Firmware UUID
  - https://github.com/kubevirt/community/pull/347


## SIG-network 1.6 feature prioritization:

- Enhanced Passt  binding with seamless migration (preserving TCP connectivity).
- Integrate network vnic hotplug into the rollout strategy feature.
- Enhanced the network binding plugin framework to support a shared folder with the virt-launcher compute container.


## SIG-storage 1.6 feature prioritization

- Split live migration VEP https://github.com/kubevirt/enhancements/pull/6
    - Object graph involved in this too https://github.com/kubevirt/community/pull/385
- Storage class migration (hot and cold)
- Incremental Backup VEP
    - Issue #1: Data-file over an existing raw image destroys current raw data
        - Shelly has a workaround using qemu-img amend but looking for a command for this
    - Issue #2: Need to restart the VM when configuring the qcow2 overlay
    - Issue #3: Checkpoint metadata needs to be recreated each time the VM is started (ephemeral libvirt installation)
    - Issue #4: How to back up a VM that is shut down?
        - Start the VM in a paused state and perform an "online" backup
        - This option reduces the amount of code to write because it's the same flow as an online backup
- [cfillekes] CDI test suite feature support
    - Accepts CSI driver capabilities as input
        - https://github.com/kubevirt/containerized-data-importer/blob/main/hack/build/run-functional-tests.sh#L69-L71
