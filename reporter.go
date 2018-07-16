package wflambda

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
	"log"
	"os"
	"strings"
	"time"
)

func incrementCounter(counter metrics.Counter, value int64, report bool) {
  if report {
      counter.Inc(value)
  }
}

func updateGauge(gauge metrics.Gauge, value int64, report bool) {
  if report {
      gauge.Update(value)
  }
}

func registerStandardLambdaMetrics() {
	// Register cold start counter.
	csCounter = metrics.NewCounter()
	csEventCounter = metrics.NewCounter()
	coldStartsCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("coldstarts", false))
	wavefront.RegisterMetric(coldStartsCounterName, csCounter, nil)
	wavefront.RegisterMetric(getStandardLambdaMetricName("coldstart", true), csEventCounter, nil)

	// Register invocations counter.
	invocationsCounter = metrics.NewCounter()
	invocationEventCounter = metrics.NewCounter()
	invocationsCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("invocations", false))
	wavefront.RegisterMetric(invocationsCounterName, invocationsCounter, nil)
	wavefront.RegisterMetric(getStandardLambdaMetricName("invocation", true), invocationEventCounter, nil)

	// Register Error counter
	errCounter = metrics.NewCounter()
	errEventCounter = metrics.NewCounter()
	errorsCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("errors", false))
	wavefront.RegisterMetric(errorsCounterName, errCounter, nil)
	wavefront.RegisterMetric(getStandardLambdaMetricName("error", true), errEventCounter, nil)

	// Register duration gauge
	dGauge = metrics.NewGauge()
	wavefront.RegisterMetric(getStandardLambdaMetricName("duration", false), dGauge, nil)
}

// Method to send all metrics in the registry to wavefront.
func reportMetrics(ctx context.Context) {
	// Get standard lambda point tags
	lc, _ := lambdacontext.FromContext(ctx)
	invokedFunctionArn := lc.InvokedFunctionArn
	splitArn := strings.Split(invokedFunctionArn, ":")

  // Expected formats for Lambda ARN are:
  // https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-lambda
	hostTags := map[string]string{
		"LambdaArn":       invokedFunctionArn,
		"source":          lambdacontext.FunctionName,
		"FunctionName":    lambdacontext.FunctionName,
		"ExecutedVersion": lambdacontext.FunctionVersion,
		"Region":          splitArn[3],
		"accountId":       splitArn[4],
	}
	if splitArn[5] == "function" {
		hostTags["Resource"] = splitArn[6]
		if len(splitArn) == 8 {
			hostTags["Resource"] += ":" + splitArn[7]
		}
	} else if splitArn[5] == "event-source-mappings" {
		hostTags["EventSourceMappings"] = splitArn[6]
	}

	err := wavefront.WavefrontOnce(wavefront.WavefrontConfig{
		DirectReporter: wavefront.NewDirectReporter(server, authToken),
		Registry:       metrics.DefaultRegistry,
		DurationUnit:   time.Nanosecond,
		Prefix:         "",
		HostTags:       hostTags,
	})
	if err != nil {
		log.Println("ERROR :: ", err)
	}
}

// Util method to returns the standard lambda metric name.
func getStandardLambdaMetricName(metric string, isEvent bool) string {
	const metric_prefix string = "aws.lambda.wf."
	const metric_event_suffix string = "_event"
	if isEvent {
		return strings.Join([]string{metric_prefix, metric, metric_event_suffix}, "")
	}
	return strings.Join([]string{metric_prefix, metric}, "")
}

// Util method to validate the specified environmental variables and return if standard lambda Metrics
// should be collected by the wrapper.
func getAndValidateLambdaEnvironment() bool {
	// Validate environmental variables required by wavefrontLambda wrapper.
	server = os.Getenv("WAVEFRONT_URL")
	if server == "" {
		log.Panicf("Environmental variable WAVEFRONT_URL is not set.")
	}
	authToken = os.Getenv("WAVEFRONT_API_TOKEN")
	if authToken == "" {
		log.Panicf("Environmental variable WAVEFRONT_API_TOKEN is not set.")
	}
	reportStandardLambdaMetrics := os.Getenv("IS_REPORT_STANDARD_METRICS")
	reportStandardMetrics := true
	if reportStandardLambdaMetrics == "False" || reportStandardLambdaMetrics == "false" {
		reportStandardMetrics = false
	}
	return reportStandardMetrics
}
