package projection

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"os"
	"reflect"
	"time"
	"unsafe"
)

var timeType = reflect.TypeOf(time.Time{})

func WithImmutabilityCheck[T Entity](reducer Reducer[T]) Reducer[T] {
	if !isDevMode() {
		return reducer
	}

	return func(current *T, event *commands.AnyMessage) *T {
		result := reducer(current, event)

		if IsReplayMode() {
			return result
		}

		if current != nil && result != nil && current != result {
			if hasSharedState(current, result) {
				panic(fmt.Sprintf(
					"IMMUTABILITY VIOLATION: reducer returned new entity but shares memory with input. Action=%s AggregateType=%s",
					event.Action,
					event.AggregateType,
				))
			}
		}

		return result
	}
}

func isDevMode() bool {
	return os.Getenv("HERMES_DEV") == "true"
}

func HasSharedState(before, after any) bool {
	return checkSharedMemory(reflect.ValueOf(before), reflect.ValueOf(after), make(map[uintptr]bool))
}

func hasSharedState(before, after any) bool {
	return HasSharedState(before, after)
}

func checkSharedMemory(before, after reflect.Value, visited map[uintptr]bool) bool {
	if !before.IsValid() || !after.IsValid() {
		return false
	}

	if before.Type() != after.Type() {
		return false
	}

	if before.Type() == timeType {
		return false
	}

	switch before.Kind() {
	case reflect.Ptr:
		if before.IsNil() || after.IsNil() {
			return false
		}
		return checkSharedMemory(before.Elem(), after.Elem(), visited)

	case reflect.Struct:
		for i := 0; i < before.NumField(); i++ {
			if checkSharedMemory(before.Field(i), after.Field(i), visited) {
				return true
			}
		}
		return false

	case reflect.Map:
		return mapsShareMemory(before, after, visited)

	case reflect.Slice:
		return slicesShareMemory(before, after, visited)

	default:
		return false
	}
}

func mapsShareMemory(before, after reflect.Value, visited map[uintptr]bool) bool {
	if before.IsNil() || after.IsNil() {
		return false
	}

	beforePtr := before.UnsafePointer()
	afterPtr := after.UnsafePointer()

	if beforePtr == afterPtr {
		return true
	}

	addr := uintptr(beforePtr)
	if visited[addr] {
		return false
	}
	visited[addr] = true

	return false
}

func slicesShareMemory(before, after reflect.Value, visited map[uintptr]bool) bool {
	if before.IsNil() || after.IsNil() {
		return false
	}

	if before.Len() == 0 || after.Len() == 0 {
		return false
	}

	beforeData := unsafe.Pointer(before.UnsafePointer())
	afterData := unsafe.Pointer(after.UnsafePointer())

	if beforeData == afterData {
		return true
	}

	addr := uintptr(beforeData)
	if visited[addr] {
		return false
	}
	visited[addr] = true

	for i := 0; i < before.Len() && i < after.Len(); i++ {
		if checkSharedMemory(before.Index(i), after.Index(i), visited) {
			return true
		}
	}

	return false
}
