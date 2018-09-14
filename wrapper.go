package wflambda

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/rcrowley/go-metrics"
)

type lambdaHandlerFunction func(context.Context, json.RawMessage) (interface{}, error)

var (
	server                    string
	authToken                 string
	reportStandardMetrics     bool
	lambdaHandlerTakesContext bool
	handlerType               reflect.Type
	handlerValue              reflect.Value
	coldStart                 = true
	csCounter                 metrics.Counter
	invocationsCounter        metrics.Counter
	errCounter                metrics.Counter
	durationGauge             metrics.GaugeFloat64
)

// Returns the Wavefront Lambda wrapper. The wrapper collects aws lambda's
//   standard metrics and reports it directly to the specified wavefront url. It
//   requires the following Environment variables to be set:
//   1.WAVEFRONT_URL : https://<INSTANCE>.wavefront.com
//   2.WAVEFRONT_API_TOKEN : Wavefront API token with Direct Data Ingestion permission
//   3.REPORT_STANDARD_METRICS : Set to False to not report standard lambda
//                                 metrics directly to wavefront.

func Wrapper(lambdaHandler interface{}) interface{} {
	// Validate wrapper environment variables.
	reportStandardMetrics = getAndValidateLambdaEnvironment()
	// Check if lambdaHandler is a valid handler.
	handlerTakesContext, err := validateLambdaHandler(lambdaHandler)
	lambdaHandlerTakesContext = handlerTakesContext
	if err != nil {
		return lambdaErrorHandler(err)
	}
	handlerType = reflect.TypeOf(lambdaHandler)
	handlerValue = reflect.ValueOf(lambdaHandler)

	//Returns a wavefrontLambda wrapper function with standard lambda metrics.
	return lambdaHandlerWrapper
}

func lambdaHandlerWrapper(ctx context.Context, payload json.RawMessage) (response interface{}, lambdaHandlerError error) {
	defer func() {
		var err interface{}
		// Increment error count if there is a panic or non nil error is returned
		// by users lambda handler function.
		if e := recover(); e != nil {
			err = e
			// Set error counters
			incrementCounter(errCounter, 1, reportStandardMetrics)
		} else if lambdaHandlerError != nil {
			// Set error counters
			incrementCounter(errCounter, 1, reportStandardMetrics)
		}
		reportMetrics(ctx)
		metrics.DefaultRegistry.UnregisterAll()
		if err != nil {
			panic(err)
		}
	}()

	var args []reflect.Value
	if lambdaHandlerTakesContext {
		args = append(args, reflect.ValueOf(ctx))
	}
	if (handlerType.NumIn() == 1 && !lambdaHandlerTakesContext) || handlerType.NumIn() == 2 {
		inputParamType := handlerType.In(handlerType.NumIn() - 1)
		paramValue := reflect.New(inputParamType)
		if e := json.Unmarshal(payload, paramValue.Interface()); e != nil {
			return nil, e
		}
		elem := paramValue.Elem()
		args = append(args, elem)
	}
	if reportStandardMetrics {
		registerStandardLambdaMetrics()
	}

	if coldStart {
		// Set cold start counter
		incrementCounter(csCounter, 1, reportStandardMetrics)
		coldStart = false
	}
	// Set invocations counter.
	incrementCounter(invocationsCounter, 1, reportStandardMetrics)
	start := time.Now()
	lambdaResponse := handlerValue.Call(args)
	executionDuration := time.Since(start)
	// Set duration gauge value in milliseconds.
	updateGaugeFloat64(durationGauge, executionDuration.Seconds()*1000, reportStandardMetrics)
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

func lambdaErrorHandler(e error) lambdaHandlerFunction {
	return func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		return nil, e
	}
}
