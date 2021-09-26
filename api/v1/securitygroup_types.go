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

package v1

import (
	// "k8s.io/api/core/v1"

	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecurityGroupSpec defines the desired state of SecurityGroup
type SecurityGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	AccountId   string `json:"accountId"`
	UserId      string `json:"userId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

const (
	ConditionTrue    string = "True"
	ConditionFalse   string = "False"
	ConditionUnknown string = "Unknown"
)

// Condition types.
const (
	// TypeReady resources are believed to be ready to handle work.
	TypeReady string = "Ready"

	// TypeSynced resources are believed to be in sync with the
	// Kubernetes resources that manage their lifecycle.
	TypeSynced string = "Synced"
)

// SecurityGroupStatus defines the observed state of SecurityGroup
type SecurityGroupStatus struct {
	// Represents the latest available observations of a securitygroup's current state.
	// +optional
	Conditions []SecurityGroupCondition `json:"conditions,omitempty"`
	Id         string                   `json:"id,omitempty"`
}

// SecurityCondition describes the state of a deployment at a certain point.
type SecurityGroupCondition struct {
	// Type of securitygroup condition.
	Type string `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status string `json:"status"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
	// The last time this condition was updated.
	LastTransitionTime string `json:"lastTransitionTime"`
}

// Equal returns true if the condition is identical to the supplied condition, ignoring the LastTransitionTime.
func (c SecurityGroupCondition) Equal(other SecurityGroupCondition) bool {
	return c.Type == other.Type &&
		c.Status == other.Status &&
		c.Reason == other.Reason &&
		c.Message == other.Message
}

// WithMessage returns a condition by adding the provided message to existing condition.
func (c SecurityGroupCondition) WithMessage(msg string) SecurityGroupCondition {
	c.Message = msg
	return c
}

// GetCondition returns the condition for the given ConditionType if exists,
// otherwise returns nil
func (s *SecurityGroupStatus) GetCondition(ct string) SecurityGroupCondition {
	for _, c := range s.Conditions {
		if c.Type == ct {
			return c
		}
	}

	return SecurityGroupCondition{Type: ct, Status: ConditionUnknown}
}

// SetConditions sets the supplied conditions, replacing any existing conditions of the same type.
//This is a no-op if all supplied conditions are identical, ignoring the last transition time, to those already set.
func (s *SecurityGroupStatus) SetConditions(c ...SecurityGroupCondition) {
	for _, new := range c {
		exists := false
		for i, existing := range s.Conditions {
			if existing.Type != new.Type {
				continue
			}

			if existing.Equal(new) {
				exists = true
				continue
			}

			s.Conditions[i] = new
			exists = true
		}
		if !exists {
			s.Conditions = append(s.Conditions, new)
		}
	}
}

// Equal returns true if the status is identical to the supplied status, ignoring the LastTransitionTimes and order of statuses.
func (s *SecurityGroupStatus) Equal(other *SecurityGroupStatus) bool {
	if s == nil || other == nil {
		return s == nil && other == nil
	}

	if len(other.Conditions) != len(s.Conditions) {
		return false
	}

	sc := make([]SecurityGroupCondition, len(s.Conditions))
	copy(sc, s.Conditions)

	oc := make([]SecurityGroupCondition, len(other.Conditions))
	copy(oc, other.Conditions)

	// We should not have more than one condition of each type.
	sort.Slice(sc, func(i, j int) bool { return sc[i].Type < sc[j].Type })
	sort.Slice(oc, func(i, j int) bool { return oc[i].Type < oc[j].Type })

	for i := range sc {
		if !sc[i].Equal(oc[i]) {
			return false
		}
	}

	return true
}

// NewConditionedStatus returns a stat with the supplied conditions set.
func NewConditionedStatus(c ...SecurityGroupCondition) *SecurityGroupStatus {
	s := &SecurityGroupStatus{}
	s.SetConditions(c...)
	return s
}

// NewCondition returns a condition
func NewCondition(conditionType, status, reason string) SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             reason,
	}
}

// Reasons a resource is or is not synced.
const (
	ReasonReconcileSuccess string = "ReconcileSuccess"
	ReasonReconcileError   string = "ReconcileError"
)

// Reasons a resource is or is not ready.
const (
	ReasonAvailable              string = "Available"
	ReasonUnavailable            string = "Unavailable"
	ReasonCreating               string = "Creating"
	ReasonDeleting               string = "Deleting"
	ReasonDSpecificationChanging string = "Updating"
)

// Creating returns a condition that indicates the resource is currently
// being created.
func Creating() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeReady,
		Status:             ConditionFalse,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonCreating,
	}
}

// Deleting returns a condition that indicates the resource is currently
// being deleted.
func Deleting() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeReady,
		Status:             ConditionFalse,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonDeleting,
	}
}

// SpecificationChanging returns a condition that indicates the resource is currently
// being Specification Change.
func SpecificationChanging() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeReady,
		Status:             ConditionFalse,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonDSpecificationChanging,
	}
}

// Available returns a condition that indicates the resource is
// currently observed to be available for use.
func Available() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeReady,
		Status:             ConditionTrue,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonAvailable,
	}
}

// Unavailable returns a condition that indicates the resource is not
// currently available for use. Unavailable should be set only when Crossplane
// expects the resource to be available but knows it is not, for example
// because its API reports it is unhealthy.
func Unavailable() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeReady,
		Status:             ConditionFalse,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonUnavailable,
	}
}

// ReconcileSuccess returns a condition indicating that Crossplane successfully
// completed the most recent reconciliation of the resource.
func ReconcileSuccess() SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeSynced,
		Status:             ConditionTrue,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonReconcileSuccess,
	}
}

// ReconcileError returns a condition indicating that Crossplane encountered an
// error while reconciling the resource. This could mean Crossplane was
// unable to update the resource to reflect its desired state, or that
// Crossplane was unable to determine the current actual state of the resource.
func ReconcileError(err error) SecurityGroupCondition {
	return SecurityGroupCondition{
		Type:               TypeSynced,
		Status:             ConditionFalse,
		LastTransitionTime: time.Now().Format(time.RFC3339),
		Reason:             ReasonReconcileError,
		Message:            err.Error(),
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=sg

// SecurityGroup is the Schema for the securitygroups API
type SecurityGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecurityGroupSpec   `json:"spec,omitempty"`
	Status            SecurityGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecurityGroupList contains a list of SecurityGroup
type SecurityGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityGroup{}, &SecurityGroupList{})
}
