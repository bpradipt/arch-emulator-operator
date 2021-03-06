/*


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

package v1alpha1

import (
	conditions "github.com/openshift/custom-resource-status/conditions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArchEmulatorSpec defines the desired state of ArchEmulator
type ArchEmulatorSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Specifies details of the emulator
	EmulatorType ArchEmulatorType `json:"emulatorType"`

	// ArchEmulatorNodeSelector is used to filter the nodes on which to install
	// emulator
	// if not specified, all worker nodes are selected
	// +optional
	// +nullable
	EmulatorNodeSelector *metav1.LabelSelector `json:"emulatorNodeSelector"`
}

type ArchEmulatorType struct {

	// Emulator name - eg. Qemu
	EmulatorName string `json:"emulatorName"`

	// Container image providing the Emulator binary
	// +optional
	EmulatorImage string `json:"emulatorImage"`
}

// ArchEmulatorStatus defines the observed state of Arch Emulator
type ArchEmulatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Nodes on which the emulator is installed
	Nodes []string `json:"nodes"`

	// Emulator name - Qemu etc
	EmulatorType ArchEmulatorType `json:"emulatorType"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +optional
	// Conditions is a list of conditions related to operator reconciliation
	Conditions []conditions.Condition `json:"conditions,omitempty"  patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ArchEmulator is the Schema for the archemulators API
type ArchEmulator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArchEmulatorSpec   `json:"spec,omitempty"`
	Status ArchEmulatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ArchEmulatorList contains a list of ArchEmulator
type ArchEmulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArchEmulator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArchEmulator{}, &ArchEmulatorList{})
}
