package reflectutil

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customByte byte
type customBytes []byte
type customStruct struct {
	intS     int
	float32S float32
	boolS    bool
}
type structWithSelfPtr struct {
	p *structWithSelfPtr
	s string
}
type loop *loop

var loop1, loop2 loop

func filterInt(v1, v2 any) bool {
	ref1 := v1.(reflect.Value)
	ref2 := v2.(reflect.Value)
	if ref1.Kind() != reflect.Int || ref2.Kind() != reflect.Int {
		return false
	}
	return true
}

func filterString(v1, v2 any) bool {
	ref1 := v1.(reflect.Value)
	ref2 := v2.(reflect.Value)
	if ref1.Kind() == reflect.String || ref2.Kind() == reflect.String {
		return false
	}
	if ref1.Kind() == reflect.Interface && reflect.TypeOf(ref1.Interface()).Kind() == reflect.String {
		return false
	}
	if ref2.Kind() == reflect.Interface && reflect.TypeOf(ref2.Interface()).Kind() == reflect.String {
		return false
	}
	return true
}

func filterSlice(v1, v2 any) bool {
	ref1 := v1.(reflect.Value)
	ref2 := v2.(reflect.Value)
	if ref1.Kind() != reflect.Slice || ref2.Kind() != reflect.Slice {
		return false
	}
	return true
}

func filterPointer(v1, v2 any) bool {
	ref1 := v1.(reflect.Value)
	ref2 := v2.(reflect.Value)
	if ref1.Kind() != reflect.Pointer || ref2.Kind() != reflect.Pointer {
		return false
	}
	return true
}

func TestDeepEquals(t *testing.T) {
	var (
		intS  int = -3
		intS2 int = -3
	)
	var intSP *int
	var (
		uintS  uint = 3
		uintS2 uint = 3
	)
	var (
		stringS  string = "string"
		stringS2 string = "string"
	)
	var (
		float32S  float32 = 3.141569
		float32S2 float32 = 3.141569
	)
	var (
		float64S  float64 = 3.141569
		float64S2 float64 = 3.141569
	)

	var structS1 = customStruct{
		intS:     3,
		float32S: 3.141569,
		boolS:    true,
	}
	var structS2 = customStruct{
		intS:     3,
		float32S: 3.141569,
		boolS:    true,
	}
	var structS1P = &structS1
	var structS2P = &structS2
	var func1 = func(v1, v2 int) bool {
		return v1 == v2
	}
	var func2 = func(v1, v2 int) bool {
		return v1 == v2
	}
	var funcp1 func()
	var funcp2 func()
	var (
		any1 any
		any2 any
	)

	var tests = []struct {
		name       string
		arg1       interface{}
		arg2       interface{}
		filterFunc FilterFunc
		expected   bool
	}{
		// Equalities
		// primary types: same types and different types(but same value)
		{"int to int", int32(-3), int32(-3), nil, true},
		{"uint to uint", uint32(3), uint32(3), nil, true},
		{"float to float", float32(3.141569), float32(3.141569), nil, true},
		{"string to string", "string", "string", nil, true},
		{"bool to bool", true, true, nil, true},
		// pointers: single pointer, double pointer, function pointer, struct pointer, nil
		{"single pointer, same address, int", &intS, &intS, nil, true},
		{"single pointer, different address, int", &intS, &intS2, nil, true},
		{"single pointer, same address, uint", &uintS, &uintS, nil, true},
		{"single pointer, different address, uint", &uintS, &uintS2, nil, true},
		{"single pointer, same address, string", &stringS, &stringS, nil, true},
		{"single pointer, different address, string", &stringS, &stringS2, nil, true},
		{"single pointer, same address, float32", &float32S, &float32S, nil, true},
		{"single pointer, different address, float32", &float32S, &float32S2, nil, true},
		{"single pointer, same address, float64", &float64S, &float64S, nil, true},
		{"single pointer, different address, float64", &float64S, &float64S2, nil, true},
		{"double pointer", &structS1P, &structS2P, nil, true},
		{"nil", nil, nil, nil, true},
		// struct
		{"struct", structS1, structS2, nil, true},
		{"struct pointer", &structS1, &structS2, nil, true},
		{"struct with self pointer", &structWithSelfPtr{p: &structWithSelfPtr{s: "a"}}, &structWithSelfPtr{p: &structWithSelfPtr{s: "a"}}, nil, true},
		// map
		{"map", map[int]string{1: "one", 2: "two"}, map[int]string{1: "one", 2: "two"}, nil, true},
		{"map pointer", &map[int]string{1: "one", 2: "two"}, &map[int]string{2: "two", 1: "one"}, nil, true},
		// func
		{"func pointer", funcp1, funcp2, nil, true},
		// slice
		{"slice", []byte{1, 2, 3}, []byte{1, 2, 3}, nil, true},
		{"double slice", [][]byte{{1, 2, 3}}, [][]byte{{1, 2, 3}}, nil, true},
		{"slice, type", []customByte{1, 2, 3}, []customByte{1, 2, 3}, nil, true},
		{"slice, type", customBytes{1, 2, 3}, customBytes{1, 2, 3}, nil, true},

		// Inequalities
		// primary types: same types and different types(but same value)
		{"int to int", int32(-3), int32(-4), nil, false},
		{"uint to uint", uint32(3), uint32(4), nil, false},
		{"float to float", float32(3.141569), float64(3.141569), nil, false},
		{"string to string", "string", "String", nil, false},
		{"bool to bool", true, false, nil, false},
		{"int to uint", int(314), uint(314), nil, false},
		{"int to string", int(314), "314", nil, false},
		{"int to bool", 1, true, nil, false},
		{"int to bool", 0, false, nil, false},
		{"int to float", int(3), float32(3), nil, false},
		// pointers
		{"single pointer, different address, int to uint", &intS, &uintS, nil, false},
		{"single pointer, different address, float32 to float64", &float32S, &float64S, nil, false},
		{"single pointer, different address, *int to nil", intSP, nil, nil, false},
		{"double pointer", &intSP, nil, nil, false},
		// struct
		{"struct", customStruct{1, 3.141569, false}, customStruct{1, 3.141569, true}, nil, false},
		{"struct pointer", &customStruct{1, 3.141569, false}, &customStruct{1, 3.141569, true}, nil, false},
		{"struct with self pointer", &structWithSelfPtr{p: &structWithSelfPtr{s: "a"}}, &structWithSelfPtr{p: &structWithSelfPtr{s: "b"}}, nil, false},
		// map
		{"map", map[int]string{1: "one", 2: "two"}, map[int]string{1: "One", 2: "Two"}, nil, false},
		{"map pointer", &map[int]string{1: "one", 2: "two"}, &map[int]string{2: "Two", 1: "One"}, nil, false},
		{"map with different type", map[int]string{1: "one", 2: "two"}, map[uint]string{1: "one", 2: "two"}, nil, false},
		// func
		{"function pointer", func1, func2, nil, false},
		// slice
		{"slice", []byte{1, 2, 3}, []byte{0, 2, 3}, nil, false},
		{"slice, different len", []byte{1, 2, 3}, []byte{2, 3}, nil, false},
		{"double slice", [][]byte{{1, 2, 3}}, [][]byte{{0, 2, 3}}, nil, false},
		{"slice, same type", []customByte{1, 2, 3}, []customByte{0, 2, 3}, nil, false},
		{"slice, same type", customBytes{1, 2, 3}, customBytes{0, 2, 3}, nil, false},
		{"slice, different type", []customByte{1, 2, 3}, []byte{1, 2, 3}, nil, false},
		{"slice, different type", customBytes{1, 2, 3}, []customByte{1, 2, 3}, nil, false},

		// Floating points
		{"NaN", math.NaN(), math.NaN(), nil, false},
		{"NaN slice", []float64{math.NaN()}, []float64{math.NaN()}, nil, false},
		{"NaN slice pointer", &[1]float64{math.NaN()}, &[1]float64{math.NaN()}, nil, false},

		// Empty
		{"empty slice", []int{}, []int{}, nil, true},
		{"nil slice", []int(nil), []int(nil), nil, true},
		{"nil slice and empty", []int{}, []int(nil), nil, false},
		{"empty map", map[int]int{}, map[int]int{}, nil, true},
		{"nil map", map[int]int(nil), map[int]int(nil), nil, true},
		{"nil map and empty", map[int]int{}, map[int]int(nil), nil, false},

		// Others
		{"loop", &loop1, &loop2, nil, true},
		{"any", &any1, &any2, nil, true},

		// Filter func
		{"filter, different type", 1, "one", filterInt, false},
		{"filter, different type, reverse", "one", 1, filterInt, false},
		{"filter slice", []byte{1, 2, 3}, []byte{0, 2, 3}, filterSlice, true},
		{"filter map", map[int]any{1: "one", 2: "two"}, map[int]any{1: "one", 2: 2}, filterString, true},
		{"filter pointer", &map[int]string{1: "one"}, &map[int]string{2: "one"}, filterPointer, true},
		{"filter struct", customStruct{1, 3.141569, true}, customStruct{0, 3.141569, true}, filterInt, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.filterFunc == nil {
				assert.Equal(t, tt.expected, reflect.DeepEqual(tt.arg1, tt.arg2))
			}
			assert.Equal(t, tt.expected, DeepEqual(tt.arg1, tt.arg2, tt.filterFunc))
		})
	}
}
