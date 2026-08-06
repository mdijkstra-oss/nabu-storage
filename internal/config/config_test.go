package config

import (
	"os"
	"path/filepath"
	"testing"

	th "nabu-storage/internal/lib/testutil"
)

func TestExpandHome(t *testing.T) {
	tests := []struct {
		Name     string
		Path     string
		Home     string
		Expected string
		Error    string
	}{
		{Name: "tilde prefix", Path: "~/Documents/nabu-persistence", Home: "/Users/tester", Expected: "/Users/tester/Documents/nabu-persistence"},
		{Name: "tilde prefix with trailing slash", Path: "~/persist/", Home: "/Users/tester", Expected: "/Users/tester/persist"},
		{Name: "bare tilde", Path: "~", Home: "/Users/tester", Expected: "/Users/tester"},
		{Name: "absolute path untouched", Path: "/var/lib/nabu", Home: "/Users/tester", Expected: "/var/lib/nabu"},
		{Name: "relative path untouched", Path: "./persist", Home: "/Users/tester", Expected: "./persist"},
		{Name: "tilde inside path untouched", Path: "/var/~/nabu", Home: "/Users/tester", Expected: "/var/~/nabu"},
		{Name: "tilde user untouched", Path: "~other/nabu", Home: "/Users/tester", Expected: "~other/nabu"},
		{Name: "unknown home fails", Path: "~/persist", Home: "", Error: "no home directory"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got, err := expandHome(tt.Path, tt.Home)
			th.AssertError(t, err, tt.Error, "expand home")
			th.AssertEqual(t, got, tt.Expected, "expanded path")
		})
	}
}

func TestProjectsDir(t *testing.T) {
	existing := t.TempDir()
	file := filepath.Join(existing, "a-file")
	th.AssertError(t, os.WriteFile(file, []byte("x"), 0644), "", "write fixture file")

	tests := []struct {
		Name     string
		Value    string
		Set      bool
		Expected string
		Error    string
	}{
		{Name: "unset", Set: false, Error: "PERSISTENCE_DIR is not set"},
		{Name: "empty", Value: "", Set: true, Error: "PERSISTENCE_DIR is not set"},
		{Name: "writable directory", Value: existing, Set: true, Expected: existing},
		{Name: "relative path", Value: "./persist", Set: true, Error: "must be an absolute path"},
		{Name: "missing directory", Value: filepath.Join(existing, "absent"), Set: true, Error: "directory does not exist"},
		{Name: "path is a file", Value: file, Set: true, Error: "not a directory"},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Set {
				t.Setenv("PERSISTENCE_DIR", tt.Value)
			} else {
				th.AssertError(t, os.Unsetenv("PERSISTENCE_DIR"), "", "unset PERSISTENCE_DIR")
			}

			got, err := projectsDir()
			th.AssertError(t, err, tt.Error, "projects dir")
			th.AssertEqual(t, got, tt.Expected, "projects dir")
		})
	}
}

func TestCheckWritableDirRejectsReadOnly(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "read-only")
	th.AssertError(t, os.Mkdir(dir, 0500), "", "create read-only dir")

	th.AssertError(t, checkWritableDir(dir), "not writable", "read-only dir")
}
