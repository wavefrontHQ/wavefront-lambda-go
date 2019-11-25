package wflambda

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAndValidateLambdaEnvironment(t *testing.T) {
	assert := assert.New(t)

	assert.PanicsWithValue("Environment variable WAVEFRONT_URL is not set.", func() { getAndValidateLambdaEnvironment() })

	os.Setenv("WAVEFRONT_URL", "https://demo.wavefront.com")
	assert.PanicsWithValue("Environment variable WAVEFRONT_API_TOKEN is not set.", func() { getAndValidateLambdaEnvironment() })

	os.Setenv("WAVEFRONT_API_TOKEN", "demo-api-token")
	assert.NotPanics(func() { getAndValidateLambdaEnvironment() })

	b := getAndValidateLambdaEnvironment()
	assert.Equal(true, b)

	os.Setenv("REPORT_STANDARD_METRICS", "False")
	b = getAndValidateLambdaEnvironment()
	assert.Equal(false, b)

	os.Setenv("REPORT_STANDARD_METRICS", "FaLsE")
	b = getAndValidateLambdaEnvironment()
	assert.Equal(false, b)

	b = getAndValidateLambdaEnvironment()
	assert.Equal(true, enabled)

	os.Setenv("WAVEFRONT_ENABLED", "False")
	b = getAndValidateLambdaEnvironment()
	assert.Equal(false, b)

	os.Setenv("WAVEFRONT_ENABLED", "FaLsE")
	b = getAndValidateLambdaEnvironment()
	assert.Equal(false, b)
}

func TestUpdateCounter(t *testing.T) {
	assert := assert.New(t)

	counter := Float()
	assert.Equal(*counter, float64(0))

	updateMetric(counter, 2, true)
	assert.Equal(*counter, float64(2))

	updateMetric(counter, -3, true)
	assert.Equal(*counter, float64(-1))
}
