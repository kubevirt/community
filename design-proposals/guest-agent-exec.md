## Virtctl Guest Agent Exec (ga-exec)

**Authors:** Ryan Hallisey NVIDIA

## Motivation
The guest-agent is a powerful tool for querying a guest, but it isn’t easy for
a user to do from kubectl.  To run a guest-agent command, the user has to locate
a launcher pod, exec into the pod, run the guest-agent-exec command, capture the pid
from the result, and run a second command - guest-agent-status - to get the result.  This
process can be simplified.

KubeVirt already has a lot of code using the guest-agent, so the idea for this proposal is
to extend the existing connections to the guest-agent all the way up to the virtctl client
so users can have ease of access too.

#### Goals
* Provide an client command for the user to run qemu guest-agent commands and return the output to stdout
* Reduce the number of client commands it takes to run a guest-agent command

#### Non-Goals
* Allowing a user to open a shell in a guest

#### Use-cases
* As a user, I want to be able to run commands inside a guest and view the output in stout
* As a user, I want to be able to view cloud-init logs with 1 command
* As a user, I want to be able to get guest OS information
* As a user, I want to be able to list users
* As a user, I want to be able to able to run filesystem commands
* As a user, I want to be able to interact with my custom guest agent
* As a user, I want my command to be cleaned up after execution

### Design
KubeVirt already has the necessary functions in place to run the guest-agent exec
command in the virt-hander and the virt-launcher, but those functions need to be
exposed through the virt-api and virtctl in order to implement this feature.

1. Create a guest-agent subresource
2. Create a rest call in virt-handler
3. Wire up the rest call to the guest-agent client
4. Create an abstraction run a command, then return the output in human readable format (if the guest-agent returns a pid, run a second command to get the commad's result)
5. Create `virtctl guest-agent <agent> <command>`

Example:
Before adding `virtctl guest-agent qemu exec`:
```bash
$ kubectl exec -it pod virt-launcher-123 - /bin/bash
$ virsh qemu-agent-command 1 '{"execute": "guest-exec", "arguments":{"capture-output": true, "path":"powershell.exe", "arg": ["type", "\\user\\logs\\cloudinit\\cloudinit.log"]}}'`
1234
$ virsh qemu-agent-command 1 '{"execute": "guest-exec-status", "arguments":{"pid": 1234}}'`
{result:somebase64string}
```

With `virtctl guest-agent qemu exec`:
```bash
$ virtctl guest-agent qemu exec powershell.exe \user\logs\cloudinit\cloudinit.log
{result:somebase64string}
```

#### Agent
The VMI can support multiple types of agents.  So having an
agent param instead of assuming `qemu` will allow room for expansion.
This proposal is only to add the qemu agent support.

#### Command
|-----|-----|
| Command | Description |
| exec | Execute a command in the guest |
| timeout | How long the 'exec' has to complete |

#### API changes
Adds the `guest-agent` subresource.
Subresources are always added to V1 so this will be added to the V1 API.
