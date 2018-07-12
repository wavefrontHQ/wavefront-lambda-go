package wflambda

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/rcrowley/go-metrics"
	"github.com/wavefronthq/go-metrics-wavefront"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

type lambdaHandlerFunction func(context.Context, json.RawMessage) (interface{}, error)

var (
	server                  string
	authToken               string
	isReportStandardMetrics bool
	csCounter               metrics.Counter
	csEventCounter          metrics.Counter
	invocationsCounter      metrics.Counter
	invocationEventCounter  metrics.Counter
	eCounter                metrics.Counter
	eEventCounter           metrics.Counter
	dGauge                  metrics.Gauge
	isColdStart             = true
)

func Wrapper(lambdaHandler interface{}) interface{} {
	handlerType := reflect.TypeOf(lambdaHandler)
	handlerValue := reflect.ValueOf(lambdaHandler)
	// Validate lambdaHandler Kind.
	if handlerType.Kind() != reflect.Func {
		return lambdaErrorHandler(fmt.Errorf("Expected lambda handler function type : %s , but Actual : %s", handlerType.Kind(), reflect.Func))
	}

	// Validate lambdaHandler Input Arguments.
	isLambdaHandlerTakesContext := false
	switch numberTIn := handlerType.NumIn(); numberTIn {
	case 0: // do nothing
	case 1:
		cxtType := reflect.TypeOf((*context.Context)(nil)).Elem()
		inputArgType := handlerType.In(0)
		isLambdaHandlerTakesContext = inputArgType.Implements(cxtType)
	case 2:
		cxtType := reflect.TypeOf((*context.Context)(nil)).Elem()
		inputArgType := handlerType.In(0)
		isLambdaHandlerTakesContext = inputArgType.Implements(cxtType)
		if !isLambdaHandlerTakesContext {
			return lambdaErrorHandler(fmt.Errorf("There lambda handler function takes 2 arguments. Expected First Argument of to be of kind : %s , but Actual : %s", cxtType, inputArgType))
		}
	default:
		return lambdaErrorHandler(fmt.Errorf("The lambda handler function takes incorrect number of arguments. It takes %d input arguments", handlerType.NumIn()))
	}

	isReportStandardMetrics = getAndValidateLambdaEnvironment()
	if !isReportStandardMetrics {
		return func(ctx context.Context, payload json.RawMessage) (response interface{}, lambdaHandlerError error) {
			defer func() {
				reportMetrics(ctx)
				metrics.DefaultRegistry.UnregisterAll()
			}()
			var args []reflect.Value
			if isLambdaHandlerTakesContext {
				args = append(args, reflect.ValueOf(ctx))
			}
			if (handlerType.NumIn() == 1 && !isLambdaHandlerTakesContext) || handlerType.NumIn() == 2 {
				inputParamType := handlerType.In(handlerType.NumIn() - 1)
				paramValue := reflect.New(inputParamType)
				if e := json.Unmarshal(payload, paramValue.Interface()); e != nil {
					return nil, e
				}
				elem := paramValue.Elem()
				args = append(args, elem)
			}
			lambdaResponse := handlerValue.Call(args)
			if len(lambdaResponse) == 0 {
				return nil, nil
			}
			var err error
			if len(lambdaResponse) > 0 {
				// The last value must always implement error.
				if e, ok := lambdaResponse[len(lambdaResponse)-1].Interface().(error); ok {
					err = e
				}
			}
			var val interface{}
			if len(lambdaResponse) == 2 {
				// In case lambda handler returns 2 arguments(i.e. Maximum allowed return arguments), first
				// argument represents a valid non-error value compatible with the encoding/json standard library.
				val = lambdaResponse[0].Interface()
			}
			return val, err
		}
	}

	//Returns a wavefrontLambda wrapper function which gathers standard lambda metrics.
	return func(ctx context.Context, payload json.RawMessage) (response interface{}, lambdaHandlerError error) {
		defer func() {
			var err interface{}
			// Increment error count if there is a panic or non nil error is returned
			// by users lambda handler function.
			if e := recover(); e != nil {
				err = e
				// Set error counters
				eCounter.Inc(1)
				eEventCounter.Inc(1)
			} else if lambdaHandlerError != nil {
				// Set error counters
				eCounter.Inc(1)
				eEventCounter.Inc(1)
			}
			reportMetrics(ctx)
			metrics.DefaultRegistry.UnregisterAll()
			if err != nil {
				panic(err)
			}
		}()

		var args []reflect.Value
		if isLambdaHandlerTakesContext {
			args = append(args, reflect.ValueOf(ctx))
		}
		if (handlerType.NumIn() == 1 && !isLambdaHandlerTakesContext) || handlerType.NumIn() == 2 {
			inputParamType := handlerType.In(handlerType.NumIn() - 1)
			paramValue := reflect.New(inputParamType)
			if e := json.Unmarshal(payload, paramValue.Interface()); e != nil {
				return nil, e
			}
			elem := paramValue.Elem()
			args = append(args, elem)
		}
		registerStandardLambdaMetrics()
		// Set cold start counter.
		if isColdStart {
			// Set cold start counter
			csCounter.Inc(1)
			csEventCounter.Inc(1)
			isColdStart = false
		}
		// Set invocations counter.
		invocationsCounter.Inc(1)
		invocationEventCounter.Inc(1)
		start := time.Now()
		lambdaResponse := handlerValue.Call(args)
		executionDuration := time.Since(start)
		// Set duration gauge value in milliseconds.
		dGauge.Update(executionDuration.Nanoseconds() / 1e6)
		if len(lambdaResponse) == 0 {
			return nil, nil
		}
		var err error
		if len(lambdaResponse) > 0 {
			// The last value must always implement error.
			if e, ok := lambdaResponse[len(lambdaResponse)-1].Interface().(error); ok {
				err = e
			}
		}
		var val interface{}
		if len(lambdaResponse) == 2 {
			// In case lambda handler returns 2 arguments(i.e. Maximum allowed return arguments), first
			// argument represents a valid non-error value compatible with the encoding/json standard library.
			val = lambdaResponse[0].Interface()
		}

		return val, err
	}
}

func lambdaErrorHandler(e error) lambdaHandlerFunction {
	return func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		return nil, e
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
	eCounter = metrics.NewCounter()
	eEventCounter = metrics.NewCounter()
	errorsCounterName := wavefront.DeltaCounterName(getStandardLambdaMetricName("errors", false))
	wavefront.RegisterMetric(errorsCounterName, eCounter, nil)
	wavefront.RegisterMetric(getStandardLambdaMetricName("error", true), eEventCounter, nil)

	// Register duration gauge
	dGauge = metrics.NewGauge()
	wavefront.RegisterMetric(getStandardLambdaMetricName("duration", false), dGauge, nil)
}

func reportMetrics(ctx context.Context) {
	// Get standard lambda point tags
	lc, _ := lambdacontext.FromContext(ctx)
	invokedFunctionArn := lc.InvokedFunctionArn
	splitArn := strings.Split(invokedFunctionArn, ":")

	/*
	   Expected formats for Lambda ARN are:
	   https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-lambda
	*/
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

func getStandardLambdaMetricName(metric string, isEvent bool) string {
	const metric_prefix string = "aws.lambda.wf."
	const metric_event_suffix string = "_event"
	if isEvent {
		return strings.Join([]string{metric_prefix, metric, metric_event_suffix}, "")
	}
	return strings.Join([]string{metric_prefix, metric}, "")
}

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
	reportStandardMetrics := os.Getenv("IS_REPORT_STANDARD_METRICS")
	isReportStandardMetrics := true
	if reportStandardMetrics == "False" || reportStandardMetrics == "false" {
		isReportStandardMetrics = false
	}
	return isReportStandardMetrics
}
