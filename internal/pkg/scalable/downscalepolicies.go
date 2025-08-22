package scalable

import (
	"context"
	"fmt"
	"time"

	"github.com/caas-team/gokubedownscaler/internal/pkg/values"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	downscalePolicyGroup   = "downscaler.io"
	downscalePolicyVersion = "v1"
	downscalePolicyKind    = "DownscalePolicy"
)

var downscalePolicyGVR = schema.GroupVersionResource{
	Group:    downscalePolicyGroup,
	Version:  downscalePolicyVersion,
	Resource: "downscalepolicies",
}

// getDownscalePolicies is the getResourceFunc for DownscalePolicies.
func getDownscalePolicies(namespace string, clientsets *Clientsets, ctx context.Context) ([]Workload, error) {
	// For now, return empty list since we need dynamic client support first
	// This is a placeholder implementation
	return []Workload{}, nil
}

// downscalePolicy is a wrapper for DownscalePolicy to implement the Workload interface.
type downscalePolicy struct {
	*unstructured.Unstructured
	dynamicClient dynamic.Interface
}

// GetAnnotations gets the annotations of the DownscalePolicy.
func (d *downscalePolicy) GetAnnotations() map[string]string {
	annotations := d.Unstructured.GetAnnotations()
	if annotations == nil {
		return map[string]string{}
	}
	return annotations
}

// SetAnnotations sets the annotations on the DownscalePolicy.
func (d *downscalePolicy) SetAnnotations(annotations map[string]string) {
	d.Unstructured.SetAnnotations(annotations)
}

// Update updates the DownscalePolicy resource.
func (d *downscalePolicy) Update(clientsets *Clientsets, ctx context.Context) error {
	// For now, just return nil since we need dynamic client support
	return nil
}

// Reget regets the DownscalePolicy from the Kubernetes API.
func (d *downscalePolicy) Reget(clientsets *Clientsets, ctx context.Context) error {
	// For now, just return nil since we need dynamic client support
	return nil
}

// ScaleUp scales up the DownscalePolicy by updating its status.
func (d *downscalePolicy) ScaleUp() error {
	// Update the status to indicate upscaling
	status := map[string]interface{}{
		"phase":      "Active",
		"lastScaled": time.Now().Format(time.RFC3339),
	}

	if err := unstructured.SetNestedMap(d.Object, status, "status"); err != nil {
		return fmt.Errorf("failed to set status for upscaling: %w", err)
	}

	return nil
}

// ScaleDown scales down the DownscalePolicy by updating its status.
func (d *downscalePolicy) ScaleDown(downscaleReplicas values.Replicas) error {
	// Get the replicas value
	replicasVal, err := downscaleReplicas.AsInt32()
	if err != nil {
		return fmt.Errorf("failed to get downscale replicas: %w", err)
	}

	// Update the status to indicate downscaling
	status := map[string]interface{}{
		"phase":           "Downscaled",
		"currentReplicas": int64(replicasVal), // Convert int32 to int64 for unstructured
		"lastScaled":      time.Now().Format(time.RFC3339),
	}

	if err := unstructured.SetNestedMap(d.Object, status, "status"); err != nil {
		return fmt.Errorf("failed to set status for downscaling: %w", err)
	}

	return nil
}