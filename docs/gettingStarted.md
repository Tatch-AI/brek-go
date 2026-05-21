# Getting Started

`brek-go` keeps configuration in JSON files, resolves it into a Go map at startup, and writes the resolved result back to disk for reuse.

## Install

Use it as a Go module:

```bash
go get github.com/Tatch-AI/brek-go
```

Install the CLI:

```bash
go install github.com/Tatch-AI/brek-go/cmd/brek@latest
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

If you are calling the module directly from Go code, import `github.com/Tatch-AI/brek-go` and call `brek.GetConfig()` or `brek.LoadConfig()`.

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

## CLI

Run the CLI directly:

```bash
brek load-config
```

If you are migrating from the original project, the Go equivalent workflow is:

```bash
go install github.com/Tatch-AI/brek-go/cmd/brek@latest
brek load-config
```
