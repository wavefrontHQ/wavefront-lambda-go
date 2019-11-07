package wflambda

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapper(t *testing.T) {
	assert := assert.New(t)

	os.Setenv("WAVEFRONT_URL", "https://demo.wavefront.com")
	os.Setenv("WAVEFRONT_API_TOKEN", "demo-api-token")

	handler := func(ctx context.Context, payload interface{}) (interface{}, error) { return nil, nil }
	iface := Wrapper(handler)
	assert.NotNil(iface)
}
