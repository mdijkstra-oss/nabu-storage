package config

import (
	"testing"

	th "nabu-storage/internal/lib/testutil"
)

func TestExpandHome(t *testing.T) {
	tests := []struct {
		Name     string
		Path     string
		Home     string
		Expected string
	}{
		{Name: "tilde prefix", Path: "~/Documents/nabu-persistence", Home: "/Users/tester", Expected: "/Users/tester/Documents/nabu-persistence"},
		{Name: "tilde prefix with trailing slash", Path: "~/persist/", Home: "/Users/tester", Expected: "/Users/tester/persist"},
		{Name: "bare tilde", Path: "~", Home: "/Users/tester", Expected: "/Users/tester"},
		{Name: "absolute path untouched", Path: "/var/lib/nabu", Home: "/Users/tester", Expected: "/var/lib/nabu"},
		{Name: "relative path untouched", Path: "./persist", Home: "/Users/tester", Expected: "./persist"},
		{Name: "tilde inside path untouched", Path: "/var/~/nabu", Home: "/Users/tester", Expected: "/var/~/nabu"},
		{Name: "tilde user untouched", Path: "~other/nabu", Home: "/Users/tester", Expected: "~other/nabu"},
		{Name: "unknown home falls back", Path: "~/persist", Home: "", Expected: fallbackProjectsDir},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			th.AssertEqual(t, expandHome(tt.Path, tt.Home), tt.Expected, "expanded path")
		})
	}
}
