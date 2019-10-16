# wavefront-lambda-go

[![travis build status](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go.svg?branch=master)](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wavefrontHQ/wavefront-lambda-go)](https://goreportcard.com/report/github.com/wavefrontHQ/wavefront-lambda-go)

A Go wrapper for AWS Lambda so you can monitor everything from your [Wavefront](https://wavefront.com) dashboard

## Installation

Using `go get`

```bash
go get github.com/wavefronthq/wavefront-lambda-go
```

## Usage

To connect your Lambda functions to Wavefront, you'll need to set two environment variables, import this module, and wrap your AWS Lambda handler function with `wavefront_lambda.Wrapper(LambdaHandler)`. The environment variables you'll need to set are:

* `WAVEFRONT_URL`: The URL of your Wavefront instance (like, `https://myinstance.wavefront.com`).
* `WAVEFRONT_API_TOKEN`: Your Wavefront API token (see the [docs](https://docs.wavefront.com/wavefront_api.html) how to create an API token).

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	wflambda "github.com/wavefronthq/wavefront-lambda-go"
)

func handler() (string, error){
	return "Hello World", nil
}

func main() {
	lambda.Start(wflambda.Wrapper(HandleLambdaRequest))
}
```

The wrapper will send the below point tags to Wavefront

| Point Tag             | Description                                                                                |
| --------------------- | ------------------------------------------------------------------------------------------ |
| LambdaArn             | ARN (**Amazon Resource Name**) of the Lambda function.                                     |
| Region                | AWS Region of the Lambda function.                                                         |
| accountId             | AWS Account ID from which the Lambda function was invoked.                                 |
| ExecutedVersion       | The version of Lambda function.                                                            |
| FunctionName          | The name of Lambda function.                                                               |
| Resource              | The name and version/alias of Lambda function. (like `DemoLambdaFunc:aliasProd`)           |
| EventSourceMappings   | AWS Event source mapping Id. (Set in case of Lambda invocation by AWS Poll-Based Services) |

## Standard Metrics

Based on the environment variable `REPORT_STANDARD_METRICS` the wrapper will send standard metrics to Wavefront. Set the variable to to `false` to not send the standard metrics. When the variable is not set, it will use the default value `true`.

| Metric Name                       |  Type         | Description                                                             |
| --------------------------------- | ------------- | ----------------------------------------------------------------------- |
| aws.lambda.wf.invocations.count   | Delta Counter | Count of number of lambda function invocations aggregated at the server.|
| aws.lambda.wf.errors.count        | Delta Counter | Count of number of errors aggregated at the server.                     |
| aws.lambda.wf.coldstarts.count    | Delta Counter | Count of number of cold starts aggregated at the server.                |
| aws.lambda.wf.duration.value      | Gauge         | Execution time of the Lambda handler function in milliseconds.          |

## Custom Metrics

You can send custom business metrics to Wavefront using the [go-metrics-wavefront](https://github.com/wavefrontHQ/go-metrics-wavefront) client. The below code reports a _counter_, a _delta counter_, and two _gauges_. All metric names should be unique. If you have metrics that you want to track as both _counter_ and _delta counter_, you'll have to add a suffix to one of the metrics. Having the same metric name for any two types of metrics will result in only one time series at the server and thus cause collisions.

```go
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rcrowley/go-metrics"
	wavefront "github.com/wavefronthq/go-metrics-wavefront"
	wflambda "github.com/wavefronthq/wavefront-lambda-go"
)

func handler() {
	// Point Tags
	appTags := map[string]string{
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
	gaugeValue := metrics.NewGauge()
	wavefront.RegisterMetric("gaugeValue", gaugeValue, appTags)
	gaugeValue.Update(551)

	// Register Float Gauge with desired tags.
	gaugeFloatValue := metrics.NewGaugeFloat64()
	wavefront.RegisterMetric("gaugeFloatValue", gaugeFloatValue, appTags)
	gaugeFloatValue.Update(551.4)
}

func main() {
	lambda.Start(wflambda.Wrapper(HandleLambdaRequest))
}
```
