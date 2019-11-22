# wavefront-lambda-go

[![travis build status](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go.svg?branch=master)](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/wavefrontHQ/wavefront-lambda-go)](https://goreportcard.com/report/github.com/wavefrontHQ/wavefront-lambda-go)

A Go wrapper for AWS Lambda so you can monitor everything from your [Wavefront](https://wavefront.com) dashboard

## Installation

Using `go get`

```bash
go get github.com/wavefronthq/wavefront-lambda-go
```

## Basic Usage

To connect your Lambda functions to Wavefront, you'll need to set three environment variables, import this module, and wrap your AWS Lambda handler function with `wflambda.Wrapper(handler)`. The environment variables you'll need to set are:

* `WAVEFRONT_URL`: The URL of your Wavefront instance (like, `https://myinstance.wavefront.com`).
* `WAVEFRONT_API_TOKEN`: Your Wavefront API token (see the [docs](https://docs.wavefront.com/wavefront_api.html) how to create an API token).
* `WAVEFRONT_ENABLED`: A boolean which determines if data is sent to Wavefront or not (this value defaults to `true` and can be omitted).

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	wflambda "github.com/wavefronthq/wavefront-lambda-go" // Import this library
)

func handler() (string, error){
	return "Hello World", nil
}

func main() {
	// Wrap the handler with wflambda.Wrapper()
	lambda.Start(wflambda.Wrapper(handler))
}
```

## Standard Point Tags

Point tags are key-value pairs (strings) that are associated with a point. Point tags provide additional context for your data and allow you to fine-tune your queries so the output shows just what you need. The below point tags are sent to Wavefront for each metric.

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

You can send custom business metrics to Wavefront using the [wavefront-sdk-go](https://github.com/wavefronthq/wavefront-sdk-go) client. The below code reports a _counter_ and a _delta counter_. All metric names should be unique. If you have metrics that you want to track as both _counter_ and _delta counter_, you'll have to add a suffix to one of the metrics. Having the same metric name for any two types of metrics will result in only one time series at the server and thus cause collisions.

The code below imported this module and wrapped the _handler_ function argument in _main_ with `wflambda.Wrapper(handler)`. During each execution the metrics are collected and sent to Wavefront using a `DirectSender`.

```go
package main

import (
	"time"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rcrowley/go-metrics"
	wavefront "github.com/wavefronthq/wavefront-sdk-go/senders"
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

	dc := &wavefront.DirectConfiguration{
		Server:               server, // Wavefront URL of the form https://<INSTANCE>.wavefront.com
		Token:                authToken, // Wavefront API token with direct data ingestion permission
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	sender, err := wavefront.NewDirectSender(dc)
	if err != nil {
		log.Printf("ERROR :: %s", err.Error())
	}

	// Send metrics using a Direct Wavefront Sender. The <source> will allow you to filter data in the Wavefront UI
	// A DeltaCounter is aggragated at the Wavefront server
	err = sender.SendDeltaCounter("the.answer", 42, "<source>", appTags)
	if err != nil {
		log.Printf("ERROR :: %s", err.Error())
	}

	err = sender.SendMetric("pizzas.needed", 2, time.Now().Unix(), "<source>", appTags)
	if err != nil {
		log.Printf("ERROR :: %s", err.Error())
	}
}

func main() {
	lambda.Start(wflambda.Wrapper(handler))
}
```
