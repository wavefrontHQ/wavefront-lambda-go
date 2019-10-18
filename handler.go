package wflambda

import (
	"context"
	"fmt"
	"reflect"
)

// validateLambdaHandler validates the lambdaHandler is a valid handler. When the handler is valid, the function will check
// whether a context parameter is used and returns a boolean based on that. If lambdaHandler is not a valid handler, an error
// is returned. The code of the function is based on source code at
// https://github.com/aws/aws-lambda-go/blob/ea03c2814414b2223eff860ed2286a83ed8a195c/lambda/handler.go#L75
func validateLambdaHandler(lambdaHandler interface{}) (bool, error) {
	if lambdaHandler == nil {
		return false, fmt.Errorf("handler is nil")
	}
	handlerType := reflect.TypeOf(lambdaHandler)

	// Validate lambdaHandler Kind.
	if handlerType.Kind() != reflect.Func {
		return false, fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func)
	}

	// Check if the lambdaHandler takes a context argument.
	takesContext, err := validateArguments(handlerType)
	if err != nil {
		return false, err
	}

	if err := validateReturns(handlerType); err != nil {
		return false, err
	}
	return takesContext, nil
}

// validateArguments validates whether the arguments passed as part of the lambdaHandler are valid. A valid lambdaHandler
// has a maximum of two arguments. When there are two arguments, the first one must be a Context. The function returns
// true or false depending on whether the lambdaHandler has a context argument. If the arguments are not valid, an error
// is returned. Detailed information on the valid handler signatures can be found in the AWS Lambda documentation
// https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html
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

// validateReturns validates whether the arguments returned by the lambdaHandler are valid or not. A valid lambdaHandler
// returns a maximum of two arguments. When there are two arguments, the second argument must be of type error. When there
// is only one argument, that one must be of type error. Detailed information on the valid handler signatures can be found
// in the AWS Lambda documentation
// https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html
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
