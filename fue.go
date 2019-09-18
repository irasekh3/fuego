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
	MethodDoesNotExistError                      = "the method \"%v\" for struct \"%v\" does not exist"
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
	osArgs := os.Args
	targetType := reflect.TypeOf(targets)

	switch targetType.Kind() {
	case reflect.Func:
		return fuegoPrintWrapper(fuegoFunc(targets, osArgs))
	case reflect.Struct:
		return fuegoPrintWrapper(fuegoStruct(targets, osArgs))
	case reflect.Array, reflect.Slice:
		if len(osArgs) < 2 {
			return nil, errors.Errorf(InsufficientArgumentsError)
		}

		methodTitleName := strings.Title(osArgs[1])

		// foreach element in the slice of targets provided, check to see if this is what was called by the cli
		// if so call this function passing in the element as the new target
		for _, key := range targets.([]interface{}) {
			keyType := reflect.TypeOf(key)

			if keyType.Kind() == reflect.Func && functionName(key) == methodTitleName {
				return Fuego(key)
			} else if keyType.Kind() == reflect.Struct && strings.HasPrefix(methodTitleName, keyType.Name()+".") {
				return Fuego(key)
			}
		}

		return nil, errors.Errorf(UnsupportedTargetTypeError, methodTitleName)
	default:
		err := errors.Errorf(UnsupportedTargetTypeError, targetType.Kind())
		printError(err)
		return nil, err
	}
}

func fuegoFunc(target interface{}, args []string) ([]reflect.Value, error) {
	targetVal := reflect.ValueOf(target)
	targetFuncName := runtime.FuncForPC(targetVal.Pointer()).Name()
	targetFuncName = targetFuncName[strings.LastIndex(targetFuncName, ".")+1:]
	// targetFuncParamCount := targetVal.Type().NumIn()

	targetFuncParamCount := targetVal.Type().NumIn()

	if len(args) > 1 && args[1] == targetFuncName && len(args)-2 < targetFuncParamCount {
		// the function name is explicitly called out but not enough params passed in
		return nil, errors.Errorf(InsufficientArgumentsError)
	} else if len(args) > 1 && args[1] != targetFuncName && len(args)-1 < targetFuncParamCount {
		// the function name is not passed is and there are not enough params passed in
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

		if len(args) > 2 && args[1] == targetFuncName {
			funcParams, err = setupFunctionParameterValues(targetValKind, args[2:])
			if err != nil {
				return nil, errors.Wrap(err, ParameterListGenerationError)
			}
		} else if len(args) > 1 && args[1] != targetFuncName {
			funcParams, err = setupFunctionParameterValues(targetValKind, args[1:])
			if err != nil {
				return nil, errors.Wrap(err, ParameterListGenerationError)
			}
		}
	}

	return targetVal.Call(funcParams), nil
}

func fuegoStruct(target interface{}, args []string) ([]reflect.Value, error) {
	targetVal := reflect.ValueOf(target)

	// returns <struct>
	structName := targetVal.Type().Name()

	// returns <package>.<struct>
	// packageStructName := targetVal.Type().String()

	if len(args) < 2 {
		return nil, errors.Errorf(InsufficientArgumentsError)
	}

	methodName := args[1]
	if strings.Contains(methodName, structName) {
		methodName = methodName[strings.LastIndex(methodName, ".")+1:]
	}

	method := targetVal.MethodByName(methodName)
	if !method.IsValid() {
		return nil, errors.Errorf(MethodDoesNotExistError, methodName, structName)
	}

	targetMethodParamCount := method.Type().NumIn()

	if len(args)-2 < targetMethodParamCount {
		return nil, errors.New(InsufficientArgumentsError)
	}

	var funcParams []reflect.Value
	var err error

	if targetMethodParamCount > 0 {
		targetValKind := make([]reflect.Kind, targetMethodParamCount)

		for x := 0; x < targetMethodParamCount; x++ {
			targetValKind[x] = method.Type().In(x).Kind()
		}

		funcParams, err = setupFunctionParameterValues(targetValKind, args[2:])
		if err != nil {
			return nil, errors.Wrap(err, ParameterListGenerationError)
		}
	}

	return method.Call(funcParams), nil
}

func functionName(key interface{}) string {
	funcName := runtime.FuncForPC(reflect.ValueOf(key).Pointer()).Name()
	return funcName[strings.LastIndex(funcName, ".")+1:]
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
