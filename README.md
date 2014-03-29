# Schematic

Generate Go client code for HTTP APIs described by [JSON Schemas](http://json-schema.org/).

## Installation

Download and install:

```console
$ go get github.com/heroku/schematic
```

**Warning**: schematic requires Go >= 1.2.

## Usage

Run it against your schema:

```console
$ schematic api.json > heroku/heroku.go 
```

This will generate a Go package named after your schema:

```go
package heroku
...
```

You then would be able to use the package as follow:

```go
h := heroku.NewService(nil)
addons, err := h.AddonList("schematic", nil)
if err != nil {
  ...
}
for _, addon := range addons {
  fmt.Println(addon.Name)
}
```
