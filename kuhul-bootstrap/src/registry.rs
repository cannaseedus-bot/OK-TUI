use std::{fs, path::Path};

pub struct RegistryCounts {
    pub glyphs: usize,
    pub opcodes: usize,
}

pub fn load_counts(base: &Path) -> Result<RegistryCounts, String> {
    let glyph_path = base.join("agl.registry.json");
    let flux_path = base.join("flux.registry.json");
    let glyph_raw = fs::read_to_string(glyph_path).map_err(|e| e.to_string())?;
    let flux_raw = fs::read_to_string(flux_path).map_err(|e| e.to_string())?;

    Ok(RegistryCounts {
        glyphs: glyph_raw.matches("\"@id\"").count(),
        opcodes: flux_raw.matches("\"@id\"").count(),
    })
}
