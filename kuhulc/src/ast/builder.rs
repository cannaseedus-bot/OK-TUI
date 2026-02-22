use crate::ast::{Node, Program};
use crate::parser::ParseTree;

pub fn build(parse_tree: ParseTree) -> anyhow::Result<Program> {
    let mut nodes = Vec::with_capacity(parse_tree.entries.len());

    for entry in parse_tree.entries {
        if let Some((_, phase)) = entry.split_once("phase:") {
            let phase = phase.trim().to_string();
            let phase_id = stable_id(&phase);
            nodes.push(Node::FluxPhase { phase, phase_id });
        } else if let Some((_, barrier)) = entry.split_once("barrier:") {
            nodes.push(Node::FluxBarrier {
                barrier: barrier.trim().to_string(),
            });
        } else if let Some((_, glyph)) = entry.split_once("glyph:") {
            let name = glyph.trim().to_string();
            let glyph_id = stable_id(&name);
            nodes.push(Node::Glyph { name, glyph_id });
        } else if let Some((_, size)) = entry.split_once("stream:") {
            let size = size.trim().parse::<u32>().unwrap_or_default();
            nodes.push(Node::StreamChunk { size });
        } else {
            nodes.push(Node::Noop);
        }
    }

    Ok(Program { nodes })
}

fn stable_id(input: &str) -> u32 {
    input
        .bytes()
        .fold(2166136261u32, |hash, b| hash.wrapping_mul(16777619) ^ b as u32)
}
