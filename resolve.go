package brek

import (
	"os"
	"sort"
	"strings"
)

func ResolveConf(value any, loaders LoaderDict) (any, error) {
	return resolveAny(value, loaders)
}

func resolveAny(value any, loaders LoaderDict) (any, error) {
	switch v := value.(type) {
	case map[string]any:
		if IsLoader(v) {
			return resolveLoader(v, loaders)
		}
		return resolveMap(v, loaders)
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			switch item.(type) {
			case map[string]any, []any:
				resolved, err := resolveAny(item, loaders)
				if err != nil {
					return nil, err
				}
				out[i] = resolved
			default:
				out[i] = item
			}
		}
		return out, nil
	case string:
		if IsEnvironmentVariable(v) {
			name := strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")
			return os.Getenv(name), nil
		}
		return v, nil
	default:
		return v, nil
	}
}

func resolveMap(value map[string]any, loaders LoaderDict) (map[string]any, error) {
	out := make(map[string]any, len(value))
	keys := sortedMapKeys(value)
	for _, key := range keys {
		resolved, err := resolveAny(value[key], loaders)
		if err != nil {
			return nil, err
		}
		out[key] = resolved
	}
	return out, nil
}

func resolveLoader(prop map[string]any, loaders LoaderDict) (any, error) {
	name := loaderName(prop)
	loader, ok := loaders[name]
	if !ok {
		return nil, LoaderNotFound{
			LoaderName: name,
			Available:  availableLoaderNames(loaders),
		}
	}

	var params any
	for key := range prop {
		params = prop[key]
		break
	}

	return loader(params)
}

func sortedMapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
