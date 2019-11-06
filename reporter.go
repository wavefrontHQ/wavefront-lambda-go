package wflambda

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// incrementCounter increments the counter by the given value if report is true
func incrementCounter(counter *float64, value int64, report bool) {
	if report {
		*counter += float64(value)
	}
}

// updateGaugeFloat64 increments the counter by the given value if report is true
func updateGaugeFloat64(gauge *float64, value float64, report bool) {
	if report {
		*gauge += float64(value)
	}
}

// reportMetrics sends the collected metrics in the registry to Wavefront. With each metric,
// the point tags listed in the README are sent by the reporter.
func reportMetrics(ctx context.Context) {
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

		reportTime := time.Now().Unix()

		err := sender.SendMetric("aws.lambda.wf.coldstarts", *csCounter, reportTime, lambdacontext.FunctionName, hostTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}

		err = sender.SendMetric("aws.lambda.wf.invocations", *invocationsCounter, reportTime, lambdacontext.FunctionName, hostTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}

		err = sender.SendMetric("aws.lambda.wf.errors", *errCounter, reportTime, lambdacontext.FunctionName, hostTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}

		err = sender.SendMetric("aws.lambda.wf.duration", *durationGauge, reportTime, lambdacontext.FunctionName, hostTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}

		sender.Flush()
		sender.Close()
	} else {
		log.Println("ERROR :: Couldn't report points to wavefront as retrieving lambdaContext from AWS failed.")
	}
}

// getAndValidateLambdaEnvironment validates whether the required environment variables WAVEFRONT_URL and
// WAVEFRONT_API_TOKEN have been set. If they are not set, the function will panic. The function also checks
// whether the environment variable REPORT_STANDARD_METRICS has been set to false (it will default to true).
// to determine if the standard metrics should be reported.
func getAndValidateLambdaEnvironment() bool {
	server = os.Getenv("WAVEFRONT_URL")
	if server == "" {
		log.Panicf("Environment variable WAVEFRONT_URL is not set.")
	}

	authToken = os.Getenv("WAVEFRONT_API_TOKEN")
	if authToken == "" {
		log.Panicf("Environment variable WAVEFRONT_API_TOKEN is not set.")
	}

	reportEnabled := os.Getenv("WAVEFRONT_ENABLED")
	if reportEnabled == "False" || reportEnabled == "false" {
		enabled = false
	}

	reportStandardLambdaMetrics := os.Getenv("REPORT_STANDARD_METRICS")
	reportStandardMetrics := true
	if reportStandardLambdaMetrics == "False" || reportStandardLambdaMetrics == "false" {
		reportStandardMetrics = false
	}
	return reportStandardMetrics
}
