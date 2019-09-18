/*
 * Created by Ilan Rasekh on 2019/9/17
 * Copyright (c) 2019. All rights reserved.
 */

package fuego

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
	"reflect"
)

const (
	UnsupportedTargetTypeError                   = "passing in \"%v\" is not yet supported"
	InsufficientArgumentsError                   = "not enough arguments were passed in to setup the function parameter values"
	ParameterListGenerationError                 = "could not generate the necessary function parameters list"
	CannotConvertToDesiredValueTypeError         = "cannot convert \"%v\" to \"%v\" as needed"
	UnsupportedConversionToDesiredValueTypeError = "fuego does not yet support converting attributes of type \"%v\""
)

var (
	PrintToStdOut = true
	PrintToStdErr = true
)

// Fuego handles the parsing of potential targets to call
// and then reflectively calls the function with all necessary params
func Fuego(targets interface{}) ([]reflect.Value, error) {
	targetType := reflect.TypeOf(targets)

	err := errors.Errorf(UnsupportedTargetTypeError, targetType.Kind())
	printError(err)
	return nil, err
}

func printValues(values []reflect.Value) {
	if PrintToStdOut {
		for x, val := range values {
			fmt.Print(val.Interface())
			if x < len(values)-1 {
				fmt.Print(", ")
			}
		}
	}
}

func printError(err error) {
	if PrintToStdErr {
		os.Stderr.WriteString("Error: " + err.Error())
	}
}
