# Getting Started

`brek-go` keeps configuration in JSON files and resolves it into a typed Go map at startup.

## Install

```bash
go get github.com/sushantvema-harper/brek-go
```

## Create config files

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

## Generate the config cache

Run:

```bash
brek load-config
```

That resolves the layered config and writes `config/config.json` by default.

## Read config in Go

```go
conf, err := brek.GetConfig()
if err != nil {
	panic(err)
}
```

## Register loaders

Go loaders are registered in code. Start with the bundled set:

```go
brek.SetLoaders(brek.DefaultLoaders())
```

You can add your own loaders on top of that map.

