# Overview
Linux and Windows Server VMs can log boot and application data to their serial console.
KubeVirt user can great benefit from having VMs logs, collected over the serial console, integrated into the Kubernetes logging architecture. 

## Motivation
VM logs streamed over the serial console contain really useful debug/troubleshooting info.
While we are currently already capturing KubeVirt components logs (virt-controller, virt-launcher, libvirt, virt-handler, etc…) but the VM guest OS and application logs are not currently captured or indexed using cluster logging:
currently KubeVirt lacks the level of integration that allows VM guests logs to be seamlessly streamed and captured as for generic container workloads ones.

A similar feature al been somehow already required more than once at community level, (see https://groups.google.com/g/kubevirt-dev/c/qBdboekQ_Vk for instance).
And a similar feature is a pretty common pattern across all the most popular HyperScalers, see as examples:
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instance-console.html#instance-console-console-output
- https://learn.microsoft.com/en-us/azure/virtual-machines/boot-diagnostics
- https://cloud.google.com/compute/docs/troubleshooting/viewing-serial-port-output#enable-stackdriver
- ...

## Goals
- It should let the user collect serial console logs for his VM according to the standard k8s logging architecture: the main benefit of this choice is that it will allow it to be directly consumed by all the k8s logging aggregation solutions without the need for further customizations.
- No guest OS sided changes should be required.
- It should not prevent the usage of the serial console to interactively login into the VM.

## Non Goals
- VM serial console logs could be streamed to external logging stacks like Fluentd+Elasticsearch+Kibana/Loki/... but this specific part of the integration is out of scope for this proposal.
- Defining custom/new format/structured for guest OS logs is out of scope. 

## Definition of Users
- VM owner: the user who owns a VM in his namespace on a Kubernetes cluster with KubeVirt
- cluster-admin: the administrator of the cluster

## User Stories
- As a VM owner I want to view the VM guest OS logs using standard client tooling just as I would an application running in a pod.
- As a VM user I want to query my logs pertaining to an application running within my VM workload using Kubernetes standard logging
- As a cluster-admin, I would like to view all the logs of all the VMs running in my cluster in order to find and correct reoccurring bugs, security flaws or misconfigurations.

## Repos
- https://github.com/kubevirt/kubevirt

# Design
In [Libvirt the serial port devices](https://libvirt.org/formatdomain.html#console) can already be configured via a log sub-element, with a file attribute.
```
<log file="/var/run/kubevirt-private/b802512e-d13e-4592-83b4-db153293b666/virt-serial0-log" append="on"/>
```
that translates to something like:
```
-chardev socket,id=charserial0,fd=17,server=on,wait=off,logfile=/dev/fdset/0,logappend=on -device {"driver":"isa-serial","chardev":"charserial0","id":"serial0","index":0}
```
on the QEMU side.

This means that we can add the <log/> element like:
```yaml
    <serial type='unix'>
      <source mode='bind' path='/var/run/kubevirt-private/b802512e-d13e-4592-83b4-db153293b666/virt-serial0'/>
--->  <log file='/var/run/kubevirt-private/b802512e-d13e-4592-83b4-db153293b666/virt-serial0-log' append='on'/>
      <target type='isa-serial' port='0'>
        <model name='isa-serial'/>
      </target>
      <alias name='serial0'/>
    </serial>
    <console type='unix'>
      <source mode='bind' path='/var/run/kubevirt-private/b802512e-d13e-4592-83b4-db153293b666/virt-serial0'/>
--->  <log file='/var/run/kubevirt-private/b802512e-d13e-4592-83b4-db153293b666/virt-serial0-log' append='on'/>
      <target type='serial' port='0'/>
      <alias name='serial0'/>
    </console>
```
to the serial and the console entries in the libvirt XML for the KubeVirt VMs.

Those log files are going to be rotated by `virtlogd` exactly as for libvirt logs so file size is not an issue as it's not for libvirt/qemu ones.

Now that we have the guest console logs in a file on the kubevirt-private volume in the pod, we just need another container in the pod to stream them out as for generic container logs.
This is exactly the standard and well adopted [Kubernertes Streaming sidecar container logging architecture](https://kubernetes.io/docs/concepts/cluster-administration/logging/#streaming-sidecar-container).

So in the pod we should have another (optional) container named `guest-console-log` and defined as:
```yaml
  - args:
       - --logfile
       - /var/run/kubevirt-private/bf9afcc1-fe55-4f63-919a-807b59efff76/virt-serial0-log
    command:
       - /usr/bin/virt-tail
    env:
       - name: VIRT_LAUNCHER_LOG_VERBOSITY
         value: "2"
    image: registry:5000/kubevirt/virt-launcher@sha256:fb1c5e6656501a6af8445236106eaa9d0dc242f1f5c1783db4a8a7895dfc75e5
    imagePullPolicy: IfNotPresent
    name: guest-console-log
    resources:
       limits:
          cpu: 15m
          memory: 60M
       requests:
          cpu: 5m
          memory: 35M
    securityContext:
       allowPrivilegeEscalation: false
       capabilities:
          drop:
             - ALL
       runAsNonRoot: true
       runAsUser: 107
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
       - mountPath: /var/run/kubevirt-private
         name: private
         readOnly: true
```

Ideally the `tail` binary could potentially be enough to stream a log file to the `stdout` in the log streaming container
but the relationship with other container in the `virt-launcher` pod, especially in the shutdown/cleanup phase is critical.

If not all the containers of the virt-launcher pod are correctly going to terminate together, we could potentially risk partially shut-down virt-launcher pods, blocking migrations and reaching a clear Terminated state.

In order to safely and quickly handle many different termination cases, we should implement a custom tail equivalent golang based binary that is able to terminate itself on:
- sigterm received from the container runtime
- the unix domain socket file used for the serial console port disappearing signalling that qemu process terminated
- a termination file generated by virt-launcher-monitor if previous two mechanisms failed

For security reason the virt-launcher `private` volume that contains the log can be mounted in read only mode.

Serial console logs for the VM testvm in namespace simone can then be fetched executing something like:
```bash
stirabos@t14s:~$ kubectl logs -n simone virt-launcher-testvm-t7h26 -c guest-console-log
[    0.000000] Initializing cgroup subsys cpuset
[    0.000000] Initializing cgroup subsys cpu
[    0.000000] Initializing cgroup subsys cpuacct
[    0.000000] Linux version 4.4.0-28-generic (buildd@lcy01-13) (gcc version 5.3.1 20160413 (Ubuntu 5.3.1-14ubuntu2.1) ) #47-Ubuntu SMP Fri Jun 24 10:09:13 UTC 2016 (Ubuntu 4.4.0-28.47-generic 4.4.13)
[    0.000000] Command line: LABEL=cirros-rootfs ro console=tty1 console=ttyS0
[    0.000000] KERNEL supported cpus:
[    0.000000]   Intel GenuineIntel
[    0.000000]   AMD AuthenticAMD
[    0.000000]   Centaur CentaurHauls
[    0.000000] x86/fpu: xstate_offset[2]:  576, xstate_sizes[2]:  256
[    0.000000] x86/fpu: Supporting XSAVE feature 0x01: 'x87 floating point registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x02: 'SSE registers'
[    0.000000] x86/fpu: Supporting XSAVE feature 0x04: 'AVX registers'
[    0.000000] x86/fpu: Enabled xstate features 0x7, context size is 832 bytes, using 'standard' format.
[    0.000000] x86/fpu: Using 'eager' FPU context switches.
[    0.000000] e820: BIOS-provided physical RAM map:
[    0.000000] BIOS-e820: [mem 0x0000000000000000-0x000000000009fbff] usable
[    0.000000] BIOS-e820: [mem 0x000000000009fc00-0x000000000009ffff] reserved
[    0.000000] BIOS-e820: [mem 0x00000000000f0000-0x00000000000fffff] reserved
[    0.000000] BIOS-e820: [mem 0x0000000000100000-0x0000000003ddcfff] usable
[    0.000000] BIOS-e820: [mem 0x0000000003ddd000-0x0000000003dfffff] reserved
[    0.000000] BIOS-e820: [mem 0x00000000b0000000-0x00000000bfffffff] reserved
[    0.000000] BIOS-e820: [mem 0x00000000fed1c000-0x00000000fed1ffff] reserved
[    0.000000] BIOS-e820: [mem 0x00000000feffc000-0x00000000feffffff] reserved
[    0.000000] BIOS-e820: [mem 0x00000000fffc0000-0x00000000ffffffff] reserved
[    0.000000] NX (Execute Disable) protection: active
[    0.000000] SMBIOS 2.8 present.
[    0.000000] Hypervisor detected: KVM
[    0.000000] e820: last_pfn = 0x3ddd max_arch_pfn = 0x400000000
[    0.000000] x86/PAT: Configuration [0-7]: WB  WC  UC- UC  WB  WC  UC- WT  
[    0.000000] found SMP MP-table at [mem 0x000f5bb0-0x000f5bbf] mapped at [ffff8800000f5bb0]
[    0.000000] Scanning 1 areas for low memory corruption
[    0.000000] Using GB pages for direct mapping
[    0.000000] RAMDISK: [mem 0x03917000-0x03dccfff]
[    0.000000] ACPI: Early table checksum verification disabled
[    0.000000] ACPI: RSDP 0x00000000000F5980 000014 (v00 BOCHS )
[    0.000000] ACPI: RSDT 0x0000000003DE27FB 000034 (v01 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] ACPI: FACP 0x0000000003DE262B 0000F4 (v03 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] ACPI: DSDT 0x0000000003DE0040 0025EB (v01 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] ACPI: FACS 0x0000000003DE0000 000040
[    0.000000] ACPI: APIC 0x0000000003DE271F 000078 (v01 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] ACPI: MCFG 0x0000000003DE2797 00003C (v01 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] ACPI: WAET 0x0000000003DE27D3 000028 (v01 BOCHS  BXPC     00000001 BXPC 00000001)
[    0.000000] No NUMA configuration found
[    0.000000] Faking a node at [mem 0x0000000000000000-0x0000000003ddcfff]
[    0.000000] NODE_DATA(0) allocated [mem 0x03dd8000-0x03ddcfff]
[    0.000000] kvm-clock: Using msrs 4b564d01 and 4b564d00
[    0.000000] kvm-clock: cpu 0, msr 0:3dd4001, primary cpu clock
[    0.000000] kvm-clock: using sched offset of 1139922126 cycles
[    0.000000] clocksource: kvm-clock: mask: 0xffffffffffffffff max_cycles: 0x1cd42e4dffb, max_idle_ns: 881590591483 ns
[    0.000000] Zone ranges:
[    0.000000]   DMA      [mem 0x0000000000001000-0x0000000000ffffff]
[    0.000000]   DMA32    [mem 0x0000000001000000-0x0000000003ddcfff]
[    0.000000]   Normal   empty
[    0.000000]   Device   empty
[    0.000000] Movable zone start for each node
[    0.000000] Early memory node ranges
[    0.000000]   node   0: [mem 0x0000000000001000-0x000000000009efff]
[    0.000000]   node   0: [mem 0x0000000000100000-0x0000000003ddcfff]
[    0.000000] Initmem setup node 0 [mem 0x0000000000001000-0x0000000003ddcfff]
[    0.000000] ACPI: PM-Timer IO Port: 0x608
[    0.000000] ACPI: LAPIC_NMI (acpi_id[0xff] dfl dfl lint[0x1])
[    0.000000] IOAPIC[0]: apic_id 0, version 17, address 0xfec00000, GSI 0-23
[    0.000000] ACPI: INT_SRC_OVR (bus 0 bus_irq 0 global_irq 2 dfl dfl)
[    0.000000] ACPI: INT_SRC_OVR (bus 0 bus_irq 5 global_irq 5 high level)
[    0.000000] ACPI: INT_SRC_OVR (bus 0 bus_irq 9 global_irq 9 high level)
[    0.000000] ACPI: INT_SRC_OVR (bus 0 bus_irq 10 global_irq 10 high level)
[    0.000000] ACPI: INT_SRC_OVR (bus 0 bus_irq 11 global_irq 11 high level)
[    0.000000] Using ACPI (MADT) for SMP configuration information
[    0.000000] smpboot: Allowing 1 CPUs, 0 hotplug CPUs
[    0.000000] PM: Registered nosave memory: [mem 0x00000000-0x00000fff]
[    0.000000] PM: Registered nosave memory: [mem 0x0009f000-0x0009ffff]
[    0.000000] PM: Registered nosave memory: [mem 0x000a0000-0x000effff]
[    0.000000] PM: Registered nosave memory: [mem 0x000f0000-0x000fffff]
[    0.000000] e820: [mem 0x03e00000-0xafffffff] available for PCI devices
[    0.000000] Booting paravirtualized kernel on KVM
[    0.000000] clocksource: refined-jiffies: mask: 0xffffffff max_cycles: 0xffffffff, max_idle_ns: 7645519600211568 ns
[    0.000000] setup_percpu: NR_CPUS:256 nr_cpumask_bits:256 nr_cpu_ids:1 nr_node_ids:1
[    0.000000] PERCPU: Embedded 33 pages/cpu @ffff880003600000 s98008 r8192 d28968 u2097152
[    0.000000] KVM setup async PF for cpu 0
[    0.000000] kvm-stealtime: cpu 0, msr 360d940
[    0.000000] PV qspinlock hash table entries: 256 (order: 0, 4096 bytes)
[    0.000000] Built 1 zonelists in Node order, mobility grouping on.  Total pages: 15470
[    0.000000] Policy zone: DMA32
[    0.000000] Kernel command line: LABEL=cirros-rootfs ro console=tty1 console=ttyS0
[    0.000000] PID hash table entries: 256 (order: -1, 2048 bytes)
[    0.000000] Memory: 37156K/62956K available (8368K kernel code, 1280K rwdata, 3928K rodata, 1480K init, 1292K bss, 25800K reserved, 0K cma-reserved)
[    0.000000] SLUB: HWalign=64, Order=0-3, MinObjects=0, CPUs=1, Nodes=1
[    0.000000] Hierarchical RCU implementation.
[    0.000000] 	Build-time adjustment of leaf fanout to 64.
[    0.000000] 	RCU restricting CPUs from NR_CPUS=256 to nr_cpu_ids=1.
[    0.000000] RCU: Adjusting geometry for rcu_fanout_leaf=64, nr_cpu_ids=1
[    0.000000] NR_IRQS:16640 nr_irqs:256 16
[    0.000000] Console: colour VGA+ 80x25
[    0.000000] console [tty1] enabled
[    0.000000] console [ttyS0] enabled
[    0.000000] tsc: Detected 2303.998 MHz processor
[    0.866317] Calibrating delay loop (skipped) preset value.. 4607.99 BogoMIPS (lpj=9215992)
[    0.876491] pid_max: default: 32768 minimum: 301
[    0.882034] ACPI: Core revision 20150930
[    0.887694] ACPI: 1 ACPI AML tables successfully acquired and loaded
[    0.896350] Security Framework initialized
[    0.901560] Yama: becoming mindful.
[    0.906310] AppArmor: AppArmor initialized
[    0.911346] Dentry cache hash table entries: 8192 (order: 4, 65536 bytes)
[    0.918808] Inode-cache hash table entries: 4096 (order: 3, 32768 bytes)
[    0.926002] Mount-cache hash table entries: 512 (order: 0, 4096 bytes)
[    0.933140] Mountpoint-cache hash table entries: 512 (order: 0, 4096 bytes)
[    0.940800] Initializing cgroup subsys io
[    0.945721] Initializing cgroup subsys memory
[    0.950848] Initializing cgroup subsys devices
[    0.956461] Initializing cgroup subsys freezer
[    0.962096] Initializing cgroup subsys net_cls
[    0.967493] Initializing cgroup subsys perf_event
[    0.973019] Initializing cgroup subsys net_prio
[    0.978636] Initializing cgroup subsys hugetlb
[    0.983996] Initializing cgroup subsys pids
[    0.990061] CPU: Physical Processor ID: 0
[    1.009245] mce: CPU supports 10 MCE banks
[    1.014926] Last level iTLB entries: 4KB 0, 2MB 0, 4MB 0
[    1.021091] Last level dTLB entries: 4KB 0, 2MB 0, 4MB 0, 1GB 0
[    1.136116] Freeing SMP alternatives memory: 28K (ffffffff820b4000 - ffffffff820bb000)
[    1.218892] ftrace: allocating 31920 entries in 125 pages
[    1.541690] smpboot: Max logical packages: 1
[    1.547091] smpboot: APIC(0) Converting physical 0 to logical package 0
[    1.555728] x2apic enabled
[    1.561100] Switched APIC routing to physical x2apic.
[    1.572998] ..TIMER: vector=0x30 apic1=0 pin1=2 apic2=-1 pin2=-1
[    1.580515] smpboot: CPU0: Intel Core Processor (Skylake, IBRS) (family: 0x6, model: 0x5e, stepping: 0x3)
[    1.592841] Performance Events: unsupported p6 CPU model 94 no PMU driver, software events only.
[    1.605077] KVM setup paravirtual spinlock
[    1.610750] x86: Booted up 1 node, 1 CPUs
[    1.615772] smpboot: Total of 1 processors activated (4607.99 BogoMIPS)
[    1.624323] devtmpfs: initialized
[    1.630309] evm: security.selinux
[    1.634897] evm: security.SMACK64
[    1.639363] evm: security.SMACK64EXEC
[    1.644084] evm: security.SMACK64TRANSMUTE
[    1.649162] evm: security.SMACK64MMAP
[    1.653682] evm: security.ima
[    1.657720] evm: security.capability
[    1.662389] clocksource: jiffies: mask: 0xffffffff max_cycles: 0xffffffff, max_idle_ns: 7645041785100000 ns
[    1.673189] pinctrl core: initialized pinctrl subsystem
[    1.679807] RTC time:  2:03:09, date: 07/14/23
[    1.685372] NET: Registered protocol family 16
[    1.690641] cpuidle: using governor ladder
[    1.695854] cpuidle: using governor menu
[    1.700686] PCCT header not found.
[    1.705283] ACPI: bus type PCI registered
[    1.710208] acpiphp: ACPI Hot Plug PCI Controller Driver version: 0.5
[    1.717504] PCI: MMCONFIG for domain 0000 [bus 00-ff] at [mem 0xb0000000-0xbfffffff] (base 0xb0000000)
[    1.728065] PCI: MMCONFIG at [mem 0xb0000000-0xbfffffff] reserved in E820
[    1.735464] PCI: Using configuration type 1 for base access
[    1.742766] ACPI: Added _OSI(Module Device)
[    1.747874] ACPI: Added _OSI(Processor Device)
[    1.753097] ACPI: Added _OSI(3.0 _SCP Extensions)
[    1.758537] ACPI: Added _OSI(Processor Aggregator Device)
[    1.768054] ACPI: Interpreter enabled
[    1.772875] ACPI Exception: AE_NOT_FOUND, While evaluating Sleep State [\_S1_] (20150930/hwxface-580)
[    1.784466] ACPI Exception: AE_NOT_FOUND, While evaluating Sleep State [\_S2_] (20150930/hwxface-580)
[    1.796107] ACPI Exception: AE_NOT_FOUND, While evaluating Sleep State [\_S3_] (20150930/hwxface-580)
[    1.807585] ACPI Exception: AE_NOT_FOUND, While evaluating Sleep State [\_S4_] (20150930/hwxface-580)
[    1.819384] ACPI: (supports S0 S5)
[    1.823837] ACPI: Using IOAPIC for interrupt routing
[    1.829570] PCI: Using host bridge windows from ACPI; if necessary, use "pci=nocrs" and report a bug
[    1.842230] ACPI: PCI Root Bridge [PCI0] (domain 0000 [bus 00-ff])
[    1.848802] acpi PNP0A08:00: _OSC: OS supports [ExtendedConfig ASPM ClockPM Segments MSI]
[    1.858205] acpi PNP0A08:00: _OSC: platform does not support [PCIeHotplug]
[    1.865608] acpi PNP0A08:00: _OSC: OS now controls [PME AER PCIeCapability]
[    1.874261] PCI host bridge to bus 0000:00
[    1.879198] pci_bus 0000:00: root bus resource [io  0x0000-0x0cf7 window]
[    1.886308] pci_bus 0000:00: root bus resource [io  0x0d00-0xffff window]
[    1.893383] pci_bus 0000:00: root bus resource [mem 0x000a0000-0x000bffff window]
[    1.984679] pci_bus 0000:00: root bus resource [mem 0x03e00000-0xafffffff window]
[    1.993410] pci_bus 0000:00: root bus resource [mem 0xc0000000-0xfebfffff window]
[    2.002190] pci_bus 0000:00: root bus resource [mem 0x100000000-0x8ffffffff window]
[    2.010900] pci_bus 0000:00: root bus resource [bus 00-ff]
[    2.123971] pci 0000:00:1f.0: quirk: [io  0x0600-0x067f] claimed by ICH6 ACPI/GPIO/TCO
[    2.157909] acpiphp: Slot [0] registered
[    2.178190] pci 0000:00:02.0: PCI bridge to [bus 01]
[    2.187000] acpiphp: Slot [0-2] registered
[    2.192126] pci 0000:00:02.1: PCI bridge to [bus 02]
[    2.200122] acpiphp: Slot [0-3] registered
[    2.205237] pci 0000:00:02.2: PCI bridge to [bus 03]
[    2.212967] acpiphp: Slot [0-4] registered
[    2.218411] pci 0000:00:02.3: PCI bridge to [bus 04]
[    2.226321] acpiphp: Slot [0-5] registered
[    2.247578] pci 0000:00:02.4: PCI bridge to [bus 05]
[    2.255907] acpiphp: Slot [0-6] registered
[    2.275486] pci 0000:00:02.5: PCI bridge to [bus 06]
[    2.284584] acpiphp: Slot [0-7] registered
[    2.304283] pci 0000:00:02.6: PCI bridge to [bus 07]
[    2.312953] acpiphp: Slot [0-8] registered
[    2.333585] pci 0000:00:02.7: PCI bridge to [bus 08]
[    2.342440] acpiphp: Slot [0-9] registered
[    2.359259] pci 0000:00:03.0: PCI bridge to [bus 09]
[    2.367505] acpiphp: Slot [0-10] registered
[    2.372777] pci 0000:00:03.1: PCI bridge to [bus 0a]
[    2.388996] ACPI: PCI Interrupt Link [LNKA] (IRQs 5 *10 11)
[    2.399888] ACPI: PCI Interrupt Link [LNKB] (IRQs 5 *10 11)
[    2.409729] ACPI: PCI Interrupt Link [LNKC] (IRQs 5 10 *11)
[    2.419759] ACPI: PCI Interrupt Link [LNKD] (IRQs 5 10 *11)
[    2.431476] ACPI: PCI Interrupt Link [LNKE] (IRQs 5 *10 11)
[    2.441306] ACPI: PCI Interrupt Link [LNKF] (IRQs 5 *10 11)
[    2.451301] ACPI: PCI Interrupt Link [LNKG] (IRQs 5 10 *11)
[    2.463055] ACPI: PCI Interrupt Link [LNKH] (IRQs 5 10 *11)
[    2.472736] ACPI: PCI Interrupt Link [GSIA] (IRQs *16)
[    2.480583] ACPI: PCI Interrupt Link [GSIB] (IRQs *17)
[    2.488335] ACPI: PCI Interrupt Link [GSIC] (IRQs *18)
[    2.496817] ACPI: PCI Interrupt Link [GSID] (IRQs *19)
[    2.504678] ACPI: PCI Interrupt Link [GSIE] (IRQs *20)
[    2.512396] ACPI: PCI Interrupt Link [GSIF] (IRQs *21)
[    2.520210] ACPI: PCI Interrupt Link [GSIG] (IRQs *22)
[    2.528609] ACPI: PCI Interrupt Link [GSIH] (IRQs *23)
[    2.536823] ACPI: Enabled 2 GPEs in block 00 to 3F
[    2.544198] vgaarb: setting as boot device: PCI:0000:00:01.0
[    2.550524] vgaarb: device added: PCI:0000:00:01.0,decodes=io+mem,owns=io+mem,locks=none
[    2.559777] vgaarb: loaded
[    2.564145] vgaarb: bridge control possible 0000:00:01.0
[    2.570328] SCSI subsystem initialized
[    2.575326] ACPI: bus type USB registered
[    2.580459] usbcore: registered new interface driver usbfs
[    2.586538] usbcore: registered new interface driver hub
[    2.592476] usbcore: registered new device driver usb
[    2.598403] PCI: Using ACPI for IRQ routing
[    3.022007] NetLabel: Initializing
[    3.026794] NetLabel:  domain hash size = 128
[    3.033144] NetLabel:  protocols = UNLABELED CIPSOv4
[    3.039957] NetLabel:  unlabeled traffic allowed by default
[    3.046538] clocksource: Switched to clocksource kvm-clock
[    3.056906] AppArmor: AppArmor Filesystem Enabled
[    3.062630] pnp: PnP ACPI init
[    3.066945] system 00:04: [mem 0xb0000000-0xbfffffff window] has been reserved
[    3.076302] pnp: PnP ACPI: found 5 devices
[    3.087865] clocksource: acpi_pm: mask: 0xffffff max_cycles: 0xffffff, max_idle_ns: 2085701024 ns
[    3.109128] pci 0000:00:02.0: BAR 13: assigned [io  0x1000-0x1fff]
[    3.116667] pci 0000:00:02.1: BAR 13: assigned [io  0x2000-0x2fff]
[    3.123264] pci 0000:00:02.2: BAR 13: assigned [io  0x3000-0x3fff]
[    3.129718] pci 0000:00:02.3: BAR 13: assigned [io  0x4000-0x4fff]
[    3.136768] pci 0000:00:02.4: BAR 13: assigned [io  0x5000-0x5fff]
[    3.143702] pci 0000:00:02.5: BAR 13: assigned [io  0x6000-0x6fff]
[    3.150491] pci 0000:00:02.6: BAR 13: assigned [io  0x7000-0x7fff]
[    3.158061] pci 0000:00:02.7: BAR 13: assigned [io  0x8000-0x8fff]
[    3.165474] pci 0000:00:03.0: BAR 13: assigned [io  0x9000-0x9fff]
[    3.173432] pci 0000:00:03.1: BAR 13: assigned [io  0xa000-0xafff]
[    3.181716] pci 0000:00:02.0: PCI bridge to [bus 01]
[    3.188501] pci 0000:00:02.0:   bridge window [io  0x1000-0x1fff]
[    3.196527] pci 0000:00:02.0:   bridge window [mem 0xfe800000-0xfe9fffff]
[    3.205910] pci 0000:00:02.0:   bridge window [mem 0xfd200000-0xfd3fffff 64bit pref]
[    3.216325] pci 0000:00:02.1: PCI bridge to [bus 02]
[    3.222344] pci 0000:00:02.1:   bridge window [io  0x2000-0x2fff]
[    3.229867] pci 0000:00:02.1:   bridge window [mem 0xfe600000-0xfe7fffff]
[    3.238015] pci 0000:00:02.1:   bridge window [mem 0xfd000000-0xfd1fffff 64bit pref]
[    3.248609] pci 0000:00:02.2: PCI bridge to [bus 03]
[    3.254469] pci 0000:00:02.2:   bridge window [io  0x3000-0x3fff]
[    3.261992] pci 0000:00:02.2:   bridge window [mem 0xfe400000-0xfe5fffff]
[    3.270329] pci 0000:00:02.2:   bridge window [mem 0xfce00000-0xfcffffff 64bit pref]
[    3.280357] pci 0000:00:02.3: PCI bridge to [bus 04]
[    3.286273] pci 0000:00:02.3:   bridge window [io  0x4000-0x4fff]
[    3.293954] pci 0000:00:02.3:   bridge window [mem 0xfe200000-0xfe3fffff]
[    3.302466] pci 0000:00:02.3:   bridge window [mem 0xfcc00000-0xfcdfffff 64bit pref]
[    3.312453] pci 0000:00:02.4: PCI bridge to [bus 05]
[    3.318218] pci 0000:00:02.4:   bridge window [io  0x5000-0x5fff]
[    3.325940] pci 0000:00:02.4:   bridge window [mem 0xfe000000-0xfe1fffff]
[    3.334297] pci 0000:00:02.4:   bridge window [mem 0xfca00000-0xfcbfffff 64bit pref]
[    3.344253] pci 0000:00:02.5: PCI bridge to [bus 06]
[    3.350103] pci 0000:00:02.5:   bridge window [io  0x6000-0x6fff]
[    3.357541] pci 0000:00:02.5:   bridge window [mem 0xfde00000-0xfdffffff]
[    3.365797] pci 0000:00:02.5:   bridge window [mem 0xfc800000-0xfc9fffff 64bit pref]
[    3.376107] pci 0000:00:02.6: PCI bridge to [bus 07]
[    3.382162] pci 0000:00:02.6:   bridge window [io  0x7000-0x7fff]
[    3.389623] pci 0000:00:02.6:   bridge window [mem 0xfdc00000-0xfddfffff]
[    3.397475] pci 0000:00:02.6:   bridge window [mem 0xfc600000-0xfc7fffff 64bit pref]
[    3.409508] pci 0000:00:02.7: PCI bridge to [bus 08]
[    3.416155] pci 0000:00:02.7:   bridge window [io  0x8000-0x8fff]
[    3.423909] pci 0000:00:02.7:   bridge window [mem 0xfda00000-0xfdbfffff]
[    3.432788] pci 0000:00:02.7:   bridge window [mem 0xfc400000-0xfc5fffff 64bit pref]
[    3.443089] pci 0000:00:03.0: PCI bridge to [bus 09]
[    3.449426] pci 0000:00:03.0:   bridge window [io  0x9000-0x9fff]
[    3.457595] pci 0000:00:03.0:   bridge window [mem 0xfd800000-0xfd9fffff]
[    3.466044] pci 0000:00:03.0:   bridge window [mem 0xfc200000-0xfc3fffff 64bit pref]
[    3.477037] pci 0000:00:03.1: PCI bridge to [bus 0a]
[    3.483494] pci 0000:00:03.1:   bridge window [io  0xa000-0xafff]
[    3.491938] pci 0000:00:03.1:   bridge window [mem 0xfd600000-0xfd7fffff]
[    3.501378] pci 0000:00:03.1:   bridge window [mem 0xfc000000-0xfc1fffff 64bit pref]
[    3.512556] NET: Registered protocol family 2
[    3.518629] TCP established hash table entries: 512 (order: 0, 4096 bytes)
[    3.526469] TCP bind hash table entries: 512 (order: 1, 8192 bytes)
[    3.533700] TCP: Hash tables configured (established 512 bind 512)
[    3.540907] UDP hash table entries: 256 (order: 1, 8192 bytes)
[    3.548776] UDP-Lite hash table entries: 256 (order: 1, 8192 bytes)
[    3.555478] NET: Registered protocol family 1
[    3.562115] Trying to unpack rootfs image as initramfs...
[    3.629343] Freeing initrd memory: 4824K (ffff880003917000 - ffff880003dcd000)
[    3.638308] Scanning for low memory corruption every 60 seconds
[    3.645448] futex hash table entries: 256 (order: 2, 16384 bytes)
[    3.652013] audit: initializing netlink subsys (disabled)
[    3.657728] audit: type=2000 audit(1689300190.424:1): initialized
[    3.664288] Initialise system trusted keyring
[    3.669429] HugeTLB registered 1 GB page size, pre-allocated 0 pages
[    3.675912] HugeTLB registered 2 MB page size, pre-allocated 0 pages
[    3.683240] zbud: loaded
[    3.686789] VFS: Disk quotas dquot_6.6.0
[    3.691473] VFS: Dquot-cache hash table entries: 512 (order 0, 4096 bytes)
[    3.699097] fuse init (API version 7.23)
[    3.703787] Key type big_key registered
[    3.708410] Allocating IMA MOK and blacklist keyrings.
[    3.714065] Key type asymmetric registered
[    3.718745] Asymmetric key parser 'x509' registered
[    3.724035] Block layer SCSI generic (bsg) driver version 0.4 loaded (major 249)
[    3.732530] io scheduler noop registered
[    3.737114] io scheduler deadline registered (default)
[    3.742611] io scheduler cfq registered
[    3.747414] ACPI: PCI Interrupt Link [GSIG] enabled at IRQ 22
[    3.788749] ACPI: PCI Interrupt Link [GSIH] enabled at IRQ 23
[    3.811802] pcieport 0000:00:02.0: Signaling PME through PCIe PME interrupt
[    3.819313] pci 0000:01:00.0: Signaling PME through PCIe PME interrupt
[    3.826493] pcieport 0000:00:02.1: Signaling PME through PCIe PME interrupt
[    3.834008] pcieport 0000:00:02.2: Signaling PME through PCIe PME interrupt
[    3.841584] pcieport 0000:00:02.3: Signaling PME through PCIe PME interrupt
[    3.849396] pcieport 0000:00:02.4: Signaling PME through PCIe PME interrupt
[    3.856977] pci 0000:05:00.0: Signaling PME through PCIe PME interrupt
[    3.865326] pcieport 0000:00:02.5: Signaling PME through PCIe PME interrupt
[    3.872356] pci 0000:06:00.0: Signaling PME through PCIe PME interrupt
[    3.879501] pcieport 0000:00:02.6: Signaling PME through PCIe PME interrupt
[    3.886791] pci 0000:07:00.0: Signaling PME through PCIe PME interrupt
[    3.893906] pcieport 0000:00:02.7: Signaling PME through PCIe PME interrupt
[    3.901061] pci 0000:08:00.0: Signaling PME through PCIe PME interrupt
[    3.907946] pcieport 0000:00:03.0: Signaling PME through PCIe PME interrupt
[    3.914745] pci 0000:09:00.0: Signaling PME through PCIe PME interrupt
[    3.921802] pcieport 0000:00:03.1: Signaling PME through PCIe PME interrupt
[    3.929179] pci_hotplug: PCI Hot Plug PCI Core version: 0.5
[    3.935174] pciehp: PCI Express Hot Plug Controller Driver version: 0.4
[    4.021341] input: Power Button as /devices/LNXSYSTM:00/LNXPWRBN:00/input/input0
[    4.030761] ACPI: Power Button [PWRF]
[    4.035462] GHES: HEST is not enabled!
[    4.072308] Serial: 8250/16550 driver, 32 ports, IRQ sharing enabled
[    4.113872] 00:00: ttyS0 at I/O 0x3f8 (irq = 4, base_baud = 115200) is a 16550A
[    4.163678] Linux agpgart interface v0.103
[    4.171300] brd: module loaded
[    4.175976] loop: module loaded
[    4.187014]  vda: vda1 vda15
[    4.196503] libphy: Fixed MDIO Bus: probed
[    4.201807] tun: Universal TUN/TAP device driver, 1.6
[    4.207312] tun: (C) 1999-2004 Max Krasnyansky <maxk@qualcomm.com>
[    4.218447] PPP generic driver version 2.4.2
[    4.223566] ehci_hcd: USB 2.0 'Enhanced' Host Controller (EHCI) Driver
[    4.230747] ehci-pci: EHCI PCI platform driver
[    4.236535] ehci-platform: EHCI generic platform driver
[    4.242451] ohci_hcd: USB 1.1 'Open' Host Controller (OHCI) Driver
[    4.249198] ohci-pci: OHCI PCI platform driver
[    4.254286] ohci-platform: OHCI generic platform driver
[    4.260039] uhci_hcd: USB Universal Host Controller Interface driver
[    4.266840] i8042: PNP: PS/2 Controller [PNP0303:KBD,PNP0f13:MOU] at 0x60,0x64 irq 1,12
[    4.279364] serio: i8042 KBD port at 0x60,0x64 irq 1
[    4.285242] serio: i8042 AUX port at 0x60,0x64 irq 12
[    4.291015] mousedev: PS/2 mouse device common for all mice
[    4.298386] input: AT Translated Set 2 keyboard as /devices/platform/i8042/serio0/input/input1
[    4.308116] rtc_cmos 00:03: RTC can wake from S4
[    4.314768] rtc_cmos 00:03: rtc core: registered rtc_cmos as rtc0
[    4.322206] rtc_cmos 00:03: alarms up to one day, y3k, 242 bytes nvram
[    4.329527] i2c /dev entries driver
[    4.333989] device-mapper: uevent: version 1.0.3
[    4.339420] device-mapper: ioctl: 4.34.0-ioctl (2015-10-28) initialised: dm-devel@redhat.com
[    4.349135] ledtrig-cpu: registered to indicate activity on CPUs
[    4.355724] NET: Registered protocol family 10
[    4.360751] NET: Registered protocol family 17
[    4.365653] Key type dns_resolver registered
[    4.370695] microcode: CPU0 sig=0x506e3, pf=0x1, revision=0x1
[    4.376770] microcode: Microcode Update Driver: v2.01 <tigran@aivazian.fsnet.co.uk>, Peter Oruba
[    4.386398] registered taskstats version 1
[    4.390926] Loading compiled-in X.509 certificates
[    4.397222] Loaded X.509 cert 'Build time autogenerated kernel key: 6ea974e07bd0b30541f4d838a3b7a8a80d5ca9af'
[    4.408000] zswap: loaded using pool lzo/zbud
[    4.414784] Key type trusted registered
[    4.420478] Key type encrypted registered
[    4.425284] AppArmor: AppArmor sha1 policy hashing enabled
[    4.431312] ima: No TPM chip found, activating TPM-bypass!
[    4.437379] evm: HMAC attrs: 0x1
[    4.446468]   Magic number: 15:778:55
[    4.451277] rtc_cmos 00:03: setting system clock to 2023-07-14 02:03:12 UTC (1689300192)
[    4.459871] BIOS EDD facility v0.16 2004-Jun-25, 0 devices found
[    4.465687] EDD information not available.
[    4.472073] Freeing unused kernel memory: 1480K (ffffffff81f42000 - ffffffff820b4000)
[    4.480768] Write protecting the kernel read-only data: 14336k
[    4.487093] Freeing unused kernel memory: 1860K (ffff88000182f000 - ffff880001a00000)
[    4.495790] Freeing unused kernel memory: 168K (ffff880001dd6000 - ffff880001e00000)

info: initramfs: up at 2.95
modprobe: module virtio_pci not found in modules.dep
modprobe: module virtio_blk not found in modules.dep
modprobe: module virtio_net not found in modules.dep
modprobe: module vfat not found in modules.dep
modprobe: module nls_cp437 not found in modules.dep
info: copying initramfs to /dev/vda1
info: initramfs loading root from /dev/vda1
info: /etc/init.d/rc.sysinit: up at 3.13
info: container: none
Starting logging: OK
modprobe: module virtio_pci not found in modules.dep
modprobe: module virtio_blk not found in modules.dep
modprobe: module virtio_net not found in modules.dep
modprobe: module vfat not found in modules.dep
modprobe: module nls_cp437 not found in modules.dep
WARN: /etc/rc3.d/S10-load-modules failed
Initializing random number generator... [    4.750376] random: dd urandom read with 27 bits of entropy available
done.
Starting acpid: OK
mcb [info=/dev/vdb dev=/dev/vdb target=tmp unmount=true callback=mcu_drop_dev_arg]: mount '/dev/vdb' '-o,ro' '/tmp/nocloud.mp.C2qkin'
mcudda: fn=cp dev=/dev/vdb mp=/tmp/nocloud.mp.C2qkin : -a /tmp/cirros-ds.kj2MPH/nocloud/raw
Starting network...
udhcpc (v1.23.2) started
Sending discover...
Sending select for 10.0.2.2...
Lease of 10.0.2.2 obtained, lease time 86313600
Top of dropbear init script
Starting dropbear sshd: OK
GROWROOT: NOCHANGE: partition 1 is size 71647. it cannot be grown
/dev/root resized successfully [took 0.01s]
=== system information ===
Platform: KubeVirt None/RHEL
Container: none
Arch: x86_64
CPU(s): 1 @ 2303.998 MHz
Cores/Sockets/Threads: 1/1/1
Virt-type: VT-x
RAM Size: 43MB
Disks:
NAME  MAJ:MIN     SIZE LABEL         MOUNTPOINT
vda   253:0   46137344               
vda1  253:1   36683264 cirros-rootfs /
vda15 253:15   8388608               
vdb   253:16   1048576 cidata        
=== sshd host keys ===
-----BEGIN SSH HOST KEY KEYS-----
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC4Tm/D83NfS+Duh94RUGOOQPHbtpQs9tYw0CBy5wFkMWQhQEi7NVDjRPss3RkD1L45Or3hKcbxpTy4ya3jUnYux74srBLizvwpX0ttGaJzF2MwwV7lrjE97d7ui6uPnPBCNddPeDYo+BvuMRE8yLNXbPYGlUjUgMdZ9zJrLkCq/MQxB51GkR3uNAmncj5nd+hpG0UfVJLJS46R3lThs+6xB6QQgVEPwkRyW7MIk5W3E6agRlWmKQcSGL+Yb18tCXPGyYR1tCeDSfbfQ9j8N0/387WeXmLOTFuLbhk7KUoBUxTHAN/HiEdJk248mkm8a8csi2PLcvQNTQE5pQg+pfV1 root@testvm
ssh-dss AAAAB3NzaC1kc3MAAACBAPNYEfvZ4WMdlzElTpH1fZz0O1v6YO6BeQv7Btw71ETojUJ6Q1zJEsC7U3mThW96ed6ntURHf1O+0p/iEwlIKNwj69Rh7xXCLMQZ6uEX4vBEOz6b5RRxVIhjg8/HKT0E84W2lRMkBdi6SQm4tGNpDVeFpkKqiXGXkbbXxHJVroDHAAAAFQDmLsehH1V2rmkFRozxbUZPshsDGQAAAIEAoORsv63XbXjYfMa+6eeVAzT98U1riSDfOmw83kkPs6ke2xwFTa1xJaA664rf/yeRwSfIcKTdMv+pKWdQv83VCB6R7APaKuwE4ZD9WLdYDJG5zOeApl85KyZMd8DQhFXUTqkUL8755K2E4vsWVPzN5q0RqHQyWqsKxtidUjR8smUAAACAL/b+yHUhJ/GRl0CyRjS8TgZLYJzstQxbq5Ly5Sfrm/wUUckviHlIZ/USJqCTYnkyY6UNbCyFbT4EMysvHmOqCTHO+xvpJdyR+PHm4U8m/eRu0kRzzoOfUXpHm6nMz1lpPXmS9KGw4b0vtaI91K12KJ7Jzv7K0y11NDAXl4OOv2g= root@testvm
-----END SSH HOST KEY KEYS-----
=== network info ===
if-info: lo,up,127.0.0.1,8,,
if-info: eth0,up,10.0.2.2,24,fe80::5054:ff:fee2:e5b8/64,
ip-route:default via 10.0.2.1 dev eth0 
ip-route:10.0.2.0/24 dev eth0  src 10.0.2.2 
ip-route6:fe80::/64 dev eth0  metric 256 
ip-route6:unreachable default dev lo  metric -1  error -101
ip-route6:ff00::/8 dev eth0  metric 256 
ip-route6:unreachable default dev lo  metric -1  error -101
=== datasource: nocloud local ===
instance-id: 5a9fc181-957e-5c32-9e5a-2de5e9673531
name: N/A
availability-zone: N/A
local-hostname: testvm
launch-index: N/A
=== cirros: current=0.4.0 uptime=3.71 ===
  ____               ____  ____
 / __/ __ ____ ____ / __ \/ __/
/ /__ / // __// __// /_/ /\ \ 
\___//_//_/  /_/   \____/___/ 
   http://cirros-cloud.net


login as 'cirros' user. default password: 'gocubsgo'. use 'sudo' for root.
testvm login: stirabos@t14s:~$ 
```

Please notice that the serial console will remain available for the interactive use:
```bash
stirabos@t14s:~$ virtctl console -n simone testvm
Successfully connected to testvm console. The escape sequence is ^]

login as 'cirros' user. default password: 'gocubsgo'. use 'sudo' for root.
testvm login: cirros
Password: 
$ 
$ sudo sh -c "echo 'Hello, World!' >> /dev/console"
Hello, World!
$ 
$ exit

login as 'cirros' user. default password: 'gocubsgo'. use 'sudo' for root.
testvm login: stirabos@t14s:~$ 
stirabos@t14s:~$ 
```

and whatever will be written there will also be logged:
```bash
stirabos@t14s:~$ kubectl logs -n simone virt-launcher-testvm-t7h26 -c guest-console-log
[    0.000000] Initializing cgroup subsys cpuset
[    0.000000] Initializing cgroup subsys cpu
[    0.000000] Initializing cgroup subsys cpuacct
[    0.000000] Linux version 4.4.0-28-generic (buildd@lcy01-13) (gcc version 5.3.1 20160413 (Ubuntu 5.3.1-14ubuntu2.1) ) #47-Ubuntu SMP Fri Jun 24 10:09:13 UTC 2016 (Ubuntu 4.4.0-28.47-generic 4.4.13)
...
login as 'cirros' user. default password: 'gocubsgo'. use 'sudo' for root.
testvm login: cirros
Password: 
$ 
$ sudo sh -c "echo 'Hello, World!' >> /dev/console"
Hello, World!
$ 
$ exit

login as 'cirros' user. default password: 'gocubsgo'. use 'sudo' for root.
testvm login: stirabos@t14s:~$ 
```

### Guest level options
- Boot diagnostics on Serial Console can also be enabled for Windows guest, see: https://learn.microsoft.com/en-us/troubleshoot/azure/virtual-machines/serial-console-windows
- In general, journald logs are not forwarded to the serial console by default. This can be easily enabled setting `ForwardToConsole=yes` in the journald configuration but please notice that ForwardToConsole is synchronous by design and the serial console is limited to 115200 bit/second so forwarding to the serial console by default can potentially slow down the whole systemd. See: https://github.com/systemd/systemd/issues/2815

### Alternative options
As for [k8s documentation](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#advanced-features-and-flexibility) 
`Other Subresources - Add operations other than CRUD, such as "logs" or "exec". is not supported on CRDs.`.
So something like:
`kubectl logs vmi my-vm`
is not really an option with VMIs defined from a CRD.

Introducing something like `virtctl logs -n <vmnamespace> <vmname>` as wrapper for `kubectl logs -n <vmnamespace> -l vm.kubevirt.io/name=<vmname>` is straightforward, so it can be reasonably considered as a reasonable next step.
On the other side, interactively fetching the guest logs of a running VM is just one of the possible use-cases.
We can reasonably expect that being able to stream the guest console of the VMs on the cluster to an external logging stack (like Grafana Loki) to store, analyze, and query it there it will become pretty common.
All the k8s logging stack are designed to natively manage and filter pod logs according to the standard k8s logging architecture, if we design something significantly different it will risk to require further effort to be integrated into other external tools.

## API Examples

On the VM CRD on `/spec/template/spec/domain/devices/autoattachSerialConsole` we already have:
```yaml
autoattachSerialConsole:
  description: Whether to attach the default serial console
    or not. Serial console access will not be available if
    set to false. Defaults to true.
  type: boolean
```
to let the VM owner easily attach the default serial console.
We can introduce an additional boolean `/spec/template/spec/domain/devices/logSerialConsole`:
```yaml
logSerialConsole:
  description: Whether to log the auto-attached default serial
    console or not. Serial console logs will be collect to
    a file and then streamed from a named 'guest-console-log'. Not
    relevant if autoattachSerialConsole is disabled. Defaults
    to false.
  type: boolean
```
to let the VM owner require the serial console to be logged.

Since the performance impact of this feature is moderate, we can safely enable it by default giving to the cluster admin the capability to amend this choice at cluster level or to the VM owner the option to specifically opt-out. 

We should address only the default auto-attached serial console since custom serial device can potentailly contain non-standard settings. If the user really needs something so special to require a special serial console device, we can assume he can also configure the logging there.

On the VM definition side it will be something like:
```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: testvm
spec:
  template:
    spec:
      domain:
        devices:
          autoattachSerialConsole: true
          logSerialConsole: true
...
```

We can also introduce a cluster wide setting for the cluster-admin to let him tune the default behavior when nothing is specified at VM level.  

## Scalability
Trying to flood the serial console, the log file is correctly rotated by `virtlogd` taking (with current KubeVirt configuration) up to 8MB.

Let's try to flood the serial console log file:
```bash
stirabos@t14s:~$ virtctl console -n simone testvm
Successfully connected to testvm console. The escape sequence is ^]

$ sudo sh -c "while true; do date >> /dev/console; done"
Fri Jul 14 03:47:26 UTC 2023
Fri Jul 14 03:47:26 UTC 2023
...
Fri Jul 14 03:47:31 UTC 2023
Fri Jul 14 03:47:31 UTC 2023
Fri Jul 14 03:47:31 stirabos@t14s:~$ 
```
and let's check the actual file size:
```bash
stirabos@t14s:~$ kubectl exec -ti -n simone virt-launcher-testvm-t7h26 -c guest-console-log -- /bin/bash              
bash-5.1$ ls -lh /var/run/kubevirt-private/c4f004fe-f413-483c-9b8c-dc7080419bd4/
total 3.0M
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-serial0
-rw-------. 1 qemu qemu 946K Jul 14 02:56 virt-serial0-log
-rw-------. 1 qemu qemu 2.0M Jul 14 02:54 virt-serial0-log.0
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-vnc
bash-5.1$ ls -lh /var/run/kubevirt-private/c4f004fe-f413-483c-9b8c-dc7080419bd4/
total 8.0M
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-serial0
-rw-------. 1 qemu qemu 1.9M Jul 14 03:03 virt-serial0-log
-rw-------. 1 qemu qemu 2.0M Jul 14 03:00 virt-serial0-log.0
-rw-------. 1 qemu qemu 2.0M Jul 14 02:57 virt-serial0-log.1
-rw-------. 1 qemu qemu 2.0M Jul 14 02:54 virt-serial0-log.2
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-vnc
bash-5.1$ ls -lh /var/run/kubevirt-private/c4f004fe-f413-483c-9b8c-dc7080419bd4/
total 10M
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-serial0
-rw-------. 1 qemu qemu 2.0M Jul 14 03:03 virt-serial0-log
-rw-------. 1 qemu qemu 2.0M Jul 14 03:00 virt-serial0-log.0
-rw-------. 1 qemu qemu 2.0M Jul 14 02:57 virt-serial0-log.1
-rw-------. 1 qemu qemu 2.0M Jul 14 02:54 virt-serial0-log.2
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-vnc
bash-5.1$ ls -lh /var/run/kubevirt-private/c4f004fe-f413-483c-9b8c-dc7080419bd4/
total 6.1M
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-serial0
-rw-------. 1 qemu qemu  16K Jul 14 03:03 virt-serial0-log
-rw-------. 1 qemu qemu 2.0M Jul 14 03:03 virt-serial0-log.0
-rw-------. 1 qemu qemu 2.0M Jul 14 03:00 virt-serial0-log.1
-rw-------. 1 qemu qemu 2.0M Jul 14 02:57 virt-serial0-log.2
srwxrwxr-x. 1 qemu qemu    0 Jul 14 02:03 virt-vnc
bash-5.1$
```

Giving an accurate and reproducible number in terms of CPU impact is instead harder because this is strongly affected by the amount of data generated by each guest OS.
We can expect guest OSes to be generically quite verbose during the bootstrap sequence becoming then more silent during regular execution.
We can safely assume that the impact on the logging system is comparable with the impact generated by a generic container based application started by any user on the cluster.

## Update/Rollback Compatibility

NA: it's a new feature that is not breaking any past setting/expectation.
tbd: can it be enabled for existing VMs without restarting/migrating them? The behaviour should be consistent with `autoattachSerialConsole` one.

## Functional Testing Approach

We should cover the positive and the negative flows:

### Negative flow

1. Start a VM with `/spec/template/spec/domain/devices/logSerialConsole=false`
2. Wait for the VM to be running
3. Check:
   1. The first serial console and the console device in the libvirt XML should not contain the <log/> element.
   2. virt-launcher pod should not contain a container named `guest-console-log`.
   3. a file named `/var/run/kubevirt-private/<uuid>/virt-serial0-log` should not be present in the Kubevirt private volume.

### Positive flows

#### Direct tests flows
1. Start a VM with `/spec/template/spec/domain/devices/logSerialConsole=true`
2. Wait for the VM to be running
3. Check:
    1. The first serial console and the console device in the libvirt XML should contain the <log/> element.
    2. virt-launcher pod should contain a container named `guest-console-log`.
    3. The test should be able to open a serial console, login to the OS and execute a command (drop a specific log message to the serial console) there. 
    4. Trying to get logs for the `guest-console-log` container in the virt-launcher pod it should identify the specific log message it injected.
    5. A file named `/var/run/kubevirt-private/<uuid>/virt-serial0-log` should be present in the Kubevirt private volume.
    6. The password used to login on the serial console should not apper in the logs.
    7. The test should generate a huge amount of data on the serial console (20MB) repeating many times a simple command (`ls` or `date`).
    8. `virtlogd` should correctly rotate the internal log file on the `kubevirt-private` dire: we should find no more than 4 `.N` log files up to 2MB each. 

#### No regressions on other features
The feature should be enabled by default for all the VMs and we should be able to execute, with no regression nor serious performance degradation, all the existing test lanes.  

# Implementation Phases

A working POC got submitted at https://github.com/kubevirt/kubevirt/pull/10110
The final implementation can be realistically achieved with a single PR.
