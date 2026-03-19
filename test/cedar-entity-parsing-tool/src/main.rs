use std::collections::BTreeMap;
use std::env;
use std::fs;
use std::path::Path;

use cedar_policy::{Entities, Schema};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Deserialize)]
struct TestFile {
    schema: String,
    entities: String,
}

#[derive(Serialize)]
struct Result {
    #[serde(rename = "allEntities")]
    all_entities: ParseResult,
    #[serde(rename = "perEntity")]
    per_entity: BTreeMap<String, ParseResult>,
}

#[derive(Serialize)]
struct ParseResult {
    success: bool,
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!(
            "Usage: {} <test-json-path> <output-json-path>",
            args[0]
        );
        std::process::exit(1);
    }

    let test_json_path = &args[1];
    let output_json_path = &args[2];

    let test_json = fs::read_to_string(test_json_path).expect("Failed to read test JSON");
    let test: TestFile = serde_json::from_str(&test_json).expect("Failed to parse test JSON");

    let base_dir = Path::new(test_json_path).parent().unwrap();

    let schema_str =
        fs::read_to_string(base_dir.join(&test.schema)).expect("Failed to read schema");
    let (schema, _) =
        Schema::from_cedarschema_str(&schema_str).expect("Failed to parse schema");

    let entities_str =
        fs::read_to_string(base_dir.join(&test.entities)).expect("Failed to read entities");

    // Parse all entities together with schema guidance
    let all_success = Entities::from_json_str(&entities_str, Some(&schema)).is_ok();

    // Per-entity parsing
    let mut per_entity = BTreeMap::new();
    if let Ok(entities_array) = serde_json::from_str::<Vec<Value>>(&entities_str) {
        for entity_value in &entities_array {
            let uid = &entity_value["uid"];
            let entity_type = uid["type"].as_str().unwrap_or("?");
            let entity_id = uid["id"].as_str().unwrap_or("?");
            let uid_str = format!("{}::{}", entity_type, entity_id);

            let single = serde_json::to_string(&vec![entity_value]).unwrap();
            let success = Entities::from_json_str(&single, Some(&schema)).is_ok();
            per_entity.insert(uid_str, ParseResult { success });
        }
    }

    let result = Result {
        all_entities: ParseResult { success: all_success },
        per_entity,
    };

    let output = serde_json::to_string_pretty(&result).unwrap();
    fs::write(output_json_path, output).unwrap();
}
