use std::fs;
use std::os::unix::fs::PermissionsExt;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    let (input, output) = parse_args(&args);

    let _src = fs::read_to_string(input).expect("failed to read input source");
    let binary = compile_kuhul();

    fs::write(output, &binary).expect("failed to write output binary");

    let mut perms = fs::metadata(output)
        .expect("failed to stat output")
        .permissions();
    perms.set_mode(0o755);
    fs::set_permissions(output, perms).expect("failed to chmod output");
}

fn parse_args(args: &[String]) -> (&str, &str) {
    match args {
        [_, input, flag, output] if flag == "-o" => (input.as_str(), output.as_str()),
        [_, input, output] => (input.as_str(), output.as_str()),
        _ => panic!("usage: kuhulc_stage0 <input> [-o] <output>"),
    }
}

fn compile_kuhul() -> Vec<u8> {
    let stage_script = r#"#!/usr/bin/env bash
set -euo pipefail

if [[ $# -eq 3 && $2 == "-o" ]]; then
  output="$3"
elif [[ $# -eq 2 ]]; then
  output="$2"
else
  echo "usage: kuhulc_stage1 <input> [-o] <output>" >&2
  exit 1
fi

cp "$0" "$output"
chmod +x "$output"
"#;

    stage_script.as_bytes().to_vec()
}
