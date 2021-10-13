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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SugarshopSpec defines the desired state of Sugarshop
type SugarshopSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Sugarshop. Edit sugarshop_types.go to remove/update
	SourceVolumeName string `json:"sourceVolumeName,omitempty"`
	SourceClaimName  string `json:"sourceClaimName,omitempty"`
	Namespace        string `json:"namespace,omitempty"`
	HostPath         string `json:"hostPath,omitempty"`
}

// SugarshopStatus defines the observed state of Sugarshop
type SugarshopStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Complete      bool   `json:"complete,omitempty"`
	NewVolumeName string `json:"newVolumeName,omitempty"`
	NewClaimName  string `json:"newClaimName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Sugarshop is the Schema for the sugarshops API
type Sugarshop struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SugarshopSpec   `json:"spec,omitempty"`
	Status SugarshopStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SugarshopList contains a list of Sugarshop
type SugarshopList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sugarshop `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sugarshop{}, &SugarshopList{})
}
