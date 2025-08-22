package scalable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownscalePolicyIntegration(t *testing.T) {
	t.Parallel()

	t.Run("handles missing dynamic client gracefully", func(t *testing.T) {
		t.Parallel()

		// Test with nil dynamic client
		clientsets := &Clientsets{
			Dynamic: nil,
		}

		workloads, err := getDownscalePolicies("test-namespace", clientsets, context.Background())
		assert.NoError(t, err)
		assert.Empty(t, workloads)
	})

	t.Run("dynamic client availability check", func(t *testing.T) {
		t.Parallel()

		// Test that the function exists and compiles
		clientsets := &Clientsets{}

		workloads, err := getDownscalePolicies("test-namespace", clientsets, context.Background())
		assert.NoError(t, err)
		assert.Empty(t, workloads) // Should be empty when Dynamic client is nil
	})
}