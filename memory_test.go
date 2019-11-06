package wflambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemory(t *testing.T) {
	assert := assert.New(t)
	stats := getMemoryStats()
	assert.NotNil(stats)
	assert.NotZero(stats.Total)
	assert.NotZero(stats.Used)
	assert.NotZero(stats.UsedPercentage)
}
