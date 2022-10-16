package service_test

import (
	"loader/pkg/service"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestIsAbnormalSequenceState(t *testing.T) {
	t.Run("return 1", func(t *testing.T) {
		result := service.IsAbnormalSequenceState(5, 1, 1)
		assert.Equal(t, float64(1), result)
	})
	t.Run("return 1", func(t *testing.T) {
		result := service.IsAbnormalSequenceState(1, 1, 5)
		assert.Equal(t, float64(1), result)
	})
	t.Run("return 0", func(t *testing.T) {
		result := service.IsAbnormalSequenceState(1, 5, 1)
		assert.Equal(t, float64(0), result)
	})
}
