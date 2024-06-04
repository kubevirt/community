# SIG Lifecycle Charter

## Scope

The scope of SIG Lifecycle is currently limited to the implementation and graduation of the `instancetype.kubevirt.io` API and CRDs to `v1`.

In the future it is hoped to increase the scope of the SIG to other aspects of the `VirtualMachine` lifecycle within KubeVirt as part of the process to break up the responsibilities of the current default `sig-compute`.

### In scope

- The `instancetype.kubevirt.io` API and CRDs
- `pkg/instancetype`
- `pkg/virtctl/create/instancetype`
- `pkg/virtctl/create/preference`
- `pkg/virt-api/webhooks/validating-webhook/admitters/instancetype-admitter.go`
- `pkg/virt-operator/resource/apply/instancetypes.go`
- `pkg/virt-operator/resource/generate/components/instancetypes.go`
- `pkg/virt-operator/resource/generate/components/data/common-clusterinstancetypes-bundle.yaml`
- `pkg/virt-operator/resource/generate/components/data/common-clusterpreferences-bundle.yaml`
- `tests/libinstancetype`
- `tests/instancetype`
- `kubevirt/common-instancetypes`

### Out of scope

- All other aspects of a `VirtualMachine` lifecycle

## Roles and Organization Management

This sig follows the Roles and Organization Management outlined in [OWNER_ALIASES](https://github.com/kubevirt/kubevirt/blob/main/OWNERS_ALIASES)
file.

SIG chairs:
- lyarwood
- 0xFelix

### Additional responsibilities of Chairs

- Be welcoming to new contributors
- Resolve conflicts
