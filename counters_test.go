package wflambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCounters(t *testing.T) {
	assert := assert.New(t)

	ctr := counter{}
	assert.NotNil(ctr)
	ctr.Increment(2)
	assert.Equal(ctr.val, float64(2))
	ctr.Increment(-3)
	assert.Equal(ctr.val, float64(-1))
}
