package wflambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambdacontext"
)

// lambdaHandler is the generic function type
type lambdaHandler func(context.Context, interface{}) (interface{}, error)

// wrapHandler decorates the handler with the handler wrapper
func wrapHandler(handler interface{}, wa *WavefrontAgent) lambdaHandler {
	return func(context context.Context, payload interface{}) (interface{}, error) {
		handlerWrapper := NewHandlerWrapper(handler, wa)
		return handlerWrapper.Invoke(context, payload)
	}
}

// HandlerWrapper
type HandlerWrapper struct {
	wavefrontAgent  *WavefrontAgent
	lambdaContext   *lambdacontext.LambdaContext
	originalHandler interface{}
	wrappedHandler  lambdaHandler
}

// NewHandlerWrapper creates a new wrapper
func NewHandlerWrapper(handler interface{}, wa *WavefrontAgent) *HandlerWrapper {
	return &HandlerWrapper{
		wavefrontAgent:  wa,
		originalHandler: handler,
		wrappedHandler:  newHandler(handler),
	}
}

func (hw *HandlerWrapper) Invoke(ctx context.Context, payload interface{}) (response interface{}, err error) {
	// Get the lambda context
	lc, _ := lambdacontext.FromContext(ctx)
	hw.lambdaContext = lc
	// Get the point tags
	invokedFunctionArn := hw.lambdaContext.InvokedFunctionArn
	splitArn := strings.Split(invokedFunctionArn, ":")

	// Expected formats for Lambda ARN are:
	// https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-lambda
	hw.wavefrontAgent.WavefrontConfig.PointTags["LambdaArn"] = invokedFunctionArn
	hw.wavefrontAgent.WavefrontConfig.PointTags["source"] = lambdacontext.FunctionName
	hw.wavefrontAgent.WavefrontConfig.PointTags["FunctionName"] = lambdacontext.FunctionName
	hw.wavefrontAgent.WavefrontConfig.PointTags["ExecutedVersion"] = lambdacontext.FunctionVersion
	hw.wavefrontAgent.WavefrontConfig.PointTags["Region"] = splitArn[3]
	hw.wavefrontAgent.WavefrontConfig.PointTags["accountId"] = splitArn[4]

	if splitArn[5] == "function" {
		hw.wavefrontAgent.WavefrontConfig.PointTags["Resource"] = splitArn[6]
		if len(splitArn) == 8 {
			hw.wavefrontAgent.WavefrontConfig.PointTags["Resource"] += ":" + splitArn[7]
		}
	} else if splitArn[5] == "event-source-mappings" {
		hw.wavefrontAgent.WavefrontConfig.PointTags["EventSourceMappings"] = splitArn[6]
	}

	defer func() {
		var deferedErr interface{}
		if e := recover(); e != nil {
			deferedErr = e
			// Set error counters
			errCounter.Increment(1)
			hw.wavefrontAgent.sender.SendDeltaCounter("aws.lambda.wf.errors", errCounter.val, lambdacontext.FunctionName, hw.wavefrontAgent.WavefrontConfig.PointTags)
		} else if err != nil {
			// Set error counters
			errCounter.Increment(1)
			hw.wavefrontAgent.sender.SendDeltaCounter("aws.lambda.wf.errors", errCounter.val, lambdacontext.FunctionName, hw.wavefrontAgent.WavefrontConfig.PointTags)
		}

		hw.wavefrontAgent.sender.Flush()
		hw.wavefrontAgent.sender.Close()

		if deferedErr != nil {
			panic(deferedErr)
		}
	}()

	// Start timer
	startTime := time.Now()

	// Call handler
	invocationsCounter.Increment(1)
	response, err = hw.wrappedHandler(ctx, payload)
	if err != nil {
		errCounter.Increment(1)
	}

	// Stop timer and report
	if coldStart {
		// Set cold start counter.
		csCounter.Increment(1)
		coldStart = false
	}
	duration := time.Since(startTime)

	reportTime := time.Now().Unix()

	hw.wavefrontAgent.counters["aws.lambda.wf.coldstarts"] = csCounter.val
	hw.wavefrontAgent.counters["aws.lambda.wf.invocations"] = invocationsCounter.val
	hw.wavefrontAgent.metrics["aws.lambda.wf.duration"] = float64(duration.Milliseconds())

	memstats := getMemoryStats()
	hw.wavefrontAgent.metrics["aws.lambda.wf.mem.total"] = memstats.Total
	hw.wavefrontAgent.metrics["aws.lambda.wf.mem.used"] = memstats.Used
	hw.wavefrontAgent.metrics["aws.lambda.wf.mem.percentage"] = memstats.UsedPercentage

	for metricName, metricValue := range hw.wavefrontAgent.metrics {
		err = hw.wavefrontAgent.sender.SendMetric(metricName, metricValue, reportTime, lambdacontext.FunctionName, hw.wavefrontAgent.WavefrontConfig.PointTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}
	}

	for metricName, metricValue := range hw.wavefrontAgent.counters {
		err = hw.wavefrontAgent.sender.SendDeltaCounter(metricName, metricValue, lambdacontext.FunctionName, hw.wavefrontAgent.WavefrontConfig.PointTags)
		if err != nil {
			log.Println("ERROR :: ", err)
		}
	}

	return response, err
}

func errorHandler(e error) lambdaHandler {
	return func(ctx context.Context, event interface{}) (interface{}, error) {
		return nil, e
	}
}

func validateArguments(handler reflect.Type) (bool, error) {
	handlerTakesContext := false
	if handler.NumIn() > 2 {
		return false, fmt.Errorf("handlers may not take more than two arguments, but handler takes %d", handler.NumIn())
	} else if handler.NumIn() > 0 {
		contextType := reflect.TypeOf((*context.Context)(nil)).Elem()
		argumentType := handler.In(0)
		handlerTakesContext = argumentType.Implements(contextType)
		if handler.NumIn() > 1 && !handlerTakesContext {
			return false, fmt.Errorf("handler takes two arguments, but the first is not Context. got %s", argumentType.Kind())
		}
	}

	return handlerTakesContext, nil
}

func validateReturns(handler reflect.Type) error {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if handler.NumOut() > 2 {
		return fmt.Errorf("handler may not return more than two values")
	} else if handler.NumOut() > 1 {
		if !handler.Out(1).Implements(errorType) {
			return fmt.Errorf("handler returns two values, but the second does not implement error")
		}
	} else if handler.NumOut() == 1 {
		if !handler.Out(0).Implements(errorType) {
			return fmt.Errorf("handler returns a single value, but it does not implement error")
		}
	}
	return nil
}

// newHandler Creates the base lambda handler, which will do basic payload unmarshaling before defering to handlerSymbol.
// If handlerSymbol is not a valid handler, the returned function will be a handler that just reports the validation error.
func newHandler(handlerSymbol interface{}) lambdaHandler {
	if handlerSymbol == nil {
		return errorHandler(fmt.Errorf("handler is nil"))
	}
	handler := reflect.ValueOf(handlerSymbol)
	handlerType := reflect.TypeOf(handlerSymbol)
	if handlerType.Kind() != reflect.Func {
		return errorHandler(fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func))
	}

	takesContext, err := validateArguments(handlerType)
	if err != nil {
		return errorHandler(err)
	}

	if err := validateReturns(handlerType); err != nil {
		return errorHandler(err)
	}

	return func(ctx context.Context, payload interface{}) (interface{}, error) {
		// construct arguments
		var args []reflect.Value
		if takesContext {
			args = append(args, reflect.ValueOf(ctx))
		}

		if (handlerType.NumIn() == 1 && !takesContext) || handlerType.NumIn() == 2 {
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				return nil, err
			}

			eventType := handlerType.In(handlerType.NumIn() - 1)
			event := reflect.New(eventType)

			if err := json.Unmarshal(payloadBytes, event.Interface()); err != nil {
				return nil, err
			}

			args = append(args, event.Elem())
		}

		response := handler.Call(args)

		// convert return values into (interface{}, error)
		var err error
		if len(response) > 0 {
			if errVal, ok := response[len(response)-1].Interface().(error); ok {
				err = errVal
			}
		}
		var val interface{}
		if len(response) > 1 {
			val = response[0].Interface()
		}

		return val, err
	}
}
