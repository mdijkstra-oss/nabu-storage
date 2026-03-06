package utils

import (
	"reflect"
	"testing"
)

func TestMap(t *testing.T) {
	tests := []struct {
		Name     string
		Input    []int
		Expected []int
	}{
		{Name: "doubles values", Input: []int{1, 2, 3}, Expected: []int{2, 4, 6}},
		{Name: "empty slice", Input: []int{}, Expected: []int{}},
		{Name: "nil slice returns nil", Input: nil, Expected: nil},
		{Name: "single element", Input: []int{5}, Expected: []int{10}},
	}

	double := func(x int) int { return x * 2 }

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := Map(tt.Input, double)
			if !reflect.DeepEqual(got, tt.Expected) {
				t.Errorf("Map() = %v, want %v", got, tt.Expected)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		Name     string
		Input    []int
		Expected []int
	}{
		{Name: "filters even numbers", Input: []int{1, 2, 3, 4, 5, 6}, Expected: []int{2, 4, 6}},
		{Name: "empty slice", Input: []int{}, Expected: []int{}},
		{Name: "no matches", Input: []int{1, 3, 5}, Expected: []int{}},
		{Name: "all match", Input: []int{2, 4, 6}, Expected: []int{2, 4, 6}},
	}

	isEven := func(x int) bool { return x%2 == 0 }

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := Filter(tt.Input, isEven)
			if !reflect.DeepEqual(got, tt.Expected) {
				t.Errorf("Filter() = %v, want %v", got, tt.Expected)
			}
		})
	}
}
