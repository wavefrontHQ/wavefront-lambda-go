# wavefront-lambda-go [![travis build status](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go.svg?branch=master)](https://travis-ci.com/wavefrontHQ/wavefront-lambda-go)

This is a Wavefront Go wrapper for AWS Lambda to enable reporting standard lambda metrics and custom app metrics directly to wavefront.

## Requirements
Go 1.x

## Installation
```
go get github.com/wavefronthq/wavefront-lambda-go
```

## Environment variables
WAVEFRONT_URL = https://\<INSTANCE>.wavefront.com  
WAVEFRONT_API_TOKEN = Wavefront API token with Direct Data Ingestion permission.  
REPORT_STANDARD_METRICS = Set to False or false to not report standard lambda metrics directly to wavefront.  

## Usage

Wrap your AWS Lambda handler function with wavefront_lambda.Wrapper(LambdaHandler).

```go
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
	"github.com/wavefronthq/wavefront-lambda-go"
)

// Lambda handler function that includes the code which will be executed when lambda is invoked.
func HandleLambdaRequest() {
	// your code
}

func main() {
	// Wrap your Lambda Handler Function with wflambda.Wrapper
	lambda.Start(wflambda.Wrapper(HandleLambdaRequest))
}
```

## Standard Lambda Metrics reported by Wavefront Lambda wrapper

The Lambda wrapper sends the following standard lambda metrics to wavefront:

| Metric Name                       |  Type              | Description                                                             |
| ----------------------------------|:------------------:| ----------------------------------------------------------------------- |
| aws.lambda.wf.invocations.count   | Delta Counter      | Count of number of lambda function invocations aggregated at the server.|
| aws.lambda.wf.errors.count        | Delta Counter      | Count of number of errors aggregated at the server.                     |
| aws.lambda.wf.coldstarts.count    | Delta Counter      | Count of number of cold starts aggregated at the server.                |
| aws.lambda.wf.duration.value      | Gauge              | Execution time of the Lambda handler function in milliseconds.          |

The Lambda wrapper adds the following point tags to all metrics sent to wavefront:

| Point Tag             | Description                                                                   |
| --------------------- | ----------------------------------------------------------------------------- |
| LambdaArn             | ARN(Amazon Resource Name) of the Lambda function.                             |
| Region                | AWS Region of the Lambda function.                                            |
| accountId             | AWS Account ID from which the Lambda function was invoked.                    |
| ExecutedVersion       | The version of Lambda function.                                               |
| FunctionName          | The name of Lambda function.                                                  |
| Resource              | The name and version/alias of Lambda function. (Ex: DemoLambdaFunc:aliasProd) |
| EventSourceMappings   | AWS Event source mapping Id. (Set in case of Lambda invocation by AWS Poll-Based Services)|

## Custom Lambda Metrics

The wavefront Go lambda wrapper reports custom business metrics via API's provided by the [go-metrics-wavefront client] (https://github.com/wavefrontHQ/go-metrics-wavefront).  
Please refer to the below code sample which shows how you can send custom business metrics to wavefront from your lambda function.

### Code Sample

```go
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
	"github.com/wavefronthq/wavefront-lambda-go"
)

// Lambda handler function that includes the code which will be executed when lambda is invoked.
func HandleLambdaRequest() {
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
	//Wrapping with wflambda.Wrapper
	lambda.Start(wflambda.Wrapper(HandleLambdaRequest))
}
```

Note: Having the same metric name for any two types of metrics will result in only one time series at the server and thus cause collisions.
In general, all metric names should be different. In case you have metrics that you want to track as both a Counter and Delta Counter, consider adding a relevant suffix to one of the metrics to differentiate one metric name from another.
