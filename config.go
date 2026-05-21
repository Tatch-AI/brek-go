package brek

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"syscall"
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
	return withConfigLock(func() (map[string]any, error) {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		if cachedConfig != nil {
			debug("getConfig: returning cached config")
			return cachedConfig, nil
		}

		conf, err := readConfigJSON()
		if err == nil {
			debug("getConfig: loaded config.json from disk")
			cachedConfig = conf
			return conf, nil
		}

		if !os.IsNotExist(err) {
			return nil, err
		}

		return loadConfigUnlocked()
	})
}

func withConfigLock[T any](fn func() (T, error)) (T, error) {
	lockFile, err := os.OpenFile(ConfigLockPath(), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		var zero T
		return zero, err
	}
	defer lockFile.Close()

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		var zero T
		return zero, err
	}
	defer syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)

	return fn()
}

func loadConfigUnlocked() (map[string]any, error) {
	env, err := GetEnvArguments()
	if err != nil {
		return nil, err
	}
	debug("loadConfig: env", env)

	sources, err := LoadConfFromFiles(env)
	if err != nil {
		return nil, err
	}
	debug("loadConfig: sources", sources)

	merged := MergeConfs(sources)
	debug("loadConfig: merged", merged)
	resolved, err := resolveMap(merged, currentLoaders())
	if err != nil {
		return nil, err
	}
	debug("loadConfig: resolved", resolved)

	if err := WriteConfJSON(resolved); err != nil {
		return nil, err
	}

	return resolved, nil
}

func LoadConfig() (map[string]any, error) {
	return withConfigLock(func() (map[string]any, error) {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		resolved, err := loadConfigUnlocked()
		if err != nil {
			return nil, err
		}

		cachedConfig = resolved
		return resolved, nil
	})
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
	_, err := withConfigLock(func() (struct{}, error) {
		cacheMu.Lock()
		defer cacheMu.Unlock()

		path := ConfigJSONPath()
		debug("deleteConfJSON:", path)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return struct{}{}, err
		}

		cachedConfig = nil

		return struct{}{}, nil
	})
	return err
}

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: brek load-config")
	}

	switch strings.TrimSpace(args[0]) {
	case "load-config":
		_, err := LoadConfig()
		return err
	default:
		return fmt.Errorf("usage: brek load-config")
	}
}
