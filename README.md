# brek-go

`brek-go` is a structured, typed config loader for Go. It keeps config logic declarative, supports layered configuration, resolves environment variables, and ships with a bundled AWS Secrets Manager loader.

_config/default.json_
```json
{
  "foo": "bar",
  "db": {
    "port": 5432
  },
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
	conf, err := brek.GetConfig()
	if err != nil {
		panic(err)
	}

	fmt.Println(conf["foo"])
}
```

## Features

- Structured configuration in JSON files
- Layered merge strategy: default, environment, deployment, user, CLI/env overrides
- Environment variable expansion with `${VAR}` syntax
- Loader support for dynamic runtime values
- Bundled `awsSecret` loader for AWS Secrets Manager
- Generated Go code via normal typing, with tests covering the port end to end

## Getting Started

See [docs/gettingStarted.md](docs/gettingStarted.md).

## Examples

See [examples/awssecret/README.md](examples/awssecret/README.md) for a full Go example using the bundled `awsSecret` loader.

## Loaders

See [docs/loaders.md](docs/loaders.md).

## API

- `brek.GetConfig()` reads the cached config from memory or disk, then falls back to loading from source files
- `brek.LoadConfig()` resolves configuration from source files, loaders, and overrides, then writes `config.json`
- `brek.WriteTypeDef()` writes `Config.d.ts` for downstream TypeScript consumers
- `brek.DefaultLoaders()` returns the bundled loader set, including `awsSecret`
- `brek.SetLoaders(...)` registers additional application loaders

## CLI

```bash
brek load-config
brek write-types
```

## Configuration

- `BREK_CONFIG_DIR` defaults to `config`
- `BREK_WRITE_DIR` defaults to `BREK_CONFIG_DIR`
- `BREK_LOADERS_FILE_PATH` is retained for parity with the TypeScript project, but Go loaders are registered in code
- `BREK` or `OVERRIDE` can contain JSON overrides
- `ENVIRONMENT`, `NODE_ENV`, `DEPLOYMENT`, and `USER` select layered config files

## Development

```bash
go test ./...
```
