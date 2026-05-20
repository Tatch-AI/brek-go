package brek

import (
	"sort"
	"strings"
)

type Loader func(params any) (string, error)

type LoaderDict map[string]Loader

func IsLoader(prop map[string]any) (isLoader bool) {
	if len(prop) != 1 {
		return false
	}

	for key := range prop {
		isLoader = strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]")
		break
	}

	return
}

func loaderName(prop map[string]any) string {
	for key := range prop {
		return strings.TrimSuffix(strings.TrimPrefix(key, "["), "]")
	}

	return ""
}

func availableLoaderNames(loaders LoaderDict) []string {
	if len(loaders) == 0 {
		return nil
	}

	names := make([]string, 0, len(loaders))
	for name := range loaders {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
