# Loaders

Loaders resolve dynamic configuration values during startup. They keep config files logic-free while still allowing external secret stores and other runtime fetches.

## Bundled loaders

- `awsSecret`: fetches a string secret from AWS Secrets Manager.

## Example

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

## Usage

- Loader functions must have the shape `func(params any) (string, error)`.
- A config object with exactly one key wrapped in square brackets, such as `[awsSecret]`, is treated as a loader invocation.
- Loader results replace the source object at that location.

## AWS Secrets Manager

The bundled AWS loader calls `GetSecretValue` against Secrets Manager using the provided `key` and `region`. The AWS SDK for Go recommends retrieving secrets with `GetSecretValue` and caching values client-side for speed and lower cost.
