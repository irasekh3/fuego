/*
 * Created by Ilan Rasekh on 2019/9/17
 * Copyright (c) 2019. All rights reserved.
 */

package fuego

import (
	"math"
	"os"
	"reflect"
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
			false,
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
			false,
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
