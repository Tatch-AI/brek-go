package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Tatch-AI/brek-go"
)

func TestMainSuccessAndErrorPaths(t *testing.T) {
	oldArgs := os.Args
	oldExit := exitFunc
	defer func() {
		os.Args = oldArgs
		exitFunc = oldExit
	}()

	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)

	brek.SetLoaders(nil)
	writeJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})

	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
		if code != 1 {
			t.Fatalf("exit code = %d, want 1", code)
		}
	}

	os.Args = []string{"brek", "write-types"}
	main()

	if exitCalled {
		t.Fatal("did not expect exit on successful command")
	}
	if _, err := os.Stat(filepath.Join(tmp, "Config.d.ts")); err != nil {
		t.Fatalf("expected Config.d.ts to be written: %v", err)
	}

	exitCalled = false
	os.Args = []string{"brek", "unknown"}
	main()

	if !exitCalled {
		t.Fatal("expected exit on invalid command")
	}
}

func writeJSON(t *testing.T, baseDir, rel string, value any) {
	t.Helper()

	path := filepath.Join(baseDir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
