# cedar-go

![Cedar Logo](https://github.com/cedar-policy/cedar/blob/main/logo.svg)

![Build and Test](https://github.com/cedar-policy/cedar-go/actions/workflows/build_and_test.yml/badge.svg)
![Nightly Corpus Test](https://github.com/cedar-policy/cedar-go/actions/workflows/corpus.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/cedar-policy/cedar-go.svg)](https://pkg.go.dev/github.com/cedar-policy/cedar-go)

This repository contains source code of the Go implementation of the [Cedar](https://www.cedarpolicy.com/) policy language.

Cedar is a language for writing and enforcing authorization policies in your applications. Using Cedar, you can write policies that specify your applications' fine-grained permissions. Your applications then authorize access requests by calling Cedar's authorization engine. Because Cedar policies are separate from application code, they can be independently authored, updated, analyzed, and audited. You can use Cedar's validator to check that Cedar policies are consistent with a declared schema which defines your application's authorization model.

Cedar is:

### Expressive

Cedar is a simple yet expressive language that is purpose-built to support authorization use cases for common authorization models such as RBAC and ABAC.

### Performant

Cedar is fast and scalable. The policy structure is designed to be indexed for quick retrieval and to support fast and scalable real-time evaluation, with bounded latency.

### Analyzable

Cedar is designed for analysis using Automated Reasoning. This enables analyzer tools capable of optimizing your policies and proving that your security model is what you believe it is.

## Using Cedar

Cedar can be used in your application by importing the `github.com/cedar-policy/cedar-go` package.

## Comparison to the Rust implementation

The Go implementation includes:

- the core authorizer
- JSON marshalling and unmarshalling
- all core and extended types (including [RFC 80](https://github.com/cedar-policy/rfcs/blob/main/text/0080-datetime-extension.md)'s datetime and duration)
- integration test suite
- schema parsing and programmatic construction
 
The Go implementation does not yet include:

- CLI applications
- the schema [validator](https://docs.cedarpolicy.com/policies/validation.html) (experimental support is provided in x/exp/schema - please give us feedback!)
- the formatter
- partial evaluation
- support for [policy templates](https://docs.cedarpolicy.com/policies/templates.html)

## Quick Start

Here's a simple example of using Cedar in Go:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	cedar "github.com/cedar-policy/cedar-go"
)

const policyCedar = `permit (
	principal == User::"alice",
	action == Action::"view",
	resource in Album::"jane_vacation"
  );
`

const entitiesJSON = `[
  {
    "uid": { "type": "User", "id": "alice" },
    "attrs": { "age": 18 },
    "parents": []
  },
  {
    "uid": { "type": "Photo", "id": "VacationPhoto94.jpg" },
    "attrs": {},
    "parents": [{ "type": "Album", "id": "jane_vacation" }]
  }
]`

func main() {
	var policy cedar.Policy
	if err := policy.UnmarshalCedar([]byte(policyCedar)); err != nil {
		log.Fatal(err)
	}

	ps := cedar.NewPolicySet()
	ps.Add("policy0", &policy)

	var entities cedar.EntityMap
	if err := json.Unmarshal([]byte(entitiesJSON), &entities); err != nil {
		log.Fatal(err)
	}
	
	req := cedar.Request{
		Principal: cedar.NewEntityUID("User", "alice"),
		Action:    cedar.NewEntityUID("Action", "view"),
		Resource:  cedar.NewEntityUID("Photo", "VacationPhoto94.jpg"),
		Context:   cedar.NewRecord(cedar.RecordMap{
			"demoRequest": cedar.True,
        }),
	}

	ok, _ := cedar.Authorize(ps, entities, req)
	fmt.Println(ok)
}
```

CLI output:

```
allow
```

This request is allowed because `VacationPhoto94.jpg` belongs to `Album::"jane_vacation"`, and `alice` can view photos in `Album::"jane_vacation"`.

If you'd like to see more details on what can be expressed as Cedar policies, see the [documentation](https://docs.cedarpolicy.com).

## Packages
The cedar-go module houses the following public packages:
 * [cedar](.) - The main package for interacting with the module, including parsing policies and entities, schemas, and authorizing requests.
 * [ast](ast/) - Programmatic construction of Cedar ASTs
 * [types](types/) - Basic types common to multiple packages. For convenience, most of these are also projected through the cedar package.
 * [x/exp/batch](x/exp/batch/) - An experimental batch authorization API supporting high-performance variable substitution via partial evaluation.
 * [x/exp/schema](x/exp/schema) - Experimental support for Cedar schema, including parsing the Cedar and JSON formats, programmatic Schema construction, and validation

## Documentation

General documentation for Cedar is available at [docs.cedarpolicy.com](https://docs.cedarpolicy.com), with source code in the [cedar-policy/cedar-docs](https://github.com/cedar-policy/cedar-docs/) repository.

Generated documentation for the latest version of the Go implementation can be accessed
[here](https://pkg.go.dev/github.com/cedar-policy/cedar-go).

If you're looking to integrate Cedar into a production system, please be sure the read the [security best practices](https://docs.cedarpolicy.com/other/security.html)

## Backward Compatibility Considerations
- `x/exp` - code in this directory is not subject to the semantic versioning constraints of the rest of the module and breaking changes may be made at any time.
- Variadics may be added to functions that do not have them to expand the arguments of a function or method.
- Concrete types may be replaced with compatible interfaces to expand the variety of arguments a function or method can take.
- Backwards compatibility is maintained for all Go minor versions released within 6 months of a release of cedar-go.

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## Contributing

We welcome contributions from the community. Please either file an issue, or see [CONTRIBUTING](CONTRIBUTING.md)

## License

This project is licensed under the Apache-2.0 License.
