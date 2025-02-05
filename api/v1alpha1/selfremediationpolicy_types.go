/*
Copyright 2025.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConditionType defines the type of condition to monitor
type ConditionType string

const (
	CPUUsage       ConditionType = "CPUUsage"
	MemoryUsage    ConditionType = "MemoryUsage"
	ErrorRate      ConditionType = "ErrorRate"
	PodRestarts    ConditionType = "PodRestarts"
)

// ActionType defines the type of remediation action
type ActionType string

const (
	ScaleUp           ActionType = "ScaleUp"
	ScaleDown         ActionType = "ScaleDown"
	RestartPod        ActionType = "RestartPod"
	RollbackDeployment ActionType = "RollbackDeployment"
	UpdateResources   ActionType = "UpdateResources"
	AdjustHPALimits  ActionType = "AdjustHPALimits"
)

// Condition defines what to monitor
type Condition struct {
	// Type of condition to monitor
	Type ConditionType `json:"type"`
	
	// Threshold value as a string (e.g., "80%", "100m", "2")
	Threshold string `json:"threshold"`
	
	// Duration the condition must be true before taking action
	// +optional
	Duration string `json:"duration,omitempty"`
}

// Target defines the resource to apply remediation on
type Target struct {
	// Kind of the target resource
	Kind string `json:"kind"`
	
	// Name of the target resource
	Name string `json:"name"`
	
	// Namespace of the target resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// ScalingParameters defines how scaling should be handled
type ScalingParameters struct {
	// TemporaryMaxReplicas allows temporary override of HPA/Argo maxReplicas
	// +optional
	TemporaryMaxReplicas *int32 `json:"temporaryMaxReplicas,omitempty"`

	// Duration for how long to maintain the temporary scaling
	// +optional
	ScalingDuration string `json:"scalingDuration,omitempty"`

	// RevertStrategy defines how to revert changes (Gradual or Immediate)
	// +optional
	RevertStrategy string `json:"revertStrategy,omitempty"`

	// NotificationWebhook for sending scaling decisions
	// +optional
	NotificationWebhook string `json:"notificationWebhook,omitempty"`
}

// Action defines what remediation to take
type Action struct {
	// Type of action to take
	Type ActionType `json:"type"`
	
	// Target resource for the action
	// +optional
	Target Target `json:"target,omitempty"`

	// ScalingParams for detailed scaling configuration
	// +optional
	ScalingParams *ScalingParameters `json:"scalingParams,omitempty"`

	// PreActionHook webhook to call before taking action
	// +optional
	PreActionHook string `json:"preActionHook,omitempty"`

	// PostActionHook webhook to call after taking action
	// +optional
	PostActionHook string `json:"postActionHook,omitempty"`

	// ConflictResolution defines how to handle conflicts with other controllers
	// +optional
	ConflictResolution string `json:"conflictResolution,omitempty"`
}

// Rule defines a single remediation rule
type Rule struct {
	// Name of the rule
	Name string `json:"name"`
	
	// Conditions that trigger the rule
	Conditions []Condition `json:"conditions"`
	
	// Actions to take when conditions are met
	Actions []Action `json:"actions"`
}

// GrafanaIntegration defines optional Grafana integration settings
type GrafanaIntegration struct {
	// Whether Grafana integration is enabled
	Enabled bool `json:"enabled"`
	
	// Webhook URL for Grafana alerts
	// +optional
	WebhookURL string `json:"webhookUrl,omitempty"`
}

// SelfRemediationPolicySpec defines the desired state
type SelfRemediationPolicySpec struct {
	// Rules for remediation
	Rules []Rule `json:"rules"`
	
	// CooldownPeriod between remediation actions
	// +optional
	CooldownPeriod string `json:"cooldownPeriod,omitempty"`
	
	// GrafanaIntegration configuration
	// +optional
	GrafanaIntegration *GrafanaIntegration `json:"grafanaIntegration,omitempty"`
}

// SelfRemediationPolicyStatus defines the observed state
type SelfRemediationPolicyStatus struct {
	// Last time the policy was evaluated
	LastEvaluationTime *metav1.Time `json:"lastEvaluationTime,omitempty"`
	
	// Last remediation action taken
	LastRemediationAction string `json:"lastRemediationAction,omitempty"`
	
	// Current state of the policy
	State string `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SelfRemediationPolicy is the Schema for the selfremediationpolicies API
type SelfRemediationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SelfRemediationPolicySpec   `json:"spec,omitempty"`
	Status SelfRemediationPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SelfRemediationPolicyList contains a list of SelfRemediationPolicy
type SelfRemediationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SelfRemediationPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SelfRemediationPolicy{}, &SelfRemediationPolicyList{})
}
