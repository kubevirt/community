# seitan: Security Aspects

- [Benefits over Existing Solutions](#benefits-over-existing-solutions)
  * [Declarative approach versus imperative approach](#declarative-approach-versus-imperative-approach)
  * [Enabling testing of resource access mechanisms and privileged operations](#enabling-testing-of-resource-access-mechanisms-and-privileged-operations)
  * [Bytecode is defined by the VMI specification, with hardcoded values and paths](#bytecode-is-defined-by-the-vmi-specification--with-hardcoded-values-and-paths)
  * [Access to privileged resources available per process and not per container](#access-to-privileged-resources-available-per-process-and-not-per-container)
  * [System call counters](#system-call-counters)
- [Security filters, interaction with existing ones and with Linux Security Modules](#security-filters--interaction-with-existing-ones-and-with-linux-security-modules)
  * [Usage as security filter](#usage-as-security-filter)
  * [Compatibility with other seccomp filters](#compatibility-with-other-seccomp-filters)
- [Attack Surface](#attack-surface)
  * [Malicious JSON](#malicious-json)
  * [Malicious bytecode](#malicious-bytecode)
  * [Malicious BPF programme](#malicious-bpf-programme)
  * [Malicious system call arguments](#malicious-system-call-arguments)
  * [Unprivileged process impersonation](#unprivileged-process-impersonation)
  * [Rogue target process](#rogue-target-process)
  * [Rogue seitan process](#rogue-seitan-process)
- [Frequently Asked Questions](#frequently-asked-questions)
  * [Does seitan make unprivileged processes... privileged?](#does-seitan-make-unprivileged-processes-privileged-)
  * [How do you avoid that a process impersonates another one?](#how-do-you-avoid-that-a-process-impersonates-another-one-)
  * [Does this allow a process to issue any system call?](#does-this-allow-a-process-to-issue-any-system-call-)
  * [Does this approach defy the usage of a Linux Security Module?](#does-this-approach-defy-the-usage-of-a-linux-security-module-)
  * [Could we use the log of a Linux Security Module instead?](#could-we-use-the-log-of-a-linux-security-module-instead-)
  * [Why isn't a real programming language used to describe actions?](#why-isn-t-a-real-programming-language-used-to-describe-actions-)
  * [If I use the JSON model as a security filter, can another thread in the same process context write to the memory area pointed to by system call arguments, while the calling thread is blocked, and defy the purpose of the filter?](#if-i-use-the-json-model-as-a-security-filter--can-another-thread-in-the-same-process-context-write-to-the-memory-area-pointed-to-by-system-call-arguments--while-the-calling-thread-is-blocked--and-defy-the-purpose-of-the-filter-)

## Benefits over Existing Solutions

### Declarative approach versus imperative approach

The declarative approach embodied by seitan enables expressing rules and
permissions as what the programs can do instead of how. In this way, the
security logic isn't scattered all over the application code, but it is handled
externally by seitan and modelled by relatively terse matches and actions.

A declarative model is generally preferred over the imperative model, and is
commonly employed by components implementing security policies, such as Linux
Security Modules (SELinux, AppArmor, TOMOYO, etc.).

The permitted system calls and relative actions are listed in the seitan recipe
and their execution can be registered into the auditing log output.

### Enabling testing of resource access mechanisms and privileged operations

Unifying the resource access mechanisms and privileged operations into seitan
enables better testing. Today, these operations are spread all over the entire
KubeVirt code and are hard to find, read and test.

seitan is a distinct project, it can be tested separately, and doesn't require,
for this purpose, all the additional infrastructure KubeVirt needs, such as
setting up a Kubernetes cluster and installing KubeVirt on it.

### Bytecode is defined by the VMI specification, with hardcoded values and paths

The bytecode actions and the seccomp filters are generated based on the VMI
definition before monitoring the process. All the paths and variable names need
to be solved and predefined before the generation of seitan inputs, and then
they are hardcoded into the bytecode.

In this way, it isn't possible to trick seitan into opening another file or
connecting to another socket while the monitored process is running.

### Access to privileged resources available per process and not per container

Today, KubeVirt creates or makes available the resources directly in the
container. However, once they are accessible into the container, any process
running in it can access and use these resources.

With seitan, resource access is granted per process, and not per container. For
example, the monitored process (e.g. QEMU) could ask to access and open a file,
but another process, executed for instance via `kubectl exec`, cannot read the
same file.

### System call counters

seitan offers the possibility to limit the number of the expected syscalls. This
prevents the execution of an additional malicious system calls, like for example
to open again a file.

This is a minimally stateful implementation. If the need arises, further states
could be easily introduced -- for example, it might be required that usage of
one system call is enabled only after a related system call has been performed.

## Security filters, interaction with existing ones and with Linux Security Modules

### Usage as security filter

It's also possible to use seitan as a natural extension of seccomp, as a
filtering mechanism: seitan provides deep inspection of system call arguments
and flags, with corresponding, tailored actions, in a way that's normally not
possible with BPF programs alone.

This is currently not in scope for the current KubeVirt integration, but it
might be considered to improve the usefulness of seccomp over the current
approach using "default" filters.

For instance, the current filter employed by the cri-o runtime allows
approximately 370 system calls, severely impacting the efficacy of a seccomp
filter: this is because it's very complicated to define a profile which simply
blocks or allows system calls solely based on the system call number itself.

Instead, a fine-grained, extended seccomp policy, that can rely on argument
inspection, would most likely make an allow-list approach possible.

See also the related question below: "If I use the JSON model as a security
filter, can another thread in the same process context write to the memory area
pointed to by system call arguments, while the calling thread is blocked, and
defy the purpose of the filter?".

### Compatibility with other seccomp filters

The seccomp filter installed by seitan is appended to the existing seccomp
filters. This implies that if a syscall was blocked by a previously installed
filter, then it remains blocked also for the seitan filter. In this way, it is
impossible to bypass already installed seccomp filters and the filters are
incrementally added.

## Attack Surface

### Malicious JSON

Crafting a malicious JSON input, later fed to seitan-cooker, with the intention
of attacking availability or integrity of a pod, requires cooperation on behalf
of virt-handler: this is because the JSON recipe is prepared by virt-handler,
and also fully consumed by seitan-cooker, before the virtual environment is
started.

If virt-handler is compromised in a way that enables this kind of attack,
presumably requiring arbitrary execution, it seems unlikely that seitan would
offer any additional attack vector.

### Malicious bytecode

A malicious bytecode could be the result of a malicious JSON input, see above.

A malformed bytecode, however, could be the result of programming mistakes in
seitan-cooker, and possibly compromise integrity and availability of a
virtualised workload.

Defences include:
* a small codebase for seitan-cooker, which needs to be easily auditable
* bytecode simplicity (currently under design/implementation)
* a robust evaluation path in seitan, not allowing e.g. execution loops
* LSM policies for seitan and seitan's own seccomp profile
* functional, and fuzz testing

### Malicious BPF programme

A malicious BPF programme could equally be the result of a malicious JSON input,
see above.

Similarly, though, a malformed BPF programme could be a result of programming
mistakes in seitan-cooker, and affect availability of a virtualised workload.

Defences include:
* a small codebase for seitan-cooker, which needs to be easily auditable
* functional, and fuzz testing

### Malicious system call arguments

System call arguments are processed by seitan according to bytecode directives.
A vulnerability in this processing could result in compromised integrity and
availability of both virtualised workloads and host system.

seitan focuses on making this processing clearly observable and auditable, by
relying on a simple, robust bytecode, generated by a declarative input form.

Observed system call arguments are simply compared, with very limited parsing
for some system calls only, and actions can be easily written without reusing
the observed arguments.

The proposed example for handling the container disks case compares, then
discards the path originally indicated by the unprivileged process, replacing
the whole argument in the resulting system call, which should be inherently
safer compared to a full-fledged parsing of input arguments.

For instance, if an unprivileged process tries to access or connect to a
resource that's not explicitly allowed, the access won't succeed. seitan
effectively implements a form of mandatory access control.

Defences against weaknesses in the system call processing include:
* a robust evaluation path in seitan
* LSM policies for seitan and seitan's own seccomp profile
* functional, and fuzz testing

### Unprivileged process impersonation

This appears unlikely to be a usable attack vector, see also the question "How
do you avoid that a process impersonates another one?" below.

If a process, however, succeeds in impersonating another process, it will gain
access to its granted operations.

Defences:
* identification of process as described in "How do you avoid that a process
  impersonates another one?"
* LSM policies for seitan and seitan's own seccomp profile

### Rogue target process

Arbitrary execution in an unprivileged process is assumed to be possible at any
time.

In case the rogue process tries to execute privileged operations differently
from the operations declared in the seitan recipe, it will fail because:

* if the system call isn't in the set of calls to be handled, it will be
  silently ignored

* if the system call uses different arguments than the ones that would be
  expected by a given seitan recipe, the match will fail and, depending on the
  configuration, the system call can either be blocked or executed in the
  original context

### Rogue seitan process

seitan needs to run as a privileged process in order to perform certain
operations on behalf of the unprivileged containers. For this reason, the same
care and protection for virt-handler need to be applied to seitan too.

Defences:
* LSM policies for seitan and seitan's own seccomp profile
* Small and auditable code-base for seitan
* No dynamic memory allocation in seitan
* No API exposed after unprivileged processes are monitored
* Functional, and fuzz testing

## Frequently Asked Questions

### Does seitan make unprivileged processes... privileged?

One of the intentions behind seitan is to make privileged operations, currently
triggered by unprivileged processes using a number of different mechanisms,
explicit and auditable.

This does not imply additional paths for unprivileged processes to trigger
privileged operations or a different set of privileged operations: rather, the
existing operations are outlined in a clear way, based on what's currently
intended to be possible, while recognizing that a number of privileged
operations are needed for some functionalities: reducing the privileges required
to perform those operations is certainly not conducive to improving security,
and neither is impeding functionality.

Reducing the gap between the operations that are currently intended to be
possible and the operations that are actually, unintendedly possible is a main
objective of this solution -- as they are now concisely declared, and not
implemented by thousands of lines of code.

*Long story short*: not really, we want to show what privileged operations
unprivileged processes can already trigger in a simple, centralised, declarative
fashion.

### How do you avoid that a process impersonates another one?

Unprivileged target processes are added to the list of monitored processes,
using their PID, by the same component (virt-handler) which currently handles
most of the existing mechanisms to trigger privileged operations on their
behalf. This makes the security level arguably at least equivalent to the
current mechanisms.

On top of this ( **TODO**: currently under evaluation), it's also possible to
filter monitored processes based on their executable paths (which, coupled with
a properly configured Linux Security Module, should be a fairly robust solution)
and based on their ELF signature. The corresponding signature needs to be
available to the component (virt-handler) configuring the input files.

### Does this allow a process to issue any system call?

Not at all! The list of allowed system calls for each monitored process is
clearly defined (see the examples), and anything else is either allowed or
blocked, depending on configuration options.

### Does this approach defy the usage of a Linux Security Module?

When seitan performs an operation on behalf of an unprivileged component, LSM
policies configured to prevent only given sets of processes (not including
seitan) are not applied in an equivalent way.

For instance, if the unprivileged process attempts to open `/var/run/file`, and
this operation is intercepted, with seitan configured to open
`/var/run/other_file`, and the LSM policy prevents the opening of
`/var/run/file`, the operation will succeed.

To offer an equivalent level of security, we need to ensure that the LSM policy
configured for a set of unprivileged processes is applied to the same child
process of seitan dealing with those unprivileged processes.

In AppArmor terms, if subprofiles for specific unprivileged environments are
used, seitan could cooperate by calling `aa_change_hat()` upon any `fork()` call
that's intended to create a child process dealing with that specific
environment. For SELinux, a type transition needs to be configured in such a way
that seitan processes also inherit policies equivalent to the ones intended for
the unprivileged processes.

We plan to ship example AppArmor and SELinux profiles implementing this
mechanism.

*Short answer*: no, but extra care needs to be used.

### Could we use the log of a Linux Security Module instead?

That approach can be certainly practical for threat analysis and, to some
extent, confinement: system calls and their parameters can be tracked in
auditing logs and processes can be reported or terminated based on configured
rules.

However, in this solution, we want to target the issue of executing actions at
the time they are requested by target processes. We can't execute them at a
later time, and we might need to atomically replace file descriptors (using the
`SECCOMP_ADDFD_FLAG_SEND` flag) in unprivileged processes. This level of
interaction is not possible by merely examining logs.

### Why isn't a real programming language used to describe actions?

For two reasons:

1. While a binding to programming languages would indeed offer some additional
   flexibility, it also makes the description of actions harder to audit,
   because of the very flexibility of using an imperative approach to describe
   operations.

2. It’s probably not needed: as we evaluated examples of current needs in
   KubeVirt, it looks apparent that this kind of flexibility is not actually
   desirable.

For instance, in the case of container disks, all we need to do is to open a
well-defined path from a different mount namespace -- not to tell another
component that it should open a given `../../../../path` from a different mount
namespace and return a file descriptor corresponding to it.

### If I use the JSON model as a security filter, can another thread in the same process context write to the memory area pointed to by system call arguments, while the calling thread is blocked, and defy the purpose of the filter?

Indeed, another thread could do that: this is the time-of-check-time-of-use
(TOCTOU) attack idea outlined in `seccomp_unotify(2)`. To prevent this, if the
evaluation of a portion of a filter depends on the contents of a process memory
area, unless continuation of the system call is not requested using a
`"continue": "unsafe"` directive, whenever such continuation
( `SECCOMP_USER_NOTIF_FLAG_CONTINUE`) is requested, seitan will preemptively
perform a deep copy of the arguments, perform argument evaluation on the copy,
and replace the pointers to its own copy (using `ptrace(2)`) before proceeding.

This is not a generic mechanism: the deep copy can obviously be performed only
for system calls that are supported by seitan. While support for a specific
system call will be quick to implement, if a new system call appears, and it's
unsupported at any given moment, seitan will refuse handling a model that
contains it.
