package brek

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func resetTestState() {
	cacheMu.Lock()
	cachedConfig = nil
	cacheMu.Unlock()
	SetLoaders(nil)
}

func writeTestJSON(t *testing.T, baseDir, rel string, value any) {
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

func readTestJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	value, err := readJSONFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return value
}

func TestPathHelpers(t *testing.T) {
	t.Setenv("BREK_CONFIG_DIR", "/tmp/brek-config")
	t.Setenv("BREK_WRITE_DIR", "/tmp/brek-write")
	t.Setenv("BREK_LOADERS_FILE_PATH", "/tmp/loaders.go")
	t.Cleanup(resetTestState)

	if got := ConfigDir(); got != "/tmp/brek-config" {
		t.Fatalf("ConfigDir() = %q", got)
	}
	if got := WriteDir(); got != "/tmp/brek-write" {
		t.Fatalf("WriteDir() = %q", got)
	}
	if got := LoadersFilePath(); got != "/tmp/loaders.go" {
		t.Fatalf("LoadersFilePath() = %q", got)
	}
	if got := ConfigJSONPath(); got != filepath.Join("/tmp/brek-write", "config.json") {
		t.Fatalf("ConfigJSONPath() = %q", got)
	}
}

func TestDefaultLoaders(t *testing.T) {
	loaders := DefaultLoaders()
	if _, ok := loaders["awsSecret"]; !ok {
		t.Fatal("expected awsSecret to be included in default loaders")
	}
}

func TestPathHelpersDefaults(t *testing.T) {
	t.Cleanup(resetTestState)
	t.Setenv("BREK_CONFIG_DIR", "")
	t.Setenv("BREK_WRITE_DIR", "")
	t.Setenv("BREK_LOADERS_FILE_PATH", "")

	if got := envOr("BREK_TEST_UNSET", "fallback"); got != "fallback" {
		t.Fatalf("envOr() = %q, want fallback", got)
	}
	if got := ConfigDir(); got != "config" {
		t.Fatalf("ConfigDir() = %q, want config", got)
	}
	if got := WriteDir(); got != "config" {
		t.Fatalf("WriteDir() = %q, want config", got)
	}
	if got := LoadersFilePath(); got != "brek.loaders.js" {
		t.Fatalf("LoadersFilePath() = %q, want brek.loaders.js", got)
	}
}

func TestIsLoaderAndEnvironmentVariable(t *testing.T) {
	if !IsLoader(map[string]any{"[foo]": "bar"}) {
		t.Fatal("expected loader object to be recognized")
	}

	if IsLoader(map[string]any{"foo": "bar"}) {
		t.Fatal("expected non-loader object to be rejected")
	}

	if IsLoader(map[string]any{}) {
		t.Fatal("expected empty object to be rejected")
	}

	if IsLoader(map[string]any{"[foo]": "bar", "[bar]": "baz"}) {
		t.Fatal("expected multi-key object to be rejected")
	}

	if !IsEnvironmentVariable("${FOO}") {
		t.Fatal("expected env variable syntax to be recognized")
	}

	if IsEnvironmentVariable("${1FOO}") {
		t.Fatal("expected invalid env variable syntax to be rejected")
	}
}

func TestLoaderHelpers(t *testing.T) {
	if got := loaderName(map[string]any{"[foo]": "bar"}); got != "foo" {
		t.Fatalf("loaderName() = %q", got)
	}
	if got := loaderName(map[string]any{}); got != "" {
		t.Fatalf("loaderName(empty) = %q", got)
	}
	if got := availableLoaderNames(nil); got != nil {
		t.Fatalf("availableLoaderNames(nil) = %#v", got)
	}
}

func TestGetEnvArgumentsAndOverrides(t *testing.T) {
	t.Setenv("ENVIRONMENT", "prod")
	t.Setenv("NODE_ENV", "dev")
	t.Setenv("DEPLOYMENT", "blue")
	t.Setenv("USER", "alice")
	t.Setenv("BREK", `{"nested":{"count":3}}`)
	t.Cleanup(resetTestState)

	args, err := GetEnvArguments()
	if err != nil {
		t.Fatalf("GetEnvArguments() error = %v", err)
	}

	want := EnvArguments{
		Environment: "prod",
		Deployment:  "blue",
		User:        "alice",
		Overrides:   map[string]any{"nested": map[string]any{"count": 3}},
	}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("GetEnvArguments() = %#v, want %#v", args, want)
	}

	t.Setenv("BREK", "")
	t.Setenv("OVERRIDE", `{"via":"override"}`)

	overrides, err := GetEnvOverrides()
	if err != nil {
		t.Fatalf("GetEnvOverrides() error = %v", err)
	}

	if !reflect.DeepEqual(overrides, map[string]any{"via": "override"}) {
		t.Fatalf("GetEnvOverrides() = %#v", overrides)
	}
}

func TestGetEnvArgumentsInvalidOverrides(t *testing.T) {
	t.Setenv("BREK", "{broken")
	t.Cleanup(resetTestState)

	_, err := GetEnvArguments()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetEnvOverridesEmpty(t *testing.T) {
	t.Setenv("BREK", "")
	t.Setenv("OVERRIDE", "")
	t.Cleanup(resetTestState)

	overrides, err := GetEnvOverrides()
	if err != nil {
		t.Fatalf("GetEnvOverrides() error = %v", err)
	}
	if len(overrides) != 0 {
		t.Fatalf("GetEnvOverrides() = %#v, want empty", overrides)
	}
}

func TestGetEnvOverridesNonObject(t *testing.T) {
	t.Setenv("BREK", `[]`)
	t.Cleanup(resetTestState)

	_, err := GetEnvOverrides()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetEnvOverridesInvalidJSON(t *testing.T) {
	t.Setenv("BREK", "{not-json")
	t.Cleanup(resetTestState)

	_, err := GetEnvOverrides()
	if err == nil {
		t.Fatal("expected error")
	}

	var invalid InvalidConf
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidConf, got %T", err)
	}
	if got := err.Error(); got != "INVALID_CONF: CLI overrides (BREK/OVERRIDE) is not valid JSON" {
		t.Fatalf("unexpected error string: %q", got)
	}
}

func TestLoadConfFileAndFromFiles(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)

	writeTestJSON(t, tmp, "default.json", map[string]any{
		"foo":    "default",
		"nested": map[string]any{"defaultOnly": true},
	})
	writeTestJSON(t, tmp, filepath.Join("environments", "prod.json"), map[string]any{
		"foo": "environment",
		"nested": map[string]any{
			"envOnly": true,
		},
	})
	writeTestJSON(t, tmp, filepath.Join("deployments", "blue.json"), map[string]any{
		"nested": map[string]any{
			"deploymentOnly": 1,
		},
	})
	writeTestJSON(t, tmp, filepath.Join("users", "alice.json"), map[string]any{
		"foo": "user",
	})

	sources, err := LoadConfFromFiles(EnvArguments{
		Environment: "prod",
		Deployment:  "blue",
		User:        "alice",
		Overrides:   map[string]any{"foo": "override"},
	})
	if err != nil {
		t.Fatalf("LoadConfFromFiles() error = %v", err)
	}

	got := MergeConfs(sources)
	want := map[string]any{
		"foo": "override",
		"nested": map[string]any{
			"defaultOnly":    true,
			"deploymentOnly": 1,
			"envOnly":        true,
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeConfs() = %#v, want %#v", got, want)
	}
}

func TestLoadConfFromFilesDefaultOnly(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)

	writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})

	sources, err := LoadConfFromFiles(EnvArguments{})
	if err != nil {
		t.Fatalf("LoadConfFromFiles() error = %v", err)
	}

	if !reflect.DeepEqual(sources, ConfSources{
		Default:     map[string]any{"foo": "bar"},
		Environment: map[string]any{},
		Deployment:  map[string]any{},
		User:        map[string]any{},
		Overrides:   nil,
	}) {
		t.Fatalf("LoadConfFromFiles() = %#v", sources)
	}
}

func TestLoadConfFromFilesErrorBranches(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)

	t.Run("environment", func(t *testing.T) {
		writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})
		if err := os.MkdirAll(filepath.Join(tmp, "environments"), 0o755); err != nil {
			t.Fatalf("mkdir environments: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "environments", "prod.json"), []byte("{broken"), 0o644); err != nil {
			t.Fatalf("write invalid environment file: %v", err)
		}
		if _, err := LoadConfFromFiles(EnvArguments{Environment: "prod"}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("deployment", func(t *testing.T) {
		writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})
		if err := os.MkdirAll(filepath.Join(tmp, "deployments"), 0o755); err != nil {
			t.Fatalf("mkdir deployments: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "deployments", "blue.json"), []byte("{broken"), 0o644); err != nil {
			t.Fatalf("write invalid deployment file: %v", err)
		}
		if _, err := LoadConfFromFiles(EnvArguments{Deployment: "blue"}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("user", func(t *testing.T) {
		writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})
		if err := os.MkdirAll(filepath.Join(tmp, "users"), 0o755); err != nil {
			t.Fatalf("mkdir users: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmp, "users", "alice.json"), []byte("{broken"), 0o644); err != nil {
			t.Fatalf("write invalid user file: %v", err)
		}
		if _, err := LoadConfFromFiles(EnvArguments{User: "alice"}); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestLoadConfFileMissingAndInvalid(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)

	got, err := LoadConfFile("missing.json")
	if err != nil {
		t.Fatalf("LoadConfFile missing returned error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("LoadConfFile missing = %#v, want empty map", got)
	}

	writeTestJSON(t, tmp, "bad.json", map[string]any{})
	if err := os.WriteFile(filepath.Join(tmp, "bad.json"), []byte("{not-json"), 0o644); err != nil {
		t.Fatalf("write invalid json: %v", err)
	}

	_, err = LoadConfFile("bad.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	var invalid InvalidConf
	if !errors.As(err, &invalid) {
		t.Fatalf("expected InvalidConf, got %T", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "nonobject.json"), []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write non-object json: %v", err)
	}
	if _, err := LoadConfFile("nonobject.json"); err == nil {
		t.Fatal("expected error for non-object JSON")
	}
}

func TestJSONUtilities(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()

	if _, err := readJSONFile(filepath.Join(tmp, "missing.json")); err == nil {
		t.Fatal("expected missing file error")
	}

	nonObjectPath := filepath.Join(tmp, "nonobject-read.json")
	if err := os.WriteFile(nonObjectPath, []byte(`[]`), 0o644); err != nil {
		t.Fatalf("write non-object file: %v", err)
	}
	if _, err := readJSONFile(nonObjectPath); err == nil {
		t.Fatal("expected non-object read error")
	}

	emptyPath := filepath.Join(tmp, "empty.json")
	if err := os.WriteFile(emptyPath, []byte("   "), 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}
	if _, err := readJSONFile(emptyPath); err == nil {
		t.Fatal("expected invalid JSON error for empty file")
	}

	if got := normalizeJSONValue(json.Number("1.5")); got != 1.5 {
		t.Fatalf("normalizeJSONValue() = %#v, want 1.5", got)
	}

	decoded, err := decodeJSON([]byte(`{"a":1} {"b":2}`))
	if err == nil {
		t.Fatalf("expected extra content error, got %#v", decoded)
	}

	_, err = decodeJSON([]byte(`{"a":1.5,"b":[1,2.5]}`))
	if err != nil {
		t.Fatalf("decodeJSON() float error = %v", err)
	}

	badPath := filepath.Join(tmp, "marshal", "bad.json")
	if err := writeJSONFile(badPath, map[string]any{"bad": make(chan int)}); err == nil {
		t.Fatal("expected marshal error")
	}

	if err := os.WriteFile(filepath.Join(tmp, "occupied"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write occupied file: %v", err)
	}
	if err := writeJSONFile(filepath.Join(tmp, "occupied", "child.json"), map[string]any{"ok": true}); err == nil {
		t.Fatal("expected mkdir error")
	}

	if got := normalizeJSONValue(json.Number("not-a-number")); got != "not-a-number" {
		t.Fatalf("normalizeJSONValue() = %#v, want string fallback", got)
	}
}

func TestMergeConfs(t *testing.T) {
	t.Run("basic precedence", func(t *testing.T) {
		got := MergeConfs(ConfSources{
			Default: map[string]any{
				"foo1": "default",
				"foo2": "default",
				"foo3": "default",
				"foo4": "default",
				"foo5": "default",
			},
			Environment: map[string]any{
				"foo2": "environment",
				"foo3": "environment",
				"foo4": "environment",
				"foo5": "environment",
			},
			Deployment: map[string]any{
				"foo3": "deployment",
				"foo4": "deployment",
				"foo5": "deployment",
			},
			User: map[string]any{
				"foo4": "user",
				"foo5": "user",
			},
			Overrides: map[string]any{
				"foo5": "overrides",
			},
		})

		want := map[string]any{
			"foo1": "default",
			"foo2": "environment",
			"foo3": "deployment",
			"foo4": "user",
			"foo5": "overrides",
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("MergeConfs() = %#v, want %#v", got, want)
		}
	})

	t.Run("loaders are not merged", func(t *testing.T) {
		got := MergeConfs(ConfSources{
			Default: map[string]any{
				"foo1": map[string]any{
					"[foo1]": map[string]any{
						"param1": "default",
						"param2": "default",
					},
				},
			},
			Environment: map[string]any{
				"foo1": map[string]any{
					"[foo1]": map[string]any{
						"param1": "environment",
					},
				},
			},
		})

		want := map[string]any{
			"foo1": map[string]any{
				"[foo1]": map[string]any{
					"param1": "environment",
				},
			},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("MergeConfs() = %#v, want %#v", got, want)
		}
	})

	t.Run("arrays are not merged", func(t *testing.T) {
		got := MergeConfs(ConfSources{
			Default: map[string]any{
				"foo1": []any{1, 2, 3},
			},
			Environment: map[string]any{
				"foo1": []any{4, 5, 6},
			},
		})

		want := map[string]any{
			"foo1": []any{4, 5, 6},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("MergeConfs() = %#v, want %#v", got, want)
		}
	})
}

func TestMergeConfigsNil(t *testing.T) {
	if got := mergeConfigs(nil, nil); len(got) != 0 {
		t.Fatalf("mergeConfigs(nil, nil) = %#v", got)
	}
}

func TestResolveConf(t *testing.T) {
	t.Run("returns unchanged config when nothing matches", func(t *testing.T) {
		input := map[string]any{
			"foo": "bar",
			"arr": []any{1, 2, 3},
			"obj": map[string]any{"a": 1, "b": 2},
		}

		got, err := ResolveConf(input, LoaderDict{})
		if err != nil {
			t.Fatalf("ResolveConf() error = %v", err)
		}

		if !reflect.DeepEqual(got, input) {
			t.Fatalf("ResolveConf() = %#v, want %#v", got, input)
		}
	})

	t.Run("resolves env vars but not array string items", func(t *testing.T) {
		t.Setenv("FOO", "bar123!")
		input := map[string]any{
			"foo": "${FOO}",
			"obj": map[string]any{
				"a": "${FOO}",
				"b": map[string]any{"c": "${FOO}"},
			},
			"arr": []any{"${FOO}", map[string]any{"nested": "${FOO}"}},
		}

		got, err := ResolveConf(input, LoaderDict{})
		if err != nil {
			t.Fatalf("ResolveConf() error = %v", err)
		}

		want := map[string]any{
			"foo": "bar123!",
			"obj": map[string]any{
				"a": "bar123!",
				"b": map[string]any{"c": "bar123!"},
			},
			"arr": []any{"${FOO}", map[string]any{"nested": "bar123!"}},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ResolveConf() = %#v, want %#v", got, want)
		}
	})

	t.Run("resolves loaders", func(t *testing.T) {
		loaders := LoaderDict{
			"foo": func(arg any) (string, error) {
				obj, ok := arg.(map[string]any)
				if !ok {
					t.Fatalf("unexpected arg type %T", arg)
				}
				return "foo_" + obj["a"].(string), nil
			},
			"bar": func(arg any) (string, error) {
				return "bar_" + arg.(string), nil
			},
		}

		input := map[string]any{
			"a": map[string]any{"[foo]": map[string]any{"a": "123"}},
			"b": map[string]any{"a": map[string]any{"[bar]": "321"}},
		}

		got, err := ResolveConf(input, loaders)
		if err != nil {
			t.Fatalf("ResolveConf() error = %v", err)
		}

		want := map[string]any{
			"a": "foo_123",
			"b": map[string]any{"a": "bar_321"},
		}

		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ResolveConf() = %#v, want %#v", got, want)
		}
	})

	t.Run("missing loader returns error", func(t *testing.T) {
		_, err := ResolveConf(map[string]any{
			"secret": map[string]any{"[missing]": "value"},
		}, LoaderDict{"other": func(arg any) (string, error) { return "ok", nil }})
		if err == nil {
			t.Fatal("expected error")
		}
		var notFound LoaderNotFound
		if !errors.As(err, &notFound) {
			t.Fatalf("expected LoaderNotFound, got %T", err)
		}
		if !strings.Contains(err.Error(), `Available loaders: other`) {
			t.Fatalf("unexpected error string: %q", err.Error())
		}
	})

	t.Run("loader error propagates", func(t *testing.T) {
		_, err := ResolveConf(map[string]any{
			"secret": map[string]any{"[bad]": "value"},
		}, LoaderDict{
			"bad": func(arg any) (string, error) {
				return "", errors.New("boom")
			},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("array loader error propagates", func(t *testing.T) {
		_, err := ResolveConf([]any{
			map[string]any{"[bad]": "value"},
		}, LoaderDict{
			"bad": func(arg any) (string, error) {
				return "", errors.New("boom")
			},
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("resolves nested arrays and primitive root values", func(t *testing.T) {
		t.Setenv("FOO", "baz")
		loaders := LoaderDict{}
		got, err := ResolveConf([]any{[]any{1, map[string]any{"nested": "${FOO}"}}}, loaders)
		if err != nil {
			t.Fatalf("ResolveConf() error = %v", err)
		}
		want := []any{[]any{1, map[string]any{"nested": "baz"}}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("ResolveConf() = %#v, want %#v", got, want)
		}

		primitive, err := ResolveConf(99, loaders)
		if err != nil {
			t.Fatalf("ResolveConf() primitive error = %v", err)
		}
		if primitive != 99 {
			t.Fatalf("ResolveConf() primitive = %#v, want 99", primitive)
		}

		nilValue, err := ResolveConf(nil, loaders)
		if err != nil {
			t.Fatalf("ResolveConf() nil error = %v", err)
		}
		if nilValue != nil {
			t.Fatalf("ResolveConf() nil = %#v, want nil", nilValue)
		}
	})
}

func TestGenerateTypeDefAndProps(t *testing.T) {
	t.Run("flat object", func(t *testing.T) {
		got := GenerateTypeDef(map[string]any{
			"foo":       "bar",
			"count":     42,
			"isEnabled": true,
		})

		want := `
import {Config} from "brek";
declare module "brek" {
    export interface Config {
        'count': number
        'foo': string
        'isEnabled': boolean
    }
}`

		if got != want {
			t.Fatalf("GenerateTypeDef() = %q, want %q", got, want)
		}
	})

	t.Run("nested object", func(t *testing.T) {
		got := GenerateTypeDef(map[string]any{
			"database": map[string]any{
				"host": "localhost",
				"port": 5432,
				"credentials": map[string]any{
					"password": "pass",
					"username": "user",
				},
			},
			"features": map[string]any{
				"darkMode":    true,
				"experiments": []any{"feature1", "feature2"},
			},
		})

		want := `
import {Config} from "brek";
declare module "brek" {
    export interface Config {
        'database': {
            'credentials': {
                'password': string
                'username': string
            }
            'host': string
            'port': number
        }
        'features': {
            'darkMode': boolean
            'experiments': string[]
        }
    }
}`

		if got != want {
			t.Fatalf("GenerateTypeDef() = %q, want %q", got, want)
		}
	})

	t.Run("loader becomes string", func(t *testing.T) {
		got := GenerateTypeDef(map[string]any{
			"mongoDb": map[string]any{
				"[fetchSecret]": map[string]any{"key": "MONGODB_URI"},
			},
		})

		want := `
import {Config} from "brek";
declare module "brek" {
    export interface Config {
        'mongoDb': string
    }
}`

		if got != want {
			t.Fatalf("GenerateTypeDef() = %q, want %q", got, want)
		}
	})

	t.Run("props helper handles arrays", func(t *testing.T) {
		got := props(map[string]any{"list": []any{"item1", "item2"}}, 2)
		want := "        'list': string[]"
		if got != want {
			t.Fatalf("props() = %q, want %q", got, want)
		}
	})

	t.Run("type helpers cover object and empty array branches", func(t *testing.T) {
		if got := typeForValue([]any{}, 2); got != "any[]" {
			t.Fatalf("typeForValue(empty array) = %q", got)
		}
		if got := primitiveType(map[string]any{}, 0); got != "object" {
			t.Fatalf("primitiveType(map) = %q", got)
		}
		if got := primitiveType([]any{}, 0); got != "object" {
			t.Fatalf("primitiveType(array) = %q", got)
		}
	})
}

func TestWriteTypeDefAndDeleteConfJSON(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)

	writeTestJSON(t, tmp, "default.json", map[string]any{
		"foo": "bar",
		"typeTest": map[string]any{
			"array":   []any{1, 2, 3},
			"boolean": true,
			"number":  42,
			"object":  map[string]any{"key": "value"},
			"string":  "hello",
		},
	})

	if err := WriteTypeDef(); err != nil {
		t.Fatalf("WriteTypeDef() error = %v", err)
	}

	typePath := filepath.Join(tmp, "Config.d.ts")
	data, err := os.ReadFile(typePath)
	if err != nil {
		t.Fatalf("read type file: %v", err)
	}

	if !strings.Contains(string(data), "'typeTest': {") {
		t.Fatalf("unexpected type file contents: %s", data)
	}

	if err := WriteConfJSON(map[string]any{"foo": "bar", "count": 3}); err != nil {
		t.Fatalf("WriteConfJSON() error = %v", err)
	}

	confPath := filepath.Join(tmp, "config.json")
	got := readTestJSON(t, confPath)
	want := map[string]any{"foo": "bar", "count": 3}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("config.json = %#v, want %#v", got, want)
	}

	if err := DeleteConfJSON(); err != nil {
		t.Fatalf("DeleteConfJSON() error = %v", err)
	}
	if _, err := os.Stat(confPath); !os.IsNotExist(err) {
		t.Fatalf("expected config.json to be deleted, stat err = %v", err)
	}
}

func TestWriteTypeDefInvalidAndDeleteMissing(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)

	if err := DeleteConfJSON(); err != nil {
		t.Fatalf("DeleteConfJSON() missing error = %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmp, "config.json"), []byte(`{"foo":"bar"}`), 0o644); err != nil {
		t.Fatalf("write config.json: %v", err)
	}
	if err := os.Chmod(tmp, 0o555); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	if err := DeleteConfJSON(); err == nil {
		t.Fatal("expected permission error from DeleteConfJSON")
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		t.Fatalf("restore chmod dir: %v", err)
	}

	// Invalid default.json should fail.
	if err := os.WriteFile(filepath.Join(tmp, "default.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write invalid default.json: %v", err)
	}
	if err := WriteTypeDef(); err == nil {
		t.Fatal("expected WriteTypeDef() error for invalid JSON")
	}

	fileConfigDir := filepath.Join(tmp, "config-dir-as-file")
	if err := os.WriteFile(fileConfigDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file config dir: %v", err)
	}
	t.Setenv("BREK_CONFIG_DIR", fileConfigDir)
	if err := WriteTypeDef(); err == nil {
		t.Fatal("expected WriteTypeDef() mkdir error")
	}
}

func TestLoadConfigErrorBranches(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)

	// Invalid CLI override JSON should fail before any file work.
	t.Setenv("BREK", "{broken")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected LoadConfig() error for invalid overrides")
	}

	// Loader not found.
	t.Setenv("BREK", "")
	writeTestJSON(t, tmp, "default.json", map[string]any{
		"secret": map[string]any{"[missing]": "value"},
	})
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected LoadConfig() error for missing loader")
	}

	// Invalid default.json file.
	t.Setenv("BREK", "")
	if err := os.WriteFile(filepath.Join(tmp, "default.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write broken default.json: %v", err)
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected LoadConfig() error for invalid default.json")
	}

	resetTestState()
	t.Setenv("BREK", "")
	writeTestJSON(t, tmp, "default.json", map[string]any{
		"secret": map[string]any{"[boom]": "value"},
	})
	SetLoaders(LoaderDict{
		"boom": func(arg any) (string, error) {
			return "", errors.New("boom")
		},
	})
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected LoadConfig() error for loader failure")
	}

	resetTestState()
	t.Setenv("BREK", "")
	writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})
	writeDirFile := filepath.Join(tmp, "write-dir-file")
	if err := os.WriteFile(writeDirFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write write-dir file: %v", err)
	}
	t.Setenv("BREK_WRITE_DIR", writeDirFile)
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected LoadConfig() error for write failure")
	}
}

func TestGetConfigBranches(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)
	t.Setenv("USER", "")

	writeTestJSON(t, tmp, "default.json", map[string]any{"foo": "bar"})

	// missing config.json path -> LoadConfig branch
	conf, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() missing file error = %v", err)
	}
	if !reflect.DeepEqual(conf, map[string]any{"foo": "bar"}) {
		t.Fatalf("GetConfig() = %#v", conf)
	}

	// invalid config.json branch
	resetTestState()
	if err := os.WriteFile(filepath.Join(tmp, "config.json"), []byte("{broken"), 0o644); err != nil {
		t.Fatalf("write broken config.json: %v", err)
	}
	if _, err := GetConfig(); err == nil {
		t.Fatal("expected error for invalid config.json")
	}
}

func TestLoadConfigGetConfigAndRun(t *testing.T) {
	t.Cleanup(resetTestState)
	tmp := t.TempDir()
	t.Setenv("BREK_CONFIG_DIR", tmp)
	t.Setenv("BREK_WRITE_DIR", tmp)
	t.Setenv("BREK_TEST_VALUE", "resolved")
	t.Setenv("USER", "alice")

	writeTestJSON(t, tmp, "default.json", map[string]any{
		"foo": "bar",
		"addResult": map[string]any{
			"[add]": map[string]any{"a": 5, "b": 6},
		},
		"arrayWithEnv": []any{"${BREK_TEST_VALUE}"},
		"envValue":     "${BREK_TEST_VALUE}",
		"multiplyResult": map[string]any{
			"[multiplyBy5]": 5,
		},
		"typeTest": map[string]any{
			"array":   []any{1, 2, 3},
			"boolean": true,
			"number":  42,
			"object":  map[string]any{"key": "value"},
			"string":  "hello",
		},
	})

	loaders := LoaderDict{
		"add": func(arg any) (string, error) {
			obj := arg.(map[string]any)
			return strconv.Itoa(obj["a"].(int) + obj["b"].(int)), nil
		},
		"multiplyBy5": func(arg any) (string, error) {
			return strconv.Itoa(arg.(int) * 5), nil
		},
	}
	SetLoaders(loaders)

	got, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	want := map[string]any{
		"foo":            "bar",
		"addResult":      "11",
		"arrayWithEnv":   []any{"${BREK_TEST_VALUE}"},
		"envValue":       "resolved",
		"multiplyResult": "25",
		"typeTest": map[string]any{
			"array":   []any{1, 2, 3},
			"boolean": true,
			"number":  42,
			"object":  map[string]any{"key": "value"},
			"string":  "hello",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("LoadConfig() = %#v, want %#v", got, want)
	}

	confPath := filepath.Join(tmp, "config.json")
	if gotFile := readTestJSON(t, confPath); !reflect.DeepEqual(gotFile, want) {
		t.Fatalf("config.json = %#v, want %#v", gotFile, want)
	}

	// cache branch
	if err := os.Remove(confPath); err != nil {
		t.Fatalf("remove config.json: %v", err)
	}
	cached, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}
	if !reflect.DeepEqual(cached, want) {
		t.Fatalf("cached GetConfig() = %#v, want %#v", cached, want)
	}

	// reload-from-disk branch
	resetTestState()
	writeTestJSON(t, tmp, "config.json", want)
	readBack, err := GetConfig()
	if err != nil {
		t.Fatalf("GetConfig() disk error = %v", err)
	}
	if !reflect.DeepEqual(readBack, want) {
		t.Fatalf("GetConfig() disk = %#v, want %#v", readBack, want)
	}

	// CLI wrapper
	resetTestState()
	SetLoaders(loaders)
	if err := Run([]string{"load-config"}); err != nil {
		t.Fatalf("Run(load-config) error = %v", err)
	}
	if _, err := os.Stat(confPath); err != nil {
		t.Fatalf("expected config.json after Run(load-config): %v", err)
	}

	if err := Run([]string{"write-types"}); err != nil {
		t.Fatalf("Run(write-types) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "Config.d.ts")); err != nil {
		t.Fatalf("expected Config.d.ts after Run(write-types): %v", err)
	}
}

func TestRunUsageAndErrors(t *testing.T) {
	if err := Run(nil); err == nil {
		t.Fatal("expected usage error")
	}
	if err := Run([]string{"unknown"}); err == nil {
		t.Fatal("expected usage error")
	}

	if got := (ConfNotLoaded{}).Error(); got != "CONF_NOT_LOADED" {
		t.Fatalf("ConfNotLoaded.Error() = %q", got)
	}
}
