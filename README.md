# wavefront-lambda-go

[![travis build status](https://travis-ci.com/retgits/wavefront-lambda-go.svg?branch=master)](https://travis-ci.com/retgits/wavefront-lambda-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/retgits/wavefront-lambda-go)](https://goreportcard.com/report/github.com/retgits/wavefront-lambda-go)
[![GoDoc reference](https://godoc.org/github.com/retgits/wavefront-lambda-go?status.svg)](https://godoc.org/github.com/retgits/wavefront-lambda-go)
[![Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)

> A Go wrapper for AWS Lambda so you can monitor everything from your [Wavefront](https://wavefront.com) dashboard.

The package `wavefront-lambda-go` provides a Go wrapper for AWS Lambda function so you can monitor the three pillars of observability from your [Wavefront](https://wavefront.com) dashboard.

## Installation

Using `go get`

```bash
go get github.com/retgits/wavefront-lambda-go
```

## Prerequisites

* [Go (at least Go 1.12)](https://golang.org/dl/)
* [A Wavefront API token](https://wavefront.com)

## Basic Usage

To let you your Lambda functions send metrics to Wavefront, you'll need to set two environment variables, import this module, and wrap your AWS Lambda handler function with `wfAgent.WrapHandler(handler)`. The environment variables you'll need to set are:

* `WAVEFRONT_URL`: The URL of your Wavefront instance (like, `https://myinstance.wavefront.com`).
* `WAVEFRONT_API_TOKEN`: Your Wavefront API token (see the [docs](https://docs.wavefront.com/wavefront_api.html) how to create an API token).

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	wflambda "github.com/retgits/wavefront-lambda-go" // Import this library
)

var wfAgent = wflambda.NewWavefrontAgent(&wflambda.WavefrontConfig{})

func handler() (string, error){
	return "Hello World", nil
}

func main() {
	// Wrap the handler with wfAgent.WrapHandler()
	lambda.Start(wfAgent.WrapHandler(handler))
}
```

## Configuration

The `wfAgent` variable in the previous sample can be configured using both environment variables, as well as values passed into it using the `WavefrontConfig` struct. If both WavefrontConfig and environment variables have a value for a specific setting, the environment variable takes precedence. The configuration options you can set are:

* **Enabled** (`*bool`): Enabled indicates whether metrics are sent to Wavefront. The environment variable `WAVEFRONT_ENABLED` is also used for this setting.
* **Server** (`*string`): Wavefront URL of the form `https://<INSTANCE>.wavefront.com`. The environment variable `WAVEFRONT_URL` is also used for this setting.
* **Token** (`*string`): Wavefront API token with direct data ingestion permission. The environment variable `WAVEFRONT_TOKEN` is also used for this setting.
* **BatchSize** (`*int`): Max batch of data sent per flush interval. The environment variable `WAVEFRONT_BATCH_SIZE` is also used for this setting.
* **MaxBufferSize** (`*int`): Max size of internal buffers beyond which received data is dropped. The environment variable `WAVEFRONT_MAX_BUFFER_SIZE` is also used for this setting.
* **PointTags** (`map[string]string`): Map of Key-Value pairs (strings) associated with each data point sent to Wavefront.

## Point Tags

Point tags are key-value pairs (strings) that are associated with a point. Point tags provide additional context for your data and allow you to fine-tune your queries so the output shows just what you need. 

### Standard Point Tags

The below point tags are automatically added to all data sent to Wavefront. You can add more point tags, by creating the passing in a `map[string]string` when you create a new Wavefront Agent.

| Point Tag             | Description                                                                                |
| --------------------- | ------------------------------------------------------------------------------------------ |
| LambdaArn             | ARN (**Amazon Resource Name**) of the Lambda function.                                     |
| Region                | AWS Region of the Lambda function.                                                         |
| accountId             | AWS Account ID from which the Lambda function was invoked.                                 |
| ExecutedVersion       | The version of Lambda function.                                                            |
| FunctionName          | The name of Lambda function.                                                               |
| Resource              | The name and version/alias of Lambda function. (like `DemoLambdaFunc:aliasProd`)           |
| EventSourceMappings   | AWS Event source mapping Id. (Set in case of Lambda invocation by AWS Poll-Based Services) |

### Custom Point Tags

You can add custom point tags either while instantiating the agent or inside your handler.

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	wflambda "github.com/retgits/wavefront-lambda-go" // Import this library
)

var tags = map[string]string{
	"MyTag": "NewTag",
}

var wfAgent = wflambda.NewWavefrontAgent(&wflambda.WavefrontConfig{
	PointTags: tags,
})

func handler() (string, error){
	// You also can add point tags from inside your handler function
	wfAgent.WavefrontConfig.PointTags["NewPointTag"] = "MyCustomTag"

	return "Hello World", nil
}

func main() {
	// Wrap the handler with wfAgent.WrapHandler()
	lambda.Start(wfAgent.WrapHandler(handler))
}
```

## Metrics

### Standard Metrics

The Wavefront Agent will send a set of default metrics to Wavefront when enabled. The metrics reported are:

| Metric Name                       |  Type         | Description                                                             |
| --------------------------------- | ------------- | ----------------------------------------------------------------------- |
| aws.lambda.wf.invocations.count   | Delta Counter | Count of number of Lambda function invocations aggregated at the server.|
| aws.lambda.wf.errors.count        | Delta Counter | Count of number of errors aggregated at the server.                     |
| aws.lambda.wf.coldstarts.count    | Delta Counter | Count of number of cold starts aggregated at the server.                |
| aws.lambda.wf.duration.value      | Metric        | Execution time of the Lambda handler function in milliseconds.          |
| aws.lambda.wf.mem.total           | Metric        | The total memory available to the Lambda function in megabytes.         |
| aws.lambda.wf.mem.used            | Metric        | The memory used by the Lambda function in megabytes.                    |
| aws.lambda.wf.mem.percentage      | Metric        | The percentage of memory used by the Lambda function.                   |

### Custom Metrics

You can send custom business metrics to Wavefront using the `RegisterMetric()` or `RegisterCounter()` methods. Counters are values that are aggregated at the Wavefront server (like the number of invocations and metrics are pretty much every other numerical value you want to send in.

```go
package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	wflambda "github.com/retgits/wavefront-lambda-go" // Import this library
)

var wfAgent = wflambda.NewWavefrontAgent(&wflambda.WavefrontConfig{})

func handler() (string, error){
	// Register a new Delta Counter
	wfAgent.RegisterCounter("MyAwesomeCounter", float64(1))

	// Register a new Metric
	wfAgent.RegisterCounter("MeaningOfLife", float64(42))

	return "Hello World", nil
}

func main() {
	// Wrap the handler with wfAgent.WrapHandler()
	lambda.Start(wfAgent.WrapHandler(handler))
}
```

## Contributing

[Pull requests](https://github.com/retgits/wavefront-lambda-go/pulls) are welcome. For major changes, please open [an issue](https://github.com/retgits/wavefront-lambda-go/issues) first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

See the [LICENSE](./LICENSE) file in the repository
