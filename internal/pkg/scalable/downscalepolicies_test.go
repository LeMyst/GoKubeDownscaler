package scalable

import (
	"context"
	"testing"
	"time"

	"github.com/caas-team/gokubedownscaler/internal/pkg/values"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func TestDownscalePolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "get downscale policies returns empty",
			test: func(t *testing.T) {
				t.Parallel()

				clientsets := &Clientsets{}
				workloads, err := getDownscalePolicies("test-namespace", clientsets, context.Background())

				assert.NoError(t, err)
				assert.Empty(t, workloads)
			},
		},
		{
			name: "downscale policy implements workload interface",
			test: func(t *testing.T) {
				t.Parallel()

				policy := &downscalePolicy{
					Unstructured: &unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "downscaler.io/v1",
							"kind":       "DownscalePolicy",
							"metadata": map[string]interface{}{
								"name":      "test-policy",
								"namespace": "test-namespace",
								"uid":       "test-uid",
								"labels": map[string]interface{}{
									"app": "test",
								},
								"annotations": map[string]interface{}{
									"downscaler/downtime": "Mon-Fri 19:00-08:00 UTC",
								},
								"creationTimestamp": time.Now().Format(time.RFC3339),
							},
							"spec": map[string]interface{}{
								"downtime":          "Mon-Fri 19:00-08:00 UTC",
								"downscaleReplicas": 0,
							},
							"status": map[string]interface{}{
								"phase": "Active",
							},
						},
					},
				}

				// Test basic interface methods
				assert.Equal(t, "test-policy", policy.GetName())
				assert.Equal(t, "test-namespace", policy.GetNamespace())
				assert.Equal(t, types.UID("test-uid"), policy.GetUID())
				
				labels := policy.GetLabels()
				assert.Equal(t, "test", labels["app"])
				
				annotations := policy.GetAnnotations()
				assert.Equal(t, "Mon-Fri 19:00-08:00 UTC", annotations["downscaler/downtime"])

				// Test GroupVersionKind
				gvk := policy.GroupVersionKind()
				assert.Equal(t, schema.GroupVersionKind{
					Group:   "downscaler.io",
					Version: "v1",
					Kind:    "DownscalePolicy",
				}, gvk)

				// Test ScaleUp
				err := policy.ScaleUp()
				assert.NoError(t, err)
				
				// Check phase
				phase, exists, err := unstructured.NestedString(policy.Object, "status", "phase")
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.Equal(t, "Active", phase)
				
				// Check lastScaled
				lastScaled, exists, err := unstructured.NestedString(policy.Object, "status", "lastScaled")
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.NotEmpty(t, lastScaled)

				// Test ScaleDown
				replicas := values.AbsoluteReplicas(0)
				err = policy.ScaleDown(replicas)
				assert.NoError(t, err)

				// Check phase
				phase, exists, err = unstructured.NestedString(policy.Object, "status", "phase")
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.Equal(t, "Downscaled", phase)

				// Check lastScaled
				lastScaled, exists, err = unstructured.NestedString(policy.Object, "status", "lastScaled")
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.NotEmpty(t, lastScaled)

				// Check currentReplicas
				currentReplicas, exists, err := unstructured.NestedInt64(policy.Object, "status", "currentReplicas")
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.Equal(t, int64(0), currentReplicas)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}