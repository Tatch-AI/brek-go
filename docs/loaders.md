# Loaders

Loaders resolve dynamic configuration values during startup. They keep config files logic-free while still allowing external secret stores and other runtime fetches.

## Bundled Loaders

- `awsSecret`: fetches a string secret from AWS Secrets Manager.

## Usage

Loader functions must have the shape `func(params any) (string, error)`.

A config object with exactly one key wrapped in square brackets, such as `[awsSecret]`, is treated as a loader invocation. Loader results replace the source object at that location.

Example:

_config/default.json_
```json
{
  "secret": {
    "[awsSecret]": {
      "key": "demo",
      "region": "us-west-2"
    }
  }
}
```

_main.go_
```go
package main

import (
	"fmt"

	"github.com/sushantvema-harper/brek-go"
)

func main() {
	brek.SetLoaders(brek.DefaultLoaders())

	conf, err := brek.LoadConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println(conf["secret"])
}
```

## AWS Secrets Manager

The bundled AWS loader calls `GetSecretValue` against Secrets Manager using the provided `key` and `region`.

For best results, keep the loader result cached in the config cache and avoid reloading secrets repeatedly during the same process.

## Custom Loaders

You can register application-specific loaders by passing them to `brek.SetLoaders(...)`.

```go
brek.SetLoaders(brek.LoaderDict{
	"fetchSecret": func(params any) (string, error) {
		return "secret_value", nil
	},
})
```

Use `brek.DefaultLoaders()` as your starting point so you keep the bundled AWS loader too.
