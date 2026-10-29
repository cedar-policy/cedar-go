use std::io::{self, BufReader, BufWriter, Read, Write};

use cedar_policy::{Policy, Template};
use cedar_policy_core::est;

fn main() {
    let mut input = BufReader::new(io::stdin().lock());
    let mut output = BufWriter::new(io::stdout().lock());
    loop {
        let mut op = [0u8; 1];
        match input.read_exact(&mut op) {
            Ok(()) => {}
            Err(e) if e.kind() == io::ErrorKind::UnexpectedEof => return,
            Err(e) => panic!("failed to read request: {e}"),
        }
        let payload = read_frame(&mut input);
        let result = match op[0] {
            b't' => from_text(&payload),
            b'j' => from_json(&payload),
            other => panic!("unknown operation {other}"),
        };
        let (status, body) = match result {
            Ok(text) => (0u8, text),
            Err(err) => (1u8, err),
        };
        output.write_all(&[status]).unwrap();
        write_frame(&mut output, body.as_bytes());
        output.flush().unwrap();
    }
}

fn read_frame(r: &mut impl Read) -> Vec<u8> {
    let mut len = [0u8; 4];
    r.read_exact(&mut len).expect("failed to read length");
    let mut buf = vec![0u8; u32::from_be_bytes(len) as usize];
    r.read_exact(&mut buf).expect("failed to read payload");
    buf
}

fn write_frame(w: &mut impl Write, b: &[u8]) {
    w.write_all(&(b.len() as u32).to_be_bytes()).unwrap();
    w.write_all(b).unwrap();
}

fn from_text(b: &[u8]) -> Result<String, String> {
    let src = std::str::from_utf8(b).map_err(describe)?;
    let json = match Policy::parse(None, src) {
        Ok(policy) => policy.to_json().map_err(describe)?,
        Err(policy_err) => match Template::parse(None, src) {
            Ok(template) => template.to_json().map_err(describe)?,
            Err(_) => return Err(describe(policy_err)),
        },
    };
    write_text(json)
}

fn from_json(b: &[u8]) -> Result<String, String> {
    let value: serde_json::Value = serde_json::from_slice(b).map_err(describe)?;
    let json = match Policy::from_json(None, value.clone()) {
        Ok(policy) => policy.to_json().map_err(describe)?,
        Err(policy_err) => match Template::from_json(None, value) {
            Ok(template) => template.to_json().map_err(describe)?,
            Err(_) => return Err(describe(policy_err)),
        },
    };
    write_text(json)
}

fn write_text(json: serde_json::Value) -> Result<String, String> {
    let policy: est::Policy = serde_json::from_value(json).map_err(describe)?;
    let template = policy
        .try_into_ast_policy_or_template(None)
        .map_err(describe)?;
    Ok(est::Policy::from(template).to_string())
}

fn describe(e: impl std::error::Error) -> String {
    let mut msg = e.to_string();
    let mut source = e.source();
    while let Some(s) = source {
        msg.push_str(": ");
        msg.push_str(&s.to_string());
        source = s.source();
    }
    msg
}
