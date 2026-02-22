use crate::ast::{Node, Program};
use crate::registry::Registry;

pub fn validate(program: &Program, registry: &Registry) -> anyhow::Result<()> {
    let has_glyph_registry = registry.glyphs.is_object();
    let has_flux_registry = registry.flux.is_object();

    anyhow::ensure!(
        has_glyph_registry && has_flux_registry,
        "MX2⟁☣ validation failed: registry must contain object maps"
    );

    anyhow::ensure!(
        program.nodes.iter().all(|n| !matches!(n, Node::Noop)),
        "MX2⟁☣ validation failed: unknown token in program"
    );

    Ok(())
}

pub fn validate_final(program: &Program) -> anyhow::Result<()> {
    anyhow::ensure!(
        !program.nodes.is_empty(),
        "final validation failed: empty program"
    );
    Ok(())
}
