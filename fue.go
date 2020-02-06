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
	// PrintToStdOut is used to determine if results should be printed to std out. Default is true but can be set to false prior to calling Fuego()
	PrintToStdOut = true
	// PrintToStdErr is used to determine if errors should be printed to std err. Default is true but can be set to false prior to calling Fuego()
	PrintToStdErr = true
)

// Fuego handles the parsing of potential targets to call and then reflectively calls the function with all necessary params
func Fuego(target interface{}) ([]reflect.Value, error) {
	osArgs := os.Args

	// If the user is trying to view the usage details for the binary or a
	// subcommand of the binary print them and exit
	if len(osArgs) == 2 {
		firstArg := strings.ToLower(strings.TrimSpace(osArgs[1]))
		if firstArg == "help" || firstArg == "--help" || firstArg == "-help" || firstArg == "-h" {
			fuegoUsage(target, "")
			// os.Exit(1)
			return nil, nil
		}
	} else if len(osArgs) == 3 {
		secondArg := strings.ToLower(strings.TrimSpace(osArgs[2]))
		if secondArg == "help" || secondArg == "--help" || secondArg == "-help" || secondArg == "-h" {
			fuegoUsage(target, osArgs[1])
			// os.Exit(1)
			return nil, nil
		}

	}

	// Execute the command the user is trying to run against the target
	return fuegoTarget(target)
}

// fuegoTarget actually handles the parsing of potential targets to call and then reflectively calls the function with all necessary params
func fuegoTarget(target interface{}) ([]reflect.Value, error) {
	osArgs := os.Args
	targetType := reflect.TypeOf(target)

	switch targetType.Kind() {
	case reflect.Func:
		return fuegoPrintWrapper(fuegoFunc(target, osArgs))
	case reflect.Ptr, reflect.Struct:
		return fuegoPrintWrapper(fuegoStruct(target, osArgs))
	case reflect.Array, reflect.Slice:
		if len(osArgs) < 2 {
			return nil, errors.Errorf(InsufficientArgumentsError)
		}

		methodTitleName := strings.Title(osArgs[1])

		// foreach element in the slice of targets provided, check to see if this is what was called by the cli
		// if so call this function passing in the element as the new target
		for _, key := range target.([]interface{}) {
			keyType := reflect.TypeOf(key)

			if keyType.Kind() == reflect.Func && functionName(key) == methodTitleName {
				return Fuego(key)
			} else if keyType.Kind() == reflect.Struct && strings.HasPrefix(methodTitleName, keyType.Name()+".") {
				return Fuego(key)
			} else if keyType.Kind() == reflect.Ptr && keyType.Elem().Kind() == reflect.Struct && strings.HasPrefix(methodTitleName, keyType.Elem().Name()+".") {
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

// fuegoUsage determines the usage details and prints them to stdout
func fuegoUsage(target interface{}, helpTarget string) {
	helpTarget = strings.TrimSpace(helpTarget)
	targetType := reflect.TypeOf(target)
	targetVal := reflect.ValueOf(target)

	switch targetType.Kind() {
	case reflect.Func:
		// get the target functions name
		targetFuncName := runtime.FuncForPC(targetVal.Pointer()).Name()
		targetFuncName = strings.TrimSpace(targetFuncName[strings.LastIndex(targetFuncName, ".")+1:])

		// if the input requests an invalid subcommand let them know
		if helpTarget != "" && strings.ToLower(helpTarget) != strings.ToLower(targetFuncName) {
			fmt.Printf("%s is not an available subcommand. Here are the details for the available function.\n\n", helpTarget)
		}

		// print the function signature to understand the parameter types required and expected return type
		fmt.Printf(targetFuncName)
		funcStructure := targetVal.Type().String()
		fmt.Println(" " + funcStructure[strings.Index(funcStructure, "("):])

		// print out paramter types separately
		targetFuncParamCount := targetVal.Type().NumIn()
		fmt.Printf("\tParamters: ")
		if targetFuncParamCount == 0 {
			fmt.Println("None")
		}
		params := fmt.Sprintf("(%s)", funcStructure[strings.Index(funcStructure, "(")+1:strings.Index(funcStructure, ")")])
		fmt.Println(params)

		// print out return types separately
		targetFuncReturnCount := targetVal.Type().NumOut()
		fmt.Printf("\tReturns:   ")
		if targetFuncReturnCount == 0 {
			fmt.Println("None")
		}
		ret := fmt.Sprintf("%s", funcStructure[strings.Index(funcStructure, ")")+2:])
		if !strings.Contains(ret, "(") {
			ret = fmt.Sprintf("(%s)", ret)
		}
		fmt.Println(ret)

	case reflect.Ptr, reflect.Struct:
		targetVal := reflect.ValueOf(reflect.ValueOf(&target).Elem().Interface())

		structNameSplit := strings.Split(targetVal.Type().String(), ".")
		structName := structNameSplit[len(structNameSplit)-1]
		structPackageName := strings.Replace(strings.Join(structNameSplit[:len(structNameSplit)-1], ""), "*", "", -1)

		// if the input requests an invalid subcommand let them know
		if helpTarget != "" && strings.ToLower(structName) != strings.ToLower(helpTarget) {
			fmt.Printf("%s is not an available subcommand. Here are the details for the available struct.\n\n", helpTarget)
		}

		// print usage for a struct
		fmt.Println(structName + ":")

		// print available methods for a struct that can be called as a subcommand
		fmt.Printf("\t%s Methods\n", structName)
		methodsCount := targetVal.Elem().NumMethod()
		for x := 0; x < methodsCount; x++ {
			structMethod := targetVal.Elem().Type().Method(x)
			methodSig := structMethod.Type.String()
			methodSig = strings.Replace(methodSig, "func("+structPackageName+"."+structName+", ", "(", 1)
			fmt.Printf("\t\t%s %v\n", structMethod.Name, methodSig)
		}

		// print available fields for a struct that can be set with flags "--<field>=<value>"
		fmt.Printf("\t%s Fields\n", structName)
		fieldsCount := targetVal.Elem().NumField()
		for x := 0; x < fieldsCount; x++ {
			fmt.Printf("\t\t--%s=<%v>\n", targetVal.Elem().Type().Field(x).Name, targetVal.Elem().Type().Field(x).Type)
		}

	case reflect.Array, reflect.Slice:
	}

	return
}

// fuegoFunc is used as a helper function for Fuego() to handle targets of type Func
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
			funcParams, err = convertStringsToReflectValues(targetValKind, args[2:])
			if err != nil {
				return nil, errors.Wrap(err, ParameterListGenerationError)
			}
		} else if len(args) > 1 && args[1] != targetFuncName {
			funcParams, err = convertStringsToReflectValues(targetValKind, args[1:])
			if err != nil {
				return nil, errors.Wrap(err, ParameterListGenerationError)
			}
		}
	}

	return targetVal.Call(funcParams), nil
}

// fuegoStruct is used as a helper function for Fuego() to handle targets of type Struct or pointer to a Struct
func fuegoStruct(target interface{}, args []string) ([]reflect.Value, error) {
	targetVal := reflect.ValueOf(reflect.ValueOf(&target).Elem().Interface())

	structNameSplit := strings.Split(targetVal.Type().String(), ".")
	structName := structNameSplit[len(structNameSplit)-1]
	// structPackageName := strings.Replace(strings.Join(structNameSplit[:len(structNameSplit)-1], ""), "*", "", -1)

	if len(args) < 2 {
		return nil, errors.Errorf(InsufficientArgumentsError)
	}

	for _, arg := range args {
		if strings.HasPrefix(arg, "--") && len(arg) > 2 {
			arg = arg[2:]
			argSplit := strings.Split(arg, "=")
			attribute := targetVal.Elem().FieldByName(argSplit[0])

			if attribute.CanSet() {
				vals, err := convertStringsToReflectValues([]reflect.Kind{attribute.Kind()}, []string{argSplit[1]})
				if err != nil {
					// do i error out or ignore and continue and print the error - leaning to fail
					printError(errors.Wrap(err, "the struct attribute could not be altered"))
					continue
				}
				attribute.Set(vals[0])
			}
		}
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

		funcParams, err = convertStringsToReflectValues(targetValKind, args[2:])
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

// printValues is used to handle printing out Reflect Values to std out if the user would like to allow it
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

// printError is used to handle printing out an Error to std err if the user would like to allow it
func printError(err error) {
	if PrintToStdErr {
		_, _ = os.Stderr.WriteString("Error: " + err.Error())
	}
}

// fuegoPrintWrapper is a simple wrapper function to parse the results and error that Fuego would return and print it out to std out / std err if desired
func fuegoPrintWrapper(values []reflect.Value, err error) ([]reflect.Value, error) {
	if err != nil {
		printError(err)
	} else {
		printValues(values)
	}
	return values, err
}

// convertStringsToReflectValues converts a list of strings to the desired reflect value so that it can be used as a parameter for a reflective call of a function or to be set as the value of a struct attribute.
func convertStringsToReflectValues(targetValueKindSlice []reflect.Kind, args []string) ([]reflect.Value, error) {

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
