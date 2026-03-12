use std::collections::BTreeMap;
use std::env;
use std::fs;
use std::path::Path;

use cedar_policy::{
    Context, Entities, EntityId, EntityTypeName, EntityUid, PolicySet, Request, Schema,
    ValidationMode, Validator,
};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Deserialize)]
struct TestFile {
    schema: String,
    policies: String,
    entities: String,
    requests: Vec<TestRequest>,
}

#[derive(Deserialize)]
struct TestRequest {
    description: String,
    principal: EntityRef,
    action: EntityRef,
    resource: EntityRef,
    context: Value,
    #[serde(rename = "validateRequest", default)]
    validate_request: bool,
}

#[derive(Deserialize)]
struct EntityRef {
    #[serde(rename = "type")]
    entity_type: String,
    id: String,
}

#[derive(Serialize)]
struct ValidationResult {
    #[serde(rename = "policyValidation")]
    policy_validation: PolicyValidationResult,
    #[serde(rename = "entityValidation")]
    entity_validation: EntityValidationResult,
    #[serde(rename = "requestValidation")]
    request_validation: Vec<RequestResult>,
}

#[derive(Serialize)]
struct PolicyValidationResult {
    strict: bool,
    permissive: bool,
    #[serde(rename = "strictErrors", skip_serializing_if = "Vec::is_empty")]
    strict_errors: Vec<String>,
    #[serde(rename = "permissiveErrors", skip_serializing_if = "Vec::is_empty")]
    permissive_errors: Vec<String>,
    #[serde(rename = "perPolicy")]
    per_policy: BTreeMap<String, PerPolicyResult>,
}

#[derive(Serialize)]
struct PerPolicyResult {
    strict: bool,
    permissive: bool,
    #[serde(rename = "strictErrors", skip_serializing_if = "Vec::is_empty")]
    strict_errors: Vec<String>,
    #[serde(rename = "permissiveErrors", skip_serializing_if = "Vec::is_empty")]
    permissive_errors: Vec<String>,
}

#[derive(Serialize)]
struct EntityValidationResult {
    #[serde(rename = "perEntity")]
    per_entity: BTreeMap<String, PerEntityResult>,
}

#[derive(Serialize)]
struct PerEntityResult {
    #[serde(rename = "errors", skip_serializing_if = "Vec::is_empty")]
    errors: Vec<String>,
}

#[derive(Serialize)]
struct RequestResult {
    description: String,
    strict: Option<bool>,
    permissive: Option<bool>,
    #[serde(rename = "errors", skip_serializing_if = "Vec::is_empty")]
    errors: Vec<String>,
}

fn make_entity_uid(r: &EntityRef) -> EntityUid {
    EntityUid::from_type_name_and_id(
        r.entity_type.parse::<EntityTypeName>().unwrap(),
        EntityId::new(&r.id),
    )
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: {} <test-json-path> <output-json-path>", args[0]);
        std::process::exit(1);
    }

    let test_json_path = &args[1];
    let output_json_path = &args[2];

    let test_json = fs::read_to_string(test_json_path).expect("Failed to read test JSON");
    let test: TestFile = serde_json::from_str(&test_json).expect("Failed to parse test JSON");

    // Resolve file paths relative to the parent of the corpus-tests directory
    let base_dir = Path::new(test_json_path)
        .parent()
        .unwrap()
        .parent()
        .unwrap();

    // Read schema
    let schema_str = fs::read_to_string(base_dir.join(&test.schema)).expect("Failed to read schema");
    let (schema, _warnings) =
        Schema::from_cedarschema_str(&schema_str).expect("Failed to parse schema");

    // Read policies
    let policies_str =
        fs::read_to_string(base_dir.join(&test.policies)).expect("Failed to read policies");
    let policy_set: PolicySet = policies_str.parse().expect("Failed to parse policies");

    // Read entities
    let entities_str =
        fs::read_to_string(base_dir.join(&test.entities)).expect("Failed to read entities");

    // 1. Policy validation (strict and permissive) with per-policy breakdown
    let validator = Validator::new(schema.clone());
    let strict_policy = validator.validate(&policy_set, ValidationMode::Strict);
    let permissive_policy = validator.validate(&policy_set, ValidationMode::Permissive);

    // Aggregate results
    let strict_all_errors: Vec<String> = strict_policy
        .validation_errors()
        .map(|e| format!("{e}"))
        .collect();
    let permissive_all_errors: Vec<String> = permissive_policy
        .validation_errors()
        .map(|e| format!("{e}"))
        .collect();

    // Per-policy breakdown
    let mut per_policy = BTreeMap::new();
    let mut policy_ids: Vec<String> = policy_set.policies().map(|p| p.id().to_string()).collect();
    policy_ids.sort();

    // Re-validate to get fresh iterators for per-policy grouping
    let strict_result2 = validator.validate(&policy_set, ValidationMode::Strict);
    let permissive_result2 = validator.validate(&policy_set, ValidationMode::Permissive);

    let strict_errors_all: Vec<_> = strict_result2.validation_errors().collect();
    let permissive_errors_all: Vec<_> = permissive_result2.validation_errors().collect();

    for pid_str in &policy_ids {
        let strict_for_policy: Vec<String> = strict_errors_all
            .iter()
            .filter(|e| e.policy_id().to_string() == *pid_str)
            .map(|e| format!("{e}"))
            .collect();
        let permissive_for_policy: Vec<String> = permissive_errors_all
            .iter()
            .filter(|e| e.policy_id().to_string() == *pid_str)
            .map(|e| format!("{e}"))
            .collect();
        per_policy.insert(
            pid_str.clone(),
            PerPolicyResult {
                strict: strict_for_policy.is_empty(),
                permissive: permissive_for_policy.is_empty(),
                strict_errors: strict_for_policy,
                permissive_errors: permissive_for_policy,
            },
        );
    }

    let policy_validation = PolicyValidationResult {
        strict: strict_policy.validation_passed(),
        permissive: permissive_policy.validation_passed(),
        strict_errors: strict_all_errors,
        permissive_errors: permissive_all_errors,
        per_policy,
    };

    // 2. Per-entity validation
    let entities_json: Vec<Value> =
        serde_json::from_str(&entities_str).expect("Failed to parse entities JSON");

    let mut per_entity = BTreeMap::new();
    for entity_value in &entities_json {
        let uid = &entity_value["uid"];
        // Handle both {"type": ..., "id": ...} and {"__entity": {"type": ..., "id": ...}} formats
        let (entity_type, entity_id) = if let Some(inner) = uid.get("__entity") {
            (
                inner["type"].as_str().expect("entity uid must have type"),
                inner["id"].as_str().expect("entity uid must have id"),
            )
        } else {
            (
                uid["type"].as_str().expect("entity uid must have type"),
                uid["id"].as_str().expect("entity uid must have id"),
            )
        };

        // Use raw type::id format (no Cedar quoting) to avoid escape format
        // differences between Go and Rust
        let uid_str = format!("{}::{}", entity_type, entity_id);

        // Validate single entity
        let single_entity_json =
            serde_json::to_string(&vec![entity_value]).expect("Failed to serialize single entity");
        let result = Entities::from_json_str(&single_entity_json, Some(&schema));
        let errors: Vec<String> = match &result {
            Err(e) => {
                // Collect the full error chain for maximum detail
                let mut msgs = vec![format!("{e}")];
                let err: &dyn std::error::Error = &*e;
                let mut source = err.source();
                while let Some(cause) = source {
                    msgs.push(format!("{cause}"));
                    source = cause.source();
                }
                if msgs.len() > 1 {
                    vec![msgs.join(": ")]
                } else {
                    msgs
                }
            }
            Ok(_) => vec![],
        };

        per_entity.insert(uid_str, PerEntityResult { errors });
    }

    let entity_validation = EntityValidationResult { per_entity };

    // 3. Per-request validation
    let mut request_validation = Vec::new();
    for req in &test.requests {
        if !req.validate_request {
            request_validation.push(RequestResult {
                description: req.description.clone(),
                strict: None,
                permissive: None,
                errors: vec![],
            });
            continue;
        }

        let principal = make_entity_uid(&req.principal);
        let action = make_entity_uid(&req.action);
        let resource = make_entity_uid(&req.resource);

        // Validate request: build context with schema, then build request with schema.
        // The Rust Request API doesn't take a validation mode parameter,
        // so strict and permissive produce the same result.
        let result = (|| -> Result<(), Box<dyn std::error::Error>> {
            let context =
                Context::from_json_value(req.context.clone(), Some((&schema, &action)))?;
            Request::new(principal, action, resource, context, Some(&schema))?;
            Ok(())
        })();
        let valid = result.is_ok();
        let errors: Vec<String> = match &result {
            Err(e) => vec![format!("{e}")],
            Ok(_) => vec![],
        };

        request_validation.push(RequestResult {
            description: req.description.clone(),
            strict: Some(valid),
            permissive: Some(valid),
            errors,
        });
    }

    let result = ValidationResult {
        policy_validation,
        entity_validation,
        request_validation,
    };

    let output = serde_json::to_string_pretty(&result).expect("Failed to serialize result");
    fs::write(output_json_path, output).expect("Failed to write output");
}
