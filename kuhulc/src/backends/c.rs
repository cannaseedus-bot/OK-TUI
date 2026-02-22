use crate::ast::{Node, Program};
use crate::backends::Backend;

pub struct CBackend {}

impl CBackend {
    pub fn new() -> Self {
        Self {}
    }
}

impl Backend for CBackend {
    fn emit(&self, program: &Program, output: &str) -> anyhow::Result<()> {
        let mut out = String::new();

        out.push_str("#include \"agl_runtime.h\"\n\n");
        out.push_str("int main() {\n");

        for node in &program.nodes {
            match node {
                Node::FluxPhase { phase, .. } => {
                    out.push_str(&format!("  agl_flux_phase_enter(\"{}\");\n", phase))
                }
                Node::FluxBarrier { barrier } => {
                    out.push_str(&format!("  agl_flux_barrier(\"{}\");\n", barrier))
                }
                Node::Glyph { name, .. } => {
                    out.push_str(&format!("  agl_glyph_exec(\"{}\");\n", name))
                }
                Node::StreamChunk { size } => {
                    out.push_str(&format!("  agl_stream_chunk({});\n", size))
                }
                Node::Noop => {}
            }
        }

        out.push_str("  return 0;\n}\n");
        std::fs::write(output, out)?;
        Ok(())
    }
}
