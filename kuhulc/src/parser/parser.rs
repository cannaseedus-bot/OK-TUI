use super::Token;

#[derive(Clone, Debug)]
pub struct ParseTree {
    pub entries: Vec<String>,
}

pub fn parse(tokens: Vec<Token>) -> anyhow::Result<ParseTree> {
    Ok(ParseTree {
        entries: tokens.into_iter().map(|t| t.lexeme).collect(),
    })
}
