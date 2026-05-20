# Getting Started

`brek-go` keeps configuration in JSON files, resolves it into a Go map at startup, and writes the resolved result back to disk for reuse.

## Install

```bash
go get github.com/sushantvema-harper/brek-go
```

Or build the CLI locally:

```bash
go run ./cmd/brek load-config
```

## Create Config Files

Create a `config` directory in the root of your project. Only `default.json` is required.

```text
root/
└── config/
    ├── deployments/
    ├── environments/
    ├── users/
    └── default.json
```

Example `default.json`:

```json
{
  "port": 3000,
  "postgres": {
    "host": "localhost",
    "password": "pgpassword"
  }
}
```

## Generate the Config Cache

Run:

```bash
brek load-config
```

That resolves the layered config and writes `config/config.json` by default.

## Read Config in Go

```go
conf, err := brek.GetConfig()
if err != nil {
	panic(err)
}
```

## Register Loaders

Go loaders are registered in code.

Start with the bundled set:

```go
brek.SetLoaders(brek.DefaultLoaders())
```

You can add your own loaders on top of that map before calling `LoadConfig()`.

## Write Types

`brek write-types` writes `config/Config.d.ts` from `default.json` and clears any stale `config.json` cache.

If you are migrating from the original TypeScript project, this is the Go equivalent of the same developer workflow.
