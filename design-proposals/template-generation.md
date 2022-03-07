# Overview

This design defines an approach to create Virtual Machine templates starting from an existing VM.

## Motivation

Creating a 'blueprint' of an existing VM to further tweak and use as a basis golden VM template is a common operation, also it is a tool that can be leveraged by UIs to improve the UX of the project.

## Goals

* Create a generic template of an existing VM.

## Non Goals

* Designing a VM-cloning mechanism, which copies disk data or create a snapshot of the live memory of a VM.

## User Stories

* As a cluster user, I want to create a re-usable VM spec to use as the basis for my VMs workloads starting from an existing VM.
* As a cluster user, I want to create different workloads based on the same VM image by just tweaking some template parameters.

## Repos

The idea is to implement this design by adding a controller in kubevirt/ssp-operator.

# Design

The template generation API introduces a new CRD called the 'VirtualMachineTemplateRequest' object. This CRD when posted to the cluster is an explicit request to create a VM template derived from an existing VM.

The controller watching `VirtualMachineTemplateRequest` objects will see the request and post a template of the request VirtualMachine in the specified format.

## VirtualMachineTemplateRequest API

The VirtualMachineTemplateRequest specification in Golang:

```go
// VirtualMachineTemplateRequestSource is the reference VM that will be used to generate the template
type VirtualMachineTemplateRequestSource struct {
	// VMName is the name of the reference VM that will be used to generate the template
	VMName string `json:"vmName"`
	// VMNamespace is the namespace of the VM referenced by VMName
	VMNamespace string `json:"vmNamespace"`
}

// VirtualMachineTemplateRequestTarget is the target VM template that will be created
type VirtualMachineTemplateRequestTarget struct {
	// OpenShiftTemplate is the OpenShift template format
	OpenShiftTemplate *OpenShiftTemplate `json:"openshiftTemplate,omitempty"`
}

// VirtualMachineTemplateRequestSpec contains the VirtualMachineTemplateRequest specification
type VirtualMachineTemplateRequestSpec struct {
	// Source is the reference VM from which the template will be created
	Source VirtualMachineTemplateRequestSource `json:"source"`
	// Target is the target VM template that will be created
	Target VirtualMachineTemplateRequestTarget `json:"target"`
}

type OpenShiftTemplate struct {
	// Name is the name of the generated template
	Name string `json:"name,omitempty"`
}
```

The first implemented format will be the [OpenShift Template(https://docs.openshift.com/container-platform/4.9/openshift_images/using-templates.html#templates-writing_using-templates)] format.

In order to support multiple formats, the controller will have to be extended to generate the correct output for those formats.

## API Examples

Given the following VM

```yaml
apiVersion: kubevirt.io/v1
kind: VirtualMachine
metadata:
  name: vm-ephemeral
spec:
  template:
    domain:
      devices:
        disks:
        - disk:
            bus: virtio
          name: containerdisk
      resources:
        requests:
          memory: 128Mi
    volumes:
    - containerDisk:
        image: registry:5000/kubevirt/cirros-container-disk-demo:devel
      name: containerdisk
```

After posting this VirtualMachineTemplateRequest object

```yaml
apiVersion: ssp.kubevirt.io/v1alpha1
kind: VirtualMachineTemplateRequest
metadata:
  name: virtualmachinetemplaterequest-sample
spec:
  vmName: vmi-ephemeral
  templateName: vmi-ephemeral-template
  format: template.openshift.io/v1
```

The following template will be created, beware this is an example of the OpenShift template format.

```yaml
apiVersion: v1
kind: Template
metadata:
  name: vmi-ephemeral-template
  labels: ${LABELS}
objects:
- apiVersion: kubevirt.io/v1
  kind: VirtualMachine
  metadata:
    generateName: vmi-ephemeral
  template:
    spec:
      domain:
        devices:
          disks:
          - disk:
              bus: ${DISK0_BUS}
            name: containerdisk
        resources:
          requests:
            memory: ${REQUESTS_MEMORY}
      volumes:
      - containerDisk:
          image: ${VOLUME0_IMAGE}
        name: containerdisk
parameters:
- description: disk0-bus
  name: DISK0_BUS
  value: virtio
- description: requests-memory
  name: REQUESTS_MEMORY
  value: 128Mi
- description: volume0-image
  name: VOLUME0_IMAGE
  value: registry:5000/kubevirt/cirros-container-disk-demo:devel
```

## Template generation

The template generation mechanism follows some basic rules:

* every leaf field (but name fields) in the `VirtualMachineInstanceSpec` will be parameterized and given a default value according to the reference VM which has been created from.
* the VM name will be converted into a `generateName` field.
* the `dataVolumeTemplates` field inside a `VirtualMachineSpec` will be preserved as is (as it's already templated).
* the `LABELS` parameter will always be created, its content will be merged with the existing labels of the `VirtualMachine` object.

## Further developments

* Add `virtctl` support for creating templates from the cmdline.
