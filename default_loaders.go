package brek

import "github.com/sushantvema-harper/brek-go/loaders/awssecret"

func DefaultLoaders() LoaderDict {
	return LoaderDict{
		"awsSecret": awssecret.Loader,
	}
}
