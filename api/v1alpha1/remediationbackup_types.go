package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// RemediationBackupSpec defines the desired state of RemediationBackup
type RemediationBackupSpec struct {
	// Original state of the resource before remediation
	OriginalState runtime.RawExtension `json:"originalState"`

	// Reference to the remediated resource
	ResourceRef ResourceReference `json:"resourceRef"`

	// Reference to the remediation policy that triggered the action
	PolicyRef ResourceReference `json:"policyRef"`

	// Type of remediation action taken
	ActionType string `json:"actionType"`

	// Timestamp when the backup was created
	BackupTime metav1.Time `json:"backupTime"`

	// Time to live for this backup
	TTL *metav1.Duration `json:"ttl,omitempty"`

	// Labels from the original resource
	OriginalLabels map[string]string `json:"originalLabels,omitempty"`

	// Annotations from the original resource
	OriginalAnnotations map[string]string `json:"originalAnnotations,omitempty"`
}

// ResourceReference contains information to find a Kubernetes resource
type ResourceReference struct {
	// API Group of the resource
	APIGroup string `json:"apiGroup"`

	// Kind of the resource
	Kind string `json:"kind"`

	// Name of the resource
	Name string `json:"name"`

	// Namespace of the resource
	Namespace string `json:"namespace"`
}

// RemediationBackupStatus defines the observed state of RemediationBackup
type RemediationBackupStatus struct {
	// Whether this backup can be used for rollback
	IsValid bool `json:"isValid"`

	// Last time the backup was validated
	LastValidationTime *metav1.Time `json:"lastValidationTime,omitempty"`

	// Any validation errors
	ValidationErrors []string `json:"validationErrors,omitempty"`

	// Size of the backup in bytes
	BackupSizeBytes int64 `json:"backupSizeBytes,omitempty"`

	// Hash of the backup content for integrity verification
	ContentHash string `json:"contentHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Valid",type="boolean",JSONPath=".status.isValid"
//+kubebuilder:printcolumn:name="Resource",type="string",JSONPath=".spec.resourceRef.kind"

// RemediationBackup is the Schema for the remediationbackups API
type RemediationBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemediationBackupSpec   `json:"spec,omitempty"`
	Status RemediationBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RemediationBackupList contains a list of RemediationBackup
type RemediationBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemediationBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemediationBackup{}, &RemediationBackupList{})
}
