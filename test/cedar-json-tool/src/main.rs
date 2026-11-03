//! Records how Rust Cedar decodes and re-encodes JSON inputs, so cedar-go's
//! JSON codec can be compared with it.

use std::env;
use std::fs;
use std::process;

use cedar_policy::{Entities, Entity};
use serde::Serialize;
use serde_json::Value;

#[derive(Serialize)]
struct EntitiesResult {
    accepted: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    entities: Option<Vec<Value>>,
}

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 4 {
        usage(&args[0]);
    }
    let (mode, input_path, output_path) = (&args[1], &args[2], &args[3]);

    let result = match mode.as_str() {
        "entities" => serde_json::to_value(entities(input_path)).unwrap(),
        _ => usage(&args[0]),
    };

    let output = serde_json::to_string(&result).unwrap();
    fs::write(output_path, output + "\n").expect("Failed to write output");
}

fn usage(program: &str) -> ! {
    eprintln!(
        "Usage: {} entities <input-json-path> <output-json-path>",
        program
    );
    process::exit(1);
}

/// Decodes an entities JSON file without a schema and, if Rust Cedar accepts
/// it, re-encodes each entity.
fn entities(input_path: &str) -> EntitiesResult {
    // Read through the reader API, so that invalid UTF-8 reaches the parser.
    let f = fs::File::open(input_path).expect("Failed to open input");
    if Entities::from_json_file(f, None).is_err() {
        return EntitiesResult {
            accepted: false,
            entities: None,
        };
    }

    // Entities adds every ancestor of an entity to its parents, so decode
    // each entity on its own to keep only the parents in the input.
    let input = fs::read(input_path).expect("Failed to read input");
    let values: Vec<Value> = serde_json::from_slice(&input).unwrap();
    let entities = values
        .into_iter()
        .map(|value| {
            let entity = Entity::from_json_value(value, None).unwrap();
            let mut json = entity.to_json_value().unwrap();
            // Rust Cedar writes parents in HashSet order.
            if let Some(Value::Array(parents)) = json.get_mut("parents") {
                parents.sort_by_key(|p| p.to_string());
            }
            sort_keys(&mut json);
            json
        })
        .collect();
    EntitiesResult {
        accepted: true,
        entities: Some(entities),
    }
}

/// Sorts the members of every object in a JSON value. Rust Cedar writes
/// attributes and tags in HashMap order, and serde_json keeps insertion order,
/// so this keeps the output reproducible.
fn sort_keys(value: &mut Value) {
    match value {
        Value::Array(elems) => elems.iter_mut().for_each(sort_keys),
        Value::Object(members) => {
            members.sort_keys();
            members.values_mut().for_each(sort_keys);
        }
        _ => {}
    }
}
