package brek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func readJSONFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return nil, InvalidConf{ValidationErrors: []string{path + " is not valid JSON"}}
	}

	decoded, err := decodeJSON(data)
	if err != nil {
		return nil, InvalidConf{ValidationErrors: []string{path + " is not valid JSON"}}
	}

	obj, ok := decoded.(map[string]any)
	if !ok {
		return nil, InvalidConf{ValidationErrors: []string{path + " is not valid JSON"}}
	}

	return obj, nil
}

func decodeJSON(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}

	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("extra json content")
	}

	return normalizeJSONValue(value), nil
}

func normalizeJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = normalizeJSONValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = normalizeJSONValue(item)
		}
		return out
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}

		if f, err := v.Float64(); err == nil {
			return f
		}

		return v.String()
	default:
		return value
	}
}

func writeJSONFile(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), ".brek-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Sync(); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}

	return nil
}
