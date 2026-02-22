mod cli;
mod pipeline;
mod registry;

use cli::Cli;
use std::{fs, path::Path};

fn main() -> Result<(), String> {
    let cli = Cli::parse()?;

    let counts = registry::load_counts(Path::new("registry"))?;
    eprintln!(
        "Loaded registries: {} glyphs, {} opcodes",
        counts.glyphs, counts.opcodes
    );

    let program = pipeline::compile(Path::new(&cli.input)).map_err(|e| e.to_string())?;
    let output = pipeline::emit(&program, cli.target, cli.deterministic);
    fs::write(&cli.output, output).map_err(|e| e.to_string())?;
    Ok(())
}
