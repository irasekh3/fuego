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
		{
			"StructAttributeArgument.Success",
			[]interface{}{&(MyMath{Offset: 5})},
			[]string{"Fuego.StructAttributeArgument.Success", "MyMath.Add", "5", "3", "--Offset=2"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Subtract).Type().NumOut(),
			[]interface{}{float64(10)},
			nil,
		},

		{
			"StructAttributeInvalidTypeArgument.Success",
			[]interface{}{&(MyMath{Offset: 5})},
			[]string{"Fuego.StructAttributeArgument.Success", "MyMath.Add", "5", "3", "--Offset=hi"},
			false,
			false,
			reflect.ValueOf(MyMath{Offset: 0}.Subtract).Type().NumOut(),
			[]interface{}{float64(13)},
			nil,
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

func TestConvertStringsToReflectValues(t *testing.T) {
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

	if _, err := convertStringsToReflectValues(successKind, successArgs); err != nil {
		t.Errorf(err.Error())
	}

	for _, kind := range successKind {
		if kind != reflect.String {
			val, err := convertStringsToReflectValues([]reflect.Kind{kind}, []string{"This Should Fail"})
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

func TestGetTypeInfo(t *testing.T) {
	tests := []struct {
		name       string
		target     interface{}
		wantKind   reflect.Kind
		wantName   string
		wantIsPtr  bool
		wantMinMethods int
	}{
		{
			name:       "function type",
			target:     AddInt,
			wantKind:   reflect.Func,
			wantName:   "",
			wantIsPtr:  false,
			wantMinMethods: 0,
		},
		{
			name:       "struct type",
			target:     MyMath{Offset: 1.0},
			wantKind:   reflect.Struct,
			wantName:   "MyMath",
			wantIsPtr:  false,
			wantMinMethods: 2, // Add and Subtract
		},
		{
			name:       "pointer to struct",
			target:     &MyMath{Offset: 2.0},
			wantKind:   reflect.Ptr,
			wantName:   "MyMath",
			wantIsPtr:  true,
			wantMinMethods: 2,
		},
		{
			name:       "string type",
			target:     "hello",
			wantKind:   reflect.String,
			wantName:   "string",
			wantIsPtr:  false,
			wantMinMethods: 0,
		},
		{
			name:       "int type",
			target:     42,
			wantKind:   reflect.Int,
			wantName:   "int",
			wantIsPtr:  false,
			wantMinMethods: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := GetTypeInfo(tc.target)
			if info.Kind != tc.wantKind {
				t.Errorf("Kind = %v, want %v", info.Kind, tc.wantKind)
			}
			if info.Name != tc.wantName {
				t.Errorf("Name = %v, want %v", info.Name, tc.wantName)
			}
			if info.IsPtr != tc.wantIsPtr {
				t.Errorf("IsPtr = %v, want %v", info.IsPtr, tc.wantIsPtr)
			}
			if info.NumMethods < tc.wantMinMethods {
				t.Errorf("NumMethods = %v, want at least %v", info.NumMethods, tc.wantMinMethods)
			}
			if tc.wantIsPtr && info.ElemKind != reflect.Struct {
				t.Errorf("ElemKind = %v, want %v", info.ElemKind, reflect.Struct)
			}
		})
	}
}

func TestGetStructFields(t *testing.T) {
	t.Run("struct value", func(t *testing.T) {
		fields, err := GetStructFields(MyMath{Offset: 3.14})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fields) != 1 {
			t.Fatalf("expected 1 field, got %d", len(fields))
		}
		if fields[0].Name != "Offset" {
			t.Errorf("field name = %v, want Offset", fields[0].Name)
		}
		if fields[0].Kind != reflect.Float64 {
			t.Errorf("field kind = %v, want %v", fields[0].Kind, reflect.Float64)
		}
		if !fields[0].Exported {
			t.Error("expected field to be exported")
		}
		if fields[0].Index != 0 {
			t.Errorf("field index = %v, want 0", fields[0].Index)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		fields, err := GetStructFields(&MyMath{Offset: 1.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fields) != 1 {
			t.Fatalf("expected 1 field, got %d", len(fields))
		}
		if fields[0].Name != "Offset" {
			t.Errorf("field name = %v, want Offset", fields[0].Name)
		}
	})

	t.Run("multi-field struct", func(t *testing.T) {
		type MultiField struct {
			Name    string  `json:"name"`
			Age     int     `json:"age"`
			Balance float64
		}
		fields, err := GetStructFields(MultiField{Name: "test", Age: 25, Balance: 100.0})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(fields) != 3 {
			t.Fatalf("expected 3 fields, got %d", len(fields))
		}
		// Verify struct tags are captured via reflection
		if string(fields[0].Tag) != `json:"name"` {
			t.Errorf("tag = %v, want json:\"name\"", fields[0].Tag)
		}
		if fields[1].Kind != reflect.Int {
			t.Errorf("Age kind = %v, want %v", fields[1].Kind, reflect.Int)
		}
	})

	t.Run("non-struct error", func(t *testing.T) {
		_, err := GetStructFields("not a struct")
		if err == nil {
			t.Error("expected error for non-struct type")
		}
	})

	t.Run("non-struct slice error", func(t *testing.T) {
		_, err := GetStructFields([]int{1, 2, 3})
		if err == nil {
			t.Error("expected error for slice type")
		}
	})
}

func TestIsVariadic(t *testing.T) {
	variadicFunc := func(args ...int) int {
		sum := 0
		for _, a := range args {
			sum += a
		}
		return sum
	}

	t.Run("variadic function", func(t *testing.T) {
		result, err := IsVariadic(variadicFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected variadic function to return true")
		}
	})

	t.Run("non-variadic function", func(t *testing.T) {
		result, err := IsVariadic(AddInt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected non-variadic function to return false")
		}
	})

	t.Run("method is non-variadic", func(t *testing.T) {
		result, err := IsVariadic(MyMath.Add)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected method to return false")
		}
	})

	t.Run("non-func error", func(t *testing.T) {
		_, err := IsVariadic(42)
		if err == nil {
			t.Error("expected error for non-func type")
		}
	})

	t.Run("struct error", func(t *testing.T) {
		_, err := IsVariadic(MyMath{})
		if err == nil {
			t.Error("expected error for struct type")
		}
	})
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
