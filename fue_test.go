/*
 * Created by Ilan Rasekh on 2019/9/17
 * Copyright (c) 2019. All rights reserved.
 */

package fuego

import (
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

var (
	testCases = []struct {
		Name                 string
		Targets              interface{}
		Args                 []string
		PrintToStdOut        bool
		PrintToStdErr        bool
		ExpectedReturnCount  int
		ExpectedReturnValues []interface{}
		ExpectedError        error
	}{
		{
			"FunctionAddInt.Success",
			AddInt,
			[]string{"Fuego.FunctionAddInt.Success", "3", "5"},
			false,
			false,
			reflect.ValueOf(AddInt).Type().NumOut(),
			[]interface{}{int(8)},
			nil,
		},
		{
			"FunctionExternalFunction.Success",
			math.Frexp,
			[]string{"Fuego.FunctionExternalFunction.Success", "-5"},
			true,
			false,
			reflect.ValueOf(math.Frexp).Type().NumOut(),
			[]interface{}{float64(-0.625), int(3)},
			nil,
		},
		{
			"StructFunctionSubtractFloat64.Success",
			MyMath{Offset: 0}.Subtract,
			[]string{"Fuego.TestSubtract.Success", "3.5", "5.4"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Subtract).Type().NumOut(),
			[]interface{}{float64(-1.9000000000000004)},
			nil,
		},
		{
			"FunctionAddFloat64.Failure",
			AddFloat64,
			[]string{"Fuego.FunctionAddFloat64.Failure", "3.5"},
			false,
			false,
			0,
			nil,
			errors.Errorf("%v", InsufficientArgumentsError),
		},
		{
			"FunctionSubtractInt.Failure",
			SubtractInt,
			[]string{"Fuego.FunctionSubtractInt.Failure", "5", "hi"},
			false,
			true,
			0,
			nil,
			errors.Errorf("%v: %v", ParameterListGenerationError, CannotConvertToDesiredValueTypeError),
		},
		{
			"MapNotSupportedAdd.Failure",
			map[interface{}]bool{MyMath{}: true},
			[]string{"Fuego.MapNotSupportedAdd.Failure", "Add", "5", "hi"},
			false,
			false,
			0,
			nil,
			errors.Errorf("%v", UnsupportedTargetTypeError),
		},
		{
			"StructAdd1.Success",
			MyMath{Offset: 0},
			[]string{"Fuego.StructAdd.Success", "MyMath.Add", "5", "3"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			[]interface{}{float64(8)},
			nil,
		},
		{
			"StructAdd2.Success",
			MyMath{Offset: 0},
			[]string{"Fuego.StructAdd.Success", "Add", "5", "3"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			[]interface{}{float64(8)},
			nil,
		},
		{
			"StructInsufficientArguments.Failure",
			MyMath{Offset: 0},
			[]string{"Fuego.StructInsufficientArguments.Failure"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			nil,
			errors.Errorf("%v", InsufficientArgumentsError),
		},
		{
			"StructInsufficientParameterArgument.Failure",
			MyMath{Offset: 0},
			[]string{"Fuego.StructInsufficientParameterArgument.Failure", "MyMath.Add", "5"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			nil,
			errors.Errorf("%v", InsufficientArgumentsError),
		},
		{
			"StructInsufficientParameterArgument.Failure",
			MyMath{Offset: 0},
			[]string{"Fuego.StructInsufficientParameterArgument.Failure", "MyMath.Add", "5", "hi"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			nil,
			errors.Errorf("%v: %v", ParameterListGenerationError, CannotConvertToDesiredValueTypeError),
		},

		{
			"StructInsufficientParameterArgument.Failure",
			MyMath{Offset: 0},
			[]string{"Fuego.StructInsufficientParameterArgument.Failure", "MyMath.SubtractInt", "5", "3"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			nil,
			errors.Errorf("%v", MethodDoesNotExistError),
		},

		{
			"SliceFunctions.Success",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}.Add},
			[]string{"Fuego.SliceFunctions.Success", "SubtractInt", "5", "4"},
			false,
			false,
			reflect.ValueOf(SubtractInt).Type().NumOut(),
			[]interface{}{1},
			nil,
		},
		{
			"SliceFunctionsWithAdditionalArgs.Success",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithAdditionalArgs.Success", "SubtractInt", "5", "4", "3", "5"},
			false,
			false,
			reflect.ValueOf(SubtractInt).Type().NumOut(),
			[]interface{}{1},
			nil,
		},
		{
			"SliceFunctionsWithNotEnoughArgs1.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs1.Failure"},
			false,
			false,
			0,
			nil,
			errors.New(InsufficientArgumentsError),
		},
		{
			"SliceFunctionsWithNotEnoughArgs2.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs2.Failure", "SubtractInt", "3"},
			false,
			false,
			0,
			nil,
			errors.New(InsufficientArgumentsError),
		},
		{
			"SliceFunctionWithInvalidParameterType.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs2.Failure", "SubtractInt", "3", "hi"},
			false,
			false,
			0,
			nil,
			errors.Errorf("%v: %v", ParameterListGenerationError, CannotConvertToDesiredValueTypeError),
		},
		{
			"SliceStruct1.Success",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceStruct1.Success", "MyMath.Add", "5", "4"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			[]interface{}{float64(9)},
			nil,
		},
		{
			"SliceStruct2.Success",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceStruct2.Success", "MyMath.Add", "5", "4"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Add).Type().NumOut(),
			[]interface{}{float64(9)},
			nil,
		},
		{
			"SliceStructWithAdditionalArgs.Success",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceStructWithAdditionalArgs.Success", "MyMath.Subtract", "5", "4", "3", "5"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Subtract).Type().NumOut(),
			[]interface{}{float64(1)},
			nil,
		},
		{
			"SliceStructWithNotEnoughArgs1.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs1.Failure"},
			false,
			false,
			0,
			nil,
			errors.New(InsufficientArgumentsError),
		},
		{
			"SliceStructWithNotEnoughArgs2.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs2.Failure", "MyMath.Subtract", "3.5"},
			false,
			false,
			0,
			nil,
			errors.New(InsufficientArgumentsError),
		},
		{
			"SliceStructWithInvalidParameterType.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs2.Failure", "MyMath.Subtract", "3.5", "hi"},
			false,
			false,
			0,
			nil,
			errors.Errorf("%v: %v", ParameterListGenerationError, CannotConvertToDesiredValueTypeError),
		},
		{
			"SliceUnsupportedTargetType.Failure",
			[]interface{}{AddInt, SubtractInt, MyMath{Offset: 0}},
			[]string{"Fuego.SliceFunctionsWithNotEnoughArgs2.Failure", "Square", "3.5"},
			false,
			false,
			0,
			nil,
			errors.New(UnsupportedTargetTypeError),
		},
	}
)

func TestFuego(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			PrintToStdOut = testCase.PrintToStdOut
			PrintToStdErr = testCase.PrintToStdErr
			os.Args = testCase.Args
			returnedValues, err := Fuego(testCase.Targets)

			if testCase.ExpectedError != nil {
				if err == nil {
					t.Errorf("Expected the following error but no error was returned: \"%v\"", testCase.ExpectedError)
				} else if !doErrorsMatch(testCase.ExpectedError, err) {
					t.Errorf("Expected to receive the first error but instead the second error was returned: \n\t1) \"%v\"\n\t2) \"%v\"", testCase.ExpectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Error is not expected but got %v", err)
				}

				if len(returnedValues) != testCase.ExpectedReturnCount {
					t.Errorf("%d return value expected, got %d", testCase.ExpectedReturnCount, len(returnedValues))
				}

				for x, returnedVal := range returnedValues {
					returnedValKind := returnedVal.Kind()
					expectedVal := testCase.ExpectedReturnValues[x]
					expectedValType := reflect.TypeOf(expectedVal)

					if returnedValKind != expectedValType.Kind() {
						t.Errorf("the returned value type \"%v\" does not equal the expected value type \"%v\" at the position \"%v\"", returnedValKind, expectedValType.Kind(), x)
					} else {
						if !reflect.DeepEqual(returnedVal.Interface(), expectedVal) {
							returnedValueErrorStr := "the returned value \"%v\" does not equal the expected value \"%v\" at the position \"%v\""
							t.Errorf(returnedValueErrorStr, returnedVal, expectedVal, x)
						}
					}
				}

			}
		})
	}

}

func TestSetupFunctionParameterValues(t *testing.T) {
	successArgs := []string{
		strconv.FormatInt(-1, 10),
		strconv.FormatInt(-2, 10),
		strconv.FormatInt(-3, 10),
		strconv.FormatInt(-4, 10),
		strconv.FormatInt(-5, 10),
		strconv.FormatUint(uint64(1), 10),
		strconv.FormatUint(uint64(2), 10),
		strconv.FormatUint(uint64(3), 10),
		strconv.FormatUint(uint64(4), 10),
		strconv.FormatUint(uint64(5), 10),
		strconv.FormatFloat(3.14159265359, 'f', -1, 32),
		fmt.Sprintf("%f", 3.14159265359),
		strconv.FormatBool(true),
		"hi",
	}
	successKind := []reflect.Kind{
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.Bool,
		reflect.String,
	}

	if _, err := setupFunctionParameterValues(successKind, successArgs); err != nil {
		t.Errorf(err.Error())
	}

	for _, kind := range successKind {
		if kind != reflect.String {
			val, err := setupFunctionParameterValues([]reflect.Kind{kind}, []string{"This Should Fail"})
			if err == nil {
				t.Errorf("the string was parsed to \"%v\" with the value of \"%v\"", kind, val[0].Interface())
			} else if !doErrorsMatch(errors.New(CannotConvertToDesiredValueTypeError), err) {
				t.Errorf("Expected to receive the first error but instead the second error was returned: \n\t1) \"%v\"\n\t2) \"%v\"", CannotConvertToDesiredValueTypeError, err)
			}
		}
	}
}

/* Test Help Functions */
func AddFloat64(a float64, b float64) float64 {
	return a + b
}

func AddInt(a int, b int) int {
	return a + b
}

func SubtractInt(a int, b int) int {
	return a - b
}

type MyMath struct {
	Offset float64
}

func (m MyMath) Add(a float64, b float64) float64 {
	return a + b + m.Offset
}

func (m MyMath) Subtract(a float64, b float64) float64 {
	return a - b - m.Offset
}

func doErrorsMatch(err1 error, err2 error) bool {
	if err1.Error() == err2.Error() {
		return true
	} else if getStaticErrorParts(err1) == getStaticErrorParts(err2) {
		return true
	}

	return false
}

func getStaticErrorParts(err error) string {
	errMsgParts := strings.Split(err.Error(), "\"")
	errMsg := ""
	for x, msgPart := range errMsgParts {
		if x == 0 {
			errMsg = msgPart
		} else if x%2 == 0 {
			errMsg += " " + msgPart
		}
	}
	return errMsg
}
