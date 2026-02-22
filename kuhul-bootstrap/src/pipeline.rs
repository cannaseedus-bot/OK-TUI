use crate::cli::Target;
use std::{fs, path::Path};

#[derive(Debug, Clone)]
pub enum Op {
    Init(u32),
    Phase(String),
    Barrier(String),
    Glyph(String),
}

#[derive(Debug, Clone)]
pub struct Program {
    pub ops: Vec<Op>,
}

pub fn compile(input: &Path) -> std::io::Result<Program> {
    let source = fs::read_to_string(input)?;
    let mut ops = Vec::new();

    for raw in source.lines() {
        let line = raw.trim();
        if line.is_empty() || line.starts_with("//") {
            continue;
        }
        if line.starts_with("@flux.init") {
            let rate = extract_num(line).unwrap_or(60);
            ops.push(Op::Init(rate));
        } else if line.starts_with("@flux.phase") {
            let phase = extract_string(line).unwrap_or_else(|| "compute".to_string());
            ops.push(Op::Phase(phase));
        } else if line.starts_with("@flux.barrier") {
            let barrier = extract_string(line).unwrap_or_else(|| "gpu".to_string());
            ops.push(Op::Barrier(barrier));
        } else if line.starts_with("⟁") {
            let target = extract_string(line).unwrap_or_else(|| "unknown".to_string());
            ops.push(Op::Glyph(target));
        }
    }

    Ok(Program { ops })
}

pub fn emit(program: &Program, target: Target, deterministic: bool) -> String {
    match target {
        Target::C => emit_c(program, deterministic),
        Target::Gpu => emit_gpu(program, deterministic),
        Target::Scxq2 => emit_scxq2(program, deterministic),
    }
}

fn emit_c(program: &Program, deterministic: bool) -> String {
    let mut out = String::from(
        "#include \"../runtime/agl_runtime.h\"\n\nint main(void) {\n",
    );

    for op in &program.ops {
        match op {
            Op::Init(rate) => {
                out.push_str(&format!("    // init rate={}\n", rate));
            }
            Op::Phase(name) => out.push_str(&format!("    agl_flux_phase_enter(\"{}\");\n", name)),
            Op::Barrier(name) => out.push_str(&format!("    agl_flux_barrier(\"{}\");\n", name)),
            Op::Glyph(name) => out.push_str(&format!("    agl_glyph_exec(\"{}\");\n", name)),
        }
    }

    if deterministic {
        out.push_str("    // deterministic replay seed: 0\n");
    }

    out.push_str("    return 0;\n}\n");
    out
}

fn emit_gpu(program: &Program, deterministic: bool) -> String {
    let mut out = String::from("// WebGPU bootstrap output\nexport const program = [\n");
    for op in &program.ops {
        out.push_str(&format!("  {:?},\n", op));
    }
    out.push_str("] as const;\n");
    if deterministic {
        out.push_str("export const deterministicReplay = true;\n");
    }
    out
}

fn emit_scxq2(program: &Program, deterministic: bool) -> String {
    let mut out = String::from("SCXQ2\n");
    for (i, op) in program.ops.iter().enumerate() {
        out.push_str(&format!("{}|{:?}\n", i, op));
    }
    if deterministic {
        out.push_str("replay=deterministic\n");
    }
    out
}

fn extract_string(line: &str) -> Option<String> {
    let start = line.find('"')? + 1;
    let end = line[start..].find('"')? + start;
    Some(line[start..end].to_string())
}

fn extract_num(line: &str) -> Option<u32> {
    let digits: String = line.chars().filter(|c| c.is_ascii_digit()).collect();
    digits.parse().ok()
}
