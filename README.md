# marc

[![Go Reference](https://pkg.go.dev/badge/github.com/net-of-arc/marc.svg)](https://pkg.go.dev/github.com/net-of-arc/marc)

`marc` is a pure Go codec between [ArcNet CORE](https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md) Markdown node files and bags of `(subject, predicate, object)` triples. Your application works with a node as data, iterating triples and typed objects, and never touches Markdown or YAML.

- Decoding a canonical file and encoding it again reproduces the same bytes.
- Encoding a bag of triples and decoding it again reproduces the same triples, including predicates your vocabulary does not know.
- Every malformed input yields a distinct error you can match with `errors.Is`. Input size is bounded and the decoder is fuzzed.
- The codec has no side effects: no filesystem, network, environment or global state.

## Installation

```bash
go get github.com/net-of-arc/marc
```

Supported Go versions: 1.26 and 1.27.

## Quick start

```go
package main

import (
	"fmt"
	"os"

	"github.com/net-of-arc/marc"
)

func main() {
	src, _ := os.ReadFile("Transport Layer Security.md")

	node, err := marc.Unmarshal(src)
	if err != nil {
		panic(err)
	}

	for t := range node.Triples() {
		switch v := t.Object.(type) {
		case marc.Ref:
			fmt.Println(t.Predicate, "→", v.Target())
		case marc.Number:
			f, _ := v.Float64()
			fmt.Println(t.Predicate, f)
		default:
			fmt.Println(t.Predicate, v)
		}
	}

	codec, err := marc.New(marc.Config{
		Schema: marc.Properties{
			marc.MustPredicate("text"):        {Role: marc.RoleText, Default: true},
			marc.MustPredicate("tags"):        {Role: marc.RoleMeta, Many: true},
			marc.MustPredicate("replaces"):    {Role: marc.RoleEdge},
			marc.MustPredicate("mentionedIn"): {Role: marc.RoleLink},
		},
	})
	if err != nil {
		panic(err)
	}

	out, err := codec.Marshal(node)
	if err != nil {
		panic(err)
	}
	os.Stdout.Write(out)
}
```

`marc.Properties` is the reference `Schema`. To drive placement from your own vocabulary, implement the two-method `marc.Schema` interface yourself.

## Documentation

- API reference: [pkg.go.dev/github.com/net-of-arc/marc](https://pkg.go.dev/github.com/net-of-arc/marc)
- Format: [ArcNet CORE](https://github.com/net-of-arc/arcnet-spec/blob/main/spec/CORE.md)
- Design: [ARCHITECTURE.md](ARCHITECTURE.md) and [adrs/](adrs/)

## License

[MIT](LICENSE)
