package wflambda

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	wa := NewWavefrontAgent(&WavefrontConfig{})
	assert.NotNil(wa)

	handler := func(ctx context.Context, payload interface{}) (interface{}, error) { return nil, nil }
	wrapper := newHandler(handler)
	hw := NewHandlerWrapper(handler, wa)

	assert.IsType(hw.originalHandler, handler)
	assert.IsType(hw.wrappedHandler, wrapper)
}
