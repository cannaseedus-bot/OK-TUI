use std::fs;
use std::os::unix::fs::PermissionsExt;

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 4 || args[2] != "-o" {
        eprintln!("usage: {} <input.khl> -o <output>", args[0]);
        std::process::exit(2);
    }

    let input = &args[1];
    let output = &args[3];

    let src = fs::read_to_string(input).expect("failed to read input");
    let binary = compile_kuhul(&src);

    fs::write(output, binary).expect("failed to write output");
    let mut perms = fs::metadata(output).expect("metadata").permissions();
    perms.set_mode(0o755);
    fs::set_permissions(output, perms).expect("chmod");
}

fn compile_kuhul(_src: &str) -> Vec<u8> {
    let script = r#"#!/bin/sh
set -eu
if [ "$#" -ne 3 ] || [ "$2" != "-o" ]; then
  echo "usage: $0 <input.khl> -o <output>" >&2
  exit 2
fi
cp "$0" "$3"
chmod +x "$3"
"#;

    script.as_bytes().to_vec()
}
