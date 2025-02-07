## **Overview**

This is a proposal to include an `Object Graph` API in KubeVirt to represent VM and VMI dependencies and their relationships.

## **Motivation**

As new features continue to be added to KubeVirt, the graph of objects related to VMs (DataVolumes, PersistentVolumeClaims, InstanceTypes, Preferences, Secrets, ConfigMaps, etc.) continues to expand. Identifying all the objects that a VM depends on for tasks like backup, disaster recovery, or migration can be error-prone. We should simplify this process for users and partners by creating an authoritative way to retrieve a structured object graph.

## **Goals**

- Define an Object Graph API to represent the dependency list of VMs and VMIs.
- Create an API subresource to expose the dependency list.
- Allow filtering of resources in some way.
- Ensure extensibility for future enhancements.

### **Non-Goals**

- Introducing a redundant API that reimplements the VM or VMI Spec.

## **User Stories**

1. As a KubeVirt user, I want a clear way to retrieve all VM and VMI-related dependencies.
2. As a backup partner, I want a way to identify a list of a VM's related objects so I can comprehensively backup and restore everything a VM needs.
3. As a VM owner, I want to easily define an ACM-discovered application and protect my VM with disaster recovery software.
4. As a VM owner, I want to migrate my VM from one cluster to another and identify all necessary dependencies for replication.
5. As a KubeVirt developer, I want a specific place to keep the object graph code updated when I introduce code that changes the relationship of a VM to its dependent objects.

## **Design**

### **ObjectGraph API**
The proposed API provides access to the resource dependency via the following endpoints:

```
/apis/subresources.kubevirt.io/v1/namespaces/{namespace}/virtualmachineinstances/{name}/objectgraph
/apis/subresources.kubevirt.io/v1/namespaces/{namespace}/virtualmachines/{name}/objectgraph
```

### **Data Representation Options**

Currently considering if the API should return a hierarchical representation or a flat list of dependencies.

#### **1. Hierarchical Representation**

This format visualizes dependencies in a tree structure, showing parent-child relationships between resources.

```go
// ObjectGraphNode represents an individual resource node in the graph.
type ObjectGraphNode struct {
	ObjectReference k8sv1.TypedObjectReference `json:"objectReference"`
	Labels          map[string]string          `json:"labels,omitempty"`
	Optional        bool                       `json:"optional"`
	Children        []ObjectGraphNode          `json:"children,omitempty"`
}

// ObjectGraphNodeList represents a list of object graph nodes.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ObjectGraphNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectGraphNode `json:"items"`
}
```

##### Example Output
```json
{
  "items": [
    {
      "objectReference": {
        "apiGroup": "kubevirt.io",
        "kind": "virtualmachineinstances",
        "name": "vm-cirros-source-ocs",
        "namespace": "default"
      },
      "labels": {},
      "optional": false,
      "children": [
        {
          "objectReference": {
            "apiGroup": "",
            "kind": "pods",
            "name": "virt-launcher-vm-cirros-source-ocs-frn9h",
            "namespace": "default"
          },
          "labels": {},
          "optional": false,
          "children": []
        }
      ]
    },
    {
      "objectReference": {
        "apiGroup": "cdi.kubevirt.io",
        "kind": "datavolumes",
        "name": "cirros-dv-source-ocs",
        "namespace": "default"
      },
      "labels": {
        "type": "storage"
      },
      "optional": false,
      "children": [
        {
          "objectReference": {
            "apiGroup": "",
            "kind": "persistentvolumeclaims",
            "name": "cirros-dv-source-ocs",
            "namespace": "default"
          },
          "labels": {
            "type": "storage"
          },
          "optional": false,
          "children": []
        }
      ]
    }
  ]
}
```

#### **2. Flat Dependency List**

This format provides a simple list of dependencies without indicating hierarchical relationships.

```go
// ObjectGraphNode represents a node in the object graph.
type ObjectGraphNode struct {
	ObjectReference k8sv1.TypedObjectReference `json:"objectReference"`
	Labels          map[string]string          `json:"labels,omitempty"`
}

// ObjectGraphNodeList represents a list of object graph nodes.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ObjectGraphNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ObjectGraphNode `json:"items"`
}
```

##### Example Output

```json
{
  "items": [
    {
      "objectReference": {
        "apiGroup": "kubevirt.io",
        "kind": "virtualmachineinstances",
        "name": "vm-cirros-source-ocs",
        "namespace": "default"
      },
      "labels": {
        "type": "",
        "optional": "false"
      }
    },
    {
      "objectReference": {
        "apiGroup": "",
        "kind": "pods",
        "name": "virt-launcher-vm-cirros-source-ocs-frn9h",
        "namespace": "default"
      },
      "labels": {
        "type": "",
        "optional": "false"
      }
    },
    {
      "objectReference": {
        "apiGroup": "cdi.kubevirt.io",
        "kind": "datavolumes",
        "name": "cirros-dv-source-ocs",
        "namespace": "default"
      },
      "labels": {
        "type": "storage",
        "optional": "false"
      }
    },
    {
      "objectReference": {
        "apiGroup": "",
        "kind": "persistentvolumeclaims",
        "name": "cirros-dv-source-ocs",
        "namespace": "default"
      },
      "labels": {
        "type": "storage",
        "optional": "false"
      }
    }
  ]
}
```

### **Considerations**

1. **Naming:** Is `ObjectGraph` descriptive enough even if we are returning a flat list of objects? Would `DependencyList` be more accurate?
2. **Extensibility:** How can we ensure the API is extensible for future enhancements? Should the API be made more intelligent (with fields such as `Optional`) or just rely on labels for extensibility?
3. **Filtering:** Should the user handle filtering, or should we allow some kind of filtering in the ObjectGraph request?
4. **API Design:** Should we use a hierarchical or flat list representation for the Object Graph API?

### **User Flow**

1. Access the ObjectGraph API through the subresource endpoint for a VM/VMI.
2. Parse the response and filter unnecessary objects (e.g., in backup scenarios).
3. Use the retrieved data as needed.

### **Included Resources**

- **InstanceType (`spec.instancetype.name`)**
  - **ControllerRevision (`spec.instancetype.revisionName`)**
- **Preference (`spec.preference.name`)**
  - **ControllerRevision (`spec.preference.revisionName`)**
- **VirtualMachineInstance (VMI):** Identified by VM name.  
  - **Virt-launcher Pod:** Identified by label.
  - **Volumes**:
    - **DataVolumes (`spec.template.spec.volumes[*].dataVolume`)**
    - **PersistentVolumeClaims (`spec.template.spec.volumes[*].persistentVolumeClaim`)**
    - **ConfigMaps (`spec.template.spec.volumes[*].configMap`)**
    - **Secrets (`spec.template.spec.volumes[*].secret`)**
    - **ServiceAccounts (`spec.template.spec.volumes[*].serviceAccount`)**
    - **MemoryDump (`spec.template.spec.volumes[*].memoryDump`)**
  - **AccessCredentials**
    - **SSH Secrets (`spec.template.spec.accessCredentials.sshPublicKey.source.secret`)**
    - **User Password Secrets (`spec.template.spec.accessCredentials.userPassword.source.secret`)**

**Backend Storage PVC**  
Identified by the persistent state PVC label.

**Other Resources:**  
- Should we include `networkAttachmentDefinitions` and `networks` in the ObjectGraph?  
- Should optional objects such as `VMExports` or `VMSnapshots` be considered?  

## **Implementation Phases**

1. Implement API fields to represent the ObjectGraph and its nodes.
2. Include Object Graph parser in virt-api that generates the dependency list.
3. Create subresource endpoints for VirtualMachine and VirtualMachineInstance.
4. Implement virtctl commands to interact with the ObjectGraph API.