/*
 KUHUL Compiler Driver
 Canonical MX2⟁☣ deterministic compiler
*/

mod ast;
mod backends;
mod cli;
mod emit;
mod parser;
mod pipeline;
mod registry;

use backends::{Backend, CBackend, GPUBackend, SCXQ2Backend};
use clap::Parser;
use cli::CompilerConfig;
use pipeline::CompilerPipeline;

fn main() -> anyhow::Result<()> {
    let config = CompilerConfig::parse();

    println!("KUHUL Compiler v1.0");
    println!("Input: {}", config.input);

    let registry = registry::load_all()?;
    let mut pipeline = CompilerPipeline::new(registry);
    let ast = pipeline.compile(&config.input)?;

    let backend: Box<dyn Backend> = match config.target.as_str() {
        "c" => Box::new(CBackend::new()),
        "gpu" => Box::new(GPUBackend::new()),
        "scxq2" => Box::new(SCXQ2Backend::new()),
        _ => anyhow::bail!("Unknown backend: {}", config.target),
    };

    backend.emit(&ast, &config.output)?;

    println!("Output written: {}", config.output);

    Ok(())
}
