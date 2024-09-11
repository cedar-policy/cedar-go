# cedar-go

![Cedar Logo](https://github.com/cedar-policy/cedar/blob/main/logo.svg)

![Build and Test](https://github.com/cedar-policy/cedar-go/actions/workflows/build_and_test.yml/badge.svg)
![golangci-lint](https://github.com/cedar-policy/cedar-go/actions/workflows/golangci-lint.yml/badge.svg)
![Nightly Corpus Test](https://github.com/cedar-policy/cedar-go/actions/workflows/corpus.yml/badge.svg)

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
- all core and extended types
- integration test suite

The Go implementation does not yet include:

- examples and CLI applications
- schema support and the validator
- the formatter
- partial evaluation

## Quick Start

Here's a simple example of using Cedar in Go:

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	cedar "github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"
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
	ps.Store("policy0", &policy)

	var entities types.Entities
	if err := json.Unmarshal([]byte(entitiesJSON), &entities); err != nil {
		log.Fatal(err)
	}
	req := cedar.Request{
		Principal: types.EntityUID{Type: "User", ID: "alice"},
		Action:    types.EntityUID{Type: "Action", ID: "view"},
		Resource:  types.EntityUID{Type: "Photo", ID: "VacationPhoto94.jpg"},
		Context:   types.Record{},
	}

	ok, _ := ps.IsAuthorized(entities, req)
	fmt.Println(ok)
}
```

CLI output:

```
allow
```

This request is allowed because `VacationPhoto94.jpg` belongs to `Album::"jane_vacation"`, and `alice` can view photos in `Album::"jane_vacation"`.

If you'd like to see more details on what can be expressed as Cedar policies, see the [documentation](https://docs.cedarpolicy.com).

## Documentation

General documentation for Cedar is available at [docs.cedarpolicy.com](https://docs.cedarpolicy.com), with source code in the [cedar-policy/cedar-docs](https://github.com/cedar-policy/cedar-docs/) repository.

Generated documentation for the latest version of the Go implementation can be accessed
[here](https://pkg.go.dev/github.com/cedar-policy/cedar-go).

If you're looking to integrate Cedar into a production system, please be sure the read the [security best practices](https://docs.cedarpolicy.com/other/security.html)

## Backward Compatibility Considerations

x/exp - code in this subrepository is not subject to the Go 1
compatibility promise.

While in development (0.x.y), each tagged release may contain breaking changes.

## Change log

### New features in 0.4.x

- General performance improvements to the evaluator
- An experimental batch evaluator has been added to `x/exp/batch`

### New features in 0.2.x

- A programmatic AST is now available in the `ast` package.
- Policy sets can be marshaled and unmarshaled from JSON.
- Policies can also be marshaled to Cedar text.

### Upgrading from 0.1.x to 0.2.x

- The Cedar value types have moved from the `cedar` package to the `types` package.
- The PolicyIDs are now `strings`, previously they were numeric.
- Errors and reasons use the new `PolicyID` form.
- Combining multiple parsed Cedar files now involves coming up with IDs for each
statement in those files.  It's best to create an empty `NewPolicySet` then
parse individual files using `NewPolicyListFromBytes` and subsequently use
`PolicySet.Store` to add each of the policy statements.
- The Cedar `Entity` and `Entities` types have moved from the `cedar` package to the `types` package.
- Stronger typing is being used in many places.
- The `Value` method `Cedar() string` was changed to `MarshalCedar() []byte`


## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## Contributing

We welcome contributions from the community. Please either file an issue, or see [CONTRIBUTING](CONTRIBUTING.md)

## License

This project is licensed under the Apache-2.0 License.
