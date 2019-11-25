package wflambda

import (
	"log"
	"os"

	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

var (
	// Is this a cold start or not.
	coldStart = true
	// Count the number of cold starts.
	csCounter = counter{}
	// Count the number of invocations.
	invocationsCounter = counter{}
	// Count the number of errors.
	errCounter = counter{}
)

// WavefrontConfig configures the direct ingestion sender to Wavefront.
type WavefrontConfig struct {
	// Enabled indicates whether metrics are sent to Wavefront
	Enabled *bool
	// Wavefront URL of the form https://<INSTANCE>.wavefront.com.
	Server *string
	// Wavefront API token with direct data ingestion permission.
	Token *string
	// Max batch of data sent per flush interval.
	BatchSize *int
	// Max size of internal buffers beyond which received data is dropped.
	MaxBufferSize *int
	// Map of Key-Value pairs (strings) associated with each data point sent to Wavefront.
	PointTags map[string]string
}

// WavefrontAgent is the agent instance that communicates with Wavefront.
type WavefrontAgent struct {
	*WavefrontConfig
	metrics  map[string]float64
	counters map[string]float64
	sender   wavefront.Sender
}

var (
	// Default value whether the agent is enabled or not.
	defaultEnabled = true
	// Default value for batch of data sent per flush interval.
	defaultBatchSize = 10000
	// Default size of internal buffers beyond which received data is dropped
	defaultMaxBufferSize = 50000
	// Default interval (in seconds) at which to flush data to Wavefront.
	defaultFlushIntervalSeconds = 1
)

// NewWavefrontAgent returns a new agent.
func NewWavefrontAgent(w *WavefrontConfig) *WavefrontAgent {
	// Create a new instance of the WavefrontAgent.
	wfAgent := &WavefrontAgent{
		metrics:  make(map[string]float64),
		counters: make(map[string]float64),
	}

	// Create an empty map of point tags if no tags exist yet.
	if w.PointTags == nil {
		w.PointTags = make(map[string]string)
	}

	// Create the configuration to connect to Wavefront. Details are gathered from both
	// the WavefrontConfig and the environment variables. If both WavefrontConfig and
	// environment variables have a value for a specific setting, the environment variable
	// takes precedence.
	enabled := &defaultEnabled
	envEnabled := os.Getenv("WAVEFRONT_ENABLED")
	if w.Enabled != nil {
		enabled = w.Enabled
	}
	if envEnabled != "" {
		enabled = stringToBool(envEnabled)
	}

	envServer := os.Getenv("WAVEFRONT_URL")
	server := &envServer
	if w.Server != nil && len(envServer) == 0 {
		server = w.Server
	}

	envToken := os.Getenv("WAVEFRONT_API_TOKEN")
	token := &envToken
	if w.Token != nil && len(envToken) == 0 {
		token = w.Token
	}

	batchSize := &defaultBatchSize
	envBatchSize := os.Getenv("WAVEFRONT_BATCH_SIZE")
	if w.BatchSize != nil {
		batchSize = w.BatchSize
	}
	if envBatchSize != "" {
		batchSizeInt, err := stringToInt(envBatchSize)
		if err == nil {
			w.BatchSize = batchSizeInt
			batchSize = batchSizeInt
		}
	}

	maxBufferSize := &defaultMaxBufferSize
	envMaxBufferSize := os.Getenv("WAVEFRONT_MAX_BUFFER_SIZE")
	if w.MaxBufferSize != nil {
		maxBufferSize = w.MaxBufferSize
	}
	if envMaxBufferSize != "" {
		maxBufferSizeInt, err := stringToInt(envMaxBufferSize)
		if err == nil {
			w.MaxBufferSize = maxBufferSizeInt
			maxBufferSize = maxBufferSizeInt
		}
	}

	dc := &wavefront.DirectConfiguration{
		Server:               *server,
		Token:                *token,
		BatchSize:            *batchSize,
		MaxBufferSize:        *maxBufferSize,
		FlushIntervalSeconds: 1,
	}

	sender, err := wavefront.NewDirectSender(dc)
	if err != nil {
		log.Printf("ERROR :: %s", err.Error())
	}

	wfAgent.sender = sender
	wfAgent.WavefrontConfig = w
	wfAgent.WavefrontConfig.Enabled = enabled

	return wfAgent
}

// WrapHandler wraps the handler
func (wa *WavefrontAgent) WrapHandler(handler interface{}) interface{} {
	if !*wa.Enabled {
		return handler
	}

	return wrapHandler(handler, wa)
}

func (wa *WavefrontAgent) RegisterMetric(name string, value float64) {
	wa.metrics[name] = value
}

func (wa *WavefrontAgent) RegisterCounter(name string, value float64) {
	wa.counters[name] = value
}
