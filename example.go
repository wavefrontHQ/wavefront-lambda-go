package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

// Lambda handler function that includes the code which will be executed when lambda is invoked.
func HandleLambdaRequest() {
	appTags := map[string]string{
		"source": "ExampleLambdaFunction",
		"key2":   "val1",
		"key1":   "val2",
		"key0":   "val0",
		"key4":   "val4",
		"key3":   "val3",
	}
	// Register Counter with desired tags.
	customRawCounter := metrics.NewCounter()
	wavefront.RegisterMetric("counter", customRawCounter, appTags)
	customRawCounter.Inc(1)
	// Register Delta Counter with desired tags.
	customDeltaCounter := metrics.NewCounter()
	deltaCounterName := wavefront.DeltaCounterName("deltaCounter")
	wavefront.RegisterMetric(deltaCounterName, customDeltaCounter, appTags)
	customDeltaCounter.Inc(1)
	// Register Gauge with desired tags.
	gaugeValue = metrics.NewGauge()
	wavefront.RegisterMetric("gaugeValue", gaugeValue, appTags)
	gaugeValue.Update(5.5)
}

func main() {
	//Wrap with wflambda.Wrapper
	lambda.Start(wflambda.Wrapper(HandleLambdaRequest))
}
