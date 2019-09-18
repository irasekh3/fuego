/*
 * Created by Ilan Rasekh on 2019/9/17
 * Copyright (c) 2019. All rights reserved.
 */

package fuego

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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

func fuegoFunc(target interface{}, args []string) ([]reflect.Value, error) {
	targetVal := reflect.ValueOf(target)
	targetFuncName := runtime.FuncForPC(targetVal.Pointer()).Name()
	targetFuncName = targetFuncName[strings.LastIndex(targetFuncName, ".")+1:]
	// targetFuncParamCount := targetVal.Type().NumIn()

	targetFuncParamCount := targetVal.Type().NumIn()

	if len(args)-1 < targetFuncParamCount {
		return nil, errors.Errorf(InsufficientArgumentsError)
	}

	var funcParams []reflect.Value
	var err error

	if targetFuncParamCount > 0 {
		targetValKind := make([]reflect.Kind, targetFuncParamCount)

		for x := 0; x < targetFuncParamCount; x++ {
			// parameterVal := reflect.ValueOf(args[x])
			targetValKind[x] = targetVal.Type().In(x).Kind()
		}

		funcParams, err = setupFunctionParameterValues(targetValKind, args[1:])
		if err != nil {
			return nil, errors.Wrap(err, ParameterListGenerationError)
		}
	}

	returnVals := targetVal.Call(funcParams)
	return returnVals, nil
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

func fuegoPrintWrapper(values []reflect.Value, err error) ([]reflect.Value, error) {
	if err != nil {
		printError(err)
	} else {
		printValues(values)
	}
	return values, err
}

func setupFunctionParameterValues(targetValueKindSlice []reflect.Kind, args []string) ([]reflect.Value, error) {

	size := len(targetValueKindSlice)

	funcParams := make([]reflect.Value, size)

	for x, targetValKind := range targetValueKindSlice {

		switch targetValKind {
		case reflect.Int:
			paramVal, err := strconv.ParseInt(args[x], 10, 0)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(int(paramVal))

		case reflect.Int8:
			paramVal, err := strconv.ParseInt(args[x], 10, 8)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(int8(paramVal))

		case reflect.Int16:
			paramVal, err := strconv.ParseInt(args[x], 10, 16)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(int16(paramVal))

		case reflect.Int32:
			paramVal, err := strconv.ParseInt(args[x], 10, 32)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(int32(paramVal))

		case reflect.Int64:
			paramVal, err := strconv.ParseInt(args[x], 10, 64)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(paramVal)

		case reflect.Uint:
			paramVal, err := strconv.ParseUint(args[x], 10, 0)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(uint(paramVal))

		case reflect.Uint8:
			paramVal, err := strconv.ParseUint(args[x], 10, 8)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(uint8(paramVal))

		case reflect.Uint16:
			paramVal, err := strconv.ParseUint(args[x], 10, 16)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(uint16(paramVal))

		case reflect.Uint32:
			paramVal, err := strconv.ParseUint(args[x], 10, 32)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(uint32(paramVal))

		case reflect.Uint64:
			paramVal, err := strconv.ParseUint(args[x], 10, 64)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(paramVal)

		case reflect.Float32:
			paramVal, err := strconv.ParseFloat(args[x], 32)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(float32(paramVal))

		case reflect.Float64:
			paramVal, err := strconv.ParseFloat(args[x], 64)
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(paramVal)

		case reflect.Bool:
			paramVal, err := strconv.ParseBool(args[x])
			if err != nil {
				return nil, errors.Errorf(CannotConvertToDesiredValueTypeError, args[x], targetValKind)
			}
			funcParams[x] = reflect.ValueOf(paramVal)

		case reflect.String:
			funcParams[x] = reflect.ValueOf(args[x])

		default:
			return nil, errors.Errorf(UnsupportedConversionToDesiredValueTypeError, targetValKind)
		}

	}

	return funcParams, nil
}
