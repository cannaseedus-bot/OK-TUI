pub mod builder;
pub mod validator;

use crate::parser::ParseTree;
use crate::registry::Registry;

#[derive(Clone, Debug)]
pub struct Program {
    pub nodes: Vec<Node>,
}

#[derive(Clone, Debug)]
pub enum Node {
    FluxPhase { phase: String, phase_id: u32 },
    FluxBarrier { barrier: String },
    Glyph { name: String, glyph_id: u32 },
    StreamChunk { size: u32 },
    Noop,
}

pub fn build(parse_tree: ParseTree) -> anyhow::Result<Program> {
    builder::build(parse_tree)
}

pub fn lower(program: &mut Program, _registry: &Registry) -> anyhow::Result<()> {
    program.nodes.retain(|n| !matches!(n, Node::Noop));
    Ok(())
}

pub fn graph(_program: &mut Program) -> anyhow::Result<()> {
    Ok(())
}

pub fn merge(program: &mut Program) -> anyhow::Result<()> {
    let mut merged = Vec::with_capacity(program.nodes.len());
    let mut previous: Option<&Node> = None;

    for node in &program.nodes {
        if previous.is_some_and(|p| std::mem::discriminant(p) == std::mem::discriminant(node)) {
            continue;
        }
        merged.push(node.clone());
        previous = Some(node);
    }

    program.nodes = merged;
    Ok(())
}
