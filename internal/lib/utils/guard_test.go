package utils

import "testing"

func TestGuardWith(t *testing.T) {
	tests := []struct {
		Name        string
		Fn          func()
		ShouldPanic bool
	}{
		{
			Name:        "normal execution",
			Fn:          func() {},
			ShouldPanic: false,
		},
		{
			Name:        "recovers from panic",
			Fn:          func() { panic("test panic") },
			ShouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			panicked := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				GuardWith(tt.Fn, "test", "context")
			}()

			if panicked != tt.ShouldPanic {
				t.Errorf("expected outer panic=%v, got panic=%v", tt.ShouldPanic, panicked)
			}
		})
	}
}
