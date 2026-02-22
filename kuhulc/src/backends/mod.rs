use crate::ast::Program;

pub trait Backend {
    fn emit(&self, program: &Program, output: &str) -> anyhow::Result<()>;
}

pub mod c;
pub mod gpu;
pub mod scxq2;

pub use c::CBackend;
pub use gpu::GPUBackend;
pub use scxq2::SCXQ2Backend;
