package brek

import (
	"os"
	"regexp"
)

type EnvArguments struct {
	Environment string
	Deployment  string
	User        string
	Overrides   map[string]any
}

var envValuePattern = regexp.MustCompile(`^\$\{[a-zA-Z]+.*\}$`)

func IsEnvironmentVariable(value string) bool {
	return envValuePattern.MatchString(value)
}

func GetEnvOverrides() (map[string]any, error) {
	cliOverrides := osEnvFirst("BREK", "OVERRIDE")
	if cliOverrides == "" {
		return map[string]any{}, nil
	}

	decoded, err := decodeJSON([]byte(cliOverrides))
	if err != nil {
		return nil, InvalidConf{ValidationErrors: []string{"CLI overrides (BREK/OVERRIDE) is not valid JSON"}}
	}

	obj, ok := decoded.(map[string]any)
	if !ok {
		return nil, InvalidConf{ValidationErrors: []string{"CLI overrides (BREK/OVERRIDE) is not valid JSON"}}
	}

	return obj, nil
}

func GetEnvArguments() (EnvArguments, error) {
	overrides, err := GetEnvOverrides()
	if err != nil {
		return EnvArguments{}, err
	}

	return EnvArguments{
		Environment: osEnvFirst("ENVIRONMENT", "NODE_ENV"),
		Deployment:  getenv("DEPLOYMENT"),
		User:        getenv("USER"),
		Overrides:   overrides,
	}, nil
}

func osEnvFirst(keys ...string) string {
	for _, key := range keys {
		if value := getenv(key); value != "" {
			return value
		}
	}

	return ""
}

func getenv(key string) string {
	return os.Getenv(key)
}
