package handlers

import "testing"

func TestTemplateFuncMap_Add(t *testing.T) {
	fm := templateFuncMap()
	addFn := fm["add"].(func(...int) int)
	tests := []struct {
		name     string
		args     []int
		expected int
	}{
		{name: "no_args", args: []int{}, expected: 0},
		{name: "single", args: []int{5}, expected: 5},
		{name: "multiple", args: []int{1, 2, 3}, expected: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addFn(tt.args...)
			if got != tt.expected {
				t.Errorf("add(%v) = %d, want %d", tt.args, got, tt.expected)
			}
		})
	}
}

func TestTemplateFuncMap_Sub(t *testing.T) {
	fm := templateFuncMap()
	subFn := fm["sub"].(func(int, int) int)
	if got := subFn(10, 3); got != 7 {
		t.Errorf("sub(10, 3) = %d, want 7", got)
	}
}

func TestTemplateFuncMap_Mul(t *testing.T) {
	fm := templateFuncMap()
	mulFn := fm["mul"].(func(int, int) int)
	if got := mulFn(4, 5); got != 20 {
		t.Errorf("mul(4, 5) = %d, want 20", got)
	}
}

func TestTemplateFuncMap_Div(t *testing.T) {
	fm := templateFuncMap()
	divFn := fm["div"].(func(int, int) int)
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{name: "normal", a: 10, b: 3, expected: 3},
		{name: "div_by_zero", a: 10, b: 0, expected: 0},
		{name: "exact", a: 20, b: 5, expected: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := divFn(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("div(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}
