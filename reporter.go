package wflambda

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
)

// updateMetric increments the value of a given metric if report is true
func updateMetric(metric *float64, inc int64, report bool) {
	if report {
		*metric += float64(inc)
	}
}

// reportMetrics sends the collected metrics to Wavefront. With each metric,
// the point tags listed in the README are sent as well. Sending metrics relies on the
// wavefront-go-sdk project. This function is only called when WAVEFRONT_ENABLED is
// set to true
func reportMetrics(ctx context.Context) {
	lc, ok := lambdacontext.FromContext(ctx)
	if ok {
		dc := &wavefront.DirectConfiguration{
			Server:               server,
			Token:                authToken,
			BatchSize:            10000,
			MaxBufferSize:        50000,
			FlushIntervalSeconds: 1,
		}

		sender, err := wavefront.NewDirectSender(dc)
		if err != nil {
			log.Printf("ERROR :: %s", err.Error())
			return
		}

		// The reportTime is used for all metrics sent to Wavefront, to ensure all of them
		// have the same timestamp
		reportTime := time.Now().Unix()

		// Get the lambdaContext to derive information for the point tags sent to Wavefront
		// The InvokedFunctionArn contains data on region and account. The expected formats
		// for Lambda ARN are available in the AWS docs:
		// https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-lambda
		invokedFunctionArn := lc.InvokedFunctionArn
		splitArn := strings.Split(invokedFunctionArn, ":")

		pointTags := map[string]string{
			"LambdaArn":       invokedFunctionArn,
			"source":          lambdacontext.FunctionName,
			"FunctionName":    lambdacontext.FunctionName,
			"ExecutedVersion": lambdacontext.FunctionVersion,
			"Region":          splitArn[3],
			"accountId":       splitArn[4],
		}

		if splitArn[5] == "function" {
			pointTags["Resource"] = splitArn[6]
			if len(splitArn) == 8 {
				pointTags["Resource"] += ":" + splitArn[7]
			}
		} else if splitArn[5] == "event-source-mappings" {
			pointTags["EventSourceMappings"] = splitArn[6]
		}

		// Send metrics using a Direct Wavefront Sender
		err = sender.SendDeltaCounter("aws.lambda.wf.coldstarts", *csCounter, lambdacontext.FunctionName, pointTags)
		if err != nil {
			log.Printf("ERROR :: %s", err.Error())
		}

		err = sender.SendDeltaCounter("aws.lambda.wf.invocations", *invocationsCounter, lambdacontext.FunctionName, pointTags)
		if err != nil {
			log.Printf("ERROR :: %s", err.Error())
		}

		err = sender.SendDeltaCounter("aws.lambda.wf.errors", *errCounter, lambdacontext.FunctionName, pointTags)
		if err != nil {
			log.Printf("ERROR :: %s", err.Error())
		}

		err = sender.SendMetric("aws.lambda.wf.duration", *durationGauge, reportTime, lambdacontext.FunctionName, pointTags)
		if err != nil {
			log.Printf("ERROR :: %s", err.Error())
		}

		// Make sure all metrics are actually sent to Wavefront and close the sender
		sender.Flush()
		sender.Close()
	} else {
		log.Println("ERROR :: Couldn't report points to wavefront as retrieving lambdaContext from AWS failed.")
	}
}

// getAndValidateLambdaEnvironment validates whether the required environment variables WAVEFRONT_URL and
// WAVEFRONT_API_TOKEN have been set if WAVEFRONT_ENABLED is set to true. The function also checks
// whether the environment variables REPORT_STANDARD_METRICS and WAVEFRONT_ENABLED have been set to false.
// Both environment variables will default to `true`. REPORT_STANDARD_METRICS determines whether the standard
// metrics should be reported and WAVEFRONT_ENABLED determines if any data should be sent to Wavefront at all.
// If the WAVEFRONT_ENABLED flag is set to false, all other checks are skipped because no metrics will be
// emitted.
func getAndValidateLambdaEnvironment() bool {
	reportEnabled := os.Getenv("WAVEFRONT_ENABLED")
	if strings.EqualFold(reportEnabled, "false") {
		enabled = false
		return false
	}

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
	if strings.EqualFold(reportStandardLambdaMetrics, "false") {
		reportStandardMetrics = false
	}
	return reportStandardMetrics
}
