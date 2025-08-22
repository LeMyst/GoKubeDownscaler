package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DownscalePolicy represents a custom resource for managing workload downscaling
type DownscalePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DownscalePolicySpec   `json:"spec,omitempty"`
	Status DownscalePolicyStatus `json:"status,omitempty"`
}

// DownscalePolicySpec defines the desired state of DownscalePolicy
type DownscalePolicySpec struct {
	// Downtime defines timespans where workloads will be scaled down
	Downtime string `json:"downtime,omitempty"`

	// Uptime defines timespans where workloads will be scaled up
	Uptime string `json:"uptime,omitempty"`

	// ForceDowntime defines timespans where workloads will be forced to scale down
	ForceDowntime string `json:"forceDowntime,omitempty"`

	// ForceUptime defines timespans where workloads will be forced to scale up
	ForceUptime string `json:"forceUptime,omitempty"`

	// DownscaleReplicas defines the number of replicas to scale down to
	DownscaleReplicas *int32 `json:"downscaleReplicas,omitempty"`

	// GracePeriod defines duration a workload must exist before being scaled
	GracePeriod string `json:"gracePeriod,omitempty"`

	// Exclude defines timespans when workload should be excluded from scaling
	Exclude string `json:"exclude,omitempty"`

	// ExcludeUntil excludes workload from scaling until this timestamp
	ExcludeUntil *metav1.Time `json:"excludeUntil,omitempty"`
}

// DownscalePolicyStatus defines the observed state of DownscalePolicy
type DownscalePolicyStatus struct {
	// LastScaled indicates the last time the workload was scaled
	LastScaled *metav1.Time `json:"lastScaled,omitempty"`

	// CurrentReplicas indicates the current number of replicas
	CurrentReplicas *int32 `json:"currentReplicas,omitempty"`

	// Phase indicates the current phase of the downscale policy
	Phase string `json:"phase,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DownscalePolicyList contains a list of DownscalePolicy
type DownscalePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DownscalePolicy `json:"items"`
}