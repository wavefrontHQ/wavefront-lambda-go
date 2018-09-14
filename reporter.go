package wflambda

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
)

// Increment counter if report is true
func incrementCounter(counter metrics.Counter, value int64, report bool) {
	if report {
		counter.Inc(value)
	}
}

// Update gauge value if report is true
func updateGaugeFloat64(gauge metrics.GaugeFloat64, value float64, report bool) {
	if report {
		gauge.Update(value)
	}
}

// Register the standard lambda metrics.
func registerStandardLambdaMetrics() {
	// Register cold start counter.
	csCounter = metrics.NewCounter()
	csCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("coldstarts"))
	wavefront.RegisterMetric(csCounterName, csCounter, nil)

	// Register invocations counter.
	invocationsCounter = metrics.NewCounter()
	invocationsCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("invocations"))
	wavefront.RegisterMetric(invocationsCounterName, invocationsCounter, nil)

	// Register Error counter
	errCounter = metrics.NewCounter()
	errCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("errors"))
	wavefront.RegisterMetric(errCounterName, errCounter, nil)

	// Register duration gauge
	durationGauge = metrics.NewGaugeFloat64()
	wavefront.RegisterMetric(getStandardLambdaMetricName("duration"), durationGauge, nil)
}

// Method to send all metrics in the registry to wavefront.
func reportMetrics(ctx context.Context) {
	// Get standard lambda point tags
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
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
	} else {
		log.Println("ERROR :: Couldn't report points to wavefront as retrieving lambdaContext from AWS failed.")
	}
}

// Util method that returns the standard lambda metric name.
// Ex:
// getStandardLambdaMetricName("invocation", true) returns "aws.lambda.wf.invocation_event"
// getStandardLambdaMetricName("invocations", false) returns "aws.lambda.wf.invocations"

func getStandardLambdaMetricName(metric string) string {
	const metric_prefix string = "aws.lambda.wf."
	return strings.Join([]string{metric_prefix, metric}, "")
}

// Util method to validate the specified environment variables and return if standard lambda Metrics
// should be collected by the wrapper.
func getAndValidateLambdaEnvironment() bool {
	// Validate environment variables required by wavefrontLambda wrapper.
	server = os.Getenv("WAVEFRONT_URL")
	if server == "" {
		log.Panicf("Environment variable WAVEFRONT_URL is not set.")
	}
	authToken = os.Getenv("WAVEFRONT_API_TOKEN")
	if authToken == "" {
		log.Panicf("Environment variable WAVEFRONT_API_TOKEN is not set.")
	}
	reportStandardLambdaMetrics := os.Getenv("REPORT_STANDARD_METRICS")
	reportStandardMetrics := true
	if reportStandardLambdaMetrics == "False" || reportStandardLambdaMetrics == "false" {
		reportStandardMetrics = false
	}
	return reportStandardMetrics
}
