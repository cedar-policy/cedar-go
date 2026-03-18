.PHONY: test linters corpus-update check-upstream-corpus testdata-validation

# Run all tests
test:
	go test -count 1 -cover ./...

# Run linters and coverage checks
linters:
	golangci-lint run
	go run github.com/alecthomas/go-check-sumtype/cmd/go-check-sumtype@latest -default-signifies-exhaustive=false ./...
	go test -coverprofile=coverage.out ./...
	sed -i '' '/^github.com\/cedar-policy\/cedar-go\/internal\/schema\/parser\/cedarschema.go/d' coverage.out
	go tool cover -func=coverage.out | sed 's/%$$//' | awk '$$1 == "total:" { next } { if ($$3 < 100.0) { printf "Insufficient code coverage for %s\n", $$0; failed=1 } } END { exit failed }'

# Download the latest corpus tests tarball and overwrite corpus-tests.tar.gz if changed
.PHONY: check-upstream-corpus
check-upstream-corpus:
	@tmp="$$(mktemp)" && \
	curl -fL -o "$$tmp" https://raw.githubusercontent.com/cedar-policy/cedar-integration-tests/main/corpus-tests.tar.gz && \
	if cmp -s "$$tmp" corpus-tests.tar.gz; then echo "corpus-tests.tar.gz is up to date."; rm -f "$$tmp"; else mv "$$tmp" corpus-tests.tar.gz; echo "corpus-tests.tar.gz updated."; fi

# Use an order-only prerequisite to check for changes. This allows other targets to
# reference corpus-tests.tar.gz as a dependency so that they'll only be re-created
# if the corpus tests are updated.
corpus-tests.tar.gz: | check-upstream-corpus

# Convert Cedar schemas to JSON schemas
corpus-tests-json-schemas.tar.gz: corpus-tests.tar.gz
	@echo "Generating JSON schemas from Cedar schemas..."
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-json-schemas
	@mkdir -p /tmp/corpus-tests-json-schemas
	@tar -xzmf corpus-tests.tar.gz -C /tmp/
	@for schema in /tmp/corpus-tests/*.cedarschema; do \
		basename=$$(basename $$schema .cedarschema); \
		echo "Converting $$basename.cedarschema..."; \
		cedar translate-schema --direction cedar-to-json --schema "$$schema" > "/tmp/corpus-tests-json-schemas/$$basename.cedarschema.json" 2>&1; \
	done
	@tar -czf corpus-tests-json-schemas.tar.gz -C /tmp corpus-tests-json-schemas
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-json-schemas
	@echo "Done! Created corpus-tests-json-schemas.tar.gz"

# Build cedar-validation-tool and generate validation results
test/cedar-validation-tool/target/release/cedar-validation-tool: test/cedar-validation-tool/src/main.rs test/cedar-validation-tool/Cargo.toml
	@echo "Building cedar-validation-tool..."
	@cd test/cedar-validation-tool && cargo build --release

corpus-tests-validation.tar.gz: corpus-tests.tar.gz test/cedar-validation-tool/target/release/cedar-validation-tool
	@echo "Generating validation results from Rust Cedar..."
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-validation
	@mkdir -p /tmp/corpus-tests-validation
	@tar -xzmf corpus-tests.tar.gz -C /tmp/
	@for testjson in /tmp/corpus-tests/*.json; do \
		case "$$testjson" in *.entities.json) continue ;; esac; \
		basename=$$(basename $$testjson .json); \
		test/cedar-validation-tool/target/release/cedar-validation-tool \
			"$$testjson" "/tmp/corpus-tests-validation/$${basename}.validation.json"; \
	done
	@cd /tmp && tar -czf corpus-tests-validation.tar.gz corpus-tests-validation/
	@mv /tmp/corpus-tests-validation.tar.gz .
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-validation
	@echo "Done! Created corpus-tests-validation.tar.gz"

# Regenerate validation data for x/exp/schema/validate/testdata
testdata-validation: test/cedar-validation-tool/target/release/cedar-validation-tool
	@echo "Regenerating testdata validation files..."
	@for testjson in x/exp/schema/validate/testdata/*.json; do \
		case "$$testjson" in *.entities.json|*.validation.json) continue ;; esac; \
		basename=$$(basename $$testjson .json); \
		echo "  Validating $$basename..."; \
		test/cedar-validation-tool/target/release/cedar-validation-tool \
			"$$testjson" "x/exp/schema/validate/testdata/$${basename}.validation.json"; \
	done
	@echo "Done! Regenerated testdata validation files."

# Download, convert, and validate
corpus-update: corpus-tests-json-schemas.tar.gz corpus-tests-validation.tar.gz testdata-validation
