package brek

import "strings"

type InvalidConf struct {
	ValidationErrors []string
}

func (e InvalidConf) Error() string {
	return "INVALID_CONF: " + strings.Join(e.ValidationErrors, ", ")
}

type ConfNotLoaded struct{}

func (e ConfNotLoaded) Error() string {
	return "CONF_NOT_LOADED"
}

type LoaderNotFound struct {
	LoaderName string
	Available  []string
}

func (e LoaderNotFound) Error() string {
	available := "none"
	if len(e.Available) > 0 {
		available = strings.Join(e.Available, ", ")
	}

	return `LOADER_NOT_FOUND: "` + e.LoaderName + `". Available loaders: ` + available
}
