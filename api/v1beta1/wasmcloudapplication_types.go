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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	oamv1beta1 "github.com/oam-dev/kubevela/apis/core.oam.dev/v1beta1"
)

// WasmCloudApplicationStatus defines the observed state of WasmCloudApplication
type WasmCloudApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status from wasmCloud lattice controller when requesting the change or updating
	FromLatticeController string `json:"fromLatticeController"`
	// The time at which the request was made to the lattice controller
	TimeApplied string `json:"TimeApplied"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WasmCloudApplication is the Schema for the wasmcloudapplications API
type WasmCloudApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   oamv1beta1.ApplicationSpec `json:"spec,omitempty"`
	Status WasmCloudApplicationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WasmCloudApplicationList contains a list of WasmCloudApplication
type WasmCloudApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WasmCloudApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WasmCloudApplication{}, &WasmCloudApplicationList{})
}
