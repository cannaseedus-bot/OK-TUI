use std::io::Write;

use crate::ast::{Node, Program};
use crate::backends::Backend;

pub struct SCXQ2Backend {}

impl SCXQ2Backend {
    pub fn new() -> Self {
        Self {}
    }
}

impl Backend for SCXQ2Backend {
    fn emit(&self, program: &Program, output: &str) -> anyhow::Result<()> {
        let mut file = std::fs::File::create(output)?;

        for node in &program.nodes {
            match node {
                Node::FluxPhase { phase_id, .. } => {
                    file.write_all(&[0x01])?;
                    file.write_all(&phase_id.to_le_bytes())?;
                }
                Node::Glyph { glyph_id, .. } => {
                    file.write_all(&[0x02])?;
                    file.write_all(&glyph_id.to_le_bytes())?;
                }
                Node::FluxBarrier { .. } => {
                    file.write_all(&[0x03])?;
                }
                Node::StreamChunk { size } => {
                    file.write_all(&[0x04])?;
                    file.write_all(&size.to_le_bytes())?;
                }
                Node::Noop => {}
            }
        }

        Ok(())
    }
}
