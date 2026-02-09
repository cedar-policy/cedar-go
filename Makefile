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
	@cd /tmp && tar -czf corpus-tests-json-schemas.tar.gz corpus-tests-json-schemas/
	@mv /tmp/corpus-tests-json-schemas.tar.gz .
	@rm -rf /tmp/corpus-tests /tmp/corpus-tests-json-schemas
	@echo "Done! Created corpus-tests-json-schemas.tar.gz"
