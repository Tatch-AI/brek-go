package brek

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GenerateTypeDef(config map[string]any) string {
	var b strings.Builder
	b.WriteString("\nimport {Config} from \"brek\";\ndeclare module \"brek\" {\n    export interface Config {\n")
	b.WriteString(props(config, 2))
	b.WriteString("\n    }\n}")
	return b.String()
}

func props(obj map[string]any, depth int) string {
	keys := sortedMapKeys(obj)
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s'%s': %s", indent(depth), key, typeForValue(obj[key], depth)))
	}
	return strings.Join(lines, "\n")
}

func typeForValue(value any, depth int) string {
	switch v := value.(type) {
	case []any:
		if len(v) == 0 {
			return "any[]"
		}
		return primitiveType(v[0], depth) + "[]"
	case map[string]any:
		if IsLoader(v) {
			return "string"
		}
		keys := sortedMapKeys(v)
		lines := make([]string, 0, len(keys))
		for _, key := range keys {
			lines = append(lines, fmt.Sprintf("%s'%s': %s", indent(depth+1), key, typeForValue(v[key], depth+1)))
		}
		return "{\n" + strings.Join(lines, "\n") + "\n" + indent(depth) + "}"
	default:
		return primitiveType(v, depth)
	}
}

func primitiveType(value any, _ int) string {
	switch value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return "number"
	case map[string]any, []any:
		return "object"
	default:
		return "string"
	}
}

func indent(depth int) string {
	return strings.Repeat(" ", depth*4)
}

func WriteTypeDef() error {
	defaultConfig, err := LoadConfFile("default.json")
	if err != nil {
		return err
	}

	filepath := filepath.Join(ConfigDir(), "Config.d.ts")
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath, []byte(GenerateTypeDef(defaultConfig)), 0o644)
}
