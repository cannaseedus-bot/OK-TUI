use serde::Deserialize;

#[derive(Clone, Debug, Deserialize)]
pub struct Registry {
    pub glyphs: serde_json::Value,
    pub flux: serde_json::Value,
}

pub fn load_all() -> anyhow::Result<Registry> {
    let glyphs = std::fs::read_to_string("agl.registry.json")?;
    let flux = std::fs::read_to_string("flux.registry.json")?;

    Ok(Registry {
        glyphs: serde_json::from_str(&glyphs)?,
        flux: serde_json::from_str(&flux)?,
    })
}
