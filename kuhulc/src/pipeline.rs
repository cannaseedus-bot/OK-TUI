use crate::ast;
use crate::registry::Registry;

pub struct CompilerPipeline {
    registry: Registry,
}

impl CompilerPipeline {
    pub fn new(registry: Registry) -> Self {
        Self { registry }
    }

    pub fn compile(&mut self, path: &str) -> anyhow::Result<ast::Program> {
        let source = std::fs::read_to_string(path)?;
        let tokens = crate::parser::lex(&source)?;
        let parse_tree = crate::parser::parse(tokens)?;

        let mut program = ast::build(parse_tree)?;

        ast::validator::validate(&program, &self.registry)?;
        ast::lower(&mut program, &self.registry)?;
        ast::graph(&mut program)?;
        ast::merge(&mut program)?;
        ast::validator::validate_final(&program)?;

        Ok(program)
    }
}
