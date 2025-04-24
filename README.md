# yamlfmt

Format yaml documents with a JSONPath like syntax.

### Get Started

SDK:

```go
package main

import (
	"gopkg.in/yaml.v3"
    "github.com/Emptyless/yamlfmt"
)

func main() {
	node := new(yaml.Node)
	// or use the utility method to parse []byte: yamlfmt.LintBytes

	// optionally validate rules
	err := yamlfmt.Validate(rules)
	if err != nil {
		panic(err)
	}

	// lint document
	yamlfmt.Lint(node, []yamlfmt.Rule{yamlfmt.NewRule("$.key[*].name[0]", yamlfmt.StringOrderingFn, yamlfmt.NewSimpleOrdering("first", "second"))})
}
```

### Opinionated formatting of OpenAPI files
```
$ go install github.com/Emptyless/yamlfmt/openapi-fmt@latest
$ openapi-fmt --help
opinionated formatter of openapi.yaml files

Usage:
  openapi-fmt [flags]

Flags:
      --alphabetical stringArray   path to node to sort alphabetically (e.g. '$.key')
  -f, --file string                path to openapi.yaml file
  -o, --output string              path to output file
  -h, --help                       help for openapi-fmt
  -q, --quiet count                Decrease the verbosity of the output by one level, -v hides warning logs and -vv will suppress non-fatal errors
      --simple stringArray         path=keys to node to sort (e.g. path = '$.key') with comma separated list of keys
  -v, --verbose count              Increase the verbosity of the output by one level, -v shows informational logs and -vv will output debug information.
```

Given some openapi.yaml:

```
openapi-fmt --file openapi.yaml
```

Overwrite original specification:

```
openapi-fmt --file openapi.yaml --output openapi.yaml
```

Provide additional rules:

```
openapi-fmt --file openapi.yaml --alphabetical '$.key' --simple '$.key[*]=first,second'
```



