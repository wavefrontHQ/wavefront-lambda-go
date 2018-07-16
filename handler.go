package wflambda

import (
	"context"
	"fmt"
	"reflect"
)

// Validate the input lambda handler. This is taken from the below amazon source code for go lambda handlers
// https://github.com/aws/aws-lambda-go/blob/ea03c2814414b2223eff860ed2286a83ed8a195c/lambda/handler.go#L75
func validateLambdaHandler(lambdaHandler interface{}) (bool, error){
  if lambdaHandler == nil {
    return false, fmt.Errorf("handler is nil")
  }
  handlerType := reflect.TypeOf(lambdaHandler)
  // Validate lambdaHandler Kind.
  if handlerType.Kind() != reflect.Func {
    return false, fmt.Errorf("handler kind %s is not %s", handlerType.Kind(), reflect.Func)
  }
  takesContext, err := validateArguments(handlerType)
	if err != nil {
		return false, err
	}

	if err := validateReturns(handlerType); err != nil {
		return false, err
	}
	return takesContext, nil
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
