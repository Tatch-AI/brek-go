package brek

import (
	"fmt"
	"os"
	"sync"
)

var (
	cacheMu       sync.RWMutex
	cachedConfig  map[string]any
	customLoads   = LoaderDict{}
	defaultLoadMu sync.RWMutex
)

func SetLoaders(loaders LoaderDict) {
	defaultLoadMu.Lock()
	defer defaultLoadMu.Unlock()

	customLoads = cloneLoaders(loaders)
}

func cloneLoaders(loaders LoaderDict) LoaderDict {
	if loaders == nil {
		return LoaderDict{}
	}

	out := make(LoaderDict, len(loaders))
	for name, loader := range loaders {
		out[name] = loader
	}
	return out
}

func currentLoaders() LoaderDict {
	defaultLoadMu.RLock()
	defer defaultLoadMu.RUnlock()

	merged := cloneLoaders(DefaultLoaders())
	for name, loader := range customLoads {
		merged[name] = loader
	}

	return merged
}

func GetConfig() (map[string]any, error) {
	cacheMu.RLock()
	if cachedConfig != nil {
		out := cachedConfig
		cacheMu.RUnlock()
		return out, nil
	}
	cacheMu.RUnlock()

	conf, err := readConfigJSON()
	if err == nil {
		cacheMu.Lock()
		cachedConfig = conf
		cacheMu.Unlock()
		return conf, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	return LoadConfig()
}

func LoadConfig() (map[string]any, error) {
	env, err := GetEnvArguments()
	if err != nil {
		return nil, err
	}

	sources, err := LoadConfFromFiles(env)
	if err != nil {
		return nil, err
	}

	merged := MergeConfs(sources)
	resolved, err := resolveMap(merged, currentLoaders())
	if err != nil {
		return nil, err
	}

	if err := WriteConfJSON(resolved); err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cachedConfig = resolved
	cacheMu.Unlock()

	return resolved, nil
}

func readConfigJSON() (map[string]any, error) {
	path := ConfigJSONPath()
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	return readJSONFile(path)
}

func WriteConfJSON(resolvedConf map[string]any) error {
	return writeJSONFile(ConfigJSONPath(), resolvedConf)
}

func DeleteConfJSON() error {
	if err := os.Remove(ConfigJSONPath()); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: brek [load-config|write-types]")
	}

	switch args[0] {
	case "load-config":
		_, err := LoadConfig()
		return err
	case "write-types":
		return WriteTypeDef()
	default:
		return fmt.Errorf("usage: brek [load-config|write-types]")
	}
}
