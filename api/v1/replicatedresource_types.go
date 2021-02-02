/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicatedResourceSpec defines the desired state of ReplicatedResource
type ReplicatedResourceSource struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Kind      string `json:"kind,omitempty"`
}

// ReplicatedResourceSpec defines the desired state of ReplicatedResource
type ReplicatedResourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Source ReplicatedResourceSource `json:"source,omitempty"`
}

type ReplicatedResourceConditionType string

// These are valid conditions of a ReplicatedResource.
const (
	// ReplicatedResourceComplete means the ReplicatedResource has completed its execution.
	ReplicatedResourceComplete ReplicatedResourceConditionType = "Complete"
	// ReplicatedResourceFailed means the ReplicatedResource has failed its execution.
	ReplicatedResourceFailed ReplicatedResourceConditionType = "Failed"
)

// ReplicatedResourceCondition describes current state of a ReplicatedResource.
type ReplicatedResourceCondition struct {
	// Type of ReplicatedResource condition, Complete or Failed.
	Type ReplicatedResourceConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ReplicatedResourceConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

// ReplicatedResourceStatus defines the observed state of ReplicatedResource
type ReplicatedResourceStatus struct {
	Phase      string                        `json:"phase,omitempty"`
	Conditions []ReplicatedResourceCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ReplicatedResource is the Schema for the replicatedresources API
type ReplicatedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicatedResourceSpec   `json:"spec,omitempty"`
	Status ReplicatedResourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicatedResourceList contains a list of ReplicatedResource
type ReplicatedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicatedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReplicatedResource{}, &ReplicatedResourceList{})
}
