package brek

import "github.com/Tatch-AI/brek-go/loaders/awssecret"

func DefaultLoaders() LoaderDict {
	return LoaderDict{
		"awsSecret": awssecret.Loader,
	}
}
