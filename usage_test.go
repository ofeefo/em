package em

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UninitializedSubMetrics struct {
	CounterB Counter[int64] `id:"counter_b"`
}

type UninitializedMetrics struct {
	Counter Counter[int64] `id:"counter"`
	Sub     UninitializedSubMetrics
}

func TestInitialized(t *testing.T) {
	m, err := Init[UninitializedMetrics]()
	require.NoError(t, err)

	assert.NotPanics(t, func() {
		m.Counter.Inc()
	})
	assert.NotPanics(t, func() {
		m.Sub.CounterB.Inc()
	})
}

func TestUninitialized(t *testing.T) {
	m := UninitializedMetrics{}
	assert.NotPanics(t, func() {
		m.Counter.Inc()
	})
	assert.NotPanics(t, func() {
		m.Sub.CounterB.Inc()
	})
}
