use crate::ast::{Node, Program};
use crate::backends::Backend;

pub struct GPUBackend {}

impl GPUBackend {
    pub fn new() -> Self {
        Self {}
    }
}

impl Backend for GPUBackend {
    fn emit(&self, program: &Program, output: &str) -> anyhow::Result<()> {
        let mut out = String::from("// KUHUL GPU IR\n");
        for node in &program.nodes {
            match node {
                Node::FluxPhase { phase, .. } => out.push_str(&format!("phase {}\n", phase)),
                Node::FluxBarrier { barrier } => out.push_str(&format!("barrier {}\n", barrier)),
                Node::Glyph { name, .. } => out.push_str(&format!("glyph {}\n", name)),
                Node::StreamChunk { size } => out.push_str(&format!("stream {}\n", size)),
                Node::Noop => {}
            }
        }

        std::fs::write(output, out)?;
        Ok(())
    }
}
