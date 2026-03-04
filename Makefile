.PHONY: test linters corpus-update

# Run all tests
test:
	go test -count 1 -cover ./...

# Run linters and coverage checks
linters:
	golangci-lint run
	go run github.com/alecthomas/go-check-sumtype/cmd/go-check-sumtype@latest -default-signifies-exhaustive=false ./...
	go test -coverprofile=coverage.out ./...
	sed -i '' '/^github.com\/cedar-policy\/cedar-go\/internal\/schema\/parser\/cedarschema.go/d' coverage.out
	go tool cover -func=coverage.out | sed 's/%$$//' | awk '{ if ($$3 < 100.0) { printf "Insufficient code coverage for %s\n", $$0; failed=1 } } END { exit failed }'

# Download the latest corpus tests tarball
corpus-tests.tar.gz:
	curl -L -o corpus-tests.tar.gz https://raw.githubusercontent.com/cedar-policy/cedar-integration-tests/main/corpus-tests.tar.gz

# Convert Cedar schemas to JSON schemas
corpus-tests-json-schemas.tar.gz: corpus-tests.tar.gz
	@echo "Generating JSON schemas from Cedar schemas..."
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-json-schemas
	@mkdir -p /tmp/corpus-tests-json-schemas
	@tar -xzf corpus-tests.tar.gz -C /tmp/
	@for schema in /tmp/corpus-tests/*.cedarschema; do \
		basename=$$(basename $$schema .cedarschema); \
		echo "Converting $$basename.cedarschema..."; \
		cedar translate-schema --direction cedar-to-json --schema "$$schema" > "/tmp/corpus-tests-json-schemas/$$basename.cedarschema.json" 2>&1; \
	done
	@tar -czf corpus-tests-json-schemas.tar.gz -C /tmp corpus-tests-json-schemas
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-json-schemas
	@echo "Done! Created corpus-tests-json-schemas.tar.gz"

# Download and convert
corpus-update:
	rm -f corpus-tests.tar.gz corpus-tests-json-schemas.tar.gz
	$(MAKE) corpus-tests-json-schemas.tar.gz
