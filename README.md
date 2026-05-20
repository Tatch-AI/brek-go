# brek-go

`brek-go` is a structured, typed config loader for Go. It keeps configuration declarative, supports layered config files, resolves environment variables, and ships with a bundled AWS Secrets Manager loader.

`brek` stands for **B**locking **R**esolution of **E**nvironment **K**eys.

_config/default.json_
```json
{
  "foo": "bar",
  "baz": {
    "qux": 42
  },
  "quux": {
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

- Structured configuration in JSON files.
- Layered merge strategy for default, environment, deployment, user, and CLI/env overrides.
- Environment variable expansion with `${VAR}` syntax.
- Loader support for dynamic runtime values.
- Bundled `awsSecret` loader for AWS Secrets Manager.
- Generated Go-facing config typing through `Config.d.ts` for downstream TypeScript consumers.
- A small, test-covered codebase with Go-native APIs and CLI entry points.

## Table of Contents

- [Getting Started](#getting-started)
- [Configuration Rules](#configuration-rules)
- [Configuration Merge Strategy](#configuration-merge-strategy)
- [Using CLI/ENV Overrides](#using-clienv-overrides)
- [Environment Variables in Config Files](#environment-variables-in-config-files)
- [Loaders](#loaders)
- [API Reference](#api-reference)
- [CLI Reference](#cli-reference)
- [Recommended Best Practices](#recommended-best-practices)
- [Usage with AWS Lambda](#usage-with-aws-lambda)
- [Debugging](#debugging)
- [Known Issues](#known-issues)
- [Support, Feedback, and Contributions](#support-feedback-and-contributions)
- [Why is it called brek?](#why-is-it-called-brek)
- [Licensing](#licensing)

## Getting Started

See [docs/gettingStarted.md](docs/gettingStarted.md).

## Configuration Rules

- `default.json` is required. Keep it as the canonical shape of the config object.
- Environment, deployment, user, and override files should only override keys that already exist in `default.json`.
- A property’s type should not change based on environment, deployment, or user.
- Loaders always return strings. If you need another type, encode it before returning.
- Arrays should be homogeneous.

## Configuration Merge Strategy

`brek-go` merges configuration from least to most specific:

1. `default.json`
2. environment file
3. deployment file
4. user file
5. CLI/env overrides

Which files are considered depends on these environment variables:

| `process.env` | Config file |
| --- | --- |
| `ENVIRONMENT`, `NODE_ENV` | `config/environments/[ENVIRONMENT].json` |
| `DEPLOYMENT` | `config/deployments/[DEPLOYMENT].json` |
| `USER` | `config/users/[USER].json` |

Notes:

- `ENVIRONMENT` wins over `NODE_ENV` if both are set.
- Arrays and loader objects are replaced, not merged.
- `BREK_WRITE_DIR` controls where `config.json` and `Config.d.ts` are written.

## Using CLI/ENV Overrides

Set `BREK` or `OVERRIDE` to a JSON object to override config at runtime.

Examples:

```bash
# Override a nested key
BREK='{"a":{"b":"q"}}' brek load-config
```

```bash
# Inject a value from the shell environment
DATABASE_URL="postgres://user:pass@localhost:5432/db"
BREK=$(jq -n --arg db "$DATABASE_URL" '{postgres: $db}') brek load-config
```

## Environment Variables in Config Files

Use `${VAR_NAME}` inside JSON values to read from the process environment.

```json
{
  "foo": "${FOO}"
}
```

Those values are always resolved as strings.

## Loaders

Loaders resolve dynamic config values during startup and keep config files logic-free.

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

See [docs/loaders.md](docs/loaders.md) for the bundled `awsSecret` loader and custom loader guidance.

## API Reference

### `brek.GetConfig() (map[string]any, error)`

Returns the cached configuration if it has already been loaded. If `config.json` exists, it is read from disk. Otherwise, `brek` resolves the config from source files and writes `config.json`.

### `brek.LoadConfig() (map[string]any, error)`

Loads configuration from source files, merges the layers, resolves loaders and environment variables, then writes `config.json`.

### `brek.WriteConfJSON(map[string]any) error`

Writes a resolved configuration map to `config.json`.

### `brek.DeleteConfJSON() error`

Deletes the generated `config.json` cache if it exists.

### `brek.WriteTypeDef() error`

Writes `config/Config.d.ts` from `default.json`.

### `brek.DefaultLoaders() brek.LoaderDict`

Returns the bundled loader set, including `awsSecret`.

### `brek.SetLoaders(loaders brek.LoaderDict)`

Registers application-specific loaders on top of the bundled set.

## CLI Reference

You can run the CLI via `brek` or `lambdaconf`.

```bash
brek load-config
brek write-types
lambdaconf load-config
lambdaconf write-types
```

Commands:

- `load-config` resolves and writes `config.json`
- `write-types` writes `Config.d.ts` and clears `config.json`, matching the original brek CLI behavior

## Recommended Best Practices

- Keep `default.json` complete and treat the other config files as overrides.
- Store secrets outside of JSON and fetch them through loaders.
- Register loaders once during startup before calling `LoadConfig()`.
- Prefer `go test ./...` in CI so config parsing, loaders, and CLI behavior stay covered.

## Usage with AWS Lambda

AWS Lambda has a read-only filesystem outside `/tmp`.

Set `BREK_WRITE_DIR=/tmp` if you want `config.json` or `Config.d.ts` written during Lambda startup.

Example:

```json
{
  "scripts": {
    "build": "go test ./...",
    "load-config": "BREK_WRITE_DIR=/tmp brek load-config"
  }
}
```

If you need config loaded during runtime, call `brek.LoadConfig()` in your handler before processing requests.

## Debugging

Set `BREK_DEBUG=1` or `LAMBDACONF_DEBUG=1` to enable internal debug logs.

Useful checks:

- Confirm `BREK_CONFIG_DIR` points at the directory containing `default.json`
- Confirm `BREK_WRITE_DIR` is writable
- Confirm loader names in JSON match the keys returned by `brek.DefaultLoaders()` or `brek.SetLoaders(...)`
- Confirm `BREK` / `OVERRIDE` contains valid JSON

## Known Issues

- `BREK_LOADERS_FILE_PATH` is retained for parity with the original project, but Go loaders are registered in code rather than loaded from a JS file.
- `write-types` clears `config.json` after writing `Config.d.ts`, which matches the original brek CLI but can surprise if you expect the cache to remain.
- The bundled AWS Secrets Manager loader requires AWS credentials and a valid region.

## Support, Feedback, and Contributions

Open issues or PRs against this repository if you find a parity gap or a behavioral mismatch with the original brek project.

## Why is it called brek?

Same as the original project: **B**locking **R**esolution of **E**nvironment **K**eys.

## Licensing

`brek-go` follows the same MIT licensing model as the original `brek` project.
