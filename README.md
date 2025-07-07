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
- human-readable schema parsing

The Go implementation does not yet include:

- CLI applications
- the schema [validator](https://docs.cedarpolicy.com/policies/validation.html)
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

## Change log

### 1.2.3
#### New Features
- Adds Entity.Equal()

### 1.2.2
#### New Features
- Adds experimental support for inspecting Cedar policy AST

### 1.2.1
#### New Features
- Fixes the name of `AuthorizationPolicySet` and receiver types for `PolicySet.Get()` and `PolicySet.IsAuthorized()`
- Retracts 1.2.0

### 1.2.0
#### New Features
- Support for the .isEmpty() operator.
- A new top-level Authorize() function, which allows authorization against a generic policy iterator (`PolicyIterator`) instead of requiring a `PolicySet`. Like the `EntityGetter` interface does for entities, using a generic iterator enables policy to be retrieved from external sources or for policy to be selected dynamically by the iterator implementation without having to clone an entire `PolicySet`.
- batch.Authorize() likewise now also accepts a `PolicyIterator`.
- First class iterator support for EntityUIDSet, Record, Set, and PolicySet container types.

#### Upgrading from 1.1.0
- cedar-go now requires Go 1.23

### 1.1.0
#### New features
- Support for entity tags via the .getTag() and .hasTag() operators.

### 1.0.0
### New features
- AST builder methods for Cedar datetime and duration literals and their extension methods have been added
- AST builder methods for adding extension function calls with uninterpreted strings
- Small improvement in evaluation runtime performance for large, shallow entity graphs.

#### Upgrading from 0.4.x to 1.0.0

- The `Parents` field on `types.Entity` has been changed to an immutable set type with an interface similar to `types.Set`
- The `UnsafeDecimal()` constructor for the `types.Decimal` type has been removed and replaced with the following safe constructors, which return error on overflow:
  - `NewDecimal(int64 i, int exponent) (Decimal, error)`
  - `NewDecimalFromInt[T constraints.Signed](i T) (Decimal, error)`
  - `NewDecimalFromFloat[T constraints.Float](f T) (Decimal, error)`
- The `Value` field on `types.Decimal` has been made private. Instances of `Decimal` can be compared with one another via the new `Compare` method.
- `types.DecimalPrecision` has been made private
- The following error types have been made private: `types.ErrDateitme`, `types.ErrDecimal`, `types.ErrDuration`, `types.ErrIP`, `types.ErrNotComparable`
- The following datetime and duration-related constructors have been renamed:
  - `types.FromStdTime()` has been renamed to `types.NewDatetime()`
  - `types.DatetimeFromMillis()` has been renamed to `types.NewDatetimeFromMillis()`
  - `types.FromStdDuration()` has been renamed to `types.NewDuration()`
  - `types.DurationFromMillis()` has been renamed to `types.NewDurationFromMillis()`
- `types.Entities` has been renamed to `types.EntityMap`
- Because `types.Entity` is now immutable, `types.EntityMap` now stores items by value rather than by pointer
- `PolicySet.Store()` has been renamed to `PolicySet.Add()`
- `PolicySet.Delete()` has been renamed to `PolicySet.Remove()`
- `types.Set()` now takes variadic arguments of type `types.Value` instead of a single `[]types.Value` argument

### 0.4.0
#### New features

- `types.Set` is now implemented as a hash set, turning `Set.Contains()` into an O(1) operation, on average. This mitigates a worst case quadratic runtime for the evaluation of the `containsAny()` operator.
- For convenience, public types, constructors, and constants from the `types` package are now exported via the `cedar` package as well.

#### Upgrading from 0.3.x to 0.4.x

- `types.Set` is now an immutable type which must be constructed via `types.NewSet()`
  - To iterate the values, use `Set.Iterate()`, which takes an iterator callback.
  - Duplicates are now removed from `Set`s, so they won't be rendered when calling `Set.MarshalCedar()` or `Set.MarshalJSON`.
  - All implementations of `types.Value` are now safe to copy shallowly, so `Set.DeepClone()` has been removed.
- `types.Record` is now an immutable type which must be constructed via `types.NewRecord()`
  - To iterate the keys and values, use `Record.Iterate()`, which takes an iterator callback.
  - All implementations of `types.Value` are now safe to copy shallowly, so `Record.DeepClone()` has been removed.

### 0.3.2
#### New features

- An implementation of the `datetime` and `duration` extension types specified in [RFC 80](https://github.com/cedar-policy/rfcs/blob/main/text/0080-datetime-extension.md).
  - Note: While these types have been accepted into the language, they have not yet been formally analyzed in the [specification](https://github.com/cedar-policy/cedar-spec/).

### 0.3.1
#### New features

- General performance improvements to the evaluator
- An experimental batch evaluator has been added to `x/exp/batch`
- Reserved keywords are now rejected in all appropriate places when parsing Cedar text
- A parsing ambiguity between variables, entity UIDs, and extension functions has been resolved

#### Upgrading from 0.2.x to 0.3.x

- The JSON marshaling of the Position struct now uses canonical lower-case keys for its fields

### 0.2.0
#### New features

- A programmatic AST is now available in the `ast` package.
- Policy sets can be marshaled and unmarshaled from JSON.
- Policies can also be marshaled to Cedar text.

#### Upgrading from 0.1.x to 0.2.x

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
