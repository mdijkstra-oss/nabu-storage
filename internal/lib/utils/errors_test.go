package utils_test

import (
	"errors"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

type shouldInput struct {
	Value int
	Err   error
}

func TestShould(t *testing.T) {
	tests := []struct {
		Name     string
		Input    shouldInput
		Expected int
	}{
		{
			Name: "Returns value when no error",
			Input: shouldInput{
				Value: 42,
				Err:   nil,
			},
			Expected: 42,
		},
		{
			Name: "Returns value even when error present",
			Input: shouldInput{
				Value: 123,
				Err:   errors.New("some error"),
			},
			Expected: 123,
		},
		{
			Name: "Returns zero value when error and zero value given",
			Input: shouldInput{
				Value: 0,
				Err:   errors.New("error occurred"),
			},
			Expected: 0,
		},
	}

	th.RunFunctionTests(t, tests, func(input shouldInput) int {
		return utils.Should(input.Value, input.Err)
	})
}

type shouldStringInput struct {
	Value string
	Err   error
}

func TestShouldWithString(t *testing.T) {
	tests := []struct {
		Name     string
		Input    shouldStringInput
		Expected string
	}{
		{
			Name: "Returns string with no error",
			Input: shouldStringInput{
				Value: "hello",
				Err:   nil,
			},
			Expected: "hello",
		},
		{
			Name: "Returns string even with error",
			Input: shouldStringInput{
				Value: "world",
				Err:   errors.New("failed"),
			},
			Expected: "world",
		},
	}

	th.RunFunctionTests(t, tests, func(input shouldStringInput) string {
		return utils.Should(input.Value, input.Err)
	})
}

type shouldWorkInput struct {
	Err error
}

func TestShouldWork(t *testing.T) {
	tests := []struct {
		Name     string
		Input    shouldWorkInput
		Expected bool
	}{
		{
			Name:     "Does not panic with nil error",
			Input:    shouldWorkInput{Err: nil},
			Expected: true,
		},
		{
			Name:     "Does not panic with error present",
			Input:    shouldWorkInput{Err: errors.New("test error")},
			Expected: true,
		},
	}

	th.RunFunctionTests(t, tests, func(input shouldWorkInput) bool {
		defer func() {
			if r := recover(); r != nil {
				panic(r)
			}
		}()
		utils.ShouldWork(input.Err)
		return true
	})
}
