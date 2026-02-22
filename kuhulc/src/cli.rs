use clap::Parser;

#[derive(Debug, Parser)]
pub struct CompilerConfig {
    #[arg(short, long)]
    pub input: String,

    #[arg(short, long)]
    pub output: String,

    #[arg(short, long)]
    pub target: String,
}

impl CompilerConfig {
    pub fn parse() -> Self {
        Parser::parse()
    }
}
