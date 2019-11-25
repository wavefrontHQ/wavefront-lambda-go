package wflambda

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgent(t *testing.T) {
	assert := assert.New(t)

	wc := &WavefrontConfig{}
	assert.NotNil(wc)

	wa := NewWavefrontAgent(wc)
	assert.NotNil(wa)
	assert.Nil(wa.WavefrontConfig.Server)

	wa = NewWavefrontAgent(&WavefrontConfig{Enabled: stringToBool("true")})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.Enabled, stringToBool("true"))

	os.Setenv("WAVEFRONT_ENABLED", "false")
	wa = NewWavefrontAgent(&WavefrontConfig{})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.Enabled, stringToBool("false"))

	str := "https://instance.wavefront.com"
	wa = NewWavefrontAgent(&WavefrontConfig{Server: &str})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.Server, &str)

	str = "my-api-token"
	wa = NewWavefrontAgent(&WavefrontConfig{Token: &str})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.Token, &str)

	i := 1
	wa = NewWavefrontAgent(&WavefrontConfig{BatchSize: &i})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.BatchSize, &i)

	os.Setenv("WAVEFRONT_BATCH_SIZE", "12")
	i = 12
	wa = NewWavefrontAgent(&WavefrontConfig{})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.BatchSize, &i)

	i = 10
	wa = NewWavefrontAgent(&WavefrontConfig{MaxBufferSize: &i})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.MaxBufferSize, &i)

	os.Setenv("WAVEFRONT_MAX_BUFFER_SIZE", "120")
	i = 120
	wa = NewWavefrontAgent(&WavefrontConfig{})
	assert.NotNil(wa)
	assert.Equal(wa.WavefrontConfig.MaxBufferSize, &i)

	wa.RegisterCounter("counter1", 1)
	assert.Equal(wa.counters["counter1"], float64(1))
	assert.Equal(len(wa.counters), 1)

	wa.RegisterMetric("metric1", 1)
	assert.Equal(wa.metrics["metric1"], float64(1))
	assert.Equal(len(wa.metrics), 1)

	wa = NewWavefrontAgent(&WavefrontConfig{Enabled: stringToBool("false")})
	assert.NotNil(wa)
	iface := wa.WrapHandler("bla")
	assert.Equal(iface.(string), "bla")
}
