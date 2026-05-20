package brek

import (
	"os"
	"path/filepath"
)

type ConfSources struct {
	Default     map[string]any
	Environment map[string]any
	Deployment  map[string]any
	User        map[string]any
	Overrides   map[string]any
}

func LoadConfFromFiles(env EnvArguments) (ConfSources, error) {
	defaultConf, err := LoadConfFile("default.json")
	if err != nil {
		return ConfSources{}, err
	}

	environment := map[string]any{}
	if env.Environment != "" {
		environment, err = LoadConfFile("environments", env.Environment+".json")
		if err != nil {
			return ConfSources{}, err
		}
	}

	deployment := map[string]any{}
	if env.Deployment != "" {
		deployment, err = LoadConfFile("deployments", env.Deployment+".json")
		if err != nil {
			return ConfSources{}, err
		}
	}

	user := map[string]any{}
	if env.User != "" {
		user, err = LoadConfFile("users", env.User+".json")
		if err != nil {
			return ConfSources{}, err
		}
	}

	return ConfSources{
		Default:     defaultConf,
		Environment: environment,
		Deployment:  deployment,
		User:        user,
		Overrides:   env.Overrides,
	}, nil
}

func LoadConfFile(parts ...string) (map[string]any, error) {
	pathParts := append([]string{ConfigDir()}, parts...)
	path := filepath.Join(pathParts...)
	contents, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}, nil
	}

	decoded, err := decodeJSON(contents)
	if err != nil {
		return nil, InvalidConf{ValidationErrors: []string{path + " is not valid JSON"}}
	}

	obj, ok := decoded.(map[string]any)
	if !ok {
		return nil, InvalidConf{ValidationErrors: []string{path + " is not valid JSON"}}
	}

	return obj, nil
}

func MergeConfs(sources ConfSources) map[string]any {
	configs := []map[string]any{
		sources.Default,
		sources.Environment,
		sources.Deployment,
		sources.User,
		sources.Overrides,
	}

	merged := map[string]any{}
	for _, config := range configs {
		merged = mergeConfigs(merged, config)
	}

	return merged
}

func mergeConfigs(left, right map[string]any) map[string]any {
	if left == nil {
		left = map[string]any{}
	}
	if right == nil {
		right = map[string]any{}
	}

	merged := make(map[string]any, len(left)+len(right))
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}

	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok {
			continue
		}

		leftMap, leftOK := leftValue.(map[string]any)
		rightMap, rightOK := rightValue.(map[string]any)
		if !leftOK || !rightOK {
			continue
		}

		if IsLoader(leftMap) || IsLoader(rightMap) {
			continue
		}

		merged[key] = mergeConfigs(leftMap, rightMap)
	}

	return merged
}
